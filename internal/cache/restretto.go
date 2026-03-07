package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// RistrettoCache 是基于 Ristretto 实现的缓存结构体
type RistrettoCache[K ristretto.Key, V any] struct {
	cache *ristretto.Cache[K, V]
}

// NewRistrettoCache 创建一个新的 RistrettoCache 实例
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

// Set 将键值对存入缓存
func (r *RistrettoCache[K, V]) Set(key K, value V, cost int64) bool {
	return r.cache.Set(key, value, cost)
}

// SetWithTTL 将键值对存入缓存，并设置过期时间
func (r *RistrettoCache[K, V]) SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool {
	return r.cache.SetWithTTL(key, value, cost, ttl)
}

// Get 从缓存中获取值
func (r *RistrettoCache[K, V]) Get(key K) (V, bool, error) {
	value, found := r.cache.Get(key)
	if !found {
		var zeroValue V
		return zeroValue, false, nil
	}

	return value, true, nil
}

// Delete 从缓存中删除指定的键
func (r *RistrettoCache[K, V]) Delete(key K) {
	r.cache.Del(key)
}

// Close 关闭底层缓存资源
func (r *RistrettoCache[K, V]) Close() error {
	r.cache.Close()
	return nil
}
