# Migrator 与 Snapshot 设计

## 目标

把数据的「进」与「出」统一收敛到一个 **Migrator** 域，围绕一个统一资源 **Snapshot** 展开，
不再保留独立的 “backup” 概念。词汇只有三个：**Importer（导入）/ Exporter（导出）/ Snapshot（快照）**。

## 核心模型

- **Snapshot = `data/` 的 zip 归档**（排除 `files/snapshots`、`files/tmp`）。它既是导出的产物，
  也是导入的源（导出的快照可经 `ech0` 源往返导入）。资源代码在
  `internal/migrator/snapshot`（`writer.go` 打包 / `Unpack` 解包 / `reader.go` 定位 / `s3.go` 上传与保留），
  无 DI 依赖。
- **数据库一致性副本**：在线导出（fs/s3 exporter）必须以 `snapshot.WithConsistentDB(database.SnapshotTo)`
  调用 `Create`——用 `VACUUM INTO` 产出一致性时点副本打入 zip，并排除运行中的 `ech0.db` 及
  `-wal`/`-shm`/`-journal` 伴生文件（带并发写入直接拷实时库文件可能撕裂）。副本写不出来时导出必须
  失败而非静默回退。不带该选项的 `Create`（冷目录打包）保持原样带走全部文件。
- **对称的适配器族**（`internal/migrator`）：
  - `spec.Importer{Import}` 与 `spec.Exporter{Export}` 两个对称接口；
  - 导入按「来源」：`importer/ech0`、`importer/memos`（占位）；
  - 导出按「目的地」：`exporter/fs`（落本地目录）、`exporter/s3`（产出后上传 S3，仍留本地）；
  - `factory.BuildImporter(source)` / `factory.BuildExporter(dest, storageManager)` 对称选择适配器。
- **编排体 `ImportEngine` / `ExportEngine`**（`importer.go` / `exporter.go`）：选适配器 → 运行 →
  导入侧再应用配置/失效缓存/清 tmp;导出侧按是否配了对象存储选 fs/s3。只接受裸 report 回调,
  不依赖 `job.Manager`,故 `runner → migrator(核心)`、`job.Manager → runner` 全程无构造环。
- **`internal/service/migrator` 是薄转发层**：auth + 作业生命周期（提交/查询/取消）+ DTO 映射 +
  上传编排。

## 导出的三个触发出口

| 出口 | 路径 / 入口 | 机制 |
|------|------------|------|
| 手动快照 | `POST /migration/export`、`GET /migration/export/status`、`POST /migration/export/cancel` | `job.Manager`（`TypeExport`，持久化 / 可取消 / 状态轮询）→ `ExportEngine` |
| 定时快照 | `internal/task/scheduled`（cron） | 直接同步调 `ExportEngine`（不走 job，避免与手动导出抢占单行） |
| 下载 | `GET /migration/export/download` | 同步取回「最新已产出的快照」（`snapshot.LatestPath`）并流式下发，不再现打包 |

导入/导出均为 web 形态（管理后台「数据管理」三 tab:导入 / 导出 / 快照），无 CLI 命令。
事件：手动 / 定时发 `system.snapshot`；下载发 `system.export`。

## 破坏性变更（升级须知）

本次「彻底清除 backup 语义」涉及对外/持久化契约的改名，升级时注意：

- **磁盘**：快照目录 `data/files/backups/` → `data/files/snapshots/`，文件名 `ech0_backup_*.zip` →
  `ech0_snapshot_*.zip`。旧的 `data/files/backups/` 不再被识别（可手动删除）。
- **S3**：对象前缀 `backups/` → `snapshots/`。旧对象成为孤儿（可手动清理）。
- **Webhook 事件主题**：`system.backup` → `system.snapshot`，`system.backup_schedule.updated` →
  `system.snapshot_schedule.updated`。订阅旧主题的 webhook 需更新。
- **HTTP 路由**：
  - `GET /backup/export` → `GET /migration/export/download`
  - `POST /backup/snapshot` + `GET /backup/snapshot/:taskId`（旧 sync.Map 任务）→ `POST /migration/export` + `GET /migration/export/status`（job 驱动）
  - `GET/POST /backup/schedule` → `GET/POST /snapshot/schedule`
- **设置键**：定时计划持久化键 `backup_schedule` → `snapshot_schedule`（旧计划重置，需重新配置）。
- **CLI**：移除 `ech0 backup` 命令(快照导入/导出改为纯 web 形态,无 CLI 入口)。

> 注：日志轮转的 `MaxBackups` / `ECH0_LOG_FILE_MAX_BACKUPS` 属于 lumberjack 日志域，与快照无关，保持不变。
