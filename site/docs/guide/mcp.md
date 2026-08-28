---
title: MCP 接入
description: Model Context Protocol、/mcp 端点、令牌与 Tools 能力
---

**MCP（Model Context Protocol）** 是一套让「AI 客户端」和「你的服务」用统一方式对话的协议。Ech0 内建了 MCP Server：在**不另外起进程**的前提下，把帖子、评论、文件、互联等能力以 **Tools** 和 **Resources** 暴露给 Cursor、Claude Desktop 等支持 **远程 MCP（HTTP）** 的宿主。

传输层为 **Streamable HTTP**（JSON-RPC 2.0 over HTTP），与主服务**同一端口**；鉴权与 REST API 一样使用 **`Authorization: Bearer`**，但 **Audience 必须是 `mcp-remote`**，否则返回 **403**。

---

## 这篇文档适合谁读

- 你想在 **AI 编程助手**里直接「查我的 Ech0 帖子、发评论、管 Webhook」等。
- 你已有 [访问令牌](/docs/guide/accesstoken) 的概念，知道 **Scope** 表示权限。
- 不需要先读 MCP 规范全文；下面给出**端点、配置示例与能力表**。

---

## 和「普通 API 令牌」有什么区别

| 项目          | REST `/api/...`                          | MCP `/mcp`                                              |
| ------------- | ---------------------------------------- | ------------------------------------------------------- |
| 令牌 Audience | `public-client`、`cli`、`integration` 等 | **必须** `mcp-remote`                                   |
| 协议          | HTTP REST                                | JSON-RPC（`initialize`、`tools/list`、`tools/call` 等） |
| 用途          | 任意脚本、集成                           | 主要给 **MCP Host**（AI 客户端）                        |

在后台创建令牌时，**受众**请选择 **「MCP（AI Agent）」**（内部值 `mcp-remote`），并按需勾选 **权限范围**。

---

## 第一步：创建 MCP 专用令牌

在 **系统设置 → 访问令牌** 新建：

1. **受众**：选 **MCP（AI Agent）**（值为 `mcp-remote`）。
2. **Scopes**：按最小权限勾选（示例）
   - 只读动态：`echo:read`、`profile:read`
   - 要发帖：再加 `echo:write`
   - 评论：`comment:read`、`comment:write`
   - 文件：`file:read`、`file:write`
   - 互联：`connect:read`、`connect:write`
   - 通过 MCP 管理 Webhook：`admin:settings`（慎用）
3. **有效期**：建议 **8 小时** 或 **1 个月**，避免长期不过期。

创建后**保存好整串令牌**（通常只显示一次）。详见 [访问令牌](/docs/guide/accesstoken)。

---

## 第二步：在 MCP 客户端里配置

将 `url` 指向你的实例的 **`/mcp`**，并带上 Bearer 头（占位符换成真实令牌与域名）：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "https://你的域名/mcp",
      "headers": {
        "Authorization": "Bearer <你的访问令牌>"
      }
    }
  }
}
```

本地开发（默认端口以你实际为准，常见为 **6277**）：

```json
{
  "mcpServers": {
    "ech0": {
      "url": "http://localhost:6277/mcp",
      "headers": {
        "Authorization": "Bearer <你的访问令牌>"
      }
    }
  }
}
```

若客户端只支持 **stdio** 子进程而不支持远程 URL，需要用网关或代理把请求转到上述 HTTP 端点。

---

## 端点说明

| 方法   | 路径   | 说明                    |
| ------ | ------ | ----------------------- |
| `GET`  | `/mcp` | 服务状态信息            |
| `POST` | `/mcp` | JSON-RPC 请求（主入口） |

- **协议版本**（随版本升级可能变化）：以当前实例返回为准。
- 请求限流、请求体大小限制等以服务端实现为准；生产环境建议 **HTTPS** 并置于反向代理之后。

---

## 能力总览（Tools / Resources）

下列为当前实现中暴露的主要能力；**每个 Tool/Resource 都会单独校验 Scope**，无权限会拒绝。

### 帖子与标签（`echo:read` / `echo:write`）

| 类型     | 名称                                          | 说明                    |
| -------- | --------------------------------------------- | ----------------------- |
| Tool     | `search_posts`                                | 关键词 / 标签搜索，分页 |
| Tool     | `get_post`                                    | 按 UUID 取单篇          |
| Tool     | `get_today_posts`                             | 当日帖子（可带时区）    |
| Tool     | `list_tags`                                   | 标签列表与计数          |
| Tool     | `create_post` / `update_post` / `delete_post` | 创建 / 更新 / 删除      |
| Tool     | `like_post`                                   | 点赞                    |
| Tool     | `delete_tag`                                  | 删除标签并解除关联      |
| Resource | `ech0://posts/recent`                         | 最近帖子                |
| Resource Template | `ech0://posts/{id}`                  | 单篇（见 `resources/templates/list`） |
| Resource | `ech0://tags`                                 | 全部标签                |
| Resource | `ech0://stats/heatmap`                        | 热力图数据              |

### 评论（`comment:read` / `comment:write`）

| 类型     | 名称                               | 说明                                               |
| -------- | ---------------------------------- | -------------------------------------------------- |
| Tool     | `list_comments`                    | 某帖下的公开评论                                   |
| Tool     | `create_comment`                   | 与 `create_integration_comment` 等价，集成身份发评 |
| Tool     | `create_integration_comment`       | 与 REST `POST /api/comments/integration` 同源逻辑  |
| Resource | `ech0://comments/recent`           | 全站最近评论                                       |
| Resource | `ech0://guide/integration-comment` | 集成评论说明与 curl 示例                           |

### 文件（`file:read` / `file:write`）

| 类型     | 名称                       | 说明                    |
| -------- | -------------------------- | ----------------------- |
| Tool     | `list_files` / `get_file`  | 列表与元数据            |
| Tool     | `delete_file`              | 删除                    |
| Tool     | `create_external_file`     | 用外部 URL 登记文件记录 |
| Resource | `ech0://guide/file-upload` | 上传流程与 REST 说明    |

实际上传多使用 **`POST /api/files/upload`**（multipart），可把文件 `id` 再用于发帖；详见 Resource 内说明。

### 互联 Connect（`connect:read` / `connect:write`）

| 类型     | 名称                                  | 说明            |
| -------- | ------------------------------------- | --------------- |
| Tool     | `list_connects` / `get_connects_info` | 列表与对端信息  |
| Tool     | `add_connect` / `delete_connect`      | 添加 / 删除连接 |
| Resource | `ech0://connect/self`                 | 本实例公开信息  |

### Agent（`echo:read`）

| 类型 | 名称         | 说明                                |
| ---- | ------------ | ----------------------------------- |
| Tool | `get_recent` | AI 生成的站点近况摘要（可能有缓存） |

### Webhook（`admin:settings`）

| 类型 | 名称                                                                                      | 说明                            |
| ---- | ----------------------------------------------------------------------------------------- | ------------------------------- |
| Tool | `list_webhooks` / `create_webhook` / `update_webhook` / `delete_webhook` / `test_webhook` | 管理 Webhook（列表不含 secret） |

### 用户（`profile:read`）

| 类型     | 名称                | 说明                 |
| -------- | ------------------- | -------------------- |
| Resource | `ech0://profile/me` | 当前令牌对应用户资料 |

完整方法名、参数 Schema 以客户端 **`tools/list`**、**`resources/list`** 与 **`resources/templates/list`** 返回为准。

---

## 协议版本

`/mcp` 同时支持两代 MCP 协议，客户端用哪种都能连上：

- **2026-07-28（最新）**：无状态，没有 `initialize`，每个请求自带 `_meta` 与 `MCP-Protocol-Version` / `Mcp-Method` / `Mcp-Name` 头。
- **2025-11-25 / 2025-06-18 / 2025-03-26**：传统 `initialize` 握手，之后只需带 `MCP-Protocol-Version` 头。

服务端不发放会话（无 `Mcp-Session-Id`），`GET` / `DELETE /mcp` 返回 405。详见 [MCP 接入指南](https://github.com/lin-snow/Ech0/blob/main/docs/usage/mcp-usage.md#协议兼容)。

---

## 用 curl 自测（可选）

```bash
# Initialize
curl -X POST "https://你的域名/mcp" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'

# 列出 tools
curl -X POST "https://你的域名/mcp" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

# 搜索帖子（示例）
curl -X POST "https://你的域名/mcp" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_posts","arguments":{"query":"hello","page":1,"page_size":10}}}'
```

---

## 安全提示

- 令牌与 **Scope** 决定能做什么；**不要用 `admin:settings` 除非确有需要**。
- MCP 与 API 共用限流等保护；勿在不可信网络明文传令牌（优先 HTTPS）。
- Tool 调用有**超时**（实现上约数秒级），长任务应在你方异步处理。

---

## 延伸阅读

- 仓库内实现说明：[internal/mcp/README.md](https://github.com/lin-snow/Ech0/blob/main/internal/mcp/README.md)
- [访问令牌](/docs/guide/accesstoken) · [事件推送](/docs/guide/webhook) · [评论系统](/docs/guide/comment)
