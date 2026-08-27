// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"context"
	"fmt"

	"github.com/lin-snow/ech0/internal/transaction"
	"golang.org/x/sync/singleflight"
)

var readThroughGroup singleflight.Group

func ReadThroughTypedUnlessTx[T any](
	ctx context.Context,
	c ICache[string, any],
	key string,
	cost int64,
	txLoader func(context.Context) (T, error),
	loader func() (T, error),
) (T, error) {
	if transaction.HasTx(ctx) {
		return txLoader(ctx)
	}

	return ReadThroughTyped(c, key, cost, loader)
}

func ReadThroughTyped[T any](
	c ICache[string, any],
	key string,
	cost int64,
	loader func() (T, error),
) (T, error) {
	return ReadThroughTypedWithStore(c, key, func(value T) {
		c.Set(key, value, cost)
	}, loader)
}

func ReadThroughTypedWithStore[T any](
	c ICache[string, any],
	key string,
	store func(value T),
	loader func() (T, error),
) (T, error) {
	if cached, found, err := c.Get(key); err != nil {
		var zero T
		return zero, err
	} else if found {
		if typed, ok := cached.(T); ok {
			return typed, nil
		}
	}

	loaded, err, _ := readThroughGroup.Do(key, func() (any, error) {
		if cached, found, cacheErr := c.Get(key); cacheErr != nil {
			return nil, cacheErr
		} else if found {
			if typed, ok := cached.(T); ok {
				return typed, nil
			}
		}

		value, loadErr := loader()
		if loadErr != nil {
			return nil, loadErr
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

func WriteAndPopulate(
	c ICache[string, any],
	key string,
	value any,
	cost int64,
	writer func() error,
) error {
	if err := writer(); err != nil {
		return err
	}
	c.Set(key, value, cost)
	return nil
}

func InvalidateKeys(c ICache[string, any], keys ...string) {
	for _, key := range keys {
		c.Delete(key)
	}
}
