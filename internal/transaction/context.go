// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transaction

import (
	"context"

	"gorm.io/gorm"
)

func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	if ctx == nil {
		return nil, false
	}

	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	return tx, ok
}

// WithTx returns ctx carrying tx as the transaction every repository must use.
// Code that opens its own transaction (boot-time migrations) calls it so the
// same write goes through the repositories and their caches.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, TxKey, tx)
}

func HasTx(ctx context.Context) bool {
	_, ok := TxFromContext(ctx)
	return ok
}
