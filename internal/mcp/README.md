# internal/mcp

Ech0 内建的 MCP（Model Context Protocol）Server 实现。

通过 `/mcp` 端点（复用主服务 6277 端口）对外暴露 **Streamable HTTP**，任意支持该传输方式的 MCP Host 均可统一调用 Tools、读取 Resources；鉴权与 Scope 与 REST API 共用同一套 JWT。

**双时代（dual-era）**：同一端点同时服务 2026-07-28 的无状态请求与 `initialize` 握手时代（2025-11-25 / 2025-06-18 / 2025-03-26）的旧客户端。旧客户端没有 fall-forward 机制，拒绝它们等于直接连不上，因此必须兼容。

## 架构

```
MCP-compatible client / Host
    │
    ▼
┌──────────────────────────────────────────────┐
│  Gin Router  /mcp  (POST；GET/DELETE → 405)  │
│  ├─ middleware.RateLimit                     │
│  ├─ middleware.OriginGuard（防 DNS rebinding）│
│  └─ middleware.RequireAudience(mcp-remote)   │
│     → Handler.ServeEndpoint()                │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Server                                      │
│  ├─ JSON-RPC 2.0 解析与分发                  │
│  ├─ era 判定（每请求，不记忆连接状态）        │
│  ├─ modern：传输头 + _meta 必填字段校验       │
│  ├─ legacy：initialize / ping，无头无 _meta   │
│  ├─ 内置 scope 校验（403 + WWW-Authenticate） │
│  ├─ tool 执行超时（10s context deadline）     │
│  └─ 结构化审计日志（zap，含 era）             │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Registry                                    │
│  ├─ Tool 注册表（name → binding，注册序稳定） │
│  ├─ Resource 注册表（具体 URI）               │
│  └─ ResourceTemplate 注册表（前缀路由）       │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Adapter（按业务域拆分）                      │
│  ├─ adapter_echo.go    → EchoService         │
│  ├─ adapter_user.go    → UserService         │
│  ├─ adapter_comment.go → CommentService      │
│  ├─ adapter_file.go    → FileService         │
│  ├─ adapter_common.go  → CommonService       │
│  ├─ adapter_connect.go → ConnectService      │
│  ├─ adapter_agent.go   → AgentService        │
│  ├─ adapter_webhook.go → SettingService      │
│  └─ adapter_dashboard.go → DashboardService  │
│  （不直连 Repository，强制走 Service 层）      │
└──────────────────────────────────────────────┘
```

## 文件职责

| 文件 | 职责 |
|------|------|
| `jsonrpc.go` | JSON-RPC 2.0 基础类型与错误码：规范保留码 `-32020`/`-32022`，应用自定义 `-40300`（insufficient scope，落在保留区间之外）；`RPCError` 带非序列化的 HTTP 状态/挑战头覆盖字段 |
| `capability.go` | 协议版本集合（modern + legacy）、`era`、ServerCapabilities、DiscoverResult、InitializeResult、ResultEnvelope、CacheInfo |
| `tools.go` | Tool 类型：ToolDefinition（含 `annotations`）、ToolsListResult、ToolCallResult（含 `structuredContent`） |
| `resources.go` | Resource 与 ResourceTemplate 类型、各自的 list 结果、ResourceReadResult |
| `registry.go` | Tool / Resource / ResourceTemplate 注册表；具体 URI 精确匹配优先于模板前缀匹配 |
| `adapter.go` | Adapter 结构体、RegisterAll 入口、参数/结果 helper、tool 行为注解 helper |
| `adapter_echo.go` | Echo 域：帖子 CRUD + 点赞/今日/热门/随机/历史上的今天/标签 tools，posts/tags resources，`ech0://posts/{id}` template |
| `adapter_user.go` | User 域：profile/me resource |
| `adapter_comment.go` | Comment 域：`list_comments`、`create_comment` / `create_integration_comment` tools；recent/guide resources |
| `adapter_file.go` | File 域：list/get/delete/create_external file tools，上传指南 resource |
| `adapter_common.go` | Common 域：heatmap resource |
| `adapter_connect.go` | Connect 域：list/add/delete connects tools，connect self resource |
| `adapter_agent.go` | Agent 域：get_recent tool（AI 近况摘要） |
| `adapter_webhook.go` | Webhook 域：list/create/update/delete/test webhook tools |
| `adapter_dashboard.go` | Dashboard 域：`ech0://stats/visitors` resource（近 7 天 PV/UV，需 admin scope） |
| `server.go` | MCP Server 核心：era 判定、传输校验、方法分发、scope 校验、超时控制、审计日志 |
| `handler.go` | Gin 桥接层：组装 Registry → Adapter → Server，暴露 `ServeEndpoint()` |
| `server_test.go` | 单元测试：两个时代的完整流程、版本协商、传输头校验、scope 拒绝、模板列举、缓存策略、错误映射 |

## 请求处理流程

1. HTTP 请求进入 `/mcp`，经过限流、Origin 校验、JWT 鉴权（仅 POST；GET/DELETE 返回 405）
2. `Handler.ServeEndpoint()` 将 `gin.Context` 转交 `Server.ServeHTTP()`
3. 解析 JSON-RPC；无 `id` 的通知（含 `notifications/initialized`）直接返回 202
4. `resolveEra` 按 `MCP-Protocol-Version` 头与 `params._meta` 判定时代；版本不受支持返回 HTTP 400 + `-32022`
5. **modern（2026-07-28）**：校验头与 body 一致（`MCP-Protocol-Version` / `Mcp-Method` / `Mcp-Name`，支持 Base64 sentinel），并要求 `_meta` 带 `protocolVersion` 与 `clientCapabilities`；违规分别返回 `-32020` 与 `-32602`，均为 HTTP 400
6. **legacy**：不要求任何头与 `_meta`，额外支持 `initialize`（协商版本，不发放 session）与 `ping`
7. 按 `method` 分发；未知方法返回 HTTP 404 + `-32601`
8. `tools/call` / `resources/read` 查 Registry 取 binding，比对 token scope；不足返回 HTTP 403 + `WWW-Authenticate: Bearer error="insufficient_scope", scope="…"`
9. 调用 Adapter 注册的业务函数，Adapter 转发到 Ech0 Service 层
10. modern 结果盖上 `resultType: "complete"`、`_meta.serverInfo` 与缓存提示；legacy 结果不带这些 2026 专有字段

## 扩展新 Tool / Resource

1. 新建 `adapter_<domain>.go`（如 `adapter_file.go`）
2. 实现注册函数（如 `registerFileTools(reg)`）和业务 handler
3. 在 `adapter.go` 的 `RegisterAll()` 中添加一行调用
4. Tool 声明 `InputSchema`（JSON Schema 2020-12）、`Annotations`（用 `readOnlyHints()` 等 helper）与所需 scopes
5. Resource 声明 `Cache`（`publicCache` 仅用于与调用者无关的内容）；带占位符的 URI 必须用 `RegisterResourceTemplate`

不需要修改 `server.go`、`registry.go` 或路由代码。


## 相关文档

- [MCP 接入指南](../../docs/usage/mcp-usage.md) — Token 创建、Host 配置、curl 示例
