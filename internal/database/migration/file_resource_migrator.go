// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// fileResourceMigrator rebuilds the file tables as the five-table resource
// schema. It runs once per database: AutoMigrate cannot add CHECK constraints
// or partial unique indexes to SQLite, and the legacy file tables (key/class
// columns, echo_files, instance_assets) have no meaning under the new model.
//
// Only the constraint-enforcing shape needs DDL; column additions stay with
// AutoMigrate, which runs before this migrator on every boot. A row written
// before this migrator runs cannot exist — the server writes files only after
// InitDatabase returns.
type fileResourceMigrator struct{}

func NewFileResourceMigrator() Migrator {
	return &fileResourceMigrator{}
}

func (m *fileResourceMigrator) Name() string {
	return "file_resource_migrator"
}

func (m *fileResourceMigrator) Key() string {
	return "files.resource.v1"
}

func (m *fileResourceMigrator) CanRerun() bool {
	return false
}

func (m *fileResourceMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	for _, stmt := range fileResourceStatements() {
		if err := db.Exec(indentSafe(stmt)).Error; err != nil {
			return fmt.Errorf("%w: %s", err, stmt)
		}
	}
	return nil
}

// indentSafe collapses tab indentation to spaces. gorm's sqlite migrator
// parses table DDL to introspect columns and treats a tab as a quote character,
// so a tab-indented CHECK constraint makes AutoMigrate fail on the next
// restart — the table exists, the server cannot boot.
func indentSafe(stmt string) string {
	return strings.ReplaceAll(stmt, "\t", "  ")
}

func fileResourceStatements() []string {
	return []string{
		// Legacy tables and their rows are not representable in the new model.
		`DROP TABLE IF EXISTS echo_files`,
		`DROP TABLE IF EXISTS instance_assets`,
		`DROP TABLE IF EXISTS temp_files`,
		`DROP TABLE IF EXISTS file_blobs`,
		`DROP TABLE IF EXISTS file_links`,
		`DROP TABLE IF EXISTS file_gc`,
		`DROP TABLE IF EXISTS files`,
		`DROP TABLE IF EXISTS storages`,

		`CREATE TABLE files (
			id          char(36) PRIMARY KEY,
			kind        varchar(10) NOT NULL,
			status      varchar(10) NOT NULL,
			access      varchar(10) NOT NULL,
			storage_id  char(36) NULL,
			object_key  varchar(500) NULL,
			blob_bytes  integer NULL,
			ref_url     text NULL,
			name        varchar(255) NOT NULL DEFAULT '',
			content_type varchar(100) NOT NULL,
			size        integer NOT NULL DEFAULT 0,
			width       integer NOT NULL DEFAULT 0,
			height      integer NOT NULL DEFAULT 0,
			uploader_id char(36) NOT NULL,
			created_at  integer NOT NULL DEFAULT 0,
			updated_at  integer NOT NULL DEFAULT 0,
			CONSTRAINT chk_files_kind   CHECK (kind IN ('owned','blob','ref')),
			CONSTRAINT chk_files_status CHECK (status IN ('pending','ready')),
			CONSTRAINT chk_files_access CHECK (access IN ('public','owner','private')),
			CONSTRAINT chk_files_locator CHECK (
				(kind = 'owned' AND storage_id IS NOT NULL AND object_key IS NOT NULL AND ref_url IS NULL AND blob_bytes IS NULL) OR
				(kind = 'blob'  AND blob_bytes IS NOT NULL AND storage_id IS NULL AND object_key IS NULL AND ref_url IS NULL) OR
				(kind = 'ref'   AND ref_url IS NOT NULL AND storage_id IS NULL AND object_key IS NULL AND blob_bytes IS NULL)
			)
		)`,
		`CREATE TABLE file_blobs (
			file_id char(36) PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
			data    blob NOT NULL
		)`,
		`CREATE TABLE file_links (
			file_id    char(36) NOT NULL REFERENCES files(id) ON DELETE CASCADE,
			owner_kind varchar(20) NOT NULL,
			owner_id   char(36) NOT NULL,
			role       varchar(20) NOT NULL,
			sort_order integer NOT NULL DEFAULT 0,
			created_at integer NOT NULL DEFAULT 0,
			PRIMARY KEY (file_id, owner_kind, owner_id, role, sort_order)
		)`,
		`CREATE TABLE file_gc (
			id          char(36) PRIMARY KEY,
			storage_id  char(36) NOT NULL,
			object_key  varchar(500) NOT NULL,
			attempts    integer NOT NULL DEFAULT 0,
			last_error  text NOT NULL DEFAULT '',
			next_try_at integer NOT NULL,
			created_at  integer NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE storages (
			id         char(36) PRIMARY KEY,
			driver     varchar(20) NOT NULL,
			name       varchar(100) NOT NULL,
			config     text NOT NULL,
			is_default boolean NOT NULL DEFAULT false,
			created_at integer NOT NULL DEFAULT 0,
			updated_at integer NOT NULL DEFAULT 0,
			CONSTRAINT uq_storages_driver_name UNIQUE (driver, name)
		)`,

		`CREATE UNIQUE INDEX idx_files_locator ON files(storage_id, object_key) WHERE kind = 'owned'`,
		`CREATE INDEX idx_files_status_created ON files(status, created_at)`,
		`CREATE INDEX idx_files_uploader ON files(uploader_id)`,
		`CREATE INDEX idx_file_links_owner ON file_links(owner_kind, owner_id, role)`,
		`CREATE UNIQUE INDEX uq_file_links_slot ON file_links(owner_kind, owner_id, role)
			WHERE role IN ('cover','avatar','logo','favicon','og_image')`,
		`CREATE INDEX idx_file_gc_due ON file_gc(next_try_at)`,
		`CREATE UNIQUE INDEX uq_storages_default ON storages(is_default) WHERE is_default = true`,
	}
}
