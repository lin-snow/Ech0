package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type memCache struct {
	mu   sync.RWMutex
	data map[string]any
}

func newMemCache() *memCache {
	return &memCache{data: make(map[string]any)}
}

func (m *memCache) Set(key string, value any, _ int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return true
}

func (m *memCache) SetWithTTL(key string, value any, _ int64, _ time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return true
}

func (m *memCache) Get(key string) (any, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *memCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *memCache) Close() error { return nil }

func TestReadThroughTyped(t *testing.T) {
	c := newMemCache()
	loads := 0

	v, err := ReadThroughTyped[string](c, "k", 1, func() (string, error) {
		loads++
		return "v", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "v" || loads != 1 {
		t.Fatalf("unexpected first load result, v=%q loads=%d", v, loads)
	}

	v, err = ReadThroughTyped[string](c, "k", 1, func() (string, error) {
		loads++
		return "v2", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "v" || loads != 1 {
		t.Fatalf("expected cached value, v=%q loads=%d", v, loads)
	}
}

func TestWriteAndPopulate(t *testing.T) {
	c := newMemCache()

	err := WriteAndPopulate(c, "k", "v", 1, func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok, err := c.Get("k")
	if err != nil || !ok || got.(string) != "v" {
		t.Fatalf("write and populate failed, ok=%v err=%v got=%v", ok, err, got)
	}

	err = WriteAndPopulate(c, "k2", "v2", 1, func() error { return errors.New("boom") })
	if err == nil {
		t.Fatalf("expected writer error")
	}
}

func TestInvalidateKeys(t *testing.T) {
	c := newMemCache()
	c.Set("a", 1, 1)
	c.Set("b", 2, 1)

	InvalidateKeys(c, "a", "b")

	if _, ok, _ := c.Get("a"); ok {
		t.Fatalf("key a should be invalidated")
	}
	if _, ok, _ := c.Get("b"); ok {
		t.Fatalf("key b should be invalidated")
	}
}

func TestReadThroughTypedSingleflight(t *testing.T) {
	c := newMemCache()
	var calls int32
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := ReadThroughTyped[string](c, "shared", 1, func() (string, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(30 * time.Millisecond)
				return "value", nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if v != "value" {
				t.Errorf("unexpected value: %q", v)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected loader called once, got %d", calls)
	}
}

func TestReadThroughTypedUnlessTxBypassesCache(t *testing.T) {
	c := newMemCache()
	c.Set("k", "cached", 1)

	ctx := context.WithValue(context.Background(), transaction.TxKey, &gorm.DB{})
	calls := 0

	v, err := ReadThroughTypedUnlessTx[string](
		ctx,
		c,
		"k",
		1,
		func(context.Context) (string, error) {
			calls++
			return "tx-value", nil
		},
		func() (string, error) {
			t.Fatal("cached loader should not run inside tx")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "tx-value" {
		t.Fatalf("unexpected tx value: %q", v)
	}
	if calls != 1 {
		t.Fatalf("expected tx loader called once, got %d", calls)
	}
}
