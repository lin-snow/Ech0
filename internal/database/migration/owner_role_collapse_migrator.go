// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

type ownerRoleCollapseMigrator struct{}

// NewOwnerRoleCollapseMigrator folds the removed owner role into the admin
// role. Owner accounts were always admins, so only the flag has to be carried
// over before the column stops being written.
func NewOwnerRoleCollapseMigrator() Migrator {
	return &ownerRoleCollapseMigrator{}
}

func (m *ownerRoleCollapseMigrator) Name() string {
	return "owner_role_collapse_migrator"
}

func (m *ownerRoleCollapseMigrator) Key() string {
	return commonModel.OwnerRoleCollapsedKey
}

func (m *ownerRoleCollapseMigrator) CanRerun() bool {
	return false
}

func (m *ownerRoleCollapseMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	if !db.Migrator().HasColumn(&userModel.User{}, "is_owner") {
		return nil
	}
	return db.Exec(`UPDATE users SET is_admin = ? WHERE is_owner = ?`, true, true).Error
}
