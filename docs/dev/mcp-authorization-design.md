# MCP 授权服务器设计（OAuth 2.1 + CIMD）

状态：已落地（实现以 `internal/model/oauthserver`、`internal/service/oauthserver`、`internal/handler/oauthserver`、`internal/router/oauthserver.go` 为准）
关联：`docs/usage/mcp-usage.md`、`docs/dev/access-token-scope-design.md`、`internal/mcp/README.md`

## 1. 背景

Ech0 的 `/mcp` 原先只接受后台手工创建的 Access Token（audience `mcp-remote`）。这对脚本和 curl 够用，但标准 MCP Host 的“连接 → 浏览器授权 → 自动拿到 token”链路走不通：没有 Protected Resource Metadata，没有授权服务器，客户端只能让用户手工粘贴令牌。

MCP `2026-07-28` 把授权定义为 **OPTIONAL**，但一旦支持就必须按 OAuth 2.1 + RFC 9728 + RFC 8707 + RFC 9207 实现，并优先使用 Client ID Metadata Documents（CIMD）；Dynamic Client Registration（DCR）已 deprecated，仅作兼容保留。

本设计在 Ech0 内实现一个**最小但完整**的授权服务器，复用现有 Access Token / scope / audience 体系，不引入第二套令牌格式。

## 2. 角色与端点

### 2.0 代码边界

| 包 | 职责 | 约束 |
|---|---|---|
| `internal/authz` | scope / audience 词汇与 `Allows` 判定 | 纯策略，零 IO |
| `internal/access` | 令牌铸造与校验（策略 + JWT + 落库） | 被设置面板、device flow、OAuth 换 token 共用 |
| `internal/service/oauthserver` | 授权服务器协议：元数据、CIMD/DCR、code、token | 只依赖 `access` + `authz` |
| `internal/service/setting` | 设置聚合域；保留 `/api/access-tokens` CRUD | 通过 `access` 铸造令牌，不承载协议逻辑 |

```
MCP Host ──POST /mcp──────────────▶ 资源服务器（现有 mcp.Server）
   │  401 + WWW-Authenticate       │
   ▼                               │
GET /.well-known/oauth-protected-resource[/mcp]   (RFC 9728)
   │
   ▼
GET /.well-known/oauth-authorization-server       (RFC 8414)
   │
   ├── GET  /oauth/authorize        → SPA 同意页（/oauth/authorize）
   ├── GET  /api/consent            → 校验请求，返回客户端与 scope
   ├── POST /api/consent/approve    → 批准/拒绝，签发 code 或错误跳转
   ├── POST /oauth/token            → code + PKCE 换 access token
   └── POST /oauth/register         → DCR（兼容，deprecated）
```

- **资源服务器**是原有 `/mcp`，不改协议实现；只在 401 上补 `resource_metadata` 挑战，并接受 OAuth 签发的同款 JWT。
- **授权服务器**与资源服务器同进程，issuer 取请求 origin（含 `X-Forwarded-Proto`）。RFC 8414 要求 metadata 的 `issuer` 与构造 well-known URL 的标识完全一致，因此不用后台 `server_url` 覆盖；反代场景要保证 `Host` / `X-Forwarded-Proto` 正确透传。
- **同意页**是 SPA 路由，复用浏览器会话（`typ=session`）；consent API 拒绝任何 `typ=access` 令牌——否则一个令牌可以铸出另一个更宽的令牌。`GET /api/consent` 同时返回 `resource_name`，前端据此渲染资源名，不硬编码 MCP 文案；将来注册第二个资源时页面无需改动。

## 3. 令牌模型

不新增令牌格式。OAuth 换出的就是现有 Access Token：

| 字段 | 值 |
|---|---|
| `typ` | `access` |
| `aud` | `mcp-remote` |
| `scope` | 同意时批准的 scope 列表 |
| `jti` | 唯一，支持吊销 |
| `exp` | 30 天（`oauthserver.AccessTokenTTL`） |
| 名称 | `mcp:<client_name>`，出现在后台“访问令牌”列表里 |

因此撤销路径完全复用：后台删除令牌 → `TokenRevoker` 拉黑 `jti` → `/mcp` 立即 401。

## 4. 授权码与 PKCE

授权码是 **HS256 签名的无状态 JWT**（`typ=oauth_code`，60 秒），claims 绑定：

- `uid` / `username`：批准人；
- `client_id` / `client_name`；
- `redirect_uri`：精确匹配；loopback http 按 RFC 8252 忽略端口（path/query 仍须一致）；
- `scope`：批准范围；
- `resource`：`<issuer>/mcp`；
- `code_challenge`（S256）；
- `iss` / `aud` / `exp` / `jti`。

换 token 时逐项校验，并做单次消费（进程内 `jti` 账本）。PKCE 先验证再消费，避免验签失败把合法客户端的 code 提前作废。明文 `plain` 方法不接受。

`resource` 参数（RFC 8707）在授权与换 token 两处都校验，必须等于本次请求 MCP 端点；token 的 `aud` 固定为 `mcp-remote`，与现有 `RequireAudience` 对齐。

## 5. 客户端注册

优先级与 MCP 规范一致：

1. **CIMD**（`client_id` 是 HTTPS URL）：按需抓取 JSON 文档，校验
   - URL 必须 https、带 path、无 fragment/userinfo，并通过 `pkg/egress.Validate`（拒绝 loopback、私网/保留 IP、userinfo）；
   - 文档 `client_id` 与 URL 完全一致；
   - `redirect_uris` 非空，且 https 或 loopback http；CIMD 客户端的非 loopback redirect 必须与 `client_id` 同源（scheme+host）；
   - 只接受 public client（`token_endpoint_auth_method` 为空或 `none`）。
   使用 `pkg/egress` 的 Guard 客户端：抓取前 `egress.Validate`，拨号后检查远端 IP（防 DNS rebinding），重定向逐跳复验且禁止降级到 http；响应必须是 `application/json`（或 `+json`），最终 URL 必须仍是 https；64 KiB 响应上限，按 `Cache-Control: max-age` 缓存（1 分钟–1 小时）。
2. **DCR**（`POST /oauth/register`）：生成 `ech0-…` client_id，落库 `oauth_clients`；仅支持 `authorization_code` + `code` + `none`，上限 200 个，接口单独限流。规范已标记 deprecated，文档中明确说明仅为兼容。
3. **预注册**：暂不提供 UI；需要时可直接在 DCR 表插入一行，或后续补管理页。

## 6. 发现元数据

`/.well-known/oauth-protected-resource` 与 `/…/mcp` 返回同一份：

```json
{
  "resource": "https://ech0.example.com/mcp",
  "authorization_servers": ["https://ech0.example.com"],
  "scopes_supported": ["echo:read", "…"],
  "bearer_methods_supported": ["header"],
  "resource_name": "Ech0 MCP"
}
```

`/.well-known/oauth-authorization-server` 返回 RFC 8414 文档，声明 `client_id_metadata_document_supported: true`、`authorization_response_iss_parameter_supported: true`、`code_challenge_methods_supported: ["S256"]`，不使用 OIDC discovery（Ech0 不是 OIDC Provider）。

## 7. 安全决策

- **管理员专属**：Ech0 的手工令牌签发本来就只允许管理员；OAuth 同意继承同一规则。非管理员看到 `access_denied`。
- **redirect_uri 精确匹配**：不做前缀/通配匹配；`client_id` 与 `redirect_uri` 校验通过之后出现的授权错误（非管理员、用户不存在）改为 redirect-with-error，与拒绝路径一致；校验通过前仍返回本地错误，避免向未验证的 redirect 泄漏信息。
- **MRTR requestState**：HMAC 载荷包含 16 字节 nonce，同一秒内相同 principal/工具/参数的两个确认请求不会互相消费；服务端先读取 `inputResponses`，只有真正的批准才消费状态，未作答的重试不会烧掉批准。
- **无 refresh token**：v1 只签发 30 天 access token，不实现 refresh/rotation，元数据只声明 `authorization_code`。令牌到期后重新走一次同意；撤销走后台。
- **无客户端密钥**：只支持 public client + PKCE，避免在自托管场景引入密钥分发。
- **不信任客户端自述**：同意页展示的 `client_name` 来自 CIMD 文档或注册表，不取 query。
- **日志**：沿用 `mcp_request` 审计字段；授权服务器错误直接以 RFC 6749 `{error, error_description}` 返回，不套 Ech0 信封（consent 两个接口除外，便于前端复用统一请求层）。

## 8. 与 MCP 2026-07-28 最佳实践的关系

同批落地（见 `internal/mcp/README.md`）：

- `tools/list` / `resources/list` / `resources/templates/list` **按 token scope 过滤**，`cacheScope` 改为 `private`（列表随授权上下文变化，不能再声明 public）。
- 删除类 Tool（`delete_post` / `delete_tag` / `delete_file` / `delete_connect`）使用独立的 `*:delete` scope；`echo:write` 等写 scope 不再暴露它们。账号安全与访问令牌管理是 session-only，OAuth 同意页同样拒绝 `typ=access`。
- destructive 工具（注解 `destructiveHint=true`）在客户端声明 `elicitation` 能力时走 **MRTR 确认**：返回 `resultType: "input_required"` + `inputRequests["confirm"]` + HMAC `requestState`，重试时校验 principal / 工具 / 参数摘要 / TTL，拒绝换目标重放。
- `/mcp` 的 401 带 `WWW-Authenticate: Bearer resource_metadata="…"`，让客户端能自动进入发现流程。

## 9. 非目标

- 不做 OIDC Provider（不发 `id_token`，不支持 `openid` scope）。
- 不做 refresh token、token introspection、revocation endpoint（删除令牌即撤销）。
- 不做非管理员授权（与现有令牌模型一致）。
- 不支持 confidential client / client secret / `private_key_jwt`。
