# Development Guide

This guide covers local development for **Ech0** — environment setup, hot reload, and front-/back-end integration. For higher-level architecture see [`CLAUDE.md`](../../CLAUDE.md) and [`CONTRIBUTING.md`](../../CONTRIBUTING.md).

## Backend Requirements

📌 **Go 1.27.0+**

📌 **C Compiler**
When using CGO-dependent libraries such as `go-sqlite3`, install:
- Windows:
    - [MinGW-w64](https://winlibs.com/)
    - Add the `bin` directory to `PATH` after extraction
- macOS: `brew install gcc`
- Linux: `sudo apt install build-essential`

📌 **Google Wire**
Install [wire](https://github.com/google/wire) for dependency injection file generation:
- `go install github.com/google/wire/cmd/wire@latest`

📌 **Golangci-Lint**
Install [Golangci-Lint](https://golangci-lint.run/) for linting and formatting:
- Run `golangci-lint run` in the project root for linting
- Run `golangci-lint fmt` in the project root for formatting

📌 **Air (Optional, Backend Hot Reload)**
- Recommended via justfile: `just air-install`
- Or install manually: `go install github.com/air-verse/air@latest`

📌 **Swagger**
Install [Swagger](https://github.com/swaggo/gin-swagger) to generate/use OpenAPI docs:
- Run `swag init -g internal/server/server.go -o internal/swagger` in project root to generate or update Swagger docs
- Visit `http://localhost:6277/swagger/index.html` in your browser to view and use docs

📌 **Event Runtime Parameters (Busen)**
- `ECH0_EVENT_DEFAULT_BUFFER` / `ECH0_EVENT_DEFAULT_OVERFLOW`
- `ECH0_EVENT_SYSTEM_BUFFER`
- `ECH0_EVENT_AGENT_BUFFER` / `ECH0_EVENT_AGENT_PARALLELISM`
- `ECH0_EVENT_WEBHOOK_POOL_WORKERS` / `ECH0_EVENT_WEBHOOK_POOL_QUEUE`

📌 **Agent (Copilot) Parameters**
- `ECH0_AGENT_TIMEOUT_SECONDS` — per-run timeout (seconds) for a single Copilot chat run, covering the whole tool loop; default `120`, `<=0` disables the extra timeout.

📌 **OpenAPI Docs Panel**
- `ECH0_OPENAPI_DOCS_RENDERER` — renderer for the `/api/docs` panel: `stoplight` (default, Huma's built-in Stoplight Elements, loaded from CDN) or `scalar` (self-hosted offline Scalar, asset embedded in the binary — no network needed). Unknown values fall back to `stoplight`.

## Frontend Requirements

📌 **NodeJS v26.0.0+, PNPM v10.30.0+**
> Note: if you need multiple Node.js versions, use [fnm](https://github.com/Schniz/fnm) to manage them.

## Start Backend & Frontend

**Step 1: Backend (in Ech0 root directory)**
```shell
just run # normal backend start (equivalent to go run main.go serve)
just dev # backend hot reload with Air
```
> If dependency injection relationships change, run `wire` first in `ech0/internal/di/` to regenerate `wire_gen.go`.

**Step 2: Frontend (new terminal)**
```shell
cd web # enter frontend directory

pnpm install # run if dependencies are not installed

pnpm dev # start frontend preview
# or run from project root: just web dev
```

**Step 3: After both are running**
- Frontend preview: `http://localhost:5173` (actual port shown in terminal after start)
- Backend preview: `http://localhost:6277` (default backend port is 6277)

> When importing packages in a layered architecture, use standardized alias names:
> - model layer: `xxxModel`
> - util layer: `xxxUtil`
> - handler layer: `xxxHandler`
> - service layer: `xxxService`
> - repository layer: `xxxRepository`

## Pre-PR Checklist

```shell
just check        # backend fmt + lint + openapi, web format + lint + i18n:check
just wire-check   # ensure wire_gen.go is up-to-date
go build ./...
pnpm -C web build
```

See [`CONTRIBUTING.md`](../../CONTRIBUTING.md) for the full PR workflow.

## More Reference Docs

- [`auth-design.md`](./auth-design.md) — auth model
- [`access-token-scope-design.md`](./access-token-scope-design.md) — access token scopes
- [`i18n-contract.md`](./i18n-contract.md) — i18n contract between front-end and back-end
- [`logging.md`](./logging.md) — structured logging conventions
- [`timezone-design.md`](./timezone-design.md) — timezone handling
- [`table-design-standard.md`](./table-design-standard.md) — admin panel table conventions
- [`helm-release-validation.md`](./helm-release-validation.md) — Helm chart release validation
