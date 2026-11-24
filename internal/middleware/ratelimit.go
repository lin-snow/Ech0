// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/httperr"
	"github.com/lin-snow/ech0/internal/ratelimit"
	"github.com/lin-snow/ech0/internal/util/clientip"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// Budget is the fallback allowance applied to every API request: one shared
// budget per anonymous address, one per signed-in user. It exists so endpoints
// nobody thought to protect still have a ceiling, and it must run after the
// request identity is resolved.
func Budget(limits *ratelimit.MemoryLimiter) gin.HandlerFunc {
	if limits == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		scope, key := ratelimit.ScopeAnonymous, clientAddress(c)
		if v, ok := viewer.FromContext(c.Request.Context()); ok && v.UserID() != "" {
			scope, key = ratelimit.ScopeAuthenticated, v.UserID()
		}
		if key == "" {
			c.Next()
			return
		}
		if decision := limits.Consume(scope, key, 1); !decision.Allowed {
			writeRateLimited(c, decision)
			return
		}
		c.Next()
	}
}

// ScopeLimit charges one named budget for every request, keyed by the client
// address. Registration uses it for its per-address allowance.
func ScopeLimit(limits *ratelimit.MemoryLimiter, scope ratelimit.Scope, cost int) gin.HandlerFunc {
	if limits == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		key := clientAddress(c)
		if key == "" {
			c.Next()
			return
		}
		if decision := limits.Consume(scope, key, cost); !decision.Allowed {
			writeRateLimited(c, decision)
			return
		}
		c.Next()
	}
}

// RateLimit limits each client address to rps requests per second with a
// burst allowance. The limiter is private to the route that created it.
func RateLimit(rps, burst int) gin.HandlerFunc {
	return RateLimitN(rps, time.Second, burst)
}

// RateLimitN limits each client address to limit requests per window, tolerating
// up to burst requests at once.
func RateLimitN(limit int, per time.Duration, burst int) gin.HandlerFunc {
	if limit <= 0 {
		limit = 1
	}
	if per <= 0 {
		per = time.Second
	}
	if burst < limit {
		burst = limit
	}
	// A window of burst/rate holds exactly burst requests and refills at the
	// requested sustained rate.
	window := time.Duration(float64(burst) / float64(limit) * float64(per))
	limits := ratelimit.NewMemoryLimiter(ratelimit.Policy{
		ratelimit.ScopeRoute: {Limit: burst, Window: window},
	})

	return func(c *gin.Context) {
		key := clientAddress(c)
		if key == "" {
			c.Next()
			return
		}
		if decision := limits.Consume(ratelimit.ScopeRoute, key, 1); !decision.Allowed {
			writeRateLimited(c, decision)
			return
		}
		c.Next()
	}
}

// clientAddress prefers the address resolved once per request by ClientIP and
// falls back to the transport peer for callers that did not install it.
func clientAddress(c *gin.Context) string {
	if ip := clientip.FromContext(c.Request.Context()); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return host
	}
	return c.Request.RemoteAddr
}

func writeRateLimited(c *gin.Context, decision ratelimit.Decision) {
	seconds := int((decision.RetryAfter + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(seconds))

	httperr.Abort(c, apperr.ErrRateLimited)
}

func RateLimitWithIdempotency(
	rps, burst int,
	dedupTTL time.Duration,
	resourceParam string,
	onIdempotent gin.HandlerFunc,
) gin.HandlerFunc {
	if rps <= 0 {
		rps = 1
	}
	if burst < rps {
		burst = rps
	}
	limits := ratelimit.NewMemoryLimiter(ratelimit.Policy{
		ratelimit.ScopeRoute: {Limit: burst, Window: time.Duration(burst) * time.Second / time.Duration(rps)},
	})
	dedup := newIdempotencyStore(dedupTTL)

	gcInterval := max(dedupTTL, time.Minute)
	startIdempotencyGC(dedup, gcInterval)

	return func(c *gin.Context) {
		ip := clientAddress(c)
		if ip != "" {
			if decision := limits.Consume(ratelimit.ScopeRoute, ip, 1); !decision.Allowed {
				writeRateLimited(c, decision)
				return
			}
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
