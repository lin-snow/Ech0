// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	model "github.com/lin-snow/ech0/internal/model/embedding"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const vecTable = "vec_echo"

type EmbeddingRepository struct {
	db func() *gorm.DB
}

func NewEmbeddingRepository(dbProvider func() *gorm.DB) *EmbeddingRepository {
	return &EmbeddingRepository{db: dbProvider}
}

func (r *EmbeddingRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *EmbeddingRepository) EnsureVecTable(ctx context.Context, dim int) error {
	if dim <= 0 {
		return errors.New("embedding: invalid vector dim")
	}
	ddl := fmt.Sprintf(
		"CREATE VIRTUAL TABLE IF NOT EXISTS %s USING vec0(echo_id TEXT PRIMARY KEY, embedding FLOAT[%d])",
		vecTable, dim,
	)
	return r.getDB(ctx).Exec(ddl).Error
}

func (r *EmbeddingRepository) DropVecTable(ctx context.Context) error {
	return r.getDB(ctx).Exec("DROP TABLE IF EXISTS " + vecTable).Error
}

func vecToJSON(vec []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, v := range vec {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

func (r *EmbeddingRepository) Upsert(ctx context.Context, meta *model.EchoEmbedding, vector []float32) error {
	db := r.getDB(ctx)

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "echo_id"}},
		UpdateAll: true,
	}).Create(meta).Error; err != nil {
		return err
	}

	if err := db.Exec("DELETE FROM "+vecTable+" WHERE echo_id = ?", meta.EchoID).Error; err != nil {
		return err
	}
	if err := db.Exec(
		"INSERT INTO "+vecTable+"(echo_id, embedding) VALUES (?, ?)",
		meta.EchoID, vecToJSON(vector),
	).Error; err != nil {
		return err
	}
	return nil
}

func (r *EmbeddingRepository) Delete(ctx context.Context, echoID string) error {
	db := r.getDB(ctx)
	if err := db.Where("echo_id = ?", echoID).Delete(&model.EchoEmbedding{}).Error; err != nil {
		return err
	}
	_ = db.Exec("DELETE FROM "+vecTable+" WHERE echo_id = ?", echoID).Error
	return nil
}

func (r *EmbeddingRepository) GetMeta(ctx context.Context, echoID string) (*model.EchoEmbedding, bool, error) {
	var m model.EchoEmbedding
	err := r.getDB(ctx).Where("echo_id = ?", echoID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &m, true, nil
}

const searchOverfetchFactor = 8

const searchOverfetchCap = 200

func (r *EmbeddingRepository) Search(ctx context.Context, vector []float32, k int, authorUsername string) ([]model.SearchResult, error) {
	if k <= 0 {
		k = 6
	}

	fetch := k
	if authorUsername != "" {
		fetch = min(k*searchOverfetchFactor, searchOverfetchCap)
	}

	type knnRow struct {
		EchoID   string
		Distance float64
	}
	var rows []knnRow
	if err := r.getDB(ctx).Raw(
		"SELECT echo_id, distance FROM "+vecTable+" WHERE embedding MATCH ? ORDER BY distance LIMIT ?",
		vecToJSON(vector), fetch,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.EchoID
	}

	metaQuery := r.getDB(ctx).Where("echo_id IN ?", ids)
	if authorUsername != "" {
		metaQuery = metaQuery.Where("username = ?", authorUsername)
	}
	var metas []model.EchoEmbedding
	if err := metaQuery.Find(&metas).Error; err != nil {
		return nil, err
	}
	metaByID := make(map[string]model.EchoEmbedding, len(metas))
	for _, m := range metas {
		metaByID[m.EchoID] = m
	}

	results := make([]model.SearchResult, 0, min(k, len(rows)))
	for _, row := range rows {
		m, ok := metaByID[row.EchoID]
		if !ok {
			continue
		}
		results = append(results, model.SearchResult{
			EchoID:      m.EchoID,
			Content:     m.Content,
			Username:    m.Username,
			EchoCreated: m.EchoCreated,
			Distance:    row.Distance,
		})
		if len(results) >= k {
			break
		}
	}
	return results, nil
}

func (r *EmbeddingRepository) ClearAll(ctx context.Context) error {
	db := r.getDB(ctx)
	if err := db.Where("1 = 1").Delete(&model.EchoEmbedding{}).Error; err != nil {
		return err
	}
	_ = db.Exec("DELETE FROM " + vecTable).Error
	return nil
}

func (r *EmbeddingRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.getDB(ctx).Model(&model.EchoEmbedding{}).Count(&n).Error
	return n, err
}
