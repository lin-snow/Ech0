// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	model "github.com/lin-snow/ech0/internal/model/embedding"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/driver/sqlite"
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
	return r.db().WithContext(ctx)
}

func usesLibSQL(db *gorm.DB) bool {
	dialector, ok := db.Dialector.(*sqlite.Dialector)
	return ok && dialector.DriverName == "libsql"
}

func (r *EmbeddingRepository) EnsureVecTable(ctx context.Context, dim int) error {
	if dim <= 0 {
		return errors.New("embedding: invalid vector dim")
	}
	ddl := fmt.Sprintf(
		"CREATE VIRTUAL TABLE IF NOT EXISTS %s USING vec0(echo_id TEXT PRIMARY KEY, embedding FLOAT[%d])",
		vecTable, dim,
	)
	db := r.getDB(ctx).WithContext(ctx)
	if usesLibSQL(db) {
		ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(echo_id TEXT PRIMARY KEY, embedding F32_BLOB(%d) NOT NULL CHECK(length(embedding) = %d))", vecTable, dim, dim*4)
	}
	return db.Exec(ddl).Error
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

// Upsert stores an Echo's embedding, unless the Echo is already gone.
//
// Indexing is slow — it waits on the embedding API — and runs on its own
// subscription, so an Echo deleted right after it was written can have its
// delete handled first and its vector arrive afterwards. Checking for the Echo
// inside the same write transaction closes that window: a delete that commits
// before this transaction reads is seen here, and one that commits after has
// to wait for this write and is then followed by its own RemoveEcho.
func (r *EmbeddingRepository) Upsert(ctx context.Context, meta *model.EchoEmbedding, vector []float32) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := transaction.CheckWrite(ctx, tx); err != nil {
			return err
		}
		var live int64
		if err := tx.Model(&echoModel.Echo{}).Where("id = ?", meta.EchoID).Count(&live).Error; err != nil {
			return err
		}
		if live == 0 {
			return nil
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "echo_id"}},
			UpdateAll: true,
		}).Create(meta).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM "+vecTable+" WHERE echo_id = ?", meta.EchoID).Error; err != nil {
			return err
		}
		insert := "INSERT INTO " + vecTable + "(echo_id, embedding) VALUES (?, ?)"
		if usesLibSQL(tx) {
			insert = "INSERT INTO " + vecTable + "(echo_id, embedding) VALUES (?, vector32(?))"
		}
		return tx.Exec(
			insert,
			meta.EchoID, vecToJSON(vector),
		).Error
	})
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

// Search returns the k nearest Echos to vector, optionally only those authored
// by authorID.
//
// The KNN runs on the vector table, but every hit is then resolved against the
// live echos table rather than the copy stored beside its vector. That copy is
// a snapshot from index time: it keeps a username the author has since changed
// and outlives an Echo deleted while its embedding was still being computed.
// Joining on the source of truth filters by the stable user ID, drops orphans,
// and returns the text as it reads now.
func (r *EmbeddingRepository) Search(ctx context.Context, vector []float32, k int, authorID string) ([]model.SearchResult, error) {
	if k <= 0 {
		k = 6
	}

	fetch := k
	if authorID != "" {
		fetch = min(k*searchOverfetchFactor, searchOverfetchCap)
	}

	type knnRow struct {
		EchoID   string
		Distance float64
	}
	var rows []knnRow
	db := r.getDB(ctx)
	query := "SELECT echo_id, distance FROM " + vecTable + " WHERE embedding MATCH ? ORDER BY distance LIMIT ?"
	if usesLibSQL(db) {
		query = "SELECT echo_id, vector_distance_l2(embedding, vector32(?)) AS distance FROM " + vecTable + " ORDER BY distance, echo_id LIMIT ?"
	}
	if err := db.Raw(query,
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

	echoQuery := r.getDB(ctx).Model(&echoModel.Echo{}).
		Select("id", "content", "username", "created_at").
		Where("id IN ?", ids)
	if authorID != "" {
		echoQuery = echoQuery.Where("user_id = ?", authorID)
	}
	var echos []echoModel.Echo
	if err := echoQuery.Find(&echos).Error; err != nil {
		return nil, err
	}
	byID := make(map[string]echoModel.Echo, len(echos))
	for _, e := range echos {
		byID[e.ID] = e
	}

	results := make([]model.SearchResult, 0, min(k, len(rows)))
	for _, row := range rows {
		e, ok := byID[row.EchoID]
		if !ok {
			continue
		}
		results = append(results, model.SearchResult{
			EchoID:      e.ID,
			Content:     e.Content,
			Username:    e.Username,
			EchoCreated: e.CreatedAt,
			Distance:    row.Distance,
		})
		if len(results) >= k {
			break
		}
	}
	return results, nil
}

// PruneOrphans deletes index rows whose Echo no longer exists, and reports how
// many went. An Echo deleted while its embedding was in flight leaves exactly
// such a row behind; a rebuild is the natural moment to sweep them.
func (r *EmbeddingRepository) PruneOrphans(ctx context.Context) (int64, error) {
	db := r.getDB(ctx)
	live := db.Model(&echoModel.Echo{}).Select("id")
	res := db.Where("echo_id NOT IN (?)", live).Delete(&model.EchoEmbedding{})
	if res.Error != nil {
		return 0, res.Error
	}
	if err := db.Exec("DELETE FROM " + vecTable + " WHERE echo_id NOT IN (SELECT echo_id FROM echo_embeddings)").Error; err != nil {
		return res.RowsAffected, err
	}
	return res.RowsAffected, nil
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
