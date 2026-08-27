# 存储迁移指南：更换 S3 服务商与本地 ⇄ 对象存储互迁

本文说明在 Ech0 中两类常见运维场景的可行做法与注意事项：

1. **更换 S3 兼容存储**（换服务商、换桶、换 Endpoint/CDN 等，仍使用对象存储）。
2. **本地文件存储**与**对象存储（S3）**之间的**双向迁移**。

> **重要**：当前版本**没有**提供「一键迁移」或图形化迁移向导；下列流程需在理解数据格式的前提下，由管理员自行使用对象存储工具、脚本与数据库操作完成。操作前请**完整备份**数据库与文件。

> **视频等大文件建议放对象存储（S3）**：Ech0 支持上传图片、音频、视频（视频默认单文件上限 64 MiB，可用 `ECH0_UPLOAD_VIDEO_MAX_SIZE` 调整）。视频体积远大于图片，且**快照导出会把整个 `data/` 目录打包成 zip**——本地存放大量视频会同时撑大磁盘占用与每次备份的体积。若计划频繁上传视频，建议将存储切换到对象存储（S3），本地仅保留数据库快照。

---

## 1. 必读：Ech0 如何表示与定位文件

### 1.1 `files` 表中的关键字段

业务上每条文件记录对应一个逻辑文件，核心字段包括（详见 `internal/model/file/file.go`）：

| 字段 | 含义 |
|------|------|
| `key` | **扁平存储键**（如 `uid8_时间戳_随机串.png`），上传时由服务端生成，**不是**从对象存储回写的完整 Object Key。 |
| `storage_type` | `local`（本地）或 `object`（对象存储）等；决定历史记录「登记」在哪一类后端。 |
| `provider` / `bucket` | 对象存储的**元数据**（如提供商、桶名），用于唯一性与展示；**按文件 ID 代理读取时，运行时并不用这两项去连接旧桶**（见下文）。 |
| `url` | **前端直链快照**：创建/上传时根据**当时**配置拼出的可访问 URL；换 CDN、换桶域名后可能**过期**，需视业务决定是否批量更新。 |

唯一索引 `idx_file_route` 包含 `storage_type`、`provider`、`bucket`、`key`，批量改元数据时注意避免冲突。

### 1.2 扁平 `key` 与实际磁盘 / 桶内路径

数据库只存**扁平 `key`**。真正落到磁盘或桶里的相对路径由 **VireFS Schema** 的 `Resolve` 规则决定（`internal/storage/schema.go`）：

- 按**扩展名**分到 `images/`、`audios/`、`videos/`、`documents/`，其余走 `files/`。
- 文件名部分就是整条扁平 `key`（例如 `images/abc_123_xxx.png` 中的 `abc_123_xxx.png` 即库里的 `key`）。

对象存储还会在 VireFS 层叠加 **`PathPrefix`**（管理后台 S3 设置或环境变量中的路径前缀），因此：

**桶内 Object Key（概念上）≈ `PathPrefix`（若配置） + `schema.Resolve(扁平 key)`**

本地同理，在 **`DataRoot`** 下使用 **`schema.Resolve(扁平 key)`** 作为相对路径（默认 `DataRoot` 为 `data/files`，见 `internal/config/config.go` 与 `internal/storage/manager.go` 合并逻辑）。

#### 举例：上传一张 PNG 图片之后，各层长什么样？

下面用一组**虚构但贴近真实格式**的数据，把「库里的 `key` → `schema.Resolve` → `PathPrefix` → 最终落盘/落桶」串成一条线，便于对照控制台里的对象列表或本地目录。

**假设：**

- 管理员上传文件 `screenshot.png`，服务端生成的扁平 **`key`**（入库）为：  
  `a1b2c3d4_1735689600_deadbeef.png`  
  （格式大致为：用户 ID 缩略 + Unix 时间戳 + 随机后缀 + 扩展名，见 `internal/storage/keygen.go`。）
- **`PathPrefix`**（管理后台 S3 设置或 `ECH0_S3_PATH_PREFIX`）配置为：`prod/ech0`（无首尾斜杠，合并时会规范化）。
- **对象存储**：桶名为 `my-ech0-bucket`，Endpoint 为 `s3.example.com`，**未**配置 CDN（`CDNURL` 为空）。
- **本地存储**（若走本地）：`DataRoot` 为默认的 `data/files`。

**第一步：`schema.Resolve(扁平 key)`**

`.png` 命中图片规则，得到**相对路径（含类型目录）**：

```text
images/a1b2c3d4_1735689600_deadbeef.png
```

**第二步：对象存储 —— 桶内 Object Key（逻辑）**

VireFS 会先加 **`PathPrefix/`**，再交给 `ObjectKeyFunc` 解析后的路径。概念上整条 Key 为：

```text
prod/ech0/images/a1b2c3d4_1735689600_deadbeef.png
```

在 S3 控制台或 `aws s3 ls` 里，你应能在桶 `my-ech0-bucket` 下看到上述路径（具体是否带 `prod/ech0` 前缀以你配置的 `PathPrefix` 为准；**未配置** `PathPrefix` 时则仅为 `images/...`）。

**第三步：拼进 `files.url` 的直链快照（无 CDN 时）**

`internal/storage/provider.go` 中无 CDN 时，会把 **Endpoint + 桶名** 与 **`PathPrefix` + `Resolve` 结果** 拼成对外 URL，形如：

```text
https://s3.example.com/my-ech0-bucket/prod/ech0/images/a1b2c3d4_1735689600_deadbeef.png
```

（若配置了 **`CDNURL`**，则主机部分会换成 CDN 域名，路径仍为 `PathPrefix` + `images/...`。）

**第四步：同一文件若存在本地磁盘上**

相对路径仍是 **`schema.Resolve` 的结果**，落在 `DataRoot` 下：

```text
data/files/images/a1b2c3d4_1735689600_deadbeef.png
```

注意：**本地布局没有** S3 的 **`PathPrefix`** 这一层；`PathPrefix` 只作用在对象存储侧。因此做「本地 ⇄ S3」迁移时，对比的是：

- 本地：`DataRoot` + `images/...`
- 桶：`PathPrefix` + `images/...`（与当前 S3 配置一致）

**小结对照表**

| 层级 | 示例值 |
|------|--------|
| 数据库 `key` | `a1b2c3d4_1735689600_deadbeef.png` |
| `schema.Resolve(key)` | `images/a1b2c3d4_1735689600_deadbeef.png` |
| 桶内 Object Key（含前缀） | `prod/ech0/images/a1b2c3d4_1735689600_deadbeef.png` |
| 本地文件完整路径 | `data/files/images/a1b2c3d4_1735689600_deadbeef.png` |

换桶或互迁时，只要保证「**新环境**下按上表同一规则还能找到字节」，并与 `files` 里存的 `key` / `storage_type` / `url` 策略一致即可。

### 1.3 运行时如何读文件（`File` 模型、`local` / `object`、与前端路由对照）

同一套 `File` 记录在系统里可能通过**多条 HTTP 路径**被访问；迁移时「改桶了代理还能下」和「首页图是否裂图」对应的路径**不一定相同**。下面按**路由**说明服务端行为与**当前前端**实际用法。

#### 1.3.1 访问路径总览（建议先读）

| 访问方式 | 典型 URL 形态 | 服务端实现要点 | 当前前端主要用途 |
|----------|----------------|----------------|------------------|
| **本地静态文件** | `/api/files/images/…`（相对路径，再拼站点 base） | `Engine.Static("api/files", DataRoot)`，URL 路径直接映射磁盘（`internal/router/modules.go`） | **Echo 时间线、画廊、编辑器**等：通过 `getFileUrl` 使用入库的 `url`，本地多为本表第一行 |
| **对象存储直链** | `https://` 桶域名或 **CDN** + 路径 | **无 Ech0 读盘**：浏览器直连 S3/CDN | 同上；`getFileUrl` 见绝对 URL 则原样使用（`web/src/utils/other.ts`） |
| **按文件 ID 取流** | `GET /api/file/:id/stream` | `StreamFileByID`：用 **`key` + 当前存储配置** `Get`，**不用** `url`（`internal/service/file/file.go`） | **仅管理端**：`TheStorageFileList.vue` 中「下载」调用 `fetchDownloadFileById`（`web/src/service/api/file.ts`） |
| **按存储路径取流** | `GET /api/file/stream?storage_type=…&path=…` | `StreamFileByPath`：管理员按路径读 | **仅管理端**：同上文件树，无 `file_id` 时 `fetchDownloadFileByPath` 兜底 |
| **外链重定向** | 同上 ID 流式路由但 `storage_type=external` | **302** 到 `File.url` | 外链文件场景 |

说明：

- **`/api/file/.../stream` 与 `/api/file/stream` 挂在需鉴权的路由组上**（需 `ScopeFileRead` 等，见 `internal/router/file.go`），与**直接暴露的静态** `api/files` 不是同一套中间件策略；管理端下载走 stream 时会带 token（见 `downloadFile` 与前端请求封装）。
- **RSS / 服务端拼 HTML**（如 `internal/service/common/common.go` 中 Feed）若使用 **`ef.File.url`**，则与浏览器「看 `url`」一致，**不经过** stream。

#### 1.3.2 前台与 RSS：绝大多数走 `File.url`（`getFileUrl`）

上传或创建文件时，接口返回并持久化 **`File.url`**。前端统一通过 **`getFileUrl` / `getImageUrl`**（`web/src/utils/other.ts`）解析：

1. **`url` 为 `http(s)://` 绝对地址**（对象存储入库时常见）  
   - 浏览器**直接请求该 URL**，即**按入库时的直链快照**访问桶或 CDN。  
   - **换 S3、换 CDN、换自定义域名**后，若库中 **`url` 未更新**，容易出现裂图或旧域名。

2. **`url` 为相对路径**（本地常见：`/api/files/...`）  
   - 与 **`VITE_SERVICE_BASE_URL`** 拼接后请求本站。  
   - 对应 **§1.3.1** 表格第一行：由 **Gin Static** 直接读 **`DataRoot`** 下文件，**不经过** `StreamFileByID`。

因此：**时间线/卡片上的图、音频封面等，默认不是「stream API」**；**对象**以 **HTTPS 直链**为主，**本地**以 **`/api/files/...` 静态**为主。

#### 1.3.3 本地静态映射（与 stream 的区分）

- 注册方式：`ctx.Engine.Static("api/files", root)`，`root` 为配置中的 **`DataRoot`**（默认 `data/files`）。  
- 浏览器请求 `GET /api/files/images/xxx.png` 时，等价于读磁盘 `data/files/images/xxx.png`（路径与 `schema.Resolve(key)` 一致）。  
- **这是前台本地媒体的主路径**；除非产品改为全部走代理，否则**不必**用 `/api/file/:id/stream` 展示本地图。

#### 1.3.4 `stream` API：服务端行为与前端调用面

**按 ID：`GET /api/file/:id/stream`**

- 处理函数：`StreamFileByID`（`internal/service/file/file.go`）。  
- 对 **`local` / `object`**：使用**当前**合并配置 + 记录中的 **`key`** 调用 `Get`；**不使用** `url` 读字节；**不使用**行内 **`provider` / `bucket`** 选择连接（换桶后对象路径正确时，代理仍可能成功）。  
- 对 **`external`**：**302** 到 `File.url`。

**按路径：`GET /api/file/stream?storage_type=…&path=…&name=…`**

- 处理函数：`StreamFileByPath`，供**已知存储路径**、但树节点上可能没有绑定 `file_id` 时的兜底下载。

**前端谁在用（截至当前仓库）**

| API | 封装函数 | 使用位置 |
|-----|----------|----------|
| `/file/:id/stream` | `fetchDownloadFileById` | `web/src/views/panel/modules/TheSetting/TheStorageFileList.vue`（存储文件树 · 下载） |
| `/file/stream?…` | `fetchDownloadFileByPath` | 同上，无 `file_id` 时 |

`web/src/lib/file/api/adapter.ts` 中的 **`buildStreamUrl`**（手动拼 `/file/:id/stream` + token）**已导出，但全项目无其它引用**，可视为预留；**上传队列**（`file-queue.ts`）只调用 `uploadFile`，**不使用** stream。

**结论**：**stream 面向「带鉴权的按 ID/路径下载」**（当前实现集中在**设置 → 存储文件列表**）；**不是** Echo 主站看图路径。

#### 1.3.5 与迁移、换 S3 的对应关系

| 你关心的现象 | 优先检查 |
|--------------|----------|
| **首页/时间线裂图、RSS 图裂** | 多为 **`url` 仍是旧域名或旧桶 URL**；对象存储需更新 **`url`** 或修正 CDN/桶策略 |
| **管理端文件树下载失败** | **`key` + 当前 S3/本地配置** 能否读到对象；与 §1.2 路径规则是否一致 |
| **仅代理能下、直链不能（或反之）** | 直链依赖 **`url`/桶权限/CDN**；代理依赖 **`key`** 与存储配置，两套问题独立排查 |

---

## 2. 场景一：更换 S3 服务商 / 桶 / Endpoint

### 2.1 典型目标

- 从 A 厂商（或自建 MinIO）迁到 B 厂商；
- 或同一厂商下更换桶、Region、自定义域名/CDN；
- 数据库仍为同一套 SQLite（或同一实例），希望**尽量少改库**或**只改必要字段**。

### 2.2 核心原则

1. **对象字节**：在新桶中的 **Object Key** 必须与**新配置**下的  
   `PathPrefix + schema.Resolve(key)` **一致**（与旧环境对比时，注意旧环境的 `PathPrefix` 是否不同）。
2. **应用配置**：在管理后台或环境变量中更新 Endpoint、密钥、桶名、`PathPrefix`、`CDNURL`、`UseSSL` 等，并保证 **`ECH0` 进程重载或重启**后使用新配置（`StorageManager` 会从 DB/环境合并配置）。
3. **数据库**：
   - **最低限度**：迁完对象、改好配置后，先验证**按文件 ID 访问**是否正常。
   - **建议**：将 `files` 表中仍指向对象存储的行的 **`provider`、`bucket`、必要时 `url`** 更新为与新环境一致，避免管理列表、统计与外链混乱。
4. **`url` 字段**：若前端、RSS、Webhook 等依赖**库中直链**，换 CDN 或桶公共访问域名后，应**批量重算或按新规则更新 `url`**（新 URL 的拼接逻辑可参考 `buildS3PathURLResolver`，见 `internal/storage/provider.go`：含 CDN、Endpoint+桶、`PathPrefix` 与 `schema.Resolve` 后的路径）。

### 2.3 推荐操作顺序（降低停机风险）

以下顺序可按实际窗口调整；核心是**先保证新桶里已有正确对象，再切流量与配置**。

1. **备份**：SQLite 文件、当前 S3 设置快照、旧桶对象列表（或整桶同步到本地 staging）。
2. **确定路径规则**：列出旧环境与新环境的 `PathPrefix`、是否使用 CDN；对**每条** `files.key` 计算旧桶源路径与新桶目标路径是否**一一对应**（同一 `key` 下 `Resolve` 结果相同，仅桶/前缀/域名变）。
3. **拷贝对象**：使用厂商控制台、CLI（如 `aws s3 sync`、`rclone`）或批量复制，将对象从旧桶复制到新桶，**保持相对 Key 一致**（含前缀与 `images/` 等目录）。大桶时注意限速、校验与失败重试。
4. **在测试环境或只读验证**：用新凭证 + 新桶配置启动实例（或临时改配置），抽查若干 `key` 对应的代理访问与直链。
5. **切换生产配置**：更新管理后台 S3 设置或环境变量，重启/重载服务。
6. **更新数据库（建议）**：批量 `UPDATE` `provider`、`bucket`、`url`（按需）。
7. **回归**：前台发图、RSS、备份任务（若也使用 S3）等。

### 2.4 常见踩坑

| 问题 | 说明 |
|------|------|
| **只改配置未迁对象** | 新桶为空或路径不一致，读文件 404。 |
| **`PathPrefix` 与拷贝路径不一致** | 配置里是 `upload/`，对象却放在桶根，或反之。 |
| **误以为必须改 `key`** | 一般**不需要**改 `key`；除非你在新桶刻意用了另一套命名，则需同步改库或做映射（项目内无通用映射层）。 |
| **忽略 `url`** | 界面仍显示旧 CDN 链接或 403；需更新 `url` 或改为使用本站代理 URL。 |
| **迁移期间双写** | 切换窗口内应避免同一 `key` 在旧桶与新桶被不同版本覆盖；建议维护窗口内只读或停写。 |
| **`other` 类服务连不上（寻址方式）** | provider 选 `aws` / `minio` / `r2` 时寻址方式由预设决定（minio/r2 为 path-style）；选 **`other`** 默认 virtual-hosted，若目标服务（Ceph、Garage、SeaweedFS 等自建 S3 兼容服务）只支持 path-style，在管理后台打开「Path-Style 访问」开关（或设 `ECH0_S3_USE_PATH_STYLE=true`）。 |
| **上传成功但图片打不开（COS/OSS 直链形状）** | 腾讯 COS、阿里 OSS 等**只接受 virtual-hosted 直链**（`bucket.endpoint/key`）。无 CDN 域名时，Ech0 生成的公开直链会**跟随 SDK 寻址方式**：`other` 默认按 virtual-hosted 拼 `https://bucket.endpoint/key`，开了 Path-Style 开关或 minio/r2 则拼 `endpoint/bucket/key`。因此这类服务保持 `other`、**不要**开 Path-Style 开关即可；Endpoint 填**地域域名**（如 `cos.<region>.myqcloud.com`，不含 bucket）。注意 `File.url` 是**写入时快照**，改配置只影响此后的新上传，存量记录需重刷 `url` 或改填 CDN 域名。 |

---

## 3. 场景二：本地存储 ⇄ 对象存储（S3）互迁

### 3.1 可行性说明

- **逻辑上**：本地与对象存储使用**同一套** `FileSchema`（`schema.Resolve`），同一 `key` 在两侧的**相对路径部分一致**（都是 `Resolve(key)`），差异在于根是 **`DataRoot`** 还是 **桶 + `PathPrefix`**。
- **工程上**：项目**未提供**内置「本地 ⇄ S3」迁移命令；需要自行复制文件并**更新 `files` 表**中的 `storage_type` 及关联字段。

### 3.2 路径对应关系（用于编写脚本）

设扁平文件名为数据库中的 `key`，则：

- **本地绝对路径（概念）**  
  `本地根目录 = 配置中的 DataRoot（默认 `data/files`）`  
  `相对路径 = schema.Resolve(key)`  
  **完整路径 ≈ `filepath.Join(DataRoot, Resolve 结果)`**（注意操作系统路径分隔符）。

- **对象存储 Object Key（概念）**  
  **`trim(PathPrefix) + "/" + schema.Resolve(key)`**（与 VireFS `WithPrefix` + `WithObjectKeyFunc(schema.Resolve)` 一致；具体拼接以 VireFS 实现为准，迁移前应用**少量样本**在测试桶验证）。

迁移脚本应对每条 `files` 记录读取 `key`、`storage_type`，仅处理需要从一种后端迁到另一种的行。

### 3.3 本地 → 对象存储

**目标**：文件只保留在 S3，记录改为 `object`。

1. 备份数据库与 `data/files`。
2. 启用并正确配置 S3（管理后台或 `ECH0_S3_*` 等环境变量），确认 `PathPrefix` 与计划上传路径一致。
3. 对每条需迁移的记录：从本地读出文件，上传到目标桶的 **目标 Object Key**（按上一节规则计算）。
4. 校验对象大小与 Content-Type（可选）。
5. 在事务或分批中更新 SQLite：  
   - `storage_type = 'object'`  
   - 填写新的 `provider`、`bucket`  
   - 按当前配置重算或写入 **`url`**  
6. 确认应用读取正常后，**再删除**本地对应文件（或先改名目录做回滚备份）。

### 3.4 对象存储 → 本地

**目标**：文件只保留在本地磁盘，记录改为 `local`。

1. 备份数据库；确保磁盘空间与 `DataRoot` 权限足够。
2. 对每条需迁移的记录：从桶中按 **Object Key** 下载到  
   `DataRoot + "/" + schema.Resolve(key)`  
   （需创建子目录，与本地 FS 布局一致）。
3. 更新 SQLite：  
   - `storage_type = 'local'`  
   - `provider` / `bucket` 可置空或与项目约定一致（注意唯一索引是否允许空字符串，需与现有数据约束一致）  
   - 更新 **`url`** 为本地访问方式（例如本站 `/api/files/...` 类路径，以当前部署为准）
4. 验证代理访问后，再从桶中删除对象（或生命周期策略延后删除）。

### 3.5 互迁共同注意事项

| 项目 | 说明 |
|------|------|
| **`key` 一般不改** | 保持与现有 Echo、评论等关联一致；迁移的是**字节位置**与 **`storage_type`**。 |
| **批量与事务** | 大量文件时分批提交，避免长时间锁库；记录失败行以便重试。 |
| **唯一索引** | 更新 `provider`/`bucket` 时避免与已有行冲突。 |
| **TempFile** | 若有未确认的临时上传，先完成业务确认或清理逻辑，避免迁了一半的孤立记录。 |
| **双后端并存** | 配置允许「启用对象存储」与本地并存时，新上传走哪种由产品与上传接口决定；历史行仍以 `storage_type` 为准。 |

---

## 4. 校验与回滚建议

- **抽样**：随机抽取图片/音频，对比文件大小与 Content-Type；用浏览器或 curl 访问代理 URL。
- **清单**：导出 `files` 表中 `id, key, storage_type, url` 迁移前后对比。
- **回滚**：保留旧桶或旧 `data/files` 副本直至观察期结束；数据库恢复需配合文件状态一致。

---

## 5. 相关代码与配置索引（便于对照实现）

| 说明 | 位置 |
|------|------|
| `File` 模型 | `internal/model/file/file.go` |
| Schema（`Resolve` 规则） | `internal/storage/schema.go` |
| 本地 / S3 FS 与 `PathPrefix`、`ObjectKeyFunc` | `internal/storage/provider.go` |
| 存储配置合并（DB + 环境变量） | `internal/storage/manager.go` |
| 按 ID 流式读取（仅用当前配置 + `key`） | `internal/service/file/file.go`（`StreamFileByID`） |
| S3 设置键名 | `internal/model/common/common.go`（`S3SettingKey`） |
| 环境变量前缀 | `internal/config/config.go`（如 `ECH0_S3_*`） |

---

## 6. 文档版本与免责

- 本文基于当前仓库实现整理；**路径拼接细节以 VireFS 与运行时配置为准**，大规模迁移前务必在**测试环境**用真实 `key` 验证。
- 生产变更需由具备权限的运维执行；作者不对误操作导致的数据丢失负责。

如有条件，建议在项目后续版本中提供**官方迁移工具或只读校验脚本**，以减少人为错误。
