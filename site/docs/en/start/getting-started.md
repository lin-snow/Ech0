---
title: Getting started
description: Your first Ech0 deployment — from preparing the environment to publishing a post
---

If you have just heard of Ech0, read the [Product overview](/docs/guide/overview) first, then work through this page in order. The following assumes you have a Linux server or a local machine that can run Docker, and the goal is to open the web page and finish your first login **within 10 minutes**.

---

## What you will end up with

When you finish this section, you will have:

- An Ech0 instance reachable in a browser (default port **6277**)
- An **Owner (first registered)** account that can manage the site and posting permissions
- A basic grasp of “where the data lives and how to back it up”

If you care more about **concepts** than commands, jump straight to [Terms & interface](#Terms--interface).

---

## Before deploying: what you need

| Item                | Notes                                                                                                                                                         |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A machine           | A cloud server, NAS, Raspberry Pi, or local machine will do; it needs **Docker** (recommended) or the ability to run the official binary.                     |
| Network             | Trying it locally: just open `http://127.0.0.1:6277`. Public access: allow **6277** (or your mapped port) in the security group / firewall.                   |
| A domain (optional) | You can use the IP without a domain; for HTTPS, configure certificates in a reverse proxy (Nginx, Caddy, etc.). See [Installation](/docs/start/installation). |
| Signing secret      | Nothing to prepare: the server generates the JWT secret on first boot and stores it in the database, so restarts and upgrades keep sessions valid. |

You **do not** need to know Go or frontend development beforehand; self-hosted users just run the container as described.

---

## Step 1: Start it with one command (Docker)

Run this in a terminal (identical to [Installation](/docs/start/installation)):

```bash
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  sn0wl1n/ech0:latest
```

- `-v` stores data on the host at `/opt/ech0/data`, so recreating the container does not lose the database; change the path as needed.
- If the port is taken, change the left-hand `6277` to something like `8080:6277`, then use the new port in the browser.
- There is no secret to paste: the first boot creates the signing key and keeps it in the database. `JWT_SECRET` is no longer read — if it is still set, the server ignores it and logs a warning.

**Common reasons it will not open**: the firewall does not allow the port, Docker is not running, or the IP is wrong (cloud servers need the **public IP** or the bound domain).

---

## Step 2: Finish initialization in the browser

1. Open `http://<your IP or domain>:6277`.
2. Follow the wizard to **register the first account**. That account automatically becomes the **Owner** (highest privileges).
3. Go to the admin panel or settings page and finish the basic settings (site name, service address); the signing secret needs no attention because it was generated and stored automatically at first boot.

In the current version, **posting permission is usually granted only to high-privilege accounts by default**; regular registered users may only be able to browse until an administrator grants access in settings. If you “logged in but cannot post,” first check whether you are signed in as the Owner, or see the [FAQ](/docs/start/faq).

---

## Step 3: A few things to do right away

1. **Know how to rotate the signing key** (only if you ever need to): delete the `jwt_secret` row with `sqlite3 data/ech0.db "DELETE FROM key_values WHERE key='jwt_secret';"` and restart — every issued token is invalidated, so all users sign in again.
2. **Back up the data directory**: copy at least one copy of `/opt/ech0/data` (or your mount path) somewhere else. See [Data control](/docs/guide/datacontrol).
3. **Fill in the “service address”** (if you use Connect or avatars): in **System settings → Service address**, enter the full URL of this site (including `http://` or `https://`), matching the browser address bar.
4. **Plan for HTTPS**: for public sites, terminate TLS in a reverse proxy; internal-only deployments can stay on HTTP for now.

---

## Terms & interface

| Term                      | Meaning                                                                                                                                         |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **Echo**                  | One timeline post (like a single social media update); it may contain Markdown, images, link cards, and so on.                                  |
| **Instance**              | This Ech0 service that you deployed; multiple instances can federate via Connect / Hub, see [Federation & aggregation](/docs/guide/federation). |
| **Owner / administrator** | The first registered account is the Owner; it can manage users and some site-wide options in settings.                                          |

---

## What to read next (by interest)

| What you want to do…                             | Documentation                                                                 |
| ------------------------------------------------ | ----------------------------------------------------------------------------- |
| Upgrade safely, switch image versions            | [Updating](/docs/start/update)                                                |
| Write posts, add images, use extended cards      | [Writing & editing](/docs/guide/editor)                                       |
| Change the site logo, footer, personal avatar    | [Preferences & profile](/docs/guide/preferences)                              |
| Turn on comments, prevent spam                   | [Comments](/docs/guide/comment)                                               |
| GitHub / Google login, Passkey                   | [Single sign-on](/docs/guide/sso)                                             |
| Upload attachments to the cloud, save disk space | [Object storage](/docs/guide/s3)                                              |
| Call the API from scripts, automation, AI (MCP)  | [Access tokens](/docs/guide/accesstoken) · [MCP integration](/docs/guide/mcp) |
| Search before asking                             | [FAQ](/docs/start/faq)                                                        |

Contributors should also read the [Developer guide](/docs/dev/guide).

---

## Version differences

This documentation focuses on **the getting-started path and feature explanations**. Button positions and field names may differ slightly between versions; the **interface of your own instance** and its OpenAPI (`/swagger/index.html`) are authoritative.
