# Comments

The comment commands are `comment.create` (catalog id
`comment.create.integration`), `comment.list` (`comment.list.by.echo`),
`comment.public` (`comment.list.public`) and `comment.delete`
(`comment.panel.delete`). Reading is public; creating needs `comment:write`
and deleting is moderation and needs `comment:moderate`.

## Read

```bash
ech0 comment list --echo-id <echo_id> --json     # comments under one Echo
ech0 comment public --limit 20 --json            # latest public comments
```

`comment list` requires `--echo-id`. `comment public` takes an optional
`--limit` (server default 30). Both return an array of comments:
`{id, echo_id, parent_id, nickname, content, status, hot, created_at, ...}`.

## Create

```bash
ech0 comment create --echo-id <echo_id> --content "写得真好"
ech0 comment create --echo-id <echo_id> --content "来自流水线" --nickname "CI Bot" --metadata "run 42"
```

| Flag | Meaning |
| --- | --- |
| `--echo-id` | Echo the comment belongs to |
| `--content` | comment body, at most 200 characters |
| `--nickname` | display name; defaults to `Integration` |
| `--metadata` | optional operator/debug string, logged server-side |

`comment create` calls the trusted integration endpoint
`POST /api/comments/integration`, not the guest endpoint. It needs an access
token with `comment:write` (CLI logins request it by default) and is therefore
usable from scripts: there is no anti-bot form token, no captcha and no honeypot
field to fill in.

The stored comment is tagged `source: integration`, so it is distinguishable in
the admin panel. Its status follows the instance's comment approval setting: it
is approved immediately when approval is not required, otherwise it starts as
`pending`. The endpoint also suppresses a byte-identical repeat of the same
comment on the same Echo from the same author/IP within 90 seconds, and applies
its own rate limits (roughly 5 per minute and 30 per hour per IP). Hitting either
returns a business error, not a retryable 5xx.

The response is `{id, status}`.

## Delete

```bash
ech0 comment delete --id <comment_id> --yes
```

Moderation-only: the endpoint needs `comment:moderate` and the service requires
an admin account. Interactive terminals print the exact `DELETE` target and ask
`[y/N]`; pipes and `--json`/`--yaml` never prompt and require `--yes`.
