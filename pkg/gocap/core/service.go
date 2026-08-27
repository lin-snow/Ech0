// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/store"
)

type Service struct {
	Store store.Store

	ChallengeTTL time.Duration
	RedeemTTL    time.Duration
	Now          func() time.Time
	RNG          io.Reader

	SecretPepper []byte
}

type ServiceOptions struct {
	ChallengeTTL time.Duration
	RedeemTTL    time.Duration
	Now          func() time.Time
	RNG          io.Reader
	SecretPepper []byte
}

func NewService(st store.Store, opts ServiceOptions) *Service {
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	rng := opts.RNG
	if rng == nil {
		rng = rand.Reader
	}
	challengeTTL := opts.ChallengeTTL
	if challengeTTL <= 0 {
		challengeTTL = 15 * time.Minute
	}
	redeemTTL := opts.RedeemTTL
	if redeemTTL <= 0 {
		redeemTTL = 2 * time.Hour
	}
	return &Service{
		Store:        st,
		ChallengeTTL: challengeTTL,
		RedeemTTL:    redeemTTL,
		Now:          nowFn,
		RNG:          rng,
		SecretPepper: opts.SecretPepper,
	}
}

func HashSecret(secret string, pepper []byte) []byte {
	h := hmac.New(sha256.New, pepper)
	_, _ = h.Write([]byte(secret))
	return h.Sum(nil)
}

func SecureSecretEqual(secret string, expectedHash, pepper []byte) bool {
	got := HashSecret(secret, pepper)
	return subtle.ConstantTimeCompare(got, expectedHash) == 1
}

func randomHex(r io.Reader, nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
