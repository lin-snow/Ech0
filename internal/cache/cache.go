package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// ICache 定义了缓存接口，提供基本的缓存操作方法
type ICache[K ristretto.Key, V any] interface {
	Set(key K, value V, cost int64) bool
	SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool
	Get(key K) (V, bool, error)
	Delete(key K)
	Close() error
}

// NewCache 创建一个新的缓存实例，使用 Ristretto 作为缓存实现
func NewCache[K ristretto.Key, V any]() (ICache[K, V], error) {
	return NewRistrettoCache[K, V](1000000, 1000000, 100)
}
