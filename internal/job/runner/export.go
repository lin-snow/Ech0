// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/job"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	"github.com/lin-snow/ech0/pkg/busen"
)

type SnapshotExporter interface {
	Export(ctx context.Context, report func(phase string, snapshot any)) (coreMigrator.ExportOutcome, error)
}

type CapsuleExporter interface {
	Export(
		ctx context.Context,
		includePrivate bool,
		report func(phase string, snapshot any),
	) (coreMigrator.ExportOutcome, error)
}

var (
	_ SnapshotExporter = (*coreMigrator.ExportEngine)(nil)
	_ CapsuleExporter  = (*coreMigrator.CapsuleEngine)(nil)
)

type ExportRunner struct {
	exporter        SnapshotExporter
	capsuleExporter CapsuleExporter
	bus             *busen.Bus
}

func NewExportRunner(
	exporter *coreMigrator.ExportEngine,
	capsuleExporter *coreMigrator.CapsuleEngine,
	busProvider func() *busen.Bus,
) *ExportRunner {
	return &ExportRunner{exporter: exporter, capsuleExporter: capsuleExporter, bus: busProvider()}
}

func (r *ExportRunner) Run(
	ctx context.Context,
	p migratorModel.ExportPayload,
	report job.ReportFunc,
) (any, error) {
	if p.Format == migratorModel.ExportFormatCapsule {
		return r.capsuleExporter.Export(ctx, p.IncludePrivate, report)
	}

	outcome, err := r.exporter.Export(ctx, report)
	if err != nil {
		return nil, err
	}

	eventbus.Notify(ctx, r.bus, event.SystemSnapshot{Info: "System manual snapshot completed"})

	return outcome, nil
}
