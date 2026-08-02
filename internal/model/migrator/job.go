// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// 导入来源类型(对应 importer/ 下的适配器)。
const (
	MigrationSourceEch0  = "ech0"
	MigrationSourceMemos = "memos"
)

// 导出目的地类型(对应 exporter/ 下的适配器):fs=本地目录,s3=对象存储。
const (
	ExportDestFS = "fs"
	ExportDestS3 = "s3"
)

// 导出阶段(异步导出作业用,与 MigrationPhase* 对称)。S3 上传移出关键路径后台跑,不再是作业阶段,
// 故只有 packing/completed 两段:本地产物一落盘即 completed,下载立即可用。
const (
	ExportPhasePacking   = "packing"
	ExportPhaseCompleted = "completed"
)

const (
	MigrationStatusIdle      = "idle"
	MigrationStatusPending   = "pending"
	MigrationStatusRunning   = "running"
	MigrationStatusSuccess   = "success"
	MigrationStatusFailed    = "failed"
	MigrationStatusCancelled = "cancelled"
)

const (
	MigrationPhaseExtracting   = "extracting"
	MigrationPhaseTransforming = "transforming"
	MigrationPhaseValidating   = "validating"
	MigrationPhaseLoading      = "loading"
	MigrationPhaseReporting    = "reporting"
	MigrationPhaseCompleted    = "completed"
)

// 导出格式。快照是「data/ 的 zip」(含账号与凭据,可灾难恢复),胶囊是可分享的内容子集
// (不含任何凭据)。两者产物落在不同槽位,详见 internal/migrator/artifact。
const (
	ExportFormatSnapshot = "snapshot"
	ExportFormatCapsule  = "capsule"
)

// 胶囊导入阶段。与 MigrationPhase* 分开是因为胶囊走的是另一条链路:校验是硬前置
// (spec §7),没有解压与转换两步。
const (
	ImportPhaseChecking  = "checking"
	ImportPhaseImporting = "importing"
)

// MigrationSourceCapsule 是胶囊导入的来源标识(与 ech0/memos 并列,走 source_type 分派)。
const MigrationSourceCapsule = "capsule"
