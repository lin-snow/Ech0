// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type oauthBindingsDropMigrator struct{}

func NewOAuthBindingsDropMigrator() Migrator {
	return &oauthBindingsDropMigrator{}
}

func (m *oauthBindingsDropMigrator) Name() string {
	return "oauth_bindings_drop_migrator"
}

func (m *oauthBindingsDropMigrator) Key() string {
	return commonModel.OAuthBindingsDroppedKey
}

func (m *oauthBindingsDropMigrator) CanRerun() bool {
	return false
}

func (m *oauthBindingsDropMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	for _, stmt := range []string{
		`DROP TABLE IF EXISTS oauth_bindings`,
		`DROP TABLE IF EXISTS o_auth_bindings`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

type legacyInboxesDropMigrator struct{}

func NewLegacyInboxesDropMigrator() Migrator {
	return &legacyInboxesDropMigrator{}
}

func (m *legacyInboxesDropMigrator) Name() string {
	return "legacy_inboxes_drop_migrator"
}

func (m *legacyInboxesDropMigrator) Key() string {
	return commonModel.LegacyInboxesDroppedKey
}

func (m *legacyInboxesDropMigrator) CanRerun() bool {
	return false
}

func (m *legacyInboxesDropMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	for _, stmt := range []string{
		`DROP TABLE IF EXISTS inboxes`,
		`DROP TABLE IF EXISTS inbox`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}
