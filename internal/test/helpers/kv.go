// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

// NewTestKV returns a KV store over db's key_values table, so code that writes
// settings through a Spec lands in the row the test asserts on. Writes join a
// transaction carried in ctx, matching the repository layer.
func NewTestKV(t *testing.T, db *gorm.DB) kvstore.Store {
	t.Helper()
	return kvstore.NewPersistent(gormBackend{db: db})
}

type gormBackend struct {
	db *gorm.DB
}

func (b gormBackend) conn(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return b.db.WithContext(ctx)
}

func (b gormBackend) GetKeyValue(ctx context.Context, key string) (string, error) {
	var row commonModel.KeyValue
	if err := b.conn(ctx).Where("key = ?", key).First(&row).Error; err != nil {
		return "", err
	}
	return row.Value, nil
}

func (b gormBackend) AddOrUpdateKeyValue(ctx context.Context, key, value string) error {
	return b.conn(ctx).Exec(
		`INSERT INTO key_values(key, value) VALUES(?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	).Error
}

func (b gormBackend) DeleteKeyValue(ctx context.Context, key string) error {
	return b.conn(ctx).Exec(`DELETE FROM key_values WHERE key = ?`, key).Error
}
