---
title: 常见问题
description: 部署、权限、存储与集成问答
---

若这里没有你的情况，可到 [GitHub Discussions](https://github.com/lin-snow/Ech0/discussions) 搜索或提问。

---

## 产品与定位

### Ech0 是什么？

轻量、开源、**自托管**的发布工具：用一条**时间线**发短文、链接和媒体，数据在你自己的服务器上，强调简洁阅读和自主可控。更完整的产品说明见 [产品概览](/docs/guide/overview)。

### Ech0 不是什么？

不是 Obsidian / Notion 那种**知识库或团队文档**；更接近「朋友圈 / 说说」式的**个人动态**，而不是专业笔记工作流。

### 收费吗？

开源免费，**AGPL-3.0**，无广告、无订阅绑架。

---

## 部署、数据与备份

### 备份和导入怎么做？

用面板里的 **快照导出 / 迁移导入**。平时请定期备份你映射的**数据目录**（例如 Docker 的 `/opt/ech0/data`）。若启用对象存储，媒体在存储后端，库在本地或按你的部署而定。详见 [数据管理](/docs/guide/datacontrol)。

### 存储结构、换桶、本地⇄对象迁移？

见 [对象存储](/docs/guide/s3) 与 [数据管理](/docs/guide/datacontrol)；操作前务必备份。

---

## 阅读、RSS 与互联

### 支持 RSS 吗？

支持；可用 RSS 阅读器订阅你的实例上提供的订阅地址（入口以当前界面为准）。

### 为什么别人看不到我的 Connect 头像？

在 **系统设置 → 服务地址** 填写**你自己实例**的完整 URL（必须带 `http://` 或 `https://`），与浏览器访问地址一致。

### 为什么 Connect 列表不全？

后端会逐个拉取实例信息；**连不上的实例会被丢弃**，只显示成功拉取到的。详见 [互联聚合](/docs/guide/federation)。

---

## 账号、权限与发帖

### 为什么我不能发帖？

当前版本下**发帖权限默认限制在高权限账号**。**第一个注册的账号**为 Owner；其他人要发帖需由管理员在设置里授权。首次部署请确认是否已用 Owner 登录。

### 权限模型复杂吗？

采用较轻量的模型（如 Owner / Admin / 普通用户等），目标是日常够用、少折腾；细节以设置页为准。

---

## 评论、登录与存储

### 怎么开评论？

在 **评论设置 / 评论管理** 中开启，并按需配置审核与验证码。详见 [评论系统](/docs/guide/comment)。

### Passkey / OAuth 怎么配？

见 [统一登录](/docs/guide/sso)。

### S3 怎么配？

在存储设置填写 Endpoint、密钥、桶等；`endpoint` 一般**不要**带 `http/https` 前缀；若前端要直接访问媒体，需配置好桶/CDN 的访问策略。详见 [对象存储](/docs/guide/s3)。

### MetingAPI 是什么？

给**音乐卡片**解析流媒体直链用的 API；不填可用默认解析；生产环境建议用你可控的端点。

---

## 自动化、API 与第三方

### AI 摘要、Webhook、API？

- Agent：**系统设置 → Agent**，见 [智能摘要](/docs/guide/agent)。
- Webhook：[事件推送](/docs/guide/webhook)。
- 通用 API：用 [访问令牌](/docs/guide/accesstoken) 调用，文档在 **`/swagger/index.html`**。
- AI 客户端（Cursor 等）连实例：见 [MCP 接入](/docs/guide/mcp)，需 **MCP（AI Agent）** 受众的令牌。

### 自动化或 AI 想发评论？

可使用 README 中介绍的 **集成评论接口**，配合 **访问令牌**（需含 `comment:write` 与 `integration` audience 等）。详见实例 **`/swagger/index.html`** 与 [评论系统](/docs/guide/comment)。

### 第三方集成平台免责声明

未经官方授权的第三方集成服务**不属于官方支持范围**；使用产生的问题由使用方与第三方自行承担。
