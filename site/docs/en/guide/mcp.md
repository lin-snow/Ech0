---
title: MCP integration
description: Connect Ech0 to AI clients such as Cursor and Claude
---

Once MCP is connected, you can read and write your own Ech0 right from an AI client: browse the timeline, post, read comments, and manage webhooks without switching to the web UI.

Ech0 ships its own MCP server on the same address and port as the site, so there is nothing extra to deploy.

---

## Before you start

- The instance is reachable (HTTPS recommended).
- An **administrator account**: the first authorization must be confirmed in the browser.
- A client that supports **remote MCP**, such as Cursor or Claude Desktop. Clients that only support local stdio need a gateway to forward requests.

---

## Option 1: browser authorization (recommended)

1. Add a remote MCP server in the client, using `https://your-domain.com/mcp` as the URL.
2. The client opens a browser; sign in with the administrator account and confirm the authorization.
3. Back in the client, the tool list appears and you are ready to go.

- Clients that support remote OAuth discover the authorization entry point automatically. If the client asks you to pick an authentication method, choose **OAuth / browser login** instead of pasting a token.
- The authorization page lists the permissions the client requests; they take effect only after you confirm, and you can revoke them at any time under **System settings → Access tokens**.
- Authorization can only be completed by an administrator in the browser; a token cannot replace that confirmation.

---

## Option 2: manual token

For clients that do not support OAuth, or when you only want to use it in a script.

1. Open **System settings → Access tokens** and create a new token.
2. Set **Audience** to **MCP (AI Agent)**.
3. Check the permissions you need (see the table below); an expiry of 8 hours or 1 month is recommended.
4. Configure the client like this:

```json
{
  "mcpServers": {
    "ech0": {
      "url": "https://your-domain.com/mcp",
      "headers": {
        "Authorization": "Bearer <your token>"
      }
    }
  }
}
```

For local debugging, use `http://localhost:6277/mcp` instead.

| What you want it to do | Check these scopes              |
| ---------------------- | ------------------------------- |
| Read the timeline      | `echo:read`                     |
| Create and edit Echoes | `echo:write`                    |
| Delete Echoes and tags | `echo:delete`                   |
| Read private Echoes    | plus `echo:private`             |
| Read / write comments  | `comment:read`, `comment:write` |
| Delete comments        | plus `comment:delete`           |
| Files                  | `file:read`, `file:write`       |
| Federated instances    | `connect:read`, `connect:write` |
| Manage webhooks        | `admin:webhook`                 |

Do not check admin permissions beyond your current purpose.

---

## What you can ask it to do

- Find Echoes: by keyword, tag, or date, or pull a random one
- Create, edit, and delete Echoes, and like them
- Read and write comments
- Look up files and tags
- Inspect federated instances
- Manage webhooks

Capabilities you lack permission for do not appear in the client's tool list; on clients that support confirmation, destructive operations ask you first.

---

## FAQ

**Connected but no tools?** The token lacks permissions, or the authorization was not completed. Check the permissions or authorize again.

**403?** Wrong audience. A manual token must use the **MCP (AI Agent)** audience.

**401, or it keeps asking you to sign in?** The token expired, or the client's OAuth flow did not finish; authorize again.

**What URL do I use?** `https://your-domain.com/mcp` — do not append `/api`.

**The client only supports local commands?** Use a gateway or proxy to forward requests to the HTTP address above.

---

## Advanced

- Protocol: prefer **2026-07-28** (stateless); 2025-11-25 / 2025-06-18 / 2025-03-26 are also supported.
- Only `POST /mcp` is available; `GET` and `DELETE` return 405, and the server does not issue sessions.
- Capability manifest: `GET /api/mcp/manifest` (requires `mcp:read`), returning tool names, permissions, and whether each is destructive.
- Authorization discovery: `/.well-known/oauth-protected-resource/mcp`, `/.well-known/oauth-authorization-server`.

---

## Further reading

[Access tokens](/docs/guide/accesstoken) · [CLI](/docs/guide/cli) · [Webhooks](/docs/guide/webhook) · [Comments](/docs/guide/comment)
