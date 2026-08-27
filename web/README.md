# Ech0 Web (Frontend)

The Vue 3 SPA for [Ech0](../README.md) — a self-hosted personal microblog / timeline platform.
This package is built with Vite, type-checked with `vue-tsc`, and its production output is emitted to
[`../template/dist`](../template/dist), where the Go binary embeds and serves it from the same process
that exposes the REST API.

> For repo-wide architecture (Wire DI, event bus, storage abstraction, contributing rules) see the root
> [`CLAUDE.md`](../CLAUDE.md) and [`CONTRIBUTING.md`](../CONTRIBUTING.md).
> This README focuses on what lives under [`web/`](.) and how to develop against it.

---

## Tech stack

| Area              | Choice                                                                    |
| ----------------- | ------------------------------------------------------------------------- |
| Framework         | **Vue 3.5** (SFC, `<script setup>`, Composition API)                      |
| Build / dev       | **Vite 8** + `@vitejs/plugin-vue 6`                                       |
| Language          | **TypeScript 6** with `vue-tsc` for `.vue` type-checking                  |
| Routing           | `vue-router` 5                                                            |
| State             | **Pinia** 3 (composition-style stores)                                    |
| HTTP client       | **ofetch** (universal `$fetch`) + custom interceptor layer                |
| Styling           | **UnoCSS** with `@unocss/preset-wind4` + SCSS design tokens               |
| i18n              | `vue-i18n` 11 (`legacy: false`), zh-CN / en-US / de-DE                    |
| Markdown / Editor | `markdown-it` + `highlight.js` for rendering, **Vditor** for authoring    |
| Media             | `photoswipe`, `aplayer` / `meting`, `vue-virtual-scroller`, `gsap`        |
| Uploads           | `@uppy/core` + `aws-s3` / `xhr-upload` / `compressor` / `dashboard`       |
| Notifications     | `vue-sonner` (toast), `floating-vue` (popovers/tooltips)                  |
| Captcha           | `@cap.js/widget` (registered as a Vue custom element)                     |
| Avatars           | `@dicebear/core` + `@dicebear/micah`                                      |
| Lint / format     | ESLint 10 (`@vue/eslint-config-typescript`, `eslint-plugin-vue`) + Prettier 3 |
| Tests             | **Vitest 4** + `@vue/test-utils` + `jsdom`                                |
| Package manager   | **pnpm 10** (see `packageManager` field)                                  |

**Engine:** Node `>=26.0.0` (declared in [package.json:6-8](package.json#L6-L8)). Use Volta / nvm to pin.

---

## Quick start

```bash
# from web/
pnpm install
pnpm dev          # http://localhost:5173 (proxies API to backend on :6277)
```

Run the Go backend in another terminal so API requests resolve:

```bash
# from repo root
just dev          # Air hot-reload on :6277
# or
just run
```

The frontend talks to the backend via the URL configured in `VITE_SERVICE_BASE_URL`
(see [Environment variables](#environment-variables)). In dev that points at
`http://localhost:6277`; in production it's a relative path (`/`) because the Go
binary serves the SPA from the same origin.

---

## Scripts

All scripts come from [`package.json`](package.json). Run them inside `web/` (or via the
`just web …` module recipes from the repo root).

| Script                  | What it does                                                                     |
| ----------------------- | -------------------------------------------------------------------------------- |
| `pnpm dev`              | Vite dev server with HMR + Vue DevTools plugin                                   |
| `pnpm build`            | Parallel `type-check` + `build-only` via `npm-run-all2`                          |
| `pnpm build-only`       | `vite build` → emits to [`../template/dist`](../template/dist) for Go embed      |
| `pnpm type-check`       | `vue-tsc --build` (project references via `tsconfig.json`)                       |
| `pnpm preview`          | Serve the built `dist` locally for smoke-testing                                 |
| `pnpm test:unit`        | Run the Vitest suite once (CI mode)                                              |
| `pnpm lint`             | `eslint . --fix`                                                                 |
| `pnpm lint:style`       | `stylelint "src/**/*.{vue,css,scss}" --fix` — CSS/SCSS/Vue style blocks (runs in `just check`) |
| `pnpm format`           | `prettier --write src/`                                                          |
| `pnpm i18n:check`       | Composite: runs all four i18n guardrails below                                   |
| `pnpm i18n:key-check`   | en-US / de-DE must mirror the zh-CN key tree (no missing, no extras)             |
| `pnpm i18n:unused-check`| Flag locale keys that no source file references                                  |
| `pnpm i18n:hardcoded-check` | Flag user-visible strings in `.vue` / `.ts` that bypass `t()`                |
| `pnpm i18n:pseudo-smoke`| Render with a pseudo-locale to surface untranslated text and layout overflow     |
| `pnpm token:check`      | Scan for hardcoded credentials / leaked tokens in source                         |

The i18n scripts live at [`web/scripts/`](scripts/) and are mandatory before opening a PR
(`just check` from the repo root chains them in).

Single test (file or by name):

```bash
pnpm exec vitest run tests/editor/markdown.spec.ts
pnpm exec vitest run -t "escapes html"
```

---

## Project layout

```
web/
├── index.html               # Vite entry; mounts <div id="app">
├── vite.config.ts           # plugins, alias, build chunks, vitest config
├── uno.config.ts            # UnoCSS preset / shortcuts
├── eslint.config.ts         # flat ESLint config
├── tsconfig.{json,app,node}.json
├── env.d.ts                 # vite/client + global ambient types
├── public/                  # static assets copied as-is
├── scripts/                 # i18n + token guardrail node scripts
├── tests/                   # vitest suites (editor/, gallery/, stores/, utils/)
└── src/
    ├── main.ts              # app bootstrap (Pinia, router, i18n, UnoCSS)
    ├── App.vue              # root layout shell
    ├── router/              # vue-router config + guards
    ├── views/               # route-level pages (home, hub, panel/*, auth, init…)
    ├── layout/              # PanelCard, MetricCard, shell layouts
    ├── components/
    │   ├── common/          # BaseDialog, BaseInput, BaseSelect, BaseSwitch, …
    │   ├── advanced/        # TheComment, TheRecentCard, TheEchoCard, …
    │   └── icons/           # SVG icon components
    ├── composables/         # useBaseDialog, useSeoHead, useBfCacheRestore, …
    ├── stores/              # Pinia stores (auth, user, echo, editor, setting, theme, …)
    ├── service/
    │   ├── request/         # ofetch wrapper + auth/refresh/locale interceptors
    │   └── api/             # typed API endpoint modules (auth, user, echo, file, …)
    ├── editor/              # Vditor authoring + custom markdown-it renderer
    ├── locales/
    │   ├── index.ts         # vue-i18n setup, lazy-load + persistence
    │   └── messages/        # zh-CN.json (canonical), en-US.json, de-DE.json
    ├── themes/              # SCSS tokens + light/dark/sunny variants
    ├── lib/                 # third-party adapters (S3 providers, …)
    ├── enums/               # shared TS enums (provider kinds, layouts, modes)
    ├── typings/             # global ambient declarations (App.Api.* DTOs)
    ├── utils/               # storage, toast, time, image, asset loader, …
    ├── plugins/             # custom Vite plugins (welcome banner)
    ├── constants/           # shared constants
    ├── assets/              # imported assets (videos, images)
    └── scripts/             # build-time helpers
```

### Routing ([`src/router/index.ts`](src/router/index.ts))

Public routes:

| Path              | View                | Notes                                               |
| ----------------- | ------------------- | --------------------------------------------------- |
| `/`               | Home / timeline     | `meta.optionalAuth` — public read, gated publish    |
| `/publish`        | Redirect            | → `/?tab=publish`                                   |
| `/hub`            | Discovery hub       | Curated cross-instance feed                         |
| `/echo/:echoId`   | Echo detail         | Shareable single-post page                          |
| `/auth`           | Login / signup      | Blocked when already authenticated                  |
| `/init`           | First-run setup     | Redirects to `/auth` once initialized               |
| `/widget`         | Embeddable widget   | `noindex`                                           |
| `/404`            | Not-found fallback  |                                                     |

Admin routes mount under `/panel/*` with `meta.requiresAuth`: dashboard, settings, user
management, storage, data export/import, SSO, extensions, comment moderation, system logs,
advanced tools.

The global `beforeEach` guard checks system-init state, hydrates the user store, and enforces
auth/optional-auth requirements.

### State (Pinia, [`src/stores/`](src/stores/))

Notable stores:

- **`auth`** — JWT in memory; silent refresh against an HttpOnly cookie; logout / blacklist support.
- **`user`** — current profile, login / signup, passkey (WebAuthn) flows.
- **`editor`** — Vditor instance, preview state, embeddable extensions (music / video / GitHub /
  website), gallery layouts (waterfall / grid / horizontal / carousel / stack).
- **`echo`** — feed pagination, single-echo cache.
- **`setting`** — system settings (title, defaults), agent / LLM provider config, S3 / OAuth2 /
  email settings.
- **`theme`** — light / dark / sunny variants, applies CSS variables on switch.
- **`hub`**, **`init`**, **`connect`** (WebSocket).

### Service layer ([`src/service/`](src/service/))

[`request/index.ts`](src/service/request/index.ts) wraps `ofetch` with:

- automatic `Authorization: Bearer …`, `X-Timezone`, and `X-Locale` headers,
- silent refresh on `TOKEN_MISSING` / `TOKEN_INVALID` / `TOKEN_REVOKED`,
- error-code → i18n translation + toast surfacing,
- three call shapes: `request()` (default), `requestWithDirectUrl()` (external, no creds),
  `downloadFile()` (blob).

[`shared.ts`](src/service/request/shared.ts) resolves the API base, builds avatar URLs, and
constructs WebSocket URLs (auto HTTP→WS, HTTPS→WSS).

[`api/`](src/service/api/) hosts one module per backend domain (`auth`, `user`, `echo`, `file`,
`comment`, `setting`, `agent`, `dashboard`, …). All responses are typed against
`App.Api.Response<T>` declared in [`typings/app.d.ts`](src/typings/app.d.ts).

### Internationalization

`vue-i18n` is configured in composition mode at [`src/locales/index.ts`](src/locales/index.ts).
Locale messages are lazy-loaded; user choice is persisted in `localStorage` and reflected on
`<html lang>`. The canonical key tree is `zh-CN.json`; `en-US.json` and `de-DE.json` must mirror
it exactly — `pnpm i18n:check` enforces this. New UI strings **must** go through `t()`; the
hardcoded-string check will fail the PR otherwise. See
[`docs/dev/i18n-contract.md`](../docs/dev/i18n-contract.md) for the backend/frontend contract
(error envelope, `message_key` / `message_params`).

### Editor & rendering

[`src/editor/`](src/editor/) exposes two components:

- **`MarkdownEditor`** — Vditor-backed authoring surface, image / video upload via Uppy,
  embeddable extensions (music players, video, GitHub project cards, website previews).
- **`MarkdownRenderer`** — `markdown-it` pipeline with `highlight.js`, task lists, XSS-safe
  HTML escaping, and code-block collapsing past ~18 lines (with expand toggle). Output is
  cache-keyed by content + options to avoid stale re-renders on locale switch.

Both surfaces share the gallery layouts handled by `vue-virtual-scroller` + `photoswipe`.

### Styling

[`uno.config.ts`](uno.config.ts) registers `presetWind4`. Project-specific tokens, theme
variants (`base`, `light`, `dark`, `sunny`), and utilities live under
[`src/themes/`](src/themes/) (see [`TOKEN_GUIDE.md`](src/themes/TOKEN_GUIDE.md)). Themes apply
via CSS custom properties so switching is GPU-cheap.

### Testing

`vitest` config is inline in [`vite.config.ts`](vite.config.ts) (jsdom env, globals, shared
[`tests/setup.ts`](tests/setup.ts) that mocks `localStorage` / `sessionStorage` with a `Map`).

Coverage areas under [`tests/`](tests/):

- **`editor/`** — markdown renderer XSS escaping, code-block collapsing, task lists, cache reuse.
- **`gallery/`** — PhotoSwipe integration and image lifecycles.
- **`stores/`** — Pinia store initialization (e.g. `setting`).
- **`utils/`** — external asset loader (MathJax / APlayer / Meting), perf checks.

---

## Build output & embed model

`pnpm build` writes to **[`../template/dist`](../template/dist)** (not `web/dist`). The Go
binary embeds that directory and serves the SPA at `/`, so the frontend is just another asset
in the single-binary distribution. The build also:

- emits gzipped siblings for `.js` / `.css` / `.html` / `.svg` over 10 KB
  (`vite-plugin-compression`),
- splits heavy dependencies into dedicated chunks (`uppy`, `floating-vue`, `highlight`,
  `markdown`) for cache friendliness, see [vite.config.ts:50-74](vite.config.ts#L50-L74).

`<meting-js>` and `<cap-widget>` are registered as Vue custom elements
([vite.config.ts:18](vite.config.ts#L18)) so the compiler doesn't warn on them.

---

## Environment variables

Vite picks up files from `web/` matching `.env`, `.env.development`, `.env.production`. Common
vars used by the request layer:

| Var                      | Purpose                                                       |
| ------------------------ | ------------------------------------------------------------- |
| `VITE_APP_TITLE`         | Document title shown in the welcome banner / `<title>`        |
| `VITE_SERVICE_BASE_URL`  | Backend origin (e.g. `http://localhost:6277` in dev, `/` prod)|
| `VITE_PROXY`             | `YES` to route requests through `VITE_PROXY_URL`              |
| `VITE_PROXY_URL`         | Path prefix when proxy mode is enabled (typically `/api`)     |

Anything secret (API keys, tokens) belongs on the backend — never bake it into a `VITE_*` var,
since those ship to the browser.

---

## Pre-PR checklist

From the repo root:

```bash
just check        # backend lint + openapi drift + frontend lint + i18n checks
```

Inside `web/`, that boils down to:

```bash
pnpm lint
pnpm format
pnpm i18n:check
pnpm test:unit
pnpm build        # type-check + emit
```

See [`CONTRIBUTING.md`](../CONTRIBUTING.md) for the full workflow.

---

## Vite 8 peer warnings (DevTools)

`pnpm install` may print peer warnings under
`vite-plugin-vue-devtools → vite-plugin-inspect → vite-dev-rpc / vite-hot-client` because
upstream peer ranges haven't caught up to Vite 8 yet.

Current stance:

- Keep `vite-plugin-vue-devtools` enabled for development.
- Treat the warnings as non-blocking as long as `pnpm dev`, `pnpm build`, and `pnpm test:unit`
  all succeed.
- Re-evaluate once upstream stable releases widen their Vite 8 peer ranges.
