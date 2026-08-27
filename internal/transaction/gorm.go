// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transaction

import (
	"context"

	"gorm.io/gorm"
)

type GormTransactor struct {
	dbProvider func() *gorm.DB
}

func NewGormTransactor(dbProvider func() *gorm.DB) *GormTransactor {
	return &GormTransactor{
		dbProvider: dbProvider,
	}
}

func (tx *GormTransactor) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	return tx.dbProvider().Transaction(func(gormTx *gorm.DB) error {
		txCtx := context.WithValue(ctx, TxKey, gormTx)

		return fn(txCtx)
	})
}
