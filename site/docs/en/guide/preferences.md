---
title: Preferences & profile
description: Site icon and avatar, site-wide information and personal identity — why they live in two places, and how to configure them
---

Have you ever wanted visitors to land on your site and **know at a glance that this is your place** — while your avatar and name stay **consistent** whenever you post or comment? In Ech0, "what the whole site looks like" and "what your account looks like" are separate: **the former** gives visitors and federation scenarios a stable public face, **the latter** lets you be recognized on the timeline. This page follows that logic, with fewer menu paths and more on "what it actually does for you".

---

## First, decide: are you changing the "space" or "yourself"?

In the admin sidebar, the entries related to appearance and identity are mainly in two places:

| What you want to change                                                                  | Where to go                                              |
| ---------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| Site logo, title, service address, footer, default language, open registration, and more | **Preferences** (`/panel/setting`) → **System settings** |
| Your avatar, username, password, email, interface language                               | **User center** (`/panel/user`) → **User settings**      |
| Tokens for scripts / API (unrelated to appearance)                                       | The same **Preferences** page → **Access tokens**        |

In System settings, click **Edit**, make your changes, then click **Apply**; User settings works the same way.

---

## Preferences → System settings: what does this site say to the outside world?

What you set here decides the world seen by **visitors who are not signed in** (as well as federation and metadata) — not the preferences of any single user.

1. **Site icon, your first impression** — The uploaded site logo appears as the site icon; it is usually shown in the homepage top bar when **signed out**. When **signed in**, the top bar often switches to **your personal avatar**, so don't forget to set the site icon separately — otherwise visitors see this site missing "its badge". Images use the same storage as attachments (local or [object storage](/docs/guide/s3)).

2. **Site title and service name** — The title mostly appears in the browser tab and SEO; the service name is mostly used as the **display name** on the timeline, in federation, and so on. One can be formal and the other friendly, or they can be identical — follow your own branding habits.

3. **Service address, your street address on the internet** — Enter the full URL used for external access (with `https://`). Federation and other sites fetching your avatar or instance information often depend on this address; if it is wrong, or you changed domains without updating it, you may see problems such as "others can't see your avatar". See [FAQ](/docs/start/faq) and [Federation](/docs/guide/federation) for details.

4. **Footer: a closing line** — Copy and optional links (ICP filing, about page, and so on) give the page a quiet ending.

5. **Meting API: the resolution behind music cards** — Enter a trusted Meting service address; in production, a self-hosted or otherwise controlled endpoint is preferable, so that a key experience is not tied to an unfamiliar service.

6. **Custom CSS / Script: only for those who know what they're doing** — Styles fine-tune the mood; scripts are more powerful but affect every visitor-facing page. **Only trusted administrators** should edit these; malicious or broken scripts can cause anomalies or even security risks (XSS), so back up before changing anything.

7. **Default language and open registration** — The default language affects visitors, as well as users who have not yet changed the language in their profile; **Allow registration** decides whether self-service sign-up is available (other policies follow the current version).

---

## Preferences → Access tokens

Tokens are used by automation, scripts, and [MCP](/docs/guide/mcp); they are not responsible for "how things look". See [Access tokens](/docs/guide/accesstoken) when you need them.

---

## User center: who are you?

What you set here only affects **the currently signed-in account**; it does not change the site-wide settings for everyone else.

1. **Avatar, how you are recognized** — It appears in comments, on profile-related displays, and in the top bar after signing in. After uploading, click **Apply** in the interface to save; follow the prompt shown in the current version.

2. **Username, password, email** — Used for signing in and display; whether leaving the password blank means "do not change" depends on the interface; the email is used for contact or recovery (if the relevant capability is enabled).

3. **Interface language, your own interface only** — This only affects the language you see; it does **not** change the site-wide **default language** set in System settings.

---

## At a glance

| What you want to change                                                                                            | Where to go                       |
| ------------------------------------------------------------------------------------------------------------------ | --------------------------------- |
| Site icon, site title, service address, footer, Meting, custom styles/scripts, default language, open registration | **Preferences** → System settings |
| Personal avatar, your account details, your interface language                                                     | **User center** → User settings   |
| API tokens                                                                                                         | **Preferences** → Access tokens   |

If the interface wording differs from this document, **the interface on your instance and the OpenAPI (`/swagger/index.html`) take precedence**.
