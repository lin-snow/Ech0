// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package ratelimit is the in-process request budget used to protect the API
// from abuse. It keeps the shape of a sliding window counter: every key owns a
// current and a previous window, and the estimate decays linearly as the
// current window advances, so a client that stops hammering recovers smoothly.
//
// Two properties matter for this codebase:
//
//   - Nothing is persisted. Self-hosted Ech0 runs as a single process, and a
//     durable table would let an attacker lock an account across restarts.
//   - Refund exists. Login reserves budget before the password is checked and
//     hands it back once the credentials are proven, so bursts of legitimate
//     sign-ins are not punished while failures keep consuming budget.
package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/wire"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// Scope names one budget table inside a Policy.
type Scope string

const (
	// ScopeAnonymous is the fallback budget shared by every unauthenticated
	// caller, keyed by client address.
	ScopeAnonymous Scope = "anonymous"
	// ScopeAuthenticated is the fallback budget for signed-in callers, keyed
	// by user id so a shared NAT does not pool unrelated users.
	ScopeAuthenticated Scope = "authenticated"
	// ScopeSignInIP bounds sign-in attempts from one address.
	ScopeSignInIP Scope = "signin_ip"
	// ScopeSignInAccount bounds sign-in attempts against one account name.
	// Unknown names are counted too, so the budget cannot be used to probe
	// which accounts exist.
	ScopeSignInAccount Scope = "signin_account"
	// ScopeSignUpIP bounds account creation from one address.
	ScopeSignUpIP Scope = "signup_ip"
	// ScopeRoute is used by middlewares that own a private limiter for a
	// single route (burst control on reactions, MCP and OAuth endpoints).
	ScopeRoute Scope = "route"
)

// Rule is the allowance for one scope.
type Rule struct {
	Limit  int
	Window time.Duration
}

// Policy maps every scope to its rule. Missing scopes are allowed.
type Policy map[Scope]Rule

// DefaultPolicy is the built-in allowance. Numbers follow the shape used by
// comparable single-node software: generous per-request budgets, strict
// budgets on credential entry, and no account lockout.
func DefaultPolicy() Policy {
	return Policy{
		ScopeAnonymous:     {Limit: 300, Window: time.Minute},
		ScopeAuthenticated: {Limit: 600, Window: time.Minute},
		ScopeSignInIP:      {Limit: 30, Window: 15 * time.Minute},
		ScopeSignInAccount: {Limit: 10, Window: 15 * time.Minute},
		ScopeSignUpIP:      {Limit: 10, Window: time.Hour},
	}
}

// Decision is the outcome of one Consume call.
type Decision struct {
	Allowed    bool
	Rule       Rule
	Remaining  int
	RetryAfter time.Duration
}

// Limiter is the budget surface consumed by middlewares and services.
type Limiter interface {
	Consume(scope Scope, key string, cost int) Decision
	Refund(scope Scope, key string, cost int)
}

const (
	defaultCapacity = 10000
	sweepSample     = 32
	warnInterval    = time.Minute
	maxKeyLen       = 128
)

type entry struct {
	rule        Rule
	windowStart time.Time
	current     int
	previous    int
}

// MemoryLimiter keeps every budget in process memory. Its state is bounded:
// when the table is full it first drops entries that can no longer contribute
// to an estimate, then fails open and warns, because degrading rate limiting
// is preferable to rejecting every request or growing without limit.
type MemoryLimiter struct {
	policy   Policy
	now      func() time.Time
	capacity int

	mu       sync.Mutex
	entries  map[string]*entry
	lastWarn atomic.Int64
}

// Option customises a MemoryLimiter.
type Option func(*MemoryLimiter)

// WithClock replaces the time source, for tests.
func WithClock(now func() time.Time) Option {
	return func(l *MemoryLimiter) { l.now = now }
}

// WithCapacity overrides the maximum number of tracked keys.
func WithCapacity(capacity int) Option {
	return func(l *MemoryLimiter) { l.capacity = capacity }
}

// NewMemoryLimiter builds a limiter for the given policy.
func NewMemoryLimiter(policy Policy, opts ...Option) *MemoryLimiter {
	l := &MemoryLimiter{
		policy:   policy,
		now:      time.Now,
		capacity: defaultCapacity,
		entries:  make(map[string]*entry, 64),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

var defaultLimiter struct {
	once sync.Once
	lim  *MemoryLimiter
}

// Default returns the process-wide limiter. It is shared by the HTTP budget
// middleware and the auth service so both spend from the same tables.
func Default() *MemoryLimiter {
	defaultLimiter.once.Do(func() {
		defaultLimiter.lim = NewMemoryLimiter(DefaultPolicy())
	})
	return defaultLimiter.lim
}

// ProvideLimiter exposes the shared limiter to the Wire graph.
func ProvideLimiter() *MemoryLimiter { return Default() }

// ProviderSet wires the limiter into both the server and the handler graphs.
var ProviderSet = wire.NewSet(ProvideLimiter)

// Consume takes cost units from scope/key when the allowance permits. A denied
// call leaves the counters untouched.
func (l *MemoryLimiter) Consume(scope Scope, key string, cost int) Decision {
	rule, ok := l.policy[scope]
	if !ok || rule.Limit <= 0 || rule.Window <= 0 || cost <= 0 || key == "" {
		return Decision{Allowed: true, Rule: rule, Remaining: rule.Limit}
	}

	now := l.now()
	bucket := string(scope) + "\x00" + key

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[bucket]
	if !ok {
		if !l.makeRoomLocked(now) {
			l.warnOverflow(now)
			return Decision{Allowed: true, Rule: rule}
		}
		e = &entry{rule: rule, windowStart: now}
		l.entries[bucket] = e
	}

	used := e.estimate(now, rule)
	if used+cost > rule.Limit {
		remaining := rule.Limit - used
		if remaining < 0 {
			remaining = 0
		}
		return Decision{
			Allowed:    false,
			Rule:       rule,
			Remaining:  remaining,
			RetryAfter: e.retryAfter(now, rule, cost),
		}
	}

	e.current += cost
	remaining := rule.Limit - used - cost
	if remaining < 0 {
		remaining = 0
	}
	return Decision{Allowed: true, Rule: rule, Remaining: remaining}
}

// Refund returns previously consumed units, capped by what the key currently
// holds. It is used when a guarded operation succeeds after reserving budget.
func (l *MemoryLimiter) Refund(scope Scope, key string, cost int) {
	rule, ok := l.policy[scope]
	if !ok || cost <= 0 || key == "" {
		return
	}

	now := l.now()
	bucket := string(scope) + "\x00" + key

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[bucket]
	if !ok {
		return
	}
	e.rotate(now, rule)
	if e.current <= 0 {
		return
	}
	if cost > e.current {
		cost = e.current
	}
	e.current -= cost
}

// rotate advances the window pair. Elapsed windows older than one full window
// are dropped instead of being folded into previous, so an idle key restarts
// from zero.
func (e *entry) rotate(now time.Time, rule Rule) {
	elapsed := now.Sub(e.windowStart)
	if elapsed < rule.Window {
		return
	}
	windows := int(elapsed / rule.Window)
	if windows >= 2 {
		e.windowStart = now
		e.current = 0
		e.previous = 0
		return
	}
	e.previous = e.current
	e.current = 0
	e.windowStart = e.windowStart.Add(time.Duration(windows) * rule.Window)
}

// estimate is the sliding window count: the current window in full plus the
// share of the previous window that still overlaps the sliding window.
func (e *entry) estimate(now time.Time, rule Rule) int {
	e.rotate(now, rule)
	elapsed := now.Sub(e.windowStart)
	if elapsed < 0 {
		elapsed = 0
	}
	if e.previous == 0 {
		return e.current
	}
	weight := 1 - float64(elapsed)/float64(rule.Window)
	if weight < 0 {
		weight = 0
	}
	return e.current + int(math.Ceil(float64(e.previous)*weight))
}

// retryAfter answers "how long until cost more units fit" for the estimate
// above. It is exact for the current window and falls back to the rollover when
// the current window alone is already over the allowance.
func (e *entry) retryAfter(now time.Time, rule Rule, cost int) time.Duration {
	e.rotate(now, rule)
	elapsed := now.Sub(e.windowStart)
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed >= rule.Window {
		return time.Second
	}
	allow := rule.Limit - cost
	if allow < 0 {
		allow = 0
	}

	if e.current <= allow && e.previous > 0 {
		decay := float64(allow-e.current) * float64(rule.Window) / float64(e.previous)
		wait := float64(rule.Window-elapsed) - decay
		if wait > 0 {
			return atLeastSecond(time.Duration(wait))
		}
		return time.Second
	}

	rollover := rule.Window - elapsed
	if e.current <= allow {
		return atLeastSecond(rollover)
	}
	decay := float64(rule.Window) * float64(e.current-allow) / float64(e.current)
	return atLeastSecond(rollover + time.Duration(decay))
}

func atLeastSecond(d time.Duration) time.Duration {
	if d < time.Second {
		return time.Second
	}
	return d
}

// makeRoomLocked drops entries that can no longer contribute to any estimate.
// It reports whether a new key may be inserted.
func (l *MemoryLimiter) makeRoomLocked(now time.Time) bool {
	if l.capacity <= 0 || len(l.entries) < l.capacity {
		return true
	}
	checked := 0
	for key, e := range l.entries {
		if checked >= sweepSample {
			break
		}
		checked++
		if now.Sub(e.windowStart) >= 2*e.rule.Window {
			delete(l.entries, key)
		}
	}
	return len(l.entries) < l.capacity
}

func (l *MemoryLimiter) warnOverflow(now time.Time) {
	last := l.lastWarn.Load()
	if last > 0 && now.Sub(time.Unix(0, last)) < warnInterval {
		return
	}
	if l.lastWarn.CompareAndSwap(last, now.UnixNano()) {
		logUtil.GetLogger().Warn("Rate limit table is full; allowing requests (limiting degraded)")
	}
}

// AccountKey normalises a submitted account name into a stable bucket key and
// digests absurdly long names so the table cannot be filled with them.
func AccountKey(username string) string {
	key := strings.ToLower(strings.TrimSpace(username))
	if len(key) <= maxKeyLen {
		return key
	}
	sum := sha256.Sum256([]byte(key))
	return "sha256:" + hex.EncodeToString(sum[:])
}
