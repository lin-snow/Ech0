# Ech0 CLI 设计

日期：2026-10-30  
状态：草案（M1 已落地：目录骨架 + 代码生成链；登录方式与生成物策略已定）  
范围：`apps/cli`（Bun + TypeScript 独立客户端）与配套后端改动（device authorization）

## 1. 定位与范围

一句话：**登录到一台指定的 Ech0 实例，在终端做核心操作**（发布 / 查看 Echo 等）。

v1 目标用户：实例 owner（admin）。非目标：

- 不做 TUI（`internal/cli` 的 Go TUI 保持现状，`cmd/` 不动）。
- 不做离线操作 `data/` 或 SQLite（那是 Go 端 `capsule`/`snapshot` 的职责）。
- 不做 MCP 代理、不做全量管理台（settings / OAuth / S3 / embedding 命令后续再议）。

## 2. 技术栈（已定）

| 项 | 选择 | 依据 |
| --- | --- | --- |
| 运行时 | Bun ≥ 1.4（`bun build --compile` 交叉编译单文件二进制；`Bun.secrets` 取系统钥匙串） | 分发目标是零依赖二进制 |
| 语言 | TypeScript，`apps/cli` 内钉 **5.9** | `@hey-api/openapi-ts` 0.99 在 TS 7 原生 tsc 上崩溃（`Cannot read properties of undefined (reading 'AnyKeyword')`）；生成物本身在 TS 7 下也能编译，版本只需在本包内隔离 |
| 命令框架 | `commander` 15 | `--help`/补全/子命令树成熟，无运行时魔法 |
| 校验 | `zod` v4 | 从 Huma spec 生成的 schema 使用 v4 API（`z.uuid()`） |
| 类型/schema 生成 | `@hey-api/openapi-ts` 0.99，插件 `@hey-api/typescript` + `zod` | 单工具同时产出类型与 zod；`openapi-typescript` 只出类型；`orval` 8.45 直接失败（见 §4.3） |
| 测试 | `bun test` | 内置，无额外 runner |
| 包名 | `@ech0/cli`，bin `ech0` | npm `ech0` 已被空占位包占用（registry 有包、无 version/maintainer） |

## 3. 目录结构

```
apps/cli/
  package.json            # @ech0/cli, bin: ech0, type: module
  tsconfig.json           # ES2024 / bundler / strict / noEmit
  openapi-ts.config.ts    # 生成配置（input = ../../internal/openapi/openapi.yaml, SPDX header）
  justfile                # 根 justfile 通过 `mod cli "apps/cli/justfile"` 暴露
  src/
    index.ts              # #!/usr/bin/env bun → parseAsync
    commands/             # 每个子命令一文件，root.ts 负责注册
    api/
      types.gen.ts        # 生成：Huma spec 的 TS 类型（入库）
      zod.gen.ts          # 生成：zod v4 schema（入库）
      http.ts             # 手写：fetch、鉴权头、envelope 解包、错误映射
      auth.ts             # 手写：raw-gin 认证端点（login/refresh/device）
    config/               # profile 解析、凭据存取（keychain / 回退文件）
    output/               # 人类可读渲染、--json、退出码
    ui/                   # 交互提示、颜色、spinner
  test/
```

## 4. API 类型与校验

### 4.1 SSOT

Huma 定义是唯一事实源：根仓 `go run ./cmd/openapi-gen` → `internal/openapi/openapi.yaml`（`just openapi-check` 保证不漂移）。CLI 从这份已提交的 spec 生成，不依赖运行中的服务端：

```bash
just cli gen        # bun run gen:api → src/api/{types,zod}.gen.ts
just cli gen-check  # gen + git diff --exit-code -- src/api（与 just openapi-check 同一思路，产物入库）
```

### 4.2 envelope 与错误

所有 JSON 端点返回 `commonModel.Result[T]`：`{code, data, msg, message_key, message_params, error_code}`。CLI 约定：

- 解包：`code !== 1`（`commonModel.DEFAULT_SUCCESS_CODE = 1`）→ 抛 `Ech0Error{status, code, errorCode, messageKey, params}`；`data` 交给命令层。
- 错误文案：优先 `message_key` + `message_params` 走本地 i18n 映射，其次 `msg`（见 `docs/dev/i18n-contract.md`）。
- 请求带 `Accept-Language`，由 `--lang` / `ECH0_LANG` 控制。

### 4.3 int64 陷阱（已实测）

Huma 的 `int64` 生成 `z.coerce.bigint()`。若把 zod 的解析结果 `JSON.stringify`，会抛 `TypeError: Do not know how to serialize a BigInt`。因此：

- zod 只用于**校验形状**（错误信息、`--debug-validate`），不参与输出。
- `--json` 输出**原始响应体**（fetch 读到的文本），不做重新序列化。

`orval` 不可用的原因也记录在此：Huma 产出的 schema 名 `ResultInterface {}` 含空格与花括号，违反 OpenAPI component key 规则，orval 直接拒绝整个 spec。清理该命名属于后端可选的 spec 卫生项。

### 4.4 spec 之外的手写端点

raw gin（见 `docs/dev/huma-raw-gin-endpoints.md`）不在 spec 中，必须手写在 `src/api/auth.ts`：`POST /api/login`、`/api/auth/refresh|logout|exchange`、`/api/oauth/{provider}/login|callback`、passkey 两个端点，以及本设计新增的 device 端点。

## 5. 认证与凭据

### 5.1 三种登录方式（均已确认要做）

| 方式 | 命令 | 前提 | 说明 |
| --- | --- | --- | --- |
| 粘贴 access token | `ech0 auth login --server URL --token -` | 用户在面板铸造 token（`admin:token`） | 零后端改动，CI 首选（也支持 `--token`、`ECH0_TOKEN`） |
| 用户名 + 密码 | `ech0 auth login --server URL -u owner` | `/api/login` | 详见 5.2 |
| device authorization | `ech0 auth login --server URL` | 需后端新端点（§5.3） | 交互最少，适合服务器/CI |

### 5.2 密码登录后的凭据形态

`POST /api/login` 只返回 `access_token` + `expires_in`；refresh token 只走 httpOnly cookie（`TokenPair.RefreshToken` 是 `json:"-"`）。因此：

- **admin**：用登录得到的 JWT 调 `POST /api/access-tokens`，铸 `audience: "cli"`、`expiry: "one_month"`、scope 可配置的长期令牌并存下来；后续不再依赖 cookie。
- **非 admin**：目前 `/access-tokens` 服务层强制 admin（`CreateAccessToken` 检查 `user.IsAdmin`），只能保存 JWT + refresh cookie，过期时 `POST /api/auth/refresh` 续期。
- v1 面向 owner，非 admin 路径按“能用但不优化”处理。

`audience` 取 `authModel.AudienceCLI`（`internal/model/auth/scope.go` 已存在），与 `AudiencePublic`/`integration`/`mcp-remote` 并列。

### 5.3 Device authorization（RFC 8628，后端改造）

新增 3 个端点，沿用 raw gin + `commonModel.Result` 之外的 RFC 错误体（与 login 同族，不进 Huma spec）：

| 方法 | 路径 | 认证 | 请求 | 响应 |
| --- | --- | --- | --- | --- |
| POST | `/api/auth/device/code` | public | `{client_name, scopes[]}` | `{device_code, user_code, verification_uri, verification_uri_complete, expires_in: 600, interval: 5}` |
| POST | `/api/auth/device/token` | public | `{device_code}` | 成功 `{access_token, token_type: "Bearer", expires_in, scopes[]}`；失败 400 `{error: authorization_pending \| slow_down \| access_denied \| expired_token}` |
| POST | `/api/auth/device/approve` | 已登录会话 | `{user_code, scopes[]}` | `204` |

约束：

- `user_code` 8 位 `XXXX-XXXX`，剔除易混字符（`0O1I`）；不区分大小写、容忍空格与连字符。
- 批准者必须是 admin（与 access token 铸造一致）；批准时按请求的 `scopes` 取交集并落库为 `audience=cli` 的 access token。
- 存储：进程内 `map[deviceCode]entry` + `sync.Mutex` + 10 分钟 TTL。Ech0 单实例（SQLite）部署，重启丢挂在途授权可接受；多实例需换持久化（记录为已知限制）。
- CLI 轮询严格遵守 `interval`，收到 `slow_down` 加 5s。
- Web 批准页新增路由 `/device`：登录态下输入 `user_code` → 显示申请的 scopes → 确认。
- 新端点必须同步登记到 `docs/dev/huma-raw-gin-endpoints.md` 的清单。

### 5.4 凭据存储

优先级：`ECH0_TOKEN` / `--token` > 系统钥匙串 > 回退文件。

- 钥匙串：`Bun.secrets.set({service: "ech0-cli", name: "token:<profile>"})`。
- 回退：`~/.config/ech0/credentials.json`，权限 `0600`；`ECH0_NO_KEYCHAIN=1` 强制走文件（headless CI / 测试）。
- `ech0 auth logout` 同时清钥匙串、回退文件与内存缓存。

### 5.5 命令 → scope 映射（v1）

| 命令 | scope |
| --- | --- |
| `echo add` / `echo rm` | `echo:write` |
| `echo ls` / `echo show` | `echo:read` |
| `me` / `auth status` | `profile:read` |
| `auth login`（铸造 token） | `admin:token` |
| device approve | 批准者需 admin；签发的 token 取批准时选择的 scope 集合 |

## 6. 命令树（v1）

```
ech0 auth login        # 用 access token 登录（--token / ECH0_TOKEN，传 `-` 读 stdin）
ech0 auth status [--verify]
ech0 auth logout [--purge]
ech0 config list | use <name> | unset <name> | path
ech0 echo add [content|-] [--file PATH] [--private] [--tags a,b]
ech0 echo ls [--limit N] [--page N] [--search TEXT] [--all]
ech0 echo show <id>
ech0 echo rm <id> [--yes]
ech0 me
ech0 version [--remote]
```

全局选项（放在任意位置）：`--profile`、`--server`、`--token`、`--lang`、`--json`、`--no-color`。注意：commander 会把与父命令同名的选项归到父命令，所以 `--token/--server/--profile` 只在根命令声明一次，子命令不再重复声明。

后续（v2+）：`comment`、`file`、`tag`、`connect`、`chat`（SSE）、`completion`。

## 7. 输出契约

- 默认：人类可读（表格 / 纯文本）；`--json`：原始响应体，一行一个对象（列表）。
- stdout 只放数据；日志、进度、错误全部走 stderr。
- `echo ls --all --json` 输出 NDJSON（每行一个 Echo），其余命令的 `--json` 都是单次响应的原始体。
- 颜色：`NO_COLOR` / `--no-color` 关闭；TTY 外自动关闭。
- 退出码：`0` 成功；`1` 业务或网络错误；`2` 用法错误（commander `.exitOverride()` 映射）。
- 分页：`echo ls` 默认 `limit=20`，`--all` 自动翻页（`/api/echo/page`）。

## 8. 配置与 profile

`~/.config/ech0/config.json`（尊重 `XDG_CONFIG_HOME`，`ECH0_CONFIG` 可覆盖路径）：

```json
{
  "default": "memo",
  "profiles": {
    "memo": { "server": "https://memo.sn0w.fyi", "audience": "cli", "user": "owner" }
  }
}
```

环境变量覆盖优先级最高：`ECH0_SERVER`、`ECH0_TOKEN`、`ECH0_PROFILE`、`ECH0_LANG`。

## 9. 分发

- 一期：GitHub Release 附带 `ech0-cli-<os>-<arch>.tar.gz`（darwin/linux × arm64/amd64），由 `bun build --compile` 交叉编译；新增独立 workflow job，不并入现有 `prepare-release` 的 `ech0-linux-*` glob。
- 二期：npm `@ech0/cli` + `curl | sh` 安装脚本。
- 版本：`--version` 取 package 版本；`ech0 version --remote` 另显示服务端版本（`GET /api/hello`）。

## 10. 验证与 CI

```bash
just cli typecheck    # tsc --noEmit（TS 5.9）
just cli test         # bun test
just cli gen-check    # 生成物漂移检查
just cli build        # bun build --target=bun
```

- 单测用 `Bun.serve` 起假服务端覆盖：envelope 解包、错误映射、命令请求体、profile/token 优先级、CLI 退出码（`test/*.test.ts`，`bun test`）。
- 根 `just check` 后续接入 `just cli typecheck gen-check`；CI 需在对应 job 增加 `oven-sh/setup-bun`。

## 11. 里程碑

| 里程碑 | 内容 | 状态 |
| --- | --- | --- |
| M1 | 目录骨架、commander 根命令、Huma→TS/zod 生成链、SPDX、gen-check | 已完成 |
| M2 | token 登录 + profile/凭据存储 + `echo add/ls/show/rm` + `me`/`version` | 已完成 |
| M3 | 密码登录（admin 铸 cli token；非 admin cookie 刷新） | 待做 |
| M4 | device authorization：后端 3 端点 + Web `/device` 批准页 + CLI 轮询 | 待做 |
| M5 | Release 产物（compile 矩阵）、npm 包、curl 脚本 | 待做 |
| M6 | comment / file / tag / connect / chat 命令 | 待做 |
