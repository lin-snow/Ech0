// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const remoteTablesSQL = "SELECT name FROM sqlite_schema WHERE type='table' AND name NOT GLOB 'sqlite_*' AND name NOT GLOB '_litestream_*' AND name NOT GLOB 'libsql_*' AND name NOT GLOB 'ech0_history_*' AND name <> 'ech0_schema' ORDER BY name"

// SnapshotHistory identifies a point in time. The database owns the old rows
// and triggers, so the connection that started the capture can disappear.
type SnapshotHistory struct {
	ID string
}

type historyTable struct {
	Name    string
	Ordinal int
}

func historyPrefix(n int) string {
	return "ech0_history_" + strconv.Itoa(n)
}

func StartSnapshotHistory(ctx context.Context, db *gorm.DB) (SnapshotHistory, error) {
	h := SnapshotHistory{ID: uuid.NewV7()}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := checkRemoteSchema(tx); err != nil {
			return err
		}
		if err := tx.Exec("CREATE TABLE IF NOT EXISTS ech0_history_state (slot INTEGER PRIMARY KEY CHECK(slot=1), id TEXT NOT NULL, expires INTEGER NOT NULL, cookie INTEGER NOT NULL)").Error; err != nil {
			return err
		}
		if err := tx.Exec("CREATE TABLE IF NOT EXISTS ech0_history_tables (ordinal INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE)").Error; err != nil {
			return err
		}
		if err := cleanupExpiredCapture(tx, time.Now()); err != nil {
			return err
		}
		// The single slot prevents overlapping captures from doubling write load.
		if err := tx.Exec("INSERT INTO ech0_history_state VALUES(1,?,?,0)", h.ID, time.Now().Add(24*time.Hour).Unix()).Error; err != nil {
			return err
		}
		var tables []string
		if err := tx.Raw(remoteTablesSQL).Scan(&tables).Error; err != nil {
			return err
		}
		for n, name := range tables {
			if err := installHistory(tx, name, n); err != nil {
				return fmt.Errorf("capture %s: %w", name, err)
			}
		}
		return tx.Exec("UPDATE ech0_history_state SET cookie=(SELECT schema_version FROM pragma_schema_version) WHERE slot=1").Error
	})
	return h, err
}

func installHistory(tx *gorm.DB, name string, n int) error {
	if err := tx.Exec("INSERT INTO ech0_history_tables VALUES(?,?)", n, name).Error; err != nil {
		return err
	}
	var columns []string
	if err := tx.Raw("SELECT name FROM pragma_table_info(?) ORDER BY cid", name).Scan(&columns).Error; err != nil {
		return err
	}
	var high *int64
	if err := tx.Raw("SELECT max(rowid) FROM " + quoteIdentifier(name)).Row().Scan(&high); err != nil {
		return err
	}
	prefix := historyPrefix(n)
	old := prefix + "_old"
	table := quoteIdentifier(name)
	fields := make([]string, len(columns))
	oldValues := make([]string, len(columns))
	for i, col := range columns {
		fields[i] = quoteIdentifier(col)
		oldValues[i] = "OLD." + fields[i]
	}
	fieldsSQL := strings.Join(fields, ",")
	// No type coercion: preserve arbitrary blobs and the storage class of values.
	if err := tx.Exec("CREATE TABLE " + old + " (_ech0_id INTEGER PRIMARY KEY, _ech0_existed INTEGER NOT NULL," + fieldsSQL + ")").Error; err != nil {
		return err
	}
	if high != nil {
		bound := strconv.FormatInt(*high, 10)
		// Only the first change matters. Explicit NOT EXISTS also works under
		// an outer INSERT ... ON CONFLICT UPDATE statement.
		saveOld := "INSERT INTO " + old + " SELECT OLD.rowid,1," + strings.Join(oldValues, ",") + " WHERE OLD.rowid <= " + bound + " AND NOT EXISTS(SELECT 1 FROM " + old + " WHERE _ech0_id=OLD.rowid);"
		saveNew := "INSERT INTO " + old + "(_ech0_id,_ech0_existed) SELECT NEW.rowid,0 WHERE NEW.rowid <= " + bound + " AND NOT EXISTS(SELECT 1 FROM " + old + " WHERE _ech0_id=NEW.rowid);"
		for _, rule := range []struct{ suffix, event, body string }{
			{"_delete", "BEFORE DELETE", saveOld}, {"_update", "AFTER UPDATE", saveOld + saveNew}, {"_insert", "AFTER INSERT", saveNew},
		} {
			if err := tx.Exec("CREATE TRIGGER " + prefix + rule.suffix + " " + rule.event + " ON " + table + " BEGIN " + rule.body + " END").Error; err != nil {
				return err
			}
		}
	}
	limit := "0"
	if high != nil {
		limit = "a.rowid <= " + strconv.FormatInt(*high, 10)
	}
	view := "CREATE VIEW " + prefix + "_rows AS SELECT a.rowid AS _ech0_id,a.* FROM " + table + " a WHERE " + limit + " AND NOT EXISTS (SELECT 1 FROM " + old + " b WHERE b._ech0_id=a.rowid) UNION ALL SELECT _ech0_id," + fieldsSQL + " FROM " + old + " WHERE _ech0_existed=1"
	return tx.Exec(view).Error
}

func (h SnapshotHistory) check(tx *gorm.DB) error {
	var valid int
	if err := tx.Raw("SELECT count(*) FROM ech0_history_state WHERE id=? AND expires > unixepoch() AND cookie=(SELECT schema_version FROM pragma_schema_version)", h.ID).Row().Scan(&valid); err != nil {
		return err
	}
	if valid != 1 {
		return errors.New("snapshot expired or schema changed")
	}
	return nil
}

func (h SnapshotHistory) finish(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int
		if err := tx.Raw("SELECT count(*) FROM ech0_history_state WHERE id=?", h.ID).Row().Scan(&count); err != nil || count == 0 {
			return err
		}
		return dropHistory(tx)
	})
}

// ExpireSnapshotHistory discards captures orphaned by an interrupted worker.
func ExpireSnapshotHistory(ctx context.Context, db *gorm.DB, now time.Time) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasTable("ech0_history_state") {
			return nil
		}
		return cleanupExpiredCapture(tx, now)
	})
}

func cleanupExpiredCapture(tx *gorm.DB, now time.Time) error {
	var count int
	if err := tx.Raw("SELECT count(*) FROM ech0_history_state WHERE expires <= ?", now.Unix()).Row().Scan(&count); err != nil || count == 0 {
		return err
	}
	return dropHistory(tx)
}

func dropHistory(tx *gorm.DB) error {
	var tables []historyTable
	if err := tx.Raw("SELECT name,ordinal FROM ech0_history_tables").Scan(&tables).Error; err != nil {
		return err
	}
	for _, table := range tables {
		prefix := historyPrefix(table.Ordinal)
		for _, stmt := range []string{"DROP TRIGGER IF EXISTS " + prefix + "_delete", "DROP TRIGGER IF EXISTS " + prefix + "_update", "DROP TRIGGER IF EXISTS " + prefix + "_insert", "DROP VIEW IF EXISTS " + prefix + "_rows", "DROP TABLE IF EXISTS " + prefix + "_old"} {
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
	}
	if err := tx.Exec("DELETE FROM ech0_history_tables").Error; err != nil {
		return err
	}
	return tx.Exec("DELETE FROM ech0_history_state").Error
}
