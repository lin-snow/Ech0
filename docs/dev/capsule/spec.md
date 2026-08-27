# Ech0 Capsule 规格 v1

> **状态：已实现（v1）。** 本文是 Capsule 格式与相关命令的**规范性定义**——只写「是什么」。
> 设计依据、备选方案与讨论见 [`capsule-design.md`](./capsule-design.md)；文中 Q 编号指向该文档 §9。
> 面向用户的操作指南见 [`../../usage/capsule.md`](../../usage/capsule.md)。
> 实现位于 `internal/capsule/`（核心格式）与 `internal/capsule/{export,check,importer,build}`，
> CLI 接线在 `cmd/capsule.go` + `internal/cli/capsule.go`，Web 接线在 `internal/migrator/capsule.go`
> + `internal/service/migrator` + `internal/job/runner`（见 §9.1）。

**用词约定**：**必须** / **禁止** = 违反即校验错误；**应当** = 违反产生警告；**可选** = 消费者不得因缺失而报错。

---

## 1. 术语

| 术语 | 定义 |
|---|---|
| **Capsule（胶囊）** | 一个符合本规格的目录（或其 zip 打包），自包含承载一个 Ech0 站点的全部公开内容 |
| **生产者** | 产出胶囊的一方：`ech0 export capsule`、手写、第三方转换器 |
| **消费者** | 读取胶囊的一方：`ech0 import capsule`、`ech0 build`、`ech0 check`、第三方工具 |
| **内容级往返** | `export → import` 后内容等价（非字节等价）；保真度边界见 §11 |

## 2. 目录布局

```text
<capsule>/
  ech0.yaml         # 必须。清单文件（§3）
  echoes/           # 必须（可为空目录）。Echo 内容（§4）
    <YYYY>/
      <file>.md
  comments.yaml     # 可选。评论快照（§5）
  files/            # 可选。媒体文件（§6），内部结构 mirror 本地存储 DataRoot
    images/ audios/ videos/ documents/ files/
```

- 顶层**禁止**出现规格未定义的文件或目录？——否：消费者**必须**忽略未知路径（前向兼容，§8）。
- 胶囊内所有引用**必须**使用相对路径、`/` 分隔符，**禁止** `..` 路径穿越（校验错误）。

## 3. 清单文件 `ech0.yaml`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `schema_version` | int | **必须**，当前为 `1` | 破坏性变更时递增 |
| `generator` | string | 可选 | 生产者标识，如 `ech0 v2.x.x`、`memos-converter 0.3` |
| `exported_at` | RFC3339 | 可选 | 导出时刻；手写胶囊通常缺省 |
| `site.site_title` | string | 应当提供 | `SystemSetting.SiteTitle` 原样 |
| `site.server_name` | string | 可选 | `ServerName` 原样 |
| `site.server_logo` | string | 可选 | `ServerLogo` **原样**（URL 字符串，不做本地化改写；含实例 URL 时归入 §7 警告。托管上传的 logo 字节因记录驱动导出本就在 `files/` 内，`/api/files/…` 相对引用在静态站可正常渲染） |
| `site.server_url` | string(URL) | 可选 | 原实例地址，溯源用；import 仅空填充 |
| `site.default_locale` | string | 可选 | 如 `zh-CN` / `en-US` |
| `site.ICP_number` | string | 可选 | 备案号 |
| `site.footer_content` / `site.footer_link` | string | 可选 | 自定义页脚 |
| `site.meting_api` | string(URL) | 可选 | 音乐扩展渲染所需 |
| `site.custom_css` / `site.custom_js` | string | 可选 | 非空时 `check` **应当**告警（第三方胶囊 = 执行对方代码） |
| `owner.username` | string | **必须** | 归属兜底：Echo 未标 `username` 时的默认作者 |
| `connects` | list | 可选 | 互联实例快照，元素为 `{url: string}` |
| `files` | list | 可选 | **未挂在任何 Echo 上的文件行**，元素形状与 §4.2 的 `files[]` 完全一致。见下 |

- `site.*` 子集遵循「渲染所需皆入，运维行为皆弃」：`AllowRegister` 等行为开关**禁止**入胶囊。
- `site.*` 键名 = `SystemSetting` 的 json tag **原样**（`site_title`/`server_logo`/`ICP_number`…）：import 时整块直接反序列化进 `SystemSetting`，零映射代码（「导出即转储」原则，见 §11）。
- 凭据（密码哈希、token、OAuth/Passkey、S3 密钥、SMTP、Agent 配置）**禁止**出现在胶囊任何位置。
- **`files` 块的存在理由**：库中 `files` 是独立表，而 frontmatter 只能表达「挂在这条 Echo 上的文件」。站点 logo、以及上传后没用上的附件，其字节会随记录驱动导出进 `files/`，元数据却无处安放——没有这个块，导入侧就还原不出这些行，最直接的后果是**搬家之后 `site.server_logo` 变成死链**。生产者**必须**把这类行写入 `files`；消费者**必须**为其建行并落字节，但**禁止**建立任何 Echo 关联。归属跟随执行导入的 owner（胶囊不携带 `user_id`）。

## 4. Echo 文件

### 4.1 路径与命名

- 位置：`echoes/<YYYY>/<YYYY-MM-DD>-<id 去横线后的末 8 位>.md`，`<YYYY>`/`<YYYY-MM-DD>` 取自 `created_at`。取**末** 8 位是因为 Echo 的 id 是 UUIDv7，前 48 位为时间戳——同一批创建的条目前缀高度重合（真实实例上出现过 287 条中 270 条共用同一前 8 位、单日挤进 5 条），只能靠 `-2`/`-3` 后缀区分，等于没有辨识度；末 8 位落在随机段，天然离散。同名时**必须**在 `.md` 前追加 `-2`、`-3`… 去重。
- 命名仅为浏览友好：消费者**必须**以 frontmatter 为准，**禁止**从文件名解析任何语义。

### 4.2 Frontmatter 字段

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | string(UUID) | **必须** | 幂等键 + permalink（`/echo/:id`）。缺失时 `check --fix` 生成 UUIDv7 回写 |
| `created_at` | RFC3339 | **必须** | 语义为时刻（instant）；任意合法偏移均可，导出时**必须**统一 UTC（`Z` 后缀） |
| `username` | string | 可选 | 缺省取 `owner.username` |
| `tags` | string[] | 可选 | 标签名数组；消费者按名称 find-or-create |
| `layout` | enum | 可选，默认 `waterfall` | `waterfall\|grid\|horizontal\|carousel\|stack\|none` |
| `private` | bool | 可选，默认 `false` | 私密标记；export/import 默认排除 `private: true` 条目，`--include-private` 显式包含 |
| `fav_count` | int ≥ 0 | 可选，默认 `0` | 点赞数，随往返保留 |
| `files` | list | 可选 | 媒体引用，见下；**数组顺序即展示顺序**（对应 `EchoFile.SortOrder`） |
| `extension` | object | 可选 | `{type, payload}`，见下 |

`files[]` 元素——**字段名与取值对齐 `File` 模型列，import 1:1 入库、零转换**。`key` 与 `url` **必须二选一**：

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | string(UUID) | 可选 | `File.ID` 原样；缺省 import 生成。幂等锚点：目标库同 `id` 已存在 → 直接复用行 |
| `key` | string | 与 `url` 互斥 | `File.Key` 原样（托管文件的扁平存储键，**禁止**含 `/` 或 `..`）。字节**必须**位于 `files/ + Resolve(key)`（§6 路由表）——位置由 `key` 纯函数派生，胶囊不存路径 |
| `url` | string(URL) | 与 `key` 互斥 | 外链文件（`storage_type=external`）的 `File.URL` 原样透传，**不**本地化 |
| `category` | enum | 可选 | `File.Category`：`image\|video\|audio\|pdf\|markdown\|file`（`storage.Category` 全集，与 `NormalizeCategory` 一致）；缺省按扩展名派生，`url` 条目**应当**显式给出 |
| `name` | string | 可选 | `File.Name` 原始文件名（keygen 后的 key 不可读，此字段保留人类信息） |
| `content_type` | string | 可选 | `File.ContentType`；缺省按扩展名推导，兜底类别与无扩展名文件应当显式给出 |
| `size` | int ≥ 0 | 可选 | `File.Size`；提供时 `check` **应当**核对实际字节数（完整性校验红利） |
| `width` / `height` | int ≥ 0 | 可选 | `File.Width/Height`；瀑布流渲染防抖动，缺省可由消费者重算 |

**明确不入胶囊**（全部为运行时拓扑或派生数据）：`storage_type/provider/bucket`（由目标实例配置决定；external 由 `url` 在场表达）、托管文件的 `File.URL`（`AfterFind` 按当前配置重算）、`user_id`（归属跟随 Echo）、`File.CreatedAt`（`autoCreateTime` 行元数据）、`EchoFile.ID/SortOrder`（数组顺序表达）。

`extension`：

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `type` | enum | **必须**（若存在 extension） | `MUSIC\|VIDEO\|GITHUBPROJ\|WEBSITE\|LOCATION\|TWEET`（对齐 `model.Extension_*`） |
| `payload` | object | **必须**（若存在 extension） | 原样映射 `EchoExtension.Payload`，结构随 `type` 而异，本规格不逐一约束。内嵌实例相关 URL 时警告（§7，Q15） |

### 4.3 正文

- frontmatter 结束后即正文，**必须**与 `Echo.Content` 逐字一致（markdown，不做任何转换/转义）。
- 空正文合法（纯图片/纯扩展 Echo）。

## 5. 评论快照 `comments.yaml`

字段对齐 `PublicComment` 投影（`internal/model/comment/comment.go`）：

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `schema_version` | int | **必须** | 同 §3 |
| `comments[].id` | string(UUID) | **必须** | |
| `comments[].echo_id` | string(UUID) | **必须** | 须指向胶囊内存在的 Echo，孤儿仅警告 |
| `comments[].parent_id` | string(UUID) \| null | 可选 | 两级盖楼：null = 顶层 |
| `comments[].nickname` | string | **必须** | |
| `comments[].website` | string(URL) | 可选 | |
| `comments[].content` | string | **必须** | |
| `comments[].status` | string | 可选，默认 `approved` | 胶囊内**应当**只含 `approved` |
| `comments[].source` | string | 可选 | 评论来源（对齐 `SourceType`） |
| `comments[].created_at` | RFC3339 | **必须** | |

- **禁止**字段：`email`、`ip_hash`、`user_agent`、`user_id`（隐私投影，出现即校验错误）。
- 评论**必须**独立于 Echo 文件存放（单一 `comments.yaml`）：评论是第三方数据且变更生命周期与内容不同，混入 frontmatter 会污染内容文件身份并制造 diff 噪音。

## 6. 媒体目录 `files/`

- 内部结构 = 本地存储 `DataRoot`（默认 `data/files`）的**原样 mirror**：相对路径 = `files/ + schema.Resolve(key)`（`internal/storage/schema.go`）。
- 子目录固定五类：`images/ audios/ videos/ documents/ files/`（对应 `RouteByExt` 四类 + `DefaultRoute("files/")` 兜底；`files/files/` 即兜底类别的 mirror，属预期布局）。
- 生产者**必须**按 `files` 表记录驱动写入（`Resolve(key)` 落位），**禁止**盲目拷贝 `DataRoot`：`data/files/snapshots/` 等非托管产物不得进入胶囊。「目录拷贝」仅是近似心智模型。
- 胶囊**必须**自包含：所有 `files[].key` 对应 `files/ + Resolve(key)` 的字节都在胶囊内；S3 托管文件由生产者下载入胶囊。外链（`url`）文件除外。
- 既不被任何 Echo 的 `files[]`、也不被清单 `files` 块声明的字节：合法，`check` **应当**告警（悬空文件）。真实导出不会产生这种情况——未挂 Echo 的文件行都会进清单 `files` 块；它只在手写胶囊里出现。

## 7. 校验规则（`ech0 check`）

| 级别 | 条件 |
|---|---|
| **错误**（拒绝 import/build） | `ech0.yaml` 缺失或 `schema_version` 不识别；`id`/`created_at` 缺失或非法；有 `extension` 但 `type` 或 `payload` 缺失；`files[].key` 含 `/` 或 `..`，或字节不存在于 `files/ + Resolve(key)`；`key`+`url` 同时存在或同时缺失；`comments.yaml` 出现禁止字段；`id` 重复 |
| **警告** | **`layout`/`extension.type`/`files[].category` 取值不在已知枚举内**（见下）；孤儿评论；悬空媒体文件；未知字段/未知顶层路径；`custom_js`/`custom_css` 非空；`status != approved` 的评论；`files[].size` 与实际字节数不符；正文、`extension.payload` 或 `site.server_logo` 内嵌实例相关 URL（`site.server_url` 前缀或 `/api/files/` 引用，迁移后可能断链） |

**表现层枚举只警告、不阻断**：`layout`、`files[].category`、`extension.type` 都只影响「怎么渲染」，内容本身完好。消费者**必须**优雅降级——`layout` 回落 `waterfall`、`category` 回落 `file`、不认得的 `extension.type` 跳过渲染——而**禁止**因此拒绝整个胶囊。理由有二：其一，活实例的写路径本就如此（`service/echo` 把未知 `layout` 归一成 `waterfall`），规范没有理由比它描述的系统更严格；其二，这与 §8「消费者必须忽略未知字段」是同一类前向兼容问题——未知的枚举**取值**和未知的**字段**都可能来自更新的版本或第三方生产者。缺失 `extension.type` 仍是硬错：`payload` 的结构随 `type` 而异，没有它就无从解释。

> 降级只发生在**消费侧的渲染**（`build`）。`export`/`import` 一律逐字保留原值——库里是 `stream` 就导出 `stream`、导入回去还是 `stream`（§11 的 1:1 纪律）。

- `--fix` 可自动修复项：缺失 `id`（生成 UUIDv7 回写 frontmatter）。仅此一项，后续扩展须逐项列入本规格。
- `ech0 import capsule` / `ech0 build` 隐式执行同一套校验。

## 8. 版本与兼容

- `schema_version` 为整数，当前 `1`。字段语义变更或删除 → 递增；**追加**字段/枚举值 → 不递增。
- 消费者**必须**忽略未知字段与未知路径（警告级）；**禁止**因未知内容拒绝处理。
- 消费者遇到高于自身支持的 `schema_version` **必须**拒绝并明确报错。

## 9. CLI 命令语法

动词在前、格式为子命令（设计依据见 design §5.0）。裸 `ech0 export` / `ech0 import` 打印 help，无默认格式。

```bash
ech0 export capsule   [-o ./capsule] [--include-private] [--zip]
ech0 export snapshot  [-o ./snapshot.zip]                          # P4，语法保留
ech0 import capsule   [<path>=./capsule] [--dry-run]
ech0 import snapshot  <snapshot.zip> --yes                         # P4，语法保留；破坏性整库替换
ech0 check            [<path>=./capsule] [--fix]
ech0 build            [<path>=./capsule] [-o ./dist] [--base-url /]
```

| flag | 命令 | 语义 |
|---|---|---|
| `-o, --output` | export/build | 输出目录或文件 |
| `--include-private` | export capsule / import capsule | 包含 `private: true` 条目（双端默认排除） |
| `--zip` | export capsule | 目录打包为单文件 `.zip`（zip 内布局与目录形式一致） |
| `--dry-run` | import capsule | 只输出创建/跳过清单，不写库 |
| `--fix` | check | 回写可自动修复项（§7） |
| `--base-url` | build | 站点部署根路径（子路径部署用） |
| `--yes` | import snapshot | 破坏性操作确认门，缺失即拒绝 |

退出码：`0` 成功；`1` 校验错误或执行失败；仅警告不影响退出码。

## 9.1 HTTP 接口（Web 面板）

`export capsule` 与 `import capsule` 另有一条等价的 HTTP 路径，供面板使用；两条路复用同一批
`internal/capsule/*` 入口，语义完全一致。`check` 与 `build` **只有** CLI。全部端点沿用既有导出
/迁移域的 `admin:settings` 权限。

| 端点 | 胶囊相关增量 |
|---|---|
| `POST /migration/export` | 请求体 `{format?: "snapshot"\|"capsule", include_private?: bool}`；两者皆可省，省略即「快照、不含私密」 |
| `GET /migration/export/status` | 响应新增 `format`，标明当前产物是哪种格式 |
| `GET /migration/export/download` | 新增 `?format=` 查询参数，缺省 `snapshot` |
| `POST /migration/upload` | `source_type` 新增取值 `capsule` |
| `POST /migration/start` | `source_type: "capsule"` 时，`source_payload.include_private` 控制是否导入私密条目 |

- 未知 `format` **必须**拒绝，**禁止**静默回落到快照——悄悄给出另一种产物会让用户拿错东西。
- 胶囊产物落在 `data/files/capsules/`，与快照的 `data/files/snapshots/` 分居两个槽位，各自只保留最新
  一份。两个目录都**必须**排除在快照打包之外（`internal/migrator/artifact` 收口）。
- Web 导入与 CLI 一样把校验作为硬前置：有错误级发现即拒绝落库，错误摘要回填进作业的 `error_message`。
- 导出胶囊**禁止**发布 `SystemSnapshot` 事件：webhook 订阅它来确认「备份已完成」，而胶囊不是备份。

## 10. `ech0 build` 产物约定

```text
<dist>/
  index.html          # 注入 window.__ECH0_STATIC__ = true
  assets/…            # 内嵌 SPA 产物（template/dist/）原样拷贝
  dataset.json        # 烘焙数据：echoes + tags + comments + site
  api/files/…         # 胶囊 files/ 原样拷贝（与 serve 模式静态路由同形，URL 零改写）
  rss.xml             # 预生成 Atom feed
  sitemap.xml
  404.html            # SPA fallback（Pages 类托管深链支持）
  api/connect         # Connect 载荷快照（见下）
```

- `api/connect`：**必须**产出，内容与活实例 `GET /api/connect` 响应体**同形**（`Result` 信封 + `Connect` 载荷：`server_name/server_url/logo/total_echos/today_echos/sys_username/version`），统计值为构建时冻结快照——远端实例的既有探测路径无需改动即可消费。注意：无扩展名文件在部分静态托管上 `Content-Type` 不可控，消费端应按 body 解析 JSON。

## 11. 内容映射与往返契约

**「导出即转储，构建即转换」**：`export`/`import` 是 DB 的 1:1 序列化边界——字段名与字段值一律原样，一切消费导向的转换（URL 计算、dataset 烘焙、统计冻结）都在 `build` 内完成。仅有的表示层差异（非数据转换、均为无损双射）：① 时间 `int64` Unix 秒 ↔ RFC3339 UTC；② 行 ↔ frontmatter-markdown 文件形态；③ 关系实体以内容字段表示（`Tags` → 名称数组、`EchoFile.SortOrder` → 数组顺序）。

DB ↔ 胶囊字段映射（权威表见 design §4.3）；往返保真度：

| 数据 | 往返 |
|---|---|
| Echo 全量（正文/tags/layout/extension/private/fav_count/created_at/id）、托管媒体字节与文件行（含未挂 Echo 的）、site 公开子集、connects | ✅ 完整 |
| 评论 | ⚠️ 有损（Public 投影，无 email/ip_hash/user_id） |
| 外链文件 | ⚠️ URL 透传，字节不随胶囊 |
| 账号/凭据/运维配置/embeddings/访客统计/日志 | ❌ 不往返 |

> **一处收敛性偏差**：胶囊省略 `files[].size` 时，import 按实际字节数补齐（源库把 size 存成 `0` 的历史行会因此被修正）。这是元数据修复而非内容改写，且一轮往返后即收敛——`export → import → export` 的第二份胶囊会比第一份多出这些 `size`，之后再无差异。

### 11.1 `files[]` 落地语义（import）

- **`key` 条目**：字段 1:1 写入 `File` 列，无任何派生。幂等与去重：
  - `id` 在场且目标库存在同 `id` 行 → 直接复用，不重写字节；
  - 否则按（当前存储后端, `key`）查：不存在 → 写字节（`VireFS Put(key)`，原生写路径）、建行；存在且内容一致（size + hash）→ 复用；存在但内容不同（手写胶囊撞名）→ keygen 生成新 key 落盘，报告中列出改名。
- **`url` 条目** → 建 `storage_type=external` 行，URL 即权威表示（`AfterFind` 对 external 不重算）。
- 三种存储形态的归一化：local 与 object 统一为 `key`（字节入胶囊，import 按目标后端落地、URL 按目标配置重算）；external 统一为 `url` 透传。显示正确性由两条不变量保证：build 复用 serve 模式的 `/api/files/` URL 形状（`/api/files/ + Resolve(key)`）；import 不携带源 URL。

### 11.2 export 纪律

- 托管文件字节无法取回（S3 桶不可达、对象丢失、本地文件缺失）→ **必须**报错并列出清单，**禁止**静默产出不自包含的胶囊。

### 11.3 import 落地语义（其余）

- **幂等按 `id`**：目标库已存在同 `id` 的 Echo → 跳过并报告，不覆盖不合并；无 `--overwrite`（重复执行安全）。
- **原样导入原则**：胶囊字段值 1:1 写入对应 DB 列，**禁止**数值转换——`username` 逐字入库不改写、`fav_count`/`status`/正文原样。唯一例外是补全胶囊不携带的**内部必填外键** `Echo.UserID`：同名用户存在则挂接，否则挂到执行导入的 owner（权限归属）；展示归属始终以原样保留的 `username` 为准。
- **站点设置**：`site` 子集仅填「未配置项」，**禁止**覆盖已配置项。「未配置」= 当前值为空串**或**逐字等于 `setting.System.Default()`（config 派生）的对应值——不能只判空串：全新实例的 `site_title`/`server_logo`/`server_url` 等在 KV 缺键时本就返回非空默认值，只判空串会让站点身份在「搬站到新实例」这个头号用例里永远导不进去。
- **评论归属**：以「落库后目标 `echo_id` 在 `echos` 表中是否存在」为准——本轮新建的与此前已导入的同等对待。往既有 Echo 追加评论不构成对该 Echo 的修改，故不受「幂等跳过」牵连；幂等由评论自身 `id` 保证。宿主不存在（含因 `private` 被排除）→ 记为孤儿并跳过。
- **不发布事件**：导入不触发事件总线（webhook/embedding/agent 订阅者不响应）；报告末尾提示可在后台触发索引重建覆盖导入内容。

## 12. 待定项索引

**无待定项。** Q1–Q15 已全部裁决，完整决策记录见 design §10；本规格所有条目均为已确认共识。
