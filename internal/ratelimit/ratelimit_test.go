// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package ratelimit

import (
	"strings"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time          { return f.now }
func (f *fakeClock) Advance(d time.Duration) { f.now = f.now.Add(d) }

func newLimiter(policy Policy, capacity int) (*MemoryLimiter, *fakeClock) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	return NewMemoryLimiter(policy, WithClock(clock.Now), WithCapacity(capacity)), clock
}

func TestMemoryLimiterWindow(t *testing.T) {
	limiter, clock := newLimiter(Policy{ScopeRoute: {Limit: 2, Window: time.Minute}}, 16)

	for i := range 2 {
		if d := limiter.Consume(ScopeRoute, "k", 1); !d.Allowed {
			t.Fatalf("request %d should be allowed, remaining=%d retryAfter=%s", i, d.Remaining, d.RetryAfter)
		}
	}

	denied := limiter.Consume(ScopeRoute, "k", 1)
	if denied.Allowed {
		t.Fatal("third request should be denied")
	}
	// The current window cannot decay, so the wait spans the rollover plus the
	// decay of the previous window.
	if denied.RetryAfter <= 0 || denied.RetryAfter > 2*time.Minute {
		t.Fatalf("RetryAfter = %s, want (0, 2m]", denied.RetryAfter)
	}

	// The current window never decays early, so the wait spans the rollover.
	clock.Advance(30 * time.Second)
	if d := limiter.Consume(ScopeRoute, "k", 1); d.Allowed {
		t.Fatal("request before the window rolls over should be denied")
	}

	// Once the window has rolled over, the previous window decays linearly and
	// frees the first slot at the half-way point of the new window.
	clock.Advance(60 * time.Second)
	if d := limiter.Consume(ScopeRoute, "k", 1); !d.Allowed {
		t.Fatalf("decayed slot should be usable, retryAfter=%s", d.RetryAfter)
	}

	// A full window after the last rotation all consumption is gone.
	clock.Advance(2 * time.Minute)
	if d := limiter.Consume(ScopeRoute, "k", 1); !d.Allowed {
		t.Fatalf("request after full window should be allowed, retryAfter=%s", d.RetryAfter)
	}
}

func TestMemoryLimiterRefund(t *testing.T) {
	limiter, _ := newLimiter(Policy{ScopeSignInAccount: {Limit: 1, Window: time.Minute}}, 8)

	if d := limiter.Consume(ScopeSignInAccount, "alice", 1); !d.Allowed {
		t.Fatal("first attempt should be allowed")
	}
	if d := limiter.Consume(ScopeSignInAccount, "alice", 1); d.Allowed {
		t.Fatal("second attempt should be denied")
	}

	limiter.Refund(ScopeSignInAccount, "alice", 1)
	if d := limiter.Consume(ScopeSignInAccount, "alice", 1); !d.Allowed {
		t.Fatal("attempt after refund should be allowed")
	}

	// Refunding an unknown key must not create credit.
	limiter.Refund(ScopeSignInAccount, "bob", 100)
	if d := limiter.Consume(ScopeSignInAccount, "bob", 1); !d.Allowed {
		t.Fatal("unknown key should start with a full budget")
	}
}

func TestMemoryLimiterSeparatesScopes(t *testing.T) {
	limiter, _ := newLimiter(Policy{
		ScopeSignInIP:      {Limit: 1, Window: time.Minute},
		ScopeSignInAccount: {Limit: 1, Window: time.Minute},
	}, 8)

	if d := limiter.Consume(ScopeSignInIP, "alice", 1); !d.Allowed {
		t.Fatal("signin_ip should be allowed")
	}
	if d := limiter.Consume(ScopeSignInAccount, "alice", 1); !d.Allowed {
		t.Fatal("signin_account should not share the signin_ip bucket")
	}
}

func TestMemoryLimiterFailsOpenWhenFull(t *testing.T) {
	limiter, _ := newLimiter(Policy{ScopeAnonymous: {Limit: 1, Window: time.Minute}}, 1)

	if d := limiter.Consume(ScopeAnonymous, "first", 1); !d.Allowed {
		t.Fatal("first key should be allowed")
	}
	if d := limiter.Consume(ScopeAnonymous, "second", 1); !d.Allowed {
		t.Fatal("unknown key should be allowed while the table is full")
	}
	if d := limiter.Consume(ScopeAnonymous, "first", 1); d.Allowed {
		t.Fatal("tracked key must still be limited while degraded")
	}
}

func TestMemoryLimiterUnconfiguredScopeIsUnlimited(t *testing.T) {
	limiter, _ := newLimiter(Policy{}, 8)
	for range 100 {
		if d := limiter.Consume(ScopeAuthenticated, "u1", 1); !d.Allowed {
			t.Fatal("unconfigured scope must not limit")
		}
	}
}

func TestAccountKey(t *testing.T) {
	if got, want := AccountKey("  Alice "), "alice"; got != want {
		t.Fatalf("AccountKey = %q, want %q", got, want)
	}

	long := strings.Repeat("x", maxKeyLen+1)
	key := AccountKey(long)
	if !strings.HasPrefix(key, "sha256:") {
		t.Fatalf("long key should be hashed, got %q", key)
	}
	if key != AccountKey(strings.ToUpper(long)) {
		t.Fatal("long key digest must be case-insensitive")
	}
}

func TestDecisionRemaining(t *testing.T) {
	limiter, _ := newLimiter(Policy{ScopeAuthenticated: {Limit: 3, Window: time.Minute}}, 8)
	if d := limiter.Consume(ScopeAuthenticated, "u1", 1); d.Remaining != 2 {
		t.Fatalf("Remaining = %d, want 2", d.Remaining)
	}
	if d := limiter.Consume(ScopeAuthenticated, "u1", 5); d.Allowed {
		t.Fatal("cost above the limit should be denied")
	}
}
