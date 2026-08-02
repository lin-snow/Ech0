// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
	capsuleCheck "github.com/lin-snow/ech0/internal/capsule/check"
	capsuleExport "github.com/lin-snow/ech0/internal/capsule/export"
	capsuleImporter "github.com/lin-snow/ech0/internal/capsule/importer"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/migrator/artifact"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"gorm.io/gorm"
)

// CapsuleEngine 跑胶囊的导出与导入编排，与 ExportEngine / ImportEngine 对称：不感知作业
// 状态机，只接受裸 report 回调。
//
// 胶囊包相较快照多要一层依赖（直连 GORM 读写、事务、KV），故这里持有的是它们本身而不是
// 快照那套 StorageManager 抽象——胶囊包刻意不过 service 层，见 internal/capsule/export 包注释。
type CapsuleEngine struct {
	db             *gorm.DB
	storageManager StorageManager
	durableKV      kvstore.Store
	tx             transaction.Transactor
}

func NewCapsuleEngine(
	db *gorm.DB,
	storageManager StorageManager,
	durableKV kvstore.Store,
	tx transaction.Transactor,
) *CapsuleEngine {
	return &CapsuleEngine{db: db, storageManager: storageManager, durableKV: durableKV, tx: tx}
}

// Export 把当前实例导出成一个胶囊 zip，落在胶囊槽位里。
//
// 产出形态与快照对齐（同一个 ExportOutcome、同样「只保留最新一份」），因此下载出口、作业
// 状态机、前端进度卡三者都无需为胶囊分叉。
func (e *CapsuleEngine) Export(
	ctx context.Context,
	includePrivate bool,
	report func(phase string, snapshot any),
) (ExportOutcome, error) {
	slot := artifact.Capsules()
	if err := os.MkdirAll(slot.Dir(), 0o755); err != nil {
		return ExportOutcome{}, fmt.Errorf("create capsule dir: %w", err)
	}

	fileName := slot.Name(time.Now().UTC())
	outPath := slot.Path(fileName)

	// 胶囊导出内部不分阶段上报，整体就是一趟「收集记录 + 落媒体字节 + 打包」。
	report(migratorModel.ExportPhasePacking, nil)

	if _, err := capsuleExport.Run(ctx, capsuleExport.Deps{
		DB:       e.db,
		Selector: e.selector(),
		KV:       e.durableKV,
	}, capsuleExport.Options{
		Output:         outPath,
		IncludePrivate: includePrivate,
		Zip:            true,
		Generator:      "ech0 v" + versionPkg.Version,
	}); err != nil {
		return ExportOutcome{}, err
	}

	info, err := os.Stat(outPath)
	if err != nil {
		return ExportOutcome{}, fmt.Errorf("stat capsule artifact: %w", err)
	}

	// 与快照一致只留最新一份：胶囊是 100MB 量级的派生物，攒着只会吃满磁盘。
	if err := slot.KeepOnly(fileName); err != nil {
		return ExportOutcome{}, err
	}

	return ExportOutcome{
		ArtifactPath: outPath,
		FileName:     fileName,
		Size:         info.Size(),
		Format:       migratorModel.ExportFormatCapsule,
	}, nil
}

// Import 校验并导入 source_payload.tmp_dir 指向的胶囊，签名与 ImportEngine.Import 对称，
// 因此 MigrationRunner 只需按 source_type 二选一，作业状态机与前端轮询都无需分叉。
//
// 校验是硬前置而非建议：胶囊可以是手写的、也可以来自第三方生成器，带错误级问题的胶囊落库
// 会留下半截数据。spec §7 规定 import 隐式执行同一套校验，此处与 CLI 行为一致。
func (e *CapsuleEngine) Import(
	ctx context.Context,
	payload migratorModel.MigrationPayload,
	report func(phase string, snapshot any),
) (any, error) {
	logUtil.GetLogger().Info("capsule import started", slog.String("module", "migration"))
	defer func() {
		if err := CleanupTmpDirFromPayload(payload.SourcePayload); err != nil {
			logUtil.GetLogger().Warn("Failed to cleanup capsule temp directory",
				slog.String("module", "migration"), logUtil.Err(err))
		}
	}()

	// 复用迁移上传那套解析：它拒绝绝对路径与越级，tmp_dir 来自请求体，不能直接当路径用。
	dir, ok := resolveTmpDir(payload.SourcePayload)
	if !ok {
		return nil, errors.New("胶囊来源缺失或路径非法")
	}

	src, err := capsule.Open(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = src.Close() }()

	report(migratorModel.ImportPhaseChecking, nil)
	loaded, checkReport, err := capsuleCheck.Run(ctx, src, capsuleCheck.Options{})
	if err != nil {
		return nil, err
	}
	if checkReport.HasErrors() {
		return nil, fmt.Errorf("capsule failed validation, refusing to import: %s", checkReport.ErrorSummary())
	}

	includePrivate, _ := payload.SourcePayload["include_private"].(bool)

	report(migratorModel.ImportPhaseImporting, nil)
	result, err := capsuleImporter.Run(ctx, capsuleImporter.Deps{
		DB:       e.db,
		Tx:       e.tx,
		Selector: e.selector(),
		KV:       e.durableKV,
	}, loaded, capsuleImporter.Options{IncludePrivate: includePrivate})
	if err != nil {
		return nil, err
	}
	report(migratorModel.MigrationPhaseCompleted, nil)

	logUtil.GetLogger().Info("capsule import completed",
		slog.String("module", "migration"),
		slog.Int("echoes_created", result.EchoesCreated),
		slog.Int("files_created", result.FilesCreated),
	)

	// 与 ImportEngine 一样把结果挂回 payload 的 report 位，前端读同一个位置展示计数。
	enriched := payload
	if enriched.SourcePayload == nil {
		enriched.SourcePayload = map[string]any{}
	}
	enriched.SourcePayload["report"] = map[string]any{
		"echoes_created":   result.EchoesCreated,
		"echoes_skipped":   result.EchoesSkipped,
		"files_created":    result.FilesCreated,
		"comments_created": result.CommentsCreated,
		"warnings":         checkReport.Count(capsuleCheck.LevelWarning),
	}
	return enriched, nil
}

func (e *CapsuleEngine) selector() *storage.StorageSelector {
	if e.storageManager == nil {
		return nil
	}
	return e.storageManager.GetSelector()
}
