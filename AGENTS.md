# AGENTS.md

Compact guidance for AI agents working in the Ech0 repository.

## What this is

Ech0 is a self-hosted personal microblog platform. Single Go binary serves both the REST API and the embedded Vue SPA. Backend: Go 1.27+ (Gin + Wire DI + GORM + SQLite via CGO). Frontend: Vue 3 + Vite + TypeScript + UnoCSS under `web/`.

**Additional packages:** `hub/` (Vue 3 public directory site), `site/` (React Router marketing/docs site).

## Commands you'll need

The repo has a single task runner: **`just`** (no justfile). The root `justfile` holds
backend/repo-wide recipes and aggregates the sub-projects as `just` modules —
`just web …`, `just site …`, `just hub …`, `just docker …`. Run `just` (or `just --list`)
for the root recipes, `just --list web` for a module's own recipes.

### Backend (run from repo root)

| Purpose | Command |
|---------|---------|
| Run server | `just run` (or `go run ./cmd/ech0 serve`) |
| Hot-reload dev | `just dev` (auto-installs Air if missing) |
| Tests | `just test` (or `go test ./...`) |
| Single test | `go test ./internal/middleware -run TestAuth` |
| Race / coverage | `just test-race` / `just test-cover` |
| Lint | `just lint` (golangci-lint run) |
| Format | `just fmt` (golangci-lint fmt) |
| Regen DI | `just wire` (run after changing provider sets / constructors) |
| Verify DI | `just wire-check` |
| Regen mocks | `just mocks` (verify with `just mocks-check`) |
| Regen OpenAPI | `just openapi` (verify with `just openapi-check`) |

### Frontend (`just web …` from repo root, or pnpm from `web/`)

| Purpose | Command |
|---------|---------|
| Dev server | `just web dev` (Vite on :5173, proxies `/api` to :6277) |
| Build | `just web build` (type-check + vite build → `template/dist/`) |
| Unit tests | `just web test` |
| Single test | `pnpm -C web exec vitest run path/to/file.spec.ts` |
| Lint | `just web lint` (eslint --fix) |
| Style lint | `just web lint-style` (stylelint --fix) |
| Format | `just web format` (prettier --write src/) |
| i18n check | `just web i18n-check` (key completeness + unused + hardcoded + pseudo-smoke) |

### Other modules

| Purpose | Command |
|---------|---------|
| Docs/marketing site | `just site dev` / `just site build` / `just site typecheck` |
| Public directory site | `just hub dev` / `just hub build` |
| Container image | `just docker build` / `just docker push` (override `OS`/`ARCH`/`IMAGE_TAG`) |

### Pre-PR (mandatory)

```bash
just check          # SPDX + backend fmt/lint/openapi + web format/lint/style/i18n
just wire-check     # verify wire_gen.go is current
go build ./...      # backend compiles
just web build      # frontend compiles
```

`just check` runs all steps even if one fails, prints a summary table at the end.

## Architecture quick reference

### Backend layers (strict order)

handler → service → repository → database

Each domain (echo, user, auth, comment, connect, file, setting, dashboard, agent, backup, migration, init, common) has parallel packages under `internal/{handler,service,repository,model}/<domain>/`.

### Import aliases (required)

Cross-domain imports must use: `xxxHandler`, `xxxService`, `xxxRepository`, `xxxModel`, `xxxUtil`.

### Wire DI

`internal/di/wire.go` declares provider sets and the `BuildApp` injector. **Any constructor add/remove/binding change requires `just wire`.** CI runs `just wire-check` and will fail if `wire_gen.go` is stale.

### Event bus (Busen)

In-process async event bus. Publishers at `internal/event/publisher`, subscribers at `internal/event/subscriber`, contracts at `internal/event/contracts`. Webhooks and agent processing are event subscribers, not inline handler calls. Prefer publishing events over direct service calls for cross-cutting side effects.

### Frontend build output

`pnpm build` outputs to `../template/dist/` — this gets embedded into the Go binary via `embed`. The backend serves the SPA from there in production.

### Entry points

- CLI: `cmd/ech0/main.go` (Cobra: `ech0 serve`, `ech0 tui`, `ech0 version`, `ech0 hello`)
- HTTP routes: `internal/router/*.go`, wired in `internal/server/provider.go`
- Config singleton: `config.Config()` via `internal/config/config.go` (env vars parsed with `caarlos0/env`)
- Bootstrap: `internal/bootstrap/bootstrap.go` runs before Cobra dispatches

## Conventions that differ from defaults

- **No hardcoded UI strings** — use vue-i18n translation keys. i18n guardrails (`just web i18n-check`) are part of `just check`.
- **SPDX headers required** on all `.go`, `.ts`, `.vue` files. Run `just spdx` to add them; CI checks with `just spdx-check`.
- **OpenAPI spec must be committed** — regenerate with `just openapi` when routes/request/response shapes change. The spec lives at `internal/openapi/openapi.yaml`; `just openapi-check` fails on drift.
- **Logging** — use the zap wrapper at `internal/util/log` with a `module` field. See `docs/dev/logging.md`.
- **Comment integration endpoint** — `POST /api/comments/integration` intentionally bypasses captcha; requires access token with `comment:write` scope and `integration` audience. Preserve this.

## Testing notes

- Go tests: `go test ./...` (CGO required for SQLite)
- Frontend tests: vitest, jsdom environment, setup at `web/tests/setup.ts`, test files in `web/tests/`
- Frontend test include pattern: `tests/**/*.{test,spec}.ts`

## Environment & config

- Env vars parsed from `.env` (loaded via `joho/godotenv`). See `.env.example` for the full set.
- Defaults target `./data/` for SQLite + uploads. Docker images mount `/app/data`.
- Server port: 6277 (default).
- Node.js 26.0.0+, pnpm 10+, Go 1.27.0+, C toolchain for CGO.

## Key in-repo docs

| Topic | File |
|-------|------|
| Auth model | `docs/dev/auth-design.md` |
| Token scopes | `docs/dev/access-token-scope-design.md` |
| i18n contract | `docs/dev/i18n-contract.md` |
| Logging | `docs/dev/logging.md` |
| Timezone handling | `docs/dev/timezone-design.md` |
| Release process | `docs/dev/release-process.md` |
| Storage migration | `docs/usage/storage-migration.md` |
| MCP usage | `docs/usage/mcp-usage.md` |
| Webhook usage | `docs/usage/webhook-usage.md` |
| Capsule usage | `docs/usage/capsule.md` |
| Capsule format spec | `docs/dev/capsule/spec.md` (design rationale: `docs/dev/capsule/capsule-design.md`) |
| Contributing | `CONTRIBUTING.md` |
| Dev setup | `docs/dev/development.md` |

## Also in this repo

- `CLAUDE.md` — more detailed architecture notes (read if you need deeper context)
- `justfile` — the repo's only task runner (root recipes + `web`/`site`/`hub`/`docker` modules)
- `CHANGELOG.md` — user-visible changes per release (add entries under `[Unreleased]`)
