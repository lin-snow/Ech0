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

// TestUsersRoleMigrator_LegacyFlagsBecomeRoles replays the production sequence
// on a database that still has the boolean flags: the owner flag collapses into
// is_admin, then is_admin and the ownership marker become users.role. The
// earliest admin inherits the space and is_admin disappears.
func TestUsersRoleMigrator_LegacyFlagsBecomeRoles(t *testing.T) {
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

	// Simulate a pre-role database: both legacy boolean flags still exist.
	if err := db.Exec(`ALTER TABLE users ADD COLUMN is_owner BOOLEAN DEFAULT 0`).Error; err != nil {
		t.Fatalf("add legacy is_owner failed: %v", err)
	}
	if err := db.Exec(`ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT 0`).Error; err != nil {
		t.Fatalf("add legacy is_admin failed: %v", err)
	}
	for _, row := range []struct {
		id     string
		name   string
		admin  int
		owner  int
		role   string
		locale string
	}{
		{id: "u1", name: "first", admin: 1, owner: 0, locale: "zh-CN"},
		{id: "u2", name: "second", admin: 1, owner: 0, locale: "zh-CN"},
		{id: "u3", name: "member", admin: 0, owner: 0, locale: "zh-CN"},
	} {
		if err := db.Exec(
			`INSERT INTO users (id, username, is_admin, is_owner, locale) VALUES (?, ?, ?, ?, ?)`,
			row.id, row.name, row.admin, row.owner, row.locale,
		).Error; err != nil {
			t.Fatalf("insert %s failed: %v", row.name, err)
		}
	}

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(
			dbMigration.NewOwnerRoleCollapseMigrator(),
			dbMigration.NewUsersRoleMigrator(),
		),
	)

	roles := map[string]userModel.Role{}
	var rows []userModel.User
	if err := db.Table("users").Find(&rows).Error; err != nil {
		t.Fatalf("load users failed: %v", err)
	}
	for _, u := range rows {
		roles[u.Username] = u.Role
	}
	if roles["first"] != userModel.RoleOwner {
		t.Fatalf("earliest admin must own the space, got %q", roles["first"])
	}
	if roles["second"] != userModel.RoleAdmin {
		t.Fatalf("other admins must stay admins, got %q", roles["second"])
	}
	if roles["member"] != userModel.RoleMember {
		t.Fatalf("regular users must stay members, got %q", roles["member"])
	}

	if db.Migrator().HasColumn(&userModel.User{}, "is_admin") {
		t.Fatal("is_admin must be dropped once role exists")
	}

	// The partial unique index is the invariant, not an application convention.
	if err := db.Exec(
		`UPDATE users SET role = ? WHERE id = ?`, userModel.RoleOwner, "u2",
	).Error; err == nil {
		t.Fatal("a second owner must be rejected by the database")
	}
}
