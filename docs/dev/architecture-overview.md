# Ech0 架构全景图

> 本文是交付给新接手者的**架构全景**：从一个进程如何启动讲起，逐层展开后端分层、业务领域、两大能力层（Agent / MCP）、事件驱动子系统、基础设施组件、`pkg/` 自研库，最后到前端与端到端数据流。读完应能回答："某个功能从请求进来到落库/落盘，经过了哪些模块、谁依赖谁。"
>
> 配套细节文档：DI 约定见 `CLAUDE.md`；Agent/Chat 见 `docs/dev/llm-chat-design.md`、`docs/dev/agent-toolcall-design.md`；MCP 接入见 `docs/usage/mcp-usage.md`；事件系统见本文 §9 + `CLAUDE.md`「Event bus (Busen)」；存储见 `docs/usage/storage-migration.md`；鉴权见 `docs/dev/auth-design.md`、`docs/dev/access-token-scope-design.md`；Job/Task 见 `docs/dev/job-runner-design.md`、`docs/dev/snapshot-design.md`。

---

## 1. 一句话总览

Ech0 是一个**自托管的轻量个人微博（时间线）平台**，以**单个 Go 二进制**同时提供 REST API 与内嵌的 SPA 前端。

| 维度 | 选型 |
| --- | --- |
| 后端语言 | Go 1.27+（CGO，因 SQLite + sqlite-vec） |
| Web 框架 | Gin |
| 依赖注入 | Google Wire（编译期生成，`internal/di/wire_gen.go`） |
| ORM / 存储 | GORM + SQLite（`gorm.io/driver/sqlite` + `mattn/go-sqlite3`），向量检索用 `sqlite-vec` |
| CLI | Cobra（`ech0 serve` / `ech0 tui` / `ech0 version` / `ech0 hello`） |
| 缓存 | Ristretto（进程内） |
| 定时任务 | gocron v2 |
| 对象存储 | AWS SDK v2（S3 兼容），经自研 `pkg/virefs` 抽象 |
| LLM 能力 | 自研 `internal/agent`（OpenAI Chat Completions / OpenAI Responses / Anthropic 三协议 + ReAct 工具循环） |
| 前端 | Vue 3 + Vite + TS + Pinia + Vue Router + vue-i18n + UnoCSS(Wind4) + markdown-it/Vditor |
| 自研库（`pkg/`） | busen（事件总线）、gocap（PoW 验证码）、virefs（文件系统抽象）、viewer（请求身份上下文） |

设计基调有两条贯穿全局（详见 `CLAUDE.md`）：

1. **轻量优先**：评估任何新功能时，"会不会让项目变重"是一票否决项。
2. **依赖向内指向纯词汇**：事件契约、配置规格、LLM 工具抽象等"纯定义"包不依赖任何基础设施/领域；基础设施依赖定义，而非反过来。

---

## 2. 顶层架构全景

```
  ┌── 客户端 ──────────────────────────────────────────────────────────────────────┐
  │  浏览器(Vue 3 SPA)   API 客户端   外部 LLM(Claude…经 MCP)   Webhook 接收方        │
  └──────┬───────────────────┬────────────────┬─────────────────────▲────────────────┘
         │ HTTP + SSE        │ REST           │ JSON-RPC(/mcp)       │ 出站 HTTP(事件投递)
  ┌──────▼───────────────────▼────────────────▼──────────────────────┼────────────────┐
  │                       internal/server (Gin)                       │                │
  │  路由组: /(SPA+/api/files) · /api(public) · /api(auth) · /api(optionalAuth) · /ws · /mcp │
  │  middleware: auth · scope · cors · origin · ratelimit · captcha · maintenance · nocache │
  └──────┬────────────────────────────────────────────────────────────────────────────┘
         │ handler.Bundle 分发
  ┌──────▼──────────────────── 业务领域层（handler→service→repository→model）─────────────┐
  │  内容与互动:  echo   comment   file   connect                                          │
  │  身份与配置:  user   auth   init   setting                                             │
  │  智能与检索:  embedding   copilot ──► [Agent 能力层]      mcp ──► [MCP 能力层]          │
  │  运维与观测:  dashboard   migrator   common   web                                      │
  └──┬────────────────┬─────────────────────┬──────────────────────────┬──────────────────┘
     │ 调基础设施      │ emit 事件           │ Agent 出站(§7)            │ MCP 入站(§8)
  ┌──▼────────────────▼──── 基础设施 / 运行时 ──┐  ┌──────────▼─────────┐  ┌──────────▼────────┐
  │ event/bus(Busen) storage(VireFS) cache     │  │ internal/agent     │  │ internal/mcp      │
  │ kvstore transaction job task captcha        │  │ Provider抽象+ReAct  │  │ JSON-RPC+Registry │
  │ visitor setting migrator …                  │  │ →OpenAIx2/Anthropic│  │ →领域 service     │
  └──┬──────────────────────────────────────────┘  └────────────────────┘  └───────────────────┘
     │ 事件路由 (by Go type，§9)
  ┌──▼──────────── 事件订阅者（异步 side effects）────────────────────────────────────────┐
  │ webhook.Dispatcher   subscriber.AgentProcessor   subscriber.EmbeddingProcessor   snapshot │
  └────────────────────────────────────────────────────────────────────────────────────────┘

  生命周期编排: internal/app（Component: server / job.Manager / task.Manager / EventRegistrar）
              Wire 在 internal/di 装配整张依赖图（§3、§4）
```

四条主轴贯穿全图，后面逐一展开：

- **请求主轴（同步）**：HTTP → middleware → handler → service → repository → DB（§5、§6）。
- **Agent 主轴（LLM 出站）**：copilot 域注入领域工具 → `agent.Run` ReAct 循环 → Provider → 外部 LLM（§7）。
- **MCP 主轴（LLM 入站）**：外部 LLM 经 `/mcp` JSON-RPC → scope 校验 → Registry → 领域 service（§8）。
- **事件主轴（异步）**：service `Notify` 事件 → Busen 按 Go 类型路由 → 订阅者执行 webhook / 向量索引 / 缓存失效 / 快照调度（§9）。
- **生命周期主轴**：`main` → `bootstrap` → `cli` → Wire `BuildApp` → `app.App` 启停所有 `Component`（§3、§4）。

---

## 3. 进程启动与生命周期

### 3.1 启动链路

```
cmd/ech0/main.go
  └─ bootstrap.Bootstrap()          // ① 加载 config 单例 ② 初始化 zap logger ③ 设置 host env 默认值
  └─ cmd.Execute()  (Cobra 分发)
       ├─ ech0 serve   → cli.DoServeWithBlock()
       │     ├─ 端口可用性检查
       │     ├─ di.BuildApp()       // ④ Wire 装配整张依赖图，返回 *app.App
       │     └─ app.Run()           // ⑤ 启动所有 Component，阻塞至信号
       ├─ ech0 (裸) / ech0 tui → cli.DoTui()
       ├─ ech0 version → cli.DoVersion()
       └─ ech0 hello   → cli.DoHello()
```

`internal/bootstrap` 在 Cobra 分发**之前**跑，保证 config / logger 在任何子命令里都已就绪。`internal/cli` 是各 CLI 动词的实现层。快照导入导出**没有** CLI 动词，只在 Web 管理面板「数据管理」里。

### 3.2 组件生命周期（internal/app）

`internal/app` 是一个**通用的组件生命周期编排器**，与具体业务无关。核心契约：

```go
type Component interface {
    Start(ctx) error
    Stop(ctx) error
}
// 可选 Namer 接口提供友好名字用于日志/错误
```

`app.App` 由 `app.ProvideOptions` 注册三个 `Component` + 若干生命周期 Hook：

| 阶段 | 动作 |
| --- | --- |
| **BeforeStart Hook** | ① `setting.Seed` 幂等播种系统设置 ② `EventRegistrar.Register` 把所有订阅者挂到 Busen ③ `job.Manager` 启动时清扫遗留 running 作业 |
| **Start（按序）** | `job.Manager` → `task.Manager` → `server.Server`（HTTP 监听） |
| **Run** | 阻塞，直到收到 OS 信号 |
| **Stop（优雅）** | HTTP graceful shutdown → 定时任务 StopHook（如 flush 访客快照）→ `EventRegistrar.Stop`（退订 + 排空实现 `Draining` 的订阅者，如 webhook worker pool） |

要点：**所有有状态的运行时都被收敛成 `Component`**，由 `app.App` 统一启停，没有散落各处的 `go func()`。

---

## 4. 依赖注入（Google Wire）

`internal/di` 是整个后端的"装配车间"。`wire.go`（手写、`//go:build wireinject`）声明 ProviderSet 与注入器，`wire_gen.go`（生成、勿手改）是实际代码。**改了构造函数 / 绑定，必须 `just wire`**，CI 跑 `just wire-check`。

### 4.1 ProviderSet 与注入器

```
                         ┌──────────────── BuildApp() ─────────────────┐
                         │ InfraSet  VisitorSet  StorageSet  ProvideSeederKV │
                         │ DomainSet  RuntimeSet  AppSet                 │
                         └───┬──────────────────────────────────────────┘
   InfraSet ───────────────►│ database / eventbus / cache / transaction
   StorageSet ─────────────►│ storage.Manager（进程级单例，含一份只读 KV 读 S3 设置）
   VisitorSet ─────────────►│ visitor.Tracker（进程级单例）
   DomainSet ──┬───────────►│ BuildHandlers   → handler.Bundle（14 个领域 Handler + MCP）
               ├───────────►│ BuildMiddlewares→ middleware.Deps
               ├───────────►│ BuildTasker     → task.Manager（Cleanup/Snapshot/VisitorSnapshot）
               ├───────────►│ BuildJobManager → job.Manager（Reindex/Migration/Export Runner）
               └───────────►│ BuildEventRegistrar → 订阅者注册表
   RuntimeSet ─────────────►│ server.Server
```

### 4.2 三个被刻意"顶层引入一次"的共享单例

Wire 默认会为每个 Build 各生成一份实例，对**有状态**基础设施是 bug 源。以下三者必须在 `BuildApp`/`BuildServer` 顶层注入一次、统一下沉：

- **`visitor.Tracker`（VisitorSet）**：否则 WebHandler 写入 #1、Tasker 从 #2 读出恒为 0。
- **`storage.Manager`（StorageSet）**：否则设置页改了 S3 只 reload 了自己那份 Manager，文件服务仍用旧后端。
- **`job.Manager`（BuildJobManager）**：Runner 在构造期一次性注册；它依赖的 `EmbeddingService` / `migrator.ImportEngine` 都不含 `*job.Manager`，故无构造环。

`HandlerSet` 里还有一处跨域绑定值得记住：**copilot 的 `UserReader` 绑定到 `user` 服务**（取当前对话用户的展示名 + 按作者收口检索）。

---

## 5. 分层后端：handler → service → repository → model

### 5.1 严格四层 + 别名约定

每个领域在四层各有一个并列包：`internal/{handler,service,repository,model}/<domain>`。

```
   HTTP ─► handler/<x>            薄壳：解析请求、鉴权上下文、调 service、组装 response DTO
            │  (xxxHandler 别名)
            ▼
          service/<x>            业务编排：事务、缓存、跨域协作、emit 事件
            │  (xxxService 别名)
            ▼
          repository/<x>         数据访问：GORM 查询，封装 SQL/事务细节
            │  (xxxRepository 别名)
            ▼
          model/<x>             实体 + 请求/响应 DTO + 枚举常量
            (xxxModel / xxxUtil 别名)
```

**跨层导入必须用别名**：`xxxHandler` / `xxxService` / `xxxRepository` / `xxxModel` / `xxxUtil`（现有代码强制，新代码照做）。

### 5.2 HTTP 路由装配（internal/router）

`server` 把 `handler.Bundle`（14 领域 Handler 聚合）+ `middleware.Deps` 交给 `router.SetupRouter`。路由按"模块"注册，分**核心模块**（模板/静态文件/中间件/路由组）与**功能模块**（各领域）：

| 路由组 | 前缀 | 鉴权策略 |
| --- | --- | --- |
| `ResourceGroup` | `/` | 内嵌 SPA + `/api/files` 静态资源（带 StaticFileSecurity） |
| `PublicRouterGroup` | `/api` | 公开，无需 token |
| `AuthRouterGroup` | `/api` | `RequireAuth`：缺失/无效 token 一律 401 |
| `OptionalAuthRouterGroup` | `/api` | `OptionalAuth`：匿名可读，带 token 则按用户身份（管理员见更多） |
| `WSRouterGroup` | `/ws` | WebSocket（如系统日志流） |
| `MCPRouterGroup` | `/mcp` | `RequireAuth`，承载 MCP 协议（§8） |

两条需要长期保护的特例：
- **`POST /api/comments/integration`**：刻意绕过 captcha/form-token，靠 `comment:write` scope + `integration` audience 的 access token 鉴权。
- **`/api/files`** 静态路由服务本地内容；`stream` 路由是鉴权的。

---

## 6. 业务领域全景

后端共 **15 个业务领域**，按职能聚成四簇。每个领域（除少数纯模型域）都在 handler/service/repository/model 四层各有一个并列包。

### 6.1 内容与互动

| 领域 | 业务职责 | 关键关系 |
| --- | --- | --- |
| **echo** | 核心时间线条目（微博 post）：内容、媒体、布局（瀑布/网格/轮播）、公开/私密、音乐/视频/网站/位置扩展、标签 | 被 comment/file/embedding/copilot/migrator 引用；emit `Echo{Created,Updated,Deleted}` |
| **comment** | 访客评论：审核流（待审/通过/拒绝）、反垃圾（蜜罐/captcha/form-token）、公开投影脱敏 | 关联 echo；emit `Comment{Created,StatusUpdated,Deleted}` |
| **file** | 媒体资产：多后端存储（本地/S3）、临时文件生命周期、元数据、EchoFile 关联排序 | 经 storage.Manager 落盘；emit `ResourceUploaded` |
| **connect** | 实例互联（联邦）：发现远端 Ech0、健康检查、聚合时间线 | 独立子系统 |

### 6.2 身份与配置

| 领域 | 业务职责 | 关键关系 |
| --- | --- | --- |
| **user** | 用户资料与账号：身份、语言偏好、owner/admin 角色、头像 | 被 auth/copilot/dashboard 依赖；emit `User{Created,Updated,Deleted}` |
| **auth** | 鉴权基础设施：JWT（access+refresh）生命周期、OAuth2/OIDC（GitHub/Google/QQ/自定义）、Passkey/WebAuthn、token 撤销黑名单、access token 的 scope/audience | 给 middleware 提供 `TokenRevoker`；scope 体系被 MCP（§8）复用 |
| **init** | 系统初始化：一次性引导流程、owner 账号建立检测 | — |
| **setting** | 系统配置（落 KV）：站点品牌、OAuth2/S3/Passkey、Agent/embedding 配置、CORS、自定义 CSS/JS | 被所有域消费；改快照计划时 emit `UpdateSnapshotSchedule` |

### 6.3 智能与检索

| 领域 | 业务职责 | 关键关系 |
| --- | --- | --- |
| **embedding** | 向量检索基础设施：EchoEmbedding 元数据 + 内容快照、sqlite-vec 索引、内容哈希去重、模型/维度状态 | 索引 echo；被 copilot 检索；核心客户端在 `internal/embedding` |
| **copilot** | AI 对话助手（基于 echo 检索的 RAG）：LLM 驱动检索增强对话、时间窗摘要、量化统计、多模态图片注入 | **核心是 Agent 能力层（§7）**；绑定 embedding（检索）+ user（作者）+ echo + file（图片） |
| **mcp** | Model Context Protocol 服务端：把内部 API 暴露为 MCP tools/resources，供外部 LLM 调用 | **核心是 MCP 能力层（§8）**；经 viewer 取鉴权上下文，调各域 service |

### 6.4 运维与观测

| 领域 | 业务职责 | 关键关系 |
| --- | --- | --- |
| **dashboard** | 管理观测：访客统计（7 日热力）、系统日志（WS 流式）、运行指标 | 聚合 visitor / echo / user |
| **migrator** | 数据导入导出：上传快照、job 异步处理、状态轮询、产物下载 | 跨 echo/user/comment/file 批量读写；引擎在 `internal/migrator` |
| **common** | 共享件：全局错误/成功 DTO、分页、KeyValue 仓储、热力图工具、枚举常量 | 被各域复用 |
| **web** | SPA 托管：服务内嵌 Vue 前端、SPA fallback（避开 /api、/ws、/mcp、/swagger）、访客记录 | — |

> 注：`copilot` / `migrator` / `mcp` 的 service 层是"薄层"——真正的能力沉在专门的核心包里：copilot → `internal/agent`（§7），mcp → `internal/mcp`（§8），migrator → `internal/migrator`（导入导出引擎）。`agent` / `embedding` / `webhook` 这些核心包**不属于分层四件套**，是独立的能力/基础设施包。

---

## 7. Agent 能力层（internal/agent）

> **一句话**：`internal/agent` 是 Ech0 的 LLM 核心，把多家协议（OpenAI Chat Completions / OpenAI Responses / Anthropic）的生成能力收口为统一的 **Provider 抽象**，并提供一个 **ReAct 工具循环**，让模型在一轮对话内自主决定是否检索、检索几次。**领域零依赖**——它不 import `echo`/`embedding`，工具由上层（copilot service）注入。
>
> 方向是**出站（outbound）**：Ech0 作为 LLM 宿主，主动调用自己的领域工具。与之镜像的是 MCP 的入站（§8）。

### 7.1 分层与数据流

```
┌─────────────── internal/service/copilot（领域层，注入工具 + 消费事件）──────────────┐
│  ChatService.AskStream  → 组 system prompt + 历史 + question                         │
│     注入 3 个领域工具（Def=JSON Schema + Execute=领域闭包）：                         │
│       · search_echos     点查：top-k 向量检索（embedding.Search）                    │
│       · summarize_echos  聚合：区间穷举全部 Echo，写跨度总结（年终/季度/月度）        │
│       · stats_overview   量化：区间精确统计（纯 SQL）                                │
│  SummaryService.GetRecent → 非流式近期总结（被 MCP get_recent 复用）                 │
└───────────────┬──────────────────────────────────────────▲──────────────────────────┘
                │ agent.Run(ctx, RunRequest)                │ <-chan AgentEvent
                ▼                                            │ (AgentDelta/Searching/ToolResult/Done/Error)
┌─────────────── agent.Run —— ReAct Loop（run.go）─────────────────────────────────────┐
│  for round < maxRounds:                                                               │
│     provider.Stream(req) ─► 消费 Event 流                                             │
│        · EventTextDelta → 透传为 AgentDelta（实时上屏，跨轮连续）                     │
│        · EventToolCall  → 收集，本轮结束后执行                                        │
│     本轮有工具调用 → 去重 → 并发执行(≤maxParallelTools) → 追加 tool 结果回消息        │
│                     → emit AgentSearching / AgentToolResult → continue               │
│     否则 → emit AgentDone → break                                                     │
│  工具轮用尽 → 强制一轮「不给工具」收尾（保证作答，避免「只检索不回答」）              │
│  护栏：maxRounds · 同 turn 查询去重 · MaxContextTokens 软上限回收最旧工具结果 · Timeout · ctx 取消即停 │
└───────────────┬──────────────────────────────────────────▲──────────────────────────┘
                │ provider.Complete / Stream(Request)        │ <-chan Event
                ▼                                            │ (TextDelta/ToolCall/Done/Error)
┌─────────────── Provider 抽象（provider.go，脏活下沉）────────────────────────────────┐
│  providerFor(setting.Protocol)：                                                      │
│    · OpenAI 兼容（openaiProvider）：Chat Completions，OpenAI/DeepSeek/Qwen/Moonshot/Ollama… │
│        流式 tool_call 按 index 跨 chunk 累积 arguments（toolCallAccumulator）；       │
│        prompt cache 由服务端自动命中（前缀 >1024 token），无需客户端字段             │
│    · OpenAI Responses（openaiResponsesProvider）：/v1/responses 语义事件流；          │
│        工具调用从 output_item.done 的完整 item 取（call_id 只在 item 上）；           │
│        input 是扁平项数组（function_call / function_call_output 独立成项）           │
│    · Anthropic（anthropicProvider）：真流式 Messages.NewStreaming；                   │
│        tool input_json_delta 借 SDK Message.Accumulate 拼装；连续 tool 结果合并进单条 user │
│    · 未知协议（含已下线的 gemini）→ AGENT_PROTOCOL_NOT_FOUND                          │
└───────────────────────────────────────────────────────────────────────────────────┘
                                  │ 各家官方 SDK
                                  ▼  外部 LLM API（/v1/chat/completions · /v1/responses · /v1/messages）
```

### 7.2 关键抽象（types.go）

| 类型 | 作用 |
| --- | --- |
| `Provider` | 协议适配接口：`Complete`（非流式）+ `Stream`（语义 `Event` 流）。各家 SDK 差异（尤其流式 tool_call 分片拼接）封死在实现内部 |
| `Tool{Def, Execute}` | 工具 = 声明（`ToolDef`：名称+描述+JSON Schema）+ 执行闭包。**执行体由领域层注入**，agent 包零领域依赖 |
| `ToolOutput{Content, Meta, Images}` | 工具结果：`Content` 回喂模型；`Meta` 旁路带出领域数据（如命中检索结果 → SSE sources）；`Images` 非空则追加一条带图 user 消息（多模态） |
| `ImagePart` | 随消息发给多模态模型的图片：优先 `Base64`（自部署/私有存储 provider 拉不到内网 URL），`URL` 仅用于公开直链 |
| `Event{Kind}` | **Provider→Loop**：`EventTextDelta` / `EventToolCall` / `EventDone` / `EventError`。Provider 不懂业务，只吐文本增量与工具调用 |
| `AgentEvent{Kind}` | **Loop→Service**：`AgentDelta`(上屏) / `AgentSearching`(决定调工具) / `AgentToolResult`(Meta→sources) / `AgentDone` / `AgentError`。语义翻译（searching/sources）在此完成 |
| `RunRequest` | Loop 对领域层暴露的请求：`Setting`/`Messages`/`Tools`/`MaxRounds`/`Temp`/`Timeout`/`MaxContextTokens`/`Strings` |
| `RunStrings` | 工具循环里回喂/注入的少量提示文案，由领域层（知 locale）注入，使 agent 包 **i18n 零依赖**；留空回退中文默认 |

### 7.3 两个入口

- **`Generate`**：非流式、无工具。用于近期总结（summary）。`SummaryService.GetRecent` 即走它。
- **`Run`**：ReAct 工具循环（function calling）。用于 Chat。`ChatService.AskStream` 消费它吐出的 `AgentEvent`，翻译成 SSE：`searching | sources | delta | done | error`（+15s keep-alive）。

### 7.4 设计红线

1. **脏活下沉、语义上浮**：SDK 差异只许存在于 Provider 内部；Loop 与 Service 只消费统一的 `Event` / `AgentEvent`。
2. **领域零侵入**：`agent` 包不得 import `embedding`/`echo`；工具以「定义 + 执行闭包」由 Service 注入。
3. **错误分级**：工具执行错误 → 包装成 tool 结果回喂模型自愈；传输/协议错误 → 中止并 `AgentError`。**绝不静默吞错**。
4. **护栏先行**：轮数上限、查询去重、token 预算、超时——都在 Loop 层强制，防模型死循环烧 token。

---

## 8. MCP 能力层（internal/mcp）

> **一句话**：`internal/mcp` 是 Ech0 内置的 **Model Context Protocol 服务端**，把领域能力以标准 **tools/resources** 暴露给外部 LLM（如 Claude Desktop / IDE）。挂在 `/mcp` 路由组（`RequireAuth`），走 **JSON-RPC 2.0**，每个 tool/resource 声明所需 **scope**，用 access token 的 scope 集合做细粒度授权。
>
> 方向是**入站（inbound）**：外部 LLM 把 Ech0 当成工具集来调用。与 Agent 的出站（§7）正好镜像——两者都是 function calling，但方向相反。

### 8.1 请求处理链

```
外部 LLM (Claude…)
   │ JSON-RPC 2.0 over HTTP POST  /mcp
   ▼
middleware.RequireAuth ─► 解出 viewer.Context（带 token 的 scopes / audience）
   ▼
mcp.Server.dispatch(method)：
   ├─ initialize              → ServerCapabilities + ServerInfo（能力握手）
   ├─ notifications/initialized
   ├─ tools/list              → Registry.ToolDefinitions()（列全部工具声明）
   ├─ tools/call              → handleToolsCall：
   │        ① Registry.LookupTool(name) 取 handler + 所需 scopes
   │        ② 校验 viewer 是否具备全部 scope（不足 → insufficient scopes 错误）
   │        ③ 执行 handler → Adapter → 领域 service
   ├─ resources/list          → Registry.ResourceDefinitions()
   └─ resources/read          → handleResourcesRead（同样先校验 scope）
   ▼
internal/mcp/Adapter（adapter_*.go）── 注入 8 个领域 service：
   echo / user / comment / file / common / connect / copilot(SummaryService) / setting
   RegisterAll 依次调用：
     registerEcho{Tools,Resources} · registerUser/Comment/File/Common/ConnectResources
     registerAgentTools（get_recent，复用 copilotSvc.SummaryService）
     registerWebhookTools（list/create/update/delete/test，全部路由到 settingSvc，scope=admin:settings）
   每个 register* 把 service 方法包装成 ToolHandler / ResourceHandler 注册进 Registry
```

辅助：`RawTokenFromContext` / `BaseURLFromContext` 从 ctx 取原始 token 与外部 base URL（如 file_upload_guide 需要拼出可访问 URL）。

### 8.2 暴露的 Tools（动作 / 读写）

| 域 | 工具 | 所需 scope |
| --- | --- | --- |
| echo | `search_posts` · `get_post` · `list_tags` · `get_today_posts` | `echo:read` |
| echo | `create_post` · `update_post` · `delete_post` · `like_post` · `delete_tag` | `echo:write` |
| comment | `list_comments` | `comment:read` |
| comment | `create_comment` · `create_integration_comment` | `comment:write` |
| file | `list_files` · `get_file` | `file:read` |
| file | `create_external_file` · `delete_file` | `file:write` |
| connect | `list_connects` · `get_connects_info` | `connect:read` |
| connect | `add_connect` · `delete_connect` | `connect:write` |
| webhook | `list_webhooks` · `create_webhook` · `update_webhook` · `delete_webhook` · `test_webhook` | `admin:settings` |
| agent | `get_recent`（AI 近期总结，复用 copilot `SummaryService`） | `echo:read` |

### 8.3 暴露的 Resources（只读上下文 / 指南，URI 寻址）

| 域 | 资源 | 所需 scope |
| --- | --- | --- |
| echo | `tags` · `recent_posts` · `post` | `echo:read` |
| user | `profile` | `profile:read` |
| comment | `recent_comments` · `integration_comment_guide` | `comment:read` |
| file | `file_upload_guide` | `file:read` |
| connect | `connect_self` | `connect:read` |
| common | `heatmap` | `echo:read` |

> **Tools vs Resources**：tools 既含读也含写/动作（会改状态）；resources 是只读上下文（profile、heatmap、各类 guide），供 LLM 取背景信息。两者的 scope 都在 `Registry` 注册时静态声明，`Server` 在 call/read 前强制校验——**鉴权不在业务里散落，统一收口在 dispatch**。scope 体系本身由 auth 域定义（§6.2、`docs/dev/access-token-scope-design.md`）。

### 8.4 Agent ↔ MCP 对照

| 维度 | Agent（§7，出站） | MCP（§8，入站） |
| --- | --- | --- |
| 谁是 LLM 宿主 | Ech0 内部（copilot） | 外部（Claude 等） |
| 谁定义工具 | copilot service 注入领域闭包 | Adapter 把领域 service 注册进 Registry |
| 协议 | OpenAI Chat Completions / OpenAI Responses / Anthropic SDK | JSON-RPC 2.0（MCP 规范） |
| 鉴权 | 站内会话用户 | access token 的 scope 集合 |
| 典型场景 | 站内 Chat 问答 / 总结 | 把 Ech0 接进外部 AI 客户端读写 |

---

## 9. 事件驱动子系统（异步 side effects）

这是 Ech0 解耦"领域写"与"副作用"的关键，建立在自研 `pkg/busen` 之上。一条规则：**依赖向内指向纯词汇包**。

```
   ┌─────────────── internal/event (纯词汇，只 import 领域 model) ──────────────┐
   │  事件结构体 + 自描述方法：EventName() / OrderingKey()                       │
   │  EchoCreated/Updated/Deleted · UserCreated/Updated/Deleted                 │
   │  CommentCreated/StatusUpdated/Deleted · ResourceUploaded                   │
   │  SystemSnapshot · SystemExport · UpdateSnapshotSchedule                    │
   │  WebhookObservation（中立观测快照：Topic/EventName/Payload/Meta/OccurredAt）│
   └───────────────────────────────▲───────────────────────────────────────────┘
                                    │ 只认类型，不认 busen
   ┌──────────── internal/event/bus (eventbus，基础设施) ──────────────────────┐
   │  *busen.Bus 单例 · 泛型 Emit[T]/Notify[T]/On[T]/OnWithMeta[T]              │
   │  订阅预设: AsyncParallel（并行、无序）/ AsyncSequential（单 worker、FIFO）  │
   │  EventRegistrar：BeforeStart 注册全部订阅、AfterStop 退订 + 排空 Draining   │
   └──────┬──────────────────────────────────────────────────────────▲────────┘
          │ 路由 by Go type                                            │ Notify/Emit
   ┌──────▼──────────────── 订阅者 ───────────────────────┐   ┌────────┴──────── 生产者 ────────┐
   │ webhook.Dispatcher  ── OnWithMeta 13 类事件          │   │ service/echo   → Echo*           │
   │   → 转 WebhookObservation → worker pool → Sender 投递 │   │ service/user   → User*           │
   │ subscriber.AgentProcessor  ── Echo*/UserDeleted      │   │ service/comment→ Comment*        │
   │   → 清 agent 摘要缓存（AsyncParallel）               │   │ service/file   → ResourceUploaded│
   │ subscriber.EmbeddingProcessor ── Echo*               │   │ setting(snapshot)→ UpdateSnapshot│
   │   → 增量向量索引 IndexEcho/RemoveEcho（AsyncParallel）│   │ job/runner/export→ SystemSnapshot │
   │ snapshot scheduler ── UpdateSnapshotSchedule         │   │ task/scheduled  → SystemSnapshot  │
   │   → 重载 cron 计划（AsyncSequential）                │   │ migrator        → SystemExport    │
   └──────────────────────────────────────────────────────┘   └──────────────────────────────────┘
```

要点：
- **路由完全靠 Go 类型，没有 topic 维度**；事件用 `EventName()` 自描述对外的稳定 webhook 名。
- 生产者用 `eventbus.Notify(ctx, bus, event.EchoCreated{...})`（best-effort，失败仅 warn 日志）；要拿到 error 时用 `Emit`。**没有 publisher facade**。
- `webhook.Dispatcher` 本身就是一个订阅者，用 `OnWithMeta`（带元数据的 `On`）把每个可观测事件桥接成中立的 `WebhookObservation`，再经 worker pool 异步投递。
- 加跨切面副作用时，**优先发事件，而不是在 handler 里直接调服务**。
- Busen 的异步队列是 best-effort（关停时丢弃）；运行时调参经 `ECH0_EVENT_*` 环境变量。

---

## 10. 基础设施与运行时组件

按"是否是 `app.Component`（有启停生命周期）"分两类。

### 10.1 生命周期组件（被 app.App 启停）

| 模块 | 角色 | 关键符号 |
| --- | --- | --- |
| `internal/server` | 薄 Gin HTTP `Component` | `ProvideHTTPServer`（装路由+中间件）、`Start`(监听) / `Stop`(graceful) |
| `internal/job` | 长任务框架：Submit→goroutine 跑 Runner→落库 + 内存进度 + 取消 | `Manager`、`Runner`、`ReportFunc`、`JobRepository`；类型 `TypeReindex/TypeMigration/TypeExport`（`job/runner` 为具体 Runner） |
| `internal/task` | 定时任务（gocron）：`Manager` 持有 `Task` 列表 | `Task.Schedule`、`StopHook`；`task/scheduled` 提供 Cleanup/Snapshot/VisitorSnapshot |
| `internal/event/bus` 的 `EventRegistrar` | 订阅生命周期：BeforeStart 注册、AfterStop 退订+排空 | 见 §9 |

### 10.2 无生命周期的基础设施 / 单例

| 模块 | 角色 | 关键符号 / 备注 |
| --- | --- | --- |
| `internal/config` | env 配置单例（caarlos0/env） | `config.Config()`（sync.Once），见 `.env.example` |
| `internal/database` | GORM+SQLite 初始化、自动迁移、写锁、热切换 | `GetDB/SetDB`(atomic)、`MigrateDB`、`HotChangeDatabase`（快照用）、`EnableWriteLock` |
| `internal/cache` | 泛型缓存接口 + Ristretto 实现 | `ICache[K,V]`、`NewCache` |
| `internal/kvstore` | KV 抽象，两实现 | `Store` 接口、`Memory`(易失) / `Persistent`(落库，包 keyvalue repo)；字段命名 `durableKV` / `ephemeralKV` |
| `internal/transaction` | 事务抽象 | `Transactor.Run`、`GormTransactor` |
| `internal/storage` | 本地/S3 统一文件抽象（基于 `pkg/virefs`），**有状态单例** | `Manager`、`StorageSelector`、`S3SettingStore`(从 KV 读 S3 设置)、`ReloadFromConfigAndDB`、`ApplyS3Setting` |
| `internal/i18n` | 后端请求 locale 解析与翻译引擎（`go-i18n/v2`）：Gin 中间件 `Middleware()` 在每次请求时解析 locale，优先级：`lang` 查询参数 / `X-Locale` 请求头 → 用户偏好（DB）→ 站点默认（KV）→ 兜底 `en-US`；嵌入 4 种语言包（zh-CN / en-US / de-DE / ja-JP）；`ApplyUserLocaleFromUserID` 供 Auth 后升级当前请求 locale；内置 CVE-2022-32149 输入防护 | `ResolveLocale`、`NewLocalizer`、`Localize`、`Middleware`、`ApplyUserLocaleFromUserID`；被 `internal/middleware`（注册 locale 中间件）及 copilot `RunStrings` 使用 |
| `internal/middleware` | 中间件聚合 | `Deps{TokenRevoker}`；实现：auth/scope/cors/origin/ratelimit/maintenance/nocache/staticfile |
| `internal/captcha` | PoW 验证码（包 `pkg/gocap`），进程级共享 engine | `SiteVerify`、`NewHTTPHandler`（挂在 `/api`） |
| `internal/visitor` | PV/UV 追踪器，**actor 模型**（单 goroutine 改状态） | `Tracker.Record/Last7Days/Today/Load`、`DayStat`；由 `task/scheduled.VisitorSnapshot` 落库 |
| `internal/setting` | 配置引擎：`Spec[T]`(key+default+normalize/migrate) + 注册表 + 播种 | `Get[T]/Set[T]/Seed`；启动时由 `app.ProvideOptions` 调 Seed |
| `internal/migrator` | 导入导出引擎（两段式） | `ExportEngine`/`ImportEngine`；子包 `exporter/{fs,s3}`、`importer/{ech0,memos}`、`snapshot`、`spec`（契约） |
| `internal/agent` | LLM Provider 抽象 + ReAct loop（详见 §7） | `agent.Run`、`Generate`；Provider 适配 OpenAI Chat Completions / OpenAI Responses / Anthropic |
| `internal/mcp` | MCP JSON-RPC 服务端（详见 §8） | `Server.ServeHTTP`、`Registry`、`Adapter` |
| `internal/embedding` | 向量/RAG embedding 客户端（OpenAI 兼容 `/v1/embeddings`） | `Embed/EmbedOne`；service 层有 `Indexer`、`Search`、`Backfill` |
| `internal/util/*` | 横切工具：log(zap 包装) / crypto / jwt / img / md / timezone / async / egress / uuid / github / tui / cookie / err / format / url ... | 日志须带 `module` 字段，见 `docs/dev/logging.md` |

**无反向依赖的红线**：`setting`/`kvstore`/`app` 从不 import `service`/`handler`；`job`/`task` 的 runner 可 import service，但 `job`/`task` 核心不 import 具体 runner（靠 `Register` 发现）；`migrator` 核心引擎不依赖 `job`；`agent` 不依赖任何领域包。

---

## 11. `pkg/` 自研库

四个进程内自研库，沉淀通用能力，与业务解耦（除 virefs 用 AWS SDK 外基本零外部依赖）。

| 库 | 一句话 | 核心 API | 被谁用 |
| --- | --- | --- | --- |
| **busen**（`router`/`dispatch`） | 类型安全、有界、异步的进程内事件总线，支持 topic 路由、中间件、可观测 hook | `Bus`、`Bus.Publish[T]`/`Bus.Subscribe[T]`（Go 1.27 泛型方法）、`Bus.SubscribeTopic`、`Event[T]`、`Hooks`、`Shutdown(mode)` | `internal/event/bus` 封装为领域事件总线 |
| **gocap**（`cap`/`core`/`store`/`transport`） | 内嵌的 PoW 验证码引擎（challenge→redeem→siteverify），内存态 + 限流 + 可插存储 | `cap.Engine`(`Handler()`/`SiteVerify()`/`RegisterSite()`)、`core.Service`、`store.Store` | `internal/captcha` |
| **virefs**（`plugin/zip`） | 基于 key 的统一文件系统，覆盖本地盘与 S3，支持中间件、迁移、多后端路由 | `FS` 接口、`LocalFS`/`ObjectFS`、`MountTable`/`Schema`、`Migrate`、`Copier/Presigner/BatchDeleter` | `internal/storage`、`internal/migrator/snapshot` |
| **viewer** | 请求级身份/鉴权上下文抽象（user/token/scope/audience） | `Context` 接口、`NewUserViewer*`、`WithContext`/`FromContext`/`MustFromContext` | auth 中间件、comment/user/file/copilot service、mcp、scope 中间件 |

---

## 12. 前端（web/）

Vue 3 SPA，生产时由后端 `internal/handler/web` 内嵌 `template/dist` 提供；开发时 Vite 跑 `:5173` 并把 `/api` 代理到后端 `:6277`。

```
web/src/
  components/    可复用 SFC：按钮/表单/对话框(BaseDialog)/弹窗/图标/布局件
  views/         页面级组件（对应路由）
      ├─ home        公开时间线 + 发布
      ├─ panel/*     管理面板：dashboard / setting / user / storage /
      │              data-management / sso / extension / comment / advance / system-log
      ├─ auth        登录/注册/OAuth2/Passkey
      ├─ chat        Copilot 对话（SSE）
      ├─ hub/zen/echo/init/widget/404
  stores/        Pinia（组合式）：auth/user/echo/editor/theme/setting/connect/hub/init/migration/reindex/zen
  router/        Vue Router：懒加载、meta 守卫(requiresAuth/optionalAuth/noindex)、滚动行为
  service/
      ├─ api/        REST + SSE 绑定：echo/auth/chat(SSE)/comment/file/setting/user/embedding/init/dashboard/agent/connect
      └─ request/    统一 fetch 封装 + SSE 流解析 + 公共头(Authorization/locale) + 错误归一化
  locales/       vue-i18n：messages/ 下 zh-CN/en-US/de-DE/ja-JP，缺省回退 en-US
  composables/   useBaseDialog / useBfCacheRestore / useSeoHead
  utils/         echo/file/image/timeValue/toast/tokenSize/tweet/cron/storage/loadExternalAsset
  typings/       领域 DTO 的 TS 类型声明
```

- **公开侧**（home/hub/zen/echo）：只读浏览时间线、联邦发现。
- **管理侧**（panel/\*）：仪表盘、设置、用户、存储、数据迁移、评论审核、系统日志、高级配置。
- **后端通信**：REST `/api/*`（带鉴权头）；Chat 用 SSE（`searching/sources/delta/done/error` 事件流，对应 §7 的 `AgentEvent`）；文件走 multipart 上传。
- **i18n 红线**：禁止硬编码 UI 字符串，一律用翻译 key；`pnpm i18n:check` 是 `just check` 的一部分。

---

## 13. 端到端数据流示例

把前面各层串起来，看四条典型链路。

### 13.1 发一条 echo（同步写 + 异步副作用）

```
POST /api/echo  (AuthRouterGroup, RequireAuth)
  → middleware: RequireAuth → 解析 viewer
  → echoHandler.Create        薄壳，组 DTO
  → echoService.Create        事务内落库 (echoRepository → GORM → SQLite)
       └─ eventbus.Notify(ctx, bus, event.EchoCreated{...})   ← best-effort，立即返回
  ← 200 给前端

  〔异步〕Busen 按 EchoCreated 类型路由 →
     • EmbeddingProcessor → embedding.IndexEcho（增量向量索引，失败退避重试）
     • AgentProcessor     → 清 agent 摘要缓存
     • webhook.Dispatcher → 转 WebhookObservation → worker pool → Sender.Deliver（HMAC 签名，重试）
```

### 13.2 站内问 Copilot（Agent 出站 + RAG + SSE 流）

```
POST /api/chat  (鉴权)
  → copilotHandler.Ask         设 SSE 头，透传 service 写入
  → copilotService.AskStream   组 messages，注入 search_echos / summarize_echos / stats_overview 三工具
  → agent.Run(ctx, RunRequest) → <-chan AgentEvent     （internal/agent 的 ReAct loop，§7）
       循环: provider.Stream → TextDelta 透传 / ToolCall 收集→并发执行→回喂 → Done
  → service 把 AgentEvent 映射为 SSE：searching | sources | delta | done | error（+15s keep-alive）
  ← 前端 chat.ts 用 fetch + ReadableStream 实时上屏
```

### 13.3 外部 LLM 经 MCP 读写（入站）

```
外部 AI 客户端 (Claude…) ── JSON-RPC ──► POST /mcp  (RequireAuth, access token)
  → mcp.Server.dispatch："tools/call" name=search_posts
       ① 校验 viewer 具备 echo:read scope（§8）
       ② Registry.LookupTool → Adapter.searchPosts → echoService 查询
  ← JSON-RPC result（ToolCallResult）
```

### 13.4 导出快照（异步 job）

```
POST /migration/export  (鉴权)
  → migratorHandler → migratorService（鉴权 + DTO）→ job.Manager.Submit(TypeExport, payload)
  → 〔goroutine〕job/runner.ExportRunner → migrator.ExportEngine.Export
       选 FS/S3 后端（storage.Manager）→ 打包 data/ 为 zip（pkg/virefs + plugin/zip）
       → eventbus.Notify(SystemSnapshot)   （触发 webhook 等）
  前端轮询 job 状态（统一 job 进度卡）→ 完成后 GET 下载产物
```

> 同步下载走 `GET /migration/export/download`；定时快照走 `task/scheduled` cron → 同一个 `ExportEngine`。**没有独立的 "backup" 概念，全是 snapshot export**。

---

## 14. 模块依赖关系总图

自上而下、依赖向下（上层依赖下层，下层不反向依赖上层）：

```
        cmd (Cobra) ─────────────────────────────────────────────┐
            │ bootstrap(config+logger) → cli → di.BuildApp        │
            ▼                                                       │
        internal/app  ── 编排 Component ──► server · job · task · EventRegistrar
            │                                                       │
   ┌────────┼───────────────────────────────────────────┐         │
   ▼        ▼                                             ▼         │
 handler  router ──► middleware ──► service ──► repository ──► model ──► GORM/SQLite
   │                    │              │   │   │
   │ (Bundle)           │ TokenRevoker │   │   └─ emit ─► event(纯词汇) ◄─ event/bus(Busen) ─► 订阅者(webhook/agent/embedding/snapshot)
   │                    ▼              │   ├─ copilot ──► internal/agent ──► 外部 LLM(出站，§7)
   │              auth service        │   └─ mcp ◄────── internal/mcp ◄──── 外部 LLM(入站，§8)
   │                                   ▼
   │              storage · cache · kvstore · transaction · captcha · visitor · setting · embedding · migrator
   │                   │            │          │                                            │
   ▼                   ▼            ▼          ▼                                            ▼
 (前端 web/dist 内嵌) pkg/virefs  Ristretto  keyvalue repo / pkg/gocap / pkg/busen / pkg/viewer
```

---

## 15. 红线速查（接手前务必记住）

1. **改 DI 图必跑 `just wire`**；提 PR 前 `just check`（后端 lint+openapi + 前端 lint+i18n）是强制的，`go build ./...` 与 `pnpm build` 必须过。
2. **三个共享单例**（`visitor.Tracker` / `storage.Manager` / `job.Manager`）只能在顶层注入一次，别让 Wire 复制出第二份。
3. **跨层导入用别名**（`xxxHandler/Service/Repository/Model/Util`）。
4. **加副作用优先发事件**，别在 handler 里直接调服务；事件路由靠 Go 类型，不靠 topic。
5. **依赖向内指向纯词汇**：`event` / `setting` 的 Spec / `agent` 的 Tool 等定义包不依赖基础设施或领域。
6. **Agent 包零领域依赖**：工具由 copilot service 注入；SDK 脏活封死在 Provider 内；错误分级，绝不静默吞错（§7.4）。
7. **MCP 鉴权统一在 dispatch**：每个 tool/resource 静态声明 scope，`Server` 在 call/read 前强制校验，别把鉴权散落进业务（§8）。
8. **存储改 S3 走 `storage.Manager.ApplyS3Setting`/`ReloadFromConfigAndDB`**，确保文件服务与设置页/迁移用的是同一份 Manager。
9. **路由两特例别动**：`POST /api/comments/integration`（scope+audience 鉴权，绕 captcha）、`/api/files` 静态服务。
10. **i18n 禁硬编码字符串**；日志用 `internal/util/log` 并带 `module` 字段。
11. **快照只有一种**：snapshot export（手动 job / 定时 cron / 同步下载），没有单独的 backup。

---

*维护提示：本文是高层全景，随架构演进需同步更新。具体子系统的权威实现以代码为准，深入细节见 `docs/dev/` 下对应专题文档。*
