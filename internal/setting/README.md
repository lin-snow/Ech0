# internal/setting

Read primitives for settings: `Spec[T]` (Get / Set / Restore / Unset / DefaultValue), the registry (`registry.go`) that declares every settings block, and the one-shot `migrations` (`migrations.go`) that run before startup.

Not a service, not a repository: no auth, no business policy, no transactions, no side effects. It only turns stored JSON into a typed value with defaults, normalization, and the block's own structural validation.

## Rule: read direct, write through the domain

| Goal | Do | Need |
|---|---|---|
| Read a settings block | `coreSetting.System.Get(ctx, durableKV)` | `durableKV kvstore.Store` |
| Change settings | `SettingService` (`internal/service/setting`) | `settingService.Service` |
| Run boot migrations | nothing — `setting.Migrate` runs in `BeforeStart` | — |

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

- `Get`: no transaction. It starts from `Default()`, overlays the stored JSON on top (fields absent from the stored blob keep their default), and normalizes the result. A missing key is simply the default; an unparseable blob returns a fresh default plus the error.
- `Set`: joins a transaction only when the caller passes a tx ctx; this layer just forwards ctx. It normalizes, then runs `Validate` when one is declared, then writes.
- `Restore`: the import path. Normalizes and writes without `Validate`, because the source instance already ran with that configuration and a stricter rule on the target must not fail a valid snapshot. Capsule import and the migrator use it; nothing else should.
- `Unset`: deletes the block, so `Get` falls back to the defaults. Tolerates a missing key.
- `Migrate`: owned by the lifecycle (`BeforeStart`). It runs the one-shot `migrations` and writes nothing else — defaults are never materialized. A block appears in the KV table only once something writes it, so changing a default in a later release still reaches every instance that never overrode it.

Never seed defaults. Writing them on first boot looks harmless and quietly turns every future default change into a no-op for existing installs, and it makes "is this value the operator's choice or a stale default?" unanswerable.

## Validation

`Spec.Validate` guards writes, so every writer — `SettingService`, `capsule/importer`, any future importer — inherits the same invariant instead of re-checking it. It does not guard `Restore` (import) or the boot-time migrations.

Keep validators to structural invariants of the block itself (an enabled integration has the fields it needs to work). Business policy stays in `SettingService`; protocol capabilities stay with their owner (the agent API-key rule depends on the selected protocol and lives in `internal/agent`).

## Adding a setting

1. Add the field to the `model/setting` struct and DTO, set the default in `registry.go` `Default()`, copy it in `SettingService.UpdateSetting`.
2. `Default()` is not only a baseline: `Get` prefills it before decoding, so a blob that predates the field keeps the default instead of the zero value. Default-on switches can therefore be stored positively; if an older blob used an inverted field name, add a `migrations` entry that rewrites it (see `migrateAgentDiscoveryFlag`).
3. Behavior switches stay out of capsules: a capsule carries site content for rendering, never operator policy.

`DefaultValue()` returns that normalized default without touching the store — use it where only the baseline matters (imports diffing against defaults, the storage manager's nil-KV fallback).
