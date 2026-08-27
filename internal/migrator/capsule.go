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
