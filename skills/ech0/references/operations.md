# Running commands

Every command is described by `ech0 help <id> --json`: it lists the fields, where
they go, the result kind and the examples. Read it before guessing a flag.

## Global flags

Accepted anywhere (before or after the command):

```text
--profile <name>   --server <url>   --token <token>   --lang <locale>
--json             --yaml           --no-color
```

`--json` and `--yaml` are mutually exclusive. Prefer `--json`: it prints the
response body as received, while `--yaml` re-encodes the parsed document (very
large integers can lose precision).

## Field flags

- Names are kebab-cased: `page_size` → `--page-size`, `created_at` →
  `--created-at`.
- Booleans take no value: `--private`.
- Integers and numbers are parsed; an invalid number is a usage error.
- Arrays of strings accept a comma shorthand; arrays of objects **require JSON**:
  `--tags '[{"name":"daily"}]'`. A comma shorthand for object elements exits 2
  before any request is made.
- Objects take JSON: `--extension '{"type":"weather","payload":{}}'`.
- Uploads take a path, or `-` for stdin; add `--file-name <name>` so the
  extension (and therefore the file category) is known.
- Allowed values (`可选：…`) are listed in help but not enforced locally.

`echo create` and `echo update` also accept `--from <file>`, `--edit`,
`--tag <name>` and `--file-id <id>`; see [`echo.md`](echo.md).

## Results

JSON commands print the response body — pretty by default, as received with
`--json`.

- **Streams**: with `--json`, one JSON object per event on stdout; without it, a
  human rendering. `Ctrl-C` stops. `--yaml` is rejected.
- **Binary**: `--output <path>` writes the file, `-` means stdout (the default
  when stdout is not a terminal). `--json` requires `--output`; `--yaml` is
  rejected.
- **WebSocket**, when the server offers such commands: subscribes and prints
  events until `Ctrl-C`.

## Errors and exit codes

| Exit | Meaning |
| --- | --- |
| `0` | success (bare namespace help also exits 0) |
| `2` | usage: unknown command, bad flag value, missing required field, missing profile or token |
| `1` | server or network failure |

Errors go to stderr. `401` means log in again; `403` means the token lacks the
required scope, or the account is not an admin; no status means the instance is
unreachable — check `--server`/profile and the network. If the server's command
list changed, the CLI refreshes it and retries once.

## Keeping commands up to date

`ech0 catalog refresh` fetches the command list again; `ech0 catalog path` prints
where it is cached. Without a reachable instance, local commands and the cached
commands still work, and `ech0 help` explains what is missing.
