// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

const (
	MigrationSourceEch0  = "ech0"
	MigrationSourceMemos = "memos"
)

const (
	ExportDestFS = "fs"
	ExportDestS3 = "s3"
)

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

const (
	ExportFormatSnapshot = "snapshot"
	ExportFormatCapsule  = "capsule"
)

const (
	ImportPhaseChecking  = "checking"
	ImportPhaseImporting = "importing"
)

const MigrationSourceCapsule = "capsule"
