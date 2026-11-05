# Ech0 — official site & documentation

This package is the **official** landing page and **documentation site** for [Ech0](https://github.com/lin-snow/Ech0). It is a **client-side SPA** (no SSR): React Router 7, React 19, Vite, Tailwind CSS v4, and Markdown-backed docs.

## What lives here

Documentation is bilingual: English is the default locale and Simplified Chinese is served under `/zh`. The landing page is English-only.

| Path          | Purpose                                                                                               |
| ------------- | ----------------------------------------------------------------------------------------------------- |
| `app/routes/` | Pages: home (`/`), docs catalog (`/docs`, `/zh/docs`), doc article (`/docs/`_, `/zh/docs/`_), privacy |
| `app/docs/`   | Locale vocabulary (`locale.ts`), doc registry (`registry.ts`), Markdown rendering, table of contents  |
| `docs/`       | Markdown sources per locale (`en/**/*.md` default, `zh/**/*.md`), images under `docs/imgs/`           |
| `public/`     | Static assets (`logo.svg`, `screenshot.png`, `_redirects` for SPA fallback on static hosts)           |

The doc catalog order and featured cards are controlled in `app/docs/registry.ts` (`DOC_ORDER`, optional `DOC_HERO_SLUGS`).

## Prerequisites

- **Node.js** (LTS recommended)
- **pnpm** (workspace uses `pnpm` at the repo root)

## Commands

```bash
pnpm install
pnpm dev          # dev server (Vite; default http://localhost:5173)
pnpm build        # output: build/client/ (static assets)
pnpm start        # serve build/client locally (serve -s)
pnpm typecheck    # react-router typegen + tsc
pnpm lint
pnpm format       # or pnpm format:check
```

`prebuild` copies `docs/imgs` → `public/docs-assets` when present, so Markdown can reference `/docs-assets/imgs/...`.

## Environment

| Variable        | Purpose                                                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `VITE_SITE_URL` | Canonical site origin (no trailing slash) for OG URLs, JSON-LD, sitemap-related logic. Default in code: `https://www.ech0.app`. |

When deploying to a custom domain, align `VITE_SITE_URL` and update `public/sitemap.xml` / `public/robots.txt` if needed.

## Production build

- Run `pnpm build`.
- Deploy the contents of `**build/client/**` to any static host.
- Configure **SPA fallback** to `index.html` for client-side routes (this repo includes `public/_redirects` for Netlify-style hosts).

## Editing documentation

1. Add or change Markdown under a locale directory: `docs/en/<slug>.md` (default) and `docs/zh/<slug>.md` (Chinese). A missing translation falls back to English.
2. English pages are served at `/docs/<slug>`, Chinese at `/zh/docs/<slug>`; the language switch keeps the current slug.
3. Register the slug in `app/docs/registry.ts` `DOC_ORDER` if you care about sidebar order; unlisted files still appear but sort after known entries.
4. Use `![](imgs/...)` in Markdown; paths are rewritten to `/docs-assets/imgs/...` at render time.
5. Link between docs with absolute paths in the current locale (`/docs/...` in English, `/zh/docs/...` in Chinese).

## License

Content and code follow the same terms as the parent [Ech0](https://github.com/lin-snow/Ech0) repository unless noted otherwise.
