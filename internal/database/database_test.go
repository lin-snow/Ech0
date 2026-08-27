// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"path/filepath"
	"testing"
	"time"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBuildGormConfig_UsesUTCNowFunc(t *testing.T) {
	cfg := buildGormConfig(logger.Silent)
	if cfg.NowFunc != nil {
		t.Fatalf("expected nil NowFunc, got configured function")
	}
}

func TestOpenSQLite_AppliesConnectionParams(t *testing.T) {
	db, err := openSQLite(filepath.Join(t.TempDir(), "ech0.db"), logger.Silent)
	if err != nil {
		t.Fatalf("openSQLite failed: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	})

	var journalMode string
	if err := db.Raw("PRAGMA journal_mode").Scan(&journalMode).Error; err != nil {
		t.Fatalf("query journal_mode failed: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("expected journal_mode wal, got %q", journalMode)
	}

	var busyTimeout int
	if err := db.Raw("PRAGMA busy_timeout").Scan(&busyTimeout).Error; err != nil {
		t.Fatalf("query busy_timeout failed: %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("expected busy_timeout 5000, got %d", busyTimeout)
	}

	var synchronous int
	if err := db.Raw("PRAGMA synchronous").Scan(&synchronous).Error; err != nil {
		t.Fatalf("query synchronous failed: %v", err)
	}
	if synchronous != 1 {
		t.Fatalf("expected synchronous NORMAL(1), got %d", synchronous)
	}
}

func TestSnapshotTo_ProducesConsistentCopy(t *testing.T) {
	dir := t.TempDir()
	db, err := openSQLite(filepath.Join(dir, "ech0.db"), logger.Silent)
	if err != nil {
		t.Fatalf("openSQLite failed: %v", err)
	}
	SetDB(db)

	if err := db.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT)").Error; err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	if err := db.Exec("INSERT INTO t (v) VALUES ('hello')").Error; err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	dst := filepath.Join(dir, "copy.db")
	if err := SnapshotTo(dst); err != nil {
		t.Fatalf("SnapshotTo failed: %v", err)
	}

	copyDB, err := gorm.Open(sqlite.Open(dst), &gorm.Config{})
	if err != nil {
		t.Fatalf("open copy failed: %v", err)
	}
	var v string
	if err := copyDB.Raw("SELECT v FROM t LIMIT 1").Scan(&v).Error; err != nil {
		t.Fatalf("query copy failed: %v", err)
	}
	if v != "hello" {
		t.Fatalf("expected copied row 'hello', got %q", v)
	}
}

func TestMigrateDB_AccessTokenSettingIncludesScopeFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	SetDB(db)

	if err := MigrateDB(); err != nil {
		t.Fatalf("migrate db failed: %v", err)
	}

	expiry := time.Now().UTC().Add(8 * time.Hour).Unix()
	lastUsed := time.Now().UTC().Unix()
	record := settingModel.AccessTokenSetting{
		UserID:     "user-1",
		Token:      "token-1",
		Name:       "test-token",
		TokenType:  "access",
		Scopes:     `["echo:read"]`,
		Audience:   "public-client",
		JTI:        "jti-1",
		Expiry:     &expiry,
		LastUsedAt: &lastUsed,
		CreatedAt:  time.Now().UTC().Unix(),
	}

	if err := GetDB().Create(&record).Error; err != nil {
		t.Fatalf("create access token failed: %v", err)
	}

	var got settingModel.AccessTokenSetting
	if err := GetDB().Where("id = ?", record.ID).First(&got).Error; err != nil {
		t.Fatalf("query access token failed: %v", err)
	}
	if got.TokenType != "access" {
		t.Fatalf("expected token type access, got %s", got.TokenType)
	}
	if got.Scopes == "" {
		t.Fatal("expected scopes not empty")
	}
	if got.Audience != "public-client" {
		t.Fatalf("expected audience public-client, got %s", got.Audience)
	}
	if got.JTI == "" {
		t.Fatal("expected jti not empty")
	}
	if got.LastUsedAt == nil {
		t.Fatal("expected last_used_at to be set")
	}
}
