# Release Process

This document describes how to cut a new release of Ech0.

## TL;DR

```bash
# starting from a clean tree on main, in sync with origin
just check                                              # full pre-release lint pass
just bump 4.6.5                             # edits internal/version/version.go
$EDITOR CHANGELOG.md                                    # rename [Unreleased] → [4.6.5] - YYYY-MM-DD; add a fresh empty [Unreleased]
git commit -am 'chore(release): v4.6.5'
git tag -a v4.6.5 -m 'Release v4.6.5'
git push origin main v4.6.5                             # tag push triggers .github/workflows/release.yml
gh release view v4.6.5                                  # verify artifacts after CI completes
```

The rest of this document explains *why* each step exists and how to handle edge cases.

## Versioning

Ech0 follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html):

- **MAJOR** (`X.Y.Z` → `(X+1).0.0`) — breaking changes to the API, auth model, storage layout, config keys, or anything else self-hosted operators rely on. Requires migration notes in CHANGELOG.
- **MINOR** (`X.Y.Z` → `X.(Y+1).0`) — backwards-compatible new features.
- **PATCH** (`X.Y.Z` → `X.Y.(Z+1)`) — backwards-compatible bug fixes.

Pre-release tags (`v5.0.0-rc.1`, `v5.0.0-beta.1`) are used for major bumps where you want real users to test before the stable promotion. The `internal/version.Version` const must include the suffix verbatim (`Version = "5.0.0-rc.1"`) — the CI guard verifies this.

## The single source of truth

The version is declared **once** in [`internal/version/version.go`](../../internal/version/version.go):

```go
const (
    Version = "4.6.4"
    ...
)
```

Everywhere else (`/hello`, `/healthz`, the About page, `ech0 version`, MCP server identification, connect federation) reads from this single const. Build-time metadata (`Commit`, `BuildTime`) is injected via `-ldflags -X` from the justfile / Dockerfile / release workflow.

**Never edit the version string in any other file.** If you find a hardcoded version anywhere, that is a bug to fix, not a place to also bump.

## Pre-release sanity

Before bumping the version, make sure the working state is releasable:

1. **Clean tree on `main`, in sync with origin.**
   ```bash
   git checkout main
   git pull --ff-only
   git status   # must be empty
   ```

2. **All recently-merged PRs have CI green on `main`.** Check the Actions tab.

3. **Full local check passes.**
   ```bash
   just check   # backend fmt/lint + openapi drift check + web format/lint/i18n/style
   go test ./...
   pnpm -C web test:unit   # if you've changed anything frontend
   ```

4. **Decide the new version.** Look through merged PRs since the last tag; the highest-impact change determines the bump (breaking → MAJOR, feature → MINOR, fix → PATCH).

   ```bash
   git log v4.6.4..HEAD --oneline   # everything since the last release
   ```

## Bumping the version

```bash
just bump 4.6.5
```

This target:

- Validates `NEW_VERSION` is set and matches semver (`X.Y.Z` or `X.Y.Z-prerelease`).
- Refuses to run when the working tree is dirty (otherwise the release commit would carry unrelated drift).
- Edits `internal/version/version.go` to the new value.
- Runs `go build ./...` as a sanity check; reverts the file automatically if the build fails.
- Prints the diff plus the next-step commands to run.

`just bump` **never** auto-commits, never tags, never pushes — those are deliberate human actions. Eyeball the diff before proceeding.

## Updating CHANGELOG.md

Every release commit updates [`CHANGELOG.md`](../../CHANGELOG.md) atomically with the version bump:

1. Rename the existing `## [Unreleased]` heading to `## [X.Y.Z] - YYYY-MM-DD` (UTC date).
2. Open a new empty `## [Unreleased]` section above it (with empty `### Added`, `### Changed`, `### Deprecated`, `### Removed`, `### Fixed`, `### Security` sub-headings as you need them — don't include empty sections in the published version, but keep the template at the top for the next dev cycle).
3. Update the link references at the bottom:
   ```
   [Unreleased]: https://github.com/lin-snow/Ech0/compare/v4.6.5...HEAD
   [4.6.5]: https://github.com/lin-snow/Ech0/compare/v4.6.4...v4.6.5
   ```

CHANGELOG entries should be written for **users / operators**, not developers. Prefer:

- ✓ "Added: Per-user storage quotas configurable via `STORAGE_QUOTA_MB`."
- ✗ "Refactored storage middleware in `internal/middleware/staticfile.go` to support quota injection."

The latter belongs in commit messages, not the user-facing CHANGELOG.

## Committing and tagging

```bash
git commit -am 'chore(release): vX.Y.Z'
git tag -a vX.Y.Z -m 'Release vX.Y.Z'
```

- Commit message format is exactly `chore(release): vX.Y.Z` (lowercase `v`, no extra punctuation). This makes `git log --grep 'chore(release)'` a clean release ledger.
- Use **annotated** tags (`-a`), not lightweight ones — annotated tags carry author, date, and message, and are first-class objects.
- If you have a GPG or SSH signing key configured, use `git tag -s` instead. GitHub will display a "Verified" badge on signed tags, which lets downstream operators distinguish authentic releases from impersonations.

The tag must point at the same commit that contains the bump (i.e. tag immediately after committing). The CI guard refuses to build a release where the tag and `version.go` disagree.

## Pushing and triggering the release

```bash
git push origin main         # publish the chore(release) commit
git push origin vX.Y.Z       # push the tag separately to trigger release.yml
```

Pushing the tag is the act that commits to the release publicly:

- [`.github/workflows/release.yml`](../../.github/workflows/release.yml) fires on `tags: v*`.
- `verify-version` runs first; if `Version != tag`, the workflow fails fast.
- `build` then produces `linux/amd64` and `linux/arm64` static binaries with `Commit` and `BuildTime` ldflags-injected.
- `prepare-release` packages them as `tar.gz` and creates a **draft** GitHub release.
- `build-docker` pushes multi-arch images to GHCR and Docker Hub, tagged `vX.Y.Z` and `latest`.

After the workflow completes:

1. **Verify the artifacts.**
   ```bash
   gh release view vX.Y.Z              # check files, draft status
   gh run watch                        # if still running
   ```
   Download the linux/amd64 binary, run `./ech0 version`, and sanity-check that the version + commit hash match what you tagged.

2. **Promote the draft release to published.** GitHub release notes default to auto-generated PR titles since the last release; replace them with the relevant `[X.Y.Z]` section from `CHANGELOG.md`.

3. **Pull the Docker image** to confirm:
   ```bash
   docker pull ghcr.io/lin-snow/ech0:vX.Y.Z
   docker run --rm ghcr.io/lin-snow/ech0:vX.Y.Z version
   ```

## Hot-fix releases

If a critical bug is discovered post-release:

1. Branch off the affected tag: `git checkout -b fix/<short-name> vX.Y.Z`.
2. Apply the minimal fix; merge to `main` via PR.
3. From `main`, follow the standard procedure with a PATCH bump (`vX.Y.(Z+1)`).
4. Do **not** re-tag the original release — published tags are immutable.

## Pre-release / RC tags

For breaking changes you want validated before stable:

```bash
just bump 5.0.0-rc.1
# ... commit / tag / push as v5.0.0-rc.1
```

GitHub will mark the release as "pre-release" if the workflow detects the suffix (currently this is on by default in the `softprops/action-gh-release` step's `prerelease: false` flag — flip to `prerelease: true` for RC tags, or auto-detect by tag suffix in a future improvement).

After the RC bakes in real environments for a week or two, promote with a stable tag:

```bash
just bump 5.0.0
# ... commit / tag / push as v5.0.0
```

## Future improvements (not implemented yet)

These are options you might want as the project grows; they are deliberately not adopted today to keep the release surface simple for a single-maintainer project.

- **[release-please](https://github.com/googleapis/release-please)** — a GitHub Action that watches `feat:` / `fix:` / `chore(deps):` commits, decides the next version automatically, and opens a "Release vX.Y.Z" PR. Merging that PR auto-tags. Would replace `make bump` + manual CHANGELOG editing with a fully automated PR-based flow.
- **[git-cliff](https://github.com/orhun/git-cliff)** — auto-generates CHANGELOG from conventional commits without taking over the tag-creation step.
- **Reproducible builds** via `SOURCE_DATE_EPOCH=$(git log -1 --format=%ct)` instead of `date -u +%Y-%m-%dT%H:%M:%SZ` for `BuildTime`. Two builds of the same commit would then produce byte-identical binaries.
- **Signed tags by default** in CI (sigstore / cosign signatures on artifacts and images).
- **Auto-generated SBOM** alongside release artifacts.
