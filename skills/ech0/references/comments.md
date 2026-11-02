# Comments

The comment commands are `comment.create`, `comment.list` (catalog id
`comment.list.by.echo`), `comment.public` (`comment.list.public`) and
`comment.delete` (`comment.panel.delete`). Creating and reading are public
operations; deleting is moderation and needs `comment:moderate`.

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
ech0 comment create --echo-id <echo_id> --content "写得真好" --form-token <token>
ech0 comment create --echo-id <echo_id> --content "回复" --parent-id <comment_id> --form-token <token>
```

| Flag | Meaning |
| --- | --- |
| `--echo-id` | Echo the comment belongs to (required by the server) |
| `--content` | comment body, at most 200 characters |
| `--nickname` / `--email` / `--website` | author fields; nickname and email are required when the token is not an admin/owner |
| `--parent-id` | reply to an existing comment under the same Echo |
| `--form-token` | anti-bot token issued by `GET /api/comments/form` |
| `--captcha-token` | captcha answer when the instance enables captcha |
| `--hp-field` | honeypot; leave empty |

The server rejects a missing or stale `--form-token`, and requires a short
delay between issuing it and submitting. Ask the instance for one, then pass it:

```bash
curl -s https://ech0.example.com/api/comments/form | jq -r .data.form_token
```

The token is bound to the client IP, so request it from the machine running
`ech0`. When captcha is enabled, also pass `--captcha-token`. The response is
`{id, status}`; an admin/owner token posts as the account and is approved
immediately, other authors start as `pending` when the instance requires
approval.

## Delete

```bash
ech0 comment delete --id <comment_id> --yes
```

Moderation-only: the endpoint needs `comment:moderate` and the service requires
an admin account. Interactive terminals print the exact `DELETE` target and ask
`[y/N]`; pipes and `--json`/`--yaml` never prompt and require `--yes`.
