// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
)

type MigrationImporter interface {
	Import(ctx context.Context, payload migratorModel.MigrationPayload, report func(phase string, snapshot any)) (any, error)
}

var (
	_ MigrationImporter = (*coreMigrator.ImportEngine)(nil)
	_ MigrationImporter = (*coreMigrator.CapsuleEngine)(nil)
)

type MigrationRunner struct {
	importer        MigrationImporter
	capsuleImporter MigrationImporter
}

func NewMigrationRunner(importer *coreMigrator.ImportEngine, capsuleImporter *coreMigrator.CapsuleEngine) *MigrationRunner {
	return &MigrationRunner{importer: importer, capsuleImporter: capsuleImporter}
}

func (r *MigrationRunner) Run(ctx context.Context, p migratorModel.MigrationPayload, report job.ReportFunc) (any, error) {
	if p.SourceType == migratorModel.MigrationSourceCapsule {
		return r.capsuleImporter.Import(ctx, p, report)
	}
	return r.importer.Import(ctx, p, report)
}
