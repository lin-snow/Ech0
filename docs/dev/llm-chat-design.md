# Ech0 LLM Chat 需求与设计文档

> 适用范围：后端新增 `internal/**/chat`（handler/service/repository/model）、`internal/**/embedding` 索引管线、`internal/event`（echo 索引订阅）、`internal/config` 与 DB 设置、`internal/database`（sqlite-vec 接入）、`internal/router`、以及前端 `web/src` 中的 Chat 入口与对话面。

> 最后更新：2026-05。

> 状态：**设计草案（v1 未实现）**。本文锁定需求/价值与总体设计；标注「待决」的项在实现前确认，标注「落地核对」的项在实现时对齐源码。

本文档系统说明 Ech0「LLM Chat」功能的**需求、价值论证、范围边界与技术设计**，目标是让任何开发者在阅读后能建立全局认知，并清楚「要做什么、为什么做、改哪些文件」。

## 目录

1. 背景与价值论证
2. 需求（功能 / 非功能）
3. 范围与边界（v1）
4. 用户场景
5. 总体设计
6. 详细设计
7. 鉴权与安全
8. 配置项
9. 成本评估
10. 风险与权衡
11. 分期里程碑
12. 待决问题
13. 落地核对清单

---

## 1. 背景与价值论证

### 1.1 问题

Ech0 是**微博客**（朋友圈式的日常分享 + 偶尔的想法/思考），不是笔记 app。当用户积累到几百上千条 Echo 后：

- 他几乎不会再一条条翻找；
- 关键词搜索（FTS）只能定位「某一条」，且对中文分词不友好；
- 「那年今日 / 随机回顾」只能被动地翻出**单条**，不会归纳。

结果是：**过去写的内容几乎丧失了任何作用。**

### 1.2 关键判断：这是「综合」需求，不是「检索」需求

笔记 app 的内容是**为将来取用而写的**，检索是其本分；而微博客的 Echo 是**当下的碎片**，写的时候并没有打算以后查它。因此「把过去的 Echo 搜出来回答问题」是在问数据当初不打算回答的问题——**把 Chat 做成站内搜索的高级版，对微博客是伪需求。**

但微博客攒下的是另一种东西：**一份无意中写成的、带时间戳的「自我记录」**——情绪、关注点、立场随时间的漂移。对这份记录真正有价值、且其它功能都做不到的操作是「**跨时间的归纳与综合**」：

| 替代方案 | 能否「跨时间归纳」 |
|---|---|
| 滚动翻阅 | ❌ 几百条翻不动 |
| 全文搜索 / FTS | ❌ 只能定位单条，不会综合；中文分词差 |
| 那年今日 / 随机回顾 | ❌ 被动、单条、不归纳 |
| 标签 / 分类 | ❌ 用户不会勤打标，且仍不综合 |
| **LLM Chat（本功能）** | ✅ 唯一能把碎片跨时间综合成洞察 |

**结论：「把一批碎片 Echo 跨时间归纳成洞察」是只有 LLM 能做的事。这是本功能的独占价值。**

### 1.3 定位

> **不是「和你的笔记对话」，而是一个能跨时间归纳你过往 Echo 的「自我镜像 / 回顾搭子」。**

它与「那年今日」是同一条产品主线、同一拨受众——「那年今日」是被动、随机、单条的怀旧一瞥；LLM Chat 是它的**主动版、可提问版、会归纳版**。

### 1.4 诚实的边界

- **独占 ≠ 高频**：它解决的是「偶尔想回顾一下」的低频时刻，不是每日刚需。定位为**惊喜型 / 后置型**功能，不指望它撑使用频次。
- **价值后置**：随 Echo 数量与时间跨度才长出来，对新用户 / 少量数据基本是死的。
- 因此 UI 上**不把它当核心功能塞用户脸上**，做成「攒够了自然浮现的回报」。

---

## 2. 需求

### 2.1 功能需求

- FR-1：登录用户（owner）可在一个对话界面，就自己的全部 Echo 提问，得到基于这些 Echo 的归纳/综合回答。
- FR-2：回答以 **SSE 流式**逐字返回。
- FR-3：界面提供 **2~3 个预设入口问题**（如「年度回顾」「最近在想什么 / 心情变化」「汇总我记过的 idea」），引导用户问出具体问题，而非面对空对话框。
- FR-4：系统对 Echo 内容维护向量索引；Echo 新增/编辑/删除时**增量更新**索引。
- FR-5：提供**一次性回填**入口，为存量 Echo 补建索引。
- FR-6：embedding 服务**可独立配置**（独立于 Chat/生成所用 LLM 的 provider/base_url/key/model）。

### 2.2 非功能需求

- NFR-1（轻量）：不破坏「单 Go 二进制」分发；不引入需独立分发的运行时文件；新增外部依赖仅 embedding API。详见 [[project_lightweight_principle]] 原则。
- NFR-2（复用）：生成侧复用现有 `agent` 的多 provider 抽象（`agent.Generate`）与 `AgentSetting` 配置；流式需在该抽象上新增入口（agent 当前非流式，见 §6.6）。SSE 传输复用 dashboard 既有写法。
- NFR-3（隔离）：检索层做干净抽象，向量后端可替换（即便首版用 sqlite-vec）。
- NFR-4（鲁棒）：索引更新失败可重试，不阻塞 Echo 的正常增删改。
- NFR-5（中文友好）：检索基于多语言 embedding，天然处理中文，避免 FTS 分词问题。

---

## 3. 范围与边界（v1）

### 3.1 v1 包含

- 仅**登录用户（owner）**可用的 RAG Chat。
- 锚定场景 **A（回顾/反思）**，并把「idea 汇总」作为 A 的子场景一并实现。
- sqlite-vec 向量检索 + 增量索引 + 存量回填。
- SSE 流式生成，复用现有 agent LLM 能力。
- 独立的 embedding 配置。
- 预设入口问题。

### 3.2 v1 不包含（明确搁置）

- **公开 / 匿名可用**：v1 仅登录。公开会把它从「个人工具」变成「对外服务」（匿名滥用、烧 token、prompt 注入），留待核心体验跑通后再评估。设计上预留 `chat.public_enabled` 开关位，但 v1 不实现公开路径。
- **场景 B（创作辅助 / 编辑器旁的写作搭子）**：与「朋友圈式随手发布」的本质相冲，有效面窄，且需新开交互面。待真有思考型用户提出再做。
- **多轮会话记忆 / 会话持久化**：v1 默认**单轮独立检索问答**（是否带历史见 §12 待决）。

---

## 4. 用户场景

### 4.1 场景 A —— 回顾/反思（v1 主场景）

> 「我这半年都在焦虑什么？」「我对 AI 的看法这两年怎么变的？」「帮我写一份基于我 Echo 的年度总结。」

这些问题**滚动翻不出、关键词搜不出**，只能由 LLM 读取一批相关 Echo 后归纳。微博客的碎片化、时间戳、情绪浓度恰是此场景的理想原料。

### 4.2 子场景 —— idea 汇总（A 的一种）

> 「把我零散记过的关于 X 的想法汇总成一段。」

把高密度但被淹没的「想法」从流水账里救出来。本质是综合（A），非从零创作（B）。

### 4.3 交互形态

- 入口：后台 / 已登录态下的一个独立 Chat 面（后置浮现，不抢主导航焦点）。
- 打开后呈现预设问题卡片；点击即作为首条提问；也可自由输入。
- 回答流式逐字出现；可附带「引用了哪些 Echo」（点击可跳转原文，便于信任与溯源）。

---

## 5. 总体设计

```
                         ┌─────────────────────────────┐
  Echo 增删改  ──事件──▶  │  Embedding 索引订阅者         │
   (事件总线 Busen)       │  - 取文本 → 调 embedding API │
                         │  - 写入 vec0 向量表          │
                         │  - 失败进重试/死信           │
                         └──────────────┬──────────────┘
                                        ▼
                              ┌───────────────────┐
   回填命令 ───────────────▶  │  echo_embedding    │  (sqlite-vec vec0 虚表)
                              └─────────┬─────────┘
                                        │ top-k 相似检索
   owner 提问 ──HTTP──▶ Chat Handler ──▶ Chat Service
                                        │  1. embed(query)
                                        │  2. 检索 top-k Echo
                                        │  3. 拼 prompt
                                        │  4. 调 LLM（复用 agent 流式）
                                        ▼
                            SSE 流式逐字返回前端
```

分层遵循项目约定：`handler → service → repository → model`，跨层导入用 `xxxHandler/Service/Repository/Model/Util` 别名。

---

## 6. 详细设计

### 6.1 数据模型

**已核对的 Echo 主表字段**（`internal/model/echo/echo.go`）：`ID`（string，UUID v7）、`Content`（text，正文）、`Username`、`UserID`、`Private`（bool，可见性，非软删除）、`CreatedAt`（int64，Unix 时间戳，已建索引）、`Tags`、`EchoFiles`、`Extension`。索引文本取自 `Content`（可附带 `Tags`）。

**重要约束（已核对 sqlite-vec v0.1.6）**：`vec0` 是**虚拟表**，GORM `AutoMigrate` 无法创建/管理它，且**向量维度在建表时即固定**（`embedding FLOAT[N]`）。维度取决于运行时配置的 embedding 模型，安装时未知。因此采用**两表 + 懒建表**：

1. **元数据表 `echo_embeddings`（普通 GORM 表，进 `MigrateDB()` models 列表）**：
   - `echo_id`（`char(36)`，主键，关联 Echo）
   - `content_hash`（文本哈希，判断编辑后是否需重 embed；内容未变则跳过，省调用）
   - `model`（生成所用 embedding 模型名）/ `dim`（维度）
   - `created_at` / `updated_at`
2. **向量虚表 `vec_echo`（`vec0` 虚拟表，raw SQL 懒创建，不进 AutoMigrate）**：
   - `echo_id TEXT PRIMARY KEY, embedding FLOAT[<dim>]`
   - 由索引/检索代码在**首次需要时**按当前配置维度执行 `CREATE VIRTUAL TABLE IF NOT EXISTS` 创建（不是启动期固定 migrator，因为维度依赖配置）。

> **换模型 / 换维度的处理（v1 取简）**：检测到配置的 `model`/`dim` 与现存不一致时，**DROP 重建 `vec_echo` + 清空 `echo_embeddings` + 触发一次全量回填**。这是最轻量的策略，避免维护多维度并存。回填命令是统一入口。
>
> 幂等：`content_hash` 保证回填可安全重跑、增量更新只在内容变化时调用 embedding。

### 6.2 sqlite-vec 接入

- **方式（已验证 ✅）**：用 `github.com/asg017/sqlite-vec-go-bindings/cgo`（v0.1.6），在 `gorm.Open` 之前调用一次 `sqlite_vec.Auto()` 即可——它通过 `sqlite3_auto_extension` 把 sqlite-vec 注册为**进程级自动扩展**，之后所有新建的 mattn/go-sqlite3 连接（含热切换）都自动具备 `vec0` 能力。**无需自定义 driver / ConnectHook**，比最初设想简单得多。C 扩展随 CGO 静态编译进二进制，不分发独立 `.so`，单二进制不破。
- **接入位置**：`internal/database/database.go` 的 `InitDatabase()`，在 sqlite 分支 `gorm.Open` 前加 `sqlite_vec.Auto()`（已实现）。
- **冒烟结论**：本机（darwin/arm64）spike 验证 `vec_version()`=v0.1.6、`vec0` 建表/插入/KNN 全通过，`go build ./...` 全绿。macOS 下有 `sqlite3_auto_extension is deprecated` 编译告警（Apple 系统库提示），但因 mattn 用自带 sqlite3 而非系统库，实测功能正常；生产目标 Linux 完全支持。
- **⚠️ 仍需 CI 验证**：zig/musl 各 arch（含 riscv64/loongarch64）交叉编译冒烟——本机未装 zig，必须在 CI 跑一次确认 sqlite-vec 的 C 在所有目标平台编过。
- **跨平台**：复用现有 `release_zigcc.yml` 的 `zig cc -target ...` 工具链；多编一个 `vec0.c` 不增加新工具链负担。
- **预置风险（必做）**：sqlite-vec 是干净 C99，但目标平台均为 `*-linux-musl`（含 riscv64 / loongarch64）。**上车前先跑一次 CI 冒烟编译**确认全平台绿。低风险但不可凭「应该能」拍板。
- **为什么是 sqlite-vec 而非 Go 暴力余弦**：在 zigcc 抹平跨平台编译后，sqlite-vec 的接入成本已足够低；选它是为了**干净的 SQL 检索语义（`MATCH ... k=N`）、量化省内存、以及数据涨到十万级后无需重写检索层**。注意它当前仍是 O(n) 线性扫描（无 ANN/HNSW），在千级数据下相对手写暴力扫无可感性能差异——采用它是**工程抽象与未来可扩展**的取舍，而非当前性能需求。

### 6.3 Embedding 配置（独立）

- embedding 的 provider / base_url / api_key / model / dim **独立于** Chat 生成所用 LLM。
- 存储沿用现有 **DB 设置体系**：agent 的 LLM 配置以 `AgentSetting` 存于 KeyValue 设置表（key `agent_setting`，读取入口 `settingService.GetAgentInfo`），S3 配置同理。embedding 配置新增一个并列的设置项（如 `embedding_setting`），由 admin 面板维护，不进 env。
- 提供「测试连接」能力（可选）以便用户验证配置有效。

### 6.4 索引管线（增量 + 回填）

- **增量**：Echo 新增/编辑/删除经**事件总线（Busen）**发布事件 → embedding 订阅者消费：
  - 新增/编辑：计算 `content_hash`，变化则调 embedding 写入/更新 `vec0`；
  - 删除：删除对应向量行。
  - 遵循项目约定——**副作用走事件订阅，而非在 handler 内联调用**。
- **失败处理**：现有事件体系有**死信捕获**（`queue_dead_letters` 表 + `DeadLetterResolver`），但**没有自动重试循环**（目前是人工/手动重放）。因此索引订阅者需自带轻量重试（如有限次退避），失败兜底进死信；且索引失败**不得阻塞** Echo 本身的增删改。回填命令是最终兜底。
- **回填**：提供一次性回填（CLI 命令或 admin 触发），对缺失/模型不匹配的 Echo 补建索引；借 `content_hash` 幂等，可安全重跑。

### 6.5 检索

1. 对用户 query 调 embedding 得查询向量；
2. 在 `vec0` 取 top-k 最相似 Echo（k 默认值见 §12 待决，预期保守值 5~8，可配）；
3. 可附加过滤（如时间范围——用于「这半年」类问题；落地核对 Echo 时间字段）。

检索结果（含 Echo 原文与元信息）交给生成层，并保留「命中清单」用于前端引用展示。

### 6.6 生成与 SSE 流式

**现状（已核对，注意与最初设想不同）：**
- LLM 调用统一收口在 `internal/agent/agent.go` 的 `agent.Generate(ctx, setting, in, usePrompt, temperature...)`，按 `AgentSetting.Protocol` 路由到 **go-openai / anthropic-sdk-go / genai** 三家 SDK。
- **该路径目前是非流式的**（一次性返回完整字符串）。`internal/service/agent` 的「近期总结」也是非流式 + 结果缓存。
- 项目里**已有 SSE 实现，但在 dashboard 系统日志**（`internal/service/dashboard/dashboard.go`：`http.Flusher` + `text/event-stream` + 15s keep-alive），**不在 agent**。

**因此本功能需要新增（不是纯复用）：**
1. 在 `internal/agent` 增加一个**流式生成入口**（如 `GenerateStream(...) <-chan token`），对三家 SDK 各自的 streaming API 做封装；这是 net-new 工作量,是本功能的核心改动之一。
2. Chat Service 组装 system prompt + 检索上下文 + 用户问题，调流式生成，拿到 token 流。
3. Chat Handler 复用 dashboard 那套 SSE 写法（flusher + `text/event-stream` + keep-alive）把 token 流转发前端。
4. 配置可复用现有 `AgentSetting`（Protocol/Model/Key 等），生成侧无需另起配置；仅 embedding 需独立配置（§6.3）。

> **进阶（Agent / Tool Calling）**：把「固定先检索一次再生成」升级为「模型在一轮内自主、多次决定是否检索」的 Agent 形态，以及 `internal/agent` 的 Provider 抽象 + Loop 重构，见独立设计文档 [`docs/dev/agent-toolcall-design.md`](./agent-toolcall-design.md)。

### 6.7 Prompt 设计与预设问题

- **System prompt** 要点：
  - 角色设定为「基于用户自己过往 Echo 的回顾/综合助手」；
  - **只依据提供的 Echo 上下文作答，缺乏依据时如实说明「你的 Echo 里没有相关记录」，不得编造**；
  - 引用时尽量标注来源（便于前端关联原文）。
- **预设入口问题**（FR-3）作为引导，避免「流水账 + 含糊问题 = 废话输出」：
  1. 回顾型：「总结我最近一段时间在关注/思考什么」
  2. 情绪型：「我这段时间的心情变化是怎样的」
  3. idea 型：「汇总我记过的想法 / 关于某主题的零散思考」
- 预设问题文案须走 i18n key，**不得硬编码 UI 字符串**（见 `docs/dev/i18n-contract.md`）。

### 6.8 API 设计

- 新增 Chat 路由组，挂**登录鉴权中间件**（owner-only，复用 `internal/middleware` 现有登录态校验）。
- 端点（命名落地时统一）：
  - `POST /api/chat`（或 `/api/chat/stream`）：提交问题，返回 **SSE** 流。
  - `GET /api/chat/suggestions`：返回预设入口问题（可由后端给定或前端静态，倾向后端以便 i18n/可配）。
  - 回填：admin/CLI 触发的索引重建入口。
- 路由变更后重新生成 **OpenAPI 规格**（`just openapi`，产物 internal/openapi/openapi.yaml）。

### 6.9 前端形态

- 一个独立、克制的 Chat 面板（后置浮现，不抢主导航）。
- 预设问题卡片 → 点击即提问；支持自由输入。
- 流式渲染；回答可展示「引用的 Echo」并跳转原文。
- 所有文案走 i18n。

---

## 7. 鉴权与安全

- **仅登录（owner）**：所有 Chat / 索引 / 配置端点挂登录鉴权，沿用现有 JWT 登录态中间件（见 `docs/dev/auth-design.md`）。
- **公开模式（搁置）**：预留 `chat.public_enabled` 开关位，但 v1 不实现公开路径，避免匿名滥用 / 烧 token / prompt 注入。
- **成本可控**：embedding 成本极低（见 §9）；生成成本是主项，由 owner 自己的 LLM 配置承担，仅本人可触发，天然受控。
- **注入面**：v1 只有 owner 自己的内容进入 prompt，注入风险低；将来开放公开时需重审「访客诱导 AI 分身」的注入问题。

---

## 8. 配置项

| 配置 | 位置 | 说明 |
|---|---|---|
| embedding provider/base_url/api_key/model/dim | DB 设置 | 独立于生成 LLM |
| 生成 LLM provider 等 | DB 设置（`AgentSetting`） | 复用现有 agent 配置（Protocol/Model/Key） |
| `chat.enabled` | DB 设置 | 功能总开关 |
| `chat.public_enabled` | DB 设置（预留，v1 不启用） | 公开可用开关 |
| top-k、上下文预算 | DB 设置 / 默认值 | 检索条数与 prompt 预算 |

> 约定：功能型 LLM 配置走 **DB 设置**（与 agent/S3 一致），而非 env。`internal/config`（caarlos0/env）仅承载进程级基础配置。

---

## 9. 成本评估

以 owner 拥有约 1000 条 Echo、单条平均约 200 token 估算：

| 项目 | 量 | 成本 |
|---|---|---|
| 全量建索引（一次性） | ~20 万 token | `text-embedding-3-small`（$0.02/1M）≈ **$0.004** |
| 每新增一条 Echo | ~200 token | ≈ 0 |
| 每次提问的 query embedding | 几十 token | ≈ 0 |
| 向量存储 | 1000 × 1536 维 × 4 B | **~6 MB** |

**embedding 这层几乎免费**；真正的成本是生成（把上下文喂给 LLM 出答），但这笔钱用不用 RAG 都要花，且仅 owner 本人触发。

---

## 10. 风险与权衡

- **价值后置**：数据浅时近乎无用 → 做成「攒够了浮现的回报」，不强推。
- **含糊问题产生废话** → 预设入口问题引导 + system prompt 要求「无依据则如实说明」。
- **sqlite-vec 跨平台编译** → 上车前 CI 冒烟（§6.2）。
- **sqlite-vec 当前无 ANN** → 接受千级数据下与暴力扫无可感差异；选它是为抽象与未来扩展，非当前性能。
- **索引与 Echo 写入的一致性** → 走事件 + 重试，失败不阻塞主流程；回填兜底。
- **轻量原则** → 仅新增 embedding API 一个外部依赖；不破坏单二进制。

---

## 11. 分期里程碑

- **M0**：sqlite-vec 接入 —— 自定义 driver + ConnectHook 注册（§6.2）+ CGO 静态编译 + zigcc/musl 全平台 CI 冒烟通过。**风险最高,先做。**
- **M1**：数据模型（`echo_embedding`，登记进 `MigrateDB()` 的 models 列表）+ 增量索引（订阅 `echo.created/updated/deleted` 事件）+ 回填命令。
- **M2**：embedding 独立配置（KeyValue 设置项 + admin UI）。
- **M3**：检索 + 生成。**含 net-new 的流式生成入口**（`internal/agent` 给三家 SDK 各加 streaming 封装，§6.6）+ Chat Handler 用 dashboard 的 SSE 写法转发。
- **M4**：前端 Chat 面 + 预设问题 + 引用跳转 + i18n。
- **M5**：`just check`（lint / i18n / openapi / wire-check）+ `just wire` + 文档完善。

---

## 12. 待决问题

1. **embedding 提供方/模型**：跟 agent 用同一 provider 还是独立 endpoint？（已定方向：**独立配置**；具体默认模型待定）
2. **是否带多轮会话记忆**：~~v1 单轮独立检索，还是带历史的多轮对话（需存会话）？~~——**已决策：带历史多轮**。策略：展示 transcript（`ChatMessage`，含 `Sources`）与喂模型的 context 分离，每轮从持久化会话投影出模型历史（`historyForModel`）；旧轮只留 user/assistant 文本、剥掉过时的检索结果（模型需旧细节会经 `search_echos` 重检索），仅「最近一轮」的 `Sources` 折进文本兜住追问细节；按 token 预算（非条数）滑动窗口截断；摘要压缩留二期。
3. **top-k / 上下文预算默认值**：top-k 预期 5~8（实现取 `defaultTopK=6`）；历史上下文预算 `maxHistoryTokens=4000`（保守固定值，与模型窗口解耦，按 rune 数粗估，不引 tokenizer）。
4. **引用展示粒度**：是否在回答中逐句标注来源，还是仅在末尾列出命中 Echo。

---

## 13. 落地核对清单（实现时与源码对齐）

以下为已核对的关键事实与待办锚点：

- [x] **Echo 模型字段**：`internal/model/echo/echo.go` —— `ID`(string)/`Content`/`UserID`/`Username`/`Private`/`CreatedAt`(int64)/`Tags`。无软删除，`Private` 为可见性。
- [x] **Echo 事件契约**：`internal/event/contracts/contracts.go` 定义 `TopicEchoCreated/Updated/Deleted`（`echo.created/updated/deleted`）与 `EchoCreatedEvent` 等；发布在 `internal/service/echo/echo.go`（`publisher.EchoCreated(...)`）；订阅参考 `internal/event/subscriber/agent.go`（`registry.TopicSubscription(...)`）。死信表 `queue_dead_letters` + `DeadLetterResolver`，**无自动重试**。
- [x] **agent LLM 调用单元**：`internal/agent/agent.go` 的 `Generate(ctx, setting, in, usePrompt, temperature...)`，按 `AgentSetting.Protocol` 走 go-openai/anthropic/genai；**当前非流式**，需新增流式入口（§6.6）。配置经 `settingService.GetAgentInfo(&AgentSetting)` 读取。
- [ ] **DB 设置体系**：参考 `internal/service/setting/agent_setting_service.go` 与 KeyValue 设置（key 见 `internal/model/common/common.go`），新增 `embedding_setting`。
- [ ] **sqlite 连接**：`internal/database/database.go` 用 `gorm.io/driver/sqlite` 直接 `gorm.Open`，**无扩展钩子**；需自定义 driver + `ConnectHook` 注册 sqlite-vec（§6.2）。
- [ ] **迁移注册**：在 `database.go` 的 `MigrateDB()` models 列表登记 `echo_embedding`；如需后处理，加 `internal/database/migration/` 迁移器。
- [ ] **Wire DI**：`internal/di/wire.go` —— 新 domain 的 repo/service/handler 加入 `HandlerSet`（仿 `service.AgentSet/handler.AgentSet`）；索引订阅者加入 `EventSet` 并在 `ProvideSubscriptionProviders` 注册；改后 `just wire`。
- [ ] **鉴权路由**：chat 端点挂 `AuthRouterGroup`（`middleware.RequireAuth`），按需加 `middleware.RequireScopes(...)`；owner-only。注意现有 `/agent/recent` 用的是 `PublicRouterGroup`，chat 不可照搬。
- [x] **SSE 写法**：`internal/service/dashboard/dashboard.go` 已有 `http.Flusher` + `text/event-stream` + 15s keep-alive，可作模板。
- [ ] **i18n**：预设问题、按钮、错误提示全部走 key，跑 `pnpm i18n:check`。

---

> 本文档聚焦需求/价值与设计；具体实现 PR 前请完成 `just check` 与本文 §13 的全部核对项。
