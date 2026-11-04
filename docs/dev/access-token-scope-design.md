# Access Token 权限粒度改造设计（Ech0）

日期：2026-03-23  
状态：已落地（scope/audience 词汇与判定在 `internal/authz`，令牌铸造在 `internal/access`，JWT 在 `internal/util/jwt`，HTTP 校验在 `internal/middleware`；本文档保留设计背景与决策记录）  
范围：后端鉴权与令牌管理（不含完整 OAuth2 授权服务器）

## 1. 背景与目标

当前 Ech0 的 access token 与登录 JWT 基本等价，核心 claims 仅包含 `user_id` 和 `username`。这使得 token 在服务端的授权判断中会继承账号角色能力（尤其是 admin），在“不受信任客户端”场景下风险偏高。

本次改造目标：

- 将 access token 从“身份票据”改为“能力票据（capability token）”。
- 引入固定白名单 scope，实现最小权限授权（least privilege）。
- 保持 Ech0 轻量定位，不引入完整 OAuth2 Server 复杂度。
- 旧 JWT 不兼容（含旧 access token 与旧 session token），直接失效并重新签发。

非目标：

- 不在本阶段实现第三方 OAuth2 授权码流程（Auth Code + PKCE）。
- 不重做现有 session 登录体系。

## 2. 设计原则

- **默认拒绝**：未声明 scope 的操作默认不允许。
- **固定词表**：scope 使用白名单枚举，不允许自由字符串。
- **高危隔离**：管理能力与内容能力严格分离（`admin:*` 与非 `admin:*`）。
- **最小侵入**：复用现有 `viewer` 与路由分组结构，避免大规模重构。
- **可观测可审计**：围绕 `jti` 记录授权结果与拒绝原因。

补充边界定义：

- 本文“默认拒绝”仅针对 `typ=access` 的 API token 生效。
- `typ=session`（站内登录态）在本阶段保持现有授权模型；升级后旧 session token 需重新登录获取新 token。

## 3. 权限模型

### 3.1 Token 类型

统一通过 claims 字段 `typ` 区分：

- `session`：站内登录态 token（延续现状）
- `access`：用于 API 调用的能力型 token（本次改造重点）

授权规则（唯一事实来源）：

- 当 `typ=session` 时：不进入 `RequireScopes`，继续使用现有业务授权逻辑（如 `user.IsAdmin`）。
- 当 `typ=access` 时：必须经过 `RequireScopes`；缺少 scope 一律 `403`。
- 当 token 缺失 `typ` 或 `typ` 不在允许集合（`session`、`access`）时：直接 `401`。

上线前置要求：

- 新签发登录 token 必须写入 `typ=session`；
- 新签发 API token 必须写入 `typ=access`；
- 旧 JWT（含旧 access token 与旧 session token）在切换后全部失效，不提供兼容窗口。

### 3.2 Scope 词表

Scope 按**信任边界**拆分，而不是按 CRUD 动词无限拆分。`write` 只覆盖创建/更新；删除是不可逆能力，单独授予。

内容与账号：

- `echo:read` / `echo:write` / `echo:delete` / `echo:private`
- `comment:read` / `comment:write` / `comment:review` / `comment:delete` / `comment:settings`
- `file:read` / `file:write` / `file:delete`
- `connect:read` / `connect:write` / `connect:delete`
- `profile:read` / `profile:write`

管理域（按爆炸半径拆分，互不隐含）：

- `admin:site`：站点品牌、CORS、系统设置
- `admin:storage`：S3 凭据与连接测试
- `admin:auth`：OAuth2 / Passkey 策略
- `admin:webhook`
- `admin:agent`：Agent / Embedding / Copilot
- `admin:data`：迁移、导入导出、快照计划
- `admin:ops`：日志、更新检查、访问统计
- `admin:user`：用户列表、删除、管理员切换

MCP：

- `mcp:read`：只读 MCP manifest，不授予任何 Tool 调用能力。

约束：

- `echo:private` 是 `echo:read` 的附加能力；`echo:read` 默认只返回公开内容。
- `echo:write` / `file:write` / `connect:write` **不包含**对应的 `:delete`。
- `comment:review` 覆盖审核面板的读取、状态/置顶与批量审核（approve/reject）；批量删除在 `comment:review` 之外还要求 `comment:delete`（服务内按 action 二次校验）。评论设置与测试邮件需要 `comment:settings`。
- Passkey 注册/删除、OAuth 绑定、访问令牌 CRUD 是 **session-only**，不进入 scope 词表；任何 `typ=access` token 即使持有全部 scope 也不能调用。
- 用 access token 铸造新 token 时，新 token 的 scopes 必须是调用者 scopes 的子集；该检查是纵深防御，正常入口已被 session-only 限制。
- 服务端始终以 scope 判定 access token 是否允许调用，即便用户角色是 admin。
- 浏览器 session 不是 scope 通配符：`admin:*` 接口额外要求 session 的 `is_admin` claim，非管理员会话返回 403；系统日志 SSE/WS 同样拒绝普通会话。

### 3.3 Audience 与时效

- `aud`：声明 token 使用场景（如 `public-client`、`cli`、`integration`）。
- `exp`：默认短时效（建议 24h）；长期 token 需显式选择。
- `jti`：每个 token 唯一 ID，用于审计、吊销和定位风险。

claim 来源统一策略：

- `aud` 与 `jti` 仅使用 `jwt.RegisteredClaims` 中的 `Audience` 与 `ID`。
- 禁止在自定义 claims 中重复定义同名字段，避免解析歧义与双写冲突。

## 4. Claims 与数据结构改造

### 4.1 JWT Claims（`internal/model/auth`）

在现有 `MyClaims` 上新增：

- `Type string   json:"typ"`
- `Scopes []string json:"scope"`

说明：

- `aud` 与 `jti` 使用 `RegisteredClaims.Audience` 和 `RegisteredClaims.ID`，不新增重复字段。
- `scope` 使用数组而非空格字符串，减少解析歧义。

### 4.2 访问令牌模型（`internal/model/setting.AccessTokenSetting`）

建议新增字段：

- `TokenType string`（默认 `access`）
- `Scopes string`（JSON 序列化存储）
- `Audience string`（单值或 JSON，首版可单值）
- `JTI string`（唯一索引）
- `LastUsedAt *time.Time`

### 4.3 创建 DTO（`AccessTokenSettingDto`）

新增请求字段：

- `Scopes []string`
- `Audience string`

并在服务层校验：

- scope 全部在白名单内；
- `admin:*` scope 仅高权限用户可签发；
- audience 必须在允许列表中。

## 5. 鉴权流程与路由边界

### 5.1 中间件拆分

建议在现有 `JWTAuthMiddleware` 基础上拆分职责：

1. `ParseTokenMiddleware`
   - 负责 token 提取、验签、过期检查、解析 `typ/scope/aud/jti`；
   - 将 token 元信息注入 `viewer` 扩展上下文。

2. `RequireScopes(scopes ...string)`
   - 仅负责授权判定；
   - 缺失权限统一返回 `403`（`NO_PERMISSION_DENIED` 语义可沿用并补充错误码）。

传输约束（针对不受信任客户端）：

- 高权限 scope（任意 `admin:*`）禁止通过 query 参数传 token，仅允许 `Authorization: Bearer`。
- query token 仅用于受限场景（如媒体直链），且仅接受低风险读权限（如 `file:read`）。
- 违反传输约束时返回 `403`，并记录审计原因 `token_transport_forbidden`。

### 5.2 路由应用方式（`internal/router/*`）

- 保留 `PublicRouterGroup` 与 `AuthRouterGroup`。
- 在 `AuthRouterGroup` 内按资源挂 `RequireScopes`：
  - 写接口：例如 `POST /echo` 需要 `echo:write`
  - 删除接口：例如 `DELETE /echo/{id}` 需要 `echo:delete`
  - 管理接口：按域使用 `admin:site` / `admin:storage` / `admin:auth` / `admin:webhook` / `admin:agent` / `admin:data` / `admin:ops` / `admin:user`
  - 账号安全与令牌管理：使用 `RequireSession`，不挂 scope
- 匿名放行接口维持现有策略（首页分页、今日内容、公开详情等）。

### 5.3 Service 层策略

- 业务角色（`user.IsAdmin`）继续保留，用于业务规则本身。
- API 调用授权以 scope 为先：token scope 不足时必须拒绝，即使用户是 admin。
- 私密 Echo 的可见性在服务层统一判定：session 沿用 admin 规则；access token 必须同时具备 `echo:read` 场景和 `echo:private`。`QueryEchos`、`GetEchoById`、`LikeEcho`、今日/热门/随机/历史上的今天共用同一 helper，避免各入口漂移。

viewer 扩展约定：

- 在 `pkg/viewer` 中扩展 token 元信息读取能力（至少包含 `typ`、`scopes`、`jti`、`audience`）。
- service 层从 `viewer` 读取授权上下文，不直接反复解析 JWT。

## 6. 错误语义与可观测性

- `401 Unauthorized`：token 缺失/格式错误/签名错误/过期。
- `403 Forbidden`：token 有效但 scope 或 audience 不满足。
- 例外：`/mcp` 资源端点在 audience 不匹配时返回 `401` + `WWW-Authenticate: Bearer error="invalid_token", resource_metadata="…"`。这不是“登录失效”，而是当前 token 不能用于该资源；MCP Host 只有看到 401 挑战才会重新走授权流程。

审计日志（脱敏）建议记录：

- `jti`
- `user_id`
- `route + method`
- `decision`（allow/deny）
- `reason`（scope_missing、aud_mismatch、expired 等）

错误码建议（与 HTTP 状态组合）：

- `401`：复用现有 token 缺失/无效/解析失败错误码。
- `403`：新增或明确区分以下错误码：
  - `ErrCodeScopeForbidden`
  - `ErrCodeAudienceForbidden`
  - `ErrCodeTokenTransportForbidden`

## 7. 切换方案（旧 token 不兼容）

一次性切换：

1. **结构升级**
   - 为 `access_token_settings` 增列：`token_type`、`scopes`、`audience`、`jti`、`last_used_at`；
   - 给 `jti` 建唯一索引。

2. **强制失效**
   - 发布切换版本时，服务端拒绝所有“无 `typ` 或 `typ` 非法”的旧 JWT（含旧 access 与旧 session）；
   - 管理端仅支持创建新格式 token（必须带 scopes + audience + `typ=access`）。

3. **重新签发**
   - 管理后台提示“旧 token 已失效，请重新创建”；
   - 由管理员按最小权限重新生成并分发 token。
   - Scope 词表切换不保留 legacy 别名：旧的 `admin:settings`、`admin:token`、`comment:moderate` 不再有效；旧的 `*:write` token 也不再隐含 `*:delete`。

4. **回滚策略**
   - 回滚到旧版本可恢复旧鉴权行为；
   - 新版本内不保留 legacy 分支，避免长期双轨维护。

## 8. 测试策略

### 8.1 单元测试

- `internal/util/jwt`：覆盖 `typ/scope/aud/jti/exp` 的签发与解析。
- `internal/middleware`（`auth.go`）：覆盖 401/403 分支。
- `internal/service/setting/access_token_service`：覆盖 scope 白名单与高危 scope 拒绝逻辑。

### 8.2 集成测试

建立最小权限矩阵：

- 只读 token：可读不可写；
- 内容写 token：不可访问管理接口；
- 管理 token：仅可访问授权的 `admin:*` 路径；
- audience 不匹配：统一拒绝。

落地方式：

- 基于 `httptest` + `gin.Engine`，复用现有 `internal/router/router_test.go` 风格。
- 测试 token 由测试辅助方法统一签发，避免手写 JWT 字符串。
- 至少覆盖 `echo`、`setting`、`file` 三类代表路径。

### 8.3 回归测试

- 对新签发的 `typ=session` token，登录/鉴权代码路径与升级前一致；不含合法 `typ` 的旧 session token 应返回 `401` 并引导重新登录；
- 匿名公开接口行为不变；
- 现有基础功能（发布、评论、文件、设置）在授权正确时行为不变。
- 旧 access token 在切换后稳定返回 `401`，且新 token 行为符合 scope 约束。

## 9. 交付拆分建议

- **里程碑 A**：Claims + 数据模型 + 新建 token API 入参校验
- **里程碑 B**：中间件扩展 + 路由挂载 `RequireScopes`
- **里程碑 C**：旧 token 强制失效 + 管理端重建提示
- **里程碑 D**：测试补齐 + 文档更新（README / API 文档）

## 10. 风险与应对

- 风险：scope 词表过粗导致“看似细粒度，实则过权”  
  应对：按信任边界拆分（`admin:*` 分域、写与删分离、私密读取独立），并用 MCP `delete_*` 工具测试和 REST 删除路由测试防止回退。

- 风险：切换当日第三方集成全部失效  
  应对：发布公告 + 版本说明 + 管理后台显式提示“先创建新 token 再升级”。

- 风险：授权逻辑分散导致漏检  
  应对：统一使用 `RequireScopes`，减少 service 内手写鉴权分支。

- 风险：query token 在不受信任环境泄露  
  应对：限制 query token 仅可用于低风险读场景，并禁用高危 scope 的 query 传递。

## 11. 结论

对 Ech0 当前“个人优先、轻量自托管、支持集成”的定位，推荐采用：

- 固定白名单 scope + audience + 短时效 access token；
- 与 session token 分型管理；
- 通过中间件统一授权与错误语义；
- 采用一次性切换，旧 JWT（含旧 access 与旧 session）直接失效并重建。

该方案在安全收益、实现复杂度和维护成本之间达到平衡，适合作为当前版本的权限粒度升级路径。
