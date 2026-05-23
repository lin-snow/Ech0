# Changelog

All notable user-visible changes to Ech0 are recorded here.

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html), and this file follows the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

For releases prior to v4.6.5, see the [GitHub releases page](https://github.com/lin-snow/Ech0/releases) — earlier release notes are not retroactively imported here.


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

[Unreleased]: https://github.com/lin-snow/Ech0/compare/v4.8.0...HEAD
[4.8.0]: https://github.com/lin-snow/Ech0/compare/v4.7.5...v4.8.0
[4.7.5]: https://github.com/lin-snow/Ech0/compare/v4.7.4...v4.7.5
[4.7.4]: https://github.com/lin-snow/Ech0/compare/v4.7.3...v4.7.4
[4.7.3]: https://github.com/lin-snow/Ech0/compare/v4.7.2...v4.7.3
[4.7.2]: https://github.com/lin-snow/Ech0/compare/v4.7.1...v4.7.2
[4.7.1]: https://github.com/lin-snow/Ech0/compare/v4.7.0...v4.7.1
