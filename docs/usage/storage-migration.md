# 存储迁移指南：更换 S3 服务商与本地 ⇄ 对象存储互迁

本文说明两类常见运维场景的可行做法与注意事项：

1. **更换 S3 兼容存储**（换服务商、换桶、换 Endpoint/CDN 等，仍使用对象存储）。
2. **本地存储**与**对象存储（S3）**之间的**双向迁移**。

> **重要**：当前版本**没有**提供「一键迁移」或图形化迁移向导。下列流程需要管理员在理解数据格式的前提下，自行用对象存储工具与数据库操作完成。操作前请**完整备份**数据库与文件。

> **视频等大文件建议放对象存储（S3）**：Ech0 支持上传图片、音频、视频（视频默认单文件上限 64 MiB，可用 `ECH0_UPLOAD_VIDEO_MAX_SIZE` 调整）。视频远大于图片，而**快照导出会把整个 `data/` 打包成 zip**——本地存放大量视频会同时撑大磁盘占用与每次备份体积。

---

## 1. 必读：Ech0 如何表示与定位文件

完整设计见 `docs/dev/file-system-design.md`。这里只列运维需要知道的结论。

### 1.1 五张表

| 表 | 作用 |
|----|------|
| `storages` | 已启用的存储实例：`id`（`local` / `s3`）、`driver`（`local` / `s3`）、`config`（JSON）、`is_default`。**每字节数据都有且只有一个所属存储**（`files.storage_id`） |
| `files` | 文件行：`kind`（`owned` 自有字节 / `blob` 入库小文件 / `ref` 外链）、`access`（`public` / `owner` / `private`）、`status`（`pending` / `ready`）、`storage_id` + `object_key`（owned）、`blob_bytes`（blob）、`ref_url`（ref）、`name`、`content_type`、`size`、`width/height`、`uploader_id` |
| `file_blobs` | `kind=blob` 的字节（≤ 512 KiB，入库与行同事务） |
| `file_links` | 引用关系：哪条 Echo / 用户 / 实例槽位（`owner_kind` + `owner_id` + `role` + `sort_order`）在用这个文件 |
| `file_gc` | 待删除对象的重试队列（删行成功、删字节失败时写入） |

`files.object_key` 是**存储内的相对键**，形如 `2026/11/<uuidv7>.png`，由 `storage.KeyFor(file_id, ext)` 生成，写入后不再变动：

- 本地：`<DataRoot>/<object_key>`
- 对象存储：`<PathPrefix>/<object_key>`（`PathPrefix` 来自该 `storages` 行的 `config.path_prefix`，未配置则为空）

`blob` 与 `ref` 文件没有对象键：前者字节在库里，后者只存外链 URL。

### 1.2 访问路径只有一个

所有字节都由网关提供：

```
GET|HEAD /api/files/{file_id}/content
```

- `access=public`（实例 logo、头像）：匿名可读，`Cache-Control: public, max-age=31536000, immutable`；若该存储支持预签名（S3），可直接 302 到签名 URL 或 CDN。
- `access=owner`（Echo 附件、上传后的草稿）：可见性由引用它的业务决定——正文挂在公开 Echo 上匿名可读，落在私密 Echo 上就需要登录；缓存头随之变化（匿名 `max-age=0`、登录 `no-store`）。
- `access=private`（草稿）：仅上传者与管理员可读，恒为 `no-store`。
- 无权限与不存在**同样是 404**，不暴露文件是否存在。

`ETag` 是 `"<file_id>"`，支持 `Range` 与 `304`。旧的 `/api/files/stream/*`、`/api/files/temp/*`、`/api/files/public/*` 静态命名空间已全部移除。

### 1.3 前端拿到的 URL

接口返回的 `file.url` 是派生值 `/api/files/<file_id>/content`，**不写入磁盘**。换桶、换 CDN、换 `PathPrefix` 之后不需要批量修库，也不需要改前端。

---

## 2. 场景一：更换 S3 服务商 / 桶 / Endpoint

### 2.1 目标

数据仍在对象存储里，只是换了桶或服务商；文件 ID、`object_key` 都不变。

### 2.2 核心原则

- **先搬字节，再改配置**：`storages` 行的 `config` 是连接参数，`files.object_key` 是位置；只要能连上新桶并且对象键一致，文件就还是那些文件。
- **`storage_id` 决定去哪读**：让所有 owned 行仍指向 `s3` 即可，无需逐行更新。
- **`PathPrefix` 必须一致**（或在新桶里照搬旧的前缀布局），否则对象键会整体错位。

### 2.3 推荐操作顺序（降低停机风险）

1. 在面板「存储设置」里**先不要**改配置；先把旧桶的整个前缀（`PathPrefix/`）原样同步到新桶：
   ```bash
   rclone sync old:bucket/prefix new:bucket/prefix --checksum
   ```
2. 抽样核对：随机取几条 `SELECT id, object_key FROM files WHERE kind='owned' AND storage_id='s3' LIMIT 20;`，确认 `new:bucket/prefix/<object_key>` 可读。
3. 在面板里把 S3 的 Endpoint / AccessKey / SecretKey / Bucket / Region / PathStyle 改成新值并保存。保存时 Ech0 会用新配置做一次连通性探测，成功才落库。
4. 打开任一条带图片的 Echo，确认图片正常；再打开面板文件列表，确认缩略图正常。
5. 旧桶保留一段时间（只读）作为回滚点。

### 2.4 常见踩坑

- **`PathPrefix` 改了但字节没搬**：表现为全部 404（对象键找不到）。此时把前缀改回去即可。
- **Endpoint 需要 path-style**：大多数自建 S3（MinIO、Ceph）需要开启 `use_path_style`；virtual-hosted 风格在裸 IP 下不可用。
- **CDN 缓存旧 302**：public 文件会 302 到签名 URL，CDN 若缓存了 302，切桶后可能仍指向旧域名；清除 CDN 上 `/api/files/` 的缓存。
- **改了 `storage_id` 却没搬数据**：`files.storage_id` 指向的存储里没有对象键，网关返回 404（并写入日志 `file.gateway`），不会回退去别的存储找。

---

## 3. 场景二：本地存储 ⇄ 对象存储（S3）互迁

### 3.1 可行性

可行，但**必须同时搬字节、改 `storages`、改 `files.storage_id`**——存储所属是写死在行上的（这是设计的一部分：位置写入即固定，读路径不做隐式回退）。

### 3.2 路径对应关系

| 位置 | 本地 | 对象存储 |
|------|------|----------|
| 桶/目录 | `data/files/` | `<PathPrefix>/`（默认空） |
| 对象 | `data/files/<object_key>` | `s3://<bucket>/<PathPrefix>/<object_key>` |
| 行 | `files.storage_id='local'` | `files.storage_id='s3'` |

`object_key` 两边完全一致，所以搬迁就是**整棵目录树复制**，不需要改名：

```bash
aws s3 sync data/files/ s3://my-bucket/prefix/ --exact-timestamps
```

### 3.3 本地 → 对象存储

1. 面板里把 S3 配好并保存（此时 `storages` 表多一行 `id='s3'`），**先不要**设为默认。
2. 停服务，复制字节：
   ```bash
   aws s3 sync data/files/ s3://my-bucket/prefix/
   ```
3. 改库（SQLite）：
   ```sql
   UPDATE files SET storage_id = 's3' WHERE kind = 'owned' AND storage_id = 'local';
   ```
4. 面板里把默认存储切到 S3（影响新上传）。
5. 起服务，抽查若干 Echo 的图片；确认无误后再清理本地 `data/files/`。

### 3.4 对象存储 → 本地

同样三步，方向反过来：配好本地（`DataRoot` 由环境变量决定，`storage_id='local'`），把桶内前缀下载到 `data/files/`，`UPDATE files SET storage_id='local' WHERE kind='owned' AND storage_id='s3';`，再把默认存储切回本地。

### 3.5 两种方向共同的注意事项

- **`file_gc` 也可能指向旧存储**：迁移期间若有删除失败的补偿任务，`storage_id` 还指着旧存储名。先跑完 GC（或清空整张 `file_gc` 表，确认旧存储里没有残留对象）再迁。
- **`blob` 与 `ref` 不参与迁移**：前者字节在数据库里，后者本来就在外部。
- **默认存储只影响新上传**：切换默认不会搬任何东西，历史文件永远从它自己的 `storage_id` 读。
- **胶囊/快照不受影响**：它们按文件内容导出，与存储拓扑无关。

---

## 4. 校验与回滚

| 检查项 | 方法 |
|--------|------|
| 行数一致 | `SELECT storage_id, COUNT(*) FROM files WHERE kind='owned' AND status='ready' GROUP BY storage_id;` |
| 对象存在 | 按 5% 抽样 `object_key`，在目标存储里 `stat` |
| 网关可用 | 挑一条公开 Echo 的图片：`curl -I https://<site>/api/files/<id>/content` 应 200 且带 `ETag` |
| 回滚 | 把 `files.storage_id` 改回原值、默认存储切回原值、服务恢复旧配置。旧存储对象在清理前都还在 |

---

## 5. 相关代码与配置索引

| 内容 | 位置 |
|------|------|
| 存储驱动与 SPI | `internal/storage/{storage.go,fsstore.go,mem.go}`；`Store` 只有 `Put/Open/Stat/Delete`，可选 `Presigner`/`RangeOpener` |
| 存储注册表（启动时从 `storages` 表装配） | `internal/storage/registry.go` |
| 文件域（上传、网关、GC、引用） | `internal/service/file/`、`internal/repository/file/` |
| 网关 HTTP | `internal/handler/file/file.go`（`GET|HEAD /api/files/{id}/content`） |
| 表结构（CHECK / 部分唯一索引 / 外键） | `internal/database/migration/file_resource_migrator.go` |
| S3 配置字段 | `storages.config`：`endpoint`、`access_key`、`secret_key`、`bucket`、`region`、`use_ssl`、`use_path_style`、`path_prefix`、`provider`、`cdn_url`（非空时 public 文件 302 到 `<cdn_url>/<object_key>`） |

SQLite 连接参数包含 `_foreign_keys=on`：`file_links`、`file_blobs` 的级联删除是真实生效的，手工改库时请留意顺序（先删 link，再删 files）。

---

## 6. 文档版本

与 `docs/dev/file-system-design.md`（文件资源系统设计）同步维护；两者冲突时以设计文档为准。
