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
	defer func() { _ = pool.Close() }()
	return writeRemoteSnapshot(ctx, db, output)
}

func writeRemoteSnapshot(ctx context.Context, source *gorm.DB, output string) error {
	return publishSnapshot(output, func(staging string) error { return materializeSnapshot(ctx, source, staging, nil) })
}

func publishSnapshot(output string, write func(string) error) error {
	if err := enableLocalSQLite(); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(filepath.Dir(output), ".ech0-backup-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	staging := filepath.Join(dir, "ech0.db")
	if err := write(staging); err != nil {
		return err
	}
	if err := os.Chmod(staging, 0600); err != nil {
		return err
	}
	return os.Link(staging, output)
}

var remoteVecType = regexp.MustCompile(`(?i)^F32_BLOB\(([0-9]+)\)$`)

func materializeSnapshot(ctx context.Context, source *gorm.DB, file string, history *SnapshotHistory) error {
	target, err := gorm.Open(sqlite.Open(file), buildGormConfig(logger.Silent))
	if err != nil {
		return err
	}
	pool, err := target.DB()
	if err != nil {
		return err
	}
	defer func() { _ = pool.Close() }()
	if err := migrateModels(target.WithContext(ctx)); err != nil {
		return err
	}
	copyRows := func(src *gorm.DB) error {
		if err := checkRemoteSchema(src); err != nil {
			return err
		}
		var tables []string
		if err := src.Raw(remoteTablesSQL).Scan(&tables).Error; err != nil {
			return err
		}
		dest, err := pool.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = dest.Rollback() }()
		for n, table := range tables {
			if table == "vec_echo" {
				if err := createNativeVector(ctx, src, dest); err != nil {
					return err
				}
			}
			readTable, cursor, selectCols := quoteIdentifier(table), "rowid", "rowid,*"
			if history != nil {
				readTable, cursor, selectCols = historyPrefix(n)+"_rows", "_ech0_id", "*"
			}
			if err := copyTable(ctx, src, dest, table, readTable, cursor, selectCols); err != nil {
				return fmt.Errorf("copy table %s: %w", table, err)
			}
		}
		// A fresh libSQL deployment starts with the current model layout. Record
		// that provenance so this native snapshot can be imported again offline.
		if _, err := dest.ExecContext(ctx, "INSERT INTO key_values(key,value) VALUES(?,?) ON CONFLICT(key) DO NOTHING", commonModel.EchosPrivateCreatedDroppedKey, "current schema exported from libSQL"); err != nil {
			return err
		}
		if history != nil {
			if err := history.check(src); err != nil {
				return err
			}
		}
		return dest.Commit()
	}
	source = source.WithContext(ctx)
	if history != nil {
		// Every SELECT finishes independently of the next page. The final
		// generation and schema check governs publication of the finished file.
		return copyRows(source)
	}
	return source.Transaction(copyRows)
}

func quoteIdentifier(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// Each page has its own local prepared statement. The caller decides whether
// to continue in this connection or checkpoint the native image for another worker.
const snapshotPageRows = 64

func copyTable(ctx context.Context, src *gorm.DB, dest *sql.Tx, name, readTable, cursor, selectCols string) error {
	var after *int64
	for {
		next, done, err := copyPage(ctx, src, dest, name, readTable, cursor, selectCols, after)
		if err != nil || done {
			return err
		}
		after = &next
	}
}

func copyPage(ctx context.Context, src *gorm.DB, dest *sql.Tx, name, readTable, cursor, selectCols string, after *int64) (last int64, done bool, err error) {
	query := "SELECT " + selectCols + " FROM " + readTable
	var args []any
	if after != nil {
		query += " WHERE " + cursor + " > ?"
		args = append(args, *after)
	}
	query += " ORDER BY " + cursor + " LIMIT ?"
	rows, err := src.WithContext(ctx).Raw(query, append(args, snapshotPageRows)...).Rows()
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return 0, false, err
	}
	fields := make([]string, len(cols)-1)
	bind := make([]string, len(fields))
	for i, col := range cols[1:] {
		fields[i] = quoteIdentifier(col)
		bind[i] = "?"
	}
	insert, err := dest.PrepareContext(ctx, "INSERT INTO "+quoteIdentifier(name)+"("+strings.Join(fields, ",")+") VALUES("+strings.Join(bind, ",")+")")
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = insert.Close() }()
	values := make([]any, len(fields))
	pointers := make([]any, len(cols))
	pointers[0] = &last
	for i := range values {
		pointers[i+1] = &values[i]
	}
	n := 0
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return 0, false, err
		}
		if _, err := insert.ExecContext(ctx, values...); err != nil {
			return 0, false, err
		}
		n++
	}
	return last, n < snapshotPageRows, rows.Err()
}

func createNativeVector(ctx context.Context, src *gorm.DB, dest *sql.Tx) error {
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
	return nil
}
