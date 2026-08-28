# Ech0 MCP 接入指南

Ech0 内建 [MCP（Model Context Protocol）](https://modelcontextprotocol.io/) Server：在标准协议下把**帖子、标签、评论、文件、互联、资料与统计**等能力以 **Tools** 与 **Resources** 暴露给 AI 工作流。传输层为 **Streamable HTTP**（JSON-RPC 2.0 over HTTP），与主服务同端口，通过 **Bearer Token** 与 **Scope** 做最小权限控制。

> **Audience 要求**：MCP 端点**仅允许** Audience 为 **`mcp-remote`** 的 Access Token 访问。使用其他 audience（如 `public-client`、`cli`）的 Token 将被拒绝（HTTP 403）。

> 架构与实现细节见 [internal/mcp/README.md](../../internal/mcp/README.md)。

## 接入方式

Ech0 的 MCP 端点采用 **Streamable HTTP**（JSON-RPC over HTTP），与 [MCP 规范](https://modelcontextprotocol.io/) 一致。若你的环境支持远程 MCP，并能携带 `Authorization: Bearer <token>`，将 `url` 指向本服务即可。

若运行环境只支持本地 stdio 进程而非远程 HTTP，可通过网关或代理转发到本端点。

## 在管理后台查看

管理后台 **功能扩展 → MCP** 展示当前实例的端点地址、传输方式、令牌 audience、支持的协议版本，以及按权限分组的完整 Tool / Resource 名单。破坏性操作标红；点击任意名称弹出详情，显示类型、所需 Scope 与完整描述——Tool 的 `description` 是写给模型看的提示词，平铺在面板上只会挤占版面。

该页面读取 `GET /api/mcp/manifest`（需要 `admin:token` scope，与访问令牌管理同权限），数据直接由 MCP 注册表推导——注册新的 Tool 或 Resource 会自动出现，不存在需要手工同步的第二份清单。端点地址按打开面板的来源拼出，所以显示的地址一定是可达的。

## 快速开始

### 1. 创建 MCP 专用 Access Token

在 Ech0 管理后台 **设置 → 访问令牌** 中创建一个新 Token：

- **Audience**：选择 `MCP (AI Agent)`（即 `mcp-remote`，MCP 专用 audience，区别于 `cli`、`integration` 等）
- **Scopes**：根据需要勾选（建议最小权限）
  - 只读场景：`echo:read`、`profile:read`
  - 读写场景：再加上 `echo:write`
  - 评论场景：`comment:read`（查看）、`comment:write`（发表）
  - 文件场景：`file:read`（查看）、`file:write`（删除 / 外部文件入库）
  - 互联场景：`connect:read`（查看连接列表与对端信息）、`connect:write`（添加/删除连接）
  - 管理场景：`admin:settings`（Webhook 管理等管理员操作）
- **有效期**：建议选择 8 小时或 1 个月（不建议永不过期）

创建后妥善保存 Token，它只会显示一次。

### 2. 配置 MCP Host

在你使用的 MCP 客户端配置中（具体文件名与入口因产品而异）添加远程服务，例如：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "https://your-ech0-instance.com/mcp",
      "headers": {
        "Authorization": "Bearer <your-access-token>"
      }
    }
  }
}
```

如果是本地开发环境：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "http://localhost:6277/mcp",
      "headers": {
        "Authorization": "Bearer <your-access-token>"
      }
    }
  }
}
```

## MCP Endpoint

- **地址**：`/mcp`（复用主服务端口，默认 6277）
- **协议**：MCP Streamable HTTP（JSON-RPC 2.0 over HTTP POST），同时支持 `2026-07-28` 与 `initialize` 握手时代的 `2025-11-25` / `2025-06-18` / `2025-03-26`
- **POST /mcp**：处理 JSON-RPC 请求（唯一入口）
- **GET / DELETE /mcp**：返回 405（服务端无会话，也不提供独立 SSE 流）
- **Origin 校验**：带 `Origin` 头的跨源请求会被拒绝（403），用于防御 DNS rebinding。同源请求始终放行；需要额外放行的浏览器来源用 `ECH0_WEB_CORS_ALLOWED_ORIGINS` 配置。原生 MCP Host / CLI 不发送 `Origin`，不受影响。

## 能力总览

当前 MCP 共暴露 **29 个 Tool**、**9 个 Resource** 与 **1 个 Resource Template**，按业务域整理如下。每个 Tool 都带 `annotations` 行为提示（`readOnlyHint` / `destructiveHint` / `idempotentHint` / `openWorldHint`），Host 可据此决定是否需要人工确认；返回 **JSON 对象**的 Tool 额外提供 `structuredContent`（数组结果不带，因为 2026-07-28 之前的修订版把该字段定义为对象，旧客户端遇到数组会直接报错——完整数据始终在文本块里）。

### Posts & Tags

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `search_posts` | 按关键词 / 标签 ID 搜索帖子，返回分页结果 `{items, total, page, page_size}` | `echo:read` |
| Tool | `get_post` | 按 UUID 获取单篇帖子（含内容、标签、点赞数、附件、扩展块） | `echo:read` |
| Tool | `get_today_posts` | 获取今日发布的帖子（支持 IANA 时区参数） | `echo:read` |
| Tool | `get_hot_posts` | 获取热门帖子（按点赞 + 评论数加权排序），可选 `limit`（默认 5，1–100） | `echo:read` |
| Tool | `get_random_post` | 随机返回一篇帖子（无帖子时返回 null） | `echo:read` |
| Tool | `get_on_this_day_posts` | 获取往年同月同日的帖子（"历史上的今天"，支持 IANA 时区参数） | `echo:read` |
| Tool | `list_tags` | 列出全部标签（id、名称、使用次数） | `echo:read` |
| Tool | `create_post` | 创建帖子；支持 `content`、`echo_files`、`layout`、`extension`，至少提供其一 | `echo:write` |
| Tool | `update_post` | 更新帖子；`echo_files` / `extension` 提供时为**全量替换** | `echo:write` |
| Tool | `delete_post` | 永久删除帖子 | `echo:write` |
| Tool | `like_post` | 帖子点赞数 +1 | `echo:write` |
| Tool | `delete_tag` | 删除标签并解除与所有帖子的关联 | `echo:write` |
| Resource | `ech0://posts/recent` | 最近 20 条帖子（可附 `?limit=N`） | `echo:read` |
| Resource Template | `ech0://posts/{id}` | 按 UUID 读取单篇帖子（通过 `resources/templates/list` 公布） | `echo:read` |
| Resource | `ech0://tags` | 全部标签及使用次数 | `echo:read` |
| Resource | `ech0://stats/heatmap` | 过去 30 个日历日每日发帖数（热力图，UTC 日界） | `echo:read` |
| Resource | `ech0://stats/visitors` | 过去 7 天每日访客统计 `{date, pv, uv}`（UTC 日界）；仅管理员 | `admin:settings` |

### Comments

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_comments` | 列出指定帖子下已通过的公开评论（与 `GET /api/comments` 等价） | `comment:read` |
| Tool | `create_comment` | 以集成/AI 身份发表评论（与 `create_integration_comment` 相同，推荐在 Agent 中使用此名称） | `comment:write` |
| Tool | `create_integration_comment` | 同上；与 `POST /api/comments/integration` 共用同一套校验与落库逻辑（无验证码、无 form_token） | `comment:write` |
| Resource | `ech0://comments/recent` | 全站最近 20 条公开评论 | `comment:read` |
| Resource | `ech0://guide/integration-comment` | 集成评论说明：REST 端点、Audience（`mcp-remote` / `integration`）、请求体、curl 示例、与本 MCP 会话 Token 的对应关系 | `comment:read` |

### Files

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_files` | 分页列出已上传文件元数据；返回的 `id` 可作为 `create_post.echo_files[].file_id` 引用 | `file:read` |
| Tool | `get_file` | 获取单个文件元信息（名称、URL、尺寸等）；`id` 可用于 `echo_files` 引用 | `file:read` |
| Tool | `delete_file` | 永久删除文件 | `file:write` |
| Tool | `create_external_file` | 用外部 URL 注册文件记录（无需上传）；返回含 `id` 的文件元信息，可直接用于 `echo_files` | `file:write` |
| Resource | `ech0://guide/file-upload` | 文件上传指南：REST 上传端点、参数、curl 示例、以及如何将上传结果用于 `create_post` | `file:read` |

> **提示**：本地文件上传通过 REST API（`POST /api/files/upload`，multipart/form-data）完成。AI Agent 可读取 `ech0://guide/file-upload` 获取完整操作指南。已有 URL 的外部文件可直接使用 `create_external_file` 注册。

### Connects（实例互联）

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_connects` | 列出本实例已保存的对端连接 | `connect:read` |
| Tool | `get_connects_info` | 聚合获取所有对端的公开信息（有 30 分钟缓存） | `connect:read` |
| Tool | `add_connect` | 添加远程 Ech0 实例连接 | `connect:write` |
| Tool | `delete_connect` | 删除已保存的连接 | `connect:write` |
| Resource | `ech0://connect/self` | 本实例公开信息卡片（名称、URL、logo、帖子统计、版本） | `connect:read` |

### Agent

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `get_recent` | AI 生成的站点近况摘要（有缓存，首次可能需数秒） | `echo:read` |

### Webhooks

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Tool | `list_webhooks` | 列出所有已配置的 Webhook（不含 secret） | `admin:settings` |
| Tool | `create_webhook` | 创建 Webhook 端点 | `admin:settings` |
| Tool | `update_webhook` | 更新 Webhook（按 id，全量替换） | `admin:settings` |
| Tool | `delete_webhook` | 删除 Webhook | `admin:settings` |
| Tool | `test_webhook` | 向 Webhook 端点发送测试请求 | `admin:settings` |

### User

| 类型 | 名称 | 说明 | Scope |
|------|------|------|-------|
| Resource | `ech0://profile/me` | 当前 Token 对应用户的资料（id、username、email、avatar、admin） | `profile:read` |

## 安全说明

- MCP 使用与 Ech0 API 相同的 JWT 鉴权体系，每个 Tool/Resource 都有独立的 Scope 校验。
- 请求限流：默认 20 RPS / 40 Burst（按 IP）。
- 请求体大小限制：256 KB。
- Tool 执行超时：10 秒。
- 建议在生产环境使用 HTTPS 并配合反向代理。
- Token 遵循最小权限原则：只读场景不要赋予 `echo:write`。

## 协议兼容

Ech0 的 `/mcp` **同时支持两代协议**，无需任何配置：

| 客户端 | 开场方式 | 服务端行为 |
|---|---|---|
| 2026-07-28（最新） | 直接发请求，自带 `_meta` | 无状态处理，结果带 `resultType` 与缓存提示 |
| 2025-11-25 / 2025-06-18 / 2025-03-26 | `initialize` 握手 | 正常握手并按该版本语义服务，结果不带 2026 专有字段 |

`server/discover` 与版本错误里都会列出全部支持版本：`["2026-07-28", "2025-11-25", "2025-06-18", "2025-03-26"]`。

- 支持方法（两代通用）：`tools/list`、`tools/call`、`resources/list`、`resources/templates/list`、`resources/read`
- 仅 2026-07-28：`server/discover`
- 仅旧版：`initialize`、`notifications/initialized`、`ping`
- 传输方式：Streamable HTTP。**不发放会话**（无 `Mcp-Session-Id`），`GET` / `DELETE /mcp` 一律返回 405；旧客户端收到 405 后会自动关闭独立 SSE 流，不影响使用。

### 2026-07-28 客户端的额外要求

新版是无状态协议：没有握手，每个请求都要自带协议元数据。客户端必须：

1. 在请求体 `params._meta` 中携带 `io.modelcontextprotocol/protocolVersion: "2026-07-28"` 与 `io.modelcontextprotocol/clientCapabilities`（缺任一项返回 HTTP 400 + `-32602`）；
2. 携带 `MCP-Protocol-Version: 2026-07-28` 请求头（必须与 body 一致，否则 HTTP 400 + `-32020`）；
3. 携带 `Mcp-Method` 请求头（与 body 的 `method` 一致）；
4. `tools/call` / `resources/read` 还需携带 `Mcp-Name` 请求头（与 `params.name` / `params.uri` 一致，非 ASCII 安全值用 `=?base64?…?=` 编码）。

使用官方 SDK 时以上头与 `_meta` 均由 SDK 自动处理，无需手工构造。

### 错误对照

| 场景 | HTTP | JSON-RPC code |
|---|---|---|
| 传输头缺失或与 body 不一致 | 400 | `-32020` |
| `_meta` 必填字段缺失 | 400 | `-32602` |
| 协议版本不受支持 | 400 | `-32022`（`data.supported` 列出全部支持版本） |
| 未知方法 | 404 | `-32601` |
| Token scope 不足 | 403 + `WWW-Authenticate: Bearer error="insufficient_scope", scope="…"` | `-40300` |
| Tool / Resource 不存在 | 200 | `-32602` |
| Tool 执行失败（含超时） | 200 | 结果 `isError: true`，供模型自行纠正 |

## 示例：使用 curl 测试

### 新版（2026-07-28）

```bash
META='"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{}}'

# Discover（查询服务器支持的协议版本与能力）
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2026-07-28" \
  -H "Mcp-Method: server/discover" \
  -d '{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{'"$META"'}}'

# List tools
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2026-07-28" \
  -H "Mcp-Method: tools/list" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{'"$META"'}}'

# Search posts
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2026-07-28" \
  -H "Mcp-Method: tools/call" \
  -H "Mcp-Name: search_posts" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_posts","arguments":{"query":"hello","page":1,"page_size":10},'"$META"'}}'

# Create a post
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2026-07-28" \
  -H "Mcp-Method: tools/call" \
  -H "Mcp-Name: create_post" \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"create_post","arguments":{"content":"Hello from MCP!","tags":["mcp","test"]},'"$META"'}}'
```

### 旧版（`initialize` 握手时代）

```bash
# 1. 握手：服务端回显你支持的版本（不支持的会协商到 2025-11-25）
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"my-host","version":"1.0"}}}'

# 2. 握手确认（返回 202，无 body）
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

# 3. 之后的请求只需带版本头，不需要 Mcp-Method / Mcp-Name / _meta
curl -X POST http://localhost:6277/mcp \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "MCP-Protocol-Version: 2025-11-25" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
```
