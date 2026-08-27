// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package store

import "time"

type Site struct {
	SiteKey          string
	SecretHash       []byte
	JWTSecret        []byte
	Difficulty       int
	ChallengeCount   int
	SaltSize         int
	BlockOnRateLimit bool
}

type Store interface {
	UpsertSite(site Site) error
	GetSite(siteKey string) (Site, bool)
	DeleteSite(siteKey string) error

	TryMarkChallengeSigUsed(sig string, expiresAt time.Time, now time.Time) (bool, error)

	StoreRedeemToken(siteKey, token string, expiresAt time.Time) error

	ConsumeRedeemToken(siteKey, token string, now time.Time) (found bool, expired bool, err error)

	AllowRateLimit(scope, key string, max int, window time.Duration, now time.Time) (bool, int, error)

	Close() error
}
