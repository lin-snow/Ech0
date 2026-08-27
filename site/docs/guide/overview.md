---
title: 产品概览
description: Ech0 是什么、适合谁、能做什么
---

# 产品概览

## Ech0 是什么？

Ech0 是**自托管的个人微博客**：内容在一条**时间线**上展示，可以分享、评论，数据在你自己的服务器上。你可以把它理解成「自己域名下的轻量动态站」——偏发布与阅读，而不是重型笔记或团队协作工具。

和 Memos 这类「先快速记下想法」的工具相比，Ech0 更侧重**记录之后**：把时间线公开或半公开地发出去，让别人能持续看到和互动。它不是双链知识库，也不是 Notion 式团队空间。

---

## 新手从这里开始

1. **想先跑起来**：按 [快速上手](/docs/start/getting-started) 从 Docker 到首次登录走一遍（约 10 分钟）。
2. **想直接部署**：看 [安装部署](/docs/start/installation)（Docker / Compose / 二进制 / Helm）。
3. **遇到问题**：先查 [常见问题](/docs/start/faq)，再搜 [GitHub Issues](https://github.com/lin-snow/Ech0/issues)。

---

## 谁适合用 Ech0

**更适合你**，若你希望：

- 在自己的机器或域名上发**短文、链接、图片**，形成一条对外可见的时间线
- **数据在自己服务器**，而不是托管在某一家的封闭云笔记里
- 需要 **RSS**、可选 **评论**、可选 **多实例互联**，但不想维护一套论坛或博客引擎

**不太适合**，若你主要需要：

- Obsidian 式**双链知识库**、卡片笔记工作流
- Notion 式**多人协作**、复杂权限与数据库
- **纯私密备忘**、完全不打算对外发布时间线

---

## 部署与功能索引

| 你想做…                         | 文档                                                                                                 |
| ------------------------------- | ---------------------------------------------------------------------------------------------------- |
| 第一次部署、端口与数据目录      | [安装部署](/docs/start/installation)                                                                 |
| 从零基础到能发帖                | [快速上手](/docs/start/getting-started)                                                              |
| 升级镜像或小版本迭代            | [版本更新](/docs/start/update)                                                                       |
| 多站合并时间线                  | [互联聚合](/docs/guide/federation)（Connect + `/hub`）                                               |
| 第三方登录与 Passkey            | [统一登录](/docs/guide/sso)                                                                          |
| 访问令牌、MCP（AI 接入）        | [访问令牌](/docs/guide/accesstoken) · [MCP 接入](/docs/guide/mcp)                                    |
| 跟自己的 Echo 对话、近期摘要    | [AI 问答](/docs/guide/chat) · [AI 模型与摘要](/docs/guide/agent) · [向量检索](/docs/guide/embedding) |
| 站点 Logo、页脚、头像与面板偏好 | [偏好设置与用户资料](/docs/guide/preferences)                                                        |
| 评论与审核                      | [评论系统](/docs/guide/comment)                                                                      |
| 附件上云                        | [对象存储](/docs/guide/s3)                                                                           |
| 备份与迁移                      | [数据管理](/docs/guide/datacontrol)                                                                  |
| 换实例、长期保存、做静态归档站  | [胶囊与静态站](/docs/guide/capsule)                                                                  |

---

## 能力一览（概要）

- **写作**：Markdown 时间线，图片、链接，以及音乐/视频等扩展卡片（见 [编辑指南](/docs/guide/editor)）。
- **外观与资料**：站点 Logo、页脚、服务地址、Meting、自定义 CSS/JS；个人头像与界面语言（见 [偏好设置与用户资料](/docs/guide/preferences)）。
- **数据**：默认数据在本地目录；附件可接 **S3 兼容**存储；支持快照导出与从 v3 等来源迁移；可导出人类可读的**胶囊**用于搬家与长期保存，并一键编译成[静态归档站](/docs/guide/capsule)。
- **互动**：内建评论；**RSS**；**Connect / Hub** 聚合多个实例的公开内容。
- **账号**：OAuth、Passkey；脚本与集成使用 **访问令牌** 调用 HTTP API，AI 客户端通过 **MCP** 连接实例（见 [访问令牌](/docs/guide/accesstoken)、[MCP 接入](/docs/guide/mcp)）。
- **AI**：用自然语言[跟自己的 Echo 对话](/docs/guide/chat)、生成[近期摘要](/docs/guide/agent)；可选开启[向量检索](/docs/guide/embedding)做语义搜索。需自备 LLM（OpenAI 兼容 / Anthropic）。
- **自动化**：**Webhook** 事件推送等，按需在设置中开启。

接口与字段以你实例上的 **`/swagger/index.html`** 为准；不同版本可能有差异。

---

## 许可与开源

**AGPL-3.0** 开源，无广告、无强制云服务。二次分发与网络服务提供请遵守 AGPL 义务（向用户提供对应源码等）。
