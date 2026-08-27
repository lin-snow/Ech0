// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package ech0

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Importer struct{}

func New() *Importer {
	return &Importer{}
}

func (e *Importer) Import(ctx context.Context, req spec.ImportRequest) (spec.ImportResult, error) {
	sourceDBPath, sourceRoot, err := resolveSourceDBPath(req.SourcePayload)
	if err != nil {
		return spec.ImportResult{}, err
	}
	sourceDB, err := gorm.Open(sqlite.Open(sourceDBPath), &gorm.Config{})
	if err != nil {
		return spec.ImportResult{}, fmt.Errorf("open source sqlite: %w", err)
	}
	defer closeGormDB(sourceDB)

	var total int64
	if err := sourceDB.Table("echos").Count(&total).Error; err != nil {
		return spec.ImportResult{}, fmt.Errorf("count source echos: %w", err)
	}
	jobID := uuidUtil.NewV7()
	report := map[string]any{
		"job_id":        jobID,
		"source_db":     sourceDBPath,
		"source_root":   sourceRoot,
		"processed":     total,
		"success_count": total,
		"fail_count":    int64(0),
		"failed_items":  []spec.FailedItem{},
	}
	logUtil.GetLogger().Info("migration ech0 started",
		slog.String("module", "migration"),
		slog.String("job_id", jobID),
		slog.String("source_db", sourceDBPath),
		slog.String("source_root", sourceRoot),
		slog.Int64("total", total),
	)

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.ImportProgress{
			CurrentPhase: specPhaseExtracting,
			Processed:    0,
			Total:        total,
			SuccessCount: 0,
			FailCount:    0,
		})
	}

	if err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := migrateEchos(ctx, tx, sourceDB, sourceRoot); err != nil {
			return err
		}
		if req.UpdateProgress != nil {
			req.UpdateProgress(spec.ImportProgress{
				CurrentPhase: specPhaseLoading,
				Processed:    total,
				Total:        total,
				SuccessCount: total,
				FailCount:    0,
			})
		}
		return nil
	}); err != nil {
		logUtil.GetLogger().Error("migration ech0 failed",
			slog.String("module", "migration"),
			slog.String("job_id", jobID),
			logUtil.Err(err),
		)
		return spec.ImportResult{}, err
	}

	appendSettingToReport(sourceDB, report, commonModel.SystemSettingsKey, "source_system_setting")
	appendSettingToReport(sourceDB, report, commentModel.CommentSystemSettingKey, "source_comment_setting")
	appendSettingToReport(sourceDB, report, commonModel.S3SettingKey, "source_s3_setting")
	appendSettingToReport(sourceDB, report, commonModel.OAuth2SettingKey, "source_oauth2_setting")

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.ImportProgress{
			CurrentPhase: specPhaseReporting,
			Processed:    total,
			Total:        total,
			SuccessCount: total,
			FailCount:    0,
		})
		req.UpdateProgress(spec.ImportProgress{
			CurrentPhase: specPhaseCompleted,
			Processed:    total,
			Total:        total,
			SuccessCount: total,
			FailCount:    0,
		})
	}
	logUtil.GetLogger().Info("migration ech0 finished",
		slog.String("module", "migration"),
		slog.String("job_id", jobID),
		slog.Int64("processed", total),
		slog.Int64("success_count", total),
		slog.Int64("fail_count", 0),
	)

	return spec.ImportResult{
		Processed:    total,
		Total:        total,
		SuccessCount: total,
		FailCount:    0,
		ErrorSummary: fmt.Sprintf("迁移完成: success=%d fail=%d", total, 0),
		JobID:        jobID,
		Report:       report,
	}, nil
}

func resolveSourceDBPath(payload map[string]any) (string, string, error) {
	tmpDir, ok := payload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDir) == "" {
		return "", "", errors.New("source_payload.tmp_dir is required")
	}
	sourceRoot := filepath.Join("data", filepath.FromSlash(strings.TrimSpace(tmpDir)))
	dbPath := filepath.Join(sourceRoot, "ech0.db")
	if _, err := os.Stat(dbPath); err != nil {
		return "", "", fmt.Errorf("source db not found: %w", err)
	}
	return dbPath, sourceRoot, nil
}

func migrateEchos(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB, sourceRoot string) error {
	sourceEchos, err := loadRows[echoModel.Echo](ctx, sourceDB, "echos")
	if err != nil {
		return fmt.Errorf("load source echos: %w", err)
	}
	if len(sourceEchos) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(sourceEchos, dbBatchSize).Error; err != nil {
			return fmt.Errorf("migrate echos: %w", err)
		}
	}
	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "echos"),
		slog.Int("count", len(sourceEchos)),
	)
	sourceExtensions, err := loadRows[echoModel.EchoExtension](ctx, sourceDB, "echo_extensions")
	if err != nil {
		return fmt.Errorf("load source echo extensions: %w", err)
	}
	if len(sourceExtensions) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(sourceExtensions, dbBatchSize).Error; err != nil {
			return fmt.Errorf("migrate echo extensions: %w", err)
		}
	}
	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "echo_extensions"),
		slog.Int("count", len(sourceExtensions)),
	)

	if err := migrateTags(ctx, tx, sourceDB); err != nil {
		return err
	}
	if err := migrateFiles(ctx, tx, sourceDB, sourceRoot); err != nil {
		return err
	}
	if err := migrateComments(ctx, tx, sourceDB); err != nil {
		return err
	}
	return nil
}

func migrateTags(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB) error {
	sourceTags, err := loadRows[echoModel.Tag](ctx, sourceDB, "tags")
	if err != nil {
		return fmt.Errorf("load source tags: %w", err)
	}
	sourceEchoTags, err := loadRows[echoModel.EchoTag](ctx, sourceDB, "echo_tags")
	if err != nil {
		return fmt.Errorf("load source echo tags: %w", err)
	}
	if len(sourceTags) == 0 || len(sourceEchoTags) == 0 {
		return nil
	}

	tagNames := make([]string, 0, len(sourceTags))
	tagByID := make(map[string]echoModel.Tag, len(sourceTags))
	for i := range sourceTags {
		tag := sourceTags[i]
		if strings.TrimSpace(tag.Name) == "" {
			continue
		}
		tagNames = append(tagNames, tag.Name)
		tagByID[tag.ID] = tag
	}

	existing := make(map[string]string)
	if len(tagNames) > 0 {
		var existingTags []echoModel.Tag
		if err := tx.WithContext(ctx).Where("name IN ?", tagNames).Find(&existingTags).Error; err != nil {
			return fmt.Errorf("load existing tags: %w", err)
		}
		for i := range existingTags {
			existing[existingTags[i].Name] = existingTags[i].ID
		}
	}

	tagIDMap := make(map[string]string, len(tagByID))
	newTags := make([]echoModel.Tag, 0)
	for oldID, tag := range tagByID {
		if mappedID, ok := existing[tag.Name]; ok {
			tagIDMap[oldID] = mappedID
			continue
		}
		if strings.TrimSpace(tag.ID) == "" {
			tag.ID = uuidUtil.NewV7()
		}
		newTags = append(newTags, tag)
		existing[tag.Name] = tag.ID
		tagIDMap[oldID] = tag.ID
	}
	if len(newTags) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newTags, dbBatchSize).Error; err != nil {
			return fmt.Errorf("create tags: %w", err)
		}
	}

	newEchoTags := make([]echoModel.EchoTag, 0, len(sourceEchoTags))
	for i := range sourceEchoTags {
		row := sourceEchoTags[i]
		tagID, ok := tagIDMap[row.TagID]
		if !ok {
			continue
		}
		newEchoTags = append(newEchoTags, echoModel.EchoTag{
			EchoID: row.EchoID,
			TagID:  tagID,
		})
	}
	if len(newEchoTags) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newEchoTags, dbBatchSize).Error; err != nil {
			return fmt.Errorf("create echo_tags: %w", err)
		}
	}
	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "tags"),
		slog.Int("source_tags", len(sourceTags)),
		slog.Int("source_echo_tags", len(sourceEchoTags)),
		slog.Int("created_tags", len(newTags)),
		slog.Int("created_echo_tags", len(newEchoTags)),
	)
	return nil
}

func migrateFiles(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB, sourceRoot string) error {
	sourceFiles, err := loadRows[fileModel.File](ctx, sourceDB, "files")
	if err != nil {
		return fmt.Errorf("load source files: %w", err)
	}
	sourceEchoFiles, err := loadRows[fileModel.EchoFile](ctx, sourceDB, "echo_files")
	if err != nil {
		return fmt.Errorf("load source echo_files: %w", err)
	}
	if len(sourceFiles) == 0 || len(sourceEchoFiles) == 0 {
		return nil
	}

	routeKeyToID := make(map[string]string, len(sourceFiles))
	newFiles := make([]fileModel.File, 0, len(sourceFiles))
	for i := range sourceFiles {
		file := sourceFiles[i]
		key := fileRouteKey(file.StorageType, file.Provider, file.Bucket, file.Key)
		routeKeyToID[key] = file.ID
		if strings.TrimSpace(file.ID) == "" {
			file.ID = uuidUtil.NewV7()
			routeKeyToID[key] = file.ID
		}
		newFiles = append(newFiles, file)
	}
	if len(newFiles) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newFiles, dbBatchSize).Error; err != nil {
			return fmt.Errorf("create files: %w", err)
		}
	}

	var existingFiles []fileModel.File
	if err := tx.WithContext(ctx).Where("id IN ?", mapValues(routeKeyToID)).Find(&existingFiles).Error; err != nil {
		return fmt.Errorf("load existing files by id: %w", err)
	}
	if len(existingFiles) != len(routeKeyToID) {
		for routeKey := range routeKeyToID {
			storageType, provider, bucket, key := splitRouteKey(routeKey)
			var f fileModel.File
			if err := tx.WithContext(ctx).Where(
				"storage_type = ? AND provider = ? AND bucket = ? AND key = ?",
				storageType,
				provider,
				bucket,
				key,
			).Take(&f).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return fmt.Errorf("load existing file by route: %w", err)
			}
			routeKeyToID[routeKey] = f.ID
		}
	} else {
		for i := range existingFiles {
			file := existingFiles[i]
			routeKeyToID[fileRouteKey(file.StorageType, file.Provider, file.Bucket, file.Key)] = file.ID
		}
	}

	legacyFileIDToRoute := make(map[string]string, len(sourceFiles))
	for i := range sourceFiles {
		file := sourceFiles[i]
		legacyFileIDToRoute[file.ID] = fileRouteKey(file.StorageType, file.Provider, file.Bucket, file.Key)
	}

	newEchoFiles := make([]fileModel.EchoFile, 0, len(sourceEchoFiles))
	for i := range sourceEchoFiles {
		row := sourceEchoFiles[i]
		routeKey, ok := legacyFileIDToRoute[row.FileID]
		if !ok {
			continue
		}
		mappedFileID, ok := routeKeyToID[routeKey]
		if !ok || strings.TrimSpace(mappedFileID) == "" {
			continue
		}
		if strings.TrimSpace(row.ID) == "" {
			row.ID = uuidUtil.NewV7()
		}
		row.FileID = mappedFileID
		newEchoFiles = append(newEchoFiles, row)
	}
	if len(newEchoFiles) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newEchoFiles, dbBatchSize).Error; err != nil {
			return fmt.Errorf("create echo_files: %w", err)
		}
	}

	localCopyFailed := 0
	for i := range sourceFiles {
		file := sourceFiles[i]
		if !isLocalStorageType(file.StorageType) {
			continue
		}
		if err := copySourceLocalFileToTargetRoot(sourceRoot, file.Key); err != nil {
			localCopyFailed++
			logUtil.GetLogger().Warn("migration ech0 local file copy skipped",
				slog.String("module", "migration"),
				slog.String("file_key", file.Key),
				logUtil.Err(err),
			)
			continue
		}
	}
	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "files"),
		slog.Int("source_files", len(sourceFiles)),
		slog.Int("source_echo_files", len(sourceEchoFiles)),
		slog.Int("created_files", len(newFiles)),
		slog.Int("created_echo_files", len(newEchoFiles)),
		slog.Int("local_copy_failed", localCopyFailed),
	)
	return nil
}

func migrateComments(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB) error {
	sourceComments, err := loadRows[commentModel.Comment](ctx, sourceDB, "comments")
	if err != nil {
		return fmt.Errorf("load source comments: %w", err)
	}
	if len(sourceComments) == 0 {
		return nil
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(sourceComments, dbBatchSize).Error; err != nil {
		return fmt.Errorf("migrate comments: %w", err)
	}
	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "comments"),
		slog.Int("count", len(sourceComments)),
	)
	return nil
}

func isLocalStorageType(storageType string) bool {
	st := strings.ToLower(strings.TrimSpace(storageType))
	return st == "" || st == "local"
}

func copySourceLocalFileToTargetRoot(sourceRoot string, key string) error {
	relativePath := resolveLocalStoragePathByKey(key)
	if relativePath == "" {
		return errors.New("empty local file path")
	}

	targetRoot := filepath.Join("data", "files")
	targetPath := filepath.Join(targetRoot, filepath.FromSlash(relativePath))
	if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
		return nil
	}

	sourceCandidates := []string{
		filepath.Join(sourceRoot, "files", filepath.FromSlash(relativePath)),
		filepath.Join(sourceRoot, filepath.FromSlash(relativePath)),
	}
	if cleanKey := strings.Trim(strings.TrimSpace(key), "/"); cleanKey != "" {
		sourceCandidates = append(sourceCandidates, filepath.Join(sourceRoot, "files", filepath.FromSlash(cleanKey)))
		sourceCandidates = append(sourceCandidates, filepath.Join(sourceRoot, filepath.FromSlash(cleanKey)))
	}

	var srcPath string
	for _, candidate := range sourceCandidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			srcPath = candidate
			break
		}
	}
	if srcPath == "" {
		return errors.New("source local file not found")
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Sync()
}

func resolveLocalStoragePathByKey(key string) string {
	cleanKey := strings.Trim(strings.TrimSpace(key), "/")
	if cleanKey == "" {
		return ""
	}
	for _, route := range []string{"images/", "audios/", "videos/", "documents/", "files/"} {
		if strings.HasPrefix(cleanKey, route) {
			return cleanKey
		}
	}
	return storage.NewFileSchema().Resolve(cleanKey)
}

func appendSettingToReport(sourceDB *gorm.DB, report map[string]any, key string, reportKey string) {
	var kv commonModel.KeyValue
	if err := sourceDB.Table("key_values").Where("key = ?", key).Take(&kv).Error; err != nil {
		return
	}
	trimmed := strings.TrimSpace(kv.Value)
	if trimmed == "" {
		return
	}
	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		logUtil.GetLogger().Warn("migration ech0 setting parse skipped",
			slog.String("module", "migration"),
			slog.String("setting_key", key),
			logUtil.Err(err),
		)
		return
	}
	report[reportKey] = payload
}

func fileRouteKey(storageType string, provider string, bucket string, key string) string {
	return strings.TrimSpace(storageType) + "|" + strings.TrimSpace(provider) + "|" + strings.TrimSpace(bucket) + "|" + strings.TrimSpace(key)
}

func splitRouteKey(route string) (string, string, string, string) {
	parts := strings.SplitN(route, "|", 4)
	for len(parts) < 4 {
		parts = append(parts, "")
	}
	return parts[0], parts[1], parts[2], parts[3]
}

func mapValues(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, value := range m {
		out = append(out, value)
	}
	return out
}

func loadRows[T any](ctx context.Context, db *gorm.DB, table string) ([]T, error) {
	rows := make([]T, 0)
	if err := db.WithContext(ctx).Table(table).Find(&rows).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rows, nil
		}
		return nil, err
	}
	return rows, nil
}

func closeGormDB(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil || sqlDB == nil {
		return
	}
	_ = sqlDB.Close()
}

const (
	specPhaseExtracting = "extracting"
	specPhaseLoading    = "loading"
	specPhaseReporting  = "reporting"
	specPhaseCompleted  = "completed"
	dbBatchSize         = 200
)
