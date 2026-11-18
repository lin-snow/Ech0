# Ech0 — Product & Engineering Showcase

## 14 页内容稿 · 重写版

> **核心叙事：** 从个人表达的需求出发，介绍 Ech0 如何成为一个可自托管、可持续互动、可扩展的个人时间线，再通过工程实践与开源成果展示它是如何被构建和持续维护的。
>
> **建议时长：** 15–18 分钟。每页屏上文案保持克制，讲稿用于口头展开，不建议整段放进幻灯片。
>
> **产品定义：** A self-hosted personal microblog where your timeline can be shared, discussed, and fully owned.
>
> **Slogan：** Let your thoughts flow.

---

# Part I · The Story

## 01 · Let your thoughts flow.

### 屏上文案

**Ech0**

Let your thoughts flow.

一个属于自己的个人时间线。

Self-hosted · Open Source · Personal Timeline

GitHub: `github.com/lin-snow/Ech0`

### 讲述重点

Ech0 是一个开源、自托管的个人微博客。它让用户能够在自己的服务器上发布内容、维护个人时间线，并与访问者进行持续互动。

我希望做的不是又一个临时记录工具，而是一个可以长期运行、持续积累内容，并且由自己掌控的个人发布空间。

这一场分享会从产品出发，介绍它解决什么问题、用户如何使用，再展开背后的系统设计、工程实践和开源成果。

---

## 02 · The Need for a Personal Timeline

### 屏上文案

**表达应该持续发生，而不只是被记录。**

- 内容需要一个长期存在的归属地
- 时间线应该由发布者掌控
- 分享与互动不应该依赖单一平台

### 讲述重点

我们每天都会产生大量值得记录和分享的内容：一个想法、一段经历、一篇文章、一张照片，或者一个值得讨论的链接。

这些内容往往分散在不同平台和工具里。笔记工具擅长保存，社交平台擅长分发，但个人很难拥有一个既适合持续发布、又能长期维护，并且真正由自己控制的空间。

我关注的是记录之后的阶段：内容发布以后，如何形成一条持续更新的时间线；别人如何访问、阅读和回应；而这些内容又如何长期留在自己的掌控之中。

Ech0 就是从这个需求开始的。

---

## 03 · A Timeline You Own

### 屏上文案

**Ech0 is a personal timeline, hosted on your server.**

- **Personal** — 围绕个人表达与内容积累
- **Publishable** — 公开或按权限分享内容
- **Interactive** — 评论、Reactions 与分享
- **Self-hosted** — 数据与运行环境由自己掌控

### 讲述重点

Ech0 的核心是个人时间线。它以时间为主线，把短文、链接和多媒体内容组织在同一个持续更新的空间里。

它与传统笔记工具的侧重点不同：Ech0 不止关注内容如何被记录，也关注内容如何被发布、访问和讨论。

它也不试图成为一个完整的大型社交网络。产品围绕个人时间线展开，通过评论、Reactions、分享和 RSS 等能力，让内容具备轻量的互动与传播方式。

自托管则是这套体验的基础：用户可以选择自己的服务器、域名和存储方式，决定内容如何被保存和管理。

---

# Part II · The Product

## 04 · From Writing to Connection

### 屏上文案

**One timeline. Many ways to express.**

Write → Enrich → Publish → Interact → Discover

- Markdown 内容编辑
- 图片、音频、视频与链接卡片
- 时间线、评论、Reactions 与分享
- 标签、搜索与 RSS

### 讲述重点

Ech0 的主要体验围绕一条内容的完整生命周期展开。

用户可以通过 Markdown 编辑器发布内容，也可以加入图片、音频、视频、链接卡片等富媒体内容。发布后，内容进入个人时间线，访客可以阅读、评论、使用 Reactions，也可以通过分享和 RSS 持续关注。

不同内容形态共享同一套发布和浏览体验。用户不需要为了发一段文字、分享一个链接或展示一段媒体内容，切换到不同的产品模块。

这里建议使用真实产品截图展示三个场景：内容编辑、时间线浏览、评论与互动。让观众先看到产品，再理解功能。

---

## 05 · Your Data, Your Rules

### 屏上文案

**Ownership is built into the product.**

- 数据运行在自己的基础设施上
- 支持内容与附件的导出、备份和迁移
- 本地文件系统与 S3 兼容存储
- RSS 与开放接口，保留内容出口

### 讲述重点

自托管不应该只意味着「服务部署在自己的机器上」，还应该意味着数据可以被理解、备份和迁移。

Ech0 使用 SQLite 保存核心数据，文件存储通过 VireFS 统一抽象，可以使用本地文件系统，也可以接入 S3 兼容对象存储。

在数据流动方面，Ech0 提供快照导入导出和 Capsule 内容格式，并支持从 Memos 等来源迁移内容。RSS 和 API 则为外部阅读器与工具保留了开放的访问方式。

这些能力共同构成数据主权：用户不仅能够运行自己的实例，也能够备份数据、迁移实例，并在需要时把内容带走。

---

## 06 · Simple to Run, Easy to Keep

### 屏上文案

**A small footprint. A complete experience.**

```bash
docker run -d \
  -p 6277:6277 \
  -v ./data:/app/data \
  sn0wl1n/ech0:latest
```

- 单二进制发行，前端资源内嵌
- Docker / Docker Compose / Helm
- Linux、Windows 与 ARM 平台支持
- Web、CLI 与 TUI 多种使用入口

### 讲述重点

Ech0 的部署目标是降低长期运行和维护的成本。

对于个人实例，用户可以通过 Docker 快速启动，也可以直接运行对应平台的二进制。前端资源随应用一起构建和分发，不需要额外维护一套静态资源部署流程。

如果用户希望在 Kubernetes 环境运行，也可以使用 Helm 部署。除了 Web 界面，Ech0 还提供 CLI 和 TUI，适合脚本化管理或终端环境。

这里强调的不是单纯追求镜像体积，而是让安装、升级、备份和日常维护形成一条简单清晰的路径。

---

## 07 · Open by Default

### 屏上文案

**Your timeline should work with your tools.**

| Interface | What it enables          |
| --------- | ------------------------ |
| REST API  | 集成与自动化             |
| Webhook   | 事件通知与外部工作流     |
| CLI / TUI | 终端管理与脚本操作       |
| MCP       | 连接外部 AI 客户端       |
| Copilot   | 面向个人内容的检索与问答 |

### 讲述重点

Ech0 不希望用户只能通过 Web 页面使用自己的数据。因此，产品同时提供 API、Webhook 和 CLI 等接口，让时间线能够接入用户已有的工具和工作流。

MCP 面向外部 AI 客户端，允许在授权范围内发现和调用 Ech0 提供的能力。它解决的是外部工具如何以受控方式访问实例，而不是把数据库直接暴露出去。

Copilot 则是 Ech0 内部面向用户的 AI 功能，帮助用户检索和理解自己的内容。它与 MCP 的职责不同：一个提供产品内的 AI 体验，一个提供对外集成入口。

开放接口的价值不在于接口数量，而在于用户可以用更多方式访问、管理和延伸自己的时间线。

---

# Part III · The Engineering

## 08 · System Architecture

### 屏上文案

**A modular application, designed for a small footprint.**

- **Clients** — Vue 3 Web · CLI · TUI · External clients
- **Application** — Gin / HTTP · SSH · MCP
- **Core** — Handler · Service · Repository
- **Data** — SQLite · VireFS · Full-text / Vector Index
- **Infrastructure** — Busen · Async Tasks · Webhooks

**Clear boundaries. One deployable application.**

### 讲述重点

Ech0 采用 Go 构建后端，前端使用 Vue 3。整体以模块化单体的方式组织：多个业务能力在同一个应用中协作，但通过清晰的分层和接口控制依赖关系。

入口层包括 HTTP、终端交互以及 MCP。业务逻辑集中在 Service 层，数据访问由 Repository 层负责；底层使用 SQLite 保存核心数据，VireFS 统一处理文件存储，检索能力由全文和向量索引提供。

Busen 负责在应用内部连接业务事件与下游处理，例如 Webhook 投递、向量化和缓存更新等。

这种架构的重点是控制复杂度：对用户而言，它仍然是一个容易部署的应用；对开发者而言，模块边界让功能可以持续演进，而不必把所有逻辑堆在一起。

---

## 09 · Engineering the Foundations

### 屏上文案

**Extract abstractions from real product needs.**

**Busen · Event Bus**

- 统一业务事件分发
- 类型化订阅与同步 / 异步处理
- 将业务流程与下游消费者解耦

**VireFS · Storage Abstraction**

- 统一本地文件与 S3 兼容存储
- 业务侧使用一致的文件标识
- 降低存储后端切换对业务代码的影响

### 讲述重点

在持续开发过程中，有两类重复问题逐渐变得明显：业务事件需要被多个模块消费，文件存储也需要同时适配本地磁盘和对象存储。

第一个是 Busen。发布、评论等业务事件可能触发多个后续动作。如果把这些调用直接写进业务流程，随着订阅者增加，核心逻辑会越来越难维护。Busen 将事件发布和事件消费分开，让不同模块可以独立订阅自己关心的事件。

第二个是 VireFS。产品既要支持本地文件系统，也要支持 S3 兼容存储。如果业务代码直接依赖具体存储实现，后续切换会产生大量修改。VireFS 将存储差异收敛到统一接口，让业务侧使用一致的文件访问方式。

这两个抽象都来自实际业务中的重复成本。它们的目标不是增加架构层次，而是减少功能继续增长时的耦合。

---

## 10 · Reliability & Security

### 屏上文案

**Long-running software needs reliable boundaries.**

**Reliability**

- 异步任务与明确的任务状态
- 取消、失败处理与重试
- 数据快照、导入与迁移

**Security**

- OIDC / OAuth 2.0 / Passkey
- Scoped Token 与细粒度授权
- MCP / API 请求统一进行权限校验
- 出站请求安全控制

### 讲述重点

个人时间线不是一次性工具。它需要长期运行，也需要在升级、迁移和异常情况下保持数据可恢复。

导入、导出和快照等操作可能持续较长时间，因此 Ech0 将它们放入异步任务体系中管理。任务具有明确状态，并支持取消、失败处理和重试，避免把长时间操作绑定在单次 HTTP 请求上。

安全方面，Ech0 支持多种身份认证方式，包括 OIDC、OAuth 2.0 和 Passkey。对于 API 与 MCP 等程序化访问入口，则通过 Scoped Token 控制可执行的操作范围，并在统一入口进行授权校验。

此外，服务端出站请求也需要进行安全控制，降低不可信 URL 带来的 SSRF 等风险。

这一页要表达的是：数据主权不仅是部署位置的问题，也包括数据能否恢复、访问是否受控，以及系统能否安全地长期运行。

---

## 11 · AI, Connected to Your Data

### 屏上文案

**Bring AI to your timeline — without losing control.**

**Copilot · In-product experience**

- 面向个人内容的检索与问答
- RAG 检索增强生成
- ReAct 工具调用与多轮处理
- SSE 流式响应

**MCP · External integration**

- 为外部 AI 客户端提供标准化工具入口
- 通过令牌与作用域控制访问
- 将授权边界延伸到工具调用

### 讲述重点

Ech0 的 AI 能力分成两个层次。

第一层是产品内的 Copilot。用户可以围绕自己的时间线进行内容检索、总结和问答。系统先从实例内容中检索相关信息，再将上下文交给模型生成回答；对于需要进一步操作的任务，可以通过 ReAct 循环组织工具调用，并通过 SSE 将结果逐步返回。

第二层是 MCP。它面向外部 AI 客户端，让用户能够在其他工具中以受控方式访问 Ech0 的能力。MCP 不是 Copilot 的另一种界面，而是 Ech0 对外开放能力的一种集成协议。

两者共享一个关键原则：AI 能力应该建立在明确的数据访问与授权边界之上。用户可以选择接入方式，也应该能够控制哪些内容和操作可以被访问。

---

# Part IV · The Impact

## 12 · Growing in the Open

### 屏上文案

**Built in the open. Used by a growing community.**

- **2k+** GitHub Stars
- **150+** Forks
- **70k+** Docker Pulls
- **20+** Contributors

HelloGitHub 2025 · Go 类项目 Top 1  
被阮一峰《科技爱好者周刊》及开源社区收录推荐

### 讲述重点

Ech0 从 2025 年 3 月开始持续开发，并逐渐形成了一个有真实用户和外部贡献者参与的开源项目。

目前项目获得了 2k+ GitHub Stars、150+ Forks 和 70k+ Docker Pulls，也有 20 多位外部贡献者参与建设。项目曾获得 HelloGitHub 2025 年度榜单 Go 类项目 Top 1，并被技术周刊和开源社区推荐。

这些数字并不只是项目热度的展示。对一个自托管产品来说，更重要的是有人愿意部署、持续使用、提交反馈，甚至参与代码贡献。

开源项目的影响力最终要回到产品本身：它是否解决了真实问题，是否值得用户长期保留。

> 发布前请重新核验 Stars、Forks、Docker Pulls、贡献者数量及榜单名称，统一统计日期与口径。

---

## 13 · From Product Thinking to Engineering Ownership

### 屏上文案

**End-to-end ownership, from idea to operation.**

| Area         | Contribution                           |
| ------------ | -------------------------------------- |
| Product      | 产品定位、需求分析、功能边界           |
| Architecture | 系统设计、模块拆分、技术选型           |
| Engineering  | Go 后端、Vue 前端、CLI / TUI、基础设施 |
| Delivery     | CI/CD、部署、版本发布与迁移            |
| Community    | Issue、PR Review、文档与用户反馈       |

**Creator · Maintainer · Full-stack Engineer**

### 讲述重点

Ech0 是一个长期维护的独立开源项目。我承担的不只是某一个模块的开发，而是从产品定义到工程交付，再到社区维护的完整责任。

产品层面，需要判断目标用户是谁、核心体验是什么，以及哪些功能应该做、哪些应该保持克制。

架构层面，需要在可维护性、部署复杂度和未来扩展之间做取舍，并把这些取舍落实到模块边界和工程实现中。

实现层面，覆盖 Go 后端、Vue 前端、CLI、TUI、存储抽象、事件系统和 AI 集成。交付层面，还需要维护构建发布、部署方式、数据迁移和版本兼容。

项目让我持续经历一个完整循环：发现问题、定义产品方案、完成工程实现、发布给用户，再根据真实反馈继续迭代。这种端到端的 Owner 经验，是 Ech0 带给我的核心积累。

---

## 14 · Your Thoughts. Your Timeline.

### 屏上文案

**Your thoughts. Your timeline.**

Let your thoughts flow.

- Website: `ech0.app`
- GitHub: `github.com/lin-snow/Ech0`
- License: `AGPL-3.0-or-later`

### 讲述重点

Ech0 希望让个人表达拥有一个更长期、更自主的归属地。

它不要求用户改变表达习惯，也不要求用户加入一个新的社交网络。只需要部署自己的实例，就可以开始建立一条属于自己的时间线，并逐渐积累内容、连接读者和形成社区。

这也是我持续建设 Ech0 的原因：让个人发布和互动变得简单，同时让内容的控制权始终留在用户手中。

代码和文档都已经开放。欢迎试用、反馈，也欢迎参与贡献。

---

# Appendix · Technical Discussion

## A1 · Key Technology Choices

| Area                 | Choice             | Reason                         |
| -------------------- | ------------------ | ------------------------------ |
| Backend              | Go                 | 静态编译、并发模型与跨平台发行 |
| Web                  | Vue 3 / TypeScript | 组件化界面与类型约束           |
| Database             | SQLite             | 降低个人实例的部署与维护成本   |
| ORM                  | GORM               | 统一数据访问层                 |
| Dependency Injection | Google Wire        | 显式依赖组装                   |
| File Storage         | VireFS             | 统一本地与 S3 兼容存储         |
| Event System         | Busen              | 进程内类型化事件分发           |
| Frontend Styling     | UnoCSS             | 原子化样式与构建期生成         |
| AI Retrieval         | RAG / ReAct        | 面向实例内容的检索与工具调用   |

## A2 · Discussion Topics

**Why a modular monolith?**  
Ech0 的主要部署场景是个人或小型实例。模块化单体可以保留清晰的业务边界，同时避免微服务带来的部署、网络调用和运维成本。

**Why SQLite?**  
核心目标是降低个人实例的运行门槛。SQLite 让数据库成为实例目录中的一个文件，备份与迁移路径更简单。是否需要更高并发或多节点能力，应根据真实负载评估，而不是提前引入复杂依赖。

**Why build Busen instead of adopting a message broker?**  
Busen 解决的是单进程内的事件解耦，不承担跨进程消息持久化或分布式投递职责。它与 Kafka、RabbitMQ 等消息系统解决的问题不同，选择它是为了匹配当前应用边界和部署目标。

**Why build VireFS?**  
产品需要同时支持本地文件系统和 S3 兼容存储。统一抽象可以把后端差异限制在存储层，减少业务代码对具体实现的依赖。

**How are Copilot and MCP different?**  
Copilot 是 Ech0 内部面向用户的 AI 产品功能；MCP 是 Ech0 对外提供能力的集成协议。前者解决用户如何理解和使用自己的内容，后者解决外部 AI 客户端如何在授权范围内调用 Ech0。

**How does Ech0 protect user data?**  
通过自托管部署、数据导出与迁移能力、认证机制、Scoped Token 和统一授权校验，分别处理数据归属、数据可迁移性与访问控制问题。具体安全边界以当前版本实现和文档为准。

---

# 制作说明

1. **产品页优先使用真实界面截图。** 第 04 页至少展示编辑器、时间线和互动；不要用抽象插画替代产品本身。
2. **架构页只保留必要关系。** 展示客户端、入口、核心分层、数据层和基础设施，不要把所有包名与技术名词塞进图里。
3. **工程页使用统一叙事。** 先说遇到的实际问题，再说为什么这样设计，最后给出实现和效果；没有可靠数据时，不强行编造性能指标。
4. **影响力数据统一口径。** Stars、Forks、Docker Pulls、Contributors 应注明统计日期；提交数量不作为个人贡献的主要论据。
5. **避免绝对化表达。** 不使用「永远不会丢」「完全不出服务器」「唯一正确方案」等无法由架构本身保证的说法。
6. **把细节留给问答。** 主线负责讲清楚产品与关键决策，协议字段、状态机细节、基准测试和边界条件放在 Appendix 或 Q&A。