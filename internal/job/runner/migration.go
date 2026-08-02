// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
)

// MigrationImporter 是迁移导入执行端，便于测试解耦（由 migrator.ImportEngine 满足）。
type MigrationImporter interface {
	Import(ctx context.Context, payload migratorModel.MigrationPayload, report func(phase string, snapshot any)) (any, error)
}

var (
	_ MigrationImporter = (*coreMigrator.ImportEngine)(nil)
	// 胶囊引擎的 Import 与 ImportEngine 同签名，故复用同一接口。
	_ MigrationImporter = (*coreMigrator.CapsuleEngine)(nil)
)

// MigrationRunner 把迁移导入包成作业 Runner，按 source_type 在「快照/Memos 迁移」与
// 「胶囊导入」之间分派。两者语义差别很大——前者是整库替换式迁移，后者是按 id 幂等的内容
// 追加——但都只是「拿着 source_payload 干活并回报阶段」，共用作业类型即可。
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
