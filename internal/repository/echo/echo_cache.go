package repository

import (
	"strconv"
	"sync"

	"github.com/lin-snow/ech0/internal/cache"
)

type cacheKeyTracker struct {
	mu   sync.Mutex
	keys map[string]struct{}
}

func newCacheKeyTracker() *cacheKeyTracker {
	return &cacheKeyTracker{
		keys: make(map[string]struct{}),
	}
}

func (t *cacheKeyTracker) Track(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.keys[key] = struct{}{}
}

func (t *cacheKeyTracker) SnapshotAndReset() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	keys := make([]string, 0, len(t.keys))
	for key := range t.keys {
		keys = append(keys, key)
	}
	t.keys = make(map[string]struct{})
	return keys
}

var (
	echoPageCacheKeys  = newCacheKeyTracker()
	todayEchoCacheKeys = newCacheKeyTracker()
)

const (
	EchoPageCacheKeyPrefix = "echo_page" // echo_page:page:pageSize:search:showPrivate
)

func GetEchoPageCacheKey(page, pageSize int, search string, showPrivate bool) string {
	var showPrivateStr string
	if showPrivate {
		showPrivateStr = "true"
	} else {
		showPrivateStr = "false"
	}
	return EchoPageCacheKeyPrefix + ":" + strconv.Itoa(
		page,
	) + ":" + strconv.Itoa(
		pageSize,
	) + ":" + search + ":" + showPrivateStr
}

func ClearEchoPageCache(cache cache.ICache[string, any]) {
	for _, key := range echoPageCacheKeys.SnapshotAndReset() {
		cache.Delete(key)
	}
}

func TrackEchoPageCacheKey(cacheKey string) {
	echoPageCacheKeys.Track(cacheKey)
}

func TrackTodayEchosCacheKey(cacheKey string) {
	todayEchoCacheKeys.Track(cacheKey)
}

func ClearTodayEchosCache(cache cache.ICache[string, any]) {
	for _, key := range todayEchoCacheKeys.SnapshotAndReset() {
		cache.Delete(key)
	}
}

func GetEchoByIDCacheKey(id uint) string {
	return "echo_id:" + strconv.Itoa(int(id))
}

func GetTodayEchosCacheKey(showPrivate bool, timezone string) string {
	return "echo_today:" + strconv.FormatBool(showPrivate) + ":" + timezone
}
