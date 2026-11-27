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

	// A closure running inside an outer transaction joins it: opening a second
	// one on the same SQLite file would take its write lock and deadlock
	// against the first.
	if HasTx(ctx) {
		return fn(ctx)
	}

	return tx.dbProvider().Transaction(func(gormTx *gorm.DB) error {
		return fn(WithTx(ctx, gormTx))
	})
}
