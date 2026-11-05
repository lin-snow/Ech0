---
title: Single sign-on (SSO)
description: Configuring and troubleshooting third-party OAuth, OIDC, and Passkey (beginner-friendly)
---

"Single sign-on" means you don't need to remember a separate password just for Ech0: you can sign in with a **GitHub / Google / QQ** account, or go passwordless with a **Passkey (fingerprint, face, or security key)**. In the admin panel, go to **System settings → SSO** to configure **OAuth2** and **Passkey** separately; both can be enabled at the same time, and users choose on the sign-in page as needed.

---

## Who this document is for

- You are the **site owner/administrator** and want to enable third-party sign-in or Passkey on your instance.
- You do **not** need to understand the OAuth protocol in detail first; the sections below start with "what happens", then give the steps.
- Regular users only need to click the corresponding button on the sign-in page; **linking a third-party account** is done in personal settings.

---

## Understand two things first

### 1. What OAuth2 / OIDC is doing (in plain terms)

Think of the flow as "going to a bank counter for a co-branded card":

1. The user clicks "Sign in with GitHub" on Ech0.
2. The browser **jumps to GitHub**, where the user signs in and **authorizes** permissions such as "allow Ech0 to read your public profile".
3. GitHub brings the user **back** to your Ech0 site, along with a one-time **authorization code**.
4. The Ech0 server exchanges the authorization code with GitHub for a **token**, fetches the user's profile, **creates or matches a local account**, and finally completes the sign-in.

The **redirect URI** is "where GitHub must send the user back to" in step 3 — it must **exactly match** the address shown in Ech0 (scheme, domain, path, and trailing slash must all match), otherwise you get **Redirect URI mismatch**.

**OIDC** can be understood as "an upgraded OAuth": besides the access token, it also provides an **id_token** (JWT), which suits enterprise IdPs (such as Keycloak, Authentik). After enabling the OIDC-related options in Ech0, fields such as Issuer and JWKS are validated.

### 2. What Passkey is doing (in plain terms)

A Passkey replaces the password with **biometrics or a hardware security key on your device**: the browser and Ech0 communicate using the **WebAuthn** standard.  
There are two easily confused settings:

| Setting                       | In plain terms                                                                         | Example                    |
| ----------------------------- | -------------------------------------------------------------------------------------- | -------------------------- |
| **RP ID (WebAuthn RP ID)**    | Which "website" owns this key; usually the **domain**, and **do not** write `https://` | `memo.example.com`         |
| **Origins (allowed origins)** | The full URL from which the browser may use this key; must include the scheme          | `https://memo.example.com` |

Both must match the domain **actually used** in the browser's address bar. Changing domains, using only an IP, or mixing HTTP and HTTPS will all cause registration or sign-in failures.

---

## Before you start, confirm

- You are signed in with an **administrator account** and can reach **System settings → SSO**.
- For production, use **HTTPS** with a **fixed domain** (OAuth and Passkey often run into problems with bare IPs or self-signed certificates).
- When creating an application on a third-party platform, you need to be able to receive email or use a developer account (for example GitHub, Google Cloud, QQ Connect).

---

## OAuth2: from creating an app to signing in

### What the callback address looks like

Ech0 uses this path (note: **no** `/api` prefix):

```text
https://your-site-domain/oauth/<provider>/callback
```

Here `<provider>` can only be **`github`**, **`google`**, **`qq`**, or **`custom`** matching the template (commonly used for custom/OIDC).  
The **OAuth2** page in the admin panel shows the **auto-generated Redirect URI**; **copy it verbatim** into the third-party platform, and **do not change the path or add slashes**.

### What to fill in Ech0

1. Create an OAuth application on the third-party platform and obtain the **Client ID** and **Client Secret**.
2. Open **System settings → SSO → OAuth2**, enable the switch, and choose a template: **GitHub / Google / QQ / Custom**.
3. Fill the **Redirect URI** shown on the page back into the third-party platform.
4. Enter the **Client ID / Secret** into Ech0; set **Scope** to just what is needed, for example:
   - GitHub: `read:user` (read public profile)
   - Google: commonly `openid`, `email`, `profile` (follow the current console requirements)
5. If you use **OIDC** (a custom IdP), complete **Issuer, JWKS URL**, and so on in the interface; after saving, the server validates the **id_token**.

### Creating an app on each platform (entry points and caveats)

**GitHub**

1. Open GitHub → **Settings → Developer settings → OAuth Apps → New OAuth App**.
2. For **Authorization callback URL**, enter: `https://your-domain/oauth/github/callback` (matching what the admin panel shows).
3. After creating it, note the **Client ID**, generate a **Client Secret**, and store it safely.

**Google**

1. Create a project with the [Google Cloud Console](https://console.cloud.google.com/) and enable the **Google+ API / user info service** (follow the console's current naming).
2. Under "Credentials", create an **OAuth 2.0 Client ID** with application type **Web**.
3. For **Authorized redirect URIs**, enter: `https://your-domain/oauth/google/callback`.

**QQ**

1. Create a website application at [QQ Connect](https://connect.qq.com/); after it passes review, obtain the **APP ID / APP Key**.
2. For the **callback address**, enter: `https://your-domain/oauth/qq/callback` (matching the Ech0 admin panel).

**Custom / OIDC**

1. Choose the **Custom** template and fill in the **authorization URL, token URL, and user info URL** according to the IdP documentation.
2. When enabling OIDC, fill in **Issuer, JWKS**, and set scopes as required by the IdP.

### "Authorization callback" and CORS (when changes are needed)

If you access the same backend from **multiple frontend domains** (for example `https://a.com` and `https://b.com`), you may need to configure **Auth Redirect Allowed Return URLs** and **CORS Allowed Origins** in the OAuth advanced settings; otherwise the **redirect** or **cross-origin requests** after a successful sign-in will be rejected.  
For a typical **single-domain self-hosted** setup, the defaults are fine; after changing them, save and make sure they match the site address used by the reverse proxy and environment variables.

### Linking to an existing account

After a user **has signed in by password or another method**, they can **link** a GitHub/Google identity to the current account in **personal settings**, avoiding a duplicate account.

---

## Passkey: configuration steps and common pitfalls

### Configuration order (recommended)

1. **System settings → SSO → Passkey**.
2. Fill in the **WebAuthn RP ID** (your domain, no scheme) and **WebAuthn Origins** (for example `https://your-domain`, multiple entries allowed).
3. After saving, the interface should show a status such as **Passkey ready**.
4. **Register** a Passkey in a browser that supports WebAuthn (the system prompts for fingerprint, face, or inserting a security key).
5. On the same page you can **manage** registered devices (rename, delete).

### Common errors

| Symptom                                 | Possible cause                                                                                               |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| Cannot register / sign in with Passkey  | RP ID doesn't match the current domain, or Origins doesn't include the current full URL                      |
| Access by IP only                       | WebAuthn usually requires a **secure context** (HTTPS or localhost); for public access, use a domain + HTTPS |
| Everything breaks after a domain change | Reconfigure RP ID / Origins with the new domain and have users register their Passkeys again                 |

When you need to call the HTTP API, open **`/swagger/index.html`** on your deployed instance to see the Passkey-related endpoint documentation.

---

## Using OAuth and Passkey together

Both can be enabled at once: users can sign in with a third-party account, and after signing in they can also link a Passkey for next time. The exact buttons and wording follow the sign-in page of the current version.

---

## Troubleshooting quick reference

| Error or symptom                       | What to check                                                                                                                                    |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Redirect URI mismatch**              | The callback configured on the third party must match Ech0's **character for character**; check `http`/`https`, the trailing `/`, and subdomains |
| **Invalid client**                     | Client ID/Secret copied incorrectly, or the app is not enabled or has not finished review                                                        |
| OAuth redirects back but sign-in fails | Check whether the scope includes the permission needed to fetch user info; for OIDC, check Issuer/JWKS                                           |
| Passkey does not respond               | Whether the browser supports it; whether the site uses HTTPS; whether RP ID/Origins match the address bar                                        |

For more detailed endpoint and field descriptions, refer to the **OpenAPI (Swagger)** on your deployed instance.
