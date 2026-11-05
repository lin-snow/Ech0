---
title: Developer guide
description: Clone the repo, run the backend and frontend, repository layout, and common commands
---

This page helps you **run Ech0 from source** on your machine, so you can change code, debug the API, or contribute a PR. If all you want is a self-hosted instance, start with [Getting started](/docs/start/getting-started) and [Installation](/docs/start/installation) — there is no need to clone the whole repository.

---

## What's in the repository

The main Ech0 repository typically contains (as of the `main` you cloned):

| Path                | Description                                                                                                     |
| ------------------- | --------------------------------------------------------------------------------------------------------------- |
| Go code in the root | Backend API, business logic, embedded static assets, and so on                                                  |
| `web/`              | User-facing frontend (Vite + modern frontend stack)                                                             |
| `site/`             | **Website and documentation site** (the docs on this page are built here; see `site/README.md` for development) |
| `charts/`           | Helm chart                                                                                                      |
| `docs/`             | In-depth repository docs (usage, migration, etc.), complementary to the `site/docs` site docs                   |

For feature details, environment variables, and architecture notes, the root **README.zh.md** and **`docs/`** are authoritative.

---

## Environment requirements at a glance

### Backend (Go)

- **Go version**: no lower than the version declared in `go.mod` (currently 1.27+).
- **CGO**: builds that include SQLite need a local C compiler (MinGW-w64 on Windows, `brew install gcc` on macOS, `build-essential` on Linux).
- **Wire**: if you change dependency injection, run `wire` in the relevant package to generate `wire_gen.go` (see `internal/di/` and others).
- **Code style**: **golangci-lint** is available (`golangci-lint run`, `golangci-lint fmt`).
- **Hot reload (optional)**: **Air**, via `just air-install` or `go install github.com/air-verse/air@latest`.
- **API docs**: **swag** generates Swagger; start the service locally and open `http://localhost:6277/swagger/index.html`.

Starting up:

```bash
just run    # start the backend
just dev    # hot reload if Air is installed
```

### Frontend (`web/`)

- Install dependencies with **pnpm**; use the Node version agreed by the team (fnm or nvm work well).

```bash
cd web
pnpm install
pnpm dev
```

The dev server port is whatever the terminal prints (Vite usually **5173**), and the backend defaults to **6277**. The frontend reaches the backend through a proxy or environment configuration; see the Vite config inside `web`.

---

## Running both together

1. Start the backend with `just run` (in the repository root).
2. Start the frontend with `cd web && pnpm dev`.
3. Open the frontend URL in a browser and confirm the API points at the local backend.

If you only change the API, call `http://localhost:6277` directly with Swagger or curl; if you only change the website docs, start the docs site under `site/` following that directory's README.

---

## The documentation site (`site/`)

- Doc content lives in **`site/docs/{en,zh}/**/*.md`** (English is the default locale, Chinese lives under `zh/`), and routes and listings are registered by `site/app/docs/registry.ts`.
- Keep one file per locale at the same relative path; English pages are served at `/docs/<slug>` and Chinese at `/zh/docs/<slug>`.
- After adding a doc, add its slug to `DOC_ORDER` in `registry.ts` (and optionally `DOC_HERO_SLUGS`), or it may end up ordered late.
- Local development: `pnpm install` / `pnpm dev` (inside the `site` directory); see `site/README.md` for details.

---

## Further reading

Event bus, logging, MCP, and more are covered in the repository's `docs/` and README; this page only covers "getting a dev environment running" and repository navigation. Before submitting a PR, make sure the project's tests and lint pass (see `justfile` and the CI configuration).
