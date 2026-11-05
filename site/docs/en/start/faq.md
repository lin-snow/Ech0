---
title: FAQ
description: Deployment, permissions, storage, and integration Q&A
---

If your situation isn't covered here, search or ask in [GitHub Discussions](https://github.com/lin-snow/Ech0/discussions).

---

## Product & positioning

### What is Ech0?

A lightweight, open-source, **self-hosted** publishing tool: post short notes, links, and media to a single **timeline**, with your data on your own server. It emphasizes calm reading and full ownership. For a fuller product description, see [Product overview](/docs/guide/overview).

### What is Ech0 not?

It is not a **knowledge base or team wiki** like Obsidian / Notion; it is closer to a **personal feed** in the spirit of status updates — not a professional note-taking workflow.

### Does it cost anything?

It is open source and free, under **AGPL-3.0**, with no ads and no subscription lock-in.

---

## Deployment, data & backup

### How do I back up and import?

Use **snapshot export / migration import** in the panel. In normal operation, back up your mapped **data directory** (for example Docker's `/opt/ech0/data`) on a regular schedule. If object storage is enabled, media lives in the storage backend and the database lives locally or wherever your deployment puts it. See [Data control](/docs/guide/datacontrol).

### Storage layout, switching buckets, local ⇄ object storage migration?

See [Object storage (S3)](/docs/guide/s3) and [Data control](/docs/guide/datacontrol); always back up before making changes.

---

## Reading, RSS & Connect

### Is RSS supported?

Yes. Subscribe with an RSS reader to the feed URL provided by your instance (the exact entry point follows the current interface).

### Why can't others see my Connect avatar?

In **System settings → Service address**, enter the full URL of **your own instance** (it must include `http://` or `https://`), matching the address you use in the browser.

### Why is the Connect list incomplete?

The backend fetches each instance's information individually; **instances that cannot be reached are dropped**, so only successfully fetched entries are shown. See [Federation & aggregation](/docs/guide/federation).

---

## Accounts, permissions & posting

### Why can't I post?

In the current version, **posting permission is restricted to high-privilege accounts by default**. **The first account registered** is the Owner; other users need an administrator to grant permission in settings. On first deployment, make sure you are signed in as the Owner.

### Is the permission model complicated?

It uses a lightweight model (Owner / Admin / regular user, and so on) intended to be sufficient for day-to-day use with minimal fuss; the settings page is authoritative for details.

---

## Comments, sign-in & storage

### How do I enable comments?

Turn them on in **Comment settings / Comment management**, and configure moderation and captcha as needed. See [Comments](/docs/guide/comment).

### How do I configure Passkey / OAuth?

See [Single sign-on (SSO)](/docs/guide/sso).

### How do I configure S3?

Fill in the endpoint, credentials, bucket, and so on in storage settings; `endpoint` generally should **not** include an `http/https` prefix. If the frontend needs to access media directly, configure the bucket/CDN access policy properly. See [Object storage (S3)](/docs/guide/s3).

### What is MetingAPI?

An API used by **music cards** to resolve direct streaming links; if left empty, the default resolver is used. In production, prefer an endpoint you control.

---

## Automation, API & third parties

### AI summary, webhooks, API?

- Agent: **System settings → Agent**, see [AI models & recent summary](/docs/guide/agent).
- Webhook: [Webhooks](/docs/guide/webhook).
- General API: call it with an [access token](/docs/guide/accesstoken); docs live at **`/swagger/index.html`**.
- AI clients (Cursor and the like) connecting to the instance: see [MCP integration](/docs/guide/mcp); a token with the **MCP (AI Agent)** audience is required.

### Can automation or AI post comments?

You can use the **integration comment API** described in the README together with an **access token** (it needs `comment:write` and the `integration` audience, among others). See the instance's **`/swagger/index.html`** and [Comments](/docs/guide/comment).

### Third-party integration platform disclaimer

Third-party integration services that are not officially authorized are **outside the scope of official support**; any issues arising from their use are the responsibility of the user and the third party.
