// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"errors"
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

type usersRoleMigrator struct{}

// NewUsersRoleMigrator replaces the boolean is_admin flag with a role column
// that distinguishes the owner, admins and members. The owner is the earliest
// admin — the account that would have been treated as the primary admin by the
// previous model — and uniqueness is enforced by a partial index so a second
// owner can never be written.
func NewUsersRoleMigrator() Migrator {
	return &usersRoleMigrator{}
}

func (m *usersRoleMigrator) Name() string {
	return "users_role_migrator"
}

func (m *usersRoleMigrator) Key() string {
	return commonModel.UsersRoleMigratedKey
}

func (m *usersRoleMigrator) CanRerun() bool {
	return false
}

func (m *usersRoleMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if db.Migrator().HasColumn(&userModel.User{}, "is_admin") {
		if err := db.Exec(`UPDATE users SET role = ? WHERE is_admin = ?`, userModel.RoleAdmin, true).Error; err != nil {
			return err
		}
	}

	if err := m.ensureOwner(db); err != nil {
		return err
	}

	if db.Migrator().HasColumn(&userModel.User{}, "is_admin") {
		if err := db.Exec(`ALTER TABLE users DROP COLUMN is_admin`).Error; err != nil {
			return err
		}
	}

	return db.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_users_single_owner ON users(role) WHERE role = 'owner'`,
	).Error
}

// ensureOwner promotes the earliest admin when no owner exists yet. A database
// that already has an owner (or has no admin at all) is left untouched.
func (m *usersRoleMigrator) ensureOwner(db *gorm.DB) error {
	var owners int64
	if err := db.Model(&userModel.User{}).
		Where("role = ?", userModel.RoleOwner).
		Count(&owners).Error; err != nil {
		return err
	}
	if owners > 0 {
		return nil
	}

	var candidate userModel.User
	err := db.Where("role = ?", userModel.RoleAdmin).
		Order("id ASC").
		First(&candidate).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return db.Model(&userModel.User{}).
		Where("id = ? AND role = ?", candidate.ID, userModel.RoleAdmin).
		Update("role", userModel.RoleOwner).Error
}
