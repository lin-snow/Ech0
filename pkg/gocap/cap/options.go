// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cap

import (
	"crypto/rand"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/store"
)

type Option func(*config)

type config struct {
	challengeTTL      time.Duration
	redeemTTL         time.Duration
	gcInterval        time.Duration
	secretPepper      []byte
	customStore       store.Store
	rateLimit         RateLimitConfig
	rateLimitOnRedeem bool
	rateLimitOnVerify bool
	enableCORS        bool
	ipHeader          string
	maxBodyBytes      int64
}

func defaultConfig() config {
	pepper := make([]byte, 32)
	_, _ = rand.Read(pepper)
	return config{
		challengeTTL: 15 * time.Minute,
		redeemTTL:    2 * time.Hour,
		gcInterval:   2 * time.Second,
		secretPepper: pepper,
		rateLimit: RateLimitConfig{
			Max:    30,
			Window: 5 * time.Second,
			Scope:  "cap",
		},
		rateLimitOnRedeem: false,
		rateLimitOnVerify: false,
		maxBodyBytes:      1 << 20,
	}
}

func WithChallengeTTL(ttl time.Duration) Option {
	return func(c *config) {
		c.challengeTTL = ttl
	}
}

func WithRedeemTTL(ttl time.Duration) Option {
	return func(c *config) {
		c.redeemTTL = ttl
	}
}

func WithGCInterval(interval time.Duration) Option {
	return func(c *config) {
		c.gcInterval = interval
	}
}

func WithSecretPepper(pepper []byte) Option {
	return func(c *config) {
		if len(pepper) > 0 {
			c.secretPepper = append([]byte(nil), pepper...)
		}
	}
}

func WithStore(st store.Store) Option {
	return func(c *config) {
		c.customStore = st
	}
}

func WithInMemoryStore() Option {
	return func(c *config) {
		c.customStore = nil
	}
}

func WithRateLimit(max int, window time.Duration) Option {
	return func(c *config) {
		c.rateLimit.Max = max
		c.rateLimit.Window = window
	}
}

func WithRateLimitScope(scope string) Option {
	return func(c *config) {
		c.rateLimit.Scope = scope
	}
}

func WithEnableCORS(enabled bool) Option {
	return func(c *config) {
		c.enableCORS = enabled
	}
}

func WithIPHeader(header string) Option {
	return func(c *config) {
		c.ipHeader = header
	}
}

func WithRateLimitOnRedeem(enabled bool) Option {
	return func(c *config) {
		c.rateLimitOnRedeem = enabled
	}
}

func WithRateLimitOnSiteVerify(enabled bool) Option {
	return func(c *config) {
		c.rateLimitOnVerify = enabled
	}
}

func WithMaxBodyBytes(n int64) Option {
	return func(c *config) {
		if n > 0 {
			c.maxBodyBytes = n
		}
	}
}
