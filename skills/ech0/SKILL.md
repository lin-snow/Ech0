---
name: ech0
description: 'Use the ech0 CLI to inspect an Ech0 instance, browse Echoes and files, upload media, create and edit posts with tags, files and extensions, publish and moderate comments, manage Connect federation links and peer health, tags and webhooks, and log in with profiles or device authorization. Use when the task involves ech0, an Ech0 instance, or its timeline.'
---

# ech0

`ech0` is the command-line client for an Ech0 instance (a self-hosted microblog).
Run `ech0 help --json` to see what this instance supports, and look commands up
instead of assuming flags.

Add `--json` whenever you parse output. Exit codes: `0` success, `2` usage error,
`1` server or network error. Destructive commands (`echo delete`, `tag delete`,
`connect delete`, `comment panel delete`, `webhook delete`) never prompt in pipes
or with `--json`; they require `--yes`, so confirm with the human first.

## Before you start

```bash
ech0 info                   # server, version, site, counts, token state, catalog
ech0 auth status            # which profile and server are active
ech0 auth login --server https://ech0.example.com --device
```

Login modes, profiles and CI: [references/auth.md](references/auth.md).

## Common tasks

```bash
# Read the timeline, one Echo, or a curated view
ech0 echo list --page 1 --page-size 20 --json
ech0 echo list --search daily --sort-by fav_count --json
ech0 echo list --tag-ids <tag_id> --no-private --json   # public only
ech0 echo get --id <echo_id> --json
ech0 echo hot --limit 10 --json
ech0 echo today --json
ech0 echo random --json
ech0 echo onthisday --json

# Publish text
ech0 echo create --content "今天的想法" --tag daily

# Publish with a file: upload first, then reference the file id
ech0 file upload --file ./photo.png --json
ech0 echo create --content "配图" --file-id <file_id>

# Browse files, inspect metadata, download content
ech0 file list --search photo --json
ech0 file get --id <file_id> --json
ech0 file download --id <file_id> --output ./photo.png

# Register media that already lives elsewhere (no bytes uploaded)
ech0 file external --url https://example.com/photo.png --category image --json

# Edit: fetch it, change it, send it back
ech0 echo get --id <echo_id> --json
ech0 echo update --id <echo_id> --content "改过的正文" --tag daily

# Delete (ask the human first; non-interactive runs must add --yes)
ech0 echo delete --id <echo_id> --yes

# Read, write and moderate comments
ech0 comment list --echo-id <echo_id> --json
ech0 comment create --echo-id <echo_id> --content "写得好" --nickname "CI Bot"
ech0 comment panel list --status pending --json          # needs comment:review
ech0 comment panel status --id <comment_id> --status approved
ech0 comment panel delete --id <comment_id> --yes        # needs comment:delete

# Inspect and manage Connect federation links
ech0 connect list --json
ech0 connect info --json
ech0 connect health --json
ech0 connect add --connect-url https://peer.example
ech0 connect delete --id <uuid> --yes

# Tags: list ids for --tag-ids, or clean up a tag everywhere
ech0 tag list --json
ech0 echo list --tag-ids <tag_id>,<tag_id> --json
ech0 tag delete --id <tag_id> --yes

# Webhooks: list, create, test or remove automation targets
ech0 webhook list --json
ech0 webhook create --name ci --url https://example.com/hook --is-active
ech0 webhook test --id <webhook_id>
ech0 webhook delete --id <webhook_id> --yes
```

Echo create/update/delete, tag delete, Connect add/delete, comment deletion and
every webhook command require an admin account. Field
meanings, attachment rules and merge behaviour:
[references/echo.md](references/echo.md). Comment creation uses the trusted
integration route (`comment:write`); moderation needs `comment:review`, deletion
needs `comment:delete`: [references/comments.md](references/comments.md). Peer
links and health: [references/connect.md](references/connect.md). Tags and
webhooks: [references/tags.md](references/tags.md),
[references/webhooks.md](references/webhooks.md).

## Finding the right command

- `ech0 help --json` — everything you can run, local and server-side.
- `ech0 help <namespace> --json` — commands in one group (`ech0 help echo`).
- `ech0 help <words> --json` — search by name or field (`ech0 help upload`).
- `ech0 help <id> --json` — one command in full: fields, examples, hints.
  `<command> --help` prints the same.

## Output and errors

`--json` prints the response body; errors go to stderr with a hint. Streams print
one JSON line per event. Binary results go to a file with `--output <path>`.

Flags, arrays, payloads and error handling:
[references/operations.md](references/operations.md).
