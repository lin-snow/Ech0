# Ech0 CLI 设计

日期：2026-10-30  
状态：草案；M1/M2 已落地（目录骨架、catalog 驱动的命令发现、token 登录）  
范围：`apps/cli`（Bun + TypeScript 客户端）与配套后端改动（operation catalog、device authorization）

## 1. 定位与范围

一句话：**登录到一台指定的 Ech0 实例，在终端做核心操作**（发布 / 查看 Echo 等）。

关键决策：**CLI 不内置业务命令**。命令来自服务端下发的 operation catalog，这样服务端加接口不需要用户单独升级 CLI。

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

- 清单由 Huma 文档派生（`internal/openapi/catalog`）：`operationId` 的 `-` 换成 `.`（`echo-page-get` → `echo.page.get`）；`input` 把 path/query/body 三类参数拍平成一个对象，`$ref` 递归内联；`bind` 标记每个字段落在 `path`/`query`/`body`；`effect` 由方法推导（GET/HEAD=read、DELETE=destructive、其余=write）；`kind` 由成功响应内容类型推导（json/stream/binary）。
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

- 操作 id 按 `.` 分段：`echo.create` → `ech0 echo create`，`comment.panel.settings.update` → `ech0 comment panel settings update`。
- 同名叶与命名空间共用节点：`comment.create` 是命令，`comment.create.integration` 是它的子命令（先注册浅层 id）。
- 输入字段 → flags：snake_case/camelCase 统一转 kebab（`pageSize` → `--page-size`）；boolean 无值；整数/数字/逗号数组/JSON 对象分别解析；必填字段用 commander `requiredOption`（缺失 → 退出码 2）。
- 与 CLI 保留 flag 重名（`json/profile/server/token/lang/color/help/version`）的字段改名为 `--input-<name>`。
- help 末尾附 `METHOD /path · scopes: ...`，便于确认权限与效果。

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

## 6. 命令树

本地命令（CLI 自身能力）：

```
ech0 auth login|status|logout
ech0 config list|use|unset|path
ech0 catalog refresh|path
ech0 version
ech0 help [command]
```

服务端命令（由 catalog 生成，当前 spec 104 个操作），例如：

```
ech0 echo create --content "hi" --private --tags dev
ech0 echo page get --page 1 --page-size 20
ech0 echo get --id <uuid>
ech0 user info
ech0 settings get
ech0 catalog refresh
```

全局选项（任意位置）：`--profile`、`--server`、`--token`、`--lang`、`--json`、`--no-color`。注意：commander 会把与父命令同名的选项归到父命令，所以这些只在根命令声明一次。

## 7. 输出契约

- 结果：默认 pretty JSON（catalog 驱动的数据形状由服务端决定）；`--json` 输出**原始响应体**（不重编码，int64 保持 number）。
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

- 一期：GitHub Release 附带 `ech0-cli-<os>-<arch>.tar.gz`（darwin/linux × arm64/amd64），`bun build --compile` 交叉编译；独立 workflow job，不并入现有 `ech0-linux-*` glob。
- 二期：npm `@ech0/cli` + `curl | sh` 安装脚本。
- 版本：`--version` 取 package 版本；服务端版本用 `ech0 common hello`（catalog 命令）。

## 10. 验证

```bash
just cli typecheck    # tsc --noEmit（TS 5.9）
just cli test         # bun test
just cli gen-check    # 生成物漂移检查
just cli build        # bun build --target=bun
go test ./internal/openapi/... ./internal/router/...
```

- Go：catalog 派生单测（bind/effect/kind/scopes/$ref 内联/指纹稳定性）+ 路由测试（`/api/_catalog` 指纹与 body 哈希一致、过期指纹 412 且 `error_code=catalog_stale`）。
- CLI：`bun test` 覆盖 envelope 映射、catalog 缓存、动态命令 flag 解析、412 重试、CLI 退出码。
- 端到端实测（真实 Go 服务端 + CLI）：init owner → login → 铸造 `audience=cli` token → `auth login` → `catalog refresh`（104 ops）→ `echo create` → `echo page get`；把缓存文件改脏后 `echo create` 收到 412，CLI 自动刷新并重试成功。
- 密码登录：`printf 'pw' | ech0 auth login -u owner --password-stdin` → 自动铸造 CLI 令牌 → `echo create` 成功；device 登录：CLI 打印设备码 → 浏览器/curl 批准 → CLI 拿到令牌 → `echo create` 成功；轮询的 pending / slow_down / unknown 三个 `error_code` 在真实服务端复核。

## 11. 里程碑

| 里程碑 | 内容 | 状态 |
| --- | --- | --- |
| M1 | 目录骨架、commander 根命令、Huma→TS/zod 生成链、SPDX、gen-check | 已完成 |
| M2 | 后端 operation catalog（`/api/_catalog` + 指纹 + 412）与 CLI 动态命令；token 登录、profile/凭据存储；`auth/config/catalog/version` 本地命令 | 已完成 |
| M3 | 密码登录（admin 铸 cli token；非 admin cookie 刷新） | 已完成 |
| M4 | device authorization：后端端点 + Web `/device` 批准页 + CLI 轮询 | 已完成 |
| M5 | Release 编译产物、npm 包、curl 脚本 | 待做 |
| M6 | `kind=stream/binary` 渲染、shell completion、旧服务端 `/api/openapi.json` 兜底 | 待做 |

## 12. 未决

- 无 `/api/_catalog` 的旧服务端：当前明确报错并提示升级；是否用 `/api/openapi.json` 现场推导 catalog 作为兜底。
- `kind=stream`（SSE，如 copilot chat）与 `kind=binary`（导出/下载）的 CLI 渲染。
- catalog 是否需要声明兼容区间（server 版本 ↔ CLI 版本），用于提示「你的 CLI 太旧」。
- 上传类操作（multipart）catalog 尚未建模（`bind=file` 未落地）。
