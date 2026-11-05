// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"gorm.io/gorm"
)

type echosFavCountDropMigrator struct{}

// NewEchosFavCountDropMigrator removes the legacy like counter. Reactions are
// stored as per-actor rows now, so the old aggregate has no meaning.
func NewEchosFavCountDropMigrator() Migrator {
	return &echosFavCountDropMigrator{}
}

func (m *echosFavCountDropMigrator) Name() string {
	return "echos_fav_count_drop_migrator"
}

func (m *echosFavCountDropMigrator) Key() string {
	return commonModel.EchosFavCountColumnDroppedKey
}

func (m *echosFavCountDropMigrator) CanRerun() bool {
	return false
}

func (m *echosFavCountDropMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	if !db.Migrator().HasColumn(&echoModel.Echo{}, "fav_count") {
		return nil
	}
	return db.Exec(`ALTER TABLE echos DROP COLUMN fav_count`).Error
}
