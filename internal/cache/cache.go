// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type ICache[K ristretto.Key, V any] interface {
	Set(key K, value V, cost int64) bool
	SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool
	Get(key K) (V, bool, error)
	Delete(key K)
	Close() error
}

func NewCache[K ristretto.Key, V any]() (ICache[K, V], error) {
	return NewRistrettoCache[K, V](1000000, 1000000, 100)
}
