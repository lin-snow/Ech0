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

func HasTx(ctx context.Context) bool {
	_, ok := TxFromContext(ctx)
	return ok
}

type writeGuardKey struct{}

// WithWriteGuard binds an optimistic lease check to slow work outside a
// transaction. Repositories check it inside their final write transaction.
func WithWriteGuard(ctx context.Context, guard func(*gorm.DB) error) context.Context {
	return context.WithValue(ctx, writeGuardKey{}, guard)
}
func CheckWrite(ctx context.Context, tx *gorm.DB) error {
	if guard, ok := ctx.Value(writeGuardKey{}).(func(*gorm.DB) error); ok {
		return guard(tx)
	}
	return nil
}
