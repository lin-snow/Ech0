package cache

import (
	"testing"
	"time"
)

func TestRistrettoCacheGetMiss(t *testing.T) {
	c, err := NewRistrettoCache[string, string](1000, 1000, 64)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("close cache failed: %v", err)
		}
	}()

	_, found, err := c.Get("missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatalf("expected missing key")
	}
}

func TestRistrettoCacheSetAndGet(t *testing.T) {
	c, err := NewRistrettoCache[string, int](1000, 1000, 64)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("close cache failed: %v", err)
		}
	}()

	ok := c.Set("k", 42, 1)
	if !ok {
		t.Fatalf("set should return true")
	}
	c.cache.Wait()

	v, found, err := c.Get("k")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if !found {
		t.Fatalf("expected key hit")
	}
	if v != 42 {
		t.Fatalf("unexpected value: %d", v)
	}
}

func TestRistrettoCacheSetWithTTL(t *testing.T) {
	c, err := NewRistrettoCache[string, int](1000, 1000, 64)
	if err != nil {
		t.Fatalf("create cache failed: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("close cache failed: %v", err)
		}
	}()

	ok := c.SetWithTTL("ttl-k", 7, 1, 40_000_000) // 40ms
	if !ok {
		t.Fatalf("set with ttl should return true")
	}
	c.cache.Wait()

	if _, found, _ := c.Get("ttl-k"); !found {
		t.Fatalf("expected key hit before ttl expires")
	}

	time.Sleep(80 * time.Millisecond)
	if _, found, _ := c.Get("ttl-k"); found {
		t.Fatalf("expected key miss after ttl expires")
	}
}
