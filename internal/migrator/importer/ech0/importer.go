// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package ech0

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"uuid"

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

type Importer struct {
	registry *storage.Registry
}

// New builds the snapshot importer. The registry is where migrated bytes are
// written: a snapshot only carries bytes, never a target storage binding.
func New(registry *storage.Registry) *Importer {
	return &Importer{registry: registry}
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
		if err := migrateEchos(ctx, tx, sourceDB, sourceRoot, e.registry); err != nil {
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

func migrateEchos(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB, sourceRoot string, registry *storage.Registry) error {
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
	if err := migrateFiles(ctx, tx, sourceDB, sourceRoot, registry); err != nil {
		return err
	}
	if err := migrateComments(ctx, tx, sourceDB); err != nil {
		return err
	}
	if err := migrateReactions(ctx, tx, sourceDB); err != nil {
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

// legacyFile is the pre-rewrite files row. It is read as a bare table because
// the current file model no longer has these columns.
type legacyFile struct {
	ID          string
	Key         string
	StorageType string
	Provider    string
	Bucket      string
	URL         string
	Name        string
	ContentType string
	Size        int64
	Width       int
	Height      int
	UserID      string
	CreatedAt   int64
	UpdatedAt   int64
}

// legacyEchoFile is the pre-rewrite attachment row; it becomes a file_link.
type legacyEchoFile struct {
	EchoID    string
	FileID    string
	SortOrder int
}

func migrateFiles(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB, sourceRoot string, registry *storage.Registry) error {
	if !sourceDB.Migrator().HasTable("files") {
		return nil
	}
	sourceFiles, err := loadRows[legacyFile](ctx, sourceDB, "files")
	if err != nil {
		return fmt.Errorf("load source files: %w", err)
	}
	if len(sourceFiles) == 0 {
		return nil
	}

	// Snapshots written before attachments were links have no echo_files
	// table; their files still import, they just arrive without links.
	var sourceEchoFiles []legacyEchoFile
	if sourceDB.Migrator().HasTable("echo_files") {
		sourceEchoFiles, err = loadRows[legacyEchoFile](ctx, sourceDB, "echo_files")
		if err != nil {
			return fmt.Errorf("load source echo_files: %w", err)
		}
	}

	// A file attached to an echo is an attachment (access=owner); everything
	// else the snapshot carried is an instance asset (access=public).
	attachments := make(map[string]bool, len(sourceEchoFiles))
	for i := range sourceEchoFiles {
		if id := strings.TrimSpace(sourceEchoFiles[i].FileID); id != "" {
			attachments[id] = true
		}
	}

	// Locators are the dedupe keys: object_key for owned bytes, ref_url for
	// external references. A locator that already exists is reused, so a
	// repeated import writes no second row and no second copy of the bytes.
	locators, err := loadFileLocators(ctx, tx)
	if err != nil {
		return err
	}

	importedID := make(map[string]string, len(sourceFiles))
	skipped := 0
	for i := range sourceFiles {
		src := sourceFiles[i]
		legacyID := strings.TrimSpace(src.ID)
		if legacyID == "" {
			skipped++
			continue
		}
		row, locator, srcPath, err := prepareFile(src, sourceRoot, attachments[legacyID])
		if err != nil {
			skipped++
			logUtil.GetLogger().Warn("migration ech0 file skipped",
				slog.String("module", "migration"),
				slog.String("file_id", legacyID),
				slog.String("file_key", src.Key),
				logUtil.Err(err),
			)
			continue
		}
		if existingID, ok := locators[locator]; ok {
			importedID[legacyID] = existingID
			continue
		}
		if srcPath != "" {
			if err := putFileBytes(ctx, &row, srcPath, registry); err != nil {
				skipped++
				logUtil.GetLogger().Warn("migration ech0 file skipped",
					slog.String("module", "migration"),
					slog.String("file_id", legacyID),
					slog.String("file_key", src.Key),
					logUtil.Err(err),
				)
				continue
			}
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		locators[locator] = row.ID
		importedID[legacyID] = row.ID
	}

	newLinks := make([]fileModel.Link, 0, len(sourceEchoFiles))
	seenLinks := make(map[string]bool, len(sourceEchoFiles))
	for i := range sourceEchoFiles {
		src := sourceEchoFiles[i]
		fileID := importedID[strings.TrimSpace(src.FileID)]
		ownerID := strings.TrimSpace(src.EchoID)
		if fileID == "" || ownerID == "" {
			continue
		}
		locator := fileID + "|" + ownerID + "|" + strconv.Itoa(src.SortOrder)
		if seenLinks[locator] {
			continue
		}
		seenLinks[locator] = true
		newLinks = append(newLinks, fileModel.Link{
			FileID:    fileID,
			OwnerKind: fileModel.OwnerKindEcho,
			OwnerID:   ownerID,
			Role:      fileModel.RoleAttachment,
			SortOrder: src.SortOrder,
		})
	}
	if len(newLinks) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newLinks, dbBatchSize).Error; err != nil {
			return fmt.Errorf("create file_links: %w", err)
		}
	}

	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "files"),
		slog.Int("source_files", len(sourceFiles)),
		slog.Int("source_echo_files", len(sourceEchoFiles)),
		slog.Int("created_files", len(importedID)),
		slog.Int("created_links", len(newLinks)),
		slog.Int("skipped", skipped),
	)
	return nil
}

// prepareFile turns one legacy row into a new file row without touching any
// store, so the caller can dedupe on the returned locator first. The locator is
// the object key for owned bytes and "ref|<url>" for external references.
func prepareFile(src legacyFile, sourceRoot string, attachment bool) (fileModel.File, string, string, error) {
	row := fileModel.File{
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusReady),
		Access:      string(fileModel.AccessPublic),
		Name:        strings.TrimSpace(src.Name),
		ContentType: strings.TrimSpace(src.ContentType),
		Size:        src.Size,
		Width:       src.Width,
		Height:      src.Height,
		UploaderID:  strings.TrimSpace(src.UserID),
		CreatedAt:   src.CreatedAt,
		UpdatedAt:   src.UpdatedAt,
	}
	if attachment {
		row.Access = string(fileModel.AccessOwner)
	}
	if row.Name == "" {
		row.Name = filepath.Base(strings.TrimSpace(src.Key))
	}
	if row.ContentType == "" {
		row.ContentType = "application/octet-stream"
	}
	if row.CreatedAt <= 0 {
		row.CreatedAt = time.Now().UTC().Unix()
	}
	if row.UpdatedAt <= 0 {
		row.UpdatedAt = row.CreatedAt
	}

	srcPath := ""
	if isLocalStorageType(src.StorageType) {
		srcPath = findSourceLocalFile(sourceRoot, src.Key)
	}

	if srcPath != "" {
		ext := strings.TrimPrefix(filepath.Ext(row.Name), ".")
		if ext == "" {
			ext = strings.TrimPrefix(filepath.Ext(src.Key), ".")
		}
		fileID := stableFileID(strings.TrimSpace(src.ID), row.CreatedAt)
		key, err := storage.KeyFor(fileID, ext)
		if err != nil {
			return fileModel.File{}, "", "", err
		}
		row.ID = fileID
		row.ObjectKey = &key
		return row, key, srcPath, nil
	}

	if refURL := strings.TrimSpace(src.URL); refURL != "" {
		row.ID = stableFileID("url:"+refURL, row.CreatedAt)
		row.Kind = string(fileModel.KindRef)
		row.RefURL = &refURL
		return row, "ref|" + refURL, "", nil
	}
	return fileModel.File{}, "", "", errors.New("snapshot carried no bytes and no url")
}

// putFileBytes writes an owned row's bytes through the registry's default
// store and fills the storage columns from what the driver reports.
func putFileBytes(ctx context.Context, row *fileModel.File, srcPath string, registry *storage.Registry) error {
	if registry == nil {
		return errors.New("storage registry is required to import file bytes")
	}
	if row.ObjectKey == nil {
		return errors.New("owned file row has no object key")
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	storageID, store, err := registry.Default()
	if err != nil {
		return err
	}
	put, err := store.Put(ctx, *row.ObjectKey, f, info.Size())
	if err != nil {
		return err
	}
	row.StorageID = &storageID
	row.Size = put.Size
	if put.ContentType != "" && row.ContentType == "application/octet-stream" {
		row.ContentType = put.ContentType
	}
	return nil
}

// stableFileID gives one legacy identity one target id. A legacy id that is
// already UUIDv7 is reused as-is; anything else is mapped to a v7-shaped id
// derived from the legacy identity, so re-importing a snapshot rewrites the
// same object keys instead of a fresh copy under a minted id.
func stableFileID(legacyID string, createdAt int64) string {
	if _, err := storage.KeyFor(legacyID, ""); err == nil {
		return legacyID
	}
	sum := sha256.Sum256([]byte("ech0.migrator.file:" + legacyID))
	var raw [16]byte
	copy(raw[:], sum[:16])
	if createdAt > 0 {
		ms := uint64(createdAt) * 1000
		raw[0] = byte(ms >> 40)
		raw[1] = byte(ms >> 32)
		raw[2] = byte(ms >> 24)
		raw[3] = byte(ms >> 16)
		raw[4] = byte(ms >> 8)
		raw[5] = byte(ms)
	}
	raw[6] = raw[6]&0x0f | 0x70
	raw[8] = raw[8]&0x3f | 0x80
	return uuid.UUID(raw).String()
}

// loadFileLocators maps every existing locator to its file id so an import
// that runs twice reuses the rows it created the first time.
func loadFileLocators(ctx context.Context, tx *gorm.DB) (map[string]string, error) {
	var rows []fileModel.File
	if err := tx.WithContext(ctx).Where("object_key IS NOT NULL OR ref_url IS NOT NULL").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load existing file locators: %w", err)
	}
	locators := make(map[string]string, len(rows))
	for i := range rows {
		if rows[i].ObjectKey != nil && *rows[i].ObjectKey != "" {
			locators[*rows[i].ObjectKey] = rows[i].ID
		}
		if rows[i].RefURL != nil && *rows[i].RefURL != "" {
			locators["ref|"+*rows[i].RefURL] = rows[i].ID
		}
	}
	return locators, nil
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

// migrateReactions copies reaction rows when the source snapshot has them.
// Older snapshots predate the table, so its absence is not an error.
func migrateReactions(ctx context.Context, tx *gorm.DB, sourceDB *gorm.DB) error {
	if !sourceDB.Migrator().HasTable("echo_reactions") {
		return nil
	}
	sourceReactions, err := loadRows[echoModel.Reaction](ctx, sourceDB, "echo_reactions")
	if err != nil {
		return fmt.Errorf("load source reactions: %w", err)
	}
	if len(sourceReactions) == 0 {
		return nil
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(sourceReactions, dbBatchSize).Error; err != nil {
		return fmt.Errorf("migrate reactions: %w", err)
	}
	logUtil.GetLogger().Info("migration ech0 phase completed",
		slog.String("module", "migration"),
		slog.String("phase", "echo_reactions"),
		slog.Int("count", len(sourceReactions)),
	)
	return nil
}

func isLocalStorageType(storageType string) bool {
	st := strings.ToLower(strings.TrimSpace(storageType))
	return st == "" || st == "local"
}

// legacyMediaRoutes are the top-level directories the pre-rewrite local store
// filed bytes under. Snapshots keep that layout verbatim, and older rows may
// carry a bare key, so both the prefixed and the bare shape are tried.
var legacyMediaRoutes = []string{"images", "audios", "videos", "documents", "files"}

// findSourceLocalFile resolves a legacy key inside the snapshot.
func findSourceLocalFile(sourceRoot, key string) string {
	cleanKey := strings.Trim(strings.TrimSpace(key), "/")
	if cleanKey == "" {
		return ""
	}
	native := filepath.FromSlash(cleanKey)
	candidates := []string{
		filepath.Join(sourceRoot, "files", native),
		filepath.Join(sourceRoot, native),
	}
	for _, route := range legacyMediaRoutes {
		candidates = append(candidates,
			filepath.Join(sourceRoot, "files", route, native),
			filepath.Join(sourceRoot, route, native),
		)
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
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
