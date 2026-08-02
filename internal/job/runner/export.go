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

// SnapshotExporter 是快照导出执行端，便于测试解耦（由 migrator.ExportEngine 满足）。
type SnapshotExporter interface {
	Export(ctx context.Context, report func(phase string, snapshot any)) (coreMigrator.ExportOutcome, error)
}

// CapsuleExporter 是胶囊导出执行端（由 migrator.CapsuleEngine 满足）。
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

// ExportRunner 把导出执行包成作业 Runner（手动导出的异步出口），按 payload.Format 在快照与
// 胶囊之间分派。两种格式共用同一作业类型：它们都是重 IO 的整库打包，互斥跑才不会打满磁盘，
// 且前端那套轮询与进度卡可以原样复用。
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
		// 刻意不发 SystemSnapshot：webhook 订阅它来确认「备份已完成」，而胶囊不含账号与凭据、
		// 不能用于灾难恢复，拿它冒充备份完成会给用户错误的安全感。
		return r.capsuleExporter.Export(ctx, p.IncludePrivate, report)
	}

	outcome, err := r.exporter.Export(ctx, report)
	if err != nil {
		return nil, err
	}

	eventbus.Notify(ctx, r.bus, event.SystemSnapshot{Info: "System manual snapshot completed"})

	return outcome, nil
}
