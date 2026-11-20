// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package journal

import (
	"context"

	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	journalModel "github.com/lin-snow/ech0/internal/model/journal"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Writer flushes the upstream log ring before a Function returns. It never
// opens a separate write transaction while application code is holding one.
type Writer struct {
	mu      sync.Mutex
	source  string
	through uint64
	db      func() *gorm.DB
}

func NewWriter(db func() *gorm.DB) *Writer { return &Writer{db: db, source: uuidUtil.NewV7()} }

func (w *Writer) Flush(parent context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	through, entries := logUtil.LogsAfter(w.through)
	if len(entries) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 3*time.Second)
	defer cancel()
	first := through - uint64(len(entries)) + 1
	for len(entries) > 0 {
		n := min(64, len(entries))
		batch := make([]journalModel.Entry, 0, n)
		for i, e := range entries[:n] {
			raw, err := encodeEntry(e)
			if err != nil {
				return
			}
			batch = append(batch, journalModel.Entry{Source: w.source, Seq: first + uint64(i), Created: time.Now().Unix(), Level: e.Level, Search: strings.ToLower(raw + " " + logUtil.ParseLogLine(raw).Msg), Raw: raw})
		}
		if err := w.db().WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&batch).Error; err != nil {
			// Avoid recursion into the journal when its own database is down.
			fmt.Fprintln(os.Stderr, "Ech0 journal flush unavailable")
			return
		}
		first += uint64(n)
		w.through = first - 1
		entries = entries[n:]
	}
}

func Tail(ctx context.Context, db *gorm.DB, limit int, level, keyword string) ([]logUtil.LogEntry, error) {
	if limit <= 0 {
		limit = 200
	}
	var rows []journalModel.Entry
	query := db.WithContext(ctx).Select("raw").Order("id DESC").Limit(min(limit, 5000))
	if l := strings.ToLower(strings.TrimSpace(level)); l != "" && l != "all" {
		query = query.Where("level=?", l)
	}
	if k := strings.ToLower(strings.TrimSpace(keyword)); k != "" {
		query = query.Where("instr(search,?)>0", k)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	slices.Reverse(rows)
	entries := make([]logUtil.LogEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, logUtil.ParseLogLine(row.Raw))
	}
	return entries, nil
}

func Latest(ctx context.Context, db *gorm.DB) (int64, error) {
	var id int64
	err := db.WithContext(ctx).Raw("SELECT coalesce(max(id),0) FROM ech0_journal").Row().Scan(&id)
	return id, err
}

// Follow blocks only for the authenticated stream's lifetime. A monotonic
// cursor crosses pooled connections and expired log retention windows.
func Follow(ctx context.Context, db *gorm.DB, after int64, send func(int64, logUtil.LogEntry) error, heartbeat func() error) error {
	poll := time.NewTicker(time.Second)
	defer poll.Stop()
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()
	for {
		var rows []journalModel.Entry
		if err := db.WithContext(ctx).Select("id,raw").Where("id>?", after).Order("id").Limit(128).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if err := send(row.ID, logUtil.ParseLogLine(row.Raw)); err != nil {
				return err
			}
			after = row.ID
		}
		if len(rows) == 128 {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-poll.C:
		case <-keepAlive.C:
			if err := heartbeat(); err != nil {
				return err
			}
		}
	}
}
