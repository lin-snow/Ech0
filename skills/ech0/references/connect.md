# Connect (instance federation)

The Connect commands are `connect.list`, `connect.info`, `connect.health`,
`connect.add` and `connect.delete`. List and info are public; health needs
`connect:read`; add needs `connect:write`, delete needs `connect:delete`, and
both require an admin account. The instance self card is not a Connect command:
`ech0 info` shows it, together with site, version, counts and catalog state.

## Inspect

```bash
ech0 connect list --json      # saved peers: {id, connect_url}
ech0 connect info --json      # peer cards: name / logo / post stats / version
ech0 connect health --json    # per-peer {id, connect_url, status, version}
```

`list` returns `{id, connect_url}` records. Use the `id` with `connect delete`.
`info` returns one public card per peer (`server_name`, `server_url`, `logo`,
`total_echos`, `today_echos`, `sys_username`, `version`). The server caches these
cards for 30 minutes; the first call after expiry fans out to every peer and may
be slow. `health` probes every saved peer on each call and needs `connect:read`.

## Add and delete

```bash
ech0 connect add --connect-url https://peer.example --json
ech0 connect delete --id <uuid> --yes
```

`add` takes the peer instance root URL; the server appends `/api/connect` and
allocates the record id. Duplicate URLs are rejected. `delete` takes the
connection UUID from `connect list`.

Delete is destructive: interactive terminals print the exact `DELETE` target
and ask `[y/N]`; pipes and `--json`/`--yaml` never prompt and require `--yes`.
Verify the id first — deleting a connection only drops the local record, the
peer instance is unaffected.
