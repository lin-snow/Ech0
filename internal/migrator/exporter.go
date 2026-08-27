// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"context"
	"strings"

	"github.com/lin-snow/ech0/internal/migrator/spec"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
)

type ExportOutcome struct {
	ArtifactPath string `json:"-"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size,omitempty"`
	Format       string `json:"format,omitempty"`
}

type ExportEngine struct {
	storageManager StorageManager
}

func NewExportEngine(storageManager StorageManager) *ExportEngine {
	return &ExportEngine{storageManager: storageManager}
}

func (ex *ExportEngine) Export(
	ctx context.Context,
	report func(phase string, snapshot any),
) (ExportOutcome, error) {
	dest := migratorModel.ExportDestFS
	if ex.storageManager != nil {
		if sel := ex.storageManager.GetSelector(); sel != nil && sel.ObjectEnabled() {
			dest = migratorModel.ExportDestS3
		}
	}

	exporter, err := BuildExporter(dest, ex.storageManager)
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
