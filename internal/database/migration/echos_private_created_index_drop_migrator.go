// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type echosPrivateCreatedIndexDropMigrator struct{}

// NewEchosPrivateCreatedIndexDropMigrator removes the two-column timeline
// index. Keyset pagination orders by (created_at, id), so the index was
// replaced by idx_echos_private_created_id, which serves every query the old
// one did and sorts the id tiebreaker without a temporary b-tree.
func NewEchosPrivateCreatedIndexDropMigrator() Migrator {
	return &echosPrivateCreatedIndexDropMigrator{}
}

func (m *echosPrivateCreatedIndexDropMigrator) Name() string {
	return "echos_private_created_index_drop_migrator"
}

func (m *echosPrivateCreatedIndexDropMigrator) Key() string {
	return commonModel.EchosPrivateCreatedDroppedKey
}

func (m *echosPrivateCreatedIndexDropMigrator) CanRerun() bool {
	return false
}

func (m *echosPrivateCreatedIndexDropMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	// IF EXISTS keeps the migrator safe when a database never had the index.
	return db.Exec(`DROP INDEX IF EXISTS idx_echos_private_created`).Error
}
