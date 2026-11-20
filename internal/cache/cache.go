// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"time"

	"golang.org/x/sync/singleflight"
)

// Backend is the raw key-value store the application cache sits on.
type Backend interface {
	Set(key string, value any, cost int64) bool
	SetWithTTL(key string, value any, cost int64, ttl time.Duration) bool
	Get(key string) (any, bool, error)
	Delete(key string)
	Close() error
}

// Cache is the application cache: the raw Backend operations plus typed
// read-through loading, coalesced per key. It must not be copied.
type Cache struct {
	Backend
	loads          singleflight.Group
	sharedDatabase bool
}

func New(backend Backend) *Cache {
	return &Cache{Backend: backend}
}

// RetainsReads is false when other instances can change a value independently.
func (c *Cache) RetainsReads() bool { return !c.sharedDatabase }
