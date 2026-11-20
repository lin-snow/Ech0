// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	copilotModel "github.com/lin-snow/ech0/internal/model/copilot"
	"gorm.io/gorm"
)

type Rounds struct{ db func() *gorm.DB }

func NewRounds(db func() *gorm.DB) *Rounds { return &Rounds{db: db} }

func (r *Rounds) Register(ctx context.Context, id, owner string, deadline time.Time) error {
	return r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires <= ?", time.Now().UnixMilli()).Delete(&copilotModel.Round{}).Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO copilot_rounds(id,owner,expires) VALUES(?,?,?)", id, owner, deadline.UnixMilli()).Error
	})
}

func (r *Rounds) Reply(ctx context.Context, id, owner string, answer []byte) (bool, error) {
	result := r.db().WithContext(ctx).Exec("UPDATE copilot_rounds SET answer=? WHERE id=? AND owner=? AND expires>? AND answer IS NULL", answer, id, owner, time.Now().UnixMilli())
	return result.RowsAffected == 1, result.Error
}

func (r *Rounds) Read(ctx context.Context, id string) ([]byte, error) {
	var payload []byte
	err := r.db().WithContext(ctx).Raw("SELECT answer FROM copilot_rounds WHERE id=? AND expires>?", id, time.Now().UnixMilli()).Row().Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, context.DeadlineExceeded
	}
	if len(payload) == 0 {
		return nil, err
	}
	return payload, err
}

func (r *Rounds) Close(ctx context.Context, id string) error {
	return r.db().WithContext(ctx).Where("id=?", id).Delete(&copilotModel.Round{}).Error
}
