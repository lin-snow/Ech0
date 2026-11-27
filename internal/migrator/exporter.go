// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"context"
	"strings"

	"github.com/lin-snow/ech0/internal/migrator/spec"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	"github.com/lin-snow/ech0/internal/storage"
)

type ExportOutcome struct {
	ArtifactPath string `json:"-"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size,omitempty"`
	Format       string `json:"format,omitempty"`
}

type ExportEngine struct {
	registry *storage.Registry
}

func NewExportEngine(registry *storage.Registry) *ExportEngine {
	return &ExportEngine{registry: registry}
}

func (ex *ExportEngine) Export(
	ctx context.Context,
	report func(phase string, snapshot any),
) (ExportOutcome, error) {
	dest := migratorModel.ExportDestFS
	if ex.registry != nil {
		// The artifact already sits on local disk; only an object-storage
		// default needs a second copy under snapshots/.
		if id, _, err := ex.registry.Default(); err == nil && id != storage.DriverLocal {
			dest = migratorModel.ExportDestS3
		}
	}

	exporter, err := BuildExporter(dest, ex.registry)
	if err != nil {
		return ExportOutcome{}, err
	}

	result, err := exporter.Export(ctx, spec.ExportRequest{
		UpdateProgress: func(progress spec.ExportProgress) {
			if phase := strings.TrimSpace(progress.CurrentPhase); phase != "" {
				report(phase, nil)
			}
		},
	})
	if err != nil {
		return ExportOutcome{}, err
	}

	return ExportOutcome{
		ArtifactPath: result.ArtifactPath,
		FileName:     result.FileName,
		Size:         result.Size,
		Format:       migratorModel.ExportFormatSnapshot,
	}, nil
}
