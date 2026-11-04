# Comments

The comment surface covers publishing, reading and review: `comment.create`
(catalog id `comment.create.integration`), `comment.list` (`comment.list.by.echo`),
and the review trio `comment.panel.list` / `comment.panel.status` /
`comment.panel.delete`. Reading is public, creating needs `comment:write`,
reviewing needs `comment:review`, and deleting needs `comment:delete`.

## Read

```bash
ech0 comment list --echo-id <echo_id> --json     # approved comments under one Echo
```

`comment list` requires `--echo-id` and returns an array of comments:
`{id, echo_id, parent_id, nickname, content, status, hot, created_at, ...}`.
Only `approved` comments are returned. Pending and rejected comments are visible
through `comment panel list` (see Review).

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
`pending` and can be approved from the terminal (see Review).
The endpoint also suppresses a byte-identical repeat of the same comment on the
same Echo from the same author/IP within 90 seconds, and applies its own rate
limits (roughly 5 per minute and 30 per hour per IP). Hitting either returns a
business error, not a retryable 5xx.

The response is `{id, status}`.

## Review

```bash
ech0 comment panel list --status pending --json
ech0 comment panel list --keyword 广告 --page 1 --page-size 20 --json
ech0 comment panel status --id <comment_id> --status approved
ech0 comment panel status --id <comment_id> --status rejected
ech0 comment panel delete --id <comment_id> --yes
```

`comment panel list` requires `comment:review` and returns the panel page as
`{items: [...], total}`. Filters: `--status` (`pending` / `approved` /
`rejected`), `--keyword`, `--echo-id`, `--hot true|false`, `--page`,
`--page-size`.

`comment panel status` is a `PATCH /api/panel/comments/{id}/status`; it changes
only the status and never the body. Use `--status pending` to pull an approved
comment back into the queue. Approving makes the comment visible in
`comment list`.

`comment panel delete` needs `comment:delete` — a scope separate from review —
and deletes the comment, not the Echo it belongs to. It is destructive: pipes and
`--json`/`--yaml` require `--yes`.

Not exposed: panel batch actions, hot-pinning, comment settings and the
test-email endpoint. Moderation is one comment at a time.
