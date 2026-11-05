---
title: CLI
description: Post, moderate comments, and inspect your instance from the terminal
---

`ech0` is Ech0's command-line client, built for bulk operations, scripts, and server troubleshooting. Day-to-day use through the web UI works perfectly well too.

---

## Installation

```bash
npm install -g @useech0/cli
```

Supports macOS / Linux (arm64 / x64), ships its own runtime, and needs no separate Node or Go install. Windows is not supported yet. You can also download `ech0-cli-<os>-<arch>.tar.gz` from [GitHub Releases](https://github.com/lin-snow/Ech0/releases).

---

## Signing in

```bash
ech0 auth login --server https://your-domain.com --device
```

Open the link shown in the browser and confirm; the token is kept for 1 month by default. Credentials are stored per instance — switch between several instances with `ech0 config list` / `ech0 config use <name>`.

Other ways to sign in (scripts or CI):

```bash
ech0 auth login --server https://your-domain.com -u admin          # username + hidden password prompt
echo "$TOKEN" | ech0 auth login --server https://your-domain.com --token -
```

In CI you can also use the `ECH0_SERVER` and `ECH0_TOKEN` environment variables directly, with nothing written to disk.

---

## A first look at this instance

```bash
ech0 info
```

Shows at a glance: site name, version, number of Echoes, whether the instance is initialized, current sign-in state, and whether the command manifest is up to date. `--json` outputs for scripts.

---

## Common operations

```bash
# Browsing
ech0 echo list --page 1 --page-size 20
ech0 echo get --id <echo ID>

# Posting (upload images first, then reference the returned ID)
ech0 file upload --file photo.png
ech0 echo create --content "today's thoughts" --tag daily --file-id <file ID>

# Comment moderation
ech0 comment panel list --status pending
ech0 comment panel status --id <comment ID> --status approved

# Federation health
ech0 connect health
```

- Destructive commands ask for confirmation first; scripts or `--json` output must pass `--yes`.
- There is no delete command for files: a file may still be referenced by an Echo, so handle it in the web panel.

---

## Permissions

The default sign-in already covers the commands above. If you want a read-only token for a script, specify scopes when signing in:

```bash
ech0 auth login --server https://your-domain.com --device --scopes echo:read,file:read
```

A 403 means the token is missing a permission — sign in again for a new token. See [Access tokens](/docs/guide/accesstoken) for the complete list of permissions.

---

## Help and output

```bash
ech0 help                       # all commands
ech0 help echo                  # one group
ech0 help comment panel list    # fields and examples for a single command
```

`--json` prints the raw response, `--yaml` prints YAML. Exit codes: `0` success, `2` usage error, `1` server or network error.

New server endpoints do not require a CLI upgrade: the command manifest updates with the instance and refreshes automatically when stale.

---

## For AI assistants

```bash
ech0 install skills ~/.claude/skills
```

Installs Ech0's command documentation (including permission and safety conventions) into the skills directory of tools such as Claude Code.
