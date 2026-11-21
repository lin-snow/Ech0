// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	spy := newTestCache()
	c := cache.New(spy)
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

	ClearEchoPageCache(c)
	ClearTodayEchosCache(c)

	if spy.deletedCount() != 2*n {
		t.Fatalf("expected %d deleted keys, got %d", 2*n, spy.deletedCount())
	}

	ClearEchoPageCache(c)
	ClearTodayEchosCache(c)
	if spy.deletedCount() != 2*n {
		t.Fatalf("expected stable deleted count %d, got %d", 2*n, spy.deletedCount())
	}
}

// D1 requires that an anonymous request can never be served the private-echo
// payload cached by a signed-in request (and vice versa). The keys are derived
// from PrivateScope, so this exercises the cache end to end instead of only
// comparing strings.
func TestEchoCacheDoesNotCrossVisibilityScopes(t *testing.T) {
	repo, db := newEchoRepo(t)
	now := time.Now().Unix()
	seedEcho(t, db, "public-1", "public echo", false, now-10)
	seedEcho(t, db, "private-1", "private echo", true, now-5)

	anonymous := echoModel.PrivateScope{}
	authed := echoModel.PrivateScope{Authenticated: true}

	anonPage, anonTotal := repo.GetEchosByPage(1, 10, "", anonymous)
	require.Equal(t, int64(1), anonTotal)
	require.Len(t, anonPage, 1)
	assert.Equal(t, "public-1", anonPage[0].ID)

	authedPage, authedTotal := repo.GetEchosByPage(1, 10, "", authed)
	require.Equal(t, int64(2), authedTotal)
	require.Len(t, authedPage, 2)

	// Re-read from cache: the anonymous key must still only hold public data.
	anonPage, anonTotal = repo.GetEchosByPage(1, 10, "", anonymous)
	require.Equal(t, int64(1), anonTotal)
	require.Len(t, anonPage, 1)
	assert.Equal(t, "public-1", anonPage[0].ID)

	anonToday := repo.GetTodayEchos(anonymous, "UTC")
	for _, echo := range anonToday {
		assert.False(t, echo.Private, "anonymous today cache leaked a private echo")
	}
	authedToday := repo.GetTodayEchos(authed, "UTC")
	assert.Len(t, authedToday, 2)

	// And the signed-in cache entry survives the anonymous read unchanged.
	authedPage, authedTotal = repo.GetEchosByPage(1, 10, "", authed)
	require.Equal(t, int64(2), authedTotal)
	require.Len(t, authedPage, 2)
}
