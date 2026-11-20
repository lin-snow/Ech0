// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConvertSQLite builds a new SQLite file suitable for Turso's --from-file.
// VACUUM INTO takes a consistent snapshot including committed WAL pages.
func ConvertSQLite(ctx context.Context, source, output string) error {
	if err := enableLocalSQLite(); err != nil {
		return err
	}
	path, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	uri := url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro"}
	src, err := gorm.Open(sqlite.Open(uri.String()), buildGormConfig(logger.Silent))
	if err != nil {
		return err
	}
	pool, err := src.DB()
	if err != nil {
		return err
	}
	defer pool.Close()

	var marker commonModel.KeyValue
	if err := src.WithContext(ctx).Where("key = ?", commonModel.EchosPrivateCreatedDroppedKey).First(&marker).Error; err != nil || marker.Value == "" {
		return errors.New("upgrade the SQLite source with the audited Ech0 release before conversion")
	}
	dir, err := os.MkdirTemp(filepath.Dir(output), ".ech0-conversion-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "converted.db")
	if err := src.WithContext(ctx).Exec("VACUUM INTO ?", file).Error; err != nil {
		return err
	}
	if err := os.Chmod(file, 0600); err != nil {
		return err
	}
	if err := prepareConvertedFile(ctx, file); err != nil {
		return err
	}
	// Publish only a complete closed database, without replacing an existing backup.
	return os.Link(file, output)
}

var vecDimension = regexp.MustCompile(`(?i)using\s+vec0\s*\(\s*echo_id\s+text\s+primary\s+key\s*,\s*embedding\s+float\[(\d+)\]\s*\)\s*$`)

func prepareConvertedFile(ctx context.Context, file string) error {
	db, err := gorm.Open(sqlite.Open(file), buildGormConfig(logger.Silent))
	if err != nil {
		return err
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	db = db.WithContext(ctx)
	if err := db.Exec("PRAGMA journal_mode=DELETE").Error; err != nil {
		return err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable("ech0_schema") {
			return errors.New("source already has a remote schema marker")
		}
		if err := migrateModels(tx); err != nil {
			return err
		}
		if tx.Migrator().HasTable("vec_echo") {
			var ddl string
			if err := tx.Raw("SELECT sql FROM sqlite_schema WHERE name = 'vec_echo'").Row().Scan(&ddl); err != nil {
				return err
			}
			match := vecDimension.FindStringSubmatch(ddl)
			if len(match) != 2 {
				return errors.New("unsupported vec_echo schema")
			}
			dim, err := strconv.Atoi(match[1])
			if err != nil || dim <= 0 || dim > 1<<20 {
				return errors.New("invalid vec_echo dimension")
			}
			ddl = fmt.Sprintf("CREATE TABLE ech0_vec_copy(echo_id TEXT PRIMARY KEY, embedding F32_BLOB(%d) NOT NULL CHECK(length(embedding)=%d))", dim, dim*4)
			for _, sql := range []string{ddl, "INSERT INTO ech0_vec_copy SELECT echo_id, embedding FROM vec_echo", "DROP TABLE vec_echo", "ALTER TABLE ech0_vec_copy RENAME TO vec_echo"} {
				if err := tx.Exec(sql).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Exec("CREATE TABLE ech0_schema (id INTEGER PRIMARY KEY CHECK(id = 1), version INTEGER NOT NULL)").Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO ech0_schema VALUES(1,?)", remoteSchemaVersion).Error
	}); err != nil {
		return err
	}
	var check string
	if err := db.Raw("PRAGMA integrity_check").Row().Scan(&check); err != nil {
		return err
	}
	if check != "ok" {
		return errors.New("converted file failed integrity_check")
	}
	var violations int
	if err := db.Raw("SELECT count(*) FROM pragma_foreign_key_check").Row().Scan(&violations); err != nil {
		return err
	}
	if violations > 0 {
		return fmt.Errorf("converted file contains %d foreign key violations", violations)
	}
	return nil
}
