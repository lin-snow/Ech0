---
title: Semantic search (embeddings)
description: Build a vector index of past Echoes so AI chat retrieves by meaning — includes dimension-matching troubleshooting
---

**Semantic search** builds a vector index of your past Echoes so [AI chat](/docs/guide/chat) can find content by **meaning** rather than by literal text.

::: tip Optional feature
Chat works fine without it. The difference is retrieval quality:

- **Off**: retrieval falls back to **keyword substring matching** — rephrase something (searching for “feeling low” when you wrote “pretty down”) and it may find nothing.
- **On**: **semantic recall** — close in meaning is enough to match, which suits recollection-style questions like “did I ever write about X?”
  :::

Configure it under **System settings → Copilot → Vector index**. The current version supports only **OpenAI-compatible `/v1/embeddings`** endpoints, which covers the vast majority of providers including OpenAI, Qwen/DashScope, Ollama, and Jina.

---

## Settings

| Item                    | Description                                                                                                                                        |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Enable indexing         | Master switch                                                                                                                                      |
| Embedding model         | For example `text-embedding-3-small`, Qwen `text-embedding-v4`, and so on                                                                          |
| Vector dimensions (Dim) | **Must match the model's actual output dimensions**; see “Dimension matching” below                                                                |
| API key                 | Access key; may be left empty for local services such as Ollama                                                                                    |
| Custom API URL          | The service's **API base URL** (for example `https://api.openai.com/v1`). `/embeddings` is appended automatically — **do not add it yourself**     |
| Batch size              | Number of texts vectorized per request when rebuilding the index. Some providers cap at 64 or 25 and error beyond that; `0` uses the default of 64 |

---

## Dimension matching (the easiest trap to fall into)

The vector index table fixes the dimensions at creation time to the `Dim` you enter, so **the dimensions the model actually returns must equal `Dim`**, otherwise writes fail:

```
Dimension mismatch ... Expected 1024 dimensions but received 2048
```

This means the index table was created with `1024` dimensions, but the model returned `2048` this time. How to get it right:

- **Models that support custom dimensions** (OpenAI `text-embedding-3` family, Qwen `text-embedding-v4`, etc.): `Dim` can be any tier the model supports (such as `1024` or `1536`), and Ech0 requests that dimension.
- **Models that do not support custom dimensions**: `Dim` **must be the model's native dimension** — check the vendor documentation, for example `1536`.

::: tip What if I change the model or dimensions?
Changing the **embedding model** or **Dim** invalidates the existing index. Ech0 prompts you to **rebuild the vector index** — follow the prompt and **rebuild once** to regenerate vectors for all past Echoes with the new model and dimensions.
:::

---

## First enablement: rebuild the index

Echoes published before you enabled the feature have no vectors yet. Once configured, open “Vector index” and click **Rebuild index (Start rebuild)** to backfill vectors for all past Echoes:

- Progress is shown (indexed / total / failed) and you can **cancel** at any time.
- Afterwards, **creating / editing / deleting** Echoes maintains the index incrementally, with no further manual rebuilds.
- A rebuild is only needed again after **changing the model or dimensions**.

---

## Cost notice

Every Echo added to the index and every chat retrieval calls the embeddings API once, which may incur charges. Choose the model and provider that fit your needs, and follow their terms.
