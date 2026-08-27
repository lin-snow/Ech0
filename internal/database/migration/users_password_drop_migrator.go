// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

type usersPasswordDropMigrator struct{}

func NewUsersPasswordDropMigrator() Migrator {
	return &usersPasswordDropMigrator{}
}

func (m *usersPasswordDropMigrator) Name() string {
	return "users_password_drop_migrator"
}

func (m *usersPasswordDropMigrator) Key() string {
	return commonModel.UsersPasswordColumnDroppedKey
}

func (m *usersPasswordDropMigrator) CanRerun() bool {
	return false
}

func (m *usersPasswordDropMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	if !db.Migrator().HasColumn(&userModel.User{}, "password") {
		return nil
	}
	return db.Exec(`ALTER TABLE users DROP COLUMN password`).Error
}
