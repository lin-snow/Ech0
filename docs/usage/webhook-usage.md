# Ech0 Webhook 使用说明

这份文档基于当前项目实现编写，目标是让你可以直接把 Webhook 配起来并稳定落地到业务系统。

---

## 1. Webhook 是什么

Webhook 是 Ech0 的事件推送能力。当系统内发生关键事件（如用户、Echo、资源、系统任务相关事件）时，Ech0 会把事件通过 HTTP `POST` 推送到你配置的 URL。

你可以用它做：
- 同步数据到外部系统（检索库、BI、消息中心）
- 自动化流程触发（通知、审计、联动任务）
- 跨系统事件编排

---

## 2. 当前支持能力（实装）

- 支持管理接口：新增 / 修改 / 删除 / 列表 / 测试发送
- 支持启用开关（`is_active`）
- 支持签名头（HMAC-SHA256，`X-Ech0-Signature`）
- 支持状态记录（`last_status`、`last_trigger`）
- 支持失败即时重试（单次发送内指数退避）
- 支持事件主题白名单（非白名单事件不会发 webhook）

---

## 3. 管理端接口

以下接口都在鉴权后使用，且要求管理员权限：

- `GET /webhook`：获取 Webhook 列表
- `POST /webhook`：创建 Webhook
- `PUT /webhook/:id`：更新 Webhook
- `DELETE /webhook/:id`：删除 Webhook
- `POST /webhook/:id/test`：测试该 Webhook 连通性

创建/更新请求体（`WebhookDto`）：

```json
{
  "name": "My Receiver",
  "url": "https://example.com/ech0/webhook",
  "secret": "your-signing-secret",
  "is_active": true
}
```

字段说明：
- `name`：Webhook 名称（必填）
- `url`：接收地址（必填，必须是 `http/https`）
- `secret`：签名密钥（可选；不填则不带签名头）
- `is_active`：是否启用

---

## 4. URL 安全校验规则（后端强校验）

为了降低 SSRF 风险，Webhook URL 会被校验，不通过会直接拒绝：

- 必须是 `http` 或 `https`
- 禁止 `localhost`
- 禁止 `.local` 结尾域名
- 禁止内网/环回/链路本地等 IP（如 `127.0.0.1`、`10.x.x.x`、`192.168.x.x` 等）

这意味着：本机联调地址一般无法直接作为 Webhook 目标地址。  
本地调试建议通过可公开访问的临时隧道域名。

---

## 5. 实际推送请求格式

### 5.1 请求方法与头

- Method: `POST`
- `Content-Type: application/json`
- `User-Agent: Ech0-Webhook-Client`
- `X-Ech0-Event`: 事件 topic（例如 `echo.created`）
- `X-Ech0-Event-ID`: 事件 ID（时间戳纳秒字符串）
- `X-Ech0-Timestamp`: Unix 秒级时间戳（UTC）
- `X-Ech0-Signature`: `sha256=<hex>`（仅配置了 `secret` 时存在）

### 5.2 请求体

```json
{
  "topic": "echo.created",
  "event_name": "EchoCreated",
  "payload_raw": {
    "...": "具体事件载荷"
  },
  "metadata": {
    "...": "附加元信息"
  },
  "occurred_at": 1710000000
}
```

字段说明：
- `topic`：事件主题
- `event_name`：事件类型名（Go struct name）
- `payload_raw`：事件原始 payload（最核心业务数据）
- `metadata`：元信息（可能为空）
- `occurred_at`：事件发生时间（UTC，Unix 秒）

---

## 6. 当前可投递的事件主题（白名单）

只有以下 topic 会投递到 Webhook：

- `user.created`
- `user.updated`
- `user.deleted`
- `echo.created`
- `echo.updated`
- `echo.deleted`
- `comment.created`
- `comment.status.updated`
- `comment.deleted`
- `resource.uploaded`
- `system.snapshot`
- `system.export`
- `system.snapshot_schedule.updated`

---

## 7. 成功判定与重试机制

### 7.1 一次投递何时算成功

当接收端返回 HTTP `2xx` 时，判定成功。  
否则（网络错误、超时、4xx/5xx）判定失败。

### 7.2 即时重试（单次发送内）

Webhook 发送内置指数退避重试：
- 最大尝试次数：3 次（测试接口为 2 次）
- 每次失败后等待：`500ms` -> `1s` -> `2s`（测试接口为更短间隔）
- 请求超时：5 秒

### 7.3 状态回写

每次投递结束都会更新：
- `last_status`：`success` / `failed`
- `last_trigger`：本次触发时间（UTC）

---

## 8. 签名校验（接收端建议强制启用）

当你配置了 `secret`，Ech0 会对请求体做 HMAC-SHA256：

- 算法：`HMAC_SHA256(secret, rawBodyBytes)`
- Header：`X-Ech0-Signature: sha256=<hex>`

### Node.js 校验示例

```js
import crypto from 'node:crypto'

function verifyEch0Signature(rawBody, secret, signatureHeader) {
  if (!signatureHeader || !signatureHeader.startsWith('sha256=')) return false
  const received = signatureHeader.slice('sha256='.length)
  const expected = crypto.createHmac('sha256', secret).update(rawBody).digest('hex')
  return crypto.timingSafeEqual(Buffer.from(received), Buffer.from(expected))
}
```

建议同时校验：
- `X-Ech0-Timestamp`（防重放，限制时间窗口）
- `X-Ech0-Event-ID`（幂等去重）

---

## 9. 前端页面怎么用（管理员）

`设置 -> Webhook` 页面：

1. 点击“新建 Webhook”
2. 填写名称、URL、可选 secret，设置启用状态
3. 保存后在表格查看状态标签（最近成功/失败/未知）
4. 可以在表格里快速开关启用、编辑、删除

注意：
- 编辑表单不会回显旧 secret（安全设计）
- 若提交编辑时 secret 为空，会按当前实现更新为空（等于清空签名）

---

## 10. 手工测试（推荐）

如果你想直接用接口测试某个 webhook：

```bash
curl -X POST "https://<your-ech0-domain>/api/webhook/<webhook-id>/test" \
  -H "Authorization: <your-token>"
```

测试会发送一条 `webhook.test` 事件，用于验证连通性和签名链路。  
测试结果也会更新该 webhook 的 `last_status` 和 `last_trigger`。

---

## 11. 常见问题排查

### 11.1 一直失败（红色状态）

优先检查：
- 接收 URL 是否公网可达、证书是否有效
- 接收端是否在 5 秒内响应
- 接收端是否返回 `2xx`
- 是否启用了签名校验但 secret 不一致
- 是否被防火墙/WAF 拦截

### 11.2 为什么我的事件没收到

常见原因：
- Webhook 未启用（`is_active=false`）
- 事件不在白名单 topic 内
- URL 被安全校验拒绝（内网/localhost）

### 11.3 要不要返回业务错误码？

Webhook 接收端建议：
- 只要接收并落库成功就返回 `2xx`
- 后续异步处理失败不要返回 `5xx`，否则会触发即时重试

---

## 12. 接收端最佳实践

- 做幂等：用 `X-Ech0-Event-ID` 去重
- 做鉴权：校验 `X-Ech0-Signature`
- 做防重放：校验 `X-Ech0-Timestamp` 时间窗口
- 快速 ACK：先入队再异步处理，避免超时
- 可观测：记录 topic、event_id、status、延迟

---

如果你愿意，我可以继续补一份 `receiver-demo`（Node/Go 二选一）到 `examples/`，你拉起来就能直接接 Ech0 的 Webhook。
