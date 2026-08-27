// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	rps     int
	burst   int
}

type tokenBucket struct {
	tokens    float64
	lastTime  time.Time
	ratePerNs float64
	burst     float64
}

func newRateLimiter(rps, burst int) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		rps:     rps,
		burst:   burst,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{
			tokens:    float64(rl.burst),
			lastTime:  time.Now(),
			ratePerNs: float64(rl.rps) / float64(time.Second),
			burst:     float64(rl.burst),
		}
		rl.buckets[key] = b
	}

	now := time.Now()
	elapsed := now.Sub(b.lastTime)
	b.tokens += float64(elapsed) * b.ratePerNs
	if b.tokens > b.burst {
		b.tokens = b.burst
	}
	b.lastTime = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func RateLimit(rps, burst int) gin.HandlerFunc {
	limiter := newRateLimiter(rps, burst)
	startBucketGC(limiter, 5*time.Minute, 10*time.Minute)

	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RateLimitWithIdempotency(
	rps, burst int,
	dedupTTL time.Duration,
	resourceParam string,
	onIdempotent gin.HandlerFunc,
) gin.HandlerFunc {
	limiter := newRateLimiter(rps, burst)
	dedup := newIdempotencyStore(dedupTTL)

	gcInterval := max(dedupTTL, time.Minute)
	startBucketGC(limiter, gcInterval, 10*time.Minute)
	startIdempotencyGC(dedup, gcInterval)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		resourceID := c.Param(resourceParam)
		if ip != "" && resourceID != "" && !dedup.acquire(ip+"|"+resourceID, time.Now()) {
			onIdempotent(c)
			c.Abort()
			return
		}

		c.Next()
	}
}

func startBucketGC(rl *rateLimiter, interval, idle time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			cutoff := time.Now().Add(-idle)
			for k, b := range rl.buckets {
				if b.lastTime.Before(cutoff) {
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

type idempotencyStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
}

func newIdempotencyStore(ttl time.Duration) *idempotencyStore {
	return &idempotencyStore{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
}

func (s *idempotencyStore) acquire(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.seen[key]; ok && now.Sub(t) < s.ttl {
		return false
	}
	s.seen[key] = now
	return true
}

func (s *idempotencyStore) gc(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, t := range s.seen {
		if now.Sub(t) >= s.ttl {
			delete(s.seen, k)
		}
	}
}

func startIdempotencyGC(s *idempotencyStore, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			s.gc(time.Now())
		}
	}()
}
