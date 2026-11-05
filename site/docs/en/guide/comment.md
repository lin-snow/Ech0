---
title: Comments
description: Built-in comments, moderation, captcha, and automation
---

Ech0 uses a **built-in comment system** — no third-party comment hosting such as Twikoo or Waline is needed.

---

## Enabling and entry points

Under **System settings → Comment settings** (or the comment-related entry in the panel):

- Turn on the **global comment switch**.
- Optionally enable **post-moderation** (guest comments are shown publicly only after an administrator approves them).
- Optionally enable **captcha** (lowers the risk of automated spam; the exact implementation depends on your version).

**Moderation, deletion, and bulk operations** for comments are generally done in **Admin panel → Comment management**.

---

## Where comments come from

Common types include:

- **Guest comments**: left by visitors who are not signed in. A nickname is usually required, and an email in some scenarios; whether approval is mandatory is decided by your settings.
- **Signed-in users**: behavior and permissions follow role policies (see the UI).
- **Integration API**: automation or upstream systems can write comments through a dedicated API (see “Third-party integration” below); the source is distinguishable in the admin panel.

Comment length limits, whether rich text is allowed, and similar details depend on the **current version and the OpenAPI spec**.

---

## Moderation and statuses

Comments go through a flow like “pending → approved / rejected” (the exact status names depend on the UI). Administrators can:

- Filter by status and keyword.
- Approve, reject, or delete a single comment or a batch.
- Mark standout comments as “hot” and the like (if your version provides it).

Single comments can also be moderated from the terminal: `ech0 comment panel list --status pending` → `ech0 comment panel status --id <id> --status approved` → `ech0 comment panel delete --id <id> --yes`. See [CLI](/docs/guide/cli).

---

## Anti-spam (approach)

An instance may combine form validation, rate limiting, duplicate-content detection, captcha, and more. Switches and thresholds live in **Comment settings**; for exact parameters, check your instance's **`/swagger/index.html`**.

---

## Third-party integration (AI / automated comments)

To post comments **on your behalf** from your own scripts, n8n, or an upstream AI, use the integration endpoint documented in the repository README (path of the form `POST /api/comments/integration`): create a token under **Access tokens** with the required permissions such as **`comment:write`** and **`audience` set to `integration`**, then send `Authorization: Bearer <token>` in the request headers. Tokens minted by CLI login have `audience=cli` and are accepted by the same endpoint.

**Do not** commit that token to a public repository. For the full request body, rate limits, and field descriptions, see the **OpenAPI (/swagger)** on your instance.

---

## Webhooks

Comment creation, status changes, deletion, and similar events may trigger `comment.*` events in [Webhooks](/docs/guide/webhook), making it easy to sync them to Slack or your own services.
