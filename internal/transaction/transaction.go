// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package transaction

import (
	"context"
)

type contextKey string

const TxKey contextKey = "tx"

type Transactor interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
