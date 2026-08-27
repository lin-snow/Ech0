---
title: 事件推送
description: Webhook 是什么、如何配置、请求格式、验签与重试（新手向）
---

**事件推送（Webhook）** 的含义可以一句话说清：当 Ech0 里发生了某件事（例如发了新动态、新用户注册），你的服务器可以向一个**你自己提供的网址**自动发一条 **HTTP POST** 请求，里面带上**事件类型和 JSON 数据**。  
这样你就能把 Ech0 和 **Slack、Discord、自建机器人、审计系统** 等连起来，而不必一直轮询问「有没有新消息」。

---

## 这篇文档适合谁读

- 你是**管理员**，要在后台配置接收地址。
- 你准备在**自己的服务**里写几行代码接收 POST（或用 Zapier、n8n 等能暴露 HTTPS URL 的工具）。
- 不需要先懂分布式系统；下面会说明**配置填什么**、**请求长什么样**、**怎样算投递成功**。

---

## 在后台里怎么配

1. 使用**管理员账号**进入 **系统设置 → Webhook**。
2. **新建**：填写 **名称**、**接收 URL**（`http` 或 `https`）、可选 **Secret**（用于验签，强烈建议生产环境填写）。
3. 保存并**启用**。列表里可看到最近一次投递成功/失败等状态。
4. 编辑时：**Secret 通常不会回显**；若提交时留空，可能表示**清空签名**（以当前版本界面提示为准）。

若你更习惯用 API 管理，管理员鉴权后也可以使用（路径以你部署的 `/api` 为准）：`GET/POST/PUT/DELETE /webhook` 等，详见实例上的 **Swagger**。

---

## 为什么不能用「内网地址」当 URL

为防止 **SSRF**（服务端伪造请求攻击内网），Ech0 会拒绝明显不安全的接收地址，例如：

- `localhost`、纯内网 IP（如 `10.x`、`192.168.x`、`127.0.0.1`）
- 以 `.local` 结尾的主机名

因此：**在你自己笔记本上监听的 `http://127.0.0.1:3000` 不能直接填**。  
本地调试请使用 **公网可访问的隧道域名**（如 ngrok、Cloudflare Tunnel 等临时 HTTPS 地址），与生产环境一致地配置验签。

---

## 投递成功与失败怎么判定

- **成功**：你的服务器对这次 HTTP 请求返回 **2xx** 状态码。
- **失败**：网络错误、超时、返回 4xx/5xx 都算失败。

一次发送失败时，客户端会**在短时间内自动重试**（例如共 3 次，间隔约 500ms → 1s → 2s，**单次请求超时约 5 秒**；管理后台里的「测试」可能使用更短的重试策略）。  
若这几次即时重试仍失败，本次投递记为失败（在 Webhook 状态里体现），事件**不会**再排队补投。

**建议**：接收端收到请求后**尽快返回 2xx**，耗时逻辑先**丢进队列异步处理**，避免超过 5 秒导致反复重试。

---

## 会收到哪些「事件」

下面是一张**完整清单**（与当前后端白名单一致）。只有这些 **topic** 会作为业务事件推送到 Webhook：

| Topic                                                            | 含义（白话）                       |
| ---------------------------------------------------------------- | ---------------------------------- |
| `user.created` / `user.updated` / `user.deleted`                 | 用户创建、资料变更、删除           |
| `echo.created` / `echo.updated` / `echo.deleted`                 | 动态（Echo）发布、编辑、删除       |
| `comment.created` / `comment.status.updated` / `comment.deleted` | 评论创建、状态变更（如审核）、删除 |
| `resource.uploaded`                                              | 资源/文件上传完成                  |
| `system.snapshot` / `system.export`                                | 快照或导出任务相关                 |
| `system.snapshot_schedule.updated`                                 | 快照计划被修改                     |

说明：评论与审核相关行为也可结合 [评论系统](/docs/guide/comment) 理解；快照类与 [数据管理](/docs/guide/datacontrol) 中的计划任务相关。

---

## HTTP 请求长什么样

### 请求方法与安全头

- **Method**：`POST`
- **`Content-Type`**：`application/json`
- **`User-Agent`**：`Ech0-Webhook-Client`
- **`X-Ech0-Event`**：事件 topic（例如 `echo.created`）
- **`X-Ech0-Event-ID`**：本次投递的事件 ID（字符串，可用于**幂等去重**）
- **`X-Ech0-Timestamp`**：Unix 时间戳（秒，UTC）
- **`X-Ech0-Signature`**：仅在配置了 **Secret** 时出现，格式为 `sha256=<十六进制小写字符串>`

### JSON 正文结构

```json
{
  "topic": "echo.created",
  "event_name": "EchoCreated",
  "payload_raw": {},
  "metadata": {},
  "occurred_at": 1710000000
}
```

| 字段          | 含义                                     |
| ------------- | ---------------------------------------- |
| `topic`       | 与请求头 `X-Ech0-Event` 一致             |
| `event_name`  | 服务端事件类型名（如 Go 结构体名）       |
| `payload_raw` | **核心业务数据**（JSON），结构随事件变化 |
| `metadata`    | 附加键值，可能为空                       |
| `occurred_at` | 事件发生时间（UTC，Unix 秒）             |

完整字段与嵌套结构以当前版本及 **`/swagger`** 为准。

---

## 如何验证请求确实来自 Ech0（验签）

若配置了 **Secret**，请用同一 Secret 对**原始请求体字节**（未解析 JSON 前的完整 body）做 **HMAC-SHA256**，与请求头 `X-Ech0-Signature` 中 `sha256=` 后的十六进制字符串**常量时间比较**。

**Node.js 示例：**

```javascript
import crypto from "node:crypto";

function verifyEch0Signature(rawBody, secret, signatureHeader) {
  if (!signatureHeader?.startsWith("sha256=")) return false;
  const received = signatureHeader.slice("sha256=".length);
  const expected = crypto
    .createHmac("sha256", secret)
    .update(rawBody)
    .digest("hex");
  return crypto.timingSafeEqual(
    Buffer.from(received, "utf8"),
    Buffer.from(expected, "utf8"),
  );
}
```

**强烈建议同时：**

- 校验 **`X-Ech0-Timestamp`** 是否在可接受的时间窗口内（减轻重放攻击）。
- 用 **`X-Ech0-Event-ID`** 做去重，同一 ID 只处理一次。

---

## 测试连通性

在管理后台对某条 Webhook 执行**测试**时，会发送 **`webhook.test`** 类测试事件（用于验证 URL、证书与验签链路），并更新该 Webhook 的最近状态。  
若使用 API，一般为管理员 `POST /api/webhook/:id/test`（以 Swagger 为准）。

---

## 常见问题

**一直显示失败？**  
检查接收 URL 是否公网可达、TLS 证书是否被客户端信任、是否在 **5 秒内**返回 2xx、防火墙/WAF 是否拦截、验签 Secret 是否与 Ech0 里配置一致。

**某个业务事件从来没收到？**  
确认该 Webhook **已启用**；确认事件属于上文 **topic 白名单**；确认 URL 未被安全策略拒绝（例如误填内网地址）。

**接收端业务处理失败了要不要返回 500？**  
不建议。Webhook 侧只要**确认收到并持久化**就应返回 2xx；后续业务失败应在你方队列里重试，否则 Ech0 会认为是投递失败并触发即时重试。

---

## 接收端最佳实践（小结）

- **幂等**：按 `X-Ech0-Event-ID` 去重。
- **鉴权**：校验 `X-Ech0-Signature`。
- **防重放**：限制 `X-Ech0-Timestamp` 窗口。
- **快响应**：先入队再异步处理。
- **可观测**：记录 topic、event_id、HTTP 状态与延迟。
