# internal/mcp

Ech0 内建的 MCP（Model Context Protocol）Server 实现。

通过 `/mcp` 端点（复用主服务 6277 端口）对外暴露 **Streamable HTTP**，任意支持该传输方式的 MCP Host 均可统一调用 Tools、读取 Resources；鉴权与 Scope 与 REST API 共用同一套 JWT。

## 架构

```
MCP-compatible client / Host
    │
    ▼
┌──────────────────────────────────────────────┐
│  Gin Router  /mcp  (POST；GET/DELETE → 405)  │
│  ├─ middleware.RateLimit                     │
│  ├─ middleware.OriginGuard                   │
│  ├─ middleware.RequireAuth             │
│  └─ Handler.ServeEndpoint()                  │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Server                                      │
│  ├─ JSON-RPC 2.0 解析与分发                  │
│  ├─ 传输元数据校验（MCP-Protocol-Version /   │
│  │   Mcp-Method / Mcp-Name，2026-07-28）     │
│  ├─ server/discover / tools/* / resources/*  │
│  ├─ 内置 scope 校验（per tool/resource）      │
│  ├─ tool 执行超时（10s context deadline）     │
│  └─ 结构化审计日志（zap）                     │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│  Registry                                    │
│  ├─ Tool 注册表（name → handler + scopes）    │
│  └─ Resource 注册表（uri → handler + scopes） │
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
| `jsonrpc.go` | JSON-RPC 2.0 基础类型：Request、Response、RPCError、错误码常量（含 2026-07-28 规范保留码 `-32020`/`-32022`） |
| `capability.go` | MCP 协议版本、ServerCapabilities、DiscoverResult、ResultEnvelope（resultType + `_meta.serverInfo`）、CacheInfo（ttlMs + cacheScope）、ServerInfo |
| `tools.go` | Tool 相关类型：ToolDefinition、ToolCallParams、ToolCallResult、ContentItem |
| `resources.go` | Resource 相关类型：ResourceDefinition、ResourceReadParams、ResourceReadResult |
| `registry.go` | Tool/Resource 注册表，支持精确匹配与 URI 前缀匹配 |
| `adapter.go` | Adapter 结构体、构造函数、RegisterAll 入口、通用参数/结果 helper |
| `adapter_echo.go` | Echo 域：帖子 CRUD + 点赞/今日/热门/随机/历史上的今天/标签 tools，posts/tags resources |
| `adapter_user.go` | User 域：profile/me resource |
| `adapter_comment.go` | Comment 域：`list_comments`、`create_comment` / `create_integration_comment` tools；`ech0://comments/recent`、`ech0://guide/integration-comment` resources |
| `adapter_file.go` | File 域：list/get/delete/create_external file tools |
| `adapter_common.go` | Common 域：heatmap resource |
| `adapter_connect.go` | Connect 域：list/add/delete connects tools，connect self/info resources |
| `adapter_agent.go` | Agent 域：get_recent tool（AI 近况摘要） |
| `adapter_webhook.go` | Webhook 域：list/create/update/delete/test webhook tools |
| `adapter_dashboard.go` | Dashboard 域：`ech0://stats/visitors` resource（近 7 天 PV/UV，需 admin scope） |
| `server.go` | MCP Server 核心：请求解析、传输头校验、方法分发、scope 校验、超时控制、审计日志 |
| `handler.go` | Gin 桥接层：组装 Registry → Adapter → Server，暴露 `ServeEndpoint()` |
| `server_test.go` | 单元测试：server/discover、版本协商与传输头校验、tool 调用、scope 拒绝、resource 读取、错误处理 |

## 请求处理流程

1. HTTP 请求进入 `/mcp`，经过限流、Origin 校验、JWT 鉴权（仅 POST；GET/DELETE 返回 405）
2. `Handler.ServeEndpoint()` 将 `gin.Context` 转交 `Server.ServeHTTP()`
3. `Server` 解析 JSON-RPC，校验 2026-07-28 传输元数据：`MCP-Protocol-Version` 头与 `params._meta` 中的协议版本必须一致且受支持，`Mcp-Method` 必须与 body method 一致，`tools/call` / `resources/read` 还要求 `Mcp-Name`（支持 Base64 sentinel 编码）与 body 一致；违规返回 HTTP 400 + `-32020`/`-32022`
4. 按 `method` 分发（`server/discover`、`tools/list`、`tools/call`、`resources/list`、`resources/read`；未知方法返回 HTTP 404 + `-32601`）
5. `tools/call` 和 `resources/read` 会查 `Registry` 获取 handler 与所需 scopes
6. 从 `viewer.Context` 提取当前 token 的 scopes，做细粒度权限校验
7. 调用 `Adapter` 中注册的业务函数，Adapter 转发到 Ech0 Service 层
8. 结果统一盖上 `resultType: "complete"` 与 `_meta.serverInfo`；discover/list/read 结果附缓存提示（`ttlMs` + `cacheScope`）后返回

## 扩展新 Tool / Resource

1. 新建 `adapter_<domain>.go`（如 `adapter_file.go`）
2. 实现注册函数（如 `registerFileTools(reg)`）和业务 handler
3. 在 `adapter.go` 的 `RegisterAll()` 中添加一行调用
4. 声明 `InputSchema`（JSON Schema）和所需 `scopes`

不需要修改 `server.go`、`registry.go` 或路由代码。

## 相关文档

- [MCP 接入指南](../../docs/usage/mcp-usage.md) — Token 创建、Host 配置、curl 示例
