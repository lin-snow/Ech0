// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cap

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type challengeResp struct {
	Challenge struct {
		C int `json:"c"`
		S int `json:"s"`
		D int `json:"d"`
	} `json:"challenge"`
	Token string `json:"token"`
}

type redeemResp struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

func TestEngineHTTPFlow(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = engine.Close()
	}()

	if err := engine.RegisterSite(SiteRegistration{
		SiteKey:        "my-site",
		Secret:         "my-secret",
		Difficulty:     1,
		ChallengeCount: 1,
		SaltSize:       8,
	}); err != nil {
		t.Fatal(err)
	}

	h := engine.Handler()

	req := httptest.NewRequest(http.MethodPost, "/my-site/challenge", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("challenge status: %d, body=%s", rec.Code, rec.Body.String())
	}

	var ch challengeResp
	if err := json.Unmarshal(rec.Body.Bytes(), &ch); err != nil {
		t.Fatal(err)
	}
	sol := bruteSolution(ch.Token, ch.Challenge.C, ch.Challenge.S, ch.Challenge.D)
	redeemBody, _ := json.Marshal(map[string]any{
		"token":     ch.Token,
		"solutions": []int{sol},
	})
	req = httptest.NewRequest(http.MethodPost, "/my-site/redeem", bytes.NewReader(redeemBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("redeem status: %d, body=%s", rec.Code, rec.Body.String())
	}

	var rdm redeemResp
	if err := json.Unmarshal(rec.Body.Bytes(), &rdm); err != nil {
		t.Fatal(err)
	}
	if !rdm.Success || rdm.Token == "" {
		t.Fatalf("invalid redeem response: %s", rec.Body.String())
	}

	verifyBody, _ := json.Marshal(map[string]any{
		"secret":   "my-secret",
		"response": rdm.Token,
	})
	req = httptest.NewRequest(http.MethodPost, "/my-site/siteverify", bytes.NewReader(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("siteverify status: %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestEngineSiteVerifyInProcess(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = engine.Close()
	}()

	if err := engine.RegisterSite(SiteRegistration{
		SiteKey:        "my-site",
		Secret:         "my-secret",
		Difficulty:     1,
		ChallengeCount: 1,
		SaltSize:       8,
	}); err != nil {
		t.Fatal(err)
	}

	token := redeemViaHTTP(t, engine.Handler())
	ok, err := engine.SiteVerify("my-site", "my-secret", token)
	if err != nil {
		t.Fatalf("in-process siteverify error: %v", err)
	}
	if !ok {
		t.Fatal("expected redeem token to verify in-process")
	}

	ok, err = engine.SiteVerify("my-site", "my-secret", token)
	if err != nil {
		t.Fatalf("second siteverify error: %v", err)
	}
	if ok {
		t.Fatal("expected redeem token to be consumed after first verify")
	}

	fresh := redeemViaHTTP(t, engine.Handler())
	ok, err = engine.SiteVerify("my-site", "wrong-secret", fresh)
	if err != nil {
		t.Fatalf("wrong-secret siteverify error: %v", err)
	}
	if ok {
		t.Fatal("expected wrong secret to fail verification")
	}
}

func redeemViaHTTP(t *testing.T, h http.Handler) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/my-site/challenge", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("challenge status: %d, body=%s", rec.Code, rec.Body.String())
	}
	var ch challengeResp
	if err := json.Unmarshal(rec.Body.Bytes(), &ch); err != nil {
		t.Fatal(err)
	}
	sol := bruteSolution(ch.Token, ch.Challenge.C, ch.Challenge.S, ch.Challenge.D)
	redeemBody, _ := json.Marshal(map[string]any{
		"token":     ch.Token,
		"solutions": []int{sol},
	})
	req = httptest.NewRequest(http.MethodPost, "/my-site/redeem", bytes.NewReader(redeemBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("redeem status: %d, body=%s", rec.Code, rec.Body.String())
	}
	var rdm redeemResp
	if err := json.Unmarshal(rec.Body.Bytes(), &rdm); err != nil {
		t.Fatal(err)
	}
	if !rdm.Success || rdm.Token == "" {
		t.Fatalf("invalid redeem response: %s", rec.Body.String())
	}
	return rdm.Token
}

func bruteSolution(seed string, _ int, saltSize, difficulty int) int {
	salt, target := buildPair(seed, 1, saltSize, difficulty)
	for nonce := range 10_000_000 {
		sum := sha256.Sum256([]byte(salt + strconv.Itoa(nonce)))
		h := hex.EncodeToString(sum[:])
		if h[:len(target)] == target {
			return nonce
		}
	}
	return 0
}

func buildPair(seed string, idx, saltSize, difficulty int) (string, string) {
	salt := prng(seed+strconv.Itoa(idx), saltSize)
	target := prng(seed+strconv.Itoa(idx)+"d", difficulty)
	return salt, target
}

func prng(seed string, length int) string {
	hash := fnv1a(seed)
	out := ""
	for len(out) < length {
		hash ^= hash << 13
		hash ^= hash >> 17
		hash ^= hash << 5
		out += leftPadHex8(hash)
	}
	return out[:length]
}

func fnv1a(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h += (h << 1) + (h << 4) + (h << 7) + (h << 8) + (h << 24)
	}
	return h
}

func leftPadHex8(v uint32) string {
	const hexdigits = "0123456789abcdef"
	buf := [8]byte{}
	for i := 7; i >= 0; i-- {
		buf[i] = hexdigits[v&0xF]
		v >>= 4
	}
	return string(buf[:])
}
