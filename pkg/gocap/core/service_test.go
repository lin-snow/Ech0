// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/store"
	"github.com/lin-snow/ech0/pkg/gocap/store/memstore"
)

func TestChallengeRedeemSiteVerifyFlow(t *testing.T) {
	st := memstore.New(memstore.Options{GCInterval: 10 * time.Second})
	defer func() {
		_ = st.Close()
	}()

	pepper := []byte("pepper")
	site := store.Site{
		SiteKey:        "site1",
		SecretHash:     HashSecret("secret1", pepper),
		JWTSecret:      []byte("jwt-secret"),
		Difficulty:     1,
		ChallengeCount: 2,
		SaltSize:       8,
	}
	if err := st.UpsertSite(site); err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1700000000, 0)
	svc := NewService(st, ServiceOptions{
		ChallengeTTL: 5 * time.Minute,
		RedeemTTL:    10 * time.Minute,
		Now:          func() time.Time { return now },
		SecretPepper: pepper,
	})

	chal, err := svc.CreateChallenge("site1")
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	solutions := bruteSolutions(chal.Token, chal.Challenge.C, chal.Challenge.S, chal.Challenge.D)

	rdm, err := svc.Redeem("site1", RedeemRequest{
		Token:     chal.Token,
		Solutions: solutions,
	})
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if !rdm.Success {
		t.Fatalf("expected success redeem")
	}

	if _, err := svc.Redeem("site1", RedeemRequest{Token: chal.Token, Solutions: solutions}); err == nil {
		t.Fatalf("expected replay redeem to fail")
	}

	verifyResp, err := svc.SiteVerify("site1", SiteVerifyRequest{
		Secret:   "secret1",
		Response: rdm.Token,
	})
	if err != nil {
		t.Fatalf("siteverify: %v", err)
	}
	if !verifyResp.Success {
		t.Fatalf("expected siteverify success")
	}

	if _, err := svc.SiteVerify("site1", SiteVerifyRequest{
		Secret:   "secret1",
		Response: rdm.Token,
	}); err == nil {
		t.Fatalf("expected second siteverify to fail")
	}
}

func TestRedeemConcurrentReplayOnlyOneSuccess(t *testing.T) {
	st := memstore.New(memstore.Options{GCInterval: 10 * time.Second})
	defer func() {
		_ = st.Close()
	}()

	pepper := []byte("pepper")
	site := store.Site{
		SiteKey:        "site1",
		SecretHash:     HashSecret("secret1", pepper),
		JWTSecret:      []byte("jwt-secret"),
		Difficulty:     1,
		ChallengeCount: 1,
		SaltSize:       8,
	}
	if err := st.UpsertSite(site); err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1700000000, 0)
	svc := NewService(st, ServiceOptions{
		ChallengeTTL: 5 * time.Minute,
		RedeemTTL:    10 * time.Minute,
		Now:          func() time.Time { return now },
		SecretPepper: pepper,
	})

	chal, err := svc.CreateChallenge("site1")
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	solutions := bruteSolutions(chal.Token, chal.Challenge.C, chal.Challenge.S, chal.Challenge.D)

	var success atomic.Int32
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			_, err := svc.Redeem("site1", RedeemRequest{
				Token:     chal.Token,
				Solutions: solutions,
			})
			if err == nil {
				success.Add(1)
			}
		})
	}
	wg.Wait()

	if got := success.Load(); got != 1 {
		t.Fatalf("expected exactly 1 success, got %d", got)
	}
}

func TestChallengeExpire(t *testing.T) {
	st := memstore.New(memstore.Options{GCInterval: 10 * time.Second})
	defer func() {
		_ = st.Close()
	}()

	pepper := []byte("pepper")
	site := store.Site{
		SiteKey:        "site1",
		SecretHash:     HashSecret("secret1", pepper),
		JWTSecret:      []byte("jwt-secret"),
		Difficulty:     1,
		ChallengeCount: 1,
		SaltSize:       8,
	}
	if err := st.UpsertSite(site); err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1700000000, 0)
	svc := NewService(st, ServiceOptions{
		ChallengeTTL: 1 * time.Second,
		RedeemTTL:    5 * time.Minute,
		Now:          func() time.Time { return now },
		SecretPepper: pepper,
	})
	chal, err := svc.CreateChallenge("site1")
	if err != nil {
		t.Fatal(err)
	}

	solutions := bruteSolutions(chal.Token, chal.Challenge.C, chal.Challenge.S, chal.Challenge.D)
	now = now.Add(2 * time.Second)
	if _, err := svc.Redeem("site1", RedeemRequest{Token: chal.Token, Solutions: solutions}); err == nil {
		t.Fatalf("expected expired challenge to fail")
	}
}

func bruteSolutions(seed string, count, saltSize, difficulty int) []int {
	pairs := BuildChallengePairs(seed, count, saltSize, difficulty)
	out := make([]int, count)
	for i, p := range pairs {
		out[i] = bruteOne(p[0], p[1])
	}
	return out
}

func bruteOne(salt, target string) int {
	for nonce := range 10_000_000 {
		sum := sha256.Sum256([]byte(salt + intToString(nonce)))
		h := hex.EncodeToString(sum[:])
		if len(target) <= len(h) && h[:len(target)] == target {
			return nonce
		}
	}
	return 0
}

func intToString(v int) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
