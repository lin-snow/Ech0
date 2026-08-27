# 非 Huma 端点清单（仍走裸 gin 的接口及原因）

> 现状快照（分支 `feat/huma-openapi`）：所有 **JSON 业务端点** 已全量迁移到 Huma type-first；本文档清点**剩下仍走裸 gin、无法纳入 Huma JSON 契约**的端点及其原因。
> 关联文档：`docs/dev/architecture-overview.md`（分层与 Huma 总览）、`docs/dev/auth-design.md`（双 token / OAuth / Passkey）、`docs/dev/access-token-scope-design.md`（scope/audience）、`docs/usage/mcp-usage.md`（MCP）。
> 涉及代码：`internal/router/*.go` 的 `setupXxxRoutes`（裸 gin）vs `registerXxx`（Huma）。

## 1. 背景与分界

路由装配分两条线（见 `internal/router/router.go` 的 `SetupRouter`）：

- **Huma type-first**：13 个业务域各有一个 `registerXxx(api, h, revoker)`，由 `registerOperations`（`internal/router/huma.go`）统一聚合，约 **100 个 JSON 端点**。handler 为框架中立签名 `func(ctx context.Context, *XxxInput) (commonModel.Result[T], error)`，**不碰 `*gin.Context`**，只返回一个 typed JSON 信封；OpenAPI 从 Go 类型生成。
- **裸 gin**：散落在各域的 `setupXxxRoutes(groups, h)` 里，handler 多为 `func() gin.HandlerFunc`，直接操作 `*gin.Context`。

一个端点**只要满足下列任意一条，就上不了 Huma JSON 契约**，必须留在裸 gin：

1. **流式 / 长连接响应** —— SSE、WebSocket（Huma 是「单次请求 → 单个 typed body」，长连接不适用）；
2. **需要直接操作 `*gin.Context`** —— 写 `Set-Cookie`、读 cookie、发 302 跳转、从原始请求头推导状态（框架中立的 Huma handler 拿不到 gin 上下文）；
3. **非 JSON 请求体** —— `multipart/form-data` 文件上传；
4. **非 JSON 响应体** —— 二进制下载、XML/纯文本资源、静态文件、SPA HTML；
5. **非 REST 协议** —— JSON-RPC、第三方 `http.Handler` 直挂。

> 注：SSE / multipart 在 Huma 上游有部分支持能力，但本项目坚持「handler 框架中立、只产出 JSON 信封」的契约，因此这两类也统一留在裸 gin，与上面第 1/3 条一致处理。

## 2. 按原因分类

### A. 流式响应：SSE / WebSocket（3）

| 方法 | 路径 | Handler | 分组 / 鉴权 |
|---|---|---|---|
| POST | `/api/chat` | `CopilotHandler.Ask` | Auth · `admin:settings` |
| GET | `/api/system/logs/stream` | `DashboardHandler.SSESubscribeSystemLogs` | Auth · `admin:settings` |
| GET | `/ws/system/logs` | `DashboardHandler.WSSubscribeSystemLogs` | WS 组（鉴权在 handler 内） |

`/api/chat` 把 Agent ReAct 循环逐事件转成 Chat SSE（`searching\|sources\|delta\|done\|error`）。WebSocket 与请求-响应模型根本不兼容。

### B. multipart 上传（2）

| 方法 | 路径 | Handler | 分组 / 鉴权 |
|---|---|---|---|
| POST | `/api/files/upload` | `FileHandler.UploadFile` | Auth · `file:write` |
| POST | `/api/migration/upload` | `MigrationHandler.UploadSourceZip` | Auth · `admin:settings` |

请求体是 `multipart/form-data` 文件流，非 JSON body。

### C. 二进制下载 / 文件流（3）

| 方法 | 路径 | Handler | 分组 / 鉴权 |
|---|---|---|---|
| GET | `/api/file/stream` | `FileHandler.StreamFileByPath` | Auth · `file:read` |
| GET | `/api/file/:id/stream` | `FileHandler.StreamFileByID` | Auth · `file:read` |
| GET | `/api/migration/export/download` | `MigrationHandler.DownloadExport` | Auth · `admin:settings` |

响应是字节流（图片 / 快照 zip / octet-stream），非 JSON 信封。

### D. OAuth 302 跳转（2）

| 方法 | 路径 | Handler | 分组 / 鉴权 |
|---|---|---|---|
| GET | `/oauth/:provider/login` | `AuthHandler.OAuthLogin` | Resource（公开 · NoCache） |
| GET | `/oauth/:provider/callback` | `AuthHandler.OAuthCallback` | Resource（公开 · NoCache） |

响应是 `Location` 重定向到 IdP / 回前端，非 JSON。`:provider` 统一覆盖 github/google/qq/custom。

### E. Cookie 读写 / token 签发 / WebAuthn 仪式（8）

| 方法 | 路径 | Handler | 分组 / 鉴权 | 具体原因 |
|---|---|---|---|---|
| POST | `/api/login` | `AuthHandler.Login` | Public | 校验后写 refresh HttpOnly cookie |
| POST | `/api/auth/refresh` | `AuthHandler.Refresh` | Public | 读 refresh cookie、续签 |
| POST | `/api/auth/logout` | `AuthHandler.Logout` | Public | 清 cookie + 吊销 token |
| POST | `/api/auth/exchange` | `AuthHandler.Exchange` | Public | 一次性 code 换 token |
| POST | `/api/passkey/login/begin` | `AuthHandler.PasskeyLoginBeginV2` | Public | 从 gin 请求读 Origin 推导 RP-ID/Origin |
| POST | `/api/passkey/login/finish` | `AuthHandler.PasskeyLoginFinishV2` | Public | 同上 + 写 refresh cookie |
| POST | `/api/passkey/register/begin` | `AuthHandler.PasskeyRegisterBeginV2` | Auth · `profile:write` | 从 gin 请求读 Origin 推导 RP-ID/Origin |
| POST | `/api/passkey/register/finish` | `AuthHandler.PasskeyRegisterFinishV2` | Auth · `profile:write` | 同上 |

这批都需要直接操作 `*gin.Context`（写/读 cookie，或经 `getPasskeyOriginAndRPID` 从原始请求头推导 WebAuthn 的 RP-ID 与 Origin），超出框架中立 handler 的能力边界，整组留在裸 gin。

### F. 第三方 captcha（1）

| 方法 | 路径 | Handler | 分组 / 鉴权 |
|---|---|---|---|
| ANY | `/api/cap/*any` | `gin.WrapH(captchaHandler)` | Public |

第三方 captcha 库提供的 `http.Handler`，经 `gin.WrapH` 整子树直挂，非 REST/JSON 资源语义。

### G. MCP JSON-RPC（2）

| 方法 | 路径 | Handler | 分组 / 鉴权 |
|---|---|---|---|
| POST | `/mcp` | `MCPHandler.ServeEndpoint` | MCP 组 · RequireAuth + RateLimit + OriginGuard + Audience `mcp-remote` |
| GET | `/mcp` | `MCPHandler.ServeEndpoint` | 同上 |

JSON-RPC 2.0 协议（方法分发，非 REST 资源），且鉴权维度（audience/scope）与普通 API 不同。

### H. 非 JSON 资源 / SPA / 静态文件（6）

| 方法 | 路径 | Handler | 分组 / 鉴权 | 内容类型 |
|---|---|---|---|---|
| GET | `/robots.txt` | `CommonHandler.GetRobotsTxt` | Resource（公开） | `text/plain` |
| GET | `/sitemap.xml` | `CommonHandler.GetSitemap` | Resource（公开） | `application/xml` |
| GET | `/rss` | `CommonHandler.GetRss` | Resource（公开） | `application/atom+xml` / `application/xml`（按 UA 协商） |
| GET | `/healthz` | `CommonHandler.Healthz` | Resource（公开） | 探活（惯例裸 gin 探针） |
| —（NoRoute） | 任意未命中 | `WebHandler.Templates` | Engine 级 | SPA `index.html` fallback |
| GET | `/api/files/*` | `StaticFS` | 专用组 · `StaticFileSecurity`（目录穿越防护） | 本地上传文件静态服务 |

## 3. 汇总

| 类别 | 端点数 |
|---|---|
| A 流式（SSE/WS） | 3 |
| B multipart 上传 | 2 |
| C 二进制下载/流 | 3 |
| D OAuth 302 跳转 | 2 |
| E Cookie/token/WebAuthn | 8 |
| F captcha | 1 |
| G MCP JSON-RPC | 2 |
| H 非 JSON 资源/SPA/静态 | 6 |
| **合计裸 gin** | **27** |

对照面：13 个业务域（init / auth / common / echo / connect / user / setting / file / dashboard / copilot / comment / migration / embedding）均已在 Huma，约 100 个 JSON 端点，经 `registerOperations` 聚合。

## 4. 维护说明

- **核对当前裸 gin 端点**：`grep -rnE '\.(GET|POST|PUT|DELETE|Any|NoRoute|StaticFS)\(' internal/router/*.go`（排除 `route(api`/`register` 的 Huma 调用）。
- **新增端点该放哪**：先用 §1 的 5 条判定规则过一遍——全不沾就走 Huma（新增 `registerXxx` 里的 `route(api, ...)`），命中任意一条就放对应域的 `setupXxxRoutes`，并把它补进本文档对应类别。
- 本表为人工维护，**不随代码自动同步**；改路由时请一并更新（端点数与分组/鉴权）。
