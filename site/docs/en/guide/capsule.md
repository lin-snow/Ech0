---
title: Capsules & static sites
description: Export your content as a readable capsule for moving instances, long-term preservation, or compiling into a static archive site
---

**A capsule is a folder holding all your public content** — human-readable, editable, and safe to drop into git. It solves three things a snapshot cannot: moving to a new instance, still being readable ten years from now, and publishing the archive as a site that needs no server.

---

## Capsules vs snapshots

Both can “back up”, but their goals are completely different.

|                  | Capsule                                                                      | Snapshot                                                   |
| ---------------- | ---------------------------------------------------------------------------- | ---------------------------------------------------------- |
| Holds            | Public content: Echoes, media, comments, site settings                       | The entire instance: database files, accounts, credentials |
| Looks like       | A readable directory, one Markdown file per Echo                             | An opaque ZIP                                              |
| Import semantics | **Idempotent merge**: skips by id, never overwrites existing content         | **Destructive replacement** of the whole database          |
| Good for         | Migration, long-term preservation, publishing a static site, switching tools | Disaster recovery and rollback                             |

In one sentence: to back up “the whole machine”, use a snapshot; to take “what I wrote” with you, use a capsule.

Neither requires a terminal anymore: exporting and importing capsules can also be clicked in the **web panel** — Panel → Data control, pick “Capsule” in the “Export” tab and “Ech0 capsule” in the “Import” tab. The CLI stays exactly as it was; both paths are equivalent, so use whichever is handier.

> **Never treat a capsule as a backup.** A capsule contains content only — no accounts, passwords, or credentials of any kind. On the day you actually need to rebuild an instance, it returns your posts and images, but not “who can sign in to this site” — users and sign-in methods exist only in snapshots. Use snapshots for regular backups; capsules are for moving and archiving.

---

## What it looks like

```text
capsule/
  ech0.yaml                            # site info, ownership, federation list
  echoes/2026/2026-08-02-b302099a.md   # one Echo = one frontmatter-markdown file
  comments.yaml                        # comment snapshot (public projection, no emails or IPs)
  files/images/…                       # media bytes
```

A single Echo is ordinary Markdown with a block of metadata in front:

```markdown
---
id: 0198f0a0-0000-7000-8000-000000000001
created_at: 2026-08-02T10:00:00Z
tags:
  - journal
---

The body goes here, saved verbatim.
```

Even if Ech0 disappears one day, these files open in any editor.

---

## Commands

These require a **binary installation** (Docker users can `docker exec` into the container). All commands run from the instance root directory — the one containing `data/`.

```bash
ech0 export capsule   [-o ./capsule] [--include-private] [--zip]
ech0 import capsule   [<path>=./capsule] [--include-private] [--dry-run]
ech0 check            [<path>=./capsule] [--fix]
ech0 build            [<path>=./capsule] [-o ./dist] [--base-url /]
```

Two of them have web counterparts; the other two run only from the CLI:

| Command          | What it does                       | Web entry point                                         |
| ---------------- | ---------------------------------- | ------------------------------------------------------- |
| `export capsule` | Export content as a capsule        | Yes: Data control → Export → Capsule                    |
| `import capsule` | Import a capsule into the instance | Yes: Data control → Import → Ech0 capsule               |
| `check`          | Validate a capsule                 | No, CLI only (web import validates automatically first) |
| `build`          | Compile into a static site         | No, CLI only                                            |

While we're at it — snapshots now have CLI entry points too, so you can put them straight into cron:

```bash
ech0 export snapshot  [-o ./snapshot.zip]
ech0 import snapshot  <snapshot.zip> --yes    # destructive, explicit confirmation required
```

---

## Export

```bash
ech0 export capsule -o ./my-capsule
```

- **Private Echoes are excluded by default.** To take them along, add `--include-private` — capsules are not encrypted, so don't casually push them to a public repository.
- Attachments stored on S3 are downloaded into the capsule, so a capsule is always self-contained.
- If any single attachment cannot be retrieved (bucket unreachable, object lost), the export **fails as a whole and lists what is missing**, rather than producing a capsule with broken images. Fix it and rerun; a failure leaves no half-finished output behind.
- `--zip` produces a single file whose internal structure is exactly the same as the directory form.

### Exporting from the web

Panel → Data control → Export, then choose “Capsule” on the format card — it defaults to “Snapshot”, deliberately. Once capsule is selected, an extra notice appears (capsules contain no accounts or credentials and cannot be used for disaster recovery) along with an “Include private content” toggle, off by default.

Click export and the job runs in the background, moving through **preparing → packaging → done**, and downloads automatically when it finishes. The result row states the format, filename, and size; if you miss the download, click “Download again” — it honors the format the job actually produced, so flipping the selector back to snapshot afterwards won't fetch the wrong file.

Capsules exported from the web are zips. If you want the directory form you can `cd` into, or want to schedule it, use the CLI.

---

## Validation

```bash
ech0 check ./my-capsule
ech0 check ./my-capsule --fix
```

There are two levels, **errors** and **warnings**: errors make import and build refuse to run, while warnings are only advisories. Common warnings:

- `embeds source instance URL` — the body hard-codes the original site's address, which may break links after a domain change.
- `dangling media` — the capsule contains the bytes for this attachment, but nothing declares it. Capsules you exported yourself won't hit this (attachments not attached to an Echo are recorded in the manifest); it essentially only shows up in hand-written capsules.

`--fix` repairs exactly one thing: a missing `id`. Very handy when hand-writing a capsule.

---

## Import

```bash
ech0 import capsule ./my-capsule --dry-run   # show the manifest first, write nothing
ech0 import capsule ./my-capsule
```

A few key points:

- **Safe to run repeatedly.** Duplicates are detected by id; existing Echoes are skipped, never overwritten or merged.
- **Stored verbatim.** Author name, like count, and body are preserved as-is, with no rewriting whatsoever.
- **Site settings only fill blanks.** Nothing you have already configured is touched; only settings never configured (still at defaults) take values from the capsule — so when you move to a new instance, the site title, logo, and ICP filing number come along.
- **No webhooks fired.** Importing historical content shouldn't push another round of notifications to subscribers. The trade-off is that the vector index doesn't follow automatically; after importing, trigger a reindex once from the admin panel.

### Importing from the web

Panel → Data control → Import, choose “Ech0 capsule” as the source, and upload the zip you exported. There is the same “Include private content” toggle, off by default. Progress goes **validate → import → done**.

The page reassures you with two points up front, and this is where capsule import differs most from “snapshot import”:

- **Append-only, idempotent, non-destructive.** Deduplication is by content id, so importing the same capsule ten times gives the same result as once; existing content is neither cleared nor overwritten. Snapshot import replaces the whole database; capsule import does not — many people conflate the two.
- **No accounts or credentials**, so users and passwords cannot be restored; imported content is attributed to the currently signed-in site owner.

Before writing, it **forces a validation pass** (the same work `check` does). Any error-level issue rejects the import outright and lists the first few errors, instead of stopping halfway through.

For scale: an instance with 287 Echoes and 170 files exports to 114 MB and imports into a brand-new instance in about 1.2 seconds, all of it landing. Import the same capsule again and the result is 0 added, 287 skipped.

### Moving to a new instance

```bash
# old instance
ech0 export capsule -o ./capsule --include-private

# copy capsule/ to the new machine, then on the new instance
ech0 check ./capsule
ech0 import capsule ./capsule --include-private
```

---

## Building a static site

```bash
ech0 build ./my-capsule -o ./dist
ech0 build ./my-capsule -o ./dist --base-url /blog/   # deploy under a subpath
```

The output is a directory you can drop straight into GitHub Pages, Cloudflare Pages, or object storage:

```text
dist/
  index.html  404.html  assets/…   # full Ech0 frontend, identical to your site's UI
  dataset.json                     # pre-baked content data
  api/files/…                      # attachments
  api/connect                      # Connect card, so other Ech0 instances can still reach you
  rss.xml  sitemap.xml
```

**No Node or pnpm needed** — the frontend is already embedded in the `ech0` binary.

A static site is a **frozen presentation**: like counts and comments render read-only as they were at export time, publish/reply/sign-in entry points hide themselves automatically, and liking shows a notice that this is a static archive. Interaction traces are part of the content's history; hiding them only makes an archive look “deader” than the original site.

> **Deployment note**: platforms such as GitHub Pages / Cloudflare Pages need the **SPA deep-link fallback** pointed at `404.html`, otherwise visiting `/echo/xxx` directly returns a 404.

### What to use it for

- Freeze a blog into a static archive: shut the server down while the content stays online forever.
- Keep a browsable time capsule from every export.
- Use free static hosting and leave the dynamic instance on your LAN.

---

## Hand-written and third-party capsules

The capsule format is a **public specification**: you can hand-write one, or write a converter that turns another tool's data into a capsule and import it. The minimum usable Echo needs only `id` and `created_at` — if `id` is missing, generate it with `ech0 check --fix`. Timestamps accept any valid RFC 3339 offset (`+08:00` is fine).

Apart from `id` and `created_at`, missing fields are treated as empty values; the importer will not reject the capsule because of them.

---

## FAQ

**Export fails saying a file cannot be retrieved?**
The S3 configuration changed or the object was deleted. Fix the storage config and rerun; if it truly cannot be recovered, clear the corresponding attachment record.

**Site title unchanged after import?**
That means the target instance already has a title configured — import does not overwrite what you have configured; change it in the admin panel.

**Static site images are all broken?**
Check that the `api/files/` directory is complete; when deploying under a subpath, remember to pass `--base-url`.

**Can a capsule be managed with git?**
Yes — that is one of the design goals. Each Echo is a separate file and comments live in a single file, so new comments never pollute the diff of content files.
