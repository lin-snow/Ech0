// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import "github.com/lin-snow/ech0/internal/migrator/spec"

type (
	ImportRequest  = spec.ImportRequest
	ImportProgress = spec.ImportProgress
	ImportResult   = spec.ImportResult
	FailedItem     = spec.FailedItem
	Importer       = spec.Importer

	ExportRequest  = spec.ExportRequest
	ExportProgress = spec.ExportProgress
	ExportResult   = spec.ExportResult
	Exporter       = spec.Exporter
)
