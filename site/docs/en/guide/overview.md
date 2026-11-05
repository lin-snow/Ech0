---
title: Product overview
description: What Ech0 is, who it is for, and what it can do
---

# Product overview

## What is Ech0?

Ech0 is a **self-hosted personal microblog**: your content lives on a **timeline** where it can be shared and commented on, and the data stays on your own server. Think of it as “a lightweight status site on your own domain” — built for publishing and reading, not for heavyweight note-taking or team collaboration.

Compared with tools like Memos that help you “jot down a thought quickly,” Ech0 focuses on **what happens after you write**: publishing the timeline publicly or semi-publicly so others can keep reading and interacting. It is not a backlinked knowledge base, nor a Notion-style team workspace.

---

## Start here if you are new

1. **Just want it running**: follow [Getting started](/docs/start/getting-started) from Docker to your first login (about 10 minutes).
2. **Want to deploy directly**: see [Installation](/docs/start/installation) (Docker / Compose / binary / Helm).
3. **Hitting a problem**: check the [FAQ](/docs/start/faq) first, then search [GitHub Issues](https://github.com/lin-snow/Ech0/issues).

---

## Who Ech0 is for

**A good fit** if you want to:

- Post **short texts, links, and images** on your own machine or domain, forming a timeline that others can see
- Keep **your data on your own server** instead of a closed cloud notebook run by someone else
- Have **RSS**, optional **comments**, and optional **federation between instances**, without maintaining a forum or blog engine

**Not a great fit** if what you mainly need is:

- An Obsidian-style **backlinked knowledge base** or a card-note workflow
- Notion-style **multi-person collaboration**, complex permissions, and databases
- **Purely private memos** with no intention of ever publishing a timeline

---

## Deployment and feature index

| What you want to do…                                           | Documentation                                                                                                            |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| First deployment, ports, and data directory                    | [Installation](/docs/start/installation)                                                                                 |
| From zero to publishing your first post                        | [Getting started](/docs/start/getting-started)                                                                           |
| Upgrade the image or move to a newer minor version             | [Updating](/docs/start/update)                                                                                           |
| Merge timelines across sites                                   | [Federation & aggregation](/docs/guide/federation) (Connect + `/hub`)                                                    |
| Third-party login and Passkeys                                 | [Single sign-on](/docs/guide/sso)                                                                                        |
| Access tokens, MCP (AI access)                                 | [Access tokens](/docs/guide/accesstoken) · [MCP integration](/docs/guide/mcp)                                            |
| Manage the instance and moderate comments from a terminal      | [CLI](/docs/guide/cli)                                                                                                   |
| Chat with your own Echoes, recent summary                      | [AI chat](/docs/guide/chat) · [AI models & recent summary](/docs/guide/agent) · [Semantic search](/docs/guide/embedding) |
| Site logo, footer, avatar, and panel preferences               | [Preferences & profile](/docs/guide/preferences)                                                                         |
| Comments and moderation                                        | [Comments](/docs/guide/comment)                                                                                          |
| Upload attachments to the cloud                                | [Object storage](/docs/guide/s3)                                                                                         |
| Backup and migration                                           | [Data control](/docs/guide/datacontrol)                                                                                  |
| Switch instances, long-term preservation, static archive sites | [Capsules & static sites](/docs/guide/capsule)                                                                           |

---

## Capabilities at a glance

- **Writing**: a Markdown timeline with images, links, and extended cards for music, video, and more (see [Writing & editing](/docs/guide/editor)).
- **Appearance & profile**: site logo, footer, service address, Meting, custom CSS/JS; personal avatar and interface language (see [Preferences & profile](/docs/guide/preferences)).
- **Data**: data lives in a local directory by default; attachments can use **S3-compatible** storage; supports snapshot export and migration from sources such as v3; you can export human-readable **capsules** for moving house and long-term preservation, and compile them into a [static archive site](/docs/guide/capsule) in one click.
- **Interaction**: built-in comments; **RSS**; **Connect / Hub** aggregates public content from multiple instances.
- **Accounts**: OAuth, Passkey; scripts and integrations call the HTTP API with **access tokens**, and AI clients connect to the instance over **MCP** (see [Access tokens](/docs/guide/accesstoken) and [MCP integration](/docs/guide/mcp)).
- **AI**: [chat with your own Echoes](/docs/guide/chat) in natural language and generate a [recent summary](/docs/guide/agent); optionally enable [semantic search](/docs/guide/embedding). You need to bring your own LLM (OpenAI-compatible / Anthropic).
- **Automation**: **Webhook** event pushes and similar, enabled in settings as needed.

For endpoints and fields, your instance's **`/swagger/index.html`** is authoritative; they may differ between versions.

---

## License and open source

Open source under **AGPL-3.0**, with no ads and no mandatory cloud service. If you redistribute it or offer it as a network service, comply with your AGPL obligations (for example, making the corresponding source available to users).
