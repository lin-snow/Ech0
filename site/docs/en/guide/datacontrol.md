---
title: Data control
description: Snapshots, migration, backup schedules, and recovery
---

Ech0's data consists mainly of the **database (including business metadata)** and **attachment files** (a local directory or object storage). Day-to-day operations revolve around **export, migration, backup, and recovery**.

---

## What you can do in the web UI

Open **Admin panel → Data control** (the exact label depends on your version). It typically offers:

- **Snapshot export**: pack everything into a ZIP for offline archiving or moving to another machine.
- **Migration import**: import historical data from supported sources (see below).

Export file naming and storage paths follow what the interface tells you; if a “download link” or “background job” is offered, just follow the wizard.

---

## Command line and TUI

If you installed from a **binary**, you can run this on the server:

```bash
ech0 tui
```

Then perform backup-related operations from the menu (the exact items follow the TUI).

You can also use the command line directly, which suits scheduled packaging via **cron** or CI:

```bash
ech0 export snapshot -o ./snapshot.zip     # full database snapshot
ech0 import snapshot ./snapshot.zip --yes  # destructive restore, explicit confirmation required
```

If what you want is a **readable, portable, long-term** content backup rather than a whole-database image, see
[Capsules & static sites](/docs/guide/capsule) — `ech0 export capsule` produces a
Markdown directory tree that can also be compiled into a static archive site with a single command.

Refer to `ech0 --help` for the complete set of subcommands.

---

## Migration sources (common)

| Source                   | Notes                                                       |
| ------------------------ | ----------------------------------------------------------- |
| Another Ech0 v4 instance | Upload or select a package through the migration wizard     |
| Ech0 v3                  | **Export a snapshot** in v3 first, then run migration in v4 |
| Memos and others         | Import in the formats supported by the wizard               |

Migration usually **blocks concurrent writes** (a write lock) to prevent data corruption; when it finishes you can choose whether to merge some system settings (depending on the wizard).

> **v3 → v4**: in-place upgrades are not supported. You must: export a snapshot in v3 → deploy v4 → use migration import in v4.

---

## Scheduled automatic backups

Under **System settings → Backup schedule**:

1. Turn on the automatic backup switch.
2. Fill in a **Cron expression** (the common 5-field and 6-field formats are supported).

Example meanings (refer to the time zone you set and the documentation):

| Expression    | Meaning (illustrative)  |
| ------------- | ----------------------- |
| `0 2 * * 0`   | Every Sunday at 2:00 AM |
| `30 3 * * *`  | Every day at 3:30 AM    |
| `0 */6 * * *` | Every 6 hours           |

You can review execution results in the logs or the panel; it is advisable to **download backup files off-site regularly**.

---

## Manual recovery (high risk)

Only do this when you understand the consequences:

1. **Stop** the Ech0 service or container.
2. **Back up the current data directory** (copy the whole thing).
3. **Extract** a trusted snapshot ZIP and use its files to **overwrite** the database and files in the data directory (the paths must match the official directory structure).
4. Start the service and verify.

Overwriting by mistake or a version mismatch can prevent startup — always keep a backup you can roll back to first.

---

## Object storage and file paths

If **S3** is enabled, attachments actually live in the bucket; the database only stores references. Before switching buckets, changing domains, or migrating between local and object storage, read [Object storage (S3)](/docs/guide/s3), and validate in a test environment before going to production.

---

## OpenAPI

The HTTP endpoints related to backup and export (if exposed) are documented at your instance's **`/swagger/index.html`**; calls usually require an **access token** or an admin session. Never put tokens in public scripts.
