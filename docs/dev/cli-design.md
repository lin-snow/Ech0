# Ech0 CLI 设计

日期：2026-10-30  
状态：草案；M1/M2 已落地（目录骨架、catalog 驱动的命令发现、token 登录）  
范围：`apps/cli`（Bun + TypeScript 客户端）与配套后端改动（operation catalog、device authorization）

## 1. 定位与范围

一句话：**登录到一台指定的 Ech0 实例，在终端完成 Echo 的增删改查**。

范围的取舍：登录之后只暴露 Echo CRUD（`echo.create` / `echo.list` / `echo.get` / `echo.update` / `echo.delete`）与文件面（上传 `file.upload`、外链登记 `file.external`、浏览 `file.list` / `file.get` / `file.tree`、下载 `file.stream.by_id`），其余接口（评论、迁移、系统日志、Copilot 等）暂不进命令树。`apps/cli/src/catalog/scope.ts` 的 `CLI_OPERATIONS` 是唯一开关，加一个 id 即可放开一项能力，命令生成无需改动；`COMMAND_NAMES` 只用来把读起来别扭的 catalog id 映射成更像动词的命令名（`file.stream.by_id` → `ech0 file download`）。

关键决策：**CLI 不内置业务命令**。命令来自服务端下发的 operation catalog，这样服务端加接口不需要用户单独升级 CLI；客户端的 `scope` 白名单决定当下暴露哪些。

v1 目标用户：实例 owner（admin）。非目标：

- 不做 TUI（`internal/cli` 的 Go TUI 与 `cmd/` 保持现状）。
- 不做离线操作 `data/` 或 SQLite（那是 Go 端 `capsule`/`snapshot` 的职责）。
- 不做完整管理台前端替代方案；面板操作以 catalog 暴露的接口为准。

## 2. 技术栈（已定）

| 项 | 选择 | 依据 |
| --- | --- | --- |
| 运行时 | Bun ≥ 1.4 | `bun build --compile` 单文件二进制；`Bun.secrets` 取系统钥匙串；内置 test |
| 语言 | TypeScript，包内钉 **5.9** | `@hey-api/openapi-ts` 0.99 在 TS 7 原生 tsc 上崩溃（`AnyKeyword` undefined）；本包版本与根仓其它前端隔离 |
| 命令框架 | `commander` 15 | 动态子命令树、`--help`、必填选项校验 |
| 校验 | `zod` v4 | catalog 文档校验（v4 API，如 `z.record(k, v)`） |
| 类型/schema 生成 | `@hey-api/openapi-ts` 0.99（`@hey-api/typescript` + `zod`） | 单工具同时产出 `types.gen.ts` 与 `zod.gen.ts`；`orval` 8.45 因 Huma 的 `ResultInterface {}` schema 名失败 |
| 测试 | `bun test` | 28 用例：envelope、catalog 缓存、动态命令、412 重试、退出码 |
| 包名 | `@ech0/cli`，bin `ech0` | npm `ech0` 已被空占位包占用 |

## 3. 目录结构

```
apps/cli/
  package.json            # @ech0/cli, bin: ech0, type: module
  tsconfig.json           # ES2024 / bundler / strict / noEmit
  openapi-ts.config.ts    # 生成配置（input = ../../internal/openapi/openapi.yaml）
  justfile                # 根 justfile 通过 `mod cli "apps/cli/justfile"` 暴露
  scripts/
    dist.ts               # 发布产物：交叉编译二进制、tar.gz、SHA256SUMS、npm 包目录
  npm/
    ech0.cjs              # npm 启动器：按平台 exec optionalDependency 里的二进制
    README.md             # npm 包页面说明
  src/
    index.ts              # #!/usr/bin/env bun → createRootCommand → parseAsync
    session.ts            # profile/server/token 解析、全局 flags
    catalog/
      types.ts            # catalog zod schema + parseCatalog
      store.ts            # 缓存路径、指纹、读写
      load.ts             # cache-first 拉取与刷新
    api/
      generated/          # 生成：types.gen.ts / zod.gen.ts / index.ts（入库，gen 会整目录重建）
      http.ts             # 手写：fetch、鉴权/语言/指纹头、envelope 解包、Ech0Error
    commands/
      root.ts             # 本地命令 + 动态命令装配
      auth.ts             # login/status/logout
      config.ts           # profile 管理
      catalog.ts          # refresh/path
      dynamic.ts          # catalog op → commander 命令
      operation.ts        # 构建请求、412 重试、结果渲染
      version.ts          # 本地版本
    config/               # XDG 路径、profile 存储、凭据（keychain/文件）
    output/               # 控制台、表格
    ui/                   # stdin
  test/
```

## 4. Operation catalog：命令发现与版本解耦

### 4.1 动机

CLI 若把 `POST /echo` 之类的接口写死成 `ech0 echo add`，服务端加字段/加接口都要重新发 CLI。改为：

- 服务端发布**全部可调用操作**的清单（id、方法、路径、输入 JSON Schema、字段落点、作用域、效果）。
- CLI 把清单里的每个操作变成一条命令，并把输入 Schema 变成命令行 flags。
- 清单带指纹；CLI 缓存，服务端发现客户端缓存过期就让它刷新。

### 4.2 服务端契约

`GET /api/_catalog`（**public**，不需要 token），返回 canonical JSON（Go `encoding/json` 排序键、紧凑编码）：

```json
{
  "ops": {
    "echo.create": {
      "summary": "创建新的 Echo",
      "method": "POST",
      "path": "/echo",
      "scopes": ["echo:write"],
      "effect": "write",
      "kind": "json",
      "input": {
        "type": "object",
        "properties": { "content": { "type": "string" }, "private": { "type": "boolean" } },
        "required": ["content"]
      },
      "bind": { "content": "body", "private": "body" },
      "deprecated": false
    }
  }
}
```

- 清单由 Huma 文档派生（`internal/openapi/catalog`），再合并 `raw.go` 里手写的裸 gin 操作：`operationId` 的 `-` 换成 `.`（`echo-list` → `echo.list`）；`input` 把 path/query/body/form 四类参数拍平成一个对象，`$ref` 递归内联；`bind` 标记每个字段落在 `path`/`query`/`body`/`form`；`effect` 由方法推导（GET/HEAD=read、DELETE=destructive、其余=write）；`kind` 由成功响应内容类型推导（json/stream/binary/websocket）。
- 非 JSON 传输额外带 `transport`（见 4.5）；没有该字段即「JSON 进、JSON 信封出」。
- 指纹 = sha256(canonical bytes)。**每个 `/api` 响应**都带 `X-Ech0-Catalog: <指纹>`。
- 请求可带同一个头；带了但和当前指纹不同 → **412**，body `{"code":0,"msg":"catalog stale","error_code":"catalog_stale"}`，在 handler 之前短路。不带头的旧客户端（web、MCP、脚本）完全不受影响。
- 目录端点自身也带指纹头与 `ETag`。

### 4.3 客户端缓存

- 路径：`<config dir>/catalog/<host>.json`（如 `~/.config/ech0/catalog/memo.sn0w.fyi.json`），`0600`。
- 指纹在本地重算 sha256，不信任响应头（缓存文件才是身份）。
- 启动 cache-first：有缓存不发网络请求；无缓存才拉取。`ech0 catalog refresh` 强制重拉，`ech0 catalog path` 打印路径。
- 调用返回 412 或 404 时**刷新一次并重试**（404 可能意味着路由被移除或迁移）。
- 命令树在解析 argv 之前构建，实例取自 profile / `ECH0_SERVER`；命令行的 `--server` 只影响本地命令，不会重建命令树。首次 `auth login` 前没有动态命令（登录后下一次调用即可见）。

### 4.4 命令生成

- 操作 id 按 `.` 分段：`echo.create` → `ech0 echo create`，`system.logs.stream` → `ech0 system logs stream`。`COMMAND_NAMES` 可覆盖命令路径而不改 catalog id。
- **可调用的 id 不能再当命名空间**：commander 会先解析父命令的 flag，子命令的同名 flag 会被父命令吞掉（实测 `file.stream` + `file.stream.by_id` 都对 `--output` 生效时，子命令的值为空）。因此当某个 id 的某个前缀本身就是操作时，子命令上移一层并合并剩余段：`comment.create.integration` → `ech0 comment create-integration`，`migration.export.download` → `ech0 migration export-download`。纯分组的命名空间（`system.logs`）照常嵌套。`file.stream.by_id` 本来会变成 `ech0 file stream-by-id`，现在被 `COMMAND_NAMES` 改名为 `ech0 file download`。
- 输入字段 → flags：snake_case/camelCase 统一转 kebab（`pageSize` → `--page-size`）；boolean 无值；整数/数字/逗号数组/JSON 对象分别解析；`format: binary` 的字段是 `--file <path>`（`-` 读 stdin）。
- 数组字段两种写法：`--tags a,b`（逗号）或 `--tags '[{"name":"a"}]'`（以 `[` 开头按 JSON 解析）。
- 与 CLI 保留 flag 重名（`json/profile/server/token/lang/color/help/version`）的字段改名为 `--input-<name>`。
- 必填字段在 action 里校验（缺失 → 退出码 2 + `缺少必填参数 --x`），不用 `requiredOption`：可调用 id 一旦有子命令，父命令的必填项会连子命令一起拦下。
- 命令树只注册 `CLI_OPERATIONS` 里的 id（当前为 Echo CRUD 五条 + 文件面六条），其余操作即使出现在 catalog 里也不会生成命令；`ech0 help <关键词>` 里也搜不到。
- help 末尾附 `METHOD /path · scopes: ...`，便于确认权限与效果。

### 4.5 传输层：流式、上传与下载（后端 + CLI 已实现）

裸 gin 的 7 个操作由 `internal/openapi/catalog/raw.go` 手写，带完整 `transport`：

| id | 路由 | 传输 |
| --- | --- | --- |
| `system.logs.stream` | `GET /api/system/logs/stream` | SSE，无 `event:`，`data` 是信封 |
| `ws.system.logs` | `GET /ws/system/logs?token=` | WebSocket，token 走查询串 |
| `file.upload` | `POST /api/files/upload` | multipart：`file`(binary) + `category` + `storage_type` |
| `migration.upload` | `POST /api/migration/upload` | multipart：`file`(binary) + `source_type` |
| `file.stream` | `GET /api/file/stream?path=&storage_type=` | binary |
| `file.stream.by_id` | `GET /api/file/{id}/stream` | binary |
| `migration.export.download` | `GET /api/migration/export/download` | binary（zip） |

> CLI 开放了上表的 `file.upload` 与 `file.stream.by_id`（命令名 `ech0 file download`），以及 Huma 的 `file.external` / `file.list` / `file.get` / `file.tree`。SSE 与 WebSocket 的客户端实现与单测都在，但未开放命令，把对应 id 加进白名单即可。Copilot（`POST /api/chat`）按产品决定暂不支持，也不在 catalog 里。

`transport` 形状（Huma 操作仅在非纯 JSON 时才带）：

```json
{
  "request":  { "content_type": "multipart/form-data" },
  "response": {
    "kind": "stream", "content_type": "text/event-stream",
    "events": [{ "name": "delta", "data": { "text": "string" } }],
    "filename": "ech0-snapshot.zip"
  }
}
```

- `request.auth = "query"`：凭证必须放 `?token=`（WebSocket 握手带不了 Authorization）。
- `response.events` 只是提示词表：CLI 只对认识的名称做定制渲染，未知事件降级为 `[name] {json}`，服务端加事件不必发 CLI。
- **防漂移**：`internal/router/catalog_routes_test.go` 用 `engine.Routes()` 对全量路由做双向断言——每条 catalog 操作必须有真实路由，每条 `/api`、`/ws` 路由（除显式排除的 docs/openapi/catalog/静态文件/MCP/captcha/auth 全家桶）必须被 catalog 描述。手写注册表因此不会悄悄失效。

CLI 按 `kind` 分派：

| kind | 默认（TTY） | `--json` | 落盘 |
| --- | --- | --- | --- |
| `json` | 格式化 JSON | 原始响应体 | — |
| `stream` | `delta` 追加文本、`reasoning` 暗色、`searching` 走 stderr、`sources/coverage` 收尾汇总、`ask` 打印问题与选项、`error` 红字 + 退出码 1 | 每帧一行 NDJSON `{"event","data"}` | — |
| `binary` | 非 TTY 直接写 stdout；TTY 从 `Content-Disposition` 取名落盘 | 需配合 `--output`，输出元信息 JSON | `--output <path>`（temp + rename 原子写），TTY 且已知 `Content-Length` 时 stderr 进度 |
| `websocket` | 每帧打印 `data` | 每帧一行原始 JSON | — |

- SSE 解析：按空行切帧、跨 chunk 累积、多行 `data` 用 `
` 连接、`
` 归一、`: keep-alive` 注释丢弃、无 `event:` 默认 `message`。
- 上传：`Bun.file(path)` 直接进 `FormData`（不进内存），不手设 `Content-Type`；`-` 读 stdin。
- 中断：SIGINT → `AbortController.abort()`，退出码 130；下载中断删除 `.part`。
- 流式请求同样享受 412 刷新重试与 401 续签重试（在响应头返回、body 未消费前判断）。

### 4.6 组合 Echo：先上传，再引用

Echo 可以由「文本 + 附件 + 标签 + extension」组成。HTTP 侧附件本来就是独立的 `POST /files/upload`，CLI 保持同样的两步，不把上传藏进 `echo create`：

```bash
# 1) 上传，拿 file_id（category 省略时按扩展名推断）
ech0 file upload --file ./cover.png
# {"code":1,"data":{"id":"01a18c…","name":"cover.png","storage_type":"local","url":"/api/files/images/…"}}

# 2) 引用 file_id 发 Echo
ech0 echo create \
  --content "今天听了这张" \
  --tag music \
  --file-id 01a18c… \
  --extension '{"type":"MUSIC","payload":{"url":"https://music.example/1"}}'
```

| 命令 / flag | 行为 |
| --- | --- |
| `ech0 file upload --file <path>` | multipart 上传；`-` 读 stdin。`--category` 省略时按扩展名推断（image/video/audio/pdf/markdown/file） |
| `--file-name <name>` | 覆盖 multipart 文件名；stdin 上传必须带扩展名（服务端按文件名判定类型） |
| `--storage-type <local\|s3>` | 上传目标；默认 `local`。`external` 不是上传，传了直接报错并提示改用 `file external` |
| `ech0 file external --url <url>` | 外链文件：只登记 URL（服务端校验 http/https、按 URL 推断 `content_type` 与 category），返回 `storage_type=external` 的 file id |
| `ech0 echo create --file-id <id>` | 可重复；按顺序写入 `echo_files:[{file_id,sort_order}]` |
| `--tag <name>` | 可重复；追加到 `tags`（与 `--tags` 合并） |
| `--from <file\|->` / `--edit` | 用 JSON payload 提交；`--edit` 打开 `$EDITOR`，不保留任何状态 |

合并优先级：payload 文件打底 → 同名 flag 覆盖 → `--tags`/`--tag` 追加、`--file-id`/`--echo-files` 追加到 `echo_files` 末尾（`sort_order` 续号）。附件类别必须一致（服务端 `validateSingleFileCategory`），上传时应选同一个 `--category`。

外链媒体走登记而不是上传，之后同样用 `--file-id` 挂到 Echo：

```bash
ech0 file external --url https://example.com/photo.png --category image
ech0 echo create --content "外链图片" --file-id <返回的 id>
```

payload 文件就是 JSON 形态的 `EchoUpsertDto`：

```json
{
  "content": "从文件提交",
  "tags": [{ "name": "cli" }],
  "echo_files": [{ "file_id": "<已上传的 id>", "sort_order": 0 }],
  "private": false,
  "layout": "waterfall",
  "extension": { "type": "WEBSITE", "payload": { "title": "Ech0", "site": "https://example.com" } }
}
```

**为什么不做 stateful 的 draft/workspace**：上传出来的文件在 Echo 创建前是 temp 记录（TTL 到期清理），草稿目录会引入格式版本、profile 绑定、孤儿清理和部分失败语义，而且不可脚本化。JSON 文件本身就是草稿：可以进 git、可以由 agent 生成、可以 `--edit` 手改。注意草稿里存 **file_id** 会过期，长周期草稿应在提交前重新上传附件。

### 4.7 文件浏览、元信息与下载

文件面分两条写入路径和三条读取路径，命令名与 server operation 一一对应：

| 命令 | operation | 说明 |
| --- | --- | --- |
| `ech0 file upload --file <path>` | `file.upload` | multipart 写入存储；`-` 读 stdin |
| `ech0 file external --url <url>` | `file.external` | 只登记外链，不上传字节 |
| `ech0 file list [--page --page-size --search --storage-type]` | `file.list` | 分页列出文件 |
| `ech0 file get --id <file_id>` | `file.get` | 文件元信息（name / url / storage_type / category…） |
| `ech0 file tree --storage-type <local\|s3\|object>` | `file.tree` | 按存储前缀浏览目录树 |
| `ech0 file download --id <file_id> [--output <path>]` | `file.stream.by_id` | 下载内容；非 TTY 默认写 stdout，`--output` 原子落盘，省略时按响应文件名落盘 |

文件名的来源是服务端 `Content-Disposition`：`StreamFileByID` / `StreamFileByPath` 在流式响应上带 `inline; filename=…`（非 ASCII 名走 RFC 5987 的 `filename*=utf-8''…`），CLI 优先取 `filename*`、其次 `filename`，再做 `basename` 收口防止 `../` 逃出当前目录；两者都没有时落到 `download`（API 路径末段是动词 `stream`，不能当文件名）。

`file download` 的 catalog id 仍是 `file.stream.by_id`（`COMMAND_NAMES` 只改命令路径），所以 help、scope 检查和请求指纹不变。

**两者在数据库里是同一张表**：上传和外链都会写一行 `File`（`files` 表）。external 行只是 `storage_type` / `provider` 为 `external`、`bucket` 为空、`key` 为 `external/<category>/<sha256(normalized_url)>`、`size=0`、`url` 保存原始链接；上传行记录 local/s3/object 的路由与真实 key。因此 `file list`（不传 storage_type 时）、`file get`、`--file-id` 对两者一视同仁：

- `file download` 对 external 返回 307 重定向到原 URL，对 local/s3/object 才真的读字节；跟随重定向后由远端决定文件名，远端不给 `Content-Disposition` 时 CLI 落到 `download`。
- `file tree` 只列 local/object 的物理存储，external 没有物理 key。
- 两者都会建 `TempFile` 记录（未挂到 Echo 前有 TTL）；temp 清理对 external 只删记录，不删远端文件。
- external 的 `key` 由 URL + category 哈希得到，同 URL 同 category 重复登记会返回已有记录（unique route 索引）。

**为什么 `upload` 和 `external` 不合并成一个 `create`**：它们是两个 HTTP 操作，传输与字段都不同——`file.upload` 是 `multipart/form-data`（`file` 二进制 + category + storage_type），`file.external` 是 JSON（`url` + name/content_type/width/height）。合并只能在 CLI 里手写一个分派命令（有 `--file` 调上传、有 `--url` 调外链），这会打破「命令树完全由 catalog 生成」的约束：字段变化不再自动跟进，help/示例/必填校验都要另写一份，而且两条操作的 scope 与错误语义会被藏进同一个命令。保留两个动词更直观，也保留 help 里的传输卡片。

## 5. 认证与凭据

### 5.1 三种登录方式

| 方式 | 命令 | 状态 |
| --- | --- | --- |
| 粘贴 access token | `ech0 auth login --server URL --token -` | **已实现**（token 经 `/api/user` 校验后入 keychain/文件） |
| 用户名 + 密码 | `ech0 auth login --server URL -u <user>` | **已实现**（提示符隐藏输入；CI 用 `--password-stdin`） |
| device authorization | `ech0 auth login --server URL --device` | **已实现**（CLI 轮询 + Web `/device` 批准页） |

### 5.2 密码登录（M3）

`POST /api/login` 只返回 `access_token` + `expires_in`；refresh token 只走 httpOnly cookie。因此：

- **admin**：用登录 JWT 调 `POST /api/access-tokens`，铸 `audience: "cli"`、`expiry: "1_month"` 的长期令牌保存。
- **非 admin**：`/access-tokens` 服务层强制 admin，只能保存 JWT + refresh cookie，过期时 `/api/auth/refresh` 续期。
- `audience` 取 `authModel.AudienceCLI`（已存在）。

> 已修复的既有缺陷：`CreateAccessToken` 曾把 `time.Duration`（纳秒）当作秒传给 `CreateAccessClaimsWithExpiry`，导致铸出的令牌 `exp` 为负、一出生即失效（e2e 实测 token 解析失败）。现在统一走 `accessTokenExpirySeconds`（8 小时=28800，1 月=2592000，never=0）。

### 5.3 Device authorization（M4，后端改造）

新增 4 个端点（raw gin，与 login 同族，不进 Huma spec）：

| 方法 | 路径 | 认证 | 请求 | 响应 |
| --- | --- | --- | --- | --- |
| POST | `/api/auth/device/code` | public | `{client_name, scopes[]}` | `{device_code, user_code, verification_uri, verification_uri_complete, expires_in: 600, interval: 5}` |
| POST | `/api/auth/device/token` | public | `{device_code}` | 成功 `{access_token, token_type:"Bearer", expires_in, scopes[]}`；未批准/过期/拒绝时 HTTP 400 + 稳定 `error_code` |
| GET | `/api/auth/device/pending?user_code=` | 已登录会话 | — | `{client_name, scopes[], expires_in}` |
| POST | `/api/auth/device/approve` | 已登录会话 | `{user_code}` | `200 {code:1}` |

错误码（`error_code`，保留 `msg` 作为人类可读回退）：

| error_code | 含义 |
| --- | --- |
| `DEVICE_AUTHORIZATION_PENDING` | 尚未在浏览器批准，CLI 继续轮询 |
| `DEVICE_SLOW_DOWN` | 轮询快于 `interval`，CLI 退避 5s |
| `DEVICE_EXPIRED` | 设备码过期（10 分钟） |
| `DEVICE_ACCESS_DENIED` | 用户拒绝 |
| `DEVICE_CODE_UNKNOWN` | 设备码不存在或已被消费 |

约束：

- `user_code` 8 位 `XXXX-XXXX`，字符集剔除易混的 `0O1I`；查询/批准时忽略大小写、空格与连字符。
- 请求的 scope 在 `/device/code` 时确定，批准页只展示、不再选择，避免批准阶段提权。
- 批准复用 `/api/access-tokens` 的铸造链路（同一 admin 校验、同一 scope 白名单、同一 `audience=cli`），因此非 admin 批准会得到权限错误。
- 存储：进程内 map + mutex + 10 分钟 TTL；已批准/已拒绝的条目在取走或过期后删除（一次性）。Ech0 为单实例 SQLite 部署，重启丢挂在途授权可接受；多实例需换持久化（已知限制）。
- CLI 按 `interval` 轮询，`slow_down` 加 5s；Web 新增 `/device` 批准页（需登录，`requiresAuth`）。设备码与批准 URL 输出到 stderr，stdout 只放数据。
- 新端点已登记到 `docs/dev/huma-raw-gin-endpoints.md`。

### 5.4 凭据存储

优先级：`--token` > `ECH0_TOKEN` > 系统钥匙串 > 回退文件。

- 钥匙串：`Bun.secrets`（service `ech0-cli`，name `token:<profile>` / `cookie:<profile>`）。
- 回退：`<config dir>/credentials.json`，权限 `0600`；`ECH0_NO_KEYCHAIN=1` 强制文件（headless CI / 测试）。
- 密码登录成功后：admin 存 CLI 令牌并删除 cookie；非 admin 存会话令牌 + refresh cookie，命令遇 401 自动 `POST /api/auth/refresh` 续签一次并重试。
- `ech0 auth logout [--purge]` 清 keychain、文件与可选 profile。

### 5.5 权限

命令所需 scope 直接来自 catalog 的 `scopes` 字段并在 help 展示；403 时提示 token 缺 scope，并提示在面板重新铸造。CLI 不做本地 scope 预检（服务端仍是唯一裁判）。

登录默认申请的 scope 必须覆盖 `catalog/scope.ts` 暴露的命令面，否则默认登录后一部分命令必然 403：`echo:read`、`echo:write`、`profile:read`、`file:read`、`file:write`。用户在 `--scopes` 里显式给出时以用户为准（可以更窄，服务端照常拒绝越权调用）。

## 6. 命令树与 help

本地命令（CLI 自身能力）：

```
ech0 auth login|status|logout
ech0 config list|use|unset|path
ech0 catalog refresh|path
ech0 install skills <dir> [--from <folder>]
ech0 version
ech0 help [words...] [--full] [--json]
```

服务端命令来自 catalog，命令树只装配 `catalog/scope.ts` 允许的操作（Echo CRUD 五条 + 文件面六条），其余操作留在 catalog 缓存里但不生成命令。命令名按 `catalog` id 分段，与本地命令保留名冲突的字段改走 `--input-<name>`：

```
ech0 echo create --content "hi" --private --tag dev   # data 直接返回新 Echo（含 id）
ech0 echo list --page 1 --page-size 20
ech0 echo update --id <id> --content "edited"        # 同样返回更新后的 Echo
ech0 echo get --id <uuid>

ech0 file upload --file photo.png                    # 上传字节，返回 file id
ech0 file external --url https://example.com/a.png   # 只登记外链
ech0 file list --search cat --page-size 20
ech0 file get --id <file_id>
ech0 file tree --storage-type local
ech0 file download --id <file_id> --output ./a.png   # binary；非 TTY 默认 stdout
```

全局选项（任意位置）：`--profile`、`--server`、`--token`、`--lang`、`--json`、`--yaml`、`--no-color`。注意：commander 会把与父命令同名的选项归到父命令，所以这些只在根命令声明一次。

### 6.1 help 体系

帮助与命令树共用一套模型（`src/help/model.ts`），所有命令的 `formatHelp` 都指向同一个渲染器（`src/help/render.ts`），因此 `--help`、`ech0 help` 与 `ech0 help <路径>` 输出一致：

| 入口 | 输出 |
| --- | --- |
| `ech0` / `ech0 --help` / `ech0 help` | 地图：本地命令、服务端命名空间（操作数与前几个操作名）、全局选项、示例 |
| `ech0 help echo` | 命名空间列表：该组下所有操作的“你实际要敲的词”与摘要 |
| `ech0 help page` | 关键词搜索：按 id 段、摘要、字段名与字段描述加权排序（ASCII 按整词匹配，CJK 按子串） |
| `ech0 help echo create` / `ech0 echo create --help` | 单命令说明：用法、接口卡片、参数、选项、全局选项、示例 |
| `ech0 help --full` | 逐条打印全部命令说明 |
| 任意入口加 `--json` | 同一模型的 JSON（地图 / 列表 / 描述符），供脚本与 agent 消费 |

描述符的每一段都由 catalog 推导，不手写：

- **接口卡片**：`METHOD /api|/ws<path>`、效果（只读/写入/危险操作）、`scope`、`transport`（multipart 上传、token 走查询参数、SSE 事件名、二进制/WebSocket）。
- **参数**：`input.properties` → kebab flag；布尔无值；文件字段（`format: binary`）显示 `--file <path>`；`required`、`enum`、`default` 进标签或说明；`bind` 标出 path/query/form。字段说明取自 schema 的 `description`，也就是 Go 结构体上的 `doc:"..."` 标签——参数语义写在后端 DTO，CLI 不重复维护。对象数组（`echo_files`、`tags`）显示 `--x <json>`，只接受 JSON，不再用逗号简写。
- **示例**：所有必填字段 + 文件字段拼成一条可直接改用的命令行，占位符写作 `<字段名>`；`echo create/update` 这类组合操作改用固定示例序列（上传 → 引用 file_id、`--tag`、`--from`），因为“必填字段”表达不了两步流程。
- **分组**：catalog 生成的命名空间是纯容器（`addHelpCommand(false)`），渲染时现算操作数与操作名。
- **提示**：从 op 元数据派生的下一步——缺 token（`ech0 auth login`）、scope 与 403、流式（Ctrl-C / `--json`）、二进制（`--output`）、multipart（`<path>` / `-`）、危险操作与弃用；运行时失败按状态码给提示（400/422 指向命令 `--help`，401/403/网络/5xx 指向对应修复动作），缺少必填参数时同时打印可直接改用的用法行，不用翻文档猜。
- **CLI 专属提示**：catalog 描述不了的东西由命令自己补——组合操作的 `--tag` / `--file-id` / `--from` / `--edit` 语义与合并规则、update 的整体覆盖、登录三种模式（设备码流程、`--token -`、管理员铸造 `audience=cli` 令牌、`--scopes` 默认值）。
- **不设独立 skills 文件**：教程就是描述符本身——参数说明、示例序列、提示都在 `ech0 help <命令>` 里，且随 catalog/CLI 版本走，不会和实现脱节。
- **catalog 不可用**：地图仍然完整；服务端命令一栏说明原因并给出 `ech0 auth login` / `ECH0_SERVER` 的下一步，不静默失败。

未知命令报退出码 2，并给最近似的命令（编辑距离）与搜索提示；`ech0 <namespace>` 裸调用打印该组帮助并退出 0。

### 6.2 skills 安装

`ech0 install skills <dir>` 把内置的 Ech0 skill 写进一个 skills 根目录（例如 Claude Code 的 `~/.claude/skills`），最终布局是 `<dir>/ech0/SKILL.md` + `<dir>/ech0/references/*.md`：

```bash
ech0 install skills ~/.claude/skills
ech0 install skills ./skills --from ./my-skill   # 用本地文件夹替代内置版本
```

- skill 源码在仓库根目录 `skills/ech0/`，是普通 Markdown（`SKILL.md` 带 frontmatter 的 name/description），不绑定某个 agent；安装只是复制。
- 构建时用 Bun 的文本导入（`import … with { type: "text" }`）把文件嵌进 CLI，因此 `bun build` / `bun build --compile` 产出的二进制离线可用，不需要随包附带 `skills/` 目录。
- 安装会整体替换 `<dir>/ech0`：旧版本残留的 references 不会留下。
- `~` 按调用者 HOME 展开；`--from` 必须是包含 `SKILL.md` 的文件夹，否则退出码 2。
- 输出：人类可读的安装路径与文件列表；`--json` 返回 `{target, files}`。

## 7. 输出契约

- 结果：默认 pretty JSON（catalog 驱动的数据形状由服务端决定）；`--json` 输出**原始响应体**（不重编码，int64 保持 number）。
- `--yaml`：与 `--json` 同一个文档（信封），改用 YAML（`Bun.YAML.stringify`，block style）。它要先解析再重编码，超过 2^53 的整数会失精度——需要逐字节精确时用 `--json`；不是 JSON 的响应体原样透传。
- `--json` 与 `--yaml` 互斥（退出码 2）。流式/二进制结果没有整份文档：`--yaml` 直接报用法错误——流式用 `--json` 按事件输出，二进制用 `--output <path>`。
- 数据为空时输出 envelope 的 `msg`（如 `发布Echo成功！`）。
- stdout 只放数据；错误与提示走 stderr。
- 颜色：TTY 且未设 `NO_COLOR` 才启用；`--no-color` 关闭。
- 退出码：`0` 成功；`1` 业务/网络错误；`2` 用法错误（含缺少必填 flag）。
- 人读表格只保留在 `config list`（catalog 操作的数据形状不可预知，表格化留给后续 `kind` 渲染）。

## 8. 配置与 profile

`~/.config/ech0/config.json`（尊重 `XDG_CONFIG_HOME`，`ECH0_CONFIG` 覆盖路径）：

```json
{
  "default": "memo",
  "profiles": {
    "memo": { "server": "https://memo.sn0w.fyi", "user": "owner", "isAdmin": true }
  }
}
```

环境变量优先级最高：`ECH0_SERVER`、`ECH0_TOKEN`、`ECH0_PROFILE`、`ECH0_LANG`、`ECH0_CONFIG`、`ECH0_NO_KEYCHAIN`。

## 9. 分发

发布由 `.github/workflows/release-cli.yml` 驱动：CLI 使用独立版本线，源头是 `apps/cli/package.json`（首个版本 `0.1.0`），推送 `cli-vX.Y.Z` tag 触发（也支持 `workflow_dispatch` 指定历史 CLI tag 重跑），与 Ech0 服务端的 `vX.Y.Z` 发版互不影响。workflow 会校验 tag 与 `package.json` 版本一致；预发布版本（如 `0.2.0-rc.1`）发布到 npm 的 `next` dist-tag，稳定版发 `latest`。

- **GitHub Release**：`ech0-cli-<os>-<arch>.tar.gz`（darwin/linux × arm64/x64，包内是 `ech0` + `LICENSE`）与 `SHA256SUMS`，由 `bun build --compile --target=bun-<os>-<arch>` 在 Linux runner 上交叉编译，资产上传到 `cli-vX.Y.Z` 这次 CLI release（不存在时创建）；命名用 `ech0-cli-*`，不与服务端的 `ech0-linux-*` 混用。
- **npm**：
  - `@ech0/cli`：启动器（`bin/ech0.cjs`，Node ≥18），按 `process.platform` / `process.arch` 找到平台包并 exec 二进制；
  - `@ech0/cli-<os>-<arch>`：只含编译产物，用 `os` / `cpu` 字段限制安装平台；四个平台包作为 `optionalDependencies`，npm 只安装匹配的那一个；
  - 没有 postinstall 下载，安装后离线可用。
- **版本**：`apps/cli/package.json` 是开发版本来源，发布时 workflow 校验 `cli-vX.Y.Z` tag 后，用 `--define process.env.ECH0_CLI_VERSION="<version>"` 内联（`src/version-info.ts`），不修改源码树；`--version` / `ech0 version` 报发布版本，服务端版本仍用 `ech0 common hello`（catalog 命令）。
- **本地复现**：`just cli dist 0.1.0` → `apps/cli/dist/release/`（四个二进制、tar.gz、`SHA256SUMS` 与可直接 `npm publish` 的目录）。
- macOS 从浏览器下载 release 资产会被打上 quarantine 标记；二进制未签名，首次运行前需 `xattr -d com.apple.quarantine ech0`（走 npm 无此问题）。
- npm 发布需要仓库 secret `NPM_TOKEN`（能发布 `@ech0` scope）；启用 OIDC provenance（`id-token: write`）。已发布的版本会跳过，重跑 workflow 不报错。

## 10. 验证

```bash
just cli typecheck    # tsc --noEmit（TS 5.9）
just cli test         # bun test
just cli gen-check    # 生成物漂移检查
just cli build        # bun build --target=bun
just cli dist 0.1.0   # 发布用二进制 + tar.gz + npm 包（apps/cli/dist/release）
go test ./internal/openapi/... ./internal/router/...
```

- Go：catalog 派生单测（bind/effect/kind/scopes/$ref 内联/指纹稳定性）+ 路由测试（`/api/_catalog` 指纹与 body 哈希一致、过期指纹 412 且 `error_code=catalog_stale`）。
- CLI：`bun test` 覆盖 envelope 映射、catalog 缓存、动态命令 flag 解析、412 重试、CLI 退出码。
- 端到端实测（真实 Go 服务端 + CLI）：init owner → login → 铸造 `audience=cli` token → `auth login` → `catalog refresh`（104 ops）→ `echo create` → `echo list`；把缓存文件改脏后 `echo create` 收到 412，CLI 自动刷新并重试成功。
- 密码登录：`printf 'pw' | ech0 auth login -u owner --password-stdin` → 自动铸造 CLI 令牌 → `echo create` 成功；device 登录：CLI 打印设备码 → 浏览器/curl 批准 → CLI 拿到令牌 → `echo create` 成功；轮询的 pending / slow_down / unknown 三个 `error_code` 在真实服务端复核。

## 11. 里程碑

| 里程碑 | 内容 | 状态 |
| --- | --- | --- |
| M1 | 目录骨架、commander 根命令、Huma→TS/zod 生成链、SPDX、gen-check | 已完成 |
| M2 | 后端 operation catalog（`/api/_catalog` + 指纹 + 412）与 CLI 动态命令；token 登录、profile/凭据存储；`auth/config/catalog/version` 本地命令 | 已完成 |
| M3 | 密码登录（admin 铸 cli token；非 admin cookie 刷新） | 已完成 |
| M4 | device authorization：后端端点 + Web `/device` 批准页 + CLI 轮询 | 已完成 |
| M5 | Release 编译产物（`ech0-cli-*.tar.gz` + `SHA256SUMS`）与 npm 包（`@ech0/cli` + 平台包），由 `release-cli.yml` 在 `cli-v*` tag 上独立发版 | 已完成 |
| M6 | shell completion、旧服务端 `/api/openapi.json` 兜底、`curl \| sh` 安装脚本 | 待做 |

## 12. 未决

- 无 `/api/_catalog` 的旧服务端：当前明确报错并提示升级；是否用 `/api/openapi.json` 现场推导 catalog 作为兜底。
- catalog 是否需要声明兼容区间（server 版本 ↔ CLI 版本），用于提示「你的 CLI 太旧」。
- 文件面尚未开放 `file.delete` / `file.update-meta`（写入类操作），需要时加入 `CLI_OPERATIONS` 即可。
