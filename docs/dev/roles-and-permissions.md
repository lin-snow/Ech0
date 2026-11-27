# 角色与权限

> **本文只描述当前已实现的行为。** 身份与空间的建模（三层分离、秩、不变量与迁移）见 [space-identity-design.md](./space-identity-design.md)；目标权限模型（session 由角色派生 scope、端点四维声明）见 [permission-scope-design.md](./permission-scope-design.md)。两者冲突时，以设计文档为目标、以本文为现状。

Ech0 的账号有三种角色，另有未登录访客。角色存在 `users.role`，是唯一事实来源，**不进任何令牌**：每个请求解析 principal 时与 `status` 一起从库里读出来，降级、禁用、转让立即生效。

| 角色 | 判定 | 说明 |
|------|------|------|
| owner（所有者） | `role = 'owner'` | 空间的所有者，全库唯一（部分唯一索引 `ux_users_single_owner`）。除实例管理外，还能任免 admin / member、转让所有权 |
| admin（管理员） | `role = 'admin'` | 实例运营者，可以修改实例层面的任何配置；不能管理同级 admin，也不能触碰 owner |
| member（成员） | `role = 'member'` | 正常账号，拥有自己的内容，但不碰实例配置 |
| 访客 | 无会话 | 只能读公开内容，评论走匿名流程 |

初始化创建的第一个账号是 owner（`UserService.InitOwner`），自助注册一律是 member，注册还受账号数上限约束（`authModel.MAX_USER_COUNT`，见 `internal/service/user/user.go`）。角色只表示"能不能动实例"，不表示"能不能写内容"——登录即可发布。

## 管理关系：秩

`member < admin < owner`。所有答案来自同一个纯函数 `internal/authz.Authorize(subject, action, object)`，服务层不再手写"取当前用户 → 比较 bool"的守卫。账号管理动作的规则是"操作者秩严格高于目标"，而不是"是不是管理员"：

- admin 可以禁用、删除、踢下线、吊销令牌于 member；对另一个 admin 一律拒绝（平级互管会锁死实例）。
- owner 可以管理 admin 与 member；owner 自己不能被任何人禁用、删除或降级。
- 角色任免只有 owner 能做，且只能授予 admin / member（`PUT /user/role/{id}`）。给自己或 owner 行改角色会被拒绝。
- 转让所有权是唯一不由秩规则派生的动作：owner 指定一名 ACTIVE 的 admin 为继任者（`PUT /user/owner/{id}`），自己降为 admin。仓储层在单个事务里先降后升，顺序不可颠倒（部分唯一索引逐行检查）。

因为 owner 必在 admin 之上且不可删除，admin 集合永不为空，所以旧的 `LAST_ADMIN_REQUIRED` 守卫已删除。

## 决策层：internal/authz

```go
authz.Decide(viewer.MustFromContext(ctx), authz.EchoDelete, authz.ForEcho(echo))
```

- **Principal** 就是 `viewer.Context`（`Kind` / `UserID` / `Role` / `Scopes` / `Audience`）；匿名是 `KindAnonymous` 的值，不是 nil。**Action** 是编译期枚举（16 个，`internal/authz/action.go`），**Resource** 带资源类型、ID、归属者与目标角色（`ForEcho` / `ForFile` / `ForUser` / `Instance`）。
- 规则是一个 `switch`，没有策略文件、没有数据库存储、没有 `Authorizer` 接口。未知动作或结构不合理的对象返回 `ErrNoPolicy`——fail closed，测试逐一枚举动作兜底。
- 拒绝是 `authz.ErrDenied`（`apperr.ErrDenied` 这一个哨兵）；`internal/httperr` 把它映射成 403 problem 文档，服务层只负责返回错误。
- 判定不查库：角色来自中间件解析的 principal，归属者来自已加载的资源行。为此 `stream`、文件读取等路径把授权放在加载资源之后、写任何字节之前。
- 少数入口必须按"数据库里的目标账号"决定（转让所有权、OAuth 同意页），它们先加载该行，再用 `viewer.NewSession(user.ID, user.Role, …)` 构造 principal——授权对象就是被加载的那个账号。

## 准入门禁：Endpoints

Huma 路由的每一处声明都是一个 `endpoint`（`internal/router/admission.go`，不可导出，只能由构造函数产生）：

- 四种准入，一望即知：`public(reason)`（无凭据）/ `optional()`（匿名可进，有凭据则附带 principal）/ `session()`（仅交互会话）/ `scoped(scopes...)`（需认证且持有全部 scope）。`Access` 是显式字段而不是从 scope 反推——"仅会话"不是一种能力，公开端点也没有 scope 却必须说明理由。
- `Reason` 只属于公开端点，启动时必填；注册时它同时写进 `op.Description`，所以"为什么公开"会出现在 API 文档里，而不只是代码里的一句注释。
- 每个字段只属于一种准入：只有 `scoped` 能带 scope 与 audience，只有 `public` 能带 reason，其余组合在启动期直接报错。
- 管理门禁不靠 `admin:` 字符串前缀，而是由声明的 scope 反查 `authz.MinimumRole` 得到秩下限（`endpoint.rankGate`），再由中间件 `RequireMinRank` 执行。`admin:*` 之外的管理 scope（如 `comment:settings`）同样自动获得门禁。
- session scope 由角色派生（`authz.ScopesForRole`），`RequireScopes` 对两种凭据走同一条谓词，没有 session 旁路。
- `apiRegistry.verify()` 由 `SetupRouter` 在启动时调用（返回错误，Wire 因此拒绝构造 server）并在生成 OpenAPI 时再跑一次：重复路由、公开无理由、reason/scope/audience 用错准入类、未知 scope/audience、token 端点没有 scope、以及**词汇表里的管理 scope 从未被任何端点要求**。所有问题一次汇总返回，不会修一个重启一次。

## 内容与数据

| 能力 | member | admin / owner |
|------|--------|-------|
| 发布 echo | ✅ | ✅ |
| 编辑 / 删除自己的 echo | ✅ | ✅ |
| 编辑 / 删除他人的 echo | ❌ | ✅ |
| 查看公开 echo | ✅ | ✅ |
| 查看私密 echo（信息流与单条） | ✅ 所有登录用户共享，不按用户隔离 | ✅ 同左（匿名不可见） |
| 上传文件 / 记录外链文件 / 申请预签名 URL | ✅ | ✅ |
| 读取、修改、删除自己上传的文件 | ✅ | ✅ |
| 读取、修改、删除他人上传的文件 | ❌ | ✅ |
| 文件列表 / 目录树 / 按路径流式读取 | ❌ | ✅ |
| 给 echo 挂附件 | 文件须由本人或该 echo 作者上传 | 同左（历史无主文件放行） |
| 创建标签（显式接口，或写 echo 时按名字隐式新建） | ✅ | ✅ |
| 删除标签 | ❌ | ✅ |

## 评论

| 能力 | member | admin / owner |
|------|--------|-------|
| 发表评论 | ✅ 自动通过，来源记为 system，昵称取用户名、不带邮箱 | ✅ 同左 |
| 待审核列表 / 单条评论详情 | ❌ | ✅ |
| 审核、置顶、删除、批量操作 | ❌ | ✅ |
| 评论系统设置、发送测试邮件 | ❌ | ✅ |

私密 echo 的评论：匿名不能读、不能写；任何登录用户可读可写（与私密 echo 本身同一可见性）。成员的评论不进入审核队列；评论通知邮件只发给 owner（`GetOwner`）。评论删除属于审核动作，成员不能删除自己发出的评论。

## 实例与运营

| 能力 | member | admin | owner |
|------|--------|-------|-------|
| 个人资料 / 语言 / 头像 | ✅ | ✅ | ✅ |
| 实例设置（系统、S3、Webhook、Agent、Embedding、OAuth2、Passkey、快照计划） | ❌ | ✅ | ✅ |
| 访问令牌增删改查 | ❌ | ✅ | ✅ |
| 铸造带 `admin:*` scope 的访问令牌 | ❌ | ✅ | ✅ |
| 用户列表 / 删除用户 / 禁用用户 | ❌ | ✅（仅 member） | ✅ |
| 任免 admin / member | ❌ | ❌ | ✅ |
| 转让所有权 | ❌ | ❌ | ✅ |
| 数据迁移（快照导入 / 导出） | ❌ | ✅ | ✅ |
| 仪表盘与统计 | ❌ | ✅ | ✅ |
| Connect 配置 | ❌ | ✅ | ✅ |
| Copilot / Agent 对话 | ❌ | ✅ | ✅ |
| OAuth 授权批准、设备码批准 | ❌ | ✅（另需会话令牌） | ✅（同左） |

用户管理不再有"至少保留一位 admin"的兜底；取而代之的是秩规则：同级不可管理，owner 不可被管理。

## 访客

未登录只能读取公开 echo 和已通过审核的评论。评论必须填写昵称与邮箱，网址只接受 `http(s)`；`RequireApproval` 打开时评论进入待审核。验证码、频率限制、重复内容检查对访客生效。

## 边界与实现位置

- 私密内容的过滤在 repository 层按查看者作用域下推（`model.PrivateScope`）：匿名请求只返回公开 echo；`echo:private` scope 决定能否看到全部私密 echo——member 及以上的浏览器会话由角色派生出该 scope，access token 则必须在签发时声明。私密不提供账号间隔离。
- 内容类写操作（echo / 文件 / 评论）判"作者或管理员"；实例类操作一律 admin 及以上。新增接口时按此归类，不要给 member 打开实例写权限。
- 角色与访问令牌 scope 是两条独立的轴，在准入层汇合：
  - 端点的 scope 需求是声明（`Endpoint.Scopes`），管理端点的秩下限由 `authz.MinimumRole` 反查后交给 `RequireMinRank` 执行，不再有 `admin:` 前缀魔法，也不需要单独的 `RequireAdminSession`；
  - 会话的 scope 由角色派生（`authz.ScopesForRole`），因此"member 会话进不了管理端点"由同一条 `RequireScopes` 谓词保证；`access` token 不依赖角色，仍以签发的 scope 为准；
  - 访问令牌端点走 `authz` scope，其中 `admin:*` 只能由管理员铸造（`internal/access/service.go`），普通账号签发的令牌天然到不了管理面。
