---
title: MCP 接入
description: 把 Ech0 接到 Cursor、Claude 等 AI 客户端
---

接入 MCP 后，你可以直接在 AI 客户端里读写自己的 Ech0：查动态、发帖、看评论、管理 Webhook，不用切到网页。

Ech0 自带 MCP 服务，和站点同一个地址、同一个端口，不需要额外部署。

---

## 开始前

- 实例已能正常访问（建议使用 HTTPS）。
- 一个**管理员账号**：首次授权需要你在浏览器里确认。
- 支持**远程 MCP** 的客户端，如 Cursor、Claude Desktop。只支持本地 stdio 的客户端需要网关转发。

---

## 接入方式一：浏览器授权（推荐）

1. 在客户端里添加一个远程 MCP 服务，地址填 `https://你的域名/mcp`。
2. 客户端会打开浏览器；用管理员账号登录并确认授权。
3. 回到客户端，工具列表出现即可使用。

- 支持远程 OAuth 的客户端会自动发现授权入口。若客户端让你选认证方式，选 **OAuth / 浏览器登录**，不要填令牌。
- 授权页会列出客户端申请的权限，确认后才会生效；随时可以在 **系统设置 → 访问令牌** 里吊销。
- 授权只能由管理员在浏览器里完成，令牌不能代替确认。

---

## 接入方式二：手动令牌

适用于不支持 OAuth 的客户端，或你只想在脚本里用。

1. 打开 **系统设置 → 访问令牌**，新建令牌。
2. **受众**选 **MCP（AI Agent）**。
3. 按需勾选权限（见下表），有效期建议 8 小时或 1 个月。
4. 在客户端里这样填：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "https://你的域名/mcp",
      "headers": {
        "Authorization": "Bearer <你的令牌>"
      }
    }
  }
}
```

本地调试把地址换成 `http://localhost:6277/mcp`。

| 想让它做什么 | 勾选                            |
| ------------ | ------------------------------- |
| 看动态       | `echo:read`                     |
| 发帖、改帖   | `echo:write`                    |
| 删帖、删标签 | `echo:delete`                   |
| 读私密动态   | 再加 `echo:private`             |
| 读 / 发评论  | `comment:read`、`comment:write` |
| 删评论       | 再加 `comment:delete`           |
| 文件         | `file:read`、`file:write`       |
| 互联实例     | `connect:read`、`connect:write` |
| 管理 Webhook | `admin:webhook`                 |

不要勾选当前用途之外的管理权限。

---

## 你可以让它做什么

- 找动态：按关键词、标签、日期，或随机来一篇
- 发帖、改帖、删帖、点赞
- 读评论、发评论
- 查文件和标签
- 查看互联的实例
- 管理 Webhook

权限不够的能力不会出现在客户端的工具列表里；删除类操作在支持确认的客户端上会先询问你。

---

## 常见问题

**连上了但没有工具？** 令牌权限不足，或授权没完成。检查权限，或重新授权一次。

**返回 403？** 受众不对。手动令牌的受众必须选「MCP（AI Agent）」。

**返回 401，或反复要求登录？** 令牌过期，或客户端的 OAuth 流程没走完；重新授权。

**地址怎么填？** `https://你的域名/mcp`，不要再加 `/api`。

**客户端只支持本地命令？** 用一个网关或代理把请求转发到上面的 HTTP 地址。

---

## 进阶

- 协议：优先 **2026-07-28**（无状态）；也兼容 2025-11-25 / 2025-06-18 / 2025-03-26。
- 只有 `POST /mcp` 可用；`GET`、`DELETE` 返回 405，服务端不发放会话。
- 能力清单：`GET /api/mcp/manifest`（需要 `mcp:read`），返回工具名、权限与是否破坏性。
- 授权发现：`/.well-known/oauth-protected-resource/mcp`、`/.well-known/oauth-authorization-server`。

---

## 延伸阅读

[访问令牌](/zh/docs/guide/accesstoken) · [命令行（CLI）](/zh/docs/guide/cli) · [事件推送](/zh/docs/guide/webhook) · [评论系统](/zh/docs/guide/comment)
