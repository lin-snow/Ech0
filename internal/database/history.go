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

const remoteTablesSQL = "SELECT name FROM sqlite_schema WHERE type='table' AND name NOT GLOB 'sqlite_*' AND name NOT GLOB '_litestream_*' AND name NOT GLOB 'libsql_*' AND name NOT GLOB 'ech0_history_*' AND name NOT GLOB 'ech0_export_*' AND name NOT IN ('ech0_schema','copilot_rounds') ORDER BY name"

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
	return h, h.Begin(ctx, db)
}

func (h SnapshotHistory) Begin(ctx context.Context, db *gorm.DB) error {
	if h.ID == "" {
		return errors.New("empty snapshot generation")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		var active string
		if err := tx.Raw("SELECT coalesce(max(id),'') FROM ech0_history_state").Row().Scan(&active); err != nil {
			return err
		}
		if active == h.ID {
			return h.check(tx)
		}
		// The single slot prevents overlapping captures from doubling write load.
		if err := tx.Exec("INSERT INTO ech0_history_state VALUES(1,?,?,0)", h.ID, time.Now().Add(24*time.Hour).Unix()).Error; err != nil {
			return err
		}
		var tables []string
		if err := tx.Raw(remoteTablesSQL).Scan(&tables).Error; err != nil {
			return err
		}
		// Gather metadata and maxima in two round trips. All capture DDL is
		// sent in one Hrana batch, within the same transaction as activation.
		var info []struct {
			TableName  string
			ColumnName string
		}
		if err := tx.Raw("SELECT t.name AS table_name,p.name AS column_name FROM (" + remoteTablesSQL + ") t JOIN pragma_table_info(t.name) p ORDER BY t.name,p.cid").Scan(&info).Error; err != nil {
			return err
		}
		columns := make(map[string][]string)
		for _, col := range info {
			columns[col.TableName] = append(columns[col.TableName], col.ColumnName)
		}
		queries := make([]string, len(tables))
		for n, name := range tables {
			queries[n] = "SELECT max(rowid) AS high FROM " + quoteIdentifier(name)
		}
		var highs []struct{ High *int64 }
		if err := tx.Raw(strings.Join(queries, " UNION ALL ")).Scan(&highs).Error; err != nil {
			return err
		}
		if len(highs) != len(tables) {
			return errors.New("snapshot source layout changed")
		}
		var ddl []string
		for n, name := range tables {
			if len(columns[name]) == 0 {
				return errors.New("snapshot table has no columns")
			}
			ddl = append(ddl, historyDDL(name, n, columns[name], highs[n].High)...)
		}
		if err := tx.Exec(strings.Join(ddl, ";")).Error; err != nil {
			return err
		}
		return tx.Exec("UPDATE ech0_history_state SET cookie=(SELECT schema_version FROM pragma_schema_version) WHERE slot=1").Error
	})
}

func historyDDL(name string, n int, columns []string, high *int64) []string {
	literal := "'" + strings.ReplaceAll(name, "'", "''") + "'"
	ddl := []string{fmt.Sprintf("INSERT INTO ech0_history_tables VALUES(%d,%s)", n, literal)}
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
	ddl = append(ddl, "CREATE TABLE "+old+" (_ech0_id INTEGER PRIMARY KEY, _ech0_existed INTEGER NOT NULL,"+fieldsSQL+")")
	if high != nil {
		bound := strconv.FormatInt(*high, 10)
		// Only the first change matters. Explicit NOT EXISTS also works under
		// an outer INSERT ... ON CONFLICT UPDATE statement.
		saveOld := "INSERT INTO " + old + " SELECT OLD.rowid,1," + strings.Join(oldValues, ",") + " WHERE OLD.rowid <= " + bound + " AND NOT EXISTS(SELECT 1 FROM " + old + " WHERE _ech0_id=OLD.rowid);"
		saveNew := "INSERT INTO " + old + "(_ech0_id,_ech0_existed) SELECT NEW.rowid,0 WHERE NEW.rowid <= " + bound + " AND NOT EXISTS(SELECT 1 FROM " + old + " WHERE _ech0_id=NEW.rowid);"
		for _, rule := range []struct{ suffix, event, body string }{
			{"_delete", "BEFORE DELETE", saveOld}, {"_update", "AFTER UPDATE", saveOld + saveNew}, {"_insert", "AFTER INSERT", saveNew},
		} {
			ddl = append(ddl, "CREATE TRIGGER "+prefix+rule.suffix+" "+rule.event+" ON "+table+" BEGIN "+rule.body+" END")
		}
	}
	limit := "0"
	if high != nil {
		limit = "a.rowid <= " + strconv.FormatInt(*high, 10)
	}
	view := "CREATE VIEW " + prefix + "_rows AS SELECT a.rowid AS _ech0_id,a.* FROM " + table + " a WHERE " + limit + " AND NOT EXISTS (SELECT 1 FROM " + old + " b WHERE b._ech0_id=a.rowid) UNION ALL SELECT _ech0_id," + fieldsSQL + " FROM " + old + " WHERE _ech0_existed=1"
	return append(ddl, view)
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

func (h SnapshotHistory) Finish(ctx context.Context, db *gorm.DB) error {
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
	var statements []string
	for _, table := range tables {
		prefix := historyPrefix(table.Ordinal)
		statements = append(statements, "DROP TRIGGER IF EXISTS "+prefix+"_delete", "DROP TRIGGER IF EXISTS "+prefix+"_update", "DROP TRIGGER IF EXISTS "+prefix+"_insert", "DROP VIEW IF EXISTS "+prefix+"_rows", "DROP TABLE IF EXISTS "+prefix+"_old")
	}
	if len(statements) > 0 {
		if err := tx.Exec(strings.Join(statements, ";")).Error; err != nil {
			return err
		}
	}
	if err := tx.Exec("DELETE FROM ech0_history_tables").Error; err != nil {
		return err
	}
	return tx.Exec("DELETE FROM ech0_history_state").Error
}
