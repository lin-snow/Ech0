// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

type userLocalAuthBackfillMigrator struct{}

func NewUserLocalAuthBackfillMigrator() Migrator {
	return &userLocalAuthBackfillMigrator{}
}

func (m *userLocalAuthBackfillMigrator) Name() string {
	return "user_local_auth_backfill_migrator"
}

func (m *userLocalAuthBackfillMigrator) Key() string {
	return commonModel.UserLocalAuthBackfilledKey
}

func (m *userLocalAuthBackfillMigrator) CanRerun() bool {
	return false
}

func (m *userLocalAuthBackfillMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	if !db.Migrator().HasColumn(&userModel.User{}, "password") {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`
			INSERT OR IGNORE INTO user_local_auth (user_id, password_hash, password_algo, updated_at)
			SELECT id, password, 'md5', CAST(strftime('%s','now') AS INTEGER)
			FROM users
			WHERE password IS NOT NULL AND password <> ''
		`).Error
	})
}
