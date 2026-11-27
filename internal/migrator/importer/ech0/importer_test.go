// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package ech0

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/artifact"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestImporterMigrate_SuccessAndIdempotent(t *testing.T) {
	tmpRoot := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	sourceTmpDir := filepath.Join(artifact.TmpRoot(), "ech0_v4_test")
	if err := os.MkdirAll(sourceTmpDir, 0o755); err != nil {
		t.Fatalf("mkdir source tmp dir failed: %v", err)
	}
	sourceDBPath := filepath.Join(sourceTmpDir, "ech0.db")
	sourceDB, err := gorm.Open(sqlite.Open(sourceDBPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open source db failed: %v", err)
	}
	if err := migrateLegacyTables(sourceDB); err != nil {
		t.Fatalf("migrate source tables failed: %v", err)
	}
	if err := seedSourceDB(sourceDB); err != nil {
		t.Fatalf("seed source db failed: %v", err)
	}
	sourceImagePath := filepath.Join(sourceTmpDir, "files", "images", "source.png")
	if err := os.MkdirAll(filepath.Dir(sourceImagePath), 0o755); err != nil {
		t.Fatalf("mkdir source image dir failed: %v", err)
	}
	if err := os.WriteFile(sourceImagePath, []byte("png-bytes"), 0o644); err != nil {
		t.Fatalf("write source image failed: %v", err)
	}

	targetDBPath := filepath.Join(tmpRoot, "target.db")
	targetDB, err := gorm.Open(sqlite.Open(targetDBPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open target db failed: %v", err)
	}
	if err := migrateTargetTables(targetDB); err != nil {
		t.Fatalf("migrate target tables failed: %v", err)
	}
	database.SetDB(targetDB)

	registry := storage.NewRegistry()
	if err := registry.Reload(context.Background(), []storage.Config{{
		ID:      "mem",
		Driver:  storage.DriverMem,
		Name:    "mem",
		Default: true,
	}}); err != nil {
		t.Fatalf("load storage registry failed: %v", err)
	}

	extractor := New(registry)
	req := spec.ImportRequest{
		SourcePayload: map[string]any{
			"tmp_dir": artifact.TmpDir + "/ech0_v4_test",
		},
	}

	firstResult, err := extractor.Import(context.Background(), req)
	if err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	if firstResult.SuccessCount != 1 || firstResult.FailCount != 0 {
		t.Fatalf("unexpected first migrate result: %+v", firstResult)
	}
	if _, ok := firstResult.Report["source_system_setting"]; !ok {
		t.Fatalf("expected source_system_setting in report")
	}
	if _, ok := firstResult.Report["source_comment_setting"]; !ok {
		t.Fatalf("expected source_comment_setting in report")
	}
	if _, ok := firstResult.Report["source_s3_setting"]; !ok {
		t.Fatalf("expected source_s3_setting in report")
	}
	if _, ok := firstResult.Report["source_oauth2_setting"]; !ok {
		t.Fatalf("expected source_oauth2_setting in report")
	}

	secondResult, err := extractor.Import(context.Background(), req)
	if err != nil {
		t.Fatalf("second import failed: %v", err)
	}
	if secondResult.SuccessCount != 1 || secondResult.FailCount != 0 {
		t.Fatalf("unexpected second migrate result: %+v", secondResult)
	}

	assertCount(t, targetDB, "echos", 1)
	assertCount(t, targetDB, "echo_extensions", 1)
	assertCount(t, targetDB, "tags", 1)
	assertCount(t, targetDB, "echo_tags", 1)
	assertCount(t, targetDB, "files", 1)
	assertCount(t, targetDB, "file_links", 1)
	assertCount(t, targetDB, "comments", 1)
	assertCount(t, targetDB, "echo_reactions", 1)

	var migrated fileModel.File
	if err := targetDB.First(&migrated).Error; err != nil {
		t.Fatalf("load migrated file failed: %v", err)
	}
	if migrated.Kind != string(fileModel.KindOwned) {
		t.Fatalf("unexpected migrated file kind: %q", migrated.Kind)
	}
	if migrated.Status != string(fileModel.StatusReady) {
		t.Fatalf("unexpected migrated file status: %q", migrated.Status)
	}
	if migrated.Access != string(fileModel.AccessOwner) {
		t.Fatalf("unexpected migrated file access: %q", migrated.Access)
	}
	if migrated.StorageID == nil || migrated.ObjectKey == nil {
		t.Fatalf("expected migrated file to be stored: %+v", migrated)
	}
	if migrated.Size != int64(len("png-bytes")) {
		t.Fatalf("unexpected migrated file size: got=%d", migrated.Size)
	}

	var link fileModel.Link
	if err := targetDB.First(&link).Error; err != nil {
		t.Fatalf("load migrated link failed: %v", err)
	}
	if link.FileID != migrated.ID || link.OwnerKind != fileModel.OwnerKindEcho ||
		link.OwnerID != "echo-source-1" || link.Role != fileModel.RoleAttachment {
		t.Fatalf("unexpected migrated link: %+v", link)
	}

	_, store, err := registry.Default()
	if err != nil {
		t.Fatalf("resolve default store failed: %v", err)
	}
	reader, _, err := store.Open(context.Background(), *migrated.ObjectKey)
	if err != nil {
		t.Fatalf("open migrated object failed: %v", err)
	}
	defer func() { _ = reader.Close() }()
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read migrated object failed: %v", err)
	}
	if string(got) != "png-bytes" {
		t.Fatalf("unexpected migrated object bytes: %q", string(got))
	}
}

func migrateLegacyTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&legacyFileRow{},
		&legacyEchoFileRow{},
		&commentModel.Comment{},
		&echoModel.Reaction{},
		&commonModel.KeyValue{},
	)
}

func migrateTargetTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&fileModel.File{},
		&fileModel.Link{},
		&commentModel.Comment{},
		&echoModel.Reaction{},
		&commonModel.KeyValue{},
	)
}

// legacyFileRow mirrors the pre-rewrite files table the snapshot carries.
type legacyFileRow struct {
	ID          string `gorm:"primaryKey;type:text"`
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
	Category    string
	UserID      string
	CreatedAt   int64
	UpdatedAt   int64
}

func (legacyFileRow) TableName() string { return "files" }

type legacyEchoFileRow struct {
	ID        string `gorm:"primaryKey;type:text"`
	EchoID    string
	FileID    string
	SortOrder int
}

func (legacyEchoFileRow) TableName() string { return "echo_files" }

func seedSourceDB(db *gorm.DB) error {
	now := time.Now().UTC().Unix()
	echo := echoModel.Echo{
		ID:        "echo-source-1",
		Content:   "hello ech0 v4",
		Username:  "tester",
		Layout:    echoModel.LayoutWaterfall,
		Private:   false,
		UserID:    "user-source-1",
		CreatedAt: now,
	}
	if err := db.Create(&echo).Error; err != nil {
		return err
	}
	if err := db.Create(&echoModel.Reaction{
		ID: "reaction-source-1", EchoID: echo.ID, Emoji: "like", Actor: "v:source",
	}).Error; err != nil {
		return err
	}
	ext := echoModel.EchoExtension{
		ID:        "ext-source-1",
		EchoID:    echo.ID,
		Type:      echoModel.Extension_WEBSITE,
		Payload:   map[string]any{"title": "Ech0", "site": "https://example.com"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&ext).Error; err != nil {
		return err
	}
	tag := echoModel.Tag{
		ID:         "tag-source-1",
		Name:       "v4",
		UsageCount: 1,
		CreatedAt:  now,
	}
	if err := db.Create(&tag).Error; err != nil {
		return err
	}
	if err := db.Create(&echoModel.EchoTag{EchoID: echo.ID, TagID: tag.ID}).Error; err != nil {
		return err
	}
	file := legacyFileRow{
		ID:          "file-source-1",
		Key:         "images/source.png",
		StorageType: "local",
		Provider:    "",
		Bucket:      "",
		URL:         "/api/files/images/source.png",
		Name:        "source.png",
		ContentType: "image/png",
		Size:        100,
		Width:       10,
		Height:      10,
		Category:    "image",
		UserID:      echo.UserID,
		CreatedAt:   now,
	}
	if err := db.Create(&file).Error; err != nil {
		return err
	}
	if err := db.Create(&legacyEchoFileRow{
		ID:        "echofile-source-1",
		EchoID:    echo.ID,
		FileID:    file.ID,
		SortOrder: 0,
	}).Error; err != nil {
		return err
	}
	comment := commentModel.Comment{
		ID:        "comment-source-1",
		EchoID:    echo.ID,
		Nickname:  "tester",
		Email:     "tester@example.com",
		Content:   "hello comment",
		Status:    commentModel.StatusApproved,
		Source:    commentModel.SourceSystem,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&comment).Error; err != nil {
		return err
	}

	systemRaw, _ := json.Marshal(map[string]any{"site_title": "Ech0 v4"})
	commentRaw, _ := json.Marshal(map[string]any{
		"enable_comment":   true,
		"require_approval": true,
		"captcha_enabled":  false,
		"captcha_verify":   "",
		"captcha_secret":   "",
	})
	s3Raw, _ := json.Marshal(map[string]any{
		"enable":      true,
		"provider":    "r2",
		"endpoint":    "example.r2.cloudflarestorage.com",
		"access_key":  "ak",
		"secret_key":  "sk",
		"bucket_name": "bucket-a",
		"region":      "auto",
		"use_ssl":     true,
	})
	oauthRaw, _ := json.Marshal(map[string]any{"enable": true, "provider": "github", "client_id": "id-a", "client_secret": "sec-a"})
	settings := []commonModel.KeyValue{
		{Key: commonModel.SystemSettingsKey, Value: string(systemRaw)},
		{Key: commentModel.CommentSystemSettingKey, Value: string(commentRaw)},
		{Key: commonModel.S3SettingKey, Value: string(s3Raw)},
		{Key: commonModel.OAuth2SettingKey, Value: string(oauthRaw)},
	}
	return db.Create(&settings).Error
}

func TestImporterMigrate_RefFileWithoutAttachmentRows(t *testing.T) {
	tmpRoot := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(tmpRoot); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	sourceTmpDir := filepath.Join(artifact.TmpRoot(), "ech0_ref_test")
	if err := os.MkdirAll(sourceTmpDir, 0o755); err != nil {
		t.Fatalf("mkdir source tmp dir failed: %v", err)
	}
	sourceDB, err := gorm.Open(sqlite.Open(filepath.Join(sourceTmpDir, "ech0.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open source db failed: %v", err)
	}
	if err := migrateLegacyTables(sourceDB); err != nil {
		t.Fatalf("migrate source tables failed: %v", err)
	}
	if err := sourceDB.Migrator().DropTable("echo_files"); err != nil {
		t.Fatalf("drop source echo_files failed: %v", err)
	}

	now := time.Now().UTC().Unix()
	if err := sourceDB.Create(&echoModel.Echo{
		ID:        "echo-ref-1",
		Content:   "external image",
		Username:  "tester",
		Layout:    echoModel.LayoutWaterfall,
		UserID:    "user-ref-1",
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed source echo failed: %v", err)
	}
	if err := sourceDB.Create(&legacyFileRow{
		ID:          "legacy-ref-file",
		StorageType: "s3",
		Provider:    "r2",
		Bucket:      "bucket-a",
		URL:         "https://cdn.example.com/a.png",
		Name:        "a.png",
		ContentType: "image/png",
		Size:        42,
		Width:       1,
		Height:      1,
		Category:    "image",
		UserID:      "user-ref-1",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("seed source file failed: %v", err)
	}

	targetDB, err := gorm.Open(sqlite.Open(filepath.Join(tmpRoot, "target.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open target db failed: %v", err)
	}
	if err := migrateTargetTables(targetDB); err != nil {
		t.Fatalf("migrate target tables failed: %v", err)
	}
	database.SetDB(targetDB)

	registry := storage.NewRegistry()
	if err := registry.Reload(context.Background(), []storage.Config{{
		ID:      "mem",
		Driver:  storage.DriverMem,
		Name:    "mem",
		Default: true,
	}}); err != nil {
		t.Fatalf("load storage registry failed: %v", err)
	}

	extractor := New(registry)
	req := spec.ImportRequest{
		SourcePayload: map[string]any{
			"tmp_dir": artifact.TmpDir + "/ech0_ref_test",
		},
	}
	for i := range 2 {
		result, err := extractor.Import(context.Background(), req)
		if err != nil {
			t.Fatalf("import %d failed: %v", i+1, err)
		}
		if result.SuccessCount != 1 || result.FailCount != 0 {
			t.Fatalf("unexpected import %d result: %+v", i+1, result)
		}
	}

	assertCount(t, targetDB, "files", 1)
	assertCount(t, targetDB, "file_links", 0)

	var migrated fileModel.File
	if err := targetDB.First(&migrated).Error; err != nil {
		t.Fatalf("load migrated file failed: %v", err)
	}
	if migrated.Kind != string(fileModel.KindRef) {
		t.Fatalf("unexpected migrated file kind: %q", migrated.Kind)
	}
	if migrated.Status != string(fileModel.StatusReady) {
		t.Fatalf("unexpected migrated file status: %q", migrated.Status)
	}
	if migrated.StorageID != nil || migrated.ObjectKey != nil {
		t.Fatalf("ref file must not carry storage columns: %+v", migrated)
	}
	if migrated.RefURL == nil || *migrated.RefURL != "https://cdn.example.com/a.png" {
		t.Fatalf("unexpected ref url: %+v", migrated.RefURL)
	}
}

func assertCount(t *testing.T, db *gorm.DB, table string, expected int64) {
	t.Helper()
	var count int64
	if err := db.Table(table).Count(&count).Error; err != nil {
		t.Fatalf("count %s failed: %v", table, err)
	}
	if count != expected {
		t.Fatalf("unexpected %s count: got=%d want=%d", table, count, expected)
	}
}
