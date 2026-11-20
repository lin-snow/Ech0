// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func BackupLibSQL(ctx context.Context, cfg config.DatabaseConfig, output string) error {
	db, err := OpenRemoteDatabase(cfg)
	if err != nil {
		return err
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	return writeRemoteSnapshot(ctx, db, output)
}

func writeRemoteSnapshot(ctx context.Context, source *gorm.DB, output string) error {
	if err := enableLocalSQLite(); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(output), ".ech0-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	staging := filepath.Join(dir, "ech0.db")
	if err := materializeSnapshot(ctx, source, staging); err != nil {
		return err
	}
	if err := os.Chmod(staging, 0600); err != nil {
		return err
	}
	return os.Link(staging, output)
}

var remoteVecType = regexp.MustCompile(`(?i)^F32_BLOB\(([0-9]+)\)$`)

func materializeSnapshot(ctx context.Context, source *gorm.DB, file string) error {
	target, err := gorm.Open(sqlite.Open(file), buildGormConfig(logger.Silent))
	if err != nil {
		return err
	}
	pool, err := target.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := migrateModels(target.WithContext(ctx)); err != nil {
		return err
	}
	return source.WithContext(ctx).Transaction(func(src *gorm.DB) error {
		if err := checkRemoteSchema(src); err != nil {
			return err
		}
		var tables []string
		if err := src.Raw("SELECT name FROM sqlite_schema WHERE type='table' AND name NOT GLOB 'sqlite_*' AND name NOT GLOB '_litestream_*' AND name NOT GLOB 'libsql_*' AND name <> 'ech0_schema' ORDER BY name").Scan(&tables).Error; err != nil {
			return err
		}
		dest, err := pool.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer dest.Rollback()
		for _, table := range tables {
			if table == "vec_echo" {
				var typ string
				if err := src.Raw("SELECT type FROM pragma_table_info('vec_echo') WHERE name='embedding'").Row().Scan(&typ); err != nil {
					return err
				}
				m := remoteVecType.FindStringSubmatch(typ)
				if len(m) != 2 {
					return errors.New("unsupported remote vector type")
				}
				dim, err := strconv.Atoi(m[1])
				if err != nil || dim <= 0 || dim > 1<<20 {
					return errors.New("invalid remote vector dimension")
				}
				if _, err = dest.ExecContext(ctx, fmt.Sprintf("CREATE VIRTUAL TABLE vec_echo USING vec0(echo_id TEXT PRIMARY KEY, embedding FLOAT[%d])", dim)); err != nil {
					return err
				}
			}
			if err := copyTable(ctx, src, dest, table); err != nil {
				return fmt.Errorf("copy table %s: %w", table, err)
			}
		}
		// A fresh libSQL deployment starts with the current model layout. Record
		// that provenance so this native snapshot can be imported again offline.
		if _, err := dest.ExecContext(ctx, "INSERT INTO key_values(key,value) VALUES(?,?) ON CONFLICT(key) DO NOTHING", commonModel.EchosPrivateCreatedDroppedKey, "current schema exported from libSQL"); err != nil {
			return err
		}
		return dest.Commit()
	})
}

func quoteIdentifier(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// Seek by rowid inside a single remote transaction. Each response is bounded
// to avoid downloading the entire database into the function's heap.
func copyTable(ctx context.Context, src *gorm.DB, dest *sql.Tx, name string) error {
	table := quoteIdentifier(name)
	query := "SELECT rowid,* FROM " + table
	const pageSize = 64
	var last int64
	var insert *sql.Stmt
	for first := true; ; first = false {
		var rows *sql.Rows
		var err error
		if first {
			rows, err = src.Raw(query+" ORDER BY rowid LIMIT ?", pageSize).Rows()
		} else {
			rows, err = src.Raw(query+" WHERE rowid > ? ORDER BY rowid LIMIT ?", last, pageSize).Rows()
		}
		if err != nil {
			return err
		}
		cols, err := rows.Columns()
		if err != nil {
			rows.Close()
			return err
		}
		if insert == nil {
			// Preparing on an empty table also detects unexpected schema drift.
			fields := make([]string, len(cols)-1)
			bind := make([]string, len(fields))
			for i, col := range cols[1:] {
				fields[i] = quoteIdentifier(col)
				bind[i] = "?"
			}
			insert, err = dest.PrepareContext(ctx, "INSERT INTO "+table+"("+strings.Join(fields, ",")+") VALUES("+strings.Join(bind, ",")+")")
			if err != nil {
				rows.Close()
				return err
			}
			defer insert.Close()
		}
		n := 0
		values := make([]any, len(cols)-1)
		pointers := make([]any, len(cols))
		pointers[0] = &last
		for i := range values {
			pointers[i+1] = &values[i]
		}
		for rows.Next() {
			if err = rows.Scan(pointers...); err != nil {
				break
			}
			if _, err = insert.ExecContext(ctx, values...); err != nil {
				break
			}
			n++
		}
		if err == nil {
			err = rows.Err()
		}
		rows.Close()
		if err != nil {
			return err
		}
		if n < pageSize {
			return nil
		}
	}
}
