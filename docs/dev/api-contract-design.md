# API 契约层重构设计

> 状态：**已实施**（2026-11）。本文档记录最终形态；不要按早期草稿的 `code`/`message_key`/`apierr` 实现。
> 决策（本次修订）：**不保留 v1、不设 v2、一次性硬切**。理由见 §1.3。
> 范围：HTTP JSON 契约层（响应形态 / 错误 / 门禁 / 代码生成 / 弃用）。**不含**分层架构、鉴权模型、MCP 协议语义、OAuth 协议实现。
> 关联：[architecture-overview.md](architecture-overview.md)、[i18n-contract.md](i18n-contract.md)、[huma-raw-gin-endpoints.md](huma-raw-gin-endpoints.md)、[cli-design.md](cli-design.md)、[access-token-scope-design.md](access-token-scope-design.md)

## 0. 结论

三句话：

1. **不引入 protobuf / gRPC / ConnectRPC / grpc-gateway。** 150 条路由里有 44 条（SSE、WebSocket、multipart、二进制下载、OAuth 重定向、MCP JSON-RPC、SPA shell）结构上无法进入 proto（清单见 [huma-raw-gin-endpoints.md](huma-raw-gin-endpoints.md)），强行引入的结果是三层定义并存。
2. **不复制 `google.rpc.Status` 的 JSON 形态**（`{"error":{code:int32,message,status,details}}`）。int32 code 与 HTTP 状态重复；`details` 是 `google.protobuf.Any`，JSON 里必须带 `@type` 才能解码；`ErrorInfo.metadata` 是 `map[string]string`，塞不进结构化参数（我们后来连 `message_params` 本身也删了：服务端不产出用户可见文案，见 §2.3）。
3. **采用 RFC 9457 `application/problem+json` 作为唯一错误形态，成功响应剥掉信封直接是资源本身。** Huma v2 原生就是这个模型（`huma.ErrorModel` 的注释即 RFC 9457，`ContentType` 自动把 `application/json` 换成 `application/problem+json`），因此我们**一个扩展字段都不加**：`Problem` 就是 `huma.ErrorModel`，只保留 `type` 给必须区分同状态的两个错误。

净效果是**删抽象**：`Result[T]`、`Envelope[T]`、`ErrorBody`、`apiError`、`Problem` 自定义类型、28 个 i18n key 全部消失，`handler/response.Response` 缩成 5 个构造器加一个适配器，handler 回到 Huma 的原始签名 `(*Out, error)`。

---

## 1. 标准评估（本次调研）

### 1.1 候选矩阵

| 候选 | 判定 | 一句话理由 |
|---|---|---|
| protobuf + gRPC + grpc-gateway | **不采纳** | 44/150 路由进不了 proto；gateway 不支持 `text/event-stream`；proto3 JSON 映射会改写现有 wire 类型（int64→string、Timestamp→RFC3339） |
| ConnectRPC（HTTP + gRPC 双栈） | **不采纳** | 能解决"不需要 mux"，但只覆盖纯 CRUD 子集，剩下 44 条照样裸写，等于两套定义 |
| Google AIP 整体 | **部分采纳** | AIP 是可执行的设计提案集合，不是标准；逐条取舍见 §1.2 |
| RFC 9457 Problem Details | **采纳** | 唯一 IETF 标准（Standards Track，2023-07 取代 RFC 7807）；Huma v2 已内置 |
| JSON:API | **不采纳** | `data/attributes/relationships` 三层 + 媒体类型协商，对单一 SPA 客户端是净成本 |
| OData | **不采纳** | 查询语言与元数据文档过重，无消费者 |
| OpenAPI-first（ogen / oapi-codegen） | **不采纳** | 与 Huma 同为"一份 Go 类型即 schema"，收益相同、迁移全量 |
| Huma 原生错误模型 | **采纳为基座** | 官方 *Custom Errors* 路径：`StatusError` + `ContentTypeFilter` |

### 1.2 AIP 逐条取舍

| AIP | 采纳 | 不采纳 | 理由 |
|---|---|---|---|
| **AIP-193** Errors | "状态码是粗粒度分类，细粒度 reason 只在客户端需要区分时出现" | `code:int32`、`details:Any`、`reason` 必填 | 我们只保留前半句：状态码即分类，不上 wire 任何错误码常量；`type` 替代 `reason`，且只用于确实要被客户端区分的那几处 |
| **AIP-158** Pagination | **page token 必须不透明**（当前 `v1:` 明文前缀违反此条，见 §2.2） | `page_size` / `page_token` / `next_page_token` 命名 | 管理表格需要页码跳转与 `total`，AIP-158 明确"不鼓励 `total_size`"。为改名而失去这两个能力是负收益 |
| **AIP-127** HTTP/JSON | HTTP 状态码承载成败、JSON 字段 camelCase | 通用 `HttpBody` 包装 | 与既有路由风格一致 |
| **AIP-136** Custom methods `:verb` | — | URL 里加 `:verb` | 纯 URL churn，推翻 106 个 `OperationID` 与 catalog 契约，零功能收益 |
| **AIP-122** Resource names | — | `publishers/{p}/books/{b}` 式层级 | Ech0 是单租户个人实例，没有资源层级 |
| **AIP-132** List | 集合字段名 `items` | 强制 `next_page_token` | 见 AIP-158 行 |

**一句话**：采纳 AIP 的**词汇与不变量**（错误码命名、token 不透明、集合字段名），不采纳 AIP 的**容器形状**（`google.rpc.Status`、`page_token` 命名），因为它服务于 Google 的 gRPC 转码与客户端库生态，Ech0 没有这些约束。

### 1.3 为什么硬切而不是双版本

旧方案（`/api/v2` 双挂 + v1 冻结弃用）的成本被低估了：106 个 operation 双挂 + 目录双维度 + 两套前端解包 + 弃用期两套心智，换来的是"第三方集成不受影响"。但实际上：

- 唯一的外部消费者是 `hub/`（`hub/src/services/echoApi.ts:83` 读 `r.code !== 1 || !r.data`）与可选的 access-token 集成。
- `hub/` 与本仓同一个 release 发布；它是可选的卫星站，不是协议端点。
- 双版本期结束后仍然要做一次同样的全量改造，只是延后。

所以：**一次性切换，不加版本维度**。代价见 §7 风险 1，收益是删除全部兼容分支，`_catalog` 也不需要"同一操作两个挂载点"的 `paths` 字段。`internal/openapi/catalog/catalog.go` 的 `Doc` 保持单层。

### 1.4 Go / Huma 社区惯例（3 条硬约束）

1. **handler 返回 `(*Out, error)`，成败不由响应体字段表达。** Huma 的 `Register` 签名即如此；`huma.StatusError` 是错误携带状态码的唯一通道。Go 标准库与 Mat Ryer 的 *How I write HTTP services in Go after 13 years* 也是同一取向：状态码在 `WriteHeader`，正文只放资源。
2. **不发明业务状态位。** `code: 1/0` 与 `google.rpc.Code`（OK=0）语义相反，是纯粹的认知负担；HTTP 状态码已经是成败信号。
3. **类型来自生成物，不手写平行 DTO。** `@hey-api/openapi-ts`（已在 `apps/cli` 运行，产出 89.5 KB `types.gen.ts`）直接复用到 `web/`，替代 15 个文件 922 行手写 `.d.ts`。

---

## 2. 目标契约

### 2.1 成功

| 场景 | 状态码 | body |
|---|---|---|
| 读取 | `200` | 资源对象本身 |
| 创建 | `201`（见 Q2） | 创建后的资源本身 |
| 更新 | `200` | 更新后的资源本身 |
| 删除 / 纯动作已完成 | `204` | 无 |
| 列表 | `200` | `{ "items": [...], "total": N, "nextCursor": "..." }` |

```json
HTTP 201
{ "id": "01J...", "content": "hello", "created_at": 1764000000 }
```

- 无 `code`、无 `msg`、无 `data` 包装。
- **成功文案不下发。** 服务端不再为"发布成功"这类提示承担 i18n（迁移前是 63 个消息常量 + 112 处 `commonModel.OK` 调用，外加 28 个 `MsgKey*` 与一个用文案反查 key 的脆弱映射；这些文件都已删除）。提示语属于 UI，由 `web/` 的 i18n 提供。
- 创建用 `201`：Huma 通过 `DefaultStatus` 声明，零额外代码。

### 2.2 列表与分页

两种机制并存，语义固定，不再新增第三种：

| 场景 | 参数 | 响应字段 | 理由 |
|---|---|---|---|
| 管理表格（设置、用户、评论审核、文件） | `page` / `pageSize` | `total` 必填，`nextCursor` 省略 | 需要页码跳转与总数 |
| 时间线（Echo、评论流） | `cursor` / `pageSize` | `total` 省略，`nextCursor` 可能为空 | keyset 分页，`total` 不可廉价计算 |

改动点：
- 合并 `commonModel.PageQueryResult[T]` 与 `commentModel.PageResult[T]` 为唯一的 `commonModel.List[T]{Items []T, Total *int64, NextCursor string}`（`Total` 是指针：游标模式不填，偏移模式才填，避免把「0」和「没统计」混为一谈）（后者目前是 `PageQueryResult[[]model.Echo]` 这种带切片的双泛型，`internal/service/echo/ports.go:17`）。
- `cursor` 去掉 `v1:` 明文前缀（`internal/model/echo/cursor.go`）。AIP-158 明确"base64 一个透明 token 不算充分混淆"；明文前缀等于邀请客户端解析，任何格式变更都会变成隐性 breaking change。编码保持 base64url，文档声明"客户端必须当作不透明串"。
- **本次不做全仓 JSON 命名统一。** 现有 wire 同时存在 `created_at`（domain 模型）与 `pageSize`（分页参数），统一任一侧都是六七百处前端字段访问的 churn，零语义收益。规则改为"新字段跟随所在结构体既有风格"：分页层 `nextCursor`（与 `pageSize` 同族）。problem 文档本身就是 RFC 9457 的字段名，没有可争的地方。

### 2.3 错误

`Content-Type: application/problem+json`，HTTP 状态码是权威：

```json
HTTP 403
{
  "type": "https://ech0.app/problems/account-disabled",
  "title": "Forbidden",
  "status": 403,
  "detail": "this account has been disabled by an administrator",
  "errors": [
    { "location": "body.question", "message": "required property is missing" }
  ]
}
```

| 字段 | 来源 | 必需 | 说明 |
|---|---|---|---|
| `type` | URI | 否 | 省略即 `about:blank`。只在客户端必须区分**同一状态**的两个错误时出现：账号禁用（403）与设备流的四种状态（400） |
| `title` | `http.StatusText(status)` | 是 | 稳定、不本地化（RFC 9457：同一 `type` 的 title 不应变化） |
| `status` | HTTP 状态码 | 是 | 冗余但被 RFC 9457 要求 |
| `detail` | 服务端英文文案 | 是 | 给开发者、日志、curl。**客户端禁止解析它**，见 §4.1 |
| `errors[]` | Huma 原生 `ErrorDetail` | 否 | `location` 是机读指针（`body.x` / `query.y` / `header.Z`），`message` 是英文调试文本 |

三个关键取舍：

- **不发 `code`。** 行为由 `status` 表达（401 重登、403 无权限、404 不存在、409 冲突、412 前置条件、429 冷却、5xx 服务端故障），位置由 `errors[].location` 表达，剩下极少数需要区分的语义由 `type` 表达。再发一个 UPPER_SNAKE 字符串是同一信息的第三次编码，只会多出一份必须与状态码保持同步的词汇表。
- **不发 `message_key` / `params`。** 服务端不产出用户可见文案，也就没有译文可指向。`detail` 是英文，客户端按 `status` 选自己的 i18n 文案。
- **状态码只活在进程内的 `StatusError`**（`internal/model/common/error.go`）。`Fail(status, msg)` / `FailWithType(status, type, msg)` / `WithStatus(status, err)` 构造它，`httperr.From` 把它投影成 problem 文档；没有第二套码表，也没有需要与状态码同步的词汇表。

早期实现里的 `ErrCode*` 常量表（`INVALID_REQUEST`、`INTERNAL`、`TOKEN_REVOKED` …）已随 `BizError` 一起删除：状态码由 `Fail`/`WithStatus` 的调用点直接给出。`type` 只剩五个值：`TypeAccountDisabled` 与设备流的四个状态。

### 2.4 状态即行为

| HTTP | 客户端动作 |
|---|---|
| 400 / 422 | 表单校验失败；按 `errors[].location` 高亮字段 |
| 401 | 静默刷新一次；仍 401 则清 token、跳登录 |
| 403 | 提示无权限；`type` 是账号禁用时清 token、跳登录 |
| 404 / 409 / 412 | 提示，并刷新当前列表 |
| 405 | 客户端编程错误（路径存在、方法不对）；`Allow` 列出可用方法，不重试 |
| 429 | 按 `Retry-After` 冷却后再试 |
| 5xx | 通用失败提示；`detail` 是 `internal server error`，真实原因在服务端日志 |

### 2.5 头部

| 头 | 状态 |
|---|---|
| `X-Ech0-Catalog` + 412 保护 | 保留（`internal/router/catalog.go`） |
| `Retry-After` | 保留（429）：由中间件显式设置；Huma 渲染的响应不带头（`StatusError` 没有 header 通道） |
| `Location` | 201 响应带创建后资源的地址（RFC 9110 §6.4.2），仅当该资源有规范化 GET：`/echo/{id}`、`/files/{id}`。其余创建端点只回 201 + 资源体 |
| `Allow` | 405 响应由 gin 自动填充（`internal/router/fallback.go` 只负责把响应体换成 problem 文档） |
| `ETag` / `If-None-Match` | 已实现（`internal/middleware/conditional.go`）：只对能缓存的 JSON 读生效——带 `Cache-Control: no-store` 的响应（所有认证路由）与流式响应直接跳过；弱 ETag，命中回 304 且无体 |
| `Deprecation`（RFC 9745）+ `Sunset`（RFC 8594） | 新增，由 catalog 的 `deprecated`/`sunset` 驱动 |
| `RateLimit-*` | **不做**：仍是 draft，`Retry-After` 已满足客户端退避需求，维护 draft 版本是纯负债 |

### 2.6 字段对照（迁移映射）

| 旧 | 新 |
|---|---|
| `Result.Code`（1/0） | 删除；成败由 HTTP 状态码 / `response.ok` 表达 |
| `Result.Msg`（成功） | 删除；前端 i18n |
| `Result.Msg`（失败） | `problem.detail` |
| `Result.ErrorCode` | 删除；行为由 `problem.status`、少数由 `problem.type` 表达 |
| `Result.MessageKey` / `MessageParams` | 删除；文案由前端按状态码自行选择 |
| `Result.Data`（成功） | 响应体本身 |
| `Result.Data`（失败，恒为 `null`） | 删除 |
| `web`: `res.code === 1` | `if (res)`：`null` 才是失败；空体成功由请求层补一个真值哨兵 |
| `web`: `res.msg` | 按 `problem.status` 选前端 key |

---

## 3. 代码结构变化

### 3.1 删除（净减法）

| 文件 | 处理 |
|---|---|
| `internal/model/common/result.go`（68 行） | **删**。`Result[T]`、`OK`、`Fail`、`FailWithErrorCode`、`FailWithLocalized`、`OKWithCode`、`DEFAULT_SUCCESS_CODE`/`DEFAULT_FAILED_CODE` 全部消失 |
| `internal/model/common/success.go`（102 行，63 个常量） | **删** |
| `internal/model/common/i18n_keys.go` | **删**（28 个 `MsgKey*`、`MessageKeyFromErrorCode`、`MessageKeyFromMessage`）及其测试 |
| `internal/handler/humares/envelope.go`（23 行） | **删** |
| `internal/handler/humares/handle.go`（20 行） | **删**（`Wrap` 不再需要） |
| `internal/handler/humares/errors.go` | **改**：只剩 `installErrorModel` 与其内部 `frameworkProblem`；`ErrorBody`/`apiError`/`detailSuffix`/`Err` 全部删除 |
| `internal/handler/response/response.go` | **改**：`Response{Status, Data, Err}` + `Payload/Created/NoContent/Failed/FailedWithStatus` + `Execute`，失败分支交给 `httperr`，供 22 条裸 gin JSON 路由继续使用 |
| `internal/router/huma.go:207-216` | **改**：`route` 泛型签名换成 Huma 原生签名 |

### 3.2 新增

只新增一个文件 `internal/httperr/problem.go`。放在 `httperr` 而不是 `humares`，是因为 `internal/middleware` 也要写 problem，而中间件不该 import handler 层；反过来 `internal/model/common` 不能 import huma，否则整个领域层被 HTTP 框架绑死。

```go
package httperr

// Problem 就是 Huma 的 RFC 9457 文档 —— 没有自定义类型。
// 它的 ContentType() 会把 Content-Type 换成 application/problem+json。

// From 把任意 error 投影成 problem：
//   *commonModel.StatusError（含 WithStatus 包装）→ 沿链取最外层带状态的
//   其余 error                                    → 500，detail 固定为 "internal server error"
// 5xx 的真实原因只进服务端日志，不进响应体。
func From(err error) *huma.ErrorModel

// Write / Abort 给裸 gin 路由与中间件用。
func Write(ctx *gin.Context, err error)
func Abort(ctx *gin.Context, err error)
```

`type` 只在两处出现（`TypeAccountDisabled` 与设备流四种），由构造时传入的 `StatusError.Type` 决定；`Request` 侧没有任何自定义字段。`huma.NewError` / `huma.NewErrorWithContext` 仍被覆写为返回 `*huma.ErrorModel`——这是 Huma 官方文档的 *Custom Errors* 路径，覆写体只有几行，负责把 Huma 的框架错误（解码失败、校验失败）也投影成同一文档。

### 3.3 handler 前后对照

输出类型由"信封的别名"变成"响应体本身"，每个 op 一行结构体，**handler 不 import Huma**（保持现有的框架中立约束）。

**硬约束**：输出类型必须是具名结构体且字段名恰好是 `Body`。Huma 的 `processOutputType`（`huma.go:1701`）只认这个名字；写成裸别名（`type EchoOutput = model.Echo`）会让 `outBodyIndex == -1`，于是 `DefaultStatus` 落成 `204`、响应 schema 为空——端点静默返回空体，测试却可能照样通过。`struct{ Body T }` 是唯一正确形态，纯动作端点用 `struct{}`。

```go
// 前
type ConnectOutput = commonModel.Result[connectModel.Connect]

func (h *ConnectHandler) GetConnect(ctx context.Context, _ *GetConnectInput) (ConnectOutput, error) {
	connect, err := h.connectService.GetConnect()
	if err != nil {
		return ConnectOutput{}, err
	}
	return commonModel.OK(connect, commonModel.CONNECT_SUCCESS), nil
}

// 后
type ConnectOutput struct {
	Body connectModel.Connect
}

func (h *ConnectHandler) GetConnect(ctx context.Context, _ *GetConnectInput) (*ConnectOutput, error) {
	connect, err := h.connectService.GetConnect()
	if err != nil {
		return nil, err
	}
	return &ConnectOutput{Body: connect}, nil
}
```

纯动作端点（`AddConnect`、`DeleteConnect` 这类无返回值的操作）输出 `struct{}`，Huma 自动回 `204`（`huma.go:1761-1767`：无 `Body` 字段 → `DefaultStatus = 204`）。

路由辅助函数同步简化：

```go
// 前
func (a *apiRegistry) route[I, T any](e Endpoint, op huma.Operation,
	h func(context.Context, *I) (commonModel.Result[T], error)) {
	...
	huma.Register(a.API, op, humares.Wrap(h))
}

// 后
func (a *apiRegistry) route[I, O any](e Endpoint, op huma.Operation,
	h func(context.Context, *I) (*O, error)) {
	...
	huma.Register(a.API, op, h)   // Endpoint 与 budget 的拼接逻辑不变
}
```

### 3.4 影响的文件量

| 区域 | 量 |
|---|---|
| Huma operation | 106（14 个 `internal/router/*.go`） |
| handler 输出类型 | 106 个（`internal/handler/` 下 26 个文件） |
| `commonModel.OK(...)` 调用点 | 112 |
| 裸 gin JSON 响应（`response.Execute`） | 22 |
| `Result[T]` 引用（含测试/mock） | 190 处、39 个文件 |
| `web/` 中 `code === 1` | 126 处、45 个文件 |

`internal/service/**` 与 `internal/repository/**` **零改动**——它们只有 `internal/service/connect/connect.go:229` 一处内部变量用了 `Result`，`PageQueryResult` 合并时顺带改。

---

## 4. 客户端

### 4.1 `web/`

- `web/src/service/request/index.ts`：`request<T>()` 返回 `T | null`——成功给资源本身，失败给 `null`（toast 已在层内弹过）。非 2xx 由 ofetch 抛出，`problemOf` 取 `FetchError.data` 投影成 `App.Api.Problem`。
- `web/src/service/request/problem.ts`：`problemMessage(problem)` 按 `status` 选 `common.*` 的文案（400/422 无效请求、401 登录过期、403 无权限、404 不存在、409/412 冲突、429 过于频繁、5xx 服务故障），未知状态回落到通用失败文案。
- **空体成功**（`204`、纯动作端点）由 `settle()` 补一个真值哨兵，于是 `if (res)` 对全部端点统一表示成功，`null` 永远是失败；否则 `if (res)` 会把删除成功的 `undefined` 当成失败。
- 重试与登出只按状态与类型判断：`problem.status === 401` 触发一次静默刷新；`problem.type === ACCOUNT_DISABLED_TYPE` 清 token 跳登录。没有需要维护的码表。
- 诊断类接口（S3/agent/邮件/webhook 连通性测试）传 `onProblem: reportProblemDetail`，直接把服务端的英文 `detail` 展示给管理员——那里 `detail` 就是答案。
- 126 处 `res.code === 1` → `if (res)`，40 处 `res.msg` 删除或改用本地 key（45 / 23 个文件），成功 toast 统一用新增的 `common.success`。
- `web/src/service/request/sse.ts` 与 `web/src/lib/file/upload.ts` 的裸 fetch 错误体按 `App.Api.Problem` 解析。
- `web/src/typings/api-base.d.ts` 的 `App.Api.Response<T>` 删除，换成 `App.Api.Problem`；其余 14 个手写 `.d.ts` 本轮保留。
- **生成类型不在本轮**：`web/` 的 922 行手写 DTO 与 spec 仍是两份事实源，全部换成 `@hey-api/openapi-ts` 的输出需要改 89 个文件的 `App.Api.*` 引用，收益是"少一份定义"而不是"修一个 bug"。已把它列为下一轮，不在 `just check` 里假装有门禁。

### 4.2 `apps/cli/`

`apps/cli/src/api/http.ts` 的 `Envelope`/`parseEnvelope` 删除：`call<T>` 直接返回解码后的 body，`callWithMeta<T>` 额外给需要 `Set-Cookie` / `--json` 原文 / `X-Ech0-Catalog` 的调用点；`parseProblem` 把 problem 映射回 CLI 既有的 `Ech0Error`（`{ status, type }`，`detail` 作 message）。设备流按 `details.type` 的四类 URI 分支（RFC 8628 的 pending/slow_down/expired/access_denied 本来就是 body 语义，`type` 是它的标准表达）。`src/api/generated/` 由 `bun run gen:api` 从 spec 生成，`just cli gen-check` 守漂移。CLI 自身的 stdout 输出契约（`--json`/`--yaml`）是另一个东西，保持不变。

### 4.3 `hub/`

`hub/src/services/{echoApi,connectApi,healthz}.ts` 改为直接解析资源体（`EchoQueryPage` / `Connect` / `HealthzData`），`ApiResult<T>` 删除。**不保留兼容读取器**：hub 是聚合并发它的实例的客户端，旧实例在新契约下会整条不可用（列表里显示为离线），这是"一次性硬切、不留 v1"的直接代价，已在 CHANGELOG 标注 hub 需与服务器同版本发布。

---

## 5. 门禁（CI 必须能回答的问题）

| 问题 | 机制 | 状态 |
|---|---|---|
| 路由表每条是否在契约里 | `internal/router/catalog_routes_test.go:16-43,58` 双向对齐 + 具名排除清单 | 已有 |
| 客户端类型是否同源 | `apps/cli` 的 `bun run gen:api` + `just cli gen-check`（`git diff --exit-code`，已是 CLI 自己的 check） | 已有 |
| 是否破坏已发布客户端 | `scripts/api-breaking.sh`（`just api-breaking`，已挂进 `scripts/check.sh`） | **新增** |
| 错误是否可机读 | 状态码即行为，`type` 表达少数需区分的语义；**不引入码表** | — |

`api-breaking` 的基线优先取 `internal/openapi/baseline.yaml`，没有该文件时取最新 `v*` tag（`ECH0_API_BASE_REF` 可覆盖）。本次硬切用 `just api-baseline` 把新契约写成基线，于是 `just check` 在合并前就是绿的；**v7 打 tag 之后删掉 `baseline.yaml`**，基线自动回到 tag，不需要任何环境变量。`ECH0_API_BREAKING_ALLOW=1` 只作为紧急逃生口。`contract_test.go`（kin-openapi 逐响应校验）没有落地：`openapi-check` + `gen-check` + `oasdiff` 已经覆盖它的失败模式，再加一层只会重复断言 Huma 自己的生成逻辑。

`Deprecated` 目前全仓没有端点使用，`Deprecation`/`Sunset` 头的写入点留在 §2.4 的定义里，等第一个弃用端点出现时再配测试——现在写就是一个空转的测试。

---

## 6. 执行顺序与工作量

| 阶段 | 内容 | 可独立发布 | 估算 |
|---|---|---|---|
| **A** 门禁与生成链 | `oasdiff` 接入 `scripts/check.sh` + `baseline.yaml`；`apps/cli` 的 `gen-check` 已有 | 是（零行为变化） | 1 天 |
| **B** 契约切换 | `internal/httperr/problem.go` + `route` 签名 + 106 个输出类型 + 112 处 `OK` + `response.Execute` + `List[T]` 合并 + `cursor` 不透明化 + `web/` 全量切换 + `apps/cli` | **否**（后端与 SPA 同一二进制，必须原子合并） | 5–6 天 |
| **C** 收尾 | `hub` 硬切读取器、文档同步（`i18n-contract.md`、`huma-raw-gin-endpoints.md`、`architecture-overview.md`、`cli-design.md`、`README.md`）、残留删除 | 是 | 1.5 天 |

合计 **7.5–8.5 人日**。B 阶段是唯一的大 PR（约 120 文件，其中 100+ 是机械替换），建议先合并 A，再在 B 的分支上按 `internal/handler/<domain>` 分批提交以便 review，但**一次性合入**。

---

## 7. 风险

| 风险 | 影响 | 缓解 |
|---|---|---|
| **硬切破坏旧 hub** | 旧版 hub（部署在别处、连新版实例）立刻失败 | hub 与实例同仓同 release；发布说明置顶提示。刻意不做双形态读取器：兼容层会活过两个 minor 然后腐烂（§4.3） |
| 第三方 access-token 集成直接读 `code` | 集成方解析失败 | 这是硬切的固有代价：本次迁移没有弃用期，只能靠发布说明 + §2.6 对照表。`Sunset` 头服务于之后的**端点级**弃用，不适用于这次形态变更 |
| 112 处成功文案需在前端重建 | 有提示语可能"消失" | B 阶段逐条比对 `res.msg` 与现有 `web/` key；`i18n-check` 兜住缺失 key |
| 126 处 `code === 1` 漏改 | 成功后进入错误分支（表现为不刷新/不提示） | 门禁 `rg` 规则强制为空；`pnpm -C web build` 的类型检查会先抓到 `res.code` 不存在 |
| B 阶段是原子大 PR | review 成本高 | 按域分批提交 + 门禁先行（A 已合并） |
| `openapi.yaml` 一次性大 diff | 无法逐条 review 契约变化 | A 阶段先把 `oasdiff` 接上，B 的 diff 里只有"删信封 + 加 problem"两类变化 |
| `List[T]` 合并影响 20+ service 返回值 | 编译期全量暴露，风险可控 | 泛型改名，机械替换；`go build ./...` 即验证 |

---

## 8. 明确不做

| 不做 | 理由 |
|---|---|
| protobuf / ConnectRPC / grpc-gateway / cmux | §1.1；44 条非 JSON 路由是硬约束 |
| `{"error":{code,message,status,details}}` 形态 | §0 第 2 点 |
| URL 版本号（`/v2`、`Accept: application/vnd.ech0.v2+json`） | §1.3 |
| AIP-136 `:verb`、AIP-122 资源名 | 纯 churn |
| `page_size` / `page_token` 重命名 | 丢失页码跳转与 `total` |
| `FieldViolation` 新结构 | Huma `ErrorDetail` 已够用 |
| `RateLimit-*` 头 | draft 规范，负债 |
| GraphQL | 无多客户端聚合需求，与 scope/audience 模型无对应物 |
| 把 SSE / multipart 搬进 Huma | "handler 框架中立"是刻意约束 |
| OpenAPI-first 重写 | 与 Huma 同收益、迁移全量 |
| 动 `Endpoint` 声明 / `authz` / 分层 | 与契约形态正交，且现状优于多数仓库 |

---

## 9. 验收（可执行）

1. `just check` 9/9 通过，其中 `api-breaking` 对比 `baseline.yaml`，catalog 双向对齐测试仍在。
2. `rg 'code\s*[=!]==\s*1' web/src` 无结果。
3. `rg 'commonModel\.Result\[|humares\.Envelope|res\.Response\{' internal` 无结果。
4. `internal/model/common/success.go` 已删除，且 `rg 'MessageKeyFromMessage' internal` 无结果。
5. `internal/openapi/openapi.yaml` 中每个非 2xx 响应都引用 `ErrorModel`，媒体类型为 `application/problem+json`。
6. 任一端点：`curl -i` 成功体无 `code`/`msg`/`data` 键；失败体是 problem+json，只有 `type`/`title`/`status`/`detail`/`errors`，且 `status` 与 HTTP 状态一致。
8. `rg 'message_key|"code"' internal/httperr internal/handler/humares internal/model/common` 无结果。
9. `just web test`、`just web build`、`apps/cli` 的 `bun test`/`bun run typecheck`、`hub` 的 `vue-tsc -b` 全部通过（hub 没有单测脚本）。
10. 对只支持 `POST` 的路径发 `GET` → `405` + `Allow`，体是 problem+json（`internal/router/fallback.go`）。
11. 创建类端点 → `201`；`POST /echo` 与 `POST /files/{id}/commit` 另带 `Location`。
12. 可缓存的 JSON `GET` 带弱 `ETag`；带 `If-None-Match` 重放 → `304` 空体，响应仍带 `ETag`（`internal/middleware/conditional.go`）。

---

## 10. 已拍板（原待定项，均已实施）

| # | 问题 | 结论 |
|---|---|---|
| **Q1** | 成功文案是否由服务端下发 | 不下发。63 个常量、112 处 `OK` 调用删除，前端统一用 `common.success`，需要具体措辞的页面用页面自己的 key。 |
| **Q2** | 创建类 POST 返回 `201` 还是 `200` | `201` + 创建后资源；删除与纯动作返回 `204`。 |
| **Q3** | 列表是否统一 | 统一为 `commonModel.List[T]{items,total?,nextCursor?}`，`PageQueryResult`/`PageResult` 已删除。 |
| **Q4** | 错误里是否保留 `code` / `message_key` | 都不保留。`status` 承载行为，`type`（RFC 9457）承载必须区分的少数语义，`detail` 是英文开发者文本；服务端不再产出用户可见文案。详见 §2.3。 |
