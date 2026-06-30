# Changelog

All notable user-visible changes to Ech0 are recorded here.

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html), and this file follows the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

For releases prior to v4.6.5, see the [GitHub releases page](https://github.com/lin-snow/Ech0/releases) — earlier release notes are not retroactively imported here.


## [5.3.0] - 2026-06-30

This release is anchored by two large engineering efforts — a **type-first OpenAPI rebuild (Huma)** and the project's **first real backend test suite** — plus the security and stability fixes that writing those tests surfaced.

### Added

- **YouTube Shorts and live links now embed.** The video extension's YouTube matcher previously only recognized `watch?v=`, `youtu.be`, and `embed/` URLs, so `youtube.com/shorts/<id>` and `/live/<id>` fell through and were rejected. Both are now extracted and rendered through the existing `/embed/<id>` iframe — no renderer changes needed.
- **Selectable API-docs renderer.** A new `ECH0_OPENAPI_DOCS_RENDERER` env var chooses the panel served at `/api/docs`: `stoplight` (default, Huma's built-in Stoplight Elements) or `scalar` (a fully offline, self-hosted Scalar API Reference bundled into the binary — no CDN fetch at runtime). See `docs/dev/development.md`.

### Changed

- **API documentation rebuilt on Huma type-first OpenAPI.** The HTTP layer migrated off swaggo annotations to **Huma** on top of Gin: the OpenAPI spec is now generated from Go request/response types (no annotation comments), with interactive docs at `/api/docs` and the machine-readable spec at `/api/openapi.json` and `/api/openapi.yaml` (committed copy at `internal/openapi/openapi.yaml`, kept honest by `make openapi` / `make openapi-check`). JSON handlers became framework-neutral, and each endpoint's authentication is declared once as a single "posture" (`public` / `optional` / `secured`) that emits both the OpenAPI security declaration and the runtime middleware chain, so authn/authz can no longer drift from the docs. Retiring swaggo removed ~10,000 lines of generated swagger from the tree. Non-JSON endpoints (SSE/WebSocket, uploads, downloads, OAuth, captcha, MCP) keep their existing wire shape.
- **Auto theme-mode icon changed from a leaf to a palm tree.** The home header's "follow system" theme toggle now uses a Lucide `tree-palm` glyph instead of the old Tabler leaf.

### Security

- **Random-string generation is now cryptographically secure.** `util/crypto.GenerateRandomString` switched from `math/rand` (a time-seeded, predictable PRNG) to `crypto/rand` with rejection sampling to avoid modulo bias. This function backs **OAuth `state`, one-time OAuth exchange codes, and token JTIs** — values that must be unpredictable; the old source could in principle be predicted, enabling CSRF-`state` or exchange-code forgery. A failed secure-random read now panics rather than silently degrading to a guessable value.
- **OAuth token exchange now uses a timeout-bound HTTP client.** The code-for-token exchange was going through `http.DefaultClient` (no timeout); it now uses an explicit timeout-bound client, so a hung or slow identity provider can't pin a request open indefinitely.

### Fixed

- **The async worker pool no longer panics under a shutdown race.** `util/async.Pool.Submit` now holds its read lock through the channel send and drops the `recover`, so `close` happens only under the write lock — structurally eliminating the `send on closed channel` panic when `Submit` and `Stop` race. Covered by a new concurrent `-race` regression test. (This pool backs event/webhook dispatch.)
- **Busen's dispatch gate no longer panics on an unpaired or double `Leave`.** `Gate.Leave` used to `close` an already-closed idle channel (`panic: close of closed channel`) when called more times than `Enter`; it now no-ops when no callers are active and closes the idle channel only on the true `>0 → 0` edge, making any extra/duplicate `Leave` a safe no-op.

### Internal

- **First-class backend testing system.** Introduced **testify + mockery v3** (pinned via `go run`, kept out of `go.mod`), a shared `internal/test/helpers` scaffold (in-memory DB, viewer identities, JWT overrides, envelope parsing, fixtures) and centrally generated mocks under `internal/test/mocks` (10 domains, deterministic + SPDX). New Makefile targets (`mocks`, `mocks-check`, `test-race`, `test-cover`) and `docs/dev/testing.md` codify the conventions; coverage climbed across three rounds to **~66% (calibrated**, excluding generated mocks/`wire_gen.go`). CI (`test.yml`) now runs `go test` automatically on PRs and pushes to `main` (backend-path-filtered, coverage report-only — RAW + CALIBRATED — with no hard gate).
- **Testability seams (all zero-Wire — constructor signatures unchanged).** `ConnectService` gained an injectable peer-fetcher and an injectable retry backoff (`WithRetryBaseDelay`, cutting the connect package's wall-clock test time from ~4s to ~0); `storage` gained a `NewStorageManagerForTest` / `helpers.NewTestStorage` path on an isolated temp dir; auth/embedding gained white-box seams.
- **i18n / error-handling consistency, surfaced by the Huma review.** Framework-level errors are now classified by HTTP status (5xx → `INTERNAL_ERROR` + `common.request_failed` instead of masquerading as "invalid query parameter"; 4xx validation → neutral `common.invalid_request`); the localizer resolves lazily so business responses and validation errors share the post-auth user locale; `humares.Err` and `response.Execute` share one failure-field mapping ladder (`commonModel.ResolveFailureFields`) so the two response contracts can't drift.
- **Dependency bumps (Go, `go-patch-minor` group)**: `anthropics/anthropic-sdk-go` 1.50.1 → 1.52.0, `aws/aws-sdk-go-v2/service/s3` 1.103.3 → 1.104.0, `coreos/go-oidc/v3` 3.18.0 → 3.19.0, `aws/smithy-go` 1.27.2 → 1.27.3, `gorm.io/gorm` 1.31.1 → 1.31.2.
- **Dependency bumps (`web/`)**: `vue` 3.5.38 → 3.5.39, `@types/node` 25.9.4 → 26.0.1, `unocss` / `@unocss/preset-wind4` 66.7.2 → 66.7.3, `eslint` 10.5.0 → 10.6.0, `prettier` 3.8.4 → 3.9.1, `stylelint` 17.13.0 → 17.14.0, `vite` 8.0.16 → 8.1.0, `vite-plugin-vue-devtools` 8.1.3 → 8.1.5, `@vue/eslint-config-typescript` 14.8.0 → 14.9.0.
- **CI**: `actions/checkout` 6 → 7.


## [5.2.5] - 2026-06-21

### Added

- **Verified-user tooltip in comments.** Hovering the blue verified badge next to a commenter's name now explains what it means — the author is a registered user of this instance — instead of leaving guests to guess. Applied to both top-level comments and replies (shown when `source === 'system'`), reusing the existing floating-vue `v-tooltip`. New i18n key `commentSection.verifiedUser` across zh-CN / en-US / ja-JP / de-DE.

### Changed

- **About panel simplified, with a draggable logo sticker.** The About view dropped its dense field list (product subtitle and the version / commit / author / license / build rows) for a cleaner layout and gained a draggable Ech0 logo sticker that springs back to its origin on release; the "view source on GitHub" link (and its at-commit variant) is kept, and the footer line is shortened to "Powered by Ech0". i18n keys trimmed accordingly across zh-CN / en-US / ja-JP / de-DE.

### Fixed

- **OAuth login/binding now works out-of-the-box on single-domain self-hosts.** `parseAndValidateClientRedirect` implicitly allows the SPA's two hardcoded same-origin return paths — `/panel` (account binding) and `/auth` (login) — derived from the configured OAuth2 callback's origin, so a single-domain deployment no longer has to hand-configure the Redirect Allowlist just to bind or sign in via OAuth. These implicit entries are fixed, front-end-hardcoded paths (no arbitrary-path injection) and are still matched with the same RFC 6749 §3.1.2 exact scheme+host+path comparison, preserving the intent of advisory GHSA-p64j-f4x9-wq66; operator-configured allowlist entries continue to apply. Covered by new `oauth_service_test.go` cases.

### Internal

- **Dependency bump (`web/`)**: `@types/node` 25.9.3 → 25.9.4 (dev).


## [5.2.4] - 2026-06-19

### Added

- **Chat now shows the model's reasoning process.** When a model emits a thinking/reasoning stream (or inline `<think>` blocks), the chat panel separates that reasoning from the final answer and renders it in a collapsible section with a live "Thinking…" state and a "Thought for {seconds}s" duration once it settles. The reasoning text and its duration are persisted with the chat session so they survive a reload, and inline `<think>` blocks are stripped out of the answer body. New i18n keys `reasoningThinking` / `reasoningDone` across zh-CN / en-US / ja-JP / de-DE.
- **MCP discovery read tools and a visitor-stats resource.** The inbound MCP server (`/mcp`) now exposes three existing echo read endpoints as tools under `ScopeEchoRead` — `get_hot_posts`, `get_random_post`, `get_on_this_day_posts` — plus a new `ech0://stats/visitors` resource (past 7-day PV/UV) gated by `ScopeAdminSettings` to match the REST `/system/visitor-stats` route. README (`internal/mcp/README.md`) and `docs/usage/mcp-usage.md` updated.

### Internal

- **Removed the unused `vditor` dependency** from `web/` — it had zero references in `web/src` (a leftover after switching to the in-house `TheMdEditor`) and never entered the bundle, narrowing the dependency and supply-chain surface.
- **Hub instance health check.** A new scheduled GitHub Actions workflow (`hub-health-cleanup.yml`) periodically health-checks public-directory instances and prunes dead ones.
- **Toolchain bumps**: `pnpm` 11.8.0; Go 1.26.4 in the docker test-image workflow.
- **Dependency bumps (Go, `go-patch-minor` group)**: `anthropics/anthropic-sdk-go` 1.48.0 → 1.50.1, `aws/aws-sdk-go-v2` 1.41.12 → 1.42.0, `aws-sdk-go-v2/config` 1.32.23 → 1.32.25, `aws-sdk-go-v2/credentials` 1.19.22 → 1.19.24, `aws-sdk-go-v2/service/s3` 1.103.2 → 1.103.3, `aws/smithy-go` 1.27.1 → 1.27.2, `golang.org/x/mod` 0.36.0 → 0.37.0, `golang.org/x/net` 0.55.0 → 0.56.0, `golang.org/x/sync` 0.20.0 → 0.21.0, `golang.org/x/text` 0.37.0 → 0.38.0.
- **Dependency bumps (`web/`, `web-patch-minor` group)**: `vue` 3.5.35 → 3.5.38, `@dicebear/core` 10.1.0 → 10.3.0, `@dicebear/styles` 10.1.0 → 10.2.0, `unocss` / `@unocss/preset-wind4` 66.7.0 → 66.7.2, `eslint` 10.4.1 → 10.5.0, `prettier` 3.8.3 → 3.8.4, `vue-tsc` 3.3.3 → 3.3.5, `vite-plugin-vue-devtools` 8.1.2 → 8.1.3, `npm-run-all2` 9.0.1 → 9.0.2, `@types/node` 25.9.2 → 25.9.3.
- **Dependency bumps (`hub/` & `site/`)**: `markdown-it` 14.1.1 → 14.2.0 (hub), `react-router` 7.15.0 → 7.15.1 (site), `vite` 8.0.12 → 8.0.16 (both).
- **Code structure refactors** for readability/maintainability, and a staticcheck (SA5011) fix in the agent run-loop test.


## [5.2.3] - 2026-06-13

### Changed

- **Music cards now use Ech0's native single-track player.** The Vue player keeps Meting API metadata resolution while replacing APlayer/MetingJS with the browser audio API, Ech0 theme tokens, synchronized lyrics, seeking, and Media Session controls. Playlist responses intentionally play only the first track and no next-track action is exposed. The APlayer/MetingJS scripts and styles are dropped from `web/public`.
- **The Hub timeline re-enables virtual scrolling for music posts.** Because the native music card survives list recycling (APlayer instances did not), the Hub page no longer falls back to a plain non-virtualized list when a music extension is present — `DynamicScroller` is now used unconditionally on the standalone Hub page, restoring smooth scrolling on long timelines that contain music.

### Fixed

- **Tapping an Echo card on touch devices no longer needs two taps.** The card's "open" button used to be revealed by `:hover`, so on a touch device the first tap only triggered the sticky-hover reveal (swallowing the tap) and a second tap was required. The reveal is now gated behind `@media (hover: hover)`, so it only applies to mouse/hover-capable devices; touch taps open the Echo directly.


## [5.2.2] - 2026-06-13

### Added

- **Connection testing for S3 storage and the AI model.** Both the storage settings and the Agent settings panels gained a "test connection" button that validates the live configuration before you save it. Backend: a new `storage.Probe` (`internal/storage/probe.go`) checks bucket existence and credentials against the supplied S3 config **without touching the saved settings**; `agent.Ping` fires one minimal non-streaming request to verify the protocol / BaseURL / ApiKey / Model are actually usable (it deliberately does *not* require `Enable`, so you can test before turning the feature on, and uses `MaxTokens=16` to avoid an Anthropic empty-text false negative at `max_tokens=1`). New i18n keys for the test-connection states across zh-CN / en-US / ja-JP / de-DE.
- **Guest-facing language switcher.** A globe-icon popover in the home header lets visitors switch the UI language; logged-in users' picks persist to `user.locale` for cross-device sync. New i18n keys `homeNav.localeToggleTitle` / `homeNav.localeSyncFailed`.
- **Manual "snapshot now" button with live job progress in the snapshot-schedule settings.** The schedule panel can now trigger a one-off snapshot export and shows the job's progress inline, alongside the existing cron schedule.

### Changed

- **Visitors now see their own browser language by default.** Locale resolution was reordered to *device choice → navigator → site default → fallback*, using a nullable `toSupported()` helper so an unsupported-but-non-empty `navigator` value no longer short-circuits the site-default fallback. Previously a guest on a foreign-language site always saw the owner's `default_locale`.
- **Language options are unified across the app and labelled with endonyms** (native names — 简体中文 / English / 日本語 / Deutsch), sourced from one shared list and reused by the home filter, system settings, and user settings instead of three divergent inline lists.
- **Access-token detail modal rebuilt on Headless UI** for proper focus management, accessibility, and enter/leave transitions.
- **Tag manager popover now positions itself dynamically** relative to the tag button that opened it.

### Fixed

- **The site owner no longer emails themselves a "new comment" notification for comments they post from the admin panel** (`SourceSystem`). Guest comments (`SourceGuest`) and external integration deliveries (`SourceIntegration`) still notify the owner, and replies to a guest still notify the reply target.

### Internal

- **Dependency bumps (Go, `go-patch-minor` group)**: `anthropics/anthropic-sdk-go` 1.46.0 → 1.48.0, `aws/aws-sdk-go-v2` 1.41.9 → 1.41.12, `aws-sdk-go-v2/config` 1.32.20 → 1.32.23, `aws-sdk-go-v2/credentials` 1.19.19 → 1.19.22, `aws-sdk-go-v2/service/s3` 1.102.2 → 1.103.2, `aws/smithy-go` 1.26.0 → 1.27.1.
- **Dependency bumps (`web/`, `web-patch-minor` group)**: `@cap.js/widget` 0.1.54 → 0.1.56, `@dicebear/core` 10.0.2 → 10.1.0, `vue-i18n` 11.4.4 → 11.4.5, `@types/node` 25.9.1 → 25.9.2, `@vue/eslint-config-typescript` 14.7.0 → 14.8.0, `@vue/test-utils` 2.4.10 → 2.4.11, `stylelint` 17.12.0 → 17.13.0.


## [5.2.1] - 2026-06-06

A maintenance release with no user-visible product changes — code formatting, a README community badge (linux.do), and public-directory (`hub/`) entries only.


## [5.2.0] - 2026-06-06

### Added

- **Comment floor numbers & jump-to-parent.** Every comment now shows a floor number (`#N`) assigned in chronological order. A reply displays a clickable "Reply to #N" reference that smooth-scrolls to its parent comment and briefly flashes it for orientation. New i18n key `inReplyToFloor` across zh-CN / en-US / ja-JP / de-DE.
- **Editable publish time when editing an Echo.** `EchoUpsertDto` gained an optional `created_at` field; updating an Echo with a non-zero value rewrites its `created_at` column, so an entry can be backdated or corrected after the fact. Swagger regenerated.
- **Zen-mode monochrome toggle feedback.** Switching the Zen reading view between black-and-white and color now raises a confirmation toast. New i18n keys `bwToastOn` / `bwToastOff`.

### Fixed

- **Object file URLs are now resolved at read time.** A stored `File.url` used to be a snapshot taken at write time, so changing the CDN domain or S3 configuration left old `local`/`object` records pointing at dead links. URLs for these types are now recomputed from the current storage configuration on every read — via the `File` GORM `AfterFind` hook, which covers both direct loads and nested `Preload(EchoFiles.File)` — while `external` URLs keep their stored snapshot. (`a3bdeffa`)
- **Embedding providers that reject the `dimensions` parameter now degrade gracefully.** If the first batch is rejected because `dimensions` is unsupported, the request is retried without it and the conclusion is reused for later batches. Returned vectors whose dimension disagrees with the configured value now raise an error instead of being written, preventing `vec0` dimension conflicts. (`99f8056c`)

### Internal

- **Storage and user config now read through the setting engine.** S3 configuration is read via `setting.Get(setting.S3)` and the user domain reads its settings directly through the setting engine, removing the per-field env merge and decoupling `SettingService`. `setting.Get` now falls back to a normalized default when a stored value cannot be parsed. Wire graph regenerated.
- **New config/settings architecture doc** at `docs/dev/config-and-settings-architecture.md`.
- **UI polish.** `BaseButton` gained a `loading` state (shows a spinner and blocks clicks while busy); the share control is now a native `<button>` for correct semantics/accessibility; removed redundant `v-tooltip`s from the tag/publish buttons and reformatted `ChatBox`.


## [5.1.0] - 2026-06-06

### Added

- **Comment replies (two-level threading).** Comments now carry a `parent_id` field, enabling replies to existing comments in a two-level nested structure. The backend validates that the reply target exists and is approved; the frontend `TheComment` component has been rebuilt with a nested reply UI featuring "replying to @nickname" attribution, an inline reply composer, and a cancel action. New i18n keys (`reply` / `replyingTo` / `cancelReply` / `inReplyTo`) across zh-CN / en-US / ja-JP / de-DE.
- **Reply email notifications.** When a reply is created, the author of the parent comment is notified asynchronously (respecting the existing comment email-notification toggle). Self-replies, invalid recipient emails, and replies where the recipient is the site owner (already covered by the "new comment" notification) are silently skipped. A new `reply` mail template type is added (blue "Reply" badge).
- **Chat question navigator (right-side ToC).** A pill-shaped navigation rail on the right edge of the chat panel. Hover to reveal the full question text; click to scroll-jump to that question with active highlighting. Shows the most recent 7 questions; auto-highlights the latest one during live streaming.
- **Chat retry on failure / empty response.** When a streaming turn is interrupted or the model produces no text, the last turn shows a "No response this time" hint with a retry button for in-place resend — no need to retype the question.
- **`X-Powered-By: Ech0/<version>` response header.** A new global `PoweredBy` middleware stamps every HTTP response with the project identifier.
- **Build-output fingerprint banner.** Vite entry JS files are prefixed with a `/*! Powered by Ech0 — ... | AGPL-3.0-or-later */` copyright banner at build time.
- **`<meta name="generator" content="Ech0">`** added to `index.html` to identify the site generator.

### Changed

- **Copilot no longer persists empty turns.** When the model produces no text and there are no retrieval sources, the turn is skipped during persistence, preventing permanent blank bubbles in conversation history. Successfully retried turns are persisted normally.

### Fixed

- **Comment avatar seed no longer includes the array index**, keeping it consistent with the comment detail view (`e2a27ebb`).
- **Go version bumped to 1.26.4** (`ae5c5a1b`).

### Internal

- **Dependency bumps (`web/`)**: `@cap.js/widget` 0.1.53 → 0.1.54, `@dicebear/core` 10.0.1 → 10.0.2, `@dicebear/styles` 10.0.0 → 10.1.0, `eslint-plugin-vue` 10.9.1 → 10.9.2, `vite` 8.0.15 → 8.0.16, `vitest` 4.1.7 → 4.1.8, `pnpm` 11.5.0 → 11.5.2.


## [5.0.2] - 2026-06-04

A small patch release on top of 5.0.0: an unauthenticated denial-of-service fix in locale negotiation, and a vector-embedding dimension fix.

### Security

- **Fixed an unauthenticated DoS in locale parsing (GHSA-mqxv-9rm6-w8qc).** `language.ParseAcceptLanguage` runs in quadratic time on long lists of malformed subtags. The upstream CVE-2022-32149 guard caps `-` separators at 1000 but ignores `_` — which the parser aliases to `-` — so a ~1 MiB all-underscore `Accept-Language` (or `X-Locale`) value could burn seconds of CPU per request, **unauthenticated, on every route** through the global i18n middleware. Both real sinks (`ResolveLocale` and `NewLocalizer`) now sanitize their input, closing the `Accept-Language` header, the `X-Locale` header, the `?lang` query, and the authenticated user/settings locale writes in one place. Any single locale value with more than 32 `-`/`_` separators falls back to the default locale; normal short locales are unaffected.

### Fixed

- **Vector reindex/search no longer aborts with "Dimension mismatch" when the model's native output dimension differs from the configured `Dim`.** Embedding requests now send the configured dimension (the `dimensions` parameter), so the provider returns vectors that match the `vec0` index built from your `Dim` setting. Previously a model returning, e.g., 2048-d vectors against a 1024-d index failed every index/reindex upsert. Requires a model that supports custom output dimensions (e.g. OpenAI `text-embedding-3` family, Qwen `text-embedding-v4`); otherwise set `Dim` to the model's native dimension.

### Docs

- Added user guides for the AI features: **AI 问答 (Chat)** and **向量检索 (vector search / embedding)**, and rewrote the **AI 模型与近期摘要 (Agent)** guide to drop the removed Gemini protocol and inbox references.
- Corrected the **Webhook** docs' event names to match the 5.0.0 `backup → snapshot` rename (`system.snapshot`, `system.snapshot_schedule.updated`) and the `event_name` example (`EchoCreated`).


## [5.0.0] - 2026-06-04

A major **architecture-consolidation** release. Most of Ech0's cross-cutting subsystems — events, settings, tasks, key-value storage, data portability, outbound HTTP, and long-running jobs — were rewritten around one shared shape: a *thin manager + typed/self-describing registry*, with dependencies pointing inward to pure-data vocabulary. The result is the same product with a much smaller, more uniform internal surface. The version is bumped to 5.0 because of the breaking changes below: the **backup → snapshot** rename (on disk, in S3, in routes, events, and settings), the webhook `event_name` derivation, and the removal of the dead-letter retry queue. Self-hosters upgrading from 4.x should read the **Breaking Changes** section before deploying.

### Breaking Changes

- **"Backup" is gone — it is all **Snapshot** now.** Data import/export was consolidated into a single bidirectional **Migrator** domain built around one **Snapshot** resource (a zip of `data/`). The rename is end-to-end and is not auto-migrated:
  - On-disk layout: `data/files/backups/` → `data/files/snapshots/`; archive names `ech0_backup_*.zip` → `ech0_snapshot_*.zip`.
  - S3 object prefix: `backups/` → `snapshots/`.
  - Settings key: `backup_schedule` → `snapshot_schedule` — **the old scheduled-backup config is reset**; re-enable the schedule after upgrading.
  - HTTP routes: `/backup/*` → `/migration/export*` and `/migration/snapshot/schedule`. Manual export is now a job-driven async flow (see Added).
  - Event topics: `system.backup` → `system.snapshot`; `system.backup_schedule.updated` → `system.snapshot_schedule.updated`.
- **Removed the `ech0 backup` CLI command.** Import/export is now **web-only** (admin panel → "数据管理"). There is no snapshot CLI verb.
- **Snapshot download no longer carries the token in the URL.** Downloads are fetched as an authenticated blob with an `Authorization` header instead of a query-string token — safer (tokens stop leaking into logs/history), but any tooling that scripted the old token-in-URL download must be updated.
- **Webhook `event_name` lost its `Event` suffix.** Payload `event_name` is now derived from the event struct name with the suffix stripped — e.g. `EchoCreatedEvent` → `EchoCreated`. The `topic` field is **unchanged** (`echo.created` stays `echo.created`), so consumers keyed on `topic` are unaffected; consumers keyed on `event_name` must update.
- **Dead-letter retry queue removed → webhook delivery is now best-effort.** A failed webhook is retried inline (immediate retries) and then dropped; it is no longer parked in a dead-letter queue for later redelivery. The `ECH0_EVENT_DEADLETTER_BUFFER` config and the dead-letter DB table/column are gone. If you relied on guaranteed eventual delivery, treat webhooks as at-most-once after inline retries.

### Added

- **Generic long-running job subsystem (`internal/job`).** A reusable Manager with a real status machine, cancellation, persistence, status polling, and startup orphan-cleanup, with a generic `Adapt` boundary. Both **reindex** and **export** now run on it.
  - **Vector reindex is now asynchronous** — it kicks off a cancellable job with live progress and front-end polling instead of blocking the request.
  - **Snapshot export is now an async job** — trigger → poll phases → auto-download on completion, with cancel support.
- **Unified `JobProgressCard` for data management.** A reusable progress card (status pill + phase stepper + progress bar + metrics/meta grid + footer slot) shared by import and export, themed via design tokens and respecting `reduced-motion`.
  - **Export now surfaces progress the backend was already sending** but the UI had been discarding: `准备 → 打包 → 完成` phase stepping, plus the produced **file name / size** and a **re-download** action.
  - **Import** switched to the same card with real phase stepping (`解析 → 写入 → 汇总 → 完成`).
- **Configurable embedding batch size.** `/v1/embeddings` requests are now auto-split into batches (default **64** items/request, configurable via a new `batch_size` setting) to stay within provider input-array limits. Swagger, typings, and i18n updated to match.

### Changed

- **Settings are now organized into top tabs.** Six pages (storage / data / SSO / extensions / user center / preferences) moved to a top-tab layout via a new reusable `BaseSegmented` segmented control; the data-management page uses a three-tab segmented (导入 / 导出 / 快照) with the tabs lifted out of the card to match storage management.
- **Comment management split into two tabs** ("评论设置" / "评论管理"), and the comment list dropped its time column with page size reduced 20 → 10.
- **Redesigned comment-detail modal** — header bar + commenter row (Micah avatar + status / hot-comment pills) + quoted body block + info grid, centered on mobile.
- **Data import/export UI polish** across the board (new locale keys `jobProgress.*` and `exportSetting.*` in zh/en/de/ja).

### Removed

- **Dead-letter subsystem** in full: `model/queue`, `repository/queue`, the dead-letter subscriber and scheduled task, the `DeadLetterBuffer` config, the `AutoMigrate` registration, and the `dead_letters` migration column.
- **Legacy `internal/backup` package**, the never-invoked Extract→Transform→Validate→Load import pipeline, the event **publisher facade** (`contracts` / `publisher` / `registry` packages), the explicit EventBus drain component, and the empty `migrator.Worker` shell.
- **Redundant Docker `apk add tzdata`** — the timezone database is already embedded via `_ "time/tzdata"`.

### Security

- **All outbound HTTP unified behind `internal/util/egress`** with a single SSRF `Guard`: request validation, private/reserved-address blocking, a safe `DialContext`, and a response-body size limit. The previously duplicated safe-client logic (in `util/http` and the webhook HTTP client) was consolidated here and adopted by auth / comment / common / connect / setting / webhook.
- **Snapshot download tokens no longer appear in URLs** (moved to the `Authorization` header — see Breaking Changes).

### Internal

- **Event system rewrite** — one rule: dependencies point inward to a **pure vocabulary** package. `internal/event` holds event structs with self-describing `EventName()` / `OrderingKey()` and only imports models; `internal/event/bus` carries the infrastructure (`Emit` fire / `Notify` best-effort-with-warn / `On` type-routed subscribe, option presets, `EventRegistrar`). Routing is **by Go type** (no topic dimension); producers publish with a single `eventbus.Notify(...)` line. Fixed comment events silently swallowing publish errors along the way.
- **Webhook consolidated into a single subsystem** and demoted to a **plain Subscriber**: one outbound `webhook.Sender` (dedicated egress client + signing + retry) shared by the dispatcher and the settings-page TestWebhook; the bespoke bus bridge and registrar special-case were removed in favor of `eventbus.OnWithMeta` + the generic `Draining` capability for graceful worker-pool drain. The external webhook contract (topic / signing headers) is unchanged.
- **Settings engine (`internal/setting`)** — each KV config is a self-describing `Spec[T]` (key + default + normalize + migrate) behind a generic `Get`/`Set` engine plus a startup **seeder** (missing config is written on `BeforeStart`; `Get` no longer seeds as a read side-effect). `SettingService` slimmed down; auth / connect / snapshot / embedding / agent / comment now read through the engine instead of ad-hoc direct reads.
- **Unified key-value store (`internal/kvstore`)** — a single `Store` (`Get`/`Set`/`Delete`) with `Memory` (test double) and `Persistent` (delegates to the keyvalue repository) implementations, Wire-bound so the repository layer is no longer imported by services. Replaced five duplicated narrow interfaces; `Set` merges the old add/update/upsert variants.
- **Tasker → `task.Manager` + `scheduled` registry** — the old god-object + manual `Start` registration became a thin Manager holding `[]Task` with an optional `StopHook`, and each cron task (cleanup / visitor-snapshot / export) moved to its own self-describing `internal/task/scheduled` sub-package, eliminating the 7-arg constructor.
- **`storage.Manager` promoted to a process-wide shared singleton.**
- **Toolkit layering flattened** — `async` / `tui` sank to `util/{async,tui}`; `util/http` was dissolved (`TrimURL` → `util/url`, MIME mapping folded back into the file domain as one `canonicalMIMEForExt` table); the webhook `infra/httpclient` was flattened into the webhook root package.
- **DI graph regenerated** (`make wire`) across all of the above; CLAUDE.md and the dev docs (`snapshot-design.md` added; webhook-usage / job-runner-design / timezone-design updated) kept in sync.

## [4.9.2] - 2026-06-02

### Added

- **Copilot "year-in-review" / range summaries** — a dedicated `summarize_echos` tool that exhaustively aggregates echos over a date range instead of sampling top-k. It paginates through the *entire* range (hard cap 5000, truncating to the most recent with an honest notice) and adapts to the model's context window: small ranges are summarized in one pass, large ones via per-month map-reduce. A new optional **context window** setting (entered as a friendly `256k` / `1m`, stored as tokens) drives the aggregation budget. Coverage is reported back live via an SSE `coverage` event and a "📚 covered N echos" status bar, so nothing is silently truncated.
- **`stats_overview` Copilot tool** — pure in-memory aggregation that gives the model exact quantitative facts (total count, active days, by-month, most active month, top tags).
- **"Optional" badge on the vector-index tab** in Copilot settings, signalling that the feature is not required (new `commonUi.optional` i18n key across zh/en/de/ja).

### Changed

- **Chat streaming is noticeably faster** with zero visual change: `AnimatedMarkdown` now freezes the already-finalized prefix and only re-parses the unfinished tail (multi-paragraph answers drop from ~O(n²) to ~O(n) parse work, with stable block keys so animations never replay), `TheChatBox` skips a forced reflow on the per-token reveal hot path, and each message turn is layout-isolated via `contain: layout style`.
- **Copilot Agent tuning ("seven-piece" pass)**: timezone-correct "today" / date parsing via `X-Timezone` (fixes day-boundary off-by-one across UTC), Anthropic prompt-cache breakpoint on the static tools+system prefix, relaxed and context-window-scaled `top_k` (default 6, up to 20), configurable `ECH0_AGENT_MAX_ROUNDS` (default 4), bounded-concurrency tool execution, and a per-round token budget that recycles the oldest tool results when the context limit is hit.
- **Retrieval is now scoped to the current user.** Embedding search and echo queries filter by author, so Chat and retrieval only ever surface the conversation owner's own echos.
- **Embedding `base_url` is passed through literally** to the OpenAI-compatible client (no more silent rewriting), with a clearer hint to enter the root address without `/embeddings`.

### Fixed

- **Range/year summaries no longer miss data or over-weight recent echos.** The aggregation path now keeps paging until the range is fully covered (instead of stopping on the first non-full page), clamps oversized page sizes to 100 rather than resetting them to 10, and enriches each line with tags, extension markers (music / website / location), and image counts — image-only echos now count too.
- **Reindex success toast was blank** — the handler now returns a localized success message instead of empty data.
- **Embedding backfill failures are now surfaced** — when every item fails (`indexed=0`, `failed>0`) the underlying error (404 / auth) is propagated instead of a silent empty message.
- **Chat input box no longer covers history** — it has a max height (~5 lines) with internal scrolling, and the transcript yields space in real time; the empty-state composer is vertically centered and settles smoothly once the first message is sent.
- **Streaming source block no longer jitters or flickers** — replaced the rAF stick-to-bottom polling with intent-driven pinning + `ResizeObserver`, disabled native scroll anchoring, and moved the sources block clear of the bottom mask gradient.
- **`DeadLetterConsumeTask` scheduling-failure log** was mislabeled as `WebhookRetryTask`; backup setting now correctly documents its default as "disabled".

### Internal

- **Dependency bumps (Go)**: `go-patch-minor` group (6 updates).
- **Dependency bumps (`web/`)**: `@dicebear/core` 9.4.2 → 10.0.1 (migrated to `@dicebear/styles`), plus the `web-patch-minor` group (4 updates).
- **README**: each language's feature list now includes Ech0 Copilot (recap summaries & Chat).

## [4.9.0] - 2026-05-31

### Added

- **LLM Chat — talk to your timeline (RAG).** A new owner-only AI chat that answers questions over your own echos. Echos are incrementally indexed into a `sqlite-vec` vector store on create/update/delete (plus an admin full-reindex endpoint), retrieved top-k by semantic similarity, and answered with streaming SSE. Supports multi-turn conversation memory and tool-calling retrieval (`search_echos`, with tag / date filters). Embedding is configured independently via an OpenAI-compatible `/v1/embeddings` endpoint; the chat itself speaks the OpenAI or Anthropic protocol. An optional **multimodal** mode feeds matched echo images to the model, and retrieval hits surface their Extension shares (music / website / location) and image thumbnails in the UI. Entry point lives in the homepage sidebar with a dedicated `/chat` view; all settings are grouped under the Copilot panel.
- **"On This Day" API** — returns echos posted on this date in previous years.
- **Random Echo API** — returns a single random echo.

### Fixed

- **Editing an echo returns to the same timeline page** instead of jumping back to the top.
- **TWEET extension data is restored when editing an echo**, so Tweet/X cards no longer lose their embed on save.
- **Timeline pager stays in sync with the URL** after filter changes.

### Internal

- **`agent` package refactored into a `copilot` domain** with a protocol abstraction (renamed from "provider"), tool-calling retrieval, and a `GenerateStream` API (real streaming on OpenAI; single-block v1 fallback on Anthropic). The Gemini integration was dropped.
- **Frontend typings split** — `app.d.ts` broken into per-domain `.d.ts` files.
- **Toolchain**: pnpm bumped to 11.5.0; `check.sh` hardened.
- **CI**: auto-deploy `site` & `hub` to Cloudflare Pages.
- **Dependency bumps (Go)**: `go-patch-minor` group (6 updates).
- **Dependency bumps (`web/`)**: `vue` 3.5.35, `vue-router` 5.1.0, `vue-tsc` 3.3.2, `npm-run-all2` 8.0.4 → 9.0.1, plus the `web-patch-minor` group (4 updates).

## [4.8.2] - 2026-05-23

### Fixed

- **Timeline scroll jank**: echo cards no longer keep `4 × N` global scroll/click handlers attached while their action menu is closed; listeners now bind only while the menu is open, with `passive: true` on scroll. The always-on `will-change-transform` wrapper was also dropped.
- **`/api/files/...` images missing `Cache-Control` in browsers**: `StaticFileSecurity` was setting the header after `c.Next()`, too late on Chrome's Range-request path (curl saw it; browsers didn't). Header is now resolved from the URL extension and set before `c.Next()`.

### Internal

- **Dependency bumps (Go)**: `github.com/anthropics/anthropic-sdk-go` 1.42.0 → 1.43.0, `google.golang.org/genai` 1.56.0 → 1.57.0.
- **Dependency bumps (`web/`)**: `@cap.js/widget` 0.1.52 → 0.1.53, `vue-i18n` 11.4.2 → 11.4.4, `eslint` 10.3.0 → 10.4.0, `tsx` 4.22.0 → 4.22.1, `js-cookie` 3.0.5 → 3.0.7, `baseline-browser-mapping` 2.10.31 → 2.10.32 (transitive).

## [4.8.1] - 2026-05-15

### Added

- **Zen mode** for a cleaner, distraction-free writing and reading experience.
- **New Tweet card support**, improving how Tweet/X links are displayed in Echo content.

### Changed

- **CLI / TUI experience refined** with small usability and presentation improvements.

## [4.8.0] - 2026-05-13

### Added

- **RSS feed now renders as a styled page when opened in a browser**. A new XSLT stylesheet at `web/public/rss.xsl` turns the raw Atom feed into a paper-themed reading view (light + dark, mobile-friendly) when the visitor's `Accept` header includes `text/html`; dedicated RSS readers still receive `application/atom+xml` with the same bytes, so the subscription contract is unchanged. The Atom document gets an inline `<?xml-stylesheet href="/rss.xsl"?>` PI, and the handler in `internal/handler/common/common.go` now content-type-switches on `Accept`.

### Changed

- **Echo detail dividers restyled.** The dashed `border-bottom` under the detail-page meta strip (`TheEchoDetail.vue`) and the dashed `border-top` above the interactions zone (`TheEchoInteractions.vue`) have been replaced with a repeating linear-gradient "stitched" rule (5px dash, 3px gap), so the divider stays crisp on retina displays and aligns with the wider design system.
- **`HomeHeader` GitHub link hidden.** The Github icon next to the RSS button on the homepage header is commented out; only RSS, theme toggle, and the other built-in actions remain. The about page still surfaces the repo URL.
- **Panel dashboard meta strip** no longer prints `VERSION x.y.z` — version is now surfaced only on the About page (the single source of truth from `internal/version`).
- **Chinese license caption (about page) reworded** from "本软件以 …" ("This software is …") to "开源协议：…" ("Open-source license: …"), reading more naturally as a key/value pair rather than a sentence fragment.

### Fixed

- **`scripts/ech0.sh` install script** no longer 404s when a Helm chart release is published shortly after an app release. `chart-releaser-action` creates a `ech0-X.Y.Z` GitHub release for the Helm chart, which GitHub automatically flips to "latest" since it has a newer timestamp than the corresponding `vX.Y.Z` app release. The chart release only ships a `.tgz`, so `releases/latest/download/ech0-linux-<arch>.tar.gz` returned 404. The install script now hits the GitHub Releases API directly and picks the newest `v*` tag, hard-failing with a clear error if no matching release can be resolved.
- **`release_helm.yml` workflow** now re-marks the originating `vX.Y.Z` app release as "Latest" after publishing the chart release, so the GitHub UI and tooling that resolves `/releases/latest` (browsers, install scripts, third-party mirrors) continue to land on the platform-binary release rather than the chart-only one.

### Internal

- **Vendored three previously external libraries into `pkg/`**, so the entire runtime now builds from this repo alone:
  - `github.com/lin-snow/Busen` → `pkg/busen` (imported as `github.com/lin-snow/ech0/pkg/busen`) — async in-process event bus, ~5k LOC + tests.
  - `github.com/lin-snow/VireFS` → `pkg/virefs` (imported as `github.com/lin-snow/ech0/pkg/virefs`) — unified local/S3 filesystem abstraction backing `internal/storage`, now with first-class **zip-archive support**: `plugin/zip/Unpack` (extract a zip into a destination with a key prefix) and a read-only `ZipFS` (`Get`/`List`/`Stat`/`Walk` over the archive). `S3Config` adds presets for AWS / MinIO / R2; `schema` adds extension-based routing; `Walk` supports directory skipping.
  - `github.com/lin-snow/gocap` → `pkg/gocap` (imported via `internal/captcha` for the built-in CAPTCHA) — challenge/redeem PoW captcha core: `Service.Challenge` / `Service.Redeem` / `SiteVerify`, in-memory `memstore` with GC, HTTP transport (`/challenge`, `/redeem`, `/siteverify`), middleware (error handling, client-IP extraction), rate limiting, secret hashing (`HashSecret`, `SecureSecretEqual`), JWT-style `ChallengeClaims`. CLAUDE.md updated to point at the new import paths.
- **SPDX-License-Identifier + Copyright headers** added to every file under `pkg/busen`, `pkg/virefs`, `pkg/gocap`, completing the AGPL-3.0 header coverage for the vendored sources.
- **Dependency bumps (Go)**: `github.com/anthropics/anthropic-sdk-go` 1.38.0 → 1.41.0, `github.com/go-webauthn/webauthn` 0.17.2 → 0.17.3, `golang.org/x/mod` 0.35.0 → 0.36.0, `golang.org/x/net` 0.53.0 → 0.54.0, `golang.org/x/text` 0.36.0 → 0.37.0, `google.golang.org/genai` 1.55.0 → 1.56.0.
- **Dependency bumps (`web/`)**: `@cap.js/widget` 0.1.46 → 0.1.50, `vue-virtual-scroller` 3.0.2 → 3.0.3, `@types/node` 25.6.0 → 25.6.2, `vite-plugin-vue-devtools` 8.1.1 → 8.1.2.
- **Dependency bumps (`hub/`, `site/`)**: `fast-uri` 3.1.0 → 3.1.2 (transitive).

## [4.7.5] - 2026-05-07

### Added

- **`AGENTS.md`** provides a compact reference for AI agents working in the Ech0 repository. Documents the project architecture, available `make` / `pnpm` commands, backend layering, Wire DI, event bus, frontend build output, and key in-repo docs.
- **`justfile`** adds a `just` task runner mirroring all Makefile recipes, giving developers who prefer `just` a first-class workflow.

### Changed

- **SMTP sender address** can now be configured independently of `SMTPUsername`. A new `SMTPSender` field in `EmailNotifySetting` lets operators set the envelope `From:` address that actually appears in outbound comment-notification emails — useful when the SMTP provider requires a fixed sender (e.g. Postmark, SES) while credentials differ. The panel's comment-manager UI exposes the new field; existing deployments fall back to `SMTPUsername` when the field is empty.
- **`BaseSelect` component** restyled: the trigger button and dropdown now use CSS custom properties (`--select-*`) for background, border, focus ring, and disabled states, matching the rest of the design system. Keyboard navigation (↑/↓/Enter/Space/Escape) and `aria-expanded` semantics are unchanged.
- **Hub `TheImageGallery` async loader** now retries up to 3 times on chunk-load failure before surfacing the error, reducing transient failures on flaky networks.

### Internal

- **Dependency bumps (`web/`)**: `vue` 3.5.33 → 3.5.34, `vue-i18n` 11.4.0 → 11.4.2, `eslint-plugin-vue` 10.9.0 → 10.9.1, `jiti` 2.6.1 → 2.7.0, `stylelint` 17.10.0 → 17.11.0, `vite` 8.0.10 → 8.0.11, `vue-tsc` 3.2.7 → 3.2.8.

## [4.7.4] - 2026-05-04

### Changed

- **Hub Echo card redesign**. The card displayed in the cross-instance Hub feed has been visually rebuilt to read as a quoted post rather than a miniature timeline card:
  - Header is now a two-segment row — avatar + instance name (with the verified badge) on the left, an Ech0 logo on the right that links directly to the source Echo page (replacing the redundant footer "jump to echo" icon).
  - The `@username` line was removed; only the instance name (`server_name`) is shown, since for hub consumers the source instance is the meaningful identity.
  - Avatar shrunk from `w-10 h-10 sm:w-12 sm:h-12` (40/48px) to `w-6 h-6 sm:w-7 sm:h-7` (24/28px).
  - Card padding tightened (`p-3.5 sm:p-4` → `p-3 sm:p-3.5`) and corner radius bumped (`rounded-sm` → `rounded-lg`).
  - A subtle accent bar (`var(--color-accent)`, 3 × 16–18 px) is rendered at the left edge of the card, vertically centered with the avatar, as a visual citation marker.
  - Body text size and paragraph spacing pulled in to match a quoted-post density (`font-size: 0.9rem`, `line-height: 1.55`, paragraph margin 0.55rem).
  - Body text and embedded gallery now share the card's natural padding edge — `TheImageGallery`'s internal `w-[88%] mx-auto` was overridden at the hub-card level so gallery, body, date row, and like row all align to the same vertical guides. The override is scoped via `:deep()` so the main timeline's gallery presentation is unchanged.
  - Footer simplified to a single row: date on the left, like button + count on the right, both in `text-xs` muted style; the dedicated "jump to echo" icon was removed (already covered by the header logo).

- **CLI / TUI strings translated to English**. The interactive `ech0` TUI menu, all `cobra` command descriptions (`ech0`, `ech0 serve`, `ech0 backup`, `ech0 version`, `ech0 info`, `ech0 hello`), and the boxed startup / shutdown messages were emitted in Simplified Chinese only. They are now in English so non-Chinese-speaking operators can use the binary without guessing.

### Internal

- **`internal/cli/cli.go` + `cmd/*.go`** — strings only, no behavioural change.
- **`fix(workflow): add permissions for content access in i18n-guardrails`** — the i18n-guardrails GitHub Actions workflow needed `contents: read` to check out repository content under stricter default token permissions; without it the workflow could not read source files on protected branches.
- **Dependency bumps (`web/`)**: `vue-virtual-scroller` 3.0.0 → 3.0.2, `stylelint` 17.9.1 → 17.10.0.
- **Dependency bumps (Go)**: `github.com/caarlos0/env/v11` 11.4.0 → 11.4.1, `github.com/go-webauthn/webauthn` 0.17.0 → 0.17.2, `go.uber.org/zap` 1.27.1 → 1.28.0, `google.golang.org/genai` 1.54.0 → 1.55.0.
- **Design assets** (`docs/design/`): added social-preview templates (1280×640 JPG/PNG) and a six-frame `Ech0_carousel/` design source for marketing/release imagery. New screenshots under `docs/imgs/` for the v4.7.0 about page, dashboard, and a no-sidebar variant. Documentation only — not shipped in the binary.

## [4.7.3] - 2026-05-03

This is primarily a security release: six advisories disclosed since v4.7.2 are addressed. All deployments are encouraged to upgrade.

### Changed

- **Editor publish controls** split the old "toggle privacy" icon into two explicit actions, **Publish as public** and **Publish as private**. The previous flow required clicking a toggle and then publish, which often surprised users into publishing with the wrong visibility. New translation keys `publishEchoPublic` / `publishEchoPrivate`; legacy `togglePrivacy` / `privacySwitched` / `privacyPrivate` / `privacyPublic` removed.

### Security

- **[GHSA-rj4g-rqgh-rx9h](https://github.com/advisories/GHSA-rj4g-rqgh-rx9h)** — Commenter email PII leak on public endpoints. `GET /api/comments` and `/api/comments/public` returned the raw `Comment` struct, exposing every guest commenter's email (plus `user_id`, `ip_hash`, `user_agent`) to any unauthenticated caller. Public endpoints now serialize a `PublicComment` DTO that strips those fields; admin `/panel/comments` keeps the full struct for moderation.
- **[GHSA-3v85-fqvh-7rxf](https://github.com/advisories/GHSA-3v85-fqvh-7rxf)** — Stored XSS via the RSS feed. `GenerateRSS` interpolated tag names with `%s` and rendered echo bodies with raw-HTML markdown enabled, both wrapped inside Atom `<summary type="html">`. RSS readers that honour `type="html"` decoded the entities and executed any embedded `<script>`. Tag names are now HTML-escaped, the markdown renderer skips raw HTML for the RSS path, and tag write paths reject `<>"'&` as defence in depth.
- **[GHSA-pj6q-4vq4-r8cg](https://github.com/advisories/GHSA-pj6q-4vq4-r8cg)** — Like-spam on the public Echo endpoint. Anonymous `PUT /echo/like/:id` had no rate limit or de-duplication, so a single IP could arbitrarily inflate `fav_count` and repeatedly trigger four-key cache invalidation. New `RateLimitWithIdempotency` middleware combines a 2 rps / 5 burst per-IP token bucket with a 1-hour idempotency window keyed on `(IP, echoID)`; repeated requests inside the window return the same response shape as a fresh success, so clients see no behaviour change.
- **[GHSA-8mc6-xjpr-h98x](https://github.com/advisories/GHSA-8mc6-xjpr-h98x)** — SSRF via the Connect peer-info fetch. `fetchPeerConnectInfo` used the raw `SendRequest` helper with no URL validation, so an admin-added peer URL could point at private networks or cloud metadata (e.g. `169.254.169.254`, `kubernetes.default.svc`); the public `GET /api/connects/info` then triggered the outbound request. Switched to `SendSafeRequest` (URL allowlist + `SecureDialContext` against DNS rebinding); `AddConnect` also rejects malicious URLs at insertion time.
- **[GHSA-p64j-f4x9-wq66](https://github.com/advisories/GHSA-p64j-f4x9-wq66)** — OAuth redirect URI bypass. `parseAndValidateClientRedirect` only compared scheme+host, so an attacker could supply any same-host path; the server still appended `?code=<one-time>` there, where Referer leaks, third-party analytics, or an open-redirect on the same host could hand the code over and let the attacker exchange it for the victim's tokens. Comparison is now scheme+host+path per RFC 6749 §3.1.2 (query/fragment still excluded — the server needs to append `?code=...`). `GetOAuthLoginURL` and `BindOAuth` also reject bad redirect URIs before signing the state JWT.
- **[GHSA-fpw6-hrg5-q5x5](https://github.com/advisories/GHSA-fpw6-hrg5-q5x5)** — Access tokens issued with `NEVER_EXPIRY` could not be revoked. All three revocation paths failed: `/api/auth/logout` panicked dereferencing nil `ExpiresAt`, `RevokeToken` skipped the cache write because `remainTTL <= 0`, and admin "Delete Token" only removed the DB row without writing the blacklist — so a leaked token kept authenticating until `JWT_SECRET` was rotated. `CreateAccessClaimsWithExpiry` now falls back to a 100-year `ExpiresAt` (semantically still "never expires" but every revocation path receives a positive TTL); `Logout` tolerates legacy nil `ExpiresAt`; `DeleteAccessToken` looks up the JTI and writes it to the blacklist before deleting the row. Known limitation: the blacklist is still in-memory `ristretto` and is dropped on process restart.

### Internal

- **`tldts` / `tldts-core`** bumped to 7.0.30 in `web/`.

## [4.7.2] - 2026-05-02

### Added

- **In-header back button on the Echo detail page**, replacing the standalone arrow that used to sit above the card. Right-aligned, pill-shaped, ringed; falls back to `/` when there is no history to pop. New translation key `commonNav.back` across `zh-CN` / `en-US` / `ja-JP` / `de-DE`.

### Changed

- **LCP image priority**: the first image of the timeline's first Echo card and of the Echo detail gallery are now loaded with `loading="eager"` and `fetchpriority="high"`. A new `priority` prop is threaded through `TheImageGallery` → all gallery layouts (Carousel / Grid / Horizontal / Stack / Waterfall) → `GalleryImageItem`; everything else stays lazy + async.
- **Echo detail header divider** is now dashed instead of solid, to visually decouple the meta strip from the body.
- **Timeline enter animation** flipped: Echo cards now drop in from `translateY(-18px)` instead of rising from `+18px`, so the stagger reads as "newest falling into place" rather than continued scroll.

### Fixed

- **Mobile scroll restoration on `/`**: the homepage now also persists the `window` scroll position (key `home:window:scrollTop`) in addition to the inner timeline column. Returning from a detail page on small viewports — where scrolling happens on `window`, not on `mainColumn` — no longer snaps to the top.
- **First-paint scroll snap-back**: scroll restoration now waits for the first batch of Echos to render (`echoList.length > 0 && !isLoading`) before applying the saved offset, eliminating the "scrolls to 0, then jumps back" flicker on slow networks.
- **Router `scrollBehavior`**: non-`home` routes now honor `savedPosition` for browser back/forward and reset to top on fresh navigation; `home` continues to manage its own restore inside `HomePage`.

### Performance

- **Long-lived browser cache for uploaded media**: `StaticFileSecurity` now emits `Cache-Control: public, max-age=31536000, immutable` for inlineable MIME types (image/audio). This is safe because stored filenames are content-hashed by the storage layer — reusing a key implies identical bytes — so cached responses can never go stale against a different payload.

### Internal

- **`scripts/check.sh`** consolidates the pre-PR pipeline (SPDX header check + backend fmt/lint/swagger + frontend format/lint/stylelint/i18n) into a single orchestrator that runs every step even on failure and prints a summary table. `make check` / `make dev-lint` now delegate to it. Two new shortcuts: `make spdx` and `make spdx-check`.
- Sponsor wall: added `@star-uu` and corrected the sponsorship date.

## [4.7.1] - 2026-05-01

### Added

- **Admin tag creation** on the Tag Manager page. Admins can now create orphan tags ahead of time without having to publish an Echo first. New `POST /api/tag` endpoint (scope `echo:write`, admin-only, idempotent on duplicate name).
- **`TheEchoMeta` component** on the Echo detail page, showing creation/update time (precise to the minute), word count, the full tag list, and a private flag.
- **`TheEchoInteractions` component** that bundles share, like, and comments into one interaction zone below each Echo detail. The comment composer is collapsed by default behind a pill trigger to keep the page calm.
- **Hover "open detail" icon** on timeline Echo cards (next to the date) for one-click navigation to the detail page.

### Changed

- **Editor tag input** is now a multi-select picker over existing tags only (no free-typed `#tag` parsing). Capped at 3 tags per Echo, with a toast warning on overflow.
- **Echo detail page** redesigned: transparent canvas background (no card frame), hero header with avatar + server/username, then `TheEchoMeta`, body, and `TheEchoInteractions`.
- **Timeline cards** no longer render inline tags; tag filtering is handled via the existing sidebar / search.
- **About page footer**: "用心打造" / "Built with care" updated to use the heart glyph (`用 ❤️ 打造` / `Built with ❤️`), synced across `zh-CN`, `en-US`, `ja-JP`, and `de-DE`.

### Fixed

- **Tag picker popover** no longer overflows the right edge of the screen on mobile (< 640px). The panel now anchors to the editor toolbar and spans its full inner width on small viewports.
- **Tag picker popover** no longer renders behind the editor image preview. The action row was given an explicit stacking context so the popover layers above subsequent siblings.

## [4.7.0]

### Added

- **About page** (`/about`) reachable from the homepage banner. Displays the running instance's version, commit hash, build time, license, copyright, author, and a source-code link pinned to the exact commit. Implements AGPL-3.0 §13 (network users may obtain the corresponding source).
- **`internal/version` package** as the single source of truth for build / release metadata (Version, License, Author, RepoURL, StartYear, plus ldflags-injected Commit and BuildTime). Replaces the version constant that used to live in `internal/model/common`.
- **`make bump NEW_VERSION=X.Y.Z`** target that prepares a clean version-bump commit (does not auto-commit or tag).
- **CI guardrail**: the release workflow now refuses to build when the pushed git tag (`vX.Y.Z`) and `internal/version.Version` disagree. Prevents publishing artifacts that lie about their own version.
- **SPDX / Copyright headers** on every `.go` / `.ts` / `.vue` source file, plus a maintenance script `scripts/add-spdx-headers.mjs` (write / `--dry-run` / `--check` modes).
- **`docs/dev/release-process.md`** documenting the standard release procedure.

### Changed

- **`/api/hello` response shape**: dropped the legacy `github` field; added `commit`, `build_time`, `license`, `author`, `repo_url`, and `copyright`. The frontend reads version metadata from this endpoint instead of hardcoding it. Pre-PR consumers of the `github` field should switch to `repo_url` (no in-tree consumer existed).
- **`web/package.json`** now declares `license`, `author`, and `homepage` so npm tooling and SPDX scanners pick up project licensing without parsing the repo.

### Security

- Pinned `serialize-javascript` to `^7.0.5` in `hub/pnpm-lock.yaml` via `pnpm.overrides`, clearing two Dependabot alerts:
  - [GHSA-5c6j-r48x-rmvq](https://github.com/advisories/GHSA-5c6j-r48x-rmvq) — RCE via `RegExp.flags` and `Date.prototype.toISOString` (HIGH)
  - [GHSA-qj8w-gfj5-8c6v](https://github.com/advisories/GHSA-qj8w-gfj5-8c6v) — CPU-exhaustion DoS via crafted array-like objects (MEDIUM)

  Practical risk in this repo was negligible (the vulnerable code only runs at PWA build time on developer-controlled input), but the alerts are now resolved at the supply-chain level.

[Unreleased]: https://github.com/lin-snow/Ech0/compare/v5.3.0...HEAD
[5.3.0]: https://github.com/lin-snow/Ech0/compare/v5.2.5...v5.3.0
[5.2.5]: https://github.com/lin-snow/Ech0/compare/v5.2.4...v5.2.5
[5.2.4]: https://github.com/lin-snow/Ech0/compare/v5.2.3...v5.2.4
[5.2.3]: https://github.com/lin-snow/Ech0/compare/v5.2.2...v5.2.3
[5.2.2]: https://github.com/lin-snow/Ech0/compare/v5.2.1...v5.2.2
[5.2.1]: https://github.com/lin-snow/Ech0/compare/v5.2.0...v5.2.1
[5.2.0]: https://github.com/lin-snow/Ech0/compare/v5.1.0...v5.2.0
[5.1.0]: https://github.com/lin-snow/Ech0/compare/v5.0.2...v5.1.0
[5.0.2]: https://github.com/lin-snow/Ech0/compare/v5.0.0...v5.0.2
[5.0.0]: https://github.com/lin-snow/Ech0/compare/v4.9.2...v5.0.0
[4.9.2]: https://github.com/lin-snow/Ech0/compare/v4.9.0...v4.9.2
[4.9.0]: https://github.com/lin-snow/Ech0/compare/v4.8.2...v4.9.0
[4.8.2]: https://github.com/lin-snow/Ech0/compare/v4.8.1...v4.8.2
[4.8.1]: https://github.com/lin-snow/Ech0/compare/v4.8.0...v4.8.1
[4.8.0]: https://github.com/lin-snow/Ech0/compare/v4.7.5...v4.8.0
[4.7.5]: https://github.com/lin-snow/Ech0/compare/v4.7.4...v4.7.5
[4.7.4]: https://github.com/lin-snow/Ech0/compare/v4.7.3...v4.7.4
[4.7.3]: https://github.com/lin-snow/Ech0/compare/v4.7.2...v4.7.3
[4.7.2]: https://github.com/lin-snow/Ech0/compare/v4.7.1...v4.7.2
[4.7.1]: https://github.com/lin-snow/Ech0/compare/v4.7.0...v4.7.1
