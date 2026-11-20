// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lin-snow/ech0/internal/database/pagefile"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ImageCursor is checkpointed with the modified native SQLite blocks. An
// interrupted worker can repeat a step from the previous committed cursor.
type ImageCursor struct {
	Table int    `json:"table"`
	After *int64 `json:"after,omitempty"`
	Ready bool   `json:"ready"`
}

func (h SnapshotHistory) CopyImagePage(ctx context.Context, src *gorm.DB, file *pagefile.File, at ImageCursor) (ImageCursor, error) {
	if err := enableLocalSQLite(); err != nil {
		return at, err
	}
	src = src.WithContext(ctx)
	if err := h.check(src); err != nil {
		return at, err
	}
	next := at
	err := pagefile.UseSQLite(ctx, file, func(pool *sql.DB) error {
		if file.Size == 0 {
			db, err := gorm.Open(sqlite.New(sqlite.Config{Conn: pool}), buildGormConfig(logger.Silent))
			if err != nil {
				return err
			}
			if err := migrateModels(db.WithContext(ctx)); err != nil {
				return err
			}
		}
		var name string
		err := src.Raw("SELECT name FROM ech0_history_tables WHERE ordinal=?", at.Table).Row().Scan(&name)
		if errors.Is(err, sql.ErrNoRows) {
			_, err = pool.ExecContext(ctx, "INSERT INTO key_values(key,value) VALUES(?,?) ON CONFLICT(key) DO NOTHING", commonModel.EchosPrivateCreatedDroppedKey, "current schema exported from libSQL")
			next.Ready = err == nil
			return err
		}
		if err != nil {
			return err
		}
		tx, err := pool.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if name == "vec_echo" && at.After == nil {
			if err := createNativeVector(ctx, src, tx); err != nil {
				return err
			}
		}
		last, done, err := copyPage(ctx, src, tx, name, historyPrefix(at.Table)+"_rows", "_ech0_id", "*", at.After)
		if err != nil {
			return err
		}
		next.After = &last
		if done {
			next.Table++
			next.After = nil
		}
		return tx.Commit()
	})
	if err != nil {
		return at, err
	}
	if err := h.check(src); err != nil {
		return at, err
	}
	return next, nil
}
