// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
)

func NewTestCache() cache.ICache[string, any] {
	return &testCache{m: make(map[string]any)}
}

type testCache struct {
	mu sync.RWMutex
	m  map[string]any
}

func (c *testCache) Set(key string, value any, _ int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return true
}

func (c *testCache) SetWithTTL(key string, value any, cost int64, _ time.Duration) bool {
	return c.Set(key, value, cost)
}

func (c *testCache) Get(key string) (any, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok, nil
}

func (c *testCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

func (c *testCache) Close() error { return nil }
