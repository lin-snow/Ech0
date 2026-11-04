# Comments

The comment commands are `comment.create` (catalog id
`comment.create.integration`) and `comment.list` (`comment.list.by.echo`).
Reading is public; creating needs `comment:write`. The CLI deliberately has no
moderation commands, so it never asks for `comment:review` or `comment:delete`.

## Read

```bash
ech0 comment list --echo-id <echo_id> --json     # approved comments under one Echo
```

`comment list` requires `--echo-id` and returns an array of comments:
`{id, echo_id, parent_id, nickname, content, status, hot, created_at, ...}`.
Only `approved` comments are returned. Pending and rejected comments, and the
approve/reject actions, live in the web admin panel.

## Create

```bash
ech0 comment create --echo-id <echo_id> --content "写得真好"
ech0 comment create --echo-id <echo_id> --content "来自流水线" --nickname "CI Bot" --metadata "run 42"
```

| Flag | Meaning |
| --- | --- |
| `--echo-id` | Echo the comment belongs to (required) |
| `--content` | comment body, at most 200 characters (required) |
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
`pending` and must be approved in the web panel — the CLI has no review command.
The endpoint also suppresses a byte-identical repeat of the same comment on the
same Echo from the same author/IP within 90 seconds, and applies its own rate
limits (roughly 5 per minute and 30 per hour per IP). Hitting either returns a
business error, not a retryable 5xx.

The response is `{id, status}`.
