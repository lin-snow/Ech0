---
name: ech0
description: 'Use the ech0 CLI to read and publish on an Ech0 timeline: browse Echoes and files, upload media, create and edit posts with tags, files and extensions, and log in with profiles or device authorization. Use when the task involves ech0, an Ech0 instance, or its timeline.'
---

# ech0

`ech0` is the command-line client for an Ech0 instance (a self-hosted microblog).
Run `ech0 help --json` to see what this instance supports, and look commands up
instead of assuming flags.

Add `--json` whenever you parse output. Exit codes: `0` success, `2` usage error,
`1` server or network error. Destructive commands never prompt, so confirm with
the human first.

## Before you start

```bash
ech0 auth status            # which profile and server are active
ech0 auth login --server https://ech0.example.com --device
```

Login modes, profiles and CI: [references/auth.md](references/auth.md).

## Common tasks

```bash
# Read the timeline, or one Echo
ech0 echo list --page 1 --page-size 20 --json
ech0 echo get --id <echo_id> --json

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
```

Create, update and delete require an admin account. Field meanings, attachment
rules and merge behaviour: [references/echo.md](references/echo.md).

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
