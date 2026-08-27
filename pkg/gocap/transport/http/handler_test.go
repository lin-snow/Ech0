// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package caphttp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/core"
	"github.com/lin-snow/ech0/pkg/gocap/store"
	"github.com/lin-snow/ech0/pkg/gocap/store/memstore"
)

func TestPathStrictAndMethodAllow(t *testing.T) {
	h, _, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/site1/challenge", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow == "" {
		t.Fatalf("expected Allow header")
	}

	req = httptest.NewRequest(http.MethodPost, "/x/site1/challenge", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected strict 404 for invalid path, got %d", rec.Code)
	}
}

func TestRedeemCompatUnknownFieldsAndSiteverifyNoRateLimit(t *testing.T) {
	h, _, _ := newTestHandler(t)

	challengeToken, challenge := requestChallenge(t, h, "/site1/challenge")
	solution := bruteOneLocal(challengeToken, challenge.C, challenge.S, challenge.D)

	body, _ := json.Marshal(map[string]any{
		"token":         challengeToken,
		"solutions":     []int{solution},
		"instr_timeout": true,
		"instr_blocked": false,
		"unknown_extra": "ok",
	})
	req := httptest.NewRequest(http.MethodPost, "/site1/redeem", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("redeem expected 200 got %d body=%s", rec.Code, rec.Body.String())
	}

	var redeemResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &redeemResp)
	if redeemResp.Token == "" {
		t.Fatalf("missing redeem token")
	}

	verifyBody, _ := json.Marshal(map[string]any{
		"secret":   "secret1",
		"response": redeemResp.Token,
	})
	req = httptest.NewRequest(http.MethodPost, "/site1/siteverify", bytes.NewReader(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("siteverify expected 200 got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMaxBodyBytes(t *testing.T) {
	h, _, _ := newTestHandler(t)

	large := bytes.Repeat([]byte("a"), 2048)
	req := httptest.NewRequest(http.MethodPost, "/site1/redeem", bytes.NewReader(large))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized body, got %d", rec.Code)
	}
}

func newTestHandler(t *testing.T) (http.Handler, *core.Service, []byte) {
	t.Helper()
	st := memstore.New(memstore.Options{GCInterval: time.Hour})
	pepper := []byte("pepper")
	err := st.UpsertSite(store.Site{
		SiteKey:        "site1",
		SecretHash:     core.HashSecret("secret1", pepper),
		JWTSecret:      []byte("jwt-secret"),
		Difficulty:     1,
		ChallengeCount: 1,
		SaltSize:       8,
	})
	if err != nil {
		t.Fatal(err)
	}
	svc := core.NewService(st, core.ServiceOptions{
		ChallengeTTL: 5 * time.Minute,
		RedeemTTL:    10 * time.Minute,
		Now:          time.Now,
		SecretPepper: pepper,
	})
	h := NewHandler(svc, Options{
		RateLimitMax:    1,
		RateLimitWindow: time.Hour,
		RateLimitScope:  "cap",
		MaxBodyBytes:    1024,
	})
	return h, svc, pepper
}

func requestChallenge(t *testing.T, h http.Handler, path string) (string, core.ChallengeParams) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("challenge failed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token     string               `json:"token"`
		Challenge core.ChallengeParams `json:"challenge"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal challenge: %v", err)
	}
	return resp.Token, resp.Challenge
}

func bruteOneLocal(seed string, c, s, d int) int {
	pairs := core.BuildChallengePairs(seed, c, s, d)
	return brutePowNonce(pairs[0][0], pairs[0][1])
}

func brutePowNonce(salt, target string) int {
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
