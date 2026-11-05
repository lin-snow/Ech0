// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration_test

import (
	"fmt"
	"testing"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestEchosFavCountDropMigrator_DropsColumnKeepsRows(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	database.SetDB(db)
	if err := database.MigrateDB(); err != nil {
		t.Fatalf("migrate db failed: %v", err)
	}

	// Simulate a pre-reaction database: the legacy counter still exists.
	if err := db.Exec(`ALTER TABLE echos ADD COLUMN fav_count INTEGER DEFAULT 0`).Error; err != nil {
		t.Fatalf("add legacy fav_count failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO echos (id, content, username, private, user_id, fav_count, created_at) VALUES ('e1', 'kept', 'tester', 0, 'u1', 7, 100)`,
	).Error; err != nil {
		t.Fatalf("insert echo failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewEchosFavCountDropMigrator()),
	)

	if db.Migrator().HasColumn(&echoModel.Echo{}, "fav_count") {
		t.Fatal("expected echos.fav_count to be dropped")
	}

	var count int64
	if err := db.Table("echos").Where("id = ?", "e1").Count(&count).Error; err != nil {
		t.Fatalf("count echos failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected the echo row to survive, got %d rows", count)
	}

	var marker commonModel.KeyValue
	if err := db.Where("key = ?", commonModel.EchosFavCountColumnDroppedKey).First(&marker).Error; err != nil {
		t.Fatalf("expected migrator marker, got err: %v", err)
	}

	// Running it again after the column is gone must be a no-op.
	if err := dbMigration.NewEchosFavCountDropMigrator().Migrate(db); err != nil {
		t.Fatalf("rerun should be a no-op, got %v", err)
	}
}
