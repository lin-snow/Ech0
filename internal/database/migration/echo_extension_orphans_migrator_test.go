// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration_test

import (
	"fmt"
	"testing"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestEchoExtensionOrphansMigrator_RemovesOrphansAndMarkDone(t *testing.T) {
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

	if err := db.Exec(
		`INSERT INTO echos (id, content, user_id, private, created_at) VALUES ('e1', 'alive', 'u1', false, 100)`,
	).Error; err != nil {
		t.Fatalf("insert echo failed: %v", err)
	}
	for _, stmt := range []string{
		`INSERT INTO echo_extensions (id, echo_id, type, payload, created_at, updated_at) VALUES ('ext-alive', 'e1', 'demo', '{}', 100, 100)`,
		`INSERT INTO echo_extensions (id, echo_id, type, payload, created_at, updated_at) VALUES ('ext-orphan', 'ghost', 'demo', '{}', 100, 100)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("insert extension failed: %v", err)
		}
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewEchoExtensionOrphansMigrator()),
	)

	var ids []string
	if err := db.Raw("SELECT id FROM echo_extensions ORDER BY id").Scan(&ids).Error; err != nil {
		t.Fatalf("query extensions failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != "ext-alive" {
		t.Fatalf("expected only ext-alive to survive, got %v", ids)
	}

	var marker commonModel.KeyValue
	if err := db.Where("key = ?", commonModel.EchoExtensionOrphansCleanedKey).First(&marker).Error; err != nil {
		t.Fatalf("expected migrator marker, got err: %v", err)
	}
}
