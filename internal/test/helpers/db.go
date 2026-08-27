// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/lin-snow/ech0/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbSeq       atomic.Int64
	vecAutoOnce sync.Once
)

func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return newTestDB(t)
}

func NewTestDBWithVec(t *testing.T) *gorm.DB {
	t.Helper()
	vecAutoOnce.Do(sqlite_vec.Auto)
	return newTestDB(t)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:ech0test_%s_%d?mode=memory&cache=shared", dsnSafe(t.Name()), dbSeq.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("helpers: open in-memory sqlite: %v", err)
	}

	prev := currentDB()
	database.SetDB(db)
	t.Cleanup(func() {
		if sqlDB, derr := db.DB(); derr == nil {
			_ = sqlDB.Close()
		}
		database.SetDB(prev)
	})

	if err := database.MigrateDB(); err != nil {
		t.Fatalf("helpers: migrate in-memory db: %v", err)
	}
	return db
}

func currentDB() (db *gorm.DB) {
	defer func() { _ = recover() }()
	return database.GetDB()
}

func dsnSafe(name string) string {
	return strings.NewReplacer("/", "_", " ", "_", "#", "_", ":", "_").Replace(name)
}
