# Changelog

All notable user-visible changes to Ech0 are recorded here.

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html), and this file follows the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format.

For releases prior to v4.6.5, see the [GitHub releases page](https://github.com/lin-snow/Ech0/releases) — earlier release notes are not retroactively imported here.

## [4.7.0](https://github.com/lin-snow/Ech0/compare/v4.6.4...v4.7.0) (2026-05-01)


### Added

* About page, internal/version package, SPDX headers (AGPL §13) ([03f6c49](https://github.com/lin-snow/Ech0/commit/03f6c49e23e4414e77dca87fb833d6e505d8e088))
* add jump action to video card and update translations ([80bf93b](https://github.com/lin-snow/Ech0/commit/80bf93b6128002505da93e2e4e9c43509ddba4e2))
* **backend:** extract internal/version package + inject build metadata ([0beda9b](https://github.com/lin-snow/Ech0/commit/0beda9beab7e5985c5f2319cfe2cd1d75531fb61))
* **web:** add lite upload primitives under lib/file ([c794d74](https://github.com/lin-snow/Ech0/commit/c794d74d2a5e34d98c2a6f1a7a69d46bea0280a8))
* **web:** add TheUploader component ([11c6449](https://github.com/lin-snow/Ech0/commit/11c644943d90c85ad30fb76978cebc2d3bbed618))
* **web:** cap image uploads at 20 MiB on the frontend ([9e5ee70](https://github.com/lin-snow/Ech0/commit/9e5ee70a76fdd79bb1f6246ab9125ea4f3e03a1b))
* **web:** enhance uploader UI with improved layout and compression feedback ([e8575a4](https://github.com/lin-snow/Ech0/commit/e8575a4d25c8082b74e15660cc0ab028e94536d5))
* **web:** make TheUploader configurable per scenario ([dd0444e](https://github.com/lin-snow/Ech0/commit/dd0444e4a27c86f5119dac73f508ef4b63dc9e90))
* **web:** redesign about page and add icon entry on home banner ([d0bcd36](https://github.com/lin-snow/Ech0/commit/d0bcd36c351d61a3bb770bf01a46b1dd9c3ca7a4))
* **web:** redesign about page and add icon entry on home banner ([7c42555](https://github.com/lin-snow/Ech0/commit/7c42555fbed9f6f53b873a303c2f99dd594923d7))
* **web:** redesign uploader queue (size info + drag-to-reorder) ([0e3b521](https://github.com/lin-snow/Ech0/commit/0e3b521cf5e803b334ab892649f1363c4b5f5abb))
* **web:** replace Uppy with lightweight in-house uploader ([f638dda](https://github.com/lin-snow/Ech0/commit/f638dda96a43f2b4f3e5ff8da10488e49ab1f386))
* **web:** wire commit hash + license metadata into About page ([d5c361e](https://github.com/lin-snow/Ech0/commit/d5c361e8cc920da13192954f45cc43c4373887b4))


### Fixed

* replace static import with async component for TheImageGallery to avoid circular dependency ([9eb811b](https://github.com/lin-snow/Ech0/commit/9eb811b628601d28a0bce5b67e735cb8ce286498))
* update import for TheMdPreview to avoid circular dependency issues ([ca2b9df](https://github.com/lin-snow/Ech0/commit/ca2b9dffe398c978eaf85c45f07873671aa2b94f))
* **web:** correct delete behaviour in TheEditorImage ([a36a064](https://github.com/lin-snow/Ech0/commit/a36a0648289aae63d8378925d42f057e93b40960))
* **web:** correct delete bug + make TheUploader scenario-configurable + 20 MiB image cap ([e6c7cb3](https://github.com/lin-snow/Ech0/commit/e6c7cb3c90cddcdd80d090aeee3f6ac472eb1cc8))
* **web:** format copyright computation in AboutPage.vue ([7836cf1](https://github.com/lin-snow/Ech0/commit/7836cf130b6f5e0248cf9bbb0d79011f1e71d4eb))


### Changed

* **web:** centralize upload status/kind enums and dedupe checks ([fef495a](https://github.com/lin-snow/Ech0/commit/fef495acd68520ab370aadba7392f825c24aad6f))

## [Unreleased]

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

[Unreleased]: https://github.com/lin-snow/Ech0/compare/v4.6.4...HEAD
