// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transaction

import (
	"context"

	"gorm.io/gorm"
)

type GormTransactor struct {
	dbProvider func() *gorm.DB
	durable    bool
}

func NewGormTransactor(dbProvider func() *gorm.DB) *GormTransactor {
	return &GormTransactor{
		dbProvider: dbProvider,
	}
}

func NewFunctionTransactor(db func() *gorm.DB) *GormTransactor {
	return &GormTransactor{dbProvider: db, durable: true}
}

type commitKey struct{}
type commitHooks struct {
	durable bool
	hooks   []func()
}

func Durable(ctx context.Context) bool {
	h, _ := ctx.Value(commitKey{}).(*commitHooks)
	return h != nil && h.durable
}

func AfterCommit(ctx context.Context, fn func()) {
	if h, _ := ctx.Value(commitKey{}).(*commitHooks); h != nil {
		h.hooks = append(h.hooks, fn)
		return
	}
	fn()
}

func (tx *GormTransactor) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	if ctx == nil {
		ctx = context.Background()
	}

	db, ok := TxFromContext(ctx)
	if !ok {
		db = tx.dbProvider()
	}
	hooks := &commitHooks{durable: tx.durable}
	err := db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		txCtx := context.WithValue(ctx, TxKey, gormTx)

		return fn(context.WithValue(txCtx, commitKey{}, hooks))
	})
	if err == nil {
		for _, fn := range hooks.hooks {
			AfterCommit(ctx, fn)
		}
	}
	return err
}
