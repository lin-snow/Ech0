// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration_test

import (
	"fmt"
	"testing"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOwnerRoleCollapseMigrator_PromotesOwnerToAdmin(t *testing.T) {
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

	// Simulate a pre-two-role database: the removed owner flag still exists.
	if err := db.Exec(`ALTER TABLE users ADD COLUMN is_owner BOOLEAN DEFAULT 0`).Error; err != nil {
		t.Fatalf("add legacy is_owner failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO users (id, username, is_admin, is_owner, locale) VALUES ('u1', 'owner', 1, 1, 'zh-CN')`,
	).Error; err != nil {
		t.Fatalf("insert owner failed: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO users (id, username, is_admin, is_owner, locale) VALUES ('u2', 'member', 0, 0, 'zh-CN')`,
	).Error; err != nil {
		t.Fatalf("insert member failed: %v", err)
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewOwnerRoleCollapseMigrator()),
	)

	users := map[string]userModel.User{}
	var rows []userModel.User
	if err := db.Table("users").Find(&rows).Error; err != nil {
		t.Fatalf("load users failed: %v", err)
	}
	for _, u := range rows {
		users[u.Username] = u
	}
	if !users["owner"].IsAdmin {
		t.Fatal("expected the owner account to keep admin rights")
	}
	if users["member"].IsAdmin {
		t.Fatal("expected a regular account to stay non-admin")
	}
}
