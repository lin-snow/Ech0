# Ech0 架构重构 TODO

> 分支：`design/architecture-redesign`
> 状态：P1（Storage 合同与启动迁移）已实现并通过测试；P0/P2–P7 仍待开始。
> 排序原则：先冻结不可逆合同（URL / 事件 / 表），再改可逆代码（API / Service）。
> 每个阶段独立可合并；先加规则与测试，再改实现。

## 约束

- 单进程、单二进制、SQLite、CGO；不引入 broker / 外部服务。
- 不做：微服务、Casbin/OPA/OpenFGA、BaseRepository 泛型 CRUD、全量 vertical slice、proto。
- 稳定面（不能破坏）：`/mcp`、`/api/comments/integration`、`/api/files`、webhook JSON、CLI catalog。
- 所有新增 `.go/.ts/.vue` 带 SPDX；改 DI 跑 `just wire`；改路由跑 `just openapi`。

## P0 止血（独立 PR，可先合 main）

- [ ] access token 校验 `access_token_settings` 行存在；不存在即 401。
- [ ] `/api/comments/integration` 校验 token kind；session 拒绝。
- [ ] `?token=` 只允许只读 scope（`*:read`、`profile:read`）。
- 验收：三条回归测试；现有登录 / 集成流程不回归。

## P1 Storage 合同（最优先）

Why：对象 key / URL 已进入用户 markdown，是最不可逆的表面；同时修私有附件公开。

- [x] `docs/dev/storage-contract.md`：四类对象、`Ref`、`Visibility`、`ResolveURL`、public/private 规则、启动迁移与 URL 重写。
- [x] 失败测试先行：`TestPublicPathPrefixesOnlyPublic`、`TestRegisterStaticFilesServesOnlyPublicNamespace`、`TestUploadedTempFileIsNotPubliclyAddressable`；`storage_contract` tag 下的目标 API 规格。
- [x] `storage.Ref{Class, StorageType, Provider, Bucket, Key}`；与 `File` 唯一索引一致。
- [x] `Visibility`：`Public | Private`；public 只走 `/api/files/public/*`，content/temp 走认证 stream。
- [x] `ResolveURL(ref, visibility)`：managed 对象统一返回本站路径（content → stream，instance/avatar → public），不再依赖当前 provider；`external` 原样返回。
- [x] `File.URL` 降级为派生 / 外部地址；删除 `AfterFind` 全局 `resolveURL`。
- [ ] `external` 从 `storageType` 拆成独立 kind / 字段（P5 或后续；当前仍用 `storage_type=external`）。
- [x] Asset class：`Instance` / `Avatar` / `Content` / `Temp`；路径 `public/instance`、`public/avatars`、`content/`、`temp/`。
- [x] `File` 增 `OwnerKind` / `OwnerID` / `Class` / `ExpiresAt`；`users.avatar_file_id`、`system_settings.server_logo_file_id` 取代 URL 字符串。
- [x] 合同测试：公开命名空间只含 `public/`；私有 key 公开路由 404；private echo 的 content 登录可读、匿名拒绝；迁移幂等；头像/logo 优先于正文引用。
- 已定（见 storage-contract §5/§7/§8）：
  - Visibility 按引用推导：任一公开 echo 引用即 public；不存可漂移的列。
  - 私有媒体走认证 stream（`Authorization`），不用 `?token=`，不发 presigned URL。
  - S3 private content 全部经 Ech0 代理；public 才允许直链/CDN。
  - 旧对象在启动时一次性搬迁 + URL 重写，阻塞到 listener 之前。
- 已定并实现：temp 预览走 `GET /api/files/temp/:fileID`；旧 `/api/files/images/*` 与 `/api/file/*` 直接 404，不做 302；迁移前只要求文档提示手动 snapshot。
- 验收：`go test ./...` 绿（3300+）；OpenAPI 已重生成；SQLite 备份后可回滚（迁移失败阻止启动，旧对象保留）。
- 实现偏差：content 的 owner 是 echo（`owner_kind=echo`），不是 user；`files.expires_at` 取代 `temp_files`；`user_id` 保留 nullable 作为 orphan 回退。

## P2 Event 合同

- [ ] `docs/dev/event-contract.md`：事件目录、版本、投递语义、事务边界。
- [ ] `event.Catalog`：name / version / payload / ordering / delivery，唯一注册点。
- [ ] `TestCatalogComplete`、`TestEventGoldenJSON`、`reserved` 字段清单。
- [ ] 决策并实现 outbox：webhook / snapshot 同事务写事件行，提交后派发；缓存失效保持 best-effort。
- 验收：13 个事件全部登记；golden diff 进 CI；回滚事务不发 webhook。

## P3 State ownership + Migration registry

- [ ] `docs/dev/state-ownership.md`：22 张表 + 5 类对象（附件 / 头像 / logo / 快照 / 临时）→ 唯一 owner。
- [ ] `migration.Registry`：模块注册 models + steps；先包一个 `legacy` 模块，行为不变。
- [ ] `echo_file` 归 file；`EchoExtension` 归 echo；embedding 只引用 ID。
- [ ] 删除 `repository/common`；`repository/job` 只经 job `api.go`。
- [ ] arch test：禁止跨模块 model import；repo 只能被同模块 service import。
- 验收：迁移后的 schema 与今天一致；新模块加表不改中央列表。

## P4 API 声明

- [ ] `httpapi.API` + `Register` / `Seal`：每个 op 必须声明 posture 与 audience。
- [ ] 107 个 Huma op 迁入；raw gin（SSE/WS/upload/download/captcha/MCP）纳入同一注册表。
- [ ] 重复 method + path 启动期失败；catalog 测试。
- [ ] `oasdiff breaking` 进 CI。
- 验收：无游离路由；新增 op 不声明 posture 无法启动。

## P5 Authz 统一

- [x] `authz.Decide(principal, action, facts)`；API 准入 + service 资源判断共用词表。
- [ ] `authz.Filter(principal, action)` 产出 repo 谓词（替代散落 `PrivateScope`）。
- [x] 去掉 `RequireAdminSession` 的 access-token 豁免；admin 能力由 role × scope 共同决定。
- [ ] 策略矩阵测试：role × kind × audience × ownership × visibility × token-row-exists。
- [ ] 审计日志（append-only）：allow/deny + action + resource + principal。
- 验收：session 与 access token 对同一 action 结果一致；降级 admin 立即失效。

## P6 Service 边界

- [ ] 每模块 `api.go`；跨模块只 import `api.go`。
- [ ] `usecase`：`Initialize`、`RemoveUser` 收编跨模块事务。
- [ ] `echo→file`、`copilot→echo/embedding` 走端口；`setting→file`、`user→file`、`init→*`、`comment/connect/migrator→common` 消除。
- [ ] `service/common.GenerateRSS(*gin.Context)` 改中性请求。
- [ ] 删除 `service/common`。
- 验收：arch test 只剩 2 条合法同层边；`go list` 无环。

## P7 前端

- [ ] 能力驱动 UI：后端返回 capabilities，前端不再散落 `is_admin`。
- [ ] 私有 Echo 文案改“仅登录可见”。
- [ ] 可选：OpenAPI → TS 类型生成。

## 验收总表

| 阶段 | 可观测结果 |
|---|---|
| P0 | 3 个安全回归测试 |
| P1 | 私有附件不可公开访问；provider 切换不破旧 URL |
| P2 | 事件目录完整；webhook 可版本化 |
| P3 | 表有唯一 owner；迁移按模块注册 |
| P4 | 路由不可游离；OpenAPI breaking 进 CI |
| P5 | 一套 authz 决策；审计可查 |
| P6 | 8 条横调降到 2 条合法边 |
| P7 | UI 不再判断角色 |

## 非目标

- 不改 SQLite/GORM 技术栈；不换 VireFS/Busen；不引入 proto；不做多租户。
