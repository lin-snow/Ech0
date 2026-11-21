# release/

This directory holds **source templates** used to generate the public-facing
[GitHub Pages release site](https://lin-snow.github.io/Ech0/).  
It is **not** where compiled binaries or packaged artifacts are stored.

## Contents

| File | Purpose |
|---|---|
| `index.html.tmpl` | HTML template for the release landing page. Contains a `<!-- RELEASES -->` placeholder that is populated at build time with per-version download links. |
| `README.md` | This file. |

## How It Works

The generation is fully automated inside the
[`release_helm.yml`](../.github/workflows/release_helm.yml) workflow, which
runs on every published GitHub Release (and can also be triggered manually).

The pipeline performs the following steps (in order):

1. **Sync chart metadata** – Sets `version` and `appVersion` in `charts/ech0/Chart.yaml`
   from the release tag (or manual workflow input), then validates. Ensures each
   app release gets a new `ech0-<semver>.tgz`.
2. **Chart Releaser** – Packages `charts/ech0/` and updates the Helm repository
   on `gh-pages` (`index.yaml` and chart `.tgz` artifacts).
3. **Generate the release landing page** – Uses `gh release list` for published
   releases; a Python script filters to `v4.4.0+` and builds `<li>` rows with
   tarball links and release notes links. `awk` injects them into
   `index.html.tmpl` at `<!-- RELEASES -->`. Logo (`docs/imgs/logo.svg`),
   favicon (`web/public/favicon.svg`), and `.nojekyll` are copied into `_site/`.
4. **Deploy static page to `gh-pages`** – Commits `index.html`, `logo.svg`,
   `favicon.svg`, and `.nojekyll` alongside the Helm artifacts. GitHub Pages
   serves the site at `https://lin-snow.github.io/Ech0/`.

```
release/index.html.tmpl          ─┐
                                   ├─ awk replace ──► _site/index.html ──► gh-pages branch
GitHub Releases API (v4.4.0+)    ─┘                                         │
docs/imgs/logo.svg, favicon      ─────────────────────────────────────────── ├─ index.html, logo.svg, favicon.svg, .nojekyll
charts/ech0/ ─── Chart Releaser ─────────────────────────────────────────── ├─ index.yaml, ech0-*.tgz
```

## Relationship to Other Directories

| Directory / File | Role |
|---|---|
| `charts/ech0/` | Helm chart source. Packaged into `.tgz` and indexed on `gh-pages`. |
| `.github/workflows/release.yml` | Builds binaries, creates the GitHub Release (draft), and pushes Docker images. |
| `.github/workflows/release_helm.yml` | Publishes the Helm chart **and** regenerates the release page from this template. |
| `.github/workflows/release_zigcc.yml` | Experimental Zig-cc cross-compilation builds (artifacts only, not published to releases yet). |

## Editing the Release Page

To change the layout, styles, or static content of the public release page,
edit `index.html.tmpl` directly. The `<!-- RELEASES -->` comment must remain
intact — the CI pipeline uses it as the injection point for dynamic release
entries. Keep Docker / Compose install snippets aligned with the repository
root `README.md` (including `TZ`). After merging your changes,
the next published release (or a manual workflow dispatch) will regenerate and
deploy the updated page.
