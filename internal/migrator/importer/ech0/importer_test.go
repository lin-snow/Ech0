// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package ech0

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/migrator/spec"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
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

	sourceTmpDir := filepath.Join("data", "files", "tmp", "ech0_v4_test")
	if err := os.MkdirAll(sourceTmpDir, 0o755); err != nil {
		t.Fatalf("mkdir source tmp dir failed: %v", err)
	}
	sourceDBPath := filepath.Join(sourceTmpDir, "ech0.db")
	sourceDB, err := gorm.Open(sqlite.Open(sourceDBPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open source db failed: %v", err)
	}
	if err := migrateBaseTables(sourceDB); err != nil {
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
	if err := migrateBaseTables(targetDB); err != nil {
		t.Fatalf("migrate target tables failed: %v", err)
	}
	database.SetDB(targetDB)

	extractor := New()
	req := spec.ImportRequest{
		SourcePayload: map[string]any{
			"tmp_dir": "files/tmp/ech0_v4_test",
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
	assertCount(t, targetDB, "echo_files", 1)
	assertCount(t, targetDB, "comments", 1)
	if _, err := os.Stat(filepath.Join("data", "files", "images", "source.png")); err != nil {
		t.Fatalf("expected migrated source image exists: %v", err)
	}
}

func migrateBaseTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&fileModel.File{},
		&fileModel.EchoFile{},
		&commentModel.Comment{},
		&commonModel.KeyValue{},
	)
}

func seedSourceDB(db *gorm.DB) error {
	now := time.Now().UTC().Unix()
	echo := echoModel.Echo{
		ID:        "echo-source-1",
		Content:   "hello ech0 v4",
		Username:  "tester",
		Layout:    echoModel.LayoutWaterfall,
		Private:   false,
		UserID:    "user-source-1",
		FavCount:  3,
		CreatedAt: now,
	}
	if err := db.Create(&echo).Error; err != nil {
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
	file := fileModel.File{
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
	if err := db.Create(&fileModel.EchoFile{
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
