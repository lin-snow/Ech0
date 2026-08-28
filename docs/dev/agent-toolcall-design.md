# Ech0 Agent 包重构与 Tool Calling 架构设计

> 关联文档：`docs/dev/llm-chat-design.md`（§6.5 检索、§6.6 生成与 SSE 流式）。
> 本文聚焦 `internal/agent` 包的重构，使 Chat 从「单次 RAG 生成」升级为「模型在一轮对话内自主、多次决定是否检索」的 Agent 形态。

## 1. 目标与非目标

### 1.1 目标

- **G1**：让模型在一次问答内**自主决定要不要检索、检索几次、用什么 query**（agentic loop），而不是固定先检索一次再生成。
- **G2**：把 `internal/agent` 重构成**清晰的 Provider 抽象 + 统一的 Agent Loop**，消除当前散落在 `Generate`/`GenerateStream` 里的 `switch protocol`。
- **G3**：tool calling 覆盖两家主流协议 **OpenAI 兼容 / Anthropic**（OpenAI 兼容已覆盖 DeepSeek/Qwen/Moonshot/Groq/OpenRouter 等绝大多数托管服务商）。Gemini 整协议下线，见 §15.1。
- **G4**：把"各家 SDK 流式分片的脏活"封死在 Provider 内部，上层（Loop、Chat Service）只见**干净的语义事件**。
- **G5**：复用现有轻量件——检索复用 `embedding.Search`，SSE 复用 `flusher + text/event-stream + 15s keep-alive`，配置复用 `AgentSetting`，**不新增重依赖**。

### 1.2 非目标（本期明确搁置）

- **本地小模型（Ollama / llama.cpp / LM Studio）的 tool calling 兼容**：这些端点 function calling 支持参差不齐。本期只保证主流服务商；不支持工具调用的模型走「能力错误」明确报错，**不做 ReAct 文本兜底**（见 §7.3）。
- **跨轮对话历史的持久化**：消息抽象会预留历史位（§11），但是否启用多轮是独立的 scope 开关，不在本重构强绑。
- **并行工具的并发执行**：一轮内多个 tool_call 先**顺序执行**，简单可控；并发留作后续优化。

## 2. 设计原则

1. **脏活下沉，语义上浮**：SDK 差异（尤其流式 tool_call 的分片拼接）只允许存在于各 Provider 实现内部；Loop 与 Chat Service 只消费统一的 `Event` / `AgentEvent`。
2. **领域零侵入**：`agent` 包不得 import `embedding` / `echo` 等领域包。工具以「定义 + 执行闭包」的形式**由 Chat Service 注入**，agent 包只认 `Tool` 抽象。
3. **错误分级**：工具执行错误 → 作为 tool 结果**喂回模型**让它自愈；传输/协议错误 → 中止并 SSE `error`。绝不静默吞错（见 §8）。
4. **加法而非替换**：SSE 事件契约**只新增**（`searching`），保留现有 `sources / delta / done / error`，前端旧逻辑不破坏。
5. **护栏先行**：轮数上限、查询去重、token 预算、超时——在 Loop 层强制，防止模型死循环狂调工具烧 token。

## 3. 架构总览

```
┌──────────────────────────── Frontend (web/) ────────────────────────────┐
│  TheChatBox.vue ── chat.ts (原生 fetch + ReadableStream 读 SSE)          │
│  SSE 事件: searching | sources | delta | done | error                    │
└───────────────────────────────▲──────────────────────────────────────────┘
                                 │ SSE (text/event-stream)
┌──────────────────── internal/handler/copilot ───────────────────────────┐
│  CopilotHandler.Ask     薄壳：设头 + 把 service 的写入透传                │
└───────────────────────────────▲──────────────────────────────────────────┘
┌──────────────────── internal/service/copilot ───────────────────────────┐
│  CopilotService.AskStream   (chat.go)   ── 另有 summary.go：GenRecent     │
│   1. 组装 messages（system prompt + [history] + question）               │
│   2. 构造领域工具 search_echos：Def(JSON Schema) + Execute→embedding.Search│
│   3. 调 agent.Run(...)，消费 <-chan AgentEvent                           │
│   4. AgentEvent → SSE（searching/sources/delta/done/error）+ 15s keep-alive│
└──────────────▲───────────────────────────────┬───────────────────────────┘
               │ <-chan AgentEvent              │ []agent.Tool (注入)
┌──────────────┴──────────────── internal/agent（重构核心）────────────────┐
│                                                                          │
│  agent.Run(ctx, RunRequest) (<-chan AgentEvent, error)                   │
│  ┌────────────────── Loop Controller（ReAct 结构）──────────────────┐    │
│  │  for round < maxRounds:                                          │    │
│  │     provider.Stream(req)  ──►  消费 Event 流                     │    │
│  │       · TextDelta  → 透传为 AgentEvent.Delta（实时上屏）         │    │
│  │       · ToolCall   → 收集；本轮结束后执行                        │    │
│  │     若本轮有 ToolCall:                                           │    │
│  │       去重 → 顺序 Execute → 追加 assistant(tool_calls)+tool 结果 │    │
│  │       发 AgentEvent.Searching/Sources → continue                │    │
│  │     否则: 发 AgentEvent.Done → break                            │    │
│  │  护栏: maxRounds / 查询去重 / token 预算 / ctx 超时              │    │
│  └────────────────────────────▲────────────────┬──────────────────┘    │
│                               │ Event           │ Request(含 Tools 的 Def)│
│  ┌────────────────────────────┴── Provider 接口 ─────────────────────┐   │
│  │  Stream(ctx, Request) (<-chan Event, error)                       │   │
│  │  职责: 内部 Message/ToolDef ↔ 自家 SDK 形状双向映射；            │   │
│  │        吞掉流式分片，拼好后 emit 语义 Event                       │   │
│  └─────────┬───────────────────┬────────────────────┬───────────────┘   │
│      openaiProvider              anthropicProvider                       │
│      (go-openai)                 (anthropic-sdk-go)                       │
│            │                            │                                 │
│      providerFor(setting AgentSetting) (Provider, error)  ← 工厂          │
└────────────┼────────────────────────────┼────────────────────────────────┘
        OpenAI/DeepSeek/                Claude
        Qwen/Moonshot/Groq...
```

**关键分界线**：
- Provider 接口 = LLM 协议适配层（脏活止于此）。
- Loop Controller = 协议无关的 ReAct 编排（护栏在此）。
- Chat Service = 领域工具 + SSE 传输（领域知识止于此）。
- agent 包对领域**零依赖**：工具靠注入。

## 4. 分层职责

| 层 | 文件（建议） | 职责 | 不该做 |
|---|---|---|---|
| Provider 适配 | `internal/agent/provider_{openai,anthropic}.go` | 内部类型 ↔ SDK 类型映射；流式分片拼接；emit `Event` | 不懂 Loop、不懂 SSE、不懂领域 |
| 类型抽象 | `internal/agent/types.go` | `Message/ToolCall/ToolDef/Tool/Event/AgentEvent/Request/RunRequest` | 无逻辑 |
| Loop 编排 | `internal/agent/run.go` | ReAct 循环 + 护栏 + 工具分发 | 不碰 SDK、不写 HTTP |
| 工厂 | `internal/agent/provider.go` | `providerFor(AgentSetting)` 按 Protocol 选实现 | — |
| 兼容壳 | `internal/agent/agent.go`（瘦身） | `Generate`/`GenerateStream` 保留为 Run 的薄封装 | — |
| Copilot Service | `internal/service/copilot/{chat,summary,prompt}.go` | chat：组装消息、定义 `search_echos`、AgentEvent→SSE；summary：缓存非流式 | 不碰 SDK |

## 5. 核心类型设计

```go
package agent

// ---- 消息（tool-aware，替换现有纯文本 Message）----
type Role string
const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleTool      Role = "tool"   // 新增：工具返回结果
)

type Message struct {
    Role       Role
    Content    string       // 文本内容
    ToolCalls  []ToolCall   // 仅 assistant：本条发起的工具调用
    ToolCallID string       // 仅 RoleTool：对应的 ToolCall.ID
}

type ToolCall struct {
    ID   string          // 各家 SDK 的调用 id
    Name string          // 工具名
    Args json.RawMessage // 入参（JSON）
}

// ---- 工具：定义 + 执行闭包（执行体由 Chat Service 注入）----
type ToolDef struct {
    Name        string
    Description string
    Parameters  json.RawMessage // JSON Schema
}
type Tool struct {
    Def     ToolDef
    Execute func(ctx context.Context, args json.RawMessage) (ToolOutput, error)
}
type ToolOutput struct {
    Content string // 回喂模型的文本
    Meta    any    // 旁路数据（如命中的 SearchResult，供 SSE sources 用）
}

// ---- Provider 层的请求/事件 ----
type Request struct {
    Messages    []Message
    Tools       []ToolDef
    Temperature *float32
    MaxTokens   int
}
type EventKind int
const (
    EventTextDelta EventKind = iota // 文本增量
    EventToolCall                   // 一个拼装完整的工具调用
    EventDone                       // 本次 provider 调用结束
    EventError
)
type Event struct {
    Kind      EventKind
    Text      string    // EventTextDelta
    ToolCall  ToolCall  // EventToolCall
    Err       error
}

// ---- Run 层（对 Chat Service 暴露）----
type RunRequest struct {
    Setting   model.AgentSetting
    Messages  []Message
    Tools     []Tool
    MaxRounds int     // 0 → 默认 3
    Temp      float32
}
type AgentEventKind int
const (
    AgentDelta     AgentEventKind = iota // 文本上屏
    AgentSearching                       // 模型决定检索（含 name+query）
    AgentToolResult                      // 工具执行完（含 Meta，供 sources）
    AgentDone
    AgentError
)
type AgentEvent struct {
    Kind     AgentEventKind
    Text     string
    ToolName string
    ToolArgs json.RawMessage
    Meta     any   // AgentToolResult: 即 ToolOutput.Meta
    Err      error
}

func Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error)
```

**设计要点**：
- `Event`（Provider→Loop）与 `AgentEvent`（Loop→Service）**刻意分层**：Provider 不知道"searching""sources"这种业务语义，只吐 TextDelta/ToolCall；语义翻译在 Loop。
- `ToolOutput.Meta` 是把领域数据（命中的 `[]embeddingModel.SearchResult`）旁路带出来给 SSE `sources` 用的通道，**不污染**回喂模型的 `Content`。

## 6. Provider 接口与两家适配

```go
type Provider interface {
    Stream(ctx context.Context, req Request) (<-chan Event, error)
}
func providerFor(s model.AgentSetting) (Provider, error) // 按 s.Protocol 选实现
```

各家把 SDK 差异封死在内部，对外只 emit 统一 `Event`：

| 协议 | SDK | tool 表示 | 流式 tool_call 分片 |
|---|---|---|---|
| OpenAI 兼容 | go-openai | assistant `tool_calls[]` + `role:"tool"` 消息带 `tool_call_id` | `arguments` JSON 按 `index` 跨 chunk 分片，Provider 内累积 |
| Anthropic | anthropic-sdk-go | `tool_use` / `tool_result` content block | `content_block_delta` 的 `input_json_delta` 累积 |

**统一约定**：Provider 内部维护「按 index/id 累积分片 → JSON 完整后 emit 一个 `EventToolCall`」。上层永远拿到的是**完整**的 ToolCall，不处理半个 JSON。这是 G4 的落点，也是整个重构「优雅」与否的关键。

## 7. Agent Loop（ReAct 结构）与护栏

### 7.1 ReAct 与原生 function calling 的关系（澄清）

我们采用 **ReAct 的循环结构**（reason → act → observe → 重复），但**用原生 function calling 作为 act/observe 的传输层**，而不是 ReAct 原始论文那种「让模型输出 `Thought:/Action:/Observation:` 纯文本再正则解析」。

- 原生 function calling：结构化、可靠、SDK 直接给 `tool_calls`。**本期采用。**
- 文本 ReAct：任何模型可用但靠 prompt 约束格式、解析脆。**本期不用**（因已放弃本地小模型，见 §1.2）。

即：循环的**骨架是 ReAct**，调用的**机制是 function calling**。

### 7.2 循环伪码

```
messages := req.Messages
seen := set()                       // 查询去重
for round := 0; round < maxRounds; round++ {
    evCh := provider.Stream(ctx, Request{messages, toolDefs, temp})
    var pendingCalls []ToolCall
    for ev := range evCh {
        switch ev.Kind {
        case EventTextDelta: emit AgentDelta(ev.Text)        // 文本实时上屏（跨轮连续）
        case EventToolCall:  pendingCalls = append(...)
        case EventError:     emit AgentError; return
        }
    }
    if len(pendingCalls) == 0 {                               // 本轮无工具 → 收尾
        emit AgentDone; return
    }
    // 追加 assistant 的 tool_calls 消息（供下一轮上下文）
    messages = append(messages, assistantMsg(pendingCalls))
    for _, tc := range pendingCalls {
        if seen.has(normalize(tc)) {                          // 去重：同 query 不重复检索
            messages = append(messages, toolMsg(tc.ID, "（已检索过，结果见上）"))
            continue
        }
        seen.add(normalize(tc))
        emit AgentSearching(tc.Name, tc.Args)                 // → SSE searching
        out, err := tool.Execute(ctx, tc.Args)
        if err != nil {                                       // 工具错误→喂回模型自愈
            messages = append(messages, toolMsg(tc.ID, "工具执行失败："+err.Error()))
            continue
        }
        emit AgentToolResult(out.Meta)                        // → SSE sources
        messages = append(messages, toolMsg(tc.ID, out.Content))
    }
    // 继续下一轮：模型带着工具结果再 reason
}
emit AgentDone   // 到达 maxRounds 仍未收尾，强制结束（已产出的文本即答案）
```

### 7.3 护栏

| 护栏 | 默认 | 作用 |
|---|---|---|
| `maxRounds` | 3 | 工具轮数上限，防死循环 |
| 查询去重 | 同 turn 内同 `(name,args)` 不重复执行 | 防模型反复搜同一词烧 token |
| token 预算 | 软上限（可由 `ECH0_CHAT_*` 配） | 累积上下文超限时丢最旧 tool 结果 |
| `ctx` 超时 | 复用请求 ctx | 客户端断开即停 |
| 能力错误 | — | 模型不支持 tools 时 SDK 报错 → 透传为 `AgentError`，**明确文案**，不静默 |

## 8. 错误处理哲学（防静默失败）

- **工具执行错误**（检索失败、参数非法）：**不中止**，包装成 tool 结果 `"工具执行失败：..."` 喂回模型，让它换 query 或如实告知用户。
- **Provider/传输错误**（鉴权失败、网络、模型不支持 tools）：**中止**，emit `AgentError` → SSE `error`，前端明确提示。
- **到达 maxRounds**：正常收尾（`AgentDone`），把已生成文本作为答案；可在末尾附一句"（检索轮数已达上限）"。
- 所有错误都有可观测落点（zap，`module=agent`，见 `docs/dev/logging.md`），**不允许 `_ = err` 式吞错**。

## 9. SSE 事件契约（扩展，向后兼容）

| 事件 | 时机 | payload | 前端 |
|---|---|---|---|
| `searching` | **新增**。模型决定检索时 | `{name, query}` | 显示「🔍 正在检索：{query}」状态条 |
| `sources` | 每次工具执行完（可多次） | `[]SearchResult` | 合并去重后渲染引用链接 |
| `delta` | 文本增量 | `{text}` | 追加上屏（跨轮连续） |
| `done` | 收尾 | `{done:true}` | 结束 loading，切完整 markdown |
| `error` | 中止 | `{message}` | 错误提示 |

注：现有前端只认 `sources/delta/error/done`，新增 `searching` 不破坏旧逻辑（未处理则忽略），符合"加法不替换"原则。`sources` 从「一次性」变为「可多次增量到达」，前端 `onSources` 需改为**累积合并去重**而非覆盖。

## 10. 流式与 tool_call 的协同

### 10.1 目标方案（推荐）

每一轮都走 `provider.Stream`，文本 delta 实时上屏、tool_call 在 Provider 内拼完整后 emit。用户体验：文字连续流出，中间穿插「🔍 正在检索」——与主流 AI 助手一致。代价：两家 Provider 都要实现**流式 tool_call 分片累积**（§6 表末列）。

### 10.2 v1 简化（备选过渡）

若某家流式 tool_call 分片累积一时难搞，可对该 Provider 临时走 **决策轮非流式、答案轮流式**：
- 工具决策轮用非流式 `Complete`（只为拿 tool_calls，好解析）；
- 判定无 tool_call（最终答案轮）时，要么再发一次流式请求（多一次答案 token），要么把非流式整段切片成假 `delta` 推送（短答案体感无差、零浪费）。

简化版砍掉「流式分片累积」这一最难点，作为先落地、后升级的过渡——接口不变，按 Provider 各自决定。OpenAI/Anthropic 都建议直接上目标方案。

## 11. 多轮对话接入点（前向兼容）

`RunRequest.Messages` 天然支持携带历史。启用多轮只需：
1. 前端 `chat.ts` 请求体从 `{question}` 扩为 `{question, history?}`；
2. Chat Service 把历史拼进 `Messages`。

并且 tool calling **天然解掉省略式追问**（"展开第三条""还有吗"）——模型带着历史自己发起新检索 query，无需 condense-question 改写。是否在本期开启由 scope 开关定，架构不阻塞。

## 12. 与现有代码的兼容/迁移

`agent` 包对外仅 **2 个调用点**，迁移面极小（两者迁移后都归入 `copilot` 域，见 §16.2）：

| 调用方（现状） | LLM 调用迁移 | 包归并 |
|---|---|---|
| `service/chat/chat.go:94` → `agent.GenerateStream(...)` | 改调 `agent.Run(...)`，消费 `AgentEvent`，新增工具与 SSE 事件 | → `service/copilot/chat.go` |
| `service/agent/agent.go:137` → `agent.Generate(...)`（近期总结，非流式、无工具） | 保留 `Generate` 为 `Run`（`Tools=nil, MaxRounds=1`）的薄封装，行为不变 | → `service/copilot/summary.go` |

- `AgentSetting` **不改**（工具是运行时注入，非配置），但**移除 `gemini` 协议**（§15.1）。
- **Wire**：包归并（agent/chat → copilot）会改 provider set 与注入，需 `just wire` 重新生成并 `just wire-check`。
- 可测性：`agent` 仍是无状态包级函数 + 工厂；如需注入 fake provider 做单测，可引入 `Runner` struct 持有 `providerFactory`。

## 13. 成本与护栏参数

- 封 3 轮下：典型"搜一次再答"≈ 2 次模型调用、token/延迟 **2–4×**；多搜 ≈ 4–6×（上下文累积近平方增长）。
- 缓解：system + tool 定义走 **prompt cache**（OpenAI/Anthropic 支持）；工具结果保持精简（仅文本快照）；查询去重；丢最旧 tool 结果。
- 新增可选环境变量（命名待定）：`ECH0_CHAT_MAX_ROUNDS`、`ECH0_CHAT_TOKEN_BUDGET`、`ECH0_CHAT_TOOL_TIMEOUT`。
- **`summarize_echos`（§18）的额外成本**：聚合一次最多拉 `maxAggregateEchos=5000` 条；map-reduce 的 LLM 调用数 ≈ 月份数（+1 次可选 reduce），但只在「窗口放不下」时触发——窗口足够（如 1M）则零额外调用、单次塞入。成本随 `AgentSetting.ContextWindow` 自适应。

## 14. 分期实施（建议顺序）

- **M0 — 包归并**：`handler/service` 层 `agent`（近期总结）+ `chat` → `copilot`（§16.2），`just wire` 重生成，对外路由契约不变。纯结构移动，先把目录摆对，后续都建在新布局上。
- **M1 — 抽象与骨架**：`types.go`（含 tool-aware Message）、`Provider` 接口、`providerFor` 工厂、空 `Run` 框架，编译通过；`Generate`/`GenerateStream` 改为薄壳，**两个旧调用点行为不变**（回归现有 Chat & 近期总结）。
- **M2 — 两家非工具流式 + Gemini 下线**：把现有流式逻辑迁进各 Provider 的 `Stream`（无工具路径），移除 Gemini 协议（§15.1）；Chat 走 `Run` 单轮，行为对齐现状。
- **M3 — 工具与 Loop**：`search_echos` 工具（Copilot Service 注入）、Loop 护栏、`AgentSearching/ToolResult`。**前端先抽 `service/request/sse.ts` 封装 + `buildCommonHeaders`（§17），再在其上加 SSE `searching` + `sources` 累积、状态条**。先 OpenAI 打通。
- **M4 — Anthropic 工具**：补 Anthropic tool 适配（流式 tool_use 分片累积）。
- **M5 — 打磨**：prompt cache、token 预算、多轮开关（可选）、`just check` + 两家联调。

## 15. 关键决策（已定）

| # | 议题 | 决策 |
|---|---|---|
| 1 | `search_echos` 是否暴露 `top_k` 给模型 | **固定 6，不暴露**。模型只传 `query`，减少自由度带来的不稳定，行为可预测。 |
| 2 | 是否加 `keyword_search`(FTS) 作第二工具 | **本期只做 vector search**。无 embedding 兜底议题另案处理，不进本重构。 |
| 3 | v1 是否开启多轮历史 | **先单轮验证工具体验**。历史接入点已预留（§11），跑通后再开多轮。 |
| 4 | Gemini 协议去留 | **整协议下线（方案 B）**。`agent` 仅保留 OpenAI 兼容 + Anthropic（见 §15.1）。 |
| 5 | 年终/区间总结怎么做 | **独立工具 `summarize_echos` + 窗口自适应 map-reduce**，不把 `search_echos` 撑成万能（见 §18）。 |
| 6 | 模型窗口如何感知 | **`AgentSetting` 新增可选 `ContextWindow`**（0=按 256k 保守默认），驱动聚合取数预算（见 §18.3）。 |

### 15.1 Gemini 整协议下线（方案 B，已定）

下线 Gemini 让 `agent` 更简洁：少一个 SDK、少掉最磨人的流式 tool_call 分片适配，而 OpenAI 兼容 + Anthropic 已覆盖绝大多数用户。**直接删除，不做迁移/校验/降级**——存量 `gemini` 配置自然落到已有的 `AGENT_PROTOCOL_NOT_FOUND` 错误路径（`providerFor`/`Generate` 的 `default` 分支），用户重配为 OpenAI 兼容即可。落地清单：

- 从 `internal/model/common` 移除 `Gemini` 协议常量。
- 删 `generateGemini` 及 `google.golang.org/genai` 依赖（`go.mod` / `go mod tidy` 清理）。
- `providerFor` 只剩 openai / anthropic 两个分支；其余协议走 `default` 返回 `AGENT_PROTOCOL_NOT_FOUND`。
- 前端协议选项移除 Gemini（i18n key 一并清理）。
- 对应 §14 的 M4 收敛为「补 Anthropic 工具适配」，无 Gemini。

### 15.2 仍开放

- `searching` 状态条是否可点击展开看模型的检索 query？（可观测性 vs 克制 UI）

## 16. 包结构与领域边界（决策）

### 16.1 `internal/agent` 是否拆子包？→ **不拆，flat 包 + 文件级内聚**

- 重构后约 6–8 文件、~800–1000 LOC，单一公共出口 `Run` + 类型集。Go 的封装单元是 **package**；拆子包会被迫把 `Message/ToolCall/Event` 等**导出并跨包共享**，反而削弱内聚、放大导出面，还可能引入 import cycle（Provider 子包要用核心类型，核心又要用 Provider）。
- "高内聚"靠**文件组织**达成，而非子包：`types.go`（抽象）/ `run.go`（Loop）/ `provider.go`（工厂）/ `provider_openai.go` / `provider_anthropic.go` / `agent.go`（兼容壳）。Provider 间共享的小工具（如 `toOpenAIRole`）保持**包内未导出**。
- 何时才值得拆：Provider 变成**可插拔**（外部注册自定义 provider），或单 provider 膨胀到自带子依赖/独立测试体量。本期都不成立——砍掉 Gemini 后更不成立。

### 16.2 chat / 近期总结 合并为一个 domain？→ **合并，命名 `copilot`**

> 修订自早期「不合并」判断。决定因素：从**产品域**看，Chat 与近期总结同属「Ech0 Copilot」——前端已有既定的「Copilot 面板」概念（见 commit `41281aa8` "设置归拢到 Copilot 面板"）。同一产品域，后端就该是同一个 domain。

- 合并为 `internal/{handler,service}/copilot`，吸收现有 `…/agent`（近期总结）与 `…/chat`（对话）两块。
- **同时消解命名异味**（见 16.3）：`agent` 从此只指 LLM 核心，`copilot` 指产品功能。技术能力（agent）与产品域（copilot）按**不同轴**分离，各自单一职责。
- **如何避免 god-service**：合并的是 *package*，不是把代码揉成一团。包内按文件分隔关注点：
  - `service/copilot/chat.go` —— SSE 流式 + 工具循环（`AskStream`）；
  - `service/copilot/summary.go` —— 缓存非流式近期总结（`GenRecent`）；
  - `service/copilot/prompt.go` —— 共享的 system prompt / persona / i18n 资产；
  - 二者共享检索 helper 与 `agent` 客户端，但**方法、文件分明**。
- **澄清**：chat 不是「纯前端展示形式」——它有真实独立的后端逻辑（agent loop + SSE），与 summary 的缓存非流式机制不同。所以 copilot service 是「一个内聚域、两个清晰分隔的操作」，而非一个方法两用。

**迁移映射**：

| 现状 | 迁移后 |
|---|---|
| `internal/{handler,service}/agent`（近期总结） | `internal/{handler,service}/copilot`（summary 部分） |
| `internal/{handler,service}/chat`（对话） | `internal/{handler,service}/copilot`（chat 部分） |
| `internal/event/subscriber/agent.go` | 触发 copilot 的 summary（可同步改名 `copilot.go`） |
| `internal/agent`（LLM 核心） | **不动**，仅内部重构（§16.1） |

路由前缀保持 `/api/chat`、近期总结端点不变（对外契约不破），仅后端包归并。

### 16.3 命名异味的消解

合并前，「agent」一词承载三义：(a) `internal/agent` LLM 核心、(b) `internal/service/agent` 近期总结 feature、(c) Agent loop / Agent 形态。

**§16.2 的 `copilot` 重命名一并解决之**：

- `internal/agent` → 只指 **LLM 核心**（provider 抽象 + Loop），语义唯一；
- 产品 feature → `copilot`（与前端「Copilot 面板」对齐）；
- "Agent loop / Agent 形态" → 作为 `agent` 包内部的技术概念，不再与 feature 名冲突。

## 17. 前端 SSE 传输统一

### 17.1 现状散落点（审计结论）

前端 HTTP 请求统一收口在 `web/src/service/request/`（ofetch 实例：`onRequest` 注入 auth/locale/timezone、`onResponseError` 统一翻译+toast、20s 超时、token 刷新），**WebSocket 也已封装**（`request/websocket.ts`）。但 **SSE 是唯一的例外**，两处各自手写且互不一致：

| 位置 | 传输 | 鉴权 | locale/tz | abort | 错误 |
|---|---|---|---|---|---|
| `service/api/chat.ts:30` `chatStream` | fetch + `getReader` + 手写 `\n\n` 解析 | `Authorization` 头 | ✅ 手写注入 | ✅ AbortController | 手写 callback |
| `views/panel/modules/TheSystemLog.vue:214` `startSSE`（WS 降级） | `EventSource` | **`?token=` query** | ❌ 缺失 | ❌ `es.close()` | 手写、不提示 |

后果：「请求公共头注入」逻辑有**三份真相**（ofetch / chat / systemlog），且已漂移。

**当初未封装的正当原因**（记录备查）：ofetch 默认整体解析 body，与流式读取冲突；`EventSource` 不能设自定义头（故系统日志只能走 query token）。但 ofetch 的**请求侧 `onRequest` 拦截器对流式仍有效**，公共头逻辑本可复用。

### 17.2 目标：SSE 归位到 `request/` + 公共头单一真相源

**① 抽公共头为唯一真相源**

```ts
// service/request/shared.ts
export function buildCommonHeaders(): Record<string, string> // Authorization + X-Locale + X-Timezone
```

ofetch 的 `onRequest` 与下文 `sseStream` 都调它，杜绝三份漂移。

**② 新增 `service/request/sse.ts`，统一 SSE 传输**

```ts
// 基于 fetch + ReadableStream（不用 EventSource——这样能带 Authorization 头、统一 abort）
export function sseStream<E = unknown>(opts: {
  url: string                                   // 走 getApiUrl()
  body?: unknown
  onEvent: (name: string, data: E) => void      // 解析后的 event:/data: 帧
  onError?: (msg: string) => void
  onClose?: () => void
}): { abort: () => void }
```

收口：`getApiUrl()` baseURL、`buildCommonHeaders()`、`AbortController`、`event:`/`data:`/`\n\n` 帧解析、idle 超时、错误兜底。

### 17.3 迁移

- `service/api/chat.ts` → 退化为薄壳：调 `sseStream`，把 `searching/sources/delta/done/error` 映射为类型化 handler（解析逻辑搬进 `sse.ts`）。
- `TheSystemLog.vue` 的 **SSE 降级**迁到 `sseStream`，鉴权从 query token 改为 `Authorization` 头（统一）。**WS 主通道仍留 `websocket.ts`**。
  - 注意权衡：`EventSource` 自带断线重连，fetch 版没有；系统日志降级若依赖自动重连，需在 `sse.ts` 或调用处补一层轻量重连，或保留该处 EventSource（二选一，落地时定）。
- 结果：三种传输（request / websocket / sse）全部归位 `request/`，公共头只有一份。

### 17.4 与本重构的衔接

此项与后端 Agent 化**正交但同区**：M3 本就要改 chat 的 SSE（新增 `searching`、`sources` 改累积）。**顺序上先抽 `sse.ts` 封装，再在干净封装上加新事件**，避免在旧手写 `chat.ts` 上叠加。故并入 §14 的 M3（见该节）。`sse.ts` 封装本身不依赖后端，可独立先行。

## 18. 区间聚合与长文总结（年终 / 月度回顾）

> 本节是工具体系落地后的增量设计：让「年终 / 月度总结」这类**穷举聚合**任务做对。实现见 `internal/service/copilot/{aggregate,budget,search,chat,prompt}.go`。

### 18.1 问题：top-k 检索 vs 穷举聚合的访问模式错配

`search_echos` 是为**点查（needle）**设计的 top-k 语义检索：`defaultTopK=6` 固定、不暴露给模型、无翻页入参，`maxRounds=3`，且早期 `queryEchos` **丢弃了 `QueryEchos` 返回的 `Total`**。当用户说「帮我写一篇年终总结」时，模型把日期设为全年 → 只拿回最新 6 条 → 基于 6~18 条生成一篇「看着很完整」的总结。这是最坏的一种失败——**静默截断**：模型不知道自己只看到冰山一角，用户也不知道，产出却言之凿凿。

根因：**top-k 语义检索（找最相关的几条）与穷举聚合（把某区间内全部读一遍再归纳）是相反的访问模式**。再加上 `AgentSetting` 原本没有窗口字段，后端无从得知模型是 256k 还是 1M，大窗口能力被浪费。

### 18.2 原则

1. **聚合是独立访问模式 ⇒ 独立工具**：新增 `summarize_echos`，不把 `search_echos` 撑成万能。两者并行注入，模型按意图选择（§18.5）。
2. **窗口自适应**：放得下就一次塞满（吃满大窗口），放不下才按月 map-reduce 分层——而非永远保守分层。
3. **绝不静默截断**：覆盖度（命中总数 / 纳入条数 / 分桶数 / 是否截断）随结果回报给模型与用户（§18.6），呼应 §8「防静默失败」。

### 18.3 `AgentSetting.ContextWindow` 与 token 预算

新增可选字段 `ContextWindow int`（token；0=未配置）。前端以 `256k`/`1m` 友好单位填写，解析成 token 数存储（`web/src/utils/tokenSize.ts`）。预算模型（`budget.go`）：

```
window  = ContextWindow>0 ? ContextWindow : 256_000   // 默认保守 256k
usable  = max(window * 0.6 - 8_000, 2_000)            // 扣掉 system/工具定义/成稿留白，留下塞物料的预算
```

度量复用 `estimateTokens`（rune 计数，CJK≈1/字），不引 tokenizer。

### 18.4 `summarize_echos` 工具

- 入参：`date_from`/`date_to`（必填，区间）、`tags`（可选限定主题）、`focus`（可选侧重）。
- `collectRange`：按 `created_at DESC` 分页（`PageSize=200`）穷举区间内**全部** Echo，累加到覆盖 `Total` 或触顶硬上限 `maxAggregateEchos=5000`（触顶因倒序保留**最近** N 条，并置 `truncated`）。
- `mapReduceSummary`（窗口自适应）：
  - `estimateTokens(全部) ≤ usable` → 直接返回完整格式化文本（`buckets=1`，大窗口在此吃满）；
  - 否则按 `YYYY-MM` 分桶，每桶 `agent.Generate` 浓缩成事实性摘要（**map**）；拼接后若仍超预算，再 `agent.Generate` 压一轮（**reduce**），保留每月要点与时间线。
  - map 阶段 v1 **顺序执行**（成本可预测）；bounded 并行留作后续（§18.8）。任一调用失败上抛，由 Loop 作为「工具执行失败」回喂模型自愈（§8）。
- 产出：`ToolOutput.Content` = 覆盖度抬头 + 聚合物料（供模型**写最终成稿**，工具不抢生成，tone/focus 可控、工具可复用）；`Meta` = `aggregateCoverage`。

### 18.5 prompt 路由

`chatSystemPrompt`（ZH/EN）声明两个工具的分工：点查用 `search_echos`，「某段时间的总结/回顾（年终/年度/季度/月度）」用 `summarize_echos`（覆盖全区间，据当前日期换算 `date_from/date_to`）。并把「1~2 次足够」收敛为**只约束 `search_echos`**，不误伤聚合。

### 18.6 SSE `coverage` 事件（加法兼容）

`AgentToolResult` 的 `Meta` 按类型分流：`[]SearchResult` → 既有 `sources`；`aggregateCoverage` → 新增 `coverage`（payload `{total, returned, buckets, truncated}`）。旧前端忽略未知事件，符合「加法不替换」。前端以「📚 已覆盖该区间 N 条」状态条展示，截断时改提示「已覆盖最近 N 条」。

### 18.7 顺带修 search_echos 的静默截断

`queryEchos` 改为同时返回 `Total`；当 `Total > 展示条数` 时，在工具结果前补一行如实告知「共命中 N 条、仅展示 M 条，要覆盖全部请用 summarize_echos」。点查路径也不再把「采样」当「全部」。

### 18.8 成本与未来优化

- map-reduce 的 LLM 调用次数 ≈ 月份数（+1 次可选 reduce），受 `maxAggregateEchos` 与「放得下直接塞」双重收敛；普通用户一整年常落在「直接塞、零额外调用」。
- **后续（本期不做，守轻量原则）**：① 月度 digest 缓存 + 事件失效（复用 `GetRecent` 的 `EchoCreated/Updated` 失效范式）大幅降重复年终请求成本；② map 阶段 bounded 并行降延迟；③ 独立「年度回顾」端点 + UI 入口（带缓存），把 chat 内的能力沉淀为一等公民功能。

## 19. 阻塞式提问与写操作确认门

本节是 Copilot 从「只读 RAG」扩展到「可管理 Echo」时新增的机制：一个可阻塞的提问原语，以及一道由 Loop 强制、模型无法绕开的写操作确认门。被否掉的备选方案记在 §19.7。

### 19.1 问题

Copilot 要能发布 / 修改 / 删除 Echo，就必须有一个人类确认环节。而 SSE 是单向的：问题送出去之后，同一条连接上收不到回答。原有的 `Tool.Execute` 只有「成功 / 失败」两种结局，Loop 里也没有任何可挂起的点。

### 19.2 决策一：确认由 Loop 强制，不靠 prompt

**prompt 不是门。** 「写之前先确认」这类指令，会在上下文变长、话题转换、或模型被说服之后被忘记，而唯一忘记的那一次就是一次没人同意的写入。所以确认必须是**类型与控制流**上的强制，而不是模型需要记住的约定：

```go
type ToolEffect uint8
const (
    EffectUnset ToolEffect = iota // 零值 = 最严格：Loop 拒绝执行
    EffectRead                    // 什么都不改
    EffectMutate                  // 改用户自己的数据，必须过确认
)

type Tool struct {
    Def      ToolDef
    Effect   ToolEffect
    Run      func(ctx, args) (ToolOutput, error) // 只读工具的函数体
    Mutation *Mutation                            // 写工具的函数体
    Interactive bool                              // 等人；EffectMutate 隐含为真
}

type Mutation struct {
    Plan    func(ctx, args) (Plan, error)   // 算出「会改什么」，并描述它；不得执行任何写
    Confirm func(ctx, Plan) (Decision, error) // 由调用方提供、由 Loop 调用
}
type Plan struct {
    Prompt any                                    // 对 agent 包不透明（领域零侵入，§2.2）
    Apply  func(ctx) (ToolOutput, error)          // 真正的写；只能被 Loop 在同意之后调用
}
type Decision struct { Approved bool; Refusal string }
```

关键在于 `Apply` **只存在于 `Plan` 里**，而 `Plan` 只由 `dispatchTool` → `applyMutation` 消费：从「模型发出 tool_call」到「数据库被改」这条路径上，没有一条不经过 `Confirm`。工具自己从不调用 `Confirm`，所以它没有「忘记调用」的可能。

`EffectUnset` 是零值，这个顺序是刻意的：以后有人加了个写工具却忘了声明，它**不会执行**，而不是默认放行。同理，`Mutation` 缺了 `Plan` / `Confirm` / `Apply` 任一半，或 `EffectRead` 却没有 `Run`，都按「本仓库有 bug」处理并拒绝执行（error 级日志 + `RunStrings.Malformed` 作为工具结果喂回模型）。

护栏由 `internal/agent/gate_test.go` 逐条钉住：拒绝路径、零值 `Decision`、`Confirm` 报错、未声明 effect、未知 effect、半声明 mutation、`Plan→Confirm→Apply` 的次序与各调用一次、以及「同一参数第二次仍然要重新问」。

### 19.3 决策二：谁提问 —— 通用 asker 在 copilot 层

`internal/agent` 仍然不知道「问题」长什么样（§2.2）。`Confirm` 与 `ask_user` 的实现都在 `internal/service/copilot`：

- `asker`（`ask.go`）绑定一次请求：一个 userID、一条事件流、一份本轮问答记录。绑定而非查找——能自己挑 asker 的工具，就是能回答别人问题的工具。
- 问题**不走 agent 事件流**，走 `AskStream` 另开的 `askEvents` channel。Loop 对提问一无所知，问题是从「停在工具调用里的那一点」发出来的。

### 19.4 决策三：会合点用进程内 channel，不用 DB 轮询

Ech0 是单进程单二进制，且 run 随发起它的请求一起结束——回答不需要跨副本送达，也没有请求结束后还活着的 run 需要恢复。所以会合点就是 `askRegistry`（`map[askID]*pendingAsk` + mutex），回答是零延迟的。

**否掉的备选：把 pending 状态落库 + 定时轮询。** 这是多副本部署下唯一可行的形状（提交回答的请求和等待它的 goroutine 可能落在不同实例上），代价是每个等待中的 round 每个 tick 一次行读、以及一套没人写得出来的孤儿清扫（run 崩掉时，没有任何活着的东西能去写那个过期标记）。Ech0 两个前提都不成立，落库只会买来这些代价而不解决任何问题。真要支持多副本，这里就得整体换成那个形状——而不是给 channel 打补丁。

「恰好只能回答一次」是**删除动作本身**保证的：`answer` 在同一把锁里查到 entry 并 `delete`，两个标签页抢同一次点击只有一个能拿到它。这是写谓词的语义，不是「先查再写」的检查——后者两个标签页会同时通过。未知 askID / 已回答 / 别人的 round 共用同一条错误信息，避免被用来探测活跃 round。

顺序上，**先登记后展示**：反过来的话，快到在 slot 建立之前就回答的客户端，它的回答会落在零行上并静默消失，而 run 会一直等一个再也不会来的答复。

### 19.5 决策四：等人的时间不花运行预算

`ECH0_AGENT_TIMEOUT_SECONDS`（默认 120s）是**生成预算**，尺度对着模型和 API 延迟。任何紧到能抓住卡死 provider 的期限，人一定会超过。所以：

- `budget`（`run.go`）持有 deadline，按步派生 context，而不是给整个 run 一个 deadline；
- interactive 调用在 `b.base`（请求 context）下执行，结束后把等待时长 `credit` 回 deadline；
- 人的等待另有预算：`ECH0_AGENT_ASK_TIMEOUT_SECONDS`（默认 300s）。

**超时不自动选择**。`Recommended` 只是个标记，永远不会被代选——沉默不能变成没人给过的同意。超时的结局是 `ask_closed` + 工具报错，模型据此用自己的话告诉用户。

### 19.6 SSE 与端点（加法兼容，§2.4）

新增两个事件：

| 事件 | payload | 含义 |
|---|---|---|
| `ask` | `{ask_id, questions:[{id,text,header?,detail?,options?,multi?,recommended?}]}` | run 停下来等人 |
| `ask_closed` | `{ask_id}` | 没等到回答就不等了（超时 / 取消）；客户端撤掉 picker |

回答走另一条已认证请求：`POST /api/chat/answer`，body `{ask_id, answers:[{question_id, selected?, custom?}]}`，`ScopeAdminSettings`。`ask_id` 是**核对**而非信任的——过期标签页手里那个 picker 的点击必须被拒绝，而不是覆盖已经给出的回答。

`ChatMessage` 新增 `asks: []AskExchange`（问了什么 + 选了什么）。保留两半，因为决定是这一轮里**唯一由人做出**的部分；只存结果的话，「这次删除是谁批的」在重开会话后就无从回答了。`persistTurn` 因此也把「有 asks」算作这一轮有内容——停在确认上的一轮可能既没有正文也没有 sources，丢掉它就等于抹掉那次交互。

`detail` 是写确认的关键：它是**具体的那条 Echo**（ID / 内容 / 标签 / 发布日期），以预格式化文本渲染。「删除 Echo 019a…」不是一个人能同意的东西。

### 19.6.1 写工具的 id 从哪来（一个已修的坑）

`update_echo` / `delete_echo` 要的是 Echo 主键，而**唯一能给出它的地方是 `search_echos` 的工具结果**。原来那份结果只有位置编号：

```
【1】(2026-03-01) 今天读了三体
```

于是「把第一条改一下」这件事，模型会老老实实地传 `id: "1"`——那是它看到的唯一像 id 的东西。`GetEchoById("1")` 查不到，模型于是告诉用户「这条 Echo 不存在」。错的不是它的推理，是我们没把主键给它。

现在每条结果都带上真实主键，`formatSearchResults` 输出：

```
【1】(2026-03-01) id=019ce0ea-82dd-774f-ae2d-5445512d42ad 今天读了三体
```

编号保留（模型在正文里说「第 1 条」时用得上），但主键是照抄用的那一份。这份格式同时用于**跨轮历史回放**（`historyForModel` 里上一轮 sources 的那段），所以下一轮说「删掉刚才那条」也拿得到 id。

另一半是**强制**：`echoIDArg` 在 Plan 阶段校验 id 必须是 UUID，不是就直接报错，且报错文案指名道姓地说「照抄 id= 后面那串」。这一步在读库之前、提问之前，所以传错编号既不会读到东西、也不会弹确认；而模型读到这条工具错误后能在同一次 run 里自我纠正。schema 与 prompt 也各写了一遍，但那两处是给注意力有限的模型看的，校验才是那条不会忘的。

### 19.6.2 写工具能写什么（以及为什么未知字段必须报错）

`create_echo` / `update_echo` 只写三样东西：**正文、标签、可见性**。一条 Echo 还有图片（`EchoFiles`）、扩展卡片（`Extension`：音乐/视频/网站/位置/GitHub 项目）、`Layout`，这些都不在对话可写的范围内 —— 附件要先走上传，扩展卡片要先解析外链，两者都属于界面里的流程，不是一次工具调用能凑出来的东西。

边界本身不新鲜，`encoding/json` 的默认行为才是问题所在：**未知字段被静默丢掉**。模型传 `{"content":"...","files":[...]}`，`json.Unmarshal` 不报错，工具照样弹确认、照样写入、照样返回「已发布」，于是模型对用户说「已经帮你带图发出去了」。这句话是假的，而链路上没有任何一环会否认它。

所以 `decodeWriteArgs` 用 `DisallowUnknownFields`：传了写不了的字段就报错，报错文案说明「图片和扩展卡片无法通过对话创建或修改，请在界面里操作」。这与 §19.6.1 的 id 校验是同一类东西 —— schema 和 prompt 也各写了一遍，但那两处是给注意力有限的模型看的，解码器才是不会忘的那一层。

两段解码的顺序是有意的：**先宽松、后严格**。JSON 截断、类型不对是模型自己的手误，必须原样回传；只有「字段不存在」才套上那句解释。反过来会把 `unexpected EOF` 说成「不支持这个参数」，把模型送去找一个它压根没传的字段。

`update_echo` 这一侧还有个更硬的约束：仓储层的 `UpdateEcho` 会**先删后建** files 与 extension 行，所以工具必须把 `GetEchoById` 读到的整条 Echo 原样带进写入（`next := *current`），confirmation 没展示的字段一个都不能丢。改一句正文顺手抹掉配图，是这个功能能犯的最严重的错误 —— `TestUpdateEchoTool_CarriesUnshownFieldsThrough` 就是守这个的，fixture 里特意放了 file、music 扩展和非默认 layout。

### 19.7 否掉的备选方案

| 备选 | 为什么否 |
|---|---|
| **prompt 里写「写之前先确认」** | 不是门。上下文一长、话题一转、或模型被说服，这条指令就没了；而唯一忘掉的那一次就是一次没人同意的写入。确认必须是控制流上的强制。 |
| **只给 `ask_user`，让模型自己在写之前调它** | 同上，而且更糟：它看起来像有门。模型少调一次，写入照样发生，而日志里看不出区别。 |
| **一个 `EffectApproval` 档：这类工具 Loop 直接拒绝、永不执行** | 对 Ech0 没有对应物。真的要发/改/删，需要的是「问过之后执行」，不是「一律拒绝」。为不存在的动作建的审批分支是永不执行的代码，而真正决定行为的会变成那个默认值。`EffectUnset` 保留了 fail-closed 的默认，但它守的是「下一个忘了声明的工具」，不是某个虚构的动作类别。 |
| **`Confirm` 由工具自己调用（`Execute` 里 plan → ask → apply）** | 最初的实现就是这样，被否了：它把「先问再写」变成每个写工具都要记得遵守的约定，而不是一条它无法绕开的路径。第四个写工具直接调 `Apply` 就能悄悄绕过，且编译通过、测试也不会红。现在 `Apply` 只存在于 `Plan` 里，只有 `applyMutation` 能拿到它。 |
| **落库 + 轮询的会合点** | 见 §19.4。 |
| **超时自动选中 `Recommended`** | 沉默不能变成没人给过的同意。超时的结局必须是「什么都没发生」。 |
| **picker 上加「跳过 / 关闭」** | 关掉 picker 只会留下一个仍在等待、却已无处回答的 run。放弃这次回答就是放弃这次回答——Stop 已经是那个动作了。 |
| **写确认的选项标签由模型生成** | `consented` 要拿人的选择和肯定项标签作比较，标签由模型给就等于让模型定义什么算同意。标签一律由 `writeStringsFor` 提供。 |
| **打字内容与选项标签做匹配** | 打「删除」两个字是在说话，不是点了按钮。对写确认，打字永远不是同意，只作为「为什么没写」回传给模型。 |
| **未知字段沿用 `encoding/json` 的静默丢弃** | 见 §19.6.2。丢掉 `files` 换来的是一句「已帮你带图发布」——工具成功、模型撒谎、用户被骗，三者都不报错。 |
| **在 args 结构体里逐个声明 `files` / `images` / `extension` 再判空** | 等于用穷举去追模型的拼法，漏一个就退回静默丢弃。`DisallowUnknownFields` 守的是「schema 里没有的一切」，不需要预判模型会怎么编字段名。 |
| **让 `update_echo` 支持改图片/扩展卡片** | 附件要先上传、扩展卡片要先解析外链，都不是一次工具调用能凑出来的；硬塞进来只会让确认界面开始展示模型没能力保证的东西。 |

### 19.8 前端契约

picker 内联在 transcript 里，**从不是模态框**：它挡住的是助手，不是读者——读者必须还能往回滚，看清那条促成这个问题的回答再做选择。没有关闭按钮：把 picker 关掉只会留下一个仍在等待、却已无处回答的 run，而 Stop 就在下面。`label` / `description` 只作为文本渲染，不参与任何客户端分支。一轮多问逐题回答，POST 在最后一题时**只发一次**。
