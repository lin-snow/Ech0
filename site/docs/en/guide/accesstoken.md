---
title: Access tokens
description: Long-lived credentials for scripts, the CLI, and AI clients
---

An access token is a password you put in a request header, letting scripts, command-line tools, or AI clients call Ech0 as you without storing your login password.

Create them in **System settings → Access tokens**; **only administrators** can create or delete them.

---

## When you need one

| What you want to do                          | What to do                                                                                                                                |
| -------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Use the `ech0` command line                  | Generated automatically at login; no need to create one manually, see [CLI](/docs/guide/cli)                                              |
| Call the API from a script / n8n             | Create a token and choose **Public client**                                                                                               |
| Post comments from an external system        | Create a token, set the audience to **System integration**, and check `comment:write`                                                     |
| Connect an AI client such as Cursor / Claude | Prefer browser authorization; if unsupported, create a token with the audience **MCP (AI Agent)**, see [MCP integration](/docs/guide/mcp) |

---

## Creating a token

1. An administrator opens **System settings → Access tokens**.
2. Enter a name, expiry, audience (who it is for), and permissions.
3. Copy the entire token immediately after creation, **it is shown only once**.

You can delete a token from the list at any time. Deleting it revokes it, and it cannot be recovered.

---

## Using a token

Include it in the request header:

```http
Authorization: Bearer <token>
```

The exact endpoints and fields are those listed at **`/swagger/index.html`** on your own instance; they may differ slightly between versions.

---

## Choosing an audience

| UI name                | Internal value  | Where it is used                                                                                |
| ---------------------- | --------------- | ----------------------------------------------------------------------------------------------- |
| **Public client**      | `public-client` | General API calls outside the browser, small in-house tools                                     |
| **Command line**       | `cli`           | The `ech0` CLI and scripts                                                                      |
| **System integration** | `integration`   | Comment integration: `POST /api/comments/integration`                                           |
| **MCP (AI Agent)**     | `mcp-remote`    | Only for connecting to `/mcp`; tokens issued by MCP browser authorization are also of this kind |

Using the wrong audience is rejected. For example, a token with the **Public client** audience calling `/mcp` returns 403.

---

## Choosing permissions

Grant the minimum required. The groups in the panel match the table below:

| Group    | Available permissions                                                                       | Description                                                 |
| -------- | ------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Echoes   | `echo:read` / `echo:write` / `echo:delete` / `echo:private`                                 | Read public / publish and edit / delete / also read private |
| Comments | `comment:read` / `comment:write` / `comment:review` / `comment:delete` / `comment:settings` | Read / post / moderate / delete / comment settings          |
| Files    | `file:read` / `file:write` / `file:delete`                                                  | Read / upload and edit / delete                             |
| Connect  | `connect:read` / `connect:write` / `connect:delete`                                         | Read / add / delete                                         |
| Profile  | `profile:read` / `profile:write`                                                            | Read / modify your profile                                  |
| Site     | `admin:site`                                                                                | Branding, CORS, system settings                             |
| Storage  | `admin:storage`                                                                             | S3 credentials and connection tests                         |
| Sign-in  | `admin:auth`                                                                                | OAuth2 / Passkey policy                                     |
| Webhook  | `admin:webhook`                                                                             | Webhook management                                          |
| AI       | `admin:agent`                                                                               | Agent / Embedding / Copilot                                 |
| Data     | `admin:data`                                                                                | Migration, import/export, snapshot schedules                |
| Ops      | `admin:ops`                                                                                 | Logs, update checks, traffic statistics                     |
| Users    | `admin:user`                                                                                | User management                                             |
| MCP      | `mcp:read`                                                                                  | Read-only MCP capability list, no tool calls                |

Two rules that are easy to miss:

- **Write does not include delete**: `write` covers publishing and editing only; deleting requires checking `*:delete` separately.
- Account security (Passkeys, OAuth linking), access token management, and device approval only accept browser sign-in; tokens cannot call them.

---

## Common combinations

| What you want to do         | Check                                                      |
| --------------------------- | ---------------------------------------------------------- |
| Read-only timeline access   | `echo:read`                                                |
| Post from a script          | `echo:read`, `echo:write`                                  |
| Comment integration         | `comment:write` (audience set to **System integration**)   |
| Everyday MCP read and write | `echo:read`, `echo:write`, `comment:read`, `comment:write` |
| Manage webhooks             | `admin:webhook`                                            |

---

## Expiry

| Option        | Suited for                                     |
| ------------- | ---------------------------------------------- |
| About 8 hours | Temporary scripts, one-off tasks               |
| About 1 month | Long-term personal automation                  |
| No expiry     | High risk; use only in controlled environments |

Tokens stop working once they expire. If you suspect a leak, delete the token and create a new one immediately.

---

## Security recommendations

- Treat a token like a password: never commit it to a public repository or share it in screenshots; use environment variables or a secret manager in CI.
- Prefer short expiry over long, and read-only over write access.
- `admin:*` tokens are very powerful; grant them only to fully trusted environments.
