# Logging in

Commands need an instance URL and a token. Log in once and the CLI reuses the
stored profile; override either per call with `--server` / `--token`.

## Interactive login (browser)

```bash
ech0 auth login --server https://ech0.example.com --device
```

The CLI prints a link and a device code, then waits until the browser approves.
The stored token is valid for 1 month. A human has to open the link, so do not
run this in a non-interactive context.

## Access token (no browser)

```bash
ech0 auth login --server https://ech0.example.com --token <token>
echo "$TOKEN" | ech0 auth login --server https://ech0.example.com --token -
```

`--token -` reads stdin, which keeps the secret out of shell history. The CLI
checks the token before storing it. For one-off commands you can skip login
entirely with `ECH0_TOKEN`.

## Password

```bash
ech0 auth login --server https://ech0.example.com -u admin
ech0 auth login --server https://ech0.example.com -u admin --password-stdin < password.txt
```

Without `--password-stdin` the password comes from a hidden prompt. Admin
accounts get a long-lived CLI token; other accounts keep a session that renews
automatically.

## Scopes

`--scopes a,b,c` applies when a login mints a token. The default is
`echo:read,echo:write,profile:read`. Tokens minted elsewhere keep their own
scopes.

## Profiles

```bash
ech0 config list            # all profiles, active one marked
ech0 config use <name>      # switch the default
ech0 config unset <name>    # remove the profile and its token
ech0 config path            # where config and credentials live
```

Which profile, server and token a command uses:

| What | Order |
| --- | --- |
| Profile | `--profile <name>` → `ECH0_PROFILE` → default profile |
| Server | `--server <url>` → `ECH0_SERVER` → the profile's server |
| Token | `--token <token>` → `ECH0_TOKEN` → the stored token |

## CI and containers

```bash
export ECH0_SERVER=https://ech0.example.com
export ECH0_TOKEN=<token>
ech0 echo page get --page 1 --json
```

Set `ECH0_NO_KEYCHAIN=1` to keep credentials in a file instead of the system
keychain (headless machines, locked keyrings).

## Checking and clearing

```bash
ech0 auth status            # profile, server, token source (no network)
ech0 auth status --verify   # also verifies the token with the server
ech0 auth logout            # delete the token for the current profile
ech0 auth logout --purge    # delete the token and the profile
```

## Errors

- `401` — token missing, invalid or expired: `ech0 auth status --verify`, then
  log in again.
- `403` — the token lacks a required scope, or the account is not an admin.
- No profile or server — the CLI exits 2 and names the missing piece.
