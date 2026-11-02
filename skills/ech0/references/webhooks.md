# Webhooks

Webhooks push Ech0 events to an HTTP(S) URL. Every webhook command requires an
admin account; the API scope is `admin:settings` (there is no finer webhook
scope yet).

## List

```bash
ech0 webhook list --json
```

Each entry has `id`, `name`, `url`, `is_active`, `last_status`, `last_trigger`,
`created_at` and `updated_at`. The signing `secret` is never returned.

## Create

```bash
ech0 webhook create --name ci --url https://example.com/hook --is-active
ech0 webhook create --name ci --url https://example.com/hook --secret <secret> --is-active
```

`--name` and `--url` are required. The URL must be reachable over HTTP(S).
`--secret` enables the `X-Ech0-Signature: sha256=<hex>` header. Omitting
`--is-active` on create leaves the server default (active).

## Update

```bash
ech0 webhook update --id <webhook_id> --name ci --url https://example.com/hook --is-active
```

Update is a full replace: fetch `webhook list` first. **Omitting `--is-active`
deactivates the webhook; omitting `--secret` clears the signing secret.**

## Test and delete

```bash
ech0 webhook test --id <webhook_id>       # send one test event
ech0 webhook delete --id <webhook_id> --yes
```

`webhook test` reports delivery failure as an error. Deleting only removes the
local webhook configuration.
