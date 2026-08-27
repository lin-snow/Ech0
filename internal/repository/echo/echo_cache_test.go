// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

type testCache struct {
	mu      sync.Mutex
	deleted map[string]struct{}
}

func newTestCache() *testCache {
	return &testCache{deleted: make(map[string]struct{})}
}

func (t *testCache) Set(string, any, int64) bool { return true }
func (t *testCache) SetWithTTL(string, any, int64, time.Duration) bool {
	return true
}
func (t *testCache) Get(string) (any, bool, error) { return nil, false, nil }
func (t *testCache) Delete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.deleted[key] = struct{}{}
}
func (t *testCache) Close() error { return nil }

func (t *testCache) deletedCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.deleted)
}

func TestEchoCacheKeyTrackerConcurrentTrackAndClear(t *testing.T) {
	const n = 200
	cache := newTestCache()
	var wg sync.WaitGroup

	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			TrackEchoPageCacheKey("echo_page:" + strconv.Itoa(i))
			TrackTodayEchosCacheKey("echo_today:" + strconv.Itoa(i))
		}(i)
	}
	wg.Wait()

	ClearEchoPageCache(cache)
	ClearTodayEchosCache(cache)

	if cache.deletedCount() != 2*n {
		t.Fatalf("expected %d deleted keys, got %d", 2*n, cache.deletedCount())
	}

	ClearEchoPageCache(cache)
	ClearTodayEchosCache(cache)
	if cache.deletedCount() != 2*n {
		t.Fatalf("expected stable deleted count %d, got %d", 2*n, cache.deletedCount())
	}
}
