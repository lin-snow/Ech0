// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"context"
	"fmt"

	"github.com/lin-snow/ech0/internal/transaction"
)

// ReadThroughUnlessTx bypasses the cache inside a transaction, so reads observe
// uncommitted writes and never populate the cache with them.
func (c *Cache) ReadThroughUnlessTx[T any](
	ctx context.Context,
	key string,
	cost int64,
	txLoad func(context.Context) (T, error),
	load func() (T, error),
) (T, error) {
	if transaction.HasTx(ctx) {
		return txLoad(ctx)
	}
	return c.ReadThrough(key, cost, load)
}

func (c *Cache) ReadThrough[T any](key string, cost int64, load func() (T, error)) (T, error) {
	return c.ReadThroughWithStore(key, func(value T) {
		c.Set(key, value, cost)
	}, load)
}

// ReadThroughWithStore returns the cached T for key, or loads it once across
// concurrent callers and hands the result to store.
func (c *Cache) ReadThroughWithStore[T any](key string, store func(T), load func() (T, error)) (T, error) {
	if !c.RetainsReads() {
		return load()
	}
	if cached, found, err := c.lookup[T](key); err != nil || found {
		return cached, err
	}

	loaded, err, _ := c.loads.Do(key, func() (any, error) {
		if cached, found, err := c.lookup[T](key); err != nil || found {
			return cached, err
		}
		value, err := load()
		if err != nil {
			return nil, err
		}
		store(value)
		return value, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}

	typed, ok := loaded.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("cache read-through type mismatch for key %q", key)
	}
	return typed, nil
}

func (c *Cache) Invalidate(keys ...string) {
	for _, key := range keys {
		c.Delete(key)
	}
}

// lookup reports a hit only when the cached value holds a T; an entry of any
// other type is treated as a miss.
func (c *Cache) lookup[T any](key string) (T, bool, error) {
	var zero T
	cached, found, err := c.Get(key)
	if err != nil || !found {
		return zero, false, err
	}
	typed, ok := cached.(T)
	return typed, ok, nil
}
