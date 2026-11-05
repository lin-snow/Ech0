---
title: AI models & recent summary
description: Configure the AI generation model (Agent) that powers the recent summary and AI chat
---

All AI capabilities in Ech0 **share a single "generation model" configuration**. Set it up once in **System settings → Copilot → Model** and it drives both:

- **Recent summary**: the large model generates a short summary from your latest Echoes, so visitors can quickly catch up on what is happening (covered in the lower half of this page).
- **[AI chat](/docs/guide/chat)**: talk to your own Echoes in natural language.

If you also want chat to search your archive by _meaning_, configure [semantic search](/docs/guide/embedding) as well (optional).

---

## Configuration

| Item           | Description                                                                                                                                                                        |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Enable Agent   | Master switch. When off, neither the recent summary nor AI chat is available                                                                                                       |
| API protocol   | **Only two options**: **OpenAI compatible**, **Anthropic**                                                                                                                         |
| Model name     | Such as `gpt-4o`, `deepseek-chat`, `claude-3-5-sonnet`, etc.; refer to the vendor's documentation                                                                                  |
| API Key        | Access key (do not disclose)                                                                                                                                                       |
| Base URL       | Address of a self-hosted or compatible gateway; for services such as DeepSeek, Qwen, and Ollama, put their endpoint here                                                           |
| Prompt         | Custom system prompt. **Applies only to the recent summary**; AI chat does not use it                                                                                              |
| Multimodal     | When enabled, **AI chat** will pass attached images to the model when a retrieved Echo contains them (the configured model must support image input)                               |
| Context window | Enter values such as `256k` or `1m`; leave empty for the default `256k`. A larger window favors reading everything at once, a smaller one switches to layered summarization sooner |

::: tip You are choosing a protocol, not a vendor
What you select here is the **API protocol**, not a specific vendor. DeepSeek, Qwen, Tongyi, Ollama, and all kinds of self-hosted compatible gateways speak the **OpenAI compatible** protocol — choose it and use `Base URL` plus **Model name** to point at the concrete implementation. Anthropic (Claude) uses its own native protocol.
:::

The exact field names follow the UI.

---

## Recent summary

Once enabled, Ech0 draws context from your **recent Echoes** and calls the model to generate a short summary for public display.

- To avoid hammering the API with repeated calls in a short window, summaries are **cached**; publishing a new Echo invalidates the cache and it is regenerated on the next visit.
- The style and length of the summary can be adjusted with the **Prompt** above.

### HTTP endpoints (excerpt)

| Method  | Path                  | Description                                         |
| ------- | --------------------- | --------------------------------------------------- |
| GET     | `/api/agent/recent`   | Fetch the publicly displayed recent summary         |
| GET     | `/api/agent/info`     | Query metadata such as whether the Agent is enabled |
| GET/PUT | `/api/agent/settings` | Read or update the configuration (sign-in required) |

See **`/swagger/index.html`** for the complete definitions.

---

## Cost and compliance

Calls to third-party LLMs may incur charges; follow the terms of the provider you choose, and never put keys or private data in prompts.
