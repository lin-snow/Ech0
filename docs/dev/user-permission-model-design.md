# 用户与权限模型（目标设计）

> **状态：设计稿，尚未实现。** 本文冻结产品规则，作为实现与评审依据；当前已实现行为见 [roles-and-permissions.md](./roles-and-permissions.md)。两者冲突时，以本文为目标、以该文为现状。

## 1. 定位

Ech0 不是"Member 与 Admin 权限对等"的多租户微博客，而是**以实例所有者为中心、允许多个注册用户参与内容表达的个人时间线系统**。

由此得到一条贯穿全文的原则：

> **内容表达可以多用户参与，内容治理始终由实例所有者负责。**

它同时解释了三个看上去"不对称"的规则：

- Member 拥有完整的内容创作能力，但不拥有任何平台治理能力；
- 评论属于平台公共交互，评论者不拥有对其评论的管理权；
- `ADMIN` 不只是普通角色，而是实例所有权的体现。

因此不引入通用 RBAC 引擎，权限模型固定为三个域（见 §3）。

## 2. 身份

| 身份 | 说明 | 判定 | 登录 |
|---|---|---|---|
| Anonymous | 未登录访客，不是数据库角色，而是一种未认证状态 | 无会话 | — |
| Member | 普通注册用户 | `users.is_admin = false` | 允许 |
| Admin | 实例所有者 / 平台管理员 | `users.is_admin = true` | 允许 |

身份产生规则（现状已实现）：

- 第一个注册的用户自动成为 Admin（`UserService.InitOwner`）；
- 后续自助注册一律为 Member，受实例账号数上限约束（`authModel.MAX_USER_COUNT = 5`）；
- Admin 可以把 Member 提升为 Admin，也可以降级其他 Admin，但系统至少保留一位 Admin（`LAST_ADMIN_REQUIRED`）。

Admin 允许多个，但系统至少保留一位（D2，见 §8）。

## 3. 权限域

| 域 | 覆盖身份 | 回答的问题 |
|---|---|---|
| Domain 1 — Content Access | Anonymous / Member / Admin | 谁能看到 Echo 与评论 |
| Domain 2 — Personal Ownership | Member / Admin | 谁能创建、编辑、删除**自己的** Echo，管理自己的 Profile 与附件 |
| Domain 3 — Instance Management | Admin only | 谁能管理全站内容、评论、附件、用户与实例配置 |

三条硬边界：

1. 内容类写操作只判"作者本人或 Admin"，不做角色继承；
2. 实例类操作一律 Admin，Member 永远不因"拥有内容"而获得实例写权限；
3. 角色位（`is_admin`）与访问令牌 scope 是两条独立轴，不可互相替代——scope 设计见 [access-token-scope-design.md](./access-token-scope-design.md)。

## 4. 权限矩阵

### 4.1 Echo

| 操作 | Anonymous | Member | Admin |
|---|---|---|---|
| 查看 Public Echo | ✅ | ✅ | ✅ |
| 查看私密 Echo | ❌ | ✅ 全部 | ✅ 全部 |
| 创建 Echo | ❌ | ✅ | ✅ |
| 编辑自己的 Echo | ❌ | ✅ | ✅ |
| 编辑他人的 Echo | ❌ | ❌ | ✅ |
| 删除自己的 Echo | ❌ | ✅ | ✅ |
| 删除他人的 Echo | ❌ | ❌ | ✅ |

**已冻结（D1）：私密 = 仅匿名不可见，不是用户之间的隐私隔离。** 任何登录用户（Member / Admin）都能在信息流和单条读取中看到全部私密 Echo；未登录访客看不到。私密既不加密，也不限制同实例的其他账号。

附带的权限细则：

- **写权限不随可见性扩大**：Member 能看到他人的私密 Echo，但仍不能编辑或删除它；只有作者本人或 Admin 可以。
- **访问令牌另受 scope 约束**：会话（`typ: session`）直接按"登录"放行；access token 仍需 `echo:private`，只有 `echo:read` 不能枚举私密内容。
- **命名风险**：语义已从"仅作者可见"变为"仅匿名不可见"，UI 继续显示"私密 / Private"会误导用户。实现时应把界面文案改为"仅登录可见"或附一行说明；API 字段 `private` 保留不改，避免破坏兼容性。

### 4.2 Comment

| 操作 | Anonymous | Member | Admin |
|---|---|---|---|
| 查看 Public Echo 的评论 | ✅ | ✅ | ✅ |
| 查看私密 Echo 的评论 | ❌ | ✅ | ✅ |
| 评论 Public Echo | ✅ | ✅ | ✅ |
| 评论私密 Echo | ❌ | ✅ | ✅ |
| 删除评论 | ❌ | ❌ | ✅ |
| 审核、置顶、批量操作 | ❌ | ❌ | ✅ |

两条冻结规则：

1. **匿名评论允许**（已实现）。访客必须填写昵称与邮箱，走验证码、频率限制、重复内容检查；`RequireApproval` 打开时进入待审核。登录用户的评论自动通过，昵称取用户名、不带邮箱（`SourceSystem`）。
2. **Member 不允许删除自己发布的评论**（已实现）。评论删除是审核动作，不是内容所有权动作；即使评论属于自己，也只能由 Admin 处理。

> 这条规则是产品选择而非技术约束：它让"评论"成为平台公共交互的一部分，而不是评论者可以随意收回的独立资源。

### 4.3 附件与文件

| 操作 | Anonymous | Member | Admin |
|---|---|---|---|
| 上传文件 / 记录外链 / 申请预签名 URL | ❌ | ✅ | ✅ |
| 读取、修改、删除自己上传的文件 | ❌ | ✅ | ✅ |
| 读取、修改、删除他人上传的文件 | ❌ | ❌ | ✅ |
| 文件列表 / 目录树 / 按路径流式读取 | ❌ | ❌ | ✅ |
| 给 Echo 挂附件 | ❌ | 文件须由本人或该 Echo 作者上传 | 同左 |
| 创建 / 删除标签 | ❌ | ❌ | ✅ |

附件权限从 Echo 所有权继承：Member 管得了自己 Echo 的附件，管不了别人的。

## 5. 账户状态

`users.status` 只有两个值，**不引入软删除、不引入 `deleted_at`**：

| 状态 | 含义 | 登录 | 内容操作 |
|---|---|---|---|
| `ACTIVE` | 正常账户 | 允许 | 正常使用 |
| `DISABLED` | 已禁用账户 | 禁止 | 禁止 |

禁用账户时：

- 禁止登录；**已有 Token 立即失效**（session / refresh / access 三种都要覆盖）；
- 禁止创建、编辑 Echo，发表评论，修改 Profile；
- 保留 User、Echo、Comment、File 等全部数据；
- 其已发布的公开内容仍然可读（禁用不等于隐藏）；
- Admin 可以恢复为 `ACTIVE`，恢复后数据与登录能力原样回来。

### 5.1 Token 失效的实现取向

现有吊销机制是**按 JTI 的内存黑名单**（`TokenRevoker`），只能吊销单个 token，无法按用户批量吊销，且重启即丢失。禁用账户需要按用户生效，推荐：

- **在认证中间件按 `user_id` 读取用户状态**，`DISABLED` 返回 `403 ACCOUNT_DISABLED`；读取走认证仓储的直读（与每请求读取 locale 同一量级），不加缓存，避免禁用后仍在 TTL 内放行；
- 登录、静默刷新、OAuth 回调、设备码批准入口同样校验状态；
- access token（含 MCP / CLI / 集成）走同一条中间件，天然覆盖；
- 不依赖"批量吊销 JTI"，因此重启、多实例都不会漏。

代价：每次带身份请求多一次按主键的单行查询。这比"给用户加 `token_version` 并写进 claim"更简单，也不需要在禁用时枚举 token，且重启、多实例都不会漏。

## 6. 账户生命周期

只有两种管理操作：

| 操作 | 结果 | 数据保留 | 可恢复 |
|---|---|---|---|
| 禁用 | 禁止账户使用 | 全部保留 | ✅ |
| 永久删除 | 删除用户及其关联数据 | ❌ | ❌ |

### 6.1 永久删除的数据范围

以 `user_id` 定位用户创建的所有 Echo，再以 `echo_id` 清理关联数据：

```
User
├── UserLocalAuth / UserExternalIdentity / WebAuthnCredential / AccessToken
├── File (user_id, 上传者)
│   └── TempFile
└── Echo (user_id)
    ├── EchoFile (echo_id)
    ├── EchoExtension (echo_id)
    ├── EchoTag / Tag.usage_count
    ├── Reaction (echo_id, actor = u:<userID>)
    └── Comment (echo_id)
Comment (user_id)   ← 用户发表在其他 Echo 下的评论
```

两类 Comment 必须同时处理：

1. **用户自己的 Echo 下的所有评论**：全部删除，包括其他用户发表的评论；
2. **用户在其他 Echo 下发表的评论**：也要删除，`comment.user_id = 被删用户`。

其他用户的 Echo 及附件不受影响。

删除动作的事务边界：

- **数据库**：单个事务内完成用户行、身份、令牌、Echo 及其从属行、评论、文件元数据的删除；
- **对象存储**：事务提交后再删物理对象（本地磁盘 / S3），失败只记录 warning 日志、不回滚数据库，也不阻塞删除结果；当前没有自动重试队列，残留对象需人工清理；
- **缓存与索引**：清理用户缓存、Echo 缓存、AI 摘要 / 向量索引，发布 `event.UserDeleted` 让订阅者失效相关缓存（当前订阅者仅清空 Agent 的最近摘要缓存，需补齐 Echo / 向量维度）。

### 6.2 删除确认

Admin 删除用户时必须二次确认，并明确告知影响范围。文案基线：

```
永久删除用户

此操作不可撤销，将永久删除该用户及其关联数据：
- 该用户创建的全部 Echo
- 这些 Echo 关联的所有文件
- 这些 Echo 下的全部评论（包括其他用户发表的评论）
- 该用户在其他 Echo 下发表的评论
其他用户创建的 Echo 不受影响。

确认后无法恢复。
```

前端需要求输入确认（例如输入用户名）而不是仅一个"确定"按钮；后端不做"软删除兜底"，删除即真实删除。

## 7. 现状与目标差距

| 项 | 当前实现 | 目标 | 影响 |
|---|---|---|---|
| `users.status` | 不存在 | `ACTIVE` / `DISABLED` | 需迁移 + 中间件校验 + 管理端 UI |
| 禁用后的 Token | 无按用户失效机制 | 三种 token 全部拒绝 | 中间件按 user_id 校验状态 |
| 永久删除范围 | 只删 `user` + `user_local_auth` | §6.1 全量级联 | 现有数据会留下孤儿 Echo / 评论 / 文件 |
| 匿名评论 | 已实现，昵称+邮箱+验证码+审核 | 保持一致 | 无需改动 |
| Member 删自己的评论 | 已禁止 | 保持一致 | 无需改动 |
| 私密 Echo 可见性 | per-user（`PrivateScope.ViewerID`） | 仅匿名不可见（D1） | 重写 scope、`privateScope()`、`canReadPrivate()` 与缓存键 |
| 私密 Echo 评论 | 一律拒绝（`assertEchoCommentable`） | 登录用户可评论 | 评论入口与评论列表可见性都改按"是否登录"判定 |
| Admin 数量 | 允许多个，至少保留一位 | 多 Admin，至少保留一位（D2） | 维持现状，不做唯一化改造 |
| 用户上传的文件 | 删除用户时不处理 | 按 §6.1 | File 行 + 存储对象 + TempFile |
| 反应 | 不处理 | 删除 `actor = u:<userID>` 的记录 | `internal/util/actor` 已定义 `u:` 前缀 |
| 用户自助删号 | 不存在 | 不提供（D4 默认） | Admin 删除是唯一注销路径 |

## 8. 决策状态

### 已冻结

**D1 — 私密 Echo = 仅匿名不可见。** 所有登录用户共享完整实例时间线；私密不提供账号间隔离，也不加密。实现要求：

- `model.PrivateScope` 从 `ViewerID` 维度简化为"匿名 / 已登录"两态（`All` 仍保留给服务端聚合）；
- `EchoService.privateScope()` 对会话直接放行，对 access token 仍要求 `echo:private`；
- `EchoService.canReadPrivate()` 的管理员破窗路径不再需要，单条读取对登录用户直接放行；
- 缓存键从 `u:<id>` 收敛为 `public` / `authed`，缓存命中面更大，必须补一条"匿名请求不得命中 authed 缓存"的测试；
- 评论入口 `assertEchoCommentable` 改为"未登录拒绝"，评论列表按同一 scope 过滤；
- UI 文案把"私密"改为"仅登录可见"（API 字段 `private` 不变）。

**D2 — 允许多个 Admin，至少保留一位。** 保留现有 `LAST_ADMIN_REQUIRED` 兜底，不引入唯一所有者、转让或首位管理员恢复流程。文档上区分"实例所有权"（产品语义）与"管理员权限"（实现能力），不新增约束。

### 默认取值（可直接推翻）

**D3 — 删除用户时其上传的文件。** 默认删除该用户上传的全部 `File` 行与存储对象，并解除其在存活 Echo 上的 `EchoFile` 关联；其他 Echo 若引用这些文件会失去附件。

**D4 — 不提供用户自助注销。** 永久删除只能由 Admin 发起，避免用户级联删除他人评论时缺少风险确认。

**D5 — 禁用后的前端行为：立即踢出。** 中间件对 `DISABLED` 返回 `403 ACCOUNT_DISABLED`；`web/src/service/request` 识别该错误码后清空内存 access token、写入 `needLoginRedirect` 并跳转登录页，刷新接口同时清 Cookie，因此旧会话无法续期。服务端不推送，用户在下一次请求时被踢出；这是"立即失效"在前端可观察的最早时机。

## 9. 实现落点（供后续拆分）

| 层 | 位置 | 需要做的事 |
|---|---|---|
| model | `internal/model/user/user.go` | 新增 `Status` 字段与常量；`UserDTO` 同步 |
| repository | `internal/repository/user` | 状态读写、按状态过滤登录、删除级联（跨域需 service 协调） |
| service | `internal/service/user` | 禁用 / 启用 / 删除的业务守卫与事务编排 |
| service | `internal/service/echo`、`comment`、`file` | 提供按 `user_id` / `echo_id` 批量清理的端口，供用户删除调用 |
| middleware | `internal/middleware/auth.go` | 解析 token 后校验用户状态（含 access token） |
| handler/router | `internal/handler/user`、`internal/router/user.go` | `PUT /user/status/{id}`（`ScopeAdminUser`）与 `DELETE /user/{id}`；OpenAPI 重新生成 |
| 前端 | `web/src/views/panel`、`web/src/components/common/BaseDialog.vue`、`web/src/service/request` | 状态列、禁用/启用、输入用户名才能确认的删除弹窗、`ACCOUNT_DISABLED` 强制退出 |
