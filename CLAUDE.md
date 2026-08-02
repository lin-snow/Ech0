# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Ech0 is a self-hosted personal microblog (timeline) platform. It is shipped as a single Go binary that serves both the REST API and the built SPA. Backend is Go 1.26+ (Gin + Wire DI + GORM + SQLite via CGO), frontend is Vue 3 + Vite + TypeScript + UnoCSS under `web/`. Two satellite web projects also live in-repo: `hub/` (Vue 3 public-directory site) and `site/` (React Router marketing/docs site) — they are independent of the Go binary.

**For a full architecture walkthrough, read `docs/dev/architecture-overview.md` first** — it covers the layered backend, business domains, Agent/MCP capability layers, event subsystem, infra modules, and `pkg/` libraries end-to-end.

## Common commands

All backend commands run from the repository root; frontend commands run from `web/` (or via the `make web-*` wrappers).

```bash
# Backend
make run             # go run ./cmd/ech0 serve (blocks on :6277)
make dev             # Air hot-reload (auto-installs Air via `make air-install` if missing)
make test            # go test ./...
make lint            # golangci-lint run
make fmt             # golangci-lint fmt
make wire            # regenerate internal/di/wire_gen.go (run after changing provider sets / DI graph)
make wire-check      # fails if wire_gen.go is stale vs. wire.go
make openapi         # regenerate OpenAPI spec (Huma type-first) -> internal/openapi/openapi.yaml
make openapi-check   # fails if the committed OpenAPI spec is stale vs. code (mirrors wire-check)

# Frontend (from web/)
pnpm install
pnpm dev             # Vite dev server on :5173, proxies to backend on :6277
pnpm build           # type-check + vite build
pnpm test:unit       # vitest run
pnpm lint            # eslint . --fix
pnpm lint:style      # stylelint --fix (CSS/SCSS/<style> blocks in Vue SFCs)
pnpm format          # prettier --write src/
pnpm i18n:check      # runs key / unused / hardcoded / pseudo-smoke checks (required before PR)

# Full pre-PR verification (mandatory per CONTRIBUTING.md)
make check           # alias of make dev-lint: backend fmt+lint+openapi + web format+lint+i18n:check

# Single Go test
go test ./internal/middleware -run TestAuth     # example
go test -run TestName ./path/to/pkg             # by name
```

Run a single frontend test: `pnpm -C web exec vitest run path/to/file.spec.ts` (or `-t "test name"`).

Binary entrypoint is `cmd/ech0/main.go`. CLI verbs (Cobra): `ech0 serve` (HTTP), bare `ech0`/`ech0 tui` (TUI), `ech0 version`, `ech0 hello`. Snapshot import/export is web-only (admin panel "数据管理"); there is no snapshot CLI verb.

## Architecture

### Layered backend with Wire DI

Backend follows a strict layered architecture — **handler → service → repository → database** — with Google Wire generating the dependency graph. Each business domain (echo, comment, file, connect, user, auth, init, setting, embedding, copilot, dashboard, migrator, common) has parallel packages under `internal/handler/<x>`, `internal/service/<x>`, `internal/repository/<x>`, and `internal/model/<x>` (plus handler-only `web` and `mcp`). Note: `internal/agent` is **not** a layered domain — it is the LLM core package consumed by the copilot service.

- `internal/di/wire.go` declares provider sets (`InfraSet`, `DomainSet`, `HandlerSet`, `EventSet`, `TaskerSet`, `MiddlewareSet`, `StorageSet`, `VisitorSet`, `RuntimeSet`, `AppSet`) and the `BuildApp` injector that composes the full runtime. **If you add/remove a constructor or change a binding, run `make wire`** before committing.
- **Three stateful singletons must be injected once at the top level** (`BuildApp`/`BuildServer`) and shared down, or Wire silently builds a second copy and breaks things: `visitor.Tracker` (VisitorSet), `storage.Manager` (StorageSet), `job.Manager` (BuildJobManager). The `wire.go` comments explain each failure mode.
- Cross-domain aliases are required when importing layers: `xxxHandler`, `xxxService`, `xxxRepository`, `xxxModel`, `xxxUtil` (enforced by existing code; see README "Start Backend & Frontend" note).
- `internal/app` is a generic component lifecycle orchestrator. `internal/server` is the thin Gin/HTTP `Component` it manages. The other managed components are `job.Manager` and `task.Manager` (started in order job → task → server); the `EventRegistrar` registers/drains subscriptions and `setting.Seed` runs via `BeforeStart`/`AfterStop` hooks.
- `internal/bootstrap/bootstrap.go` runs before Cobra dispatches: loads config, initializes the zap-based logger, sets host env defaults. Config is accessed via `config.Config()` (singleton).
- HTTP routes use **Huma (type-first OpenAPI)** on top of Gin (`humagin` adapter). JSON handlers are **framework-neutral**: `func(ctx, *XxxInput) (commonModel.Result[T], error)` that build the success envelope with `commonModel.OK(data[, msg])` and return the raw service error — they don't import `humares`/Huma. Each handler package declares its `XxxInput` (path/query/header/body) and `XxxOutput` (= `commonModel.Result[T]`) types side by side. The single adapter `humares.Wrap` localizes the result + maps errors (`humares.Err`) and wraps it into Huma's output `Envelope[T]`; endpoints register via `route(api, posture, op, handler)` in per-domain `registerXxx` funcs, aggregated by `registerOperations` (`internal/router/huma.go`). Per-endpoint auth is one **posture** — `public()` / `optional(revoker)` / `secured(revoker, scopes...)[.audience(...)]` — that emits **both** the OpenAPI Security declaration and the runtime middleware chain (existing gin middleware reused via `humares.Bridge`), so authn/authz can't drift. The shared response contract (envelope `commonModel.Result[T]`, i18n error localization, security declarations) lives in `internal/handler/humares`. The OpenAPI spec is generated from Go types (no annotations) — runtime docs at `/api/docs`, spec at `/api/openapi.json|.yaml`, committed copy at `internal/openapi/openapi.yaml`. Non-JSON endpoints (SSE/WebSocket streams, multipart upload, binary download, OAuth redirects, captcha, cookie/token-issuance auth flows, MCP JSON-RPC) stay on raw gin (the parallel `handler/response` contract: `res.Execute` → same `commonModel.Result` wire shape) in the `setupXxxRoutes` functions; `SetupRouter` (`internal/router/router.go`) wires the app explicitly.

### Event bus (Busen)

Ech0 uses the in-repo **Busen** library (vendored at `pkg/busen`, imported as `github.com/lin-snow/ech0/pkg/busen`) as an async in-process event bus. The event system follows one rule: dependencies point inward to a **pure vocabulary** package. `internal/event` (package `event`) holds the event structs + their self-describing methods (`EventName()`, `OrderingKey()`) + `WebhookObservation` — it imports only domain models, never busen/services. `internal/event/bus` (alias `eventbus`) is the infrastructure: the `*busen.Bus` singleton, the generic `Emit[T]` (fire) / `Notify[T]` (best-effort fire + warn-log) / `On[T]` (type-routed subscribe) helpers, subscribe-option presets, and the `EventRegistrar`. The registrar wires all subscriptions on `BeforeStart` and tears them down on `AfterStop`, then drains any subscriber implementing `Draining` (e.g. the webhook dispatcher's worker pool) — there is no separate bus-drain component, and the bus's async queues are best-effort (dropped on shutdown). Subscribers live at `internal/event/subscriber` (agent processor, embedding processor, snapshot scheduler) and self-register via `bus.On`; the webhook dispatcher (`internal/webhook`) is itself a subscriber too — it bridges each observable event to a neutral `WebhookObservation` via `bus.OnWithMeta` (the metadata-aware variant of `On`). **Routing is by Go type** (no topic dimension); events self-describe their stable webhook name via `EventName()`. Producers publish with `eventbus.Notify(ctx, bus, event.EchoCreated{...})` (best-effort; `Emit` when the caller wants the error) — there is no publisher facade. The bus decouples comment/echo/user events from side effects like webhooks, agent runs, and snapshots. Runtime tuning is via `ECH0_EVENT_*` env vars (buffers, parallelism, webhook worker pool) — see README "Event Runtime Parameters".

Webhook dispatch (`internal/webhook`) and the cache/index/snapshot processors (`internal/event/subscriber`) are implemented as event subscribers, not inline handler calls. When adding cross-cutting side effects, prefer publishing an event over invoking services directly from handlers. (Caution: `internal/agent` is **not** a subscriber — it is the synchronous LLM core; the bus-facing `AgentProcessor` subscriber only invalidates the AI-summary cache.)

### Agent & MCP capability layers (LLM, mirror-image directions)

Two function-calling integrations that point opposite ways:

- **`internal/agent` (outbound)** — Ech0's LLM core. Collapses OpenAI-compatible + Anthropic protocols behind one `Provider` abstraction (`Complete`/`Stream`) plus a ReAct tool loop (`agent.Run`). **Zero domain deps**: tools are injected by the caller as `Tool{Def, Execute}` closures, and i18n strings are passed in (`RunStrings`). The copilot service (`internal/service/copilot`) injects `search_echos`/`summarize_echos`/`stats_overview`, then translates the loop's `AgentEvent` stream into Chat SSE (`searching|sources|delta|done|error`). `Generate` is the non-streaming entry used for summaries. SDK quirks (streaming tool-call fragment reassembly) are sealed inside each Provider; the loop only sees clean semantic events.
- **`internal/mcp` (inbound)** — a JSON-RPC 2.0 MCP server mounted at `/mcp` (RequireAuth) that exposes domain APIs to *external* LLMs as tools/resources. Each tool/resource declares a required scope at registration; `Server.dispatch` enforces it against the caller's access-token scopes **before** invoking the domain service via `Adapter`. Keep authorization centralized in dispatch, not scattered into business code.

### Storage (VireFS)

`internal/storage` is a unified abstraction over local disk and S3-compatible object stores, layered over the in-repo **VireFS** library (vendored at `pkg/virefs`, imported as `github.com/lin-snow/ech0/pkg/virefs`). Files are addressed by a flat `key`; `schema.Resolve` + `PathPrefix` map keys to on-disk paths or S3 object keys. Stored `File.url` is a snapshot of the UI-visible URL at write time. The `/api/files` static route serves local content; the `stream` routes are authenticated. `S3SettingStore` is bound to `KeyValueRepository` so S3 config lives in the settings DB, not env. Switching providers / migrating between local and S3 is documented at `docs/usage/storage-migration.md`.

### Frontend

- Vue 3 SFCs in `web/src`, Pinia stores, Vue Router, i18n via `vue-i18n`, UnoCSS (Wind4 preset), markdown via `markdown-it` + Vditor editor.
- i18n guardrails in `web/scripts/` (key completeness, unused keys, hardcoded strings, pseudo-locale smoke) are part of `make check` — **do not introduce hardcoded UI strings**; use translation keys.
- Vite serves `:5173` during dev and proxies `/api` to the backend on `:6277`. `pnpm build` outputs to `template/dist/`, which the Go binary embeds via `//go:embed all:dist` (`template/template.go`) and serves through `internal/handler/web` in production.

### Configuration

`internal/config/config.go` is the single config source; env vars are parsed via `caarlos0/env`. See `.env.example` for the full set (JWT secret, server port, DB path, log, S3, event runtime, etc.). Defaults target `./data/` for SQLite + uploaded files; Docker images mount `/app/data`.

## Conventions to respect

- **Before a PR**: `make check` is mandatory (enforces backend lint, frontend lint, i18n checks). `go build ./...` and `pnpm build` must pass. Regenerate the OpenAPI spec (`make openapi`) whenever routes or request/response shapes change and commit `internal/openapi/openapi.yaml`; CI-style `make openapi-check` fails on drift.
- **DI changes**: regenerate with `make wire`; CI runs `make wire-check`.
- **SPDX headers**: every `.go` / `.ts` / `.vue` file needs an SPDX license header. `make spdx` adds missing ones; CI enforces via `make spdx-check`.
- **Migrator (data portability)**: the admin panel's "数据管理" page wraps a bidirectional Migrator domain. **Import** supports Ech0 snapshot → Ech0 and Memos → Ech0; **Export** produces a unified **Snapshot** (a zip of `data/`, see `internal/migrator/snapshot`) that round-trips back through the `ech0` import. Core engine (importer/exporter execution, ETL, snapshot resource) lives in `internal/migrator`; `internal/service/migrator` is a thin layer doing auth + job lifecycle + DTO + upload orchestration. Export triggers: manual snapshot (`POST /migration/export`, async via `job.Manager`, `TypeExport`), scheduled snapshot (`internal/task/scheduled` cron, syncs through the exporter), and synchronous download (`GET /migration/export/download`). There is no separate "backup" concept — it is all snapshot export.
- **Integration comment endpoint**: `POST /api/comments/integration` intentionally bypasses captcha/form-token — it requires an access token with `comment:write` scope and `integration` audience. Preserve this behavior.
- **Access tokens**: scope/audience/`typ` design is documented at `docs/dev/access-token-scope-design.md`; implementation is authoritative.
- **Layered import aliases** (required): `xxxHandler`, `xxxService`, `xxxRepository`, `xxxModel`, `xxxUtil`.
- **Logging**: use the project zap wrapper at `internal/util/log` with a `module` field; see `docs/dev/logging.md` for field conventions.

## Useful reference docs in-repo

- `docs/dev/architecture-overview.md` — **start here**: full architecture panorama (layers, domains, Agent/MCP, events, infra, `pkg/`)
- `docs/dev/llm-chat-design.md`, `docs/dev/agent-toolcall-design.md` — Chat RAG + Agent Provider/ReAct design
- `docs/dev/job-runner-design.md`, `docs/dev/snapshot-design.md` — async job framework & snapshot export
- `docs/dev/auth-design.md`, `docs/dev/access-token-scope-design.md` — auth model & token scopes
- `docs/dev/i18n-contract.md` — frontend/backend i18n contract (locale header, error field shapes, key naming)
- `docs/dev/logging.md`, `docs/dev/timezone-design.md`, `docs/dev/table-design-standard.md` — logging fields, TZ handling, admin table conventions
- `docs/usage/storage-migration.md`, `docs/usage/mcp-usage.md`, `docs/usage/webhook-usage.md`, `docs/usage/capsule.md` — operator/integration guides
- `docs/dev/capsule/spec.md` — Capsule interchange format + CLI (normative); `docs/dev/capsule/capsule-design.md` — why it is shaped that way
- `CONTRIBUTING.md` — PR workflow and pre-submission checks; `docs/README.md` — doc index
