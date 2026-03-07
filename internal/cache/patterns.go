package cache

import (
	"context"
	"fmt"

	"github.com/lin-snow/ech0/internal/transaction"
	"golang.org/x/sync/singleflight"
)

var readThroughGroup singleflight.Group

// ReadThroughTypedUnlessTx 在事务中直接走 txLoader，避免读到事务外缓存。
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

// ReadThroughTyped 统一读穿透模式：先查缓存，未命中后走 loader 并回填缓存。
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

// ReadThroughTypedWithStore 支持自定义回填逻辑（如 TTL 回填）。
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
		// double-check，避免并发下重复 load
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

// WriteAndPopulate 统一写后回填模式：先写存储，成功后回填缓存。
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

// InvalidateKeys 统一批量失效模式。
func InvalidateKeys(c ICache[string, any], keys ...string) {
	for _, key := range keys {
		c.Delete(key)
	}
}
