<!--
Thanks for contributing to Ech0!
Please keep this PR focused on a single change. For larger work, split into smaller reviewable PRs.
-->

## Summary

<!-- Why is this change needed? What does it do? Link related issues with "Closes #123" / "Refs #123". -->

## Type of change

<!-- Tick all that apply. -->

- [ ] Bug fix
- [ ] Feature / enhancement
- [ ] Refactor (no behavior change)
- [ ] Docs
- [ ] Build / CI / chore
- [ ] Breaking change

## Area

<!-- Tick all that apply. -->

- [ ] Backend / API
- [ ] Frontend / UI
- [ ] Auth / Access tokens
- [ ] Storage (local / S3)
- [ ] Webhook / Agent / Events
- [ ] Migration / Backup
- [ ] Hub / Connect
- [ ] Docs

## How to verify

<!-- Steps a reviewer can follow to confirm the change. Include curl/HTTP requests, UI steps, or test commands. -->

## Impact

<!-- Compatibility, migration, rollback, config / env var changes, breaking notes. Write "None" if not applicable. -->

## Pre-submission checklist

- [ ] `just check` passes locally.
- [ ] `go build ./...` passes.
- [ ] `pnpm build` passes (when frontend is touched).
- [ ] Tests added / updated when behavior changed.
- [ ] `just wire` re-run when DI providers/bindings changed; `wire_gen.go` committed.
- [ ] `just openapi` re-run when routes or request/response shapes changed; `internal/openapi/openapi.yaml` committed.
- [ ] Docs updated when changes affect users or deployment.
- [ ] No hardcoded UI strings — i18n keys used; `pnpm i18n:check` passes.
- [ ] No secrets, tokens, or personal data in diffs, logs, or screenshots.
