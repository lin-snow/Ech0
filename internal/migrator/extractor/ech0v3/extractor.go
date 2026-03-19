package ech0v3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	virefs "github.com/lin-snow/VireFS"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Extractor struct{}

const (
	echoLogSampleStep = int64(50)
	dbBatchSize       = 200
)

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Extract(_ context.Context, req spec.ExtractRequest) (spec.ExtractResult, error) {
	items, ok := req.SourcePayload["items"].([]any)
	if !ok || len(items) == 0 {
		return spec.ExtractResult{
			Records:        []spec.RawRecord{},
			NextCheckpoint: req.Checkpoint,
			HasMore:        false,
			TotalHint:      0,
		}, nil
	}

	start := int(req.Checkpoint)
	if start < 0 {
		start = 0
	}
	if start >= len(items) {
		return spec.ExtractResult{
			Records:        []spec.RawRecord{},
			NextCheckpoint: int64(len(items)),
			HasMore:        false,
			TotalHint:      int64(len(items)),
		}, nil
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	end := start + batchSize
	if end > len(items) {
		end = len(items)
	}

	records := make([]spec.RawRecord, 0, end-start)
	for i := start; i < end; i++ {
		obj, ok := items[i].(map[string]any)
		if !ok {
			return spec.ExtractResult{}, fmt.Errorf("ech0 v3 item at index %d is not object", i)
		}
		if _, exists := obj["content"]; !exists {
			if text, ok := obj["text"]; ok {
				obj["content"] = text
			}
		}
		sourceID := fmt.Sprintf("%v", obj["id"])
		if sourceID == "<nil>" {
			sourceID = fmt.Sprintf("ech0v3-%d", i)
		}
		records = append(records, spec.RawRecord{
			SourceID: sourceID,
			Data:     obj,
		})
	}

	return spec.ExtractResult{
		Records:        records,
		NextCheckpoint: int64(end),
		HasMore:        end < len(items),
		TotalHint:      int64(len(items)),
	}, nil
}

func (e *Extractor) Migrate(ctx context.Context, req spec.MigrateRequest) (spec.MigrateResult, error) {
	options := parseOptions(req.SourcePayload)
	sourceDBPath, sourceRoot, err := resolveSourceDBPath(req.SourcePayload)
	if err != nil {
		return spec.MigrateResult{}, err
	}

	sourceDB, err := gorm.Open(sqlite.Open(sourceDBPath), &gorm.Config{})
	if err != nil {
		return spec.MigrateResult{}, fmt.Errorf("open source sqlite: %w", err)
	}
	defer closeGormDB(sourceDB)

	var total int64
	if err := sourceDB.Table("echos").Count(&total).Error; err != nil {
		return spec.MigrateResult{}, fmt.Errorf("count source echos: %w", err)
	}

	jobID := uuidUtil.MustNewV7()
	report := map[string]any{
		"job_id":            jobID,
		"source_db":         sourceDBPath,
		"include_tags":      true,
		"include_images":    true,
		"failure_threshold": options.FailureThreshold,
		"processed":         int64(0),
		"success_count":     int64(0),
		"fail_count":        int64(0),
		"failed_items":      []spec.FailedItem{},
		"echo_id_map":       map[string]string{},
	}
	logUtil.GetLogger().Info("migration ech0_v3 started",
		zap.String("module", "migration"),
		zap.String("source_db", sourceDBPath),
		zap.String("job_id", jobID),
		zap.Int64("total", total),
	)

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: "extracting",
			Processed:    0,
			Total:        total,
		})
	}

	migrateErr := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sourceEchos []v3Echo
		if err := sourceDB.WithContext(ctx).Table("echos").Order("id ASC").Find(&sourceEchos).Error; err != nil {
			return fmt.Errorf("query source echos: %w", err)
		}
		echoIDsWithImages, err := loadEchoIDsWithImages(sourceDB)
		if err != nil {
			return fmt.Errorf("load image-linked echo ids: %w", err)
		}

		idMap := make(map[int64]string, len(sourceEchos))
		failed := make([]spec.FailedItem, 0)
		successCount := int64(0)
		failCount := int64(0)

		for i := range sourceEchos {
			if err := ctx.Err(); err != nil {
				return err
			}
			row := sourceEchos[i]
			content := strings.TrimSpace(row.Content)
			hasContent := content != ""
			hasExtension := strings.TrimSpace(row.Extension) != ""
			_, hasFiles := echoIDsWithImages[row.ID]

			if !hasContent && !hasExtension && !hasFiles {
				failCount++
				failed = append(failed, spec.FailedItem{
					SourceID: strconv.FormatInt(row.ID, 10),
					Reason:   "empty content, files and extension",
				})
				logUtil.GetLogger().Warn("migration echo skipped",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("source_echo_id", row.ID),
					zap.String("reason", "empty content, files and extension"),
				)
				if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
					return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
				}
				continue
			}

			echo := echoModel.Echo{
				Content:   content,
				Username:  normalizeUsername(row.Username),
				Layout:    normalizeLayout(row.Layout),
				Private:   row.Private,
				UserID:    options.CreatedBy,
				FavCount:  row.FavCount,
				CreatedAt: normalizeTime(row.CreatedAt),
			}
			if err := tx.Create(&echo).Error; err != nil {
				failCount++
				failed = append(failed, spec.FailedItem{
					SourceID: strconv.FormatInt(row.ID, 10),
					Reason:   err.Error(),
				})
				logUtil.GetLogger().Warn("migration echo create failed",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("source_echo_id", row.ID),
					zap.String("reason", err.Error()),
				)
				if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
					return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
				}
				continue
			}

			if err := persistExtension(tx, echo.ID, row.ExtensionType, row.Extension); err != nil {
				failCount++
				failed = append(failed, spec.FailedItem{
					SourceID: strconv.FormatInt(row.ID, 10),
					Reason:   fmt.Sprintf("save extension: %v", err),
				})
				logUtil.GetLogger().Warn("migration echo extension failed",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("source_echo_id", row.ID),
					zap.String("target_echo_id", echo.ID),
					zap.String("reason", err.Error()),
				)
				if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
					return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
				}
				continue
			}

			successCount++
			idMap[row.ID] = echo.ID
			if successCount%echoLogSampleStep == 0 {
				logUtil.GetLogger().Info("migration echo progress",
					zap.String("module", "migration"),
					zap.String("job_id", jobID),
					zap.Int64("processed", successCount+failCount),
					zap.Int64("success_count", successCount),
					zap.Int64("fail_count", failCount),
					zap.Int64("total", total),
				)
			}

			if req.UpdateProgress != nil {
				req.UpdateProgress(spec.MigrateProgress{
					CurrentPhase: "loading",
					Processed:    successCount + failCount,
					Total:        total,
					SuccessCount: successCount,
					FailCount:    failCount,
				})
			}
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		if err := migrateTags(tx, sourceDB, idMap); err != nil {
			return fmt.Errorf("migrate tags: %w", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		imageSummary, err := migrateImages(tx, sourceDB, sourceRoot, options.CreatedBy, idMap)
		if err != nil {
			return fmt.Errorf("migrate images: %w", err)
		}
		if imageSummary.SourceS3Setting != nil {
			report["source_s3_setting"] = s3SettingToMap(*imageSummary.SourceS3Setting)
		}
		if len(imageSummary.FailedItems) > 0 {
			failed = append(failed, imageSummary.FailedItems...)
			failCount += int64(len(imageSummary.FailedItems))
			if exceedsFailureThreshold(successCount+failCount, failCount, options.FailureThreshold) {
				return fmt.Errorf("failure rate exceeded %.2f%%, rollback job %s", options.FailureThreshold*100, jobID)
			}
		}

		if err := migrateSettings(tx, sourceDB); err != nil {
			return fmt.Errorf("migrate settings: %w", err)
		}

		report["processed"] = successCount + failCount
		report["success_count"] = successCount
		report["fail_count"] = failCount
		report["failed_items"] = failed
		report["echo_id_map"] = stringifyIDMap(idMap)
		return nil
	})
	if migrateErr != nil {
		logUtil.GetLogger().Error("migration ech0_v3 failed",
			zap.String("module", "migration"),
			zap.String("job_id", jobID),
			zap.Error(migrateErr),
		)
		return spec.MigrateResult{}, migrateErr
	}

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: "completed",
			Processed:    toInt64(report["processed"]),
			Total:        total,
			SuccessCount: toInt64(report["success_count"]),
			FailCount:    toInt64(report["fail_count"]),
		})
	}

	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.MigrateProgress{
			CurrentPhase: "reporting",
			Processed:    toInt64(report["processed"]),
			Total:        total,
			SuccessCount: toInt64(report["success_count"]),
			FailCount:    toInt64(report["fail_count"]),
		})
	}
	processed := toInt64(report["processed"])
	successCount := toInt64(report["success_count"])
	failCount := toInt64(report["fail_count"])
	logFields := []zap.Field{
		zap.String("module", "migration"),
		zap.String("job_id", jobID),
		zap.Int64("processed", processed),
		zap.Int64("success_count", successCount),
		zap.Int64("fail_count", failCount),
	}
	if failCount > 0 {
		logUtil.GetLogger().Warn("migration ech0_v3 finished with failures", logFields...)
	} else {
		logUtil.GetLogger().Info("migration ech0_v3 finished", logFields...)
	}

	return spec.MigrateResult{
		Processed:    processed,
		Total:        total,
		SuccessCount: successCount,
		FailCount:    failCount,
		ErrorSummary: fmt.Sprintf("迁移完成: success=%d fail=%d", successCount, failCount),
		JobID:        jobID,
		Report:       report,
	}, nil
}

type v3Echo struct {
	ID            int64     `gorm:"column:id"`
	Content       string    `gorm:"column:content"`
	Username      string    `gorm:"column:username"`
	Private       bool      `gorm:"column:private"`
	Extension     string    `gorm:"column:extension"`
	ExtensionType string    `gorm:"column:extension_type"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	FavCount      int       `gorm:"column:fav_count"`
	Layout        string    `gorm:"column:layout"`
}

type v3Tag struct {
	ID         int64     `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	UsageCount int       `gorm:"column:usage_count"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

type v3EchoTag struct {
	EchoID int64 `gorm:"column:echo_id"`
	TagID  int64 `gorm:"column:tag_id"`
}

type v3Image struct {
	ID        int64  `gorm:"column:id"`
	MessageID int64  `gorm:"column:message_id"`
	ImageURL  string `gorm:"column:image_url"`
	ImageSrc  string `gorm:"column:image_source"`
	ObjectKey string `gorm:"column:object_key"`
	Width     int    `gorm:"column:width"`
	Height    int    `gorm:"column:height"`
}

type migrationOptions struct {
	FailureThreshold float64
	CreatedBy        string
}

func parseOptions(payload map[string]any) migrationOptions {
	options := migrationOptions{
		FailureThreshold: 0.02,
		CreatedBy:        "migrator",
	}
	if v, ok := payload["failure_threshold"].(float64); ok && v > 0 && v < 1 {
		options.FailureThreshold = v
	}
	if v, ok := payload["created_by"].(string); ok && strings.TrimSpace(v) != "" {
		options.CreatedBy = strings.TrimSpace(v)
	}
	return options
}

func resolveSourceDBPath(payload map[string]any) (string, string, error) {
	tmpDir, ok := payload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDir) == "" {
		return "", "", errors.New("source_payload.tmp_dir is required")
	}
	root := filepath.Join("data", filepath.FromSlash(strings.TrimSpace(tmpDir)))
	dbPath := filepath.Join(root, "ech0.db")
	if _, err := os.Stat(dbPath); err != nil {
		return "", "", fmt.Errorf("source db not found: %w", err)
	}
	return dbPath, root, nil
}

func normalizeUsername(v string) string {
	if strings.TrimSpace(v) == "" {
		return "migrator"
	}
	return strings.TrimSpace(v)
}

func normalizeLayout(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case echoModel.LayoutGrid, echoModel.LayoutHorizontal, echoModel.LayoutCarousel, echoModel.LayoutWaterfall:
		return strings.TrimSpace(strings.ToLower(v))
	default:
		return echoModel.LayoutWaterfall
	}
}

func normalizeTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

func persistExtension(tx *gorm.DB, echoID string, extensionType string, raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		payload["raw"] = trimmed
	}
	extType := normalizeExtensionType(extensionType, payload)
	if extType == "LEGACY" {
		return nil
	}
	payload = normalizeExtensionPayload(extType, payload, trimmed)
	if len(payload) == 0 {
		return nil
	}
	return tx.Create(&echoModel.EchoExtension{
		EchoID:  echoID,
		Type:    extType,
		Payload: payload,
	}).Error
}

func normalizeExtensionType(extensionType string, payload map[string]any) string {
	rawType := strings.ToUpper(strings.TrimSpace(extensionType))
	switch rawType {
	case echoModel.Extension_MUSIC, "MUSIC163", "NETEASE":
		return echoModel.Extension_MUSIC
	case echoModel.Extension_VIDEO, "BILIBILI", "YOUTUBE":
		return echoModel.Extension_VIDEO
	case echoModel.Extension_GITHUBPROJ, "GITHUB", "GITHUB_PROJECT":
		return echoModel.Extension_GITHUBPROJ
	case echoModel.Extension_WEBSITE, "SITE", "LINK", "URL":
		return echoModel.Extension_WEBSITE
	}

	// fallback by payload shape
	if getPayloadString(payload, "repoUrl") != "" {
		return echoModel.Extension_GITHUBPROJ
	}
	if getPayloadString(payload, "videoId") != "" || getPayloadString(payload, "bvid") != "" {
		return echoModel.Extension_VIDEO
	}
	if getPayloadString(payload, "site") != "" || getPayloadString(payload, "title") != "" {
		return echoModel.Extension_WEBSITE
	}
	if getPayloadString(payload, "url") != "" || getPayloadString(payload, "raw") != "" {
		return echoModel.Extension_MUSIC
	}
	return "LEGACY"
}

func normalizeExtensionPayload(extType string, payload map[string]any, raw string) map[string]any {
	out := make(map[string]any)
	for k, v := range payload {
		out[k] = v
	}
	rawURL := strings.TrimSpace(getPayloadString(payload, "raw"))
	if rawURL == "" {
		rawURL = strings.TrimSpace(raw)
	}

	switch extType {
	case echoModel.Extension_GITHUBPROJ:
		repoURL := strings.TrimSpace(getPayloadString(payload, "repoUrl"))
		if repoURL == "" {
			repoURL = strings.TrimSpace(getPayloadString(payload, "url"))
		}
		if repoURL == "" && strings.Contains(strings.ToLower(rawURL), "github.com/") {
			repoURL = rawURL
		}
		if repoURL != "" {
			return map[string]any{"repoUrl": repoURL}
		}
		return out
	case echoModel.Extension_MUSIC:
		url := strings.TrimSpace(getPayloadString(payload, "url"))
		if url == "" {
			url = rawURL
		}
		if url != "" {
			return map[string]any{"url": url}
		}
		return out
	case echoModel.Extension_VIDEO:
		videoID := strings.TrimSpace(getPayloadString(payload, "videoId"))
		if videoID == "" {
			videoID = strings.TrimSpace(getPayloadString(payload, "bvid"))
		}
		if videoID == "" {
			videoID = rawURL
		}
		if videoID != "" {
			return map[string]any{"videoId": videoID}
		}
		return out
	case echoModel.Extension_WEBSITE:
		site := strings.TrimSpace(getPayloadString(payload, "site"))
		title := strings.TrimSpace(getPayloadString(payload, "title"))
		if site == "" && rawURL != "" && (strings.HasPrefix(strings.ToLower(rawURL), "http://") || strings.HasPrefix(strings.ToLower(rawURL), "https://")) {
			site = rawURL
		}
		if title == "" && site != "" {
			title = site
		}
		if site != "" && title != "" {
			return map[string]any{"title": title, "site": site}
		}
		return out
	default:
		return out
	}
}

func getPayloadString(payload map[string]any, key string) string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", raw))
	}
}

func migrateTags(tx *gorm.DB, sourceDB *gorm.DB, idMap map[int64]string) error {
	var sourceTags []v3Tag
	if err := sourceDB.Table("tags").Find(&sourceTags).Error; err != nil {
		return err
	}
	if len(sourceTags) == 0 {
		return nil
	}

	tagNames := make([]string, 0, len(sourceTags))
	for i := range sourceTags {
		name := strings.TrimSpace(sourceTags[i].Name)
		if name != "" {
			tagNames = append(tagNames, name)
		}
	}
	existingTags, err := loadTagsByNames(tx, tagNames)
	if err != nil {
		return err
	}

	tagIDMap := make(map[int64]string, len(sourceTags))
	newTags := make([]echoModel.Tag, 0)
	for i := range sourceTags {
		tagRow := sourceTags[i]
		name := strings.TrimSpace(tagRow.Name)
		if name == "" {
			continue
		}
		if tagID, ok := existingTags[name]; ok {
			tagIDMap[tagRow.ID] = tagID
			continue
		}
		newTagID := uuidUtil.MustNewV7()
		newTags = append(newTags, echoModel.Tag{
			ID:         newTagID,
			Name:       name,
			UsageCount: tagRow.UsageCount,
			CreatedAt:  normalizeTime(tagRow.CreatedAt),
		})
		existingTags[name] = newTagID
		tagIDMap[tagRow.ID] = newTagID
	}
	if len(newTags) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newTags, dbBatchSize).Error; err != nil {
			return err
		}
	}

	var sourceEchoTags []v3EchoTag
	if err := sourceDB.Table("echo_tags").Find(&sourceEchoTags).Error; err != nil {
		return err
	}
	relations := make([]echoModel.EchoTag, 0, len(sourceEchoTags))
	for i := range sourceEchoTags {
		row := sourceEchoTags[i]
		targetEchoID, ok := idMap[row.EchoID]
		if !ok {
			continue
		}
		targetTagID, ok := tagIDMap[row.TagID]
		if !ok {
			continue
		}
		relations = append(relations, echoModel.EchoTag{
			EchoID: targetEchoID,
			TagID:  targetTagID,
		})
	}
	if len(relations) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(relations, dbBatchSize).Error; err != nil {
			return err
		}
	}
	return nil
}

type imageMigrationSummary struct {
	FailedItems     []spec.FailedItem
	SourceS3Setting *settingModel.S3Setting
}

func migrateImages(tx *gorm.DB, sourceDB *gorm.DB, sourceRoot string, createdBy string, idMap map[int64]string) (imageMigrationSummary, error) {
	summary := imageMigrationSummary{
		FailedItems: make([]spec.FailedItem, 0),
	}
	var sourceImages []v3Image
	if err := sourceDB.Table("images").Order("id ASC").Find(&sourceImages).Error; err != nil {
		return summary, err
	}
	if len(sourceImages) == 0 {
		return summary, nil
	}
	// v3 contract:
	// - images.object_key is the physical object key used by v3 uploads/deletes.
	// - images.image_url is a display URL and may include endpoint/bucket/path_prefix.
	// v4 contract:
	// - files.key stores the business key (no schema prefix).
	// - object storage path is derived by ObjectFS schema (e.g. images/...).
	// So migration must resolve both source physical path and target schema path.
	sourceS3Setting, err := loadSourceS3Setting(sourceDB)
	if err != nil {
		return summary, fmt.Errorf("load source s3 setting: %w", err)
	}
	summary.SourceS3Setting = sourceS3Setting
	mappedS3Setting, s3SettingValid := mapSourceS3SettingToV4(sourceS3Setting)
	objectProvider := strings.ToLower(strings.TrimSpace(mappedS3Setting.Provider))
	objectBucket := strings.TrimSpace(mappedS3Setting.BucketName)
	var s3Selector *storage.StorageSelector
	var rawObjectFS virefs.FS
	if s3SettingValid {
		s3Cfg := buildStorageConfigFromS3Setting(mappedS3Setting)
		s3Selector = storage.NewStorageSelector(s3Cfg)
		if s3Selector == nil || !s3Selector.ObjectEnabled() {
			s3SettingValid = false
		}
		if s3SettingValid {
			rawFS, buildErr := buildRawObjectFSFromS3Setting(mappedS3Setting)
			if buildErr != nil {
				logUtil.GetLogger().Warn("build raw s3 fs failed",
					zap.String("module", "migration"),
					zap.Error(buildErr),
				)
				s3SettingValid = false
			} else {
				rawObjectFS = rawFS
			}
		}
	}

	type imageCandidate struct {
		sourceImageID              int64
		sourceMessageID            int64
		targetEchoID               string
		key                        string
		sourceObjectPathCandidates []string
		sourceImageURL             string
		imageURL                   string
		imageSrc                   string
		storageType                string
		provider                   string
		bucket                     string
		width                      int
		height                     int
	}
	candidates := make([]imageCandidate, 0, len(sourceImages))
	uniqueLocalKeys := make(map[string]struct{})
	uniqueExternalKeys := make(map[string]struct{})
	uniqueObjectKeys := make(map[string]struct{})
	for i := range sourceImages {
		row := sourceImages[i]
		targetEchoID, ok := idMap[row.MessageID]
		if !ok {
			continue
		}

		imageSrc := normalizeImageSource(row.ImageSrc, row.ImageURL)
		switch imageSrc {
		case "local":
			key := strings.TrimSpace(row.ObjectKey)
			if key == "" {
				key = normalizeKeyFromURL(row.ImageURL)
			}
			if key == "" {
				continue
			}
			candidates = append(candidates, imageCandidate{
				sourceImageID:   row.ID,
				sourceMessageID: row.MessageID,
				targetEchoID:    targetEchoID,
				key:             key,
				imageURL:        strings.TrimSpace(row.ImageURL),
				imageSrc:        imageSrc,
				storageType:     "local",
				provider:        "",
				bucket:          "",
				width:           row.Width,
				height:          row.Height,
			})
			uniqueLocalKeys[key] = struct{}{}
		case "s3":
			if !s3SettingValid {
				summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
					SourceID: strconv.FormatInt(row.MessageID, 10),
					Reason:   "invalid source s3 setting",
				})
				continue
			}
			migratedKey, sourceObjectCandidates := deriveS3ObjectKeyMapping(
				row.ObjectKey,
				row.ImageURL,
				mappedS3Setting.PathPrefix,
				mappedS3Setting.BucketName,
			)
			if migratedKey == "" {
				summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
					SourceID: strconv.FormatInt(row.MessageID, 10),
					Reason:   "invalid s3 object key",
				})
				continue
			}
			finalURL := buildObjectURLFromSetting(mappedS3Setting, migratedKey)
			candidates = append(candidates, imageCandidate{
				sourceImageID:              row.ID,
				sourceMessageID:            row.MessageID,
				targetEchoID:               targetEchoID,
				key:                        migratedKey,
				sourceObjectPathCandidates: sourceObjectCandidates,
				sourceImageURL:             strings.TrimSpace(row.ImageURL),
				imageURL:                   strings.TrimSpace(finalURL),
				imageSrc:                   imageSrc,
				storageType:                "object",
				provider:                   objectProvider,
				bucket:                     objectBucket,
				width:                      row.Width,
				height:                     row.Height,
			})
			uniqueObjectKeys[migratedKey] = struct{}{}
		default:
			finalURL := strings.TrimSpace(row.ImageURL)
			if !isAbsoluteURL(finalURL) {
				summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
					SourceID: strconv.FormatInt(row.MessageID, 10),
					Reason:   "invalid external image url",
				})
				continue
			}
			externalKey := buildExternalFileKey(finalURL, "image")
			candidates = append(candidates, imageCandidate{
				sourceImageID:   row.ID,
				sourceMessageID: row.MessageID,
				targetEchoID:    targetEchoID,
				key:             externalKey,
				imageURL:        finalURL,
				imageSrc:        imageSrc,
				storageType:     "external",
				provider:        "external",
				bucket:          "",
				width:           row.Width,
				height:          row.Height,
			})
			uniqueExternalKeys[externalKey] = struct{}{}
		}
	}
	if len(candidates) == 0 {
		return summary, nil
	}

	existingFileIDs := make(map[string]string)
	localKeys := mapKeys(uniqueLocalKeys)
	localMap, err := loadExistingFileIDsByRoute(tx, "local", "", "", localKeys)
	if err != nil {
		return summary, err
	}
	for k, v := range localMap {
		existingFileIDs[fileRouteKey("local", "", "", k)] = v
	}
	externalKeys := mapKeys(uniqueExternalKeys)
	externalMap, err := loadExistingFileIDsByRoute(tx, "external", "external", "", externalKeys)
	if err != nil {
		return summary, err
	}
	for k, v := range externalMap {
		existingFileIDs[fileRouteKey("external", "external", "", k)] = v
	}
	objectKeys := mapKeys(uniqueObjectKeys)
	objectMap, err := loadExistingFileIDsByRoute(tx, "object", objectProvider, objectBucket, objectKeys)
	if err != nil {
		return summary, err
	}
	for k, v := range objectMap {
		existingFileIDs[fileRouteKey("object", objectProvider, objectBucket, k)] = v
	}

	metaCache := make(map[string]fileMeta, len(candidates))
	newFileRecords := make([]fileModel.File, 0)
	for i := range candidates {
		candidate := candidates[i]
		routeKey := fileRouteKey(candidate.storageType, candidate.provider, candidate.bucket, candidate.key)
		if _, ok := existingFileIDs[routeKey]; ok {
			continue
		}
		meta, ok := metaCache[candidate.key]
		if !ok {
			if candidate.storageType == "local" {
				localPath := resolveMigratedLocalPath(candidate.key)
				meta.size, meta.contentType = inferFileMeta(sourceRoot, candidate.key, candidate.imageURL)
				if copyErr := copySourceImageToLocalStorage(sourceRoot, candidate.key, candidate.imageURL, localPath); copyErr != nil {
					summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
						SourceID: strconv.FormatInt(candidate.sourceMessageID, 10),
						Reason:   "copy local image failed: " + copyErr.Error(),
					})
					logUtil.GetLogger().Warn("copy source image to local storage failed",
						zap.String("module", "migration"),
						zap.String("source_key", candidate.key),
						zap.String("target_path", localPath),
						zap.Error(copyErr),
					)
					continue
				}
			} else if candidate.storageType == "object" {
				if s3Selector == nil {
					summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
						SourceID: strconv.FormatInt(candidate.sourceMessageID, 10),
						Reason:   "object storage selector unavailable",
					})
					continue
				}
				if copyErr := ensureS3ObjectAtSchemaPath(rawObjectFS, mappedS3Setting, candidate.key, candidate.sourceObjectPathCandidates, candidate.sourceImageURL); copyErr != nil {
					summary.FailedItems = append(summary.FailedItems, spec.FailedItem{
						SourceID: strconv.FormatInt(candidate.sourceMessageID, 10),
						Reason:   "copy s3 image failed: " + copyErr.Error(),
					})
					continue
				}
				meta.size = 0
				meta.contentType = inferContentType(candidate.imageURL, candidate.key)
			} else {
				meta.size = 0
				meta.contentType = inferContentType(candidate.imageURL, candidate.key)
			}
			metaCache[candidate.key] = meta
		}
		fileID := uuidUtil.MustNewV7()
		fileURL := strings.TrimSpace(candidate.imageURL)
		if candidate.storageType == "local" {
			fileURL = buildMigratedFileURL(candidate.key, candidate.imageURL, candidate.imageSrc)
		}
		newFileRecords = append(newFileRecords, fileModel.File{
			ID:          fileID,
			Key:         candidate.key,
			StorageType: candidate.storageType,
			Provider:    candidate.provider,
			Bucket:      candidate.bucket,
			URL:         fileURL,
			Name:        inferFileName(candidate.key, candidate.imageURL),
			ContentType: meta.contentType,
			Size:        meta.size,
			Width:       candidate.width,
			Height:      candidate.height,
			Category:    inferCategory(meta.contentType, candidate.key),
			UserID:      createdBy,
		})
		existingFileIDs[routeKey] = fileID
	}
	if len(newFileRecords) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(newFileRecords, dbBatchSize).Error; err != nil {
			return summary, err
		}
	}

	echoFiles := make([]fileModel.EchoFile, 0, len(candidates))
	for i := range candidates {
		candidate := candidates[i]
		routeKey := fileRouteKey(candidate.storageType, candidate.provider, candidate.bucket, candidate.key)
		fileID, ok := existingFileIDs[routeKey]
		if !ok || fileID == "" {
			continue
		}
		echoFiles = append(echoFiles, fileModel.EchoFile{
			EchoID: candidate.targetEchoID,
			FileID: fileID,
		})
	}
	if len(echoFiles) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(echoFiles, dbBatchSize).Error; err != nil {
			return summary, err
		}
	}
	return summary, nil
}

type fileMeta struct {
	size        int64
	contentType string
}

func inferFileMeta(sourceRoot string, key string, fallbackURL string) (int64, string) {
	contentType := inferContentType(fallbackURL, key)

	sourcePath := filepath.Join(sourceRoot, filepath.FromSlash(key))
	info, err := os.Stat(sourcePath)
	if err != nil {
		return 0, contentType
	}
	return info.Size(), contentType
}

func inferCategory(contentType string, key string) string {
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.HasPrefix(lowerType, "image/"):
		return "image"
	case strings.HasPrefix(lowerType, "audio/"):
		return "audio"
	case strings.HasPrefix(lowerType, "video/"):
		return "video"
	}
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg":
		return "image"
	case ".mp3", ".wav", ".ogg", ".m4a":
		return "audio"
	case ".mp4", ".mov", ".webm":
		return "video"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return "document"
	default:
		return "file"
	}
}

func normalizeKeyFromURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return ""
	}
	base := filepath.Base(trimmed)
	base = strings.Split(base, "?")[0]
	if base == "." || base == "/" {
		return ""
	}
	return base
}

func buildMigratedFileURL(key string, imageURL string, _ string) string {
	cleanKey := cleanMigratedFileKey(key)
	if cleanKey == "" {
		cleanKey = cleanMigratedFileKey(normalizeKeyFromURL(imageURL))
	}
	if cleanKey == "" {
		return strings.TrimSpace(imageURL)
	}
	resolvedPath := resolveMigratedLocalPath(cleanKey)
	if resolvedPath == "" {
		return strings.TrimSpace(imageURL)
	}
	return "/api/files/" + strings.TrimLeft(resolvedPath, "/")
}

func normalizeImageSource(source string, imageURL string) string {
	s := strings.ToLower(strings.TrimSpace(source))
	switch {
	case strings.Contains(s, "s3"), strings.Contains(s, "r2"), strings.Contains(s, "object"):
		return "s3"
	case strings.Contains(s, "url"), strings.Contains(s, "http"), strings.Contains(s, "external"), strings.Contains(s, "link"):
		return "url"
	case strings.Contains(s, "local"):
		return "local"
	}
	if isAbsoluteURL(strings.TrimSpace(imageURL)) {
		return "url"
	}
	return "local"
}

func loadSourceS3Setting(sourceDB *gorm.DB) (*settingModel.S3Setting, error) {
	type v3KeyValue struct {
		Key   string `gorm:"column:key"`
		Value string `gorm:"column:value"`
	}
	var kv v3KeyValue
	if err := sourceDB.Table("key_values").Where("key = ?", "s3_setting").Take(&kv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	raw := strings.TrimSpace(kv.Value)
	if raw == "" {
		return nil, nil
	}
	var setting settingModel.S3Setting
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

func buildExternalFileKey(rawURL string, category string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawURL)))
	return "external/" + strings.TrimSpace(category) + "/" + fmt.Sprintf("%x", sum[:])
}

func mapSourceS3SettingToV4(setting *settingModel.S3Setting) (settingModel.S3Setting, bool) {
	if setting == nil {
		return settingModel.S3Setting{}, false
	}
	mapped := settingModel.S3Setting{
		Enable:     true,
		Provider:   strings.ToLower(strings.TrimSpace(setting.Provider)),
		Endpoint:   strings.TrimSpace(setting.Endpoint),
		AccessKey:  strings.TrimSpace(setting.AccessKey),
		SecretKey:  strings.TrimSpace(setting.SecretKey),
		BucketName: strings.TrimSpace(setting.BucketName),
		Region:     strings.TrimSpace(setting.Region),
		UseSSL:     setting.UseSSL,
		CDNURL:     strings.TrimRight(strings.TrimSpace(setting.CDNURL), "/"),
		PathPrefix: strings.Trim(strings.TrimSpace(setting.PathPrefix), "/"),
		PublicRead: setting.PublicRead,
	}
	mapped.Endpoint = strings.TrimPrefix(mapped.Endpoint, "http://")
	mapped.Endpoint = strings.TrimPrefix(mapped.Endpoint, "https://")
	if mapped.Provider == "" || mapped.Endpoint == "" || mapped.BucketName == "" || mapped.AccessKey == "" || mapped.SecretKey == "" {
		return mapped, false
	}
	if mapped.Region == "" {
		mapped.Region = "auto"
	}
	return mapped, true
}

func normalizeObjectKeyForV4(objectKey string, pathPrefix string) string {
	clean := cleanMigratedFileKey(objectKey)
	if clean == "" {
		return ""
	}
	prefix := strings.Trim(strings.TrimSpace(pathPrefix), "/")
	if prefix != "" {
		prefixWithSlash := prefix + "/"
		clean = strings.TrimPrefix(clean, prefixWithSlash)
	}
	for _, route := range []string{"images/", "audios/", "videos/", "documents/", "files/"} {
		if strings.HasPrefix(clean, route) {
			clean = strings.TrimPrefix(clean, route)
			break
		}
	}
	return cleanMigratedFileKey(clean)
}

// ensureProtocol ensures the URL has a protocol prefix (http:// or https://).
// If the URL already has a protocol, it returns as-is.
// If no protocol, it adds https:// or http:// based on useSSL.
func ensureProtocol(url string, useSSL bool) string {
	url = strings.TrimSpace(url)
	lower := strings.ToLower(url)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return url
	}
	scheme := "https"
	if !useSSL {
		scheme = "http"
	}
	return scheme + "://" + url
}

func buildObjectURLFromSetting(setting settingModel.S3Setting, key string) string {
	resolvedKey := buildObjectStoragePath(setting, key)
	if resolvedKey == "" {
		return ""
	}
	cdn := strings.TrimRight(strings.TrimSpace(setting.CDNURL), "/")
	if cdn != "" {
		cdn = ensureProtocol(cdn, setting.UseSSL)
		return cdn + "/" + resolvedKey
	}
	endpoint := strings.TrimSpace(setting.Endpoint)
	if endpoint == "" {
		return ""
	}
	endpoint = ensureProtocol(endpoint, setting.UseSSL)
	endpoint = strings.TrimRight(endpoint, "/")
	bucket := strings.Trim(strings.TrimSpace(setting.BucketName), "/")
	if bucket == "" {
		return ""
	}
	return endpoint + "/" + bucket + "/" + resolvedKey
}

func buildObjectStoragePath(setting settingModel.S3Setting, key string) string {
	resolvedKey := strings.Trim(strings.TrimSpace(storage.NewFileSchema().Resolve(cleanMigratedFileKey(key))), "/")
	if resolvedKey == "" {
		return ""
	}
	if prefix := strings.Trim(strings.TrimSpace(setting.PathPrefix), "/"); prefix != "" {
		return prefix + "/" + resolvedKey
	}
	return resolvedKey
}

// deriveS3ObjectKeyMapping returns:
// - migrated business key for v4 files.key (schema prefix removed)
// - source object path candidates in bucket for fetching legacy objects
func deriveS3ObjectKeyMapping(
	objectKey string,
	imageURL string,
	pathPrefix string,
	bucketName string,
) (string, []string) {
	migratedKey := normalizeObjectKeyForV4(objectKey, pathPrefix)
	if migratedKey == "" {
		migratedKey = normalizeObjectKeyForV4(normalizeKeyFromURL(imageURL), pathPrefix)
	}
	if migratedKey == "" {
		return "", nil
	}
	return migratedKey, deriveSourceObjectStoragePathCandidates(objectKey, imageURL, pathPrefix, migratedKey, bucketName)
}

func deriveSourceObjectStoragePathCandidates(
	objectKey string,
	imageURL string,
	pathPrefix string,
	migratedKey string,
	bucketName string,
) []string {
	// Candidate strategy:
	// 1) raw v3 object_key and basename fallback
	// 2) parsed URL path, with/without bucket segment
	// 3) v4 target schema path variants, with/without path_prefix
	// 4) route-prefixed fallbacks (images/files/...)
	// This maximizes compatibility with v3 historical uploads.
	appendIf := func(dst []string, item string) []string {
		clean := cleanMigratedFileKey(item)
		if clean == "" {
			return dst
		}
		return append(dst, clean)
	}
	candidates := make([]string, 0, 10)
	rawObjectKey := cleanMigratedFileKey(objectKey)
	rawFromURL := cleanMigratedFileKey(normalizeKeyFromURL(imageURL))
	fullPathFromURL, strippedPathFromURL := parseObjectPathFromImageURL(imageURL, bucketName)
	prefix := strings.Trim(strings.TrimSpace(pathPrefix), "/")
	targetWithoutPrefix := strings.Trim(strings.TrimSpace(storage.NewFileSchema().Resolve(cleanMigratedFileKey(migratedKey))), "/")

	candidates = appendIf(candidates, rawObjectKey)
	candidates = appendIf(candidates, rawFromURL)
	candidates = appendIf(candidates, fullPathFromURL)
	candidates = appendIf(candidates, strippedPathFromURL)
	candidates = appendIf(candidates, targetWithoutPrefix)
	if prefix != "" {
		candidates = appendIf(candidates, prefix+"/"+rawObjectKey)
		candidates = appendIf(candidates, prefix+"/"+rawFromURL)
		candidates = appendIf(candidates, prefix+"/"+fullPathFromURL)
		candidates = appendIf(candidates, prefix+"/"+strippedPathFromURL)
		candidates = appendIf(candidates, prefix+"/"+targetWithoutPrefix)
		candidates = appendIf(candidates, strings.TrimPrefix(rawObjectKey, prefix+"/"))
		candidates = appendIf(candidates, strings.TrimPrefix(rawFromURL, prefix+"/"))
	}
	for _, route := range []string{"images/", "files/", "audios/", "videos/", "documents/"} {
		candidates = appendIf(candidates, route+rawObjectKey)
		candidates = appendIf(candidates, route+rawFromURL)
		candidates = appendIf(candidates, route+strippedPathFromURL)
	}
	return dedupePaths(candidates)
}

func parseObjectPathFromImageURL(imageURL string, bucketName string) (string, string) {
	raw := strings.TrimSpace(imageURL)
	if raw == "" {
		return "", ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		clean := cleanMigratedFileKey(raw)
		return clean, clean
	}
	fullPath := cleanMigratedFileKey(strings.TrimPrefix(u.Path, "/"))
	if fullPath == "" {
		return "", ""
	}
	bucket := strings.Trim(strings.TrimSpace(bucketName), "/")
	if bucket == "" {
		return fullPath, fullPath
	}
	stripped := strings.TrimPrefix(fullPath, bucket+"/")
	if stripped == fullPath {
		return fullPath, fullPath
	}
	return fullPath, cleanMigratedFileKey(stripped)
}

func buildStorageConfigFromS3Setting(setting settingModel.S3Setting) config.StorageConfig {
	cfg := config.Config().Storage
	cfg.ObjectEnabled = true
	cfg.Provider = strings.TrimSpace(setting.Provider)
	cfg.Endpoint = strings.TrimSpace(setting.Endpoint)
	cfg.AccessKey = strings.TrimSpace(setting.AccessKey)
	cfg.SecretKey = strings.TrimSpace(setting.SecretKey)
	cfg.BucketName = strings.TrimSpace(setting.BucketName)
	cfg.Region = strings.TrimSpace(setting.Region)
	cfg.UseSSL = setting.UseSSL
	cfg.CDNURL = strings.TrimSpace(setting.CDNURL)
	cfg.PathPrefix = strings.Trim(strings.TrimSpace(setting.PathPrefix), "/")
	return cfg
}

func buildRawObjectFSFromS3Setting(setting settingModel.S3Setting) (virefs.FS, error) {
	cfg := &virefs.S3Config{
		Provider:  mapVirefsProvider(setting.Provider),
		Endpoint:  normalizeS3Endpoint(setting.Endpoint, setting.UseSSL),
		AccessKey: strings.TrimSpace(setting.AccessKey),
		SecretKey: strings.TrimSpace(setting.SecretKey),
		Bucket:    strings.TrimSpace(setting.BucketName),
		Region:    normalizeS3Region(setting.Provider, setting.Region),
	}
	return virefs.NewObjectFSFromConfig(context.Background(), cfg)
}

func mapVirefsProvider(raw string) virefs.Provider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "minio":
		return virefs.ProviderMinIO
	case "r2", "cloudflare", "cloudflare-r2":
		return virefs.ProviderR2
	default:
		return virefs.ProviderAWS
	}
}

func normalizeS3Endpoint(endpoint string, useSSL bool) string {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return strings.TrimRight(trimmed, "/")
	}
	scheme := "http://"
	if useSSL {
		scheme = "https://"
	}
	return strings.TrimRight(scheme+trimmed, "/")
}

func normalizeS3Region(provider string, region string) string {
	p := strings.ToLower(strings.TrimSpace(provider))
	r := strings.TrimSpace(region)
	if p == "minio" && (r == "" || strings.EqualFold(r, "auto")) {
		return "us-east-1"
	}
	return r
}

func ensureS3ObjectAtSchemaPath(
	rawObjectFS virefs.FS,
	setting settingModel.S3Setting,
	businessKey string,
	sourcePathCandidates []string,
	sourceImageURL string,
) error {
	if rawObjectFS == nil {
		return errors.New("nil object fs")
	}
	targetPath := buildObjectStoragePath(setting, businessKey)
	if targetPath == "" {
		return errors.New("empty target object path")
	}
	if reader, err := rawObjectFS.Get(context.Background(), targetPath); err == nil {
		_ = reader.Close()
		return nil
	}
	var lastGetErr error
	var lastPutErr error
	for _, sourcePath := range sourcePathCandidates {
		reader, err := rawObjectFS.Get(context.Background(), sourcePath)
		if err != nil {
			lastGetErr = err
			continue
		}
		putErr := putObjectFromReadCloser(rawObjectFS, targetPath, reader)
		_ = reader.Close()
		if putErr == nil {
			return nil
		}
		lastPutErr = putErr
	}
	// Fallback for public-read legacy buckets:
	// if SDK-based object reads fail (credential/region mismatch), try source image URL directly.
	var fetchErr error
	if body, err := fetchObjectByURL(sourceImageURL); err == nil {
		putErr := putObjectFromReadCloser(rawObjectFS, targetPath, body)
		_ = body.Close()
		if putErr == nil {
			return nil
		}
		lastPutErr = putErr
	} else {
		fetchErr = err
	}
	return fmt.Errorf(
		"source object not found for key=%s candidates=%v last_get_err=%v last_put_err=%v source_url=%q source_url_err=%v",
		businessKey,
		sourcePathCandidates,
		lastGetErr,
		lastPutErr,
		strings.TrimSpace(sourceImageURL),
		fetchErr,
	)
}

func putObjectFromReadCloser(rawObjectFS virefs.FS, targetPath string, rc io.ReadCloser) error {
	if rc == nil {
		return errors.New("nil source reader")
	}
	buf, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	return rawObjectFS.Put(context.Background(), targetPath, bytes.NewReader(buf))
}

func fetchObjectByURL(rawURL string) (io.ReadCloser, error) {
	if !isAbsoluteURL(rawURL) {
		return nil, errors.New("invalid source image url")
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, strings.TrimSpace(rawURL), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected source url status=%d", resp.StatusCode)
	}
	return resp.Body, nil
}

func s3SettingToMap(setting settingModel.S3Setting) map[string]any {
	return map[string]any{
		"enable":      setting.Enable,
		"provider":    setting.Provider,
		"endpoint":    setting.Endpoint,
		"access_key":  setting.AccessKey,
		"secret_key":  setting.SecretKey,
		"bucket_name": setting.BucketName,
		"region":      setting.Region,
		"use_ssl":     setting.UseSSL,
		"cdn_url":     setting.CDNURL,
		"path_prefix": setting.PathPrefix,
		"public_read": setting.PublicRead,
	}
}

func cleanMigratedFileKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.Split(trimmed, "?")[0]
	return strings.Trim(trimmed, "/")
}

func resolveMigratedLocalPath(key string) string {
	clean := cleanMigratedFileKey(key)
	if clean == "" {
		return ""
	}
	for _, prefix := range []string{"images/", "audios/", "videos/", "documents/", "files/"} {
		if strings.HasPrefix(clean, prefix) {
			return clean
		}
	}
	return storage.NewFileSchema().Resolve(clean)
}

func inferFileName(key string, fileURL string) string {
	if base := path.Base(strings.Trim(strings.TrimSpace(key), "/")); base != "" && base != "." && base != "/" {
		return base
	}
	if base := path.Base(strings.Split(strings.TrimSpace(fileURL), "?")[0]); base != "" && base != "." && base != "/" {
		return base
	}
	return "file"
}

func inferContentType(primary string, secondary string) string {
	if ext := strings.ToLower(filepath.Ext(strings.TrimSpace(primary))); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	if ext := strings.ToLower(filepath.Ext(strings.TrimSpace(secondary))); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	return "application/octet-stream"
}

func isAbsoluteURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func resolveSourceImagePathCandidates(sourceRoot string, key string, imageURL string, targetPath string) []string {
	candidates := make([]string, 0, 8)
	appendIf := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		candidates = append(candidates, filepath.Clean(p))
	}
	cleanKey := cleanMigratedFileKey(key)
	if cleanKey != "" {
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(cleanKey)))
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(path.Base(cleanKey))))
	}
	rawURL := strings.TrimSpace(imageURL)
	rawURL = strings.Split(rawURL, "?")[0]
	trimmedURLPath := strings.Trim(strings.TrimPrefix(rawURL, "/"), "/")
	if trimmedURLPath != "" {
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(trimmedURLPath)))
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(path.Base(trimmedURLPath))))
	}
	targetPath = strings.Trim(strings.TrimSpace(targetPath), "/")
	if targetPath != "" {
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(targetPath)))
		appendIf(filepath.Join(sourceRoot, filepath.FromSlash(path.Base(targetPath))))
	}
	return dedupePaths(candidates)
}

func copySourceImageToLocalStorage(sourceRoot string, key string, imageURL string, targetPath string) error {
	targetPath = strings.Trim(strings.TrimSpace(targetPath), "/")
	if targetPath == "" {
		return errors.New("empty target path")
	}
	targetRoot := strings.TrimSpace(config.Config().Storage.DataRoot)
	if targetRoot == "" {
		targetRoot = "data/files"
	}
	targetFullPath := filepath.Join(targetRoot, filepath.FromSlash(targetPath))
	if info, err := os.Stat(targetFullPath); err == nil && !info.IsDir() {
		return nil
	}
	candidates := resolveSourceImagePathCandidates(sourceRoot, key, imageURL, targetPath)
	var lastErr error
	for _, src := range candidates {
		info, err := os.Stat(src)
		if err != nil || info.IsDir() {
			lastErr = err
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetFullPath), 0o755); err != nil {
			return err
		}
		if err := copyFileContents(src, targetFullPath); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("source file not found")
}

func dedupePaths(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, p := range input {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func copyFileContents(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func loadTagsByNames(tx *gorm.DB, names []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(names) == 0 {
		return result, nil
	}
	var tags []echoModel.Tag
	if err := tx.Where("name IN ?", names).Find(&tags).Error; err != nil {
		return nil, err
	}
	for i := range tags {
		result[tags[i].Name] = tags[i].ID
	}
	return result, nil
}

func fileRouteKey(storageType string, provider string, bucket string, key string) string {
	return strings.TrimSpace(storageType) + "|" + strings.TrimSpace(provider) + "|" + strings.TrimSpace(bucket) + "|" + strings.TrimSpace(key)
}

func loadExistingFileIDsByRoute(tx *gorm.DB, storageType string, provider string, bucket string, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result, nil
	}
	for start := 0; start < len(keys); start += dbBatchSize {
		end := start + dbBatchSize
		if end > len(keys) {
			end = len(keys)
		}
		var files []fileModel.File
		if err := tx.Where(
			"storage_type = ? AND provider = ? AND bucket = ? AND key IN ?",
			strings.TrimSpace(storageType),
			strings.TrimSpace(provider),
			strings.TrimSpace(bucket),
			keys[start:end],
		).Find(&files).Error; err != nil {
			return nil, err
		}
		for i := range files {
			result[files[i].Key] = files[i].ID
		}
	}
	return result, nil
}

func mapKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}

func stringifyIDMap(idMap map[int64]string) map[string]string {
	out := make(map[string]string, len(idMap))
	for oldID, newID := range idMap {
		out[strconv.FormatInt(oldID, 10)] = newID
	}
	return out
}

func loadEchoIDsWithImages(sourceDB *gorm.DB) (map[int64]struct{}, error) {
	var images []v3Image
	if err := sourceDB.Table("images").Select("message_id").Find(&images).Error; err != nil {
		return nil, err
	}
	out := make(map[int64]struct{}, len(images))
	for i := range images {
		out[images[i].MessageID] = struct{}{}
	}
	return out, nil
}

func exceedsFailureThreshold(processed int64, failed int64, threshold float64) bool {
	if processed == 0 {
		return false
	}
	return float64(failed)/float64(processed) > threshold
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	default:
		return 0
	}
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

type v3KeyValue struct {
	Key   string `json:"key"   gorm:"primaryKey"`
	Value string `json:"value"`
}

func (v *v3KeyValue) TableName() string {
	return "key_values"
}

const (
	// SystemSettingsKey 是系统设置的键
	SystemSettingsKey = "system_settings"
	// CommentSettingKey 是评论设置的键
	CommentSettingKey = "comment_setting"
	// S3SettingKey 是 S3 存储设置的键
	S3SettingKey = "s3_setting"
	// OAuth2SettingKey 是 OAuth2 设置的键
	OAuth2SettingKey = "oauth2_setting"
	// ServerURLKey 是服务器URL设置的键
	ServerURLKey = "server_url"
	// FediverseSettingKey 是联邦网络设置的键
	FediverseSettingKey = "fediverse_setting"
	// BackupScheduleKey 是备份计划设置的键
	BackupScheduleKey = "backup_schedule"
	// AgentSettingKey 是 Agent 设置的键
	AgentSettingKey = "agent_setting"
	// ReleaseVersionKey 是发布版本号的键
	ReleaseVersionKey = "release_version"
	// MigrationKey 是数据库迁移的标记键
	MigrationKey = "db_migration:message_to_echo:v1"
)

type v3SystemSetting struct {
	SiteTitle     string `json:"site_title"`     // 站点标题
	ServerLogo    string `json:"server_logo"`    // 服务器Logo
	ServerName    string `json:"server_name"`    // 服务器名称
	ServerURL     string `json:"server_url"`     // 服务器地址
	AllowRegister bool   `json:"allow_register"` // 是否允许注册'
	ICPNumber     string `json:"ICP_number"`     // 备案号
	MetingAPI     string `json:"meting_api"`     // Meting API 地址
	CustomCSS     string `json:"custom_css"`     // 自定义 CSS
	CustomJS      string `json:"custom_js"`      // 自定义 JS
}

func (v *v3SystemSetting) TableName() string {
	return "system_settings"
}

type v3OAuth2Setting struct {
	Enable       bool     `json:"enable"`        // 是否启用 OAuth2 登录
	Provider     string   `json:"provider"`      // OAuth2 提供商
	ClientID     string   `json:"client_id"`     // OAuth2 Client ID
	ClientSecret string   `json:"client_secret"` // OAuth2 Client Secret
	RedirectURI  string   `json:"redirect_uri"`  // OAuth2 重定向 URI
	Scopes       []string `json:"scopes"`        // OAuth2 请求的权限范围
	AuthURL      string   `json:"auth_url"`      // OAuth2 授权 URL
	TokenURL     string   `json:"token_url"`     // OAuth2 令牌 URL
	UserInfoURL  string   `json:"user_info_url"` // OAuth2 用户信息 URL

	// OIDC 扩展
	IsOIDC  bool   `json:"is_oidc"`  // 是否启用 OIDC
	Issuer  string `json:"issuer"`   // OIDC 颁发者
	JWKSURL string `json:"jwks_url"` // OIDC JWKS URL
}

func (v *v3OAuth2Setting) TableName() string {
	return "oauth2_settings"
}

type v3Connected struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ConnectURL string `                  json:"connect_url"` // 连接地址
}

func (v *v3Connected) TableName() string {
	return "connecteds"
}

type v3AgentSetting struct {
	Enable   bool   `json:"enable"`   // 是否启用 Agent 功能
	Provider string `json:"provider"` // LLM 提供商 （OpenAI、DeepSeek、Anthropic、Gemini、阿里百炼、Ollama等）
	Model    string `json:"model"`    // LLM 模型名称
	ApiKey   string `json:"api_key"`  // LLM API Key
	Prompt   string `json:"prompt"`   // Agent 额外使用的提示词
	BaseURL  string `json:"base_url"` // 自定义 API URL（可选）
}

func (v *v3AgentSetting) TableName() string {
	return "agent_settings"
}

type v3Webhook struct {
	ID          uint      `gorm:"primaryKey"   json:"id"`           // Webhook ID
	Name        string    `                    json:"name"`         // Webhook 名称
	URL         string    `                    json:"url"`          // Webhook URL
	Secret      string    `                    json:"secret"`       // 签名密钥，用于请求验证（HMAC等）
	IsActive    bool      `gorm:"default:true" json:"is_active"`    // 启用/禁用状态
	LastStatus  string    `                    json:"last_status"`  // 最近调用状态（如 success, failed）
	LastTrigger time.Time `                    json:"last_trigger"` // 最近触发时间
	CreatedAt   time.Time `                    json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `                    json:"updated_at"`   // 更新时间
}

func (v *v3Webhook) TableName() string {
	return "webhooks"
}

type v3BackupSchedule struct {
	Enable         bool   `json:"enable"`          // 是否启用备份计划
	CronExpression string `json:"cron_expression"` // 备份计划的 Cron 表达式
}

func (v *v3BackupSchedule) TableName() string {
	return "backup_schedules"
}

func migrateSettings(tx *gorm.DB, sourceDB *gorm.DB) error {
	// Migrate all settings from the source database into the target database

	var err error

	// 迁移 System 设置
	migrateSystemSettingErr := migrateSystemSetting(tx, sourceDB)
	if migrateSystemSettingErr != nil {
		logUtil.GetLogger().Warn("migration system setting failed", zap.Error(migrateSystemSettingErr))
		err = errors.Join(err, migrateSystemSettingErr)
	}

	// 迁移 OAuth2 设置
	migrateOAuth2SettingErr := migrateOAuth2Setting(tx, sourceDB)
	if migrateOAuth2SettingErr != nil {
		logUtil.GetLogger().Warn("migration oauth2 setting failed", zap.Error(migrateOAuth2SettingErr))
		err = errors.Join(err, migrateOAuth2SettingErr)
	}

	// 迁移 Connect 设置
	migrateConnectSettingErr := migrateConnectSetting(tx, sourceDB)
	if migrateConnectSettingErr != nil {
		logUtil.GetLogger().Warn("migration connect setting failed", zap.Error(migrateConnectSettingErr))
		err = errors.Join(err, migrateConnectSettingErr)
	}

	// 迁移 Agent 设置
	migrateAgentSettingErr := migrateAgentSetting(tx, sourceDB)
	if migrateAgentSettingErr != nil {
		logUtil.GetLogger().Warn("migration agent setting failed", zap.Error(migrateAgentSettingErr))
		err = errors.Join(err, migrateAgentSettingErr)
	}

	// 迁移 Webhook 设置
	migrateWebhookSettingErr := migrateWebhookSetting(tx, sourceDB)
	if migrateWebhookSettingErr != nil {
		logUtil.GetLogger().Warn("migration webhook setting failed", zap.Error(migrateWebhookSettingErr))
		err = errors.Join(err, migrateWebhookSettingErr)
	}

	// 迁移 Backup 设置
	migrateBackupSettingErr := migrateBackupSetting(tx, sourceDB)
	if migrateBackupSettingErr != nil {
		logUtil.GetLogger().Warn("migration backup setting failed", zap.Error(migrateBackupSettingErr))
		err = errors.Join(err, migrateBackupSettingErr)
	}

	return err
}

func migrateSystemSetting(tx *gorm.DB, sourceDB *gorm.DB) error {
	var v3kv v3KeyValue
	if err := sourceDB.Model(&v3KeyValue{}).
		Where("key = ?", SystemSettingsKey).
		First(&v3kv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	raw := strings.TrimSpace(v3kv.Value)
	if raw == "" {
		return nil
	}

	var v3Setting v3SystemSetting
	if err := json.Unmarshal([]byte(raw), &v3Setting); err != nil {
		return err
	}

	var systemSetting settingModel.SystemSetting
	systemSetting.SiteTitle = v3Setting.SiteTitle
	// TODO: Implement migration of server logo
	// systemSetting.ServerLogo = v3Setting.ServerLogo
	systemSetting.ServerName = v3Setting.ServerName
	systemSetting.ServerURL = v3Setting.ServerURL
	systemSetting.AllowRegister = v3Setting.AllowRegister
	systemSetting.ICPNumber = v3Setting.ICPNumber
	systemSetting.MetingAPI = v3Setting.MetingAPI
	systemSetting.CustomCSS = v3Setting.CustomCSS
	systemSetting.CustomJS = v3Setting.CustomJS

	data, err := json.Marshal(systemSetting)
	if err != nil {
		return err
	}
	dataString := string(data)

	systemSettingKV := commonModel.KeyValue{
		Key:   commonModel.SystemSettingsKey,
		Value: dataString,
	}
	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&systemSettingKV).Error; err != nil {
		return err
	}

	serverURLKV := commonModel.KeyValue{
		Key:   commonModel.ServerURLKey,
		Value: systemSetting.ServerURL,
	}
	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&serverURLKV).Error; err != nil {
		return err
	}

	return nil
}

func migrateOAuth2Setting(tx *gorm.DB, sourceDB *gorm.DB) error {
	var v3kv v3KeyValue
	if err := sourceDB.Model(&v3KeyValue{}).
		Where("key = ?", OAuth2SettingKey).
		First(&v3kv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	raw := strings.TrimSpace(v3kv.Value)
	if raw == "" {
		return nil
	}

	var v3Setting v3OAuth2Setting
	if err := json.Unmarshal([]byte(raw), &v3Setting); err != nil {
		return err
	}

	var oauth2Setting settingModel.OAuth2Setting
	oauth2Setting.Enable = v3Setting.Enable
	oauth2Setting.Provider = v3Setting.Provider
	oauth2Setting.ClientID = v3Setting.ClientID
	oauth2Setting.ClientSecret = v3Setting.ClientSecret
	oauth2Setting.RedirectURI = v3Setting.RedirectURI
	oauth2Setting.Scopes = v3Setting.Scopes
	oauth2Setting.AuthURL = v3Setting.AuthURL
	oauth2Setting.TokenURL = v3Setting.TokenURL
	oauth2Setting.UserInfoURL = v3Setting.UserInfoURL
	oauth2Setting.IsOIDC = v3Setting.IsOIDC
	oauth2Setting.Issuer = v3Setting.Issuer
	oauth2Setting.JWKSURL = v3Setting.JWKSURL

	data, err := json.Marshal(oauth2Setting)
	if err != nil {
		return err
	}
	dataString := string(data)

	oauth2SettingKV := commonModel.KeyValue{
		Key:   commonModel.OAuth2SettingKey,
		Value: dataString,
	}
	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&oauth2SettingKV).Error; err != nil {
		return err
	}
	return nil
}

func migrateConnectSetting(tx *gorm.DB, sourceDB *gorm.DB) error {
	var v3Connects []v3Connected
	if err := sourceDB.Model(&v3Connected{}).Find(&v3Connects).Error; err != nil {
		return err
	}
	v3ConnectURLs := make([]string, 0, len(v3Connects))
	for _, v3Connect := range v3Connects {
		v3ConnectURLs = append(v3ConnectURLs, v3Connect.ConnectURL)
	}

	var connects []connectModel.Connected
	if err := tx.Model(&connectModel.Connected{}).Find(&connects).Error; err != nil {
		return err
	}
	isExists := make(map[string]struct{}, len(connects))
	for _, connect := range connects {
		isExists[connect.ConnectURL] = struct{}{}
	}

	connectsToCreate := make([]connectModel.Connected, 0)
	for _, v3ConnectURL := range v3ConnectURLs {
		if _, ok := isExists[v3ConnectURL]; !ok {
			connectsToCreate = append(connectsToCreate, connectModel.Connected{
				ConnectURL: v3ConnectURL,
			})
		}
	}

	if len(connectsToCreate) > 0 {
		if err := tx.Create(&connectsToCreate).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateAgentSetting(tx *gorm.DB, sourceDB *gorm.DB) error {
	var v3kv v3KeyValue
	if err := sourceDB.Model(&v3KeyValue{}).
		Where("key = ?", AgentSettingKey).
		First(&v3kv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	raw := strings.TrimSpace(v3kv.Value)
	if raw == "" {
		return nil
	}

	var v3Setting v3AgentSetting
	if err := json.Unmarshal([]byte(raw), &v3Setting); err != nil {
		return err
	}

	var agentSetting settingModel.AgentSetting
	agentSetting.Enable = v3Setting.Enable
	agentSetting.Provider = v3Setting.Provider
	agentSetting.Model = v3Setting.Model
	agentSetting.ApiKey = v3Setting.ApiKey
	agentSetting.Prompt = v3Setting.Prompt
	agentSetting.BaseURL = v3Setting.BaseURL

	data, err := json.Marshal(agentSetting)
	if err != nil {
		return err
	}
	dataString := string(data)

	agentSettingKV := commonModel.KeyValue{
		Key:   commonModel.AgentSettingKey,
		Value: dataString,
	}
	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&agentSettingKV).Error; err != nil {
		return err
	}
	return nil
}

func migrateWebhookSetting(tx *gorm.DB, sourceDB *gorm.DB) error {
	var v3Webhooks []v3Webhook
	if err := sourceDB.Model(&v3Webhook{}).Find(&v3Webhooks).Error; err != nil {
		return err
	}

	var webhooks []webhookModel.Webhook
	for _, v3Webhook := range v3Webhooks {
		webhooks = append(webhooks, webhookModel.Webhook{
			Name:        v3Webhook.Name,
			URL:         v3Webhook.URL,
			Secret:      v3Webhook.Secret,
			IsActive:    v3Webhook.IsActive,
			LastStatus:  v3Webhook.LastStatus,
			LastTrigger: v3Webhook.LastTrigger,
			CreatedAt:   v3Webhook.CreatedAt,
			UpdatedAt:   v3Webhook.UpdatedAt,
		})
	}

	if len(webhooks) > 0 {
		if err := tx.Create(&webhooks).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateBackupSetting(tx *gorm.DB, sourceDB *gorm.DB) error {
	var v3kv v3KeyValue
	if err := sourceDB.Model(&v3KeyValue{}).
		Where("key = ?", BackupScheduleKey).
		First(&v3kv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	raw := strings.TrimSpace(v3kv.Value)
	if raw == "" {
		return nil
	}

	var v3Setting v3BackupSchedule
	if err := json.Unmarshal([]byte(raw), &v3Setting); err != nil {
		return err
	}

	var backupSchedule settingModel.BackupSchedule
	backupSchedule.Enable = v3Setting.Enable
	backupSchedule.CronExpression = v3Setting.CronExpression

	data, err := json.Marshal(backupSchedule)
	if err != nil {
		return err
	}
	dataString := string(data)

	backupScheduleKV := commonModel.KeyValue{
		Key:   commonModel.BackupScheduleKey,
		Value: dataString,
	}
	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&backupScheduleKV).Error; err != nil {
		return err
	}
	return nil
}
