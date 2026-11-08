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

func TestEchosPrivateCreatedIndexDropMigrator_DropsOnlyLegacyIndex(t *testing.T) {
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

	if !db.Migrator().HasIndex(&echoModel.Echo{}, "idx_echos_created_id") {
		t.Fatal("expected the created/id index to be created by AutoMigrate")
	}
	if !db.Migrator().HasIndex(&echoModel.Echo{}, "idx_echos_private_created_id") {
		t.Fatal("expected the private/created/id index to be created by AutoMigrate")
	}
	assertIndexColumns(t, db, "idx_echos_created_id", []string{"created_at", "id"})
	assertIndexColumns(t, db, "idx_echos_private_created_id", []string{"private", "created_at", "id"})

	// Simulate a database that still carries the pre-keyset index.
	if err := db.Exec(
		`CREATE INDEX idx_echos_private_created ON echos(private, created_at)`,
	).Error; err != nil {
		t.Fatalf("create legacy index failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewEchosPrivateCreatedIndexDropMigrator()),
	)

	if db.Migrator().HasIndex(&echoModel.Echo{}, "idx_echos_private_created") {
		t.Fatal("expected echos.idx_echos_private_created to be dropped")
	}
	if !db.Migrator().HasIndex(&echoModel.Echo{}, "idx_echos_private_created_id") {
		t.Fatal("the replacement index must survive the cleanup")
	}

	var marker commonModel.KeyValue
	if err := db.Where("key = ?", commonModel.EchosPrivateCreatedDroppedKey).First(&marker).Error; err != nil {
		t.Fatalf("expected migrator marker, got err: %v", err)
	}

	// Running it directly again after the index is gone must be a no-op.
	if err := dbMigration.NewEchosPrivateCreatedIndexDropMigrator().Migrate(db); err != nil {
		t.Fatalf("rerun should be a no-op, got %v", err)
	}
}

func assertIndexColumns(t *testing.T, db *gorm.DB, name string, want []string) {
	t.Helper()
	var rows []struct {
		Name string `gorm:"column:name"`
	}
	if err := db.Raw(fmt.Sprintf(`PRAGMA index_info(%q)`, name)).Scan(&rows).Error; err != nil {
		t.Fatalf("read index %s columns failed: %v", name, err)
	}
	got := make([]string, 0, len(rows))
	for _, row := range rows {
		got = append(got, row.Name)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("index %s columns = %v, want %v", name, got, want)
	}
}
