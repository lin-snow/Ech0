// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package kvstore

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("kvstore: key not found")

type Store interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type Backend interface {
	GetKeyValue(ctx context.Context, key string) (string, error)
	AddOrUpdateKeyValue(ctx context.Context, key, value string) error
	DeleteKeyValue(ctx context.Context, key string) error
}
