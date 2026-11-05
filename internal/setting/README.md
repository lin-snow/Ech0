# internal/setting

Read primitives for settings: `Spec[T]` (Get / Set / Seed) plus the registry (`registry.go`) that declares every settings block.

Not a service, not a repository: no auth, no business validation, no transactions, no side effects. It only turns stored JSON into a typed value with defaults and normalization.

## Rule: read direct, write through the domain

| Goal | Do | Need |
|---|---|---|
| Read a settings block | `coreSetting.System.Get(ctx, durableKV)` | `durableKV kvstore.Store` |
| Change settings | `SettingService` (`internal/service/setting`) | `settingService.Service` |
| Seed defaults | nothing — `setting.Seed` runs in `BeforeStart` | — |

Never inject `SettingService` just to read:

```go
// Wrong: drags the whole settings domain (storage manager, webhook sender,
// token revoker, bus) into the graph for one read, and often closes a Wire
// construction cycle.
type Handler struct{ settingService settingService.Service }
system, err := h.settingService.GetSetting(&system)

// Right: hold the KV store and read directly.
type Handler struct{ durableKV kvstore.Store }
system, err := coreSetting.System.Get(ctx, h.durableKV)
```

`Get` is pure, idempotent and has no atomicity needs. Crossing domains through `SettingService` readily forms `SettingService → X → SettingService` cycles — see the comment in `internal/task/scheduled/snapshot.go`.

Direct readers today: `storage/manager`, `task/scheduled`, `event/subscriber/agent`, `capsule/{export,importer}`, `service/{auth,comment,common,connect,embedding,user}`, `handler/common` (`/llms.txt`, AI catalog), `mcp` (`/mcp/server-card`).

## Writes stay in the domain

Changing settings is rarely a bare write: admin authz, read-modify-write atomicity, and side effects (reconfigure storage, revoke tokens, dispatch webhooks, bust caches). Keeping those in `SettingService` makes the settings domain the only writer.

So `coreSetting.Set` lives in `internal/service/setting/*`, with one exception: `capsule/importer` applies site settings inside its own import transaction and exposes no write entry point. Do not copy it unless you are writing another import/migration executor.

## ctx and transactions

- `Get`: no transaction; a missing key falls back to `Spec.Default()`, then `Normalize` runs.
- `Set`: joins a transaction only when the caller passes a tx ctx; this layer just forwards ctx.
- `Seed`: owned by the lifecycle (`BeforeStart`), idempotent, never overwrites user values.

## Adding a setting

1. Add the field to the `model/setting` struct and DTO, set the default in `registry.go` `Default()`, copy it in `SettingService.UpdateSetting`.
2. `Default()` applies at seed time only. An existing `system_settings` blob has no such key, so it decodes to the zero value, and `Normalize` cannot distinguish "absent" from "turned off". Therefore default-**on** behavior switches must be stored inverted (`disable_agent_discovery`, zero = on; see `TestSystem_AgentDiscoveryDefaultsOnForLegacyBlobs`), while default-off switches may use the positive field.
3. Behavior switches stay out of capsules: a capsule carries site content for rendering, never operator policy.
