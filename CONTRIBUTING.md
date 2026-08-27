# Contributing to Ech0

Thank you for your interest in contributing to Ech0.

To keep collaboration smooth, please read this document and follow the conventions below.

## Communication and collaboration

- **Bug reports:** use GitHub Issues.
- **Feature discussion:** prefer GitHub Discussions.
- **Security vulnerabilities:** do not file public issues; follow the private disclosure process in `SECURITY.md`.

## Development environment

### Backend

- Go `1.27.0+`
- A working C toolchain (CGO is used, e.g. for SQLite)

Common commands (repository root):

```bash
just run
just dev
just check     # backend fmt/lint + web format/lint + i18n (mandatory before a PR)
```

### Frontend

- Node.js `26.0.0+`
- pnpm `10+`

Common commands (`web` directory):

```bash
pnpm install
pnpm dev
pnpm build
pnpm lint
```

## Contribution workflow

1. Fork this repository and create a feature branch (e.g. `feat/xxx`, `fix/xxx`).
2. Keep changes focused: one PR should ideally address one kind of change.
3. **Before opening a PR, run `just check` from the repository root** (required; see “Pre-submission checks”).
4. Open a Pull Request with a clear description of context, approach, and verification.

## Pre-submission checks

Before opening a PR:

- **Run `just check` once from the repository root** (backend `golangci-lint` fmt/lint, `web` format/lint, and i18n guardrails). This is **mandatory**; fix any reported issues before you submit.
- Ensure the backend still builds (`go build ./...`).
- Ensure the frontend still builds (`pnpm build` from the `web` directory).
- Add or update tests when behavior changes (when applicable).
- Update documentation when changes affect users or deployment.
- **Regenerate Swagger/OpenAPI** when HTTP routes, request/response shapes, or `swag` annotations change: from the repository root run `swag init -g internal/server/server.go -o internal/swagger`, then commit the updated files under `internal/swagger/`.

## Pull Request guidelines

Your PR description should include:

- **Purpose** (why the change is needed).
- **What changed** (main edits).
- **How you verified** (how to confirm it works).
- **Impact** (compatibility, migration, rollback notes if relevant).

For large changes, consider splitting into smaller, reviewable PRs.

## Releasing

Maintainers cutting a new release should follow the documented procedure in [`docs/dev/release-process.md`](docs/dev/release-process.md). User-visible changes per release are tracked in [`CHANGELOG.md`](CHANGELOG.md); add an entry under `[Unreleased]` whenever your PR introduces a change a self-hoster needs to know about.

## Code style

- Match existing project style and layout; avoid introducing patterns that conflict with current conventions.
- Avoid unrelated refactors and large-scale formatting-only changes.
- Follow existing naming conventions (e.g. layered package aliases).

## License

By contributing code to Ech0, you agree that your contributions are licensed under the project’s current open-source license (AGPL-3.0).
