// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package memstore

import (
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/store"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	st := New(Options{GCInterval: time.Hour})
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func validSite(key string) store.Site {
	return store.Site{
		SiteKey:    key,
		SecretHash: []byte("secret-hash"),
		JWTSecret:  []byte("jwt-secret"),
	}
}

func TestTryMarkChallengeSigUsed(t *testing.T) {
	st := newTestStore(t)
	base := time.Unix(1000, 0)

	t.Run("first mark succeeds", func(t *testing.T) {
		ok, err := st.TryMarkChallengeSigUsed("sig-active", base.Add(time.Minute), base)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if !ok {
			t.Fatalf("first mark should write and return true")
		}
	})

	t.Run("re-mark while active is rejected", func(t *testing.T) {
		ok, err := st.TryMarkChallengeSigUsed("sig-active", base.Add(time.Hour), base)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if ok {
			t.Fatalf("re-marking a still-active sig must return false")
		}
	})

	t.Run("expired sig can be re-marked", func(t *testing.T) {
		if ok, err := st.TryMarkChallengeSigUsed("sig-exp", base, base.Add(-time.Second)); err != nil || !ok {
			t.Fatalf("seed mark = (%v, %v), want (true, nil)", ok, err)
		}
		later := base.Add(time.Second)
		ok, err := st.TryMarkChallengeSigUsed("sig-exp", later.Add(time.Minute), later)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if !ok {
			t.Fatalf("expired sig should be re-markable (true)")
		}
	})
}

func TestUpsertSiteValidation(t *testing.T) {
	st := newTestStore(t)

	cases := []struct {
		name    string
		site    store.Site
		wantSub string
	}{
		{
			name:    "missing site key",
			site:    store.Site{SecretHash: []byte("h"), JWTSecret: []byte("j")},
			wantSub: "site key",
		},
		{
			name:    "missing secret hash",
			site:    store.Site{SiteKey: "s", JWTSecret: []byte("j")},
			wantSub: "secret hash",
		},
		{
			name:    "missing jwt secret",
			site:    store.Site{SiteKey: "s", SecretHash: []byte("h")},
			wantSub: "jwt secret",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := st.UpsertSite(tc.site)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q should mention %q", err.Error(), tc.wantSub)
			}
		})
	}

	t.Run("valid site succeeds", func(t *testing.T) {
		if err := st.UpsertSite(validSite("ok")); err != nil {
			t.Fatalf("valid upsert err: %v", err)
		}
		if _, ok := st.GetSite("ok"); !ok {
			t.Fatalf("stored site should be retrievable")
		}
	})
}

func TestUpsertSiteDefensiveCopy(t *testing.T) {
	st := newTestStore(t)

	secret := []byte("secret-hash")
	jwt := []byte("jwt-secret")
	site := store.Site{SiteKey: "copy", SecretHash: secret, JWTSecret: jwt}
	if err := st.UpsertSite(site); err != nil {
		t.Fatalf("upsert err: %v", err)
	}

	secret[0] = 'X'
	jwt[0] = 'Y'

	got, ok := st.GetSite("copy")
	if !ok {
		t.Fatalf("site not found")
	}
	if string(got.SecretHash) != "secret-hash" {
		t.Fatalf("SecretHash mutated through caller slice: %q", got.SecretHash)
	}
	if string(got.JWTSecret) != "jwt-secret" {
		t.Fatalf("JWTSecret mutated through caller slice: %q", got.JWTSecret)
	}
}

func TestGetSiteReturnsIsolatedCopy(t *testing.T) {
	st := newTestStore(t)
	if err := st.UpsertSite(validSite("iso")); err != nil {
		t.Fatalf("upsert err: %v", err)
	}

	first, ok := st.GetSite("iso")
	if !ok {
		t.Fatalf("site not found")
	}
	first.SecretHash[0] = 'Z'
	first.JWTSecret[0] = 'Z'

	second, ok := st.GetSite("iso")
	if !ok {
		t.Fatalf("site not found on second read")
	}
	if string(second.SecretHash) != "secret-hash" {
		t.Fatalf("SecretHash leaked mutation across GetSite calls: %q", second.SecretHash)
	}
	if string(second.JWTSecret) != "jwt-secret" {
		t.Fatalf("JWTSecret leaked mutation across GetSite calls: %q", second.JWTSecret)
	}
}

func TestAllowRateLimitWindowRolling(t *testing.T) {
	st := newTestStore(t)
	now := time.Unix(2000, 0)

	allowed, remaining, err := st.AllowRateLimit("scope", "k", 1, time.Second, now)
	if err != nil || !allowed {
		t.Fatalf("first request allowed=(%v) err=%v, want allowed", allowed, err)
	}
	if remaining != 0 {
		t.Fatalf("remaining after first = %d, want 0", remaining)
	}

	allowed, remaining, err = st.AllowRateLimit("scope", "k", 1, time.Second, now)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if allowed {
		t.Fatalf("second request in same window should be blocked")
	}
	if remaining != 0 {
		t.Fatalf("remaining when over limit = %d, want 0 (clamped)", remaining)
	}

	next := now.Add(time.Second)
	allowed, _, err = st.AllowRateLimit("scope", "k", 1, time.Second, next)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !allowed {
		t.Fatalf("request in next window should be allowed again")
	}
}

func TestAllowRateLimitShortCircuit(t *testing.T) {
	st := newTestStore(t)
	now := time.Unix(3000, 0)

	cases := []struct {
		name   string
		max    int
		window time.Duration
	}{
		{name: "max zero", max: 0, window: time.Second},
		{name: "max negative", max: -5, window: time.Second},
		{name: "window zero", max: 5, window: 0},
		{name: "window negative", max: 5, window: -time.Second},
		{name: "window sub-millisecond rounds to zero", max: 7, window: time.Microsecond},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, remaining, err := st.AllowRateLimit("scope", "sc", tc.max, tc.window, now)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !allowed {
				t.Fatalf("short-circuit should always allow")
			}
			if remaining != tc.max {
				t.Fatalf("short-circuit remaining = %d, want max %d", remaining, tc.max)
			}
		})
	}
}

func TestGCOnceRemovesExpiredEntries(t *testing.T) {
	st := newTestStore(t)
	base := time.Unix(4000, 0)
	gcAt := base.Add(time.Minute)

	if ok, err := st.TryMarkChallengeSigUsed("sig-live", base.Add(time.Hour), base); err != nil || !ok {
		t.Fatalf("seed live sig: ok=%v err=%v", ok, err)
	}
	if ok, err := st.TryMarkChallengeSigUsed("sig-dead", base, base.Add(-time.Second)); err != nil || !ok {
		t.Fatalf("seed dead sig: ok=%v err=%v", ok, err)
	}

	if err := st.StoreRedeemToken("site", "tok-live", base.Add(time.Hour)); err != nil {
		t.Fatalf("store live token: %v", err)
	}
	if err := st.StoreRedeemToken("site", "tok-dead", base); err != nil {
		t.Fatalf("store dead token: %v", err)
	}

	if allowed, _, err := st.AllowRateLimit("keep", "k", 5, time.Hour, gcAt); err != nil || !allowed {
		t.Fatalf("seed kept window: allowed=%v err=%v", allowed, err)
	}
	if allowed, _, err := st.AllowRateLimit("drop", "k", 5, time.Second, base); err != nil || !allowed {
		t.Fatalf("seed dropped window: allowed=%v err=%v", allowed, err)
	}

	st.gcOnce(gcAt)

	if ok, _ := st.TryMarkChallengeSigUsed("sig-live", base.Add(2*time.Hour), gcAt); ok {
		t.Fatalf("live sig should have survived GC and remain active")
	}

	if found, expired, _ := st.ConsumeRedeemToken("site", "tok-live", gcAt); !found || expired {
		t.Fatalf("live token should survive GC: found=%v expired=%v", found, expired)
	}
	if found, _, _ := st.ConsumeRedeemToken("site", "tok-dead", base); found {
		t.Fatalf("expired token should have been removed by GC (found=false), got found=true")
	}

	if _, remaining, _ := st.AllowRateLimit("keep", "k", 5, time.Hour, gcAt); remaining != 3 {
		t.Fatalf("kept window remaining = %d, want 3 (count preserved across GC)", remaining)
	}
	if _, remaining, _ := st.AllowRateLimit("drop", "k", 5, time.Second, base); remaining != 4 {
		t.Fatalf("dropped window remaining = %d, want 4 (count reset after GC)", remaining)
	}
}
