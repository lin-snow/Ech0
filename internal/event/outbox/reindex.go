// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/lin-snow/ech0/internal/job"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	outboxModel "github.com/lin-snow/ech0/internal/model/outbox"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const reindexSubject = "@reindex"

type reindexCursor struct {
	High     int64
	After    *int64
	Prepared bool
	Model    string
	Dim      int
	Failures int
	Result   embeddingService.BackfillResult
}

func StartReindex(ctx context.Context, db *gorm.DB) (jobModel.Job, error) {
	var jobRow jobModel.Job
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		claim := tx.Exec(`INSERT INTO jobs(type,status,phase,error,payload,started_at,finished_at,updated_at) VALUES('reindex','pending','','','{}',unixepoch(),NULL,unixepoch())
            ON CONFLICT(type) DO UPDATE SET status=excluded.status,phase='',error='',payload='{}',started_at=excluded.started_at,finished_at=NULL,updated_at=excluded.updated_at WHERE jobs.status NOT IN ('pending','running')`)
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected != 1 {
			return job.ErrAlreadyRunning
		}
		var cursor reindexCursor
		if err := tx.Raw("SELECT coalesce(max(rowid),0),count(*) FROM echos").Row().Scan(&cursor.High, &cursor.Result.Total); err != nil {
			return err
		}
		data, err := json.Marshal(cursor)
		if err != nil {
			return err
		}
		if err := tx.Create(&outboxModel.Delivery{ID: uuidUtil.NewV7(), Lane: "embedding", Subject: reindexSubject, Payload: data}).Error; err != nil {
			return err
		}
		return tx.First(&jobRow, "type='reindex'").Error
	})
	return jobRow, err
}

func CancelReindex(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM ech0_deliveries WHERE lane='embedding' AND subject=?", reindexSubject).Error; err != nil {
			return err
		}
		return tx.Exec("UPDATE jobs SET status='cancelled',finished_at=unixepoch(),updated_at=unixepoch() WHERE type='reindex' AND status IN ('pending','running')").Error
	})
}

// The same lane serializes automatic indexing and administrative rebuilds.
// Each step uses at most one provider call plus its upstream fallback. A
// retry may call the provider again; vector upsert and the cursor are fenced.
func (w *Worker) reindex(ctx context.Context, db *gorm.DB, d outboxModel.Delivery) error {
	var c reindexCursor
	if err := json.Unmarshal(d.Payload, &c); err != nil {
		return err
	}
	check := func(tx *gorm.DB) error {
		var yes int
		return tx.Raw("SELECT 1 FROM ech0_deliveries WHERE id=? AND lease=? AND ready>unixepoch()", d.ID, d.Lease).Row().Scan(&yes)
	}
	guarded := transaction.WithWriteGuard(ctx, check)
	var rowID int64
	var found bool
	var batch embeddingService.BackfillResult
	var setting settingModel.EmbeddingSetting
	var done bool
	var stepErr error
	if !c.Prepared {
		stepErr = db.WithContext(guarded).Transaction(func(tx *gorm.DB) error {
			if err := check(tx); err != nil {
				return err
			}
			txCtx := context.WithValue(guarded, transaction.TxKey, tx)
			prepared, err := w.index.PrepareBackfill(txCtx)
			if err == nil {
				c.Model, c.Dim, c.Prepared = prepared.Model, prepared.Dim, true
			}
			return err
		})
	} else {
		query := db.WithContext(guarded).Table("echos").Select("rowid,id").Where("rowid<=?", c.High)
		if c.After != nil {
			query = query.Where("rowid>?", *c.After)
		}
		var id string
		stepErr = query.Order("rowid").Limit(1).Row().Scan(&rowID, &id)
		switch {
		case errors.Is(stepErr, sql.ErrNoRows):
			done, stepErr = true, nil
		case stepErr == nil:
			found = true
			setting, stepErr = coreSetting.Embedding.Get(guarded, w.settings)
			if stepErr == nil && (!setting.Enable || setting.Dim != c.Dim || setting.Model != c.Model) {
				return w.finishReindex(db, d, c, "embedding configuration changed during reindex")
			}
			if stepErr == nil {
				var echo echoModel.Echo
				stepErr = db.WithContext(guarded).Preload("Tags").First(&echo, "id=?", id).Error
				if errors.Is(stepErr, gorm.ErrRecordNotFound) {
					stepErr = nil
					batch.Skipped = 1
				} else if stepErr == nil {
					batch, stepErr = w.index.IndexBatch(guarded, setting, []echoModel.Echo{echo})
				}
			}
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if stepErr != nil || batch.Failed > 0 {
		c.Failures++
		if c.Failures < 4 {
			return w.saveReindex(db, d, c, time.Now().Add(time.Duration(1<<c.Failures)*time.Second).Unix())
		}
		if !c.Prepared || !found {
			return w.finishReindex(db, d, c, "could not prepare or read the embedding index")
		}
		batch = embeddingService.BackfillResult{Failed: 1}
	}
	if done {
		reason := ""
		if c.Result.Indexed == 0 && c.Result.Failed > 0 {
			reason = "could not embed any echo"
		}
		return w.finishReindex(db, d, c, reason)
	}
	c.Failures = 0
	if found {
		c.After = &rowID
	}
	c.Result.Indexed += batch.Indexed
	c.Result.Skipped += batch.Skipped
	c.Result.Failed += batch.Failed
	return w.saveReindex(db, d, c, 0)
}

func (w *Worker) saveReindex(db *gorm.DB, d outboxModel.Delivery, c reindexCursor, ready int64) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	progress, err := json.Marshal(c.Result)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec("UPDATE ech0_deliveries SET payload=?,attempt=attempt+1,ready=?,lease='' WHERE id=? AND lease=?", data, ready, d.ID, d.Lease)
		if result.Error != nil || result.RowsAffected != 1 {
			return result.Error
		}
		return tx.Exec("UPDATE jobs SET status='running',phase='indexing',payload=?,updated_at=unixepoch() WHERE type='reindex'", string(progress)).Error
	})
}

func (w *Worker) finishReindex(db *gorm.DB, d outboxModel.Delivery, c reindexCursor, reason string) error {
	progress, err := json.Marshal(c.Result)
	if err != nil {
		return err
	}
	status := jobModel.StatusSuccess
	if reason != "" {
		status = jobModel.StatusFailed
	}
	return db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec("DELETE FROM ech0_deliveries WHERE id=? AND lease=?", d.ID, d.Lease)
		if result.Error != nil || result.RowsAffected != 1 {
			return result.Error
		}
		return tx.Exec("UPDATE jobs SET status=?,error=?,phase='indexing',payload=?,updated_at=unixepoch(),finished_at=unixepoch() WHERE type='reindex'", status, reason, string(progress)).Error
	})
}
