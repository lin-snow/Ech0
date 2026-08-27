// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type echoExtensionOrphansMigrator struct{}

func NewEchoExtensionOrphansMigrator() Migrator {
	return &echoExtensionOrphansMigrator{}
}

func (m *echoExtensionOrphansMigrator) Name() string {
	return "echo_extension_orphans_migrator"
}

func (m *echoExtensionOrphansMigrator) Key() string {
	return commonModel.EchoExtensionOrphansCleanedKey
}

func (m *echoExtensionOrphansMigrator) CanRerun() bool {
	return false
}

func (m *echoExtensionOrphansMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Exec(
		`DELETE FROM echo_extensions WHERE echo_id NOT IN (SELECT id FROM echos)`,
	).Error
}
