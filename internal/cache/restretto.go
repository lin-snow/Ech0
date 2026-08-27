// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type RistrettoCache[K ristretto.Key, V any] struct {
	cache *ristretto.Cache[K, V]
}

func NewRistrettoCache[K ristretto.Key, V any](
	maxCost int64,
	numCounters int64,
	bufferItems int64,
) (*RistrettoCache[K, V], error) {
	cache, err := ristretto.NewCache(&ristretto.Config[K, V]{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
	})
	if err != nil {
		return nil, err
	}
	return &RistrettoCache[K, V]{cache: cache}, nil
}

func (r *RistrettoCache[K, V]) Set(key K, value V, cost int64) bool {
	return r.cache.Set(key, value, cost)
}

func (r *RistrettoCache[K, V]) SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool {
	return r.cache.SetWithTTL(key, value, cost, ttl)
}

func (r *RistrettoCache[K, V]) Get(key K) (V, bool, error) {
	value, found := r.cache.Get(key)
	if !found {
		var zeroValue V
		return zeroValue, false, nil
	}

	return value, true, nil
}

func (r *RistrettoCache[K, V]) Delete(key K) {
	r.cache.Del(key)
}

func (r *RistrettoCache[K, V]) Close() error {
	r.cache.Close()
	return nil
}
