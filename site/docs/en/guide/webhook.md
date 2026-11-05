---
title: Webhooks
description: What webhooks are, how to configure them, request format, signature verification, and retries (beginner-friendly)
---

**Webhooks** can be explained in one sentence: when something happens in Ech0 (a new post, a new user signing up), your server can automatically send an **HTTP POST** request to **a URL you provide**, carrying the **event type and JSON data**.  
That lets you connect Ech0 to **Slack, Discord, your own bots, audit systems** and more, without polling over and over asking “is there anything new?”

---

## Who this document is for

- You're an **administrator** who needs to configure a receiving URL in the admin panel.
- You're about to write a few lines of code in **your own service** to receive the POST (or use a tool such as Zapier or n8n that can expose an HTTPS URL).
- You don't need to know distributed systems first; below we cover **what to fill in**, **what the request looks like**, and **what counts as a successful delivery**.

---

## Configuring it in the admin panel

1. Sign in with an **administrator account** and go to **System settings → Webhook**.
2. **Create** one: fill in a **name**, a **receiving URL** (`http` or `https`), and optionally a **Secret** (used for signature verification; strongly recommended in production).
3. Save and **enable** it. The list shows the status of the most recent delivery, success or failure.
4. When editing: **the Secret is usually not echoed back**; leaving it blank on submit may mean **clearing the signature** (check the hints in the current version's interface).

If you prefer managing it through the API, you can do that after admin authentication (paths depend on your deployment's `/api`): `GET/POST/PUT/DELETE /webhook` and similar; see **Swagger** on your instance for details.

---

## Why you can't use an internal address as the URL

To prevent **SSRF** (server-side request forgery against internal networks), Ech0 rejects obviously unsafe receiving addresses, for example:

- `localhost` and private IPs (such as `10.x`, `192.168.x`, `127.0.0.1`)
- Hostnames ending in `.local`

So **`http://127.0.0.1:3000` listening on your own laptop can't be entered directly**.  
For local debugging, use a **publicly reachable tunnel domain** (a temporary HTTPS address from ngrok, Cloudflare Tunnel, and so on), and configure signature verification the same way as in production.

---

## How delivery success and failure are determined

- **Success**: your server returns a **2xx** status code for the HTTP request.
- **Failure**: network errors, timeouts, and 4xx/5xx responses all count as failures.

When a delivery fails, the client **retries automatically within a short window** (for example 3 attempts in total, spaced roughly 500ms → 1s → 2s, with **a timeout of about 5 seconds per request**; the “test” action in the admin panel may use a shorter retry policy).  
If those immediate retries all fail, the delivery is recorded as failed (visible in the webhook status) and the event is **not** queued for later redelivery.

**Recommendation**: return 2xx **as soon as possible** on receipt, and **push slow work into an async queue**, so you don't exceed 5 seconds and trigger repeated retries.

---

## Which events you'll receive

Below is the **complete list** (matching the current backend allowlist). Only these **topics** are pushed to webhooks as business events:

| Topic                                                            | Meaning (plain language)                                   |
| ---------------------------------------------------------------- | ---------------------------------------------------------- |
| `user.created` / `user.updated` / `user.deleted`                 | User created, profile changed, deleted                     |
| `echo.created` / `echo.updated` / `echo.deleted`                 | Echo published, edited, deleted                            |
| `comment.created` / `comment.status.updated` / `comment.deleted` | Comment created, status changed (e.g. moderation), deleted |
| `resource.uploaded`                                              | Resource/file upload completed                             |
| `system.snapshot` / `system.export`                              | Related to snapshot or export jobs                         |
| `system.snapshot_schedule.updated`                               | Snapshot schedule was modified                             |

Note: comment- and moderation-related behavior is best understood alongside [Comments](/docs/guide/comment); snapshot topics relate to the scheduled jobs described in [Data control](/docs/guide/datacontrol).

---

## What the HTTP request looks like

### Request method and security headers

- **Method**: `POST`
- **`Content-Type`**: `application/json`
- **`User-Agent`**: `Ech0-Webhook-Client`
- **`X-Ech0-Event`**: the event topic (for example `echo.created`)
- **`X-Ech0-Event-ID`**: the event ID for this delivery (a string, usable for **idempotent deduplication**)
- **`X-Ech0-Timestamp`**: Unix timestamp (seconds, UTC)
- **`X-Ech0-Signature`**: present only when a **Secret** is configured, formatted as `sha256=<lowercase hex string>`

### JSON body structure

```json
{
  "topic": "echo.created",
  "event_name": "EchoCreated",
  "payload_raw": {},
  "metadata": {},
  "occurred_at": 1710000000
}
```

| Field         | Meaning                                                          |
| ------------- | ---------------------------------------------------------------- |
| `topic`       | Matches the `X-Ech0-Event` header                                |
| `event_name`  | Server-side event type name (e.g. the Go struct name)            |
| `payload_raw` | **The core business data** (JSON); its structure varies by event |
| `metadata`    | Extra key-value pairs, possibly empty                            |
| `occurred_at` | When the event occurred (UTC, Unix seconds)                      |

For the complete fields and nested structures, refer to the current version and **`/swagger`**.

---

## How to verify a request really came from Ech0 (signature verification)

If you configured a **Secret**, compute an **HMAC-SHA256** over the **raw request body bytes** (the full body before JSON parsing) using that same Secret, then compare it with **constant-time equality** against the hex string after `sha256=` in the `X-Ech0-Signature` header.

**Node.js example:**

```javascript
import crypto from "node:crypto";

function verifyEch0Signature(rawBody, secret, signatureHeader) {
  if (!signatureHeader?.startsWith("sha256=")) return false;
  const received = signatureHeader.slice("sha256=".length);
  const expected = crypto
    .createHmac("sha256", secret)
    .update(rawBody)
    .digest("hex");
  return crypto.timingSafeEqual(
    Buffer.from(received, "utf8"),
    Buffer.from(expected, "utf8"),
  );
}
```

**It is strongly recommended that you also:**

- Check whether **`X-Ech0-Timestamp`** falls within an acceptable time window (to mitigate replay attacks).
- Deduplicate with **`X-Ech0-Event-ID`**, processing each ID only once.

---

## Testing connectivity

When you run **Test** on a webhook in the admin panel, Ech0 sends a test event of type **`webhook.test`** (to validate the URL, the certificate, and the signature chain) and updates that webhook's most recent status.  
If you use the API, it's typically the admin endpoint `POST /api/webhook/:id/test` (see Swagger).

---

## FAQ

**It always shows as failing?**  
Check whether the receiving URL is publicly reachable, whether the TLS certificate is trusted by the client, whether you return 2xx **within 5 seconds**, whether a firewall/WAF is blocking it, and whether the verification Secret matches the one configured in Ech0.

**A particular business event never arrives?**  
Make sure the webhook is **enabled**; make sure the event is in the **topic allowlist** above; make sure the URL isn't rejected by security policy (for example an internal address entered by mistake).

**Should the receiver return 500 if its own business logic fails?**  
Not advisable. On the webhook side, return 2xx as soon as you've **acknowledged and persisted** the event; retry downstream failures in your own queue. Otherwise Ech0 treats it as a delivery failure and triggers immediate retries.

---

## Receiver best practices (summary)

- **Idempotency**: deduplicate by `X-Ech0-Event-ID`.
- **Authentication**: verify `X-Ech0-Signature`.
- **Replay protection**: restrict the `X-Ech0-Timestamp` window.
- **Fast responses**: enqueue first, process asynchronously.
- **Observability**: log topic, event_id, HTTP status, and latency.
