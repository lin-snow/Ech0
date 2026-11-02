# Connect (instance federation)

The Connect commands are `connect.status` (catalog id `connect.self`),
`connect.list`, `connect.add`, `connect.health` and `connect.remove`
(`connect.delete`). Status and list are public; health needs `connect:read`;
add and remove need `connect:write`.

## Inspect

```bash
ech0 connect status --json    # this instance's public card
ech0 connect list --json      # connections added to this instance
ech0 connect health --json    # reachability and version per connection
```

`status` returns `{server_name, server_url, logo, total_echos, today_echos,
sys_username, version}`. `list` returns `{id, connect_url}` records. `health`
returns `{id, connect_url, status, version}` per record, where `status` reports
whether the peer answered.

## Add and remove

```bash
ech0 connect add --connect-url https://peer.example --json
ech0 connect remove --id <uuid> --yes
```

`add` takes the peer instance URL; the server allocates the record id (omit
`--id`). `remove` takes the connection UUID from `connect list` or
`connect health`.

Remove is destructive: interactive terminals print the exact `DELETE` target
and ask `[y/N]`; pipes and `--json`/`--yaml` never prompt and require `--yes`.
Verify the id first — removing a connection only drops the local record, the
peer instance is unaffected.
