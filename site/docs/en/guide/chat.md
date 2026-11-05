---
title: AI chat
description: Talk to your own Echoes in natural language — RAG retrieval plus multi-turn tool calls
---

**AI chat** lets you talk to your own past Echoes in natural language. It does not answer from thin air: the model **first retrieves, aggregates, and counts your Echoes, then answers from them**, citing the sources it hit. Under the hood it is "RAG retrieval plus multi-turn tool calls".

::: tip This feature is for the site owner
`POST /chat` requires **administrator privileges**; **only the site owner / administrators can use it**. Visitors cannot see or call it. Your Echo content is not exposed to the public by this feature.
:::

---

## Prerequisites

1. **Required**: first configure the [generation model](/docs/guide/agent) in **System settings → Copilot → Model** (enable Agent and set protocol/model/key). This is the brain behind chat.
2. **Optional**: also configure [semantic search](/docs/guide/embedding). With it, chat finds past content by **meaning** (a paraphrase still matches); without it, it falls back to **keyword** search, which still works.

> The custom **Prompt** under "Model" **does not affect chat**; it applies only to the recent summary.

---

## How to use it

- **Floating entry on the home page**: after signing in, use the chat entry in the bottom-right corner of the home page and start asking.
- **Dedicated chat page**: open the Chat page for multi-turn conversations.

Not sure what to ask? Try one of these:

- "Summarize what I have been paying attention to and thinking about recently"
- "How has my mood changed over this period"
- "Round up the scattered ideas I have written down"
- "What did I write about reading last year"

---

## What it can do

Before answering, the model calls the following capabilities as needed (multiple times and in combination within one conversation):

| Capability          | Purpose                                                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Echo retrieval      | Find the most relevant Echoes by query (semantic/keyword), tags, and date range                                          |
| Range summary       | Aggregate **all** Echoes within a period, for longer-span retrospectives such as year-end, quarterly, or monthly reviews |
| Statistics overview | Report totals and common tags for a range, and similar high-level facts                                                  |

So for questions like "help me write a year-end review" or "what was I busy with in September", it picks the appropriate way to gather data rather than just sampling a few Echoes.

---

## The process is visible

Answers are returned as a **stream**, with the process in the open:

- **Retrieving**: shows what the model is searching for.
- **Source citations**: lists the Echoes hit in this turn; click through to the original to check.
- **Token-by-token output**: the answer appears as it is generated.

If **multimodal support** is enabled and the model accepts image input, images attached to retrieved Echoes are passed to the model for understanding.

---

## Conversation and privacy

- Conversations have **context memory**; you can clear the session and start over at any time.
- On multi-user instances, retrieval is strictly **scoped to Echoes you published yourself**; other people's content never enters the context.
- Calls to third-party LLMs may incur charges and send relevant Echo excerpts to the provider you configured; follow their terms and choose the model accordingly.
