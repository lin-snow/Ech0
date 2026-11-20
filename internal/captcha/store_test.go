// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package captcha

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/securitystate"
	"github.com/lin-snow/ech0/pkg/gocap/store"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIndependentCaptchaStores(t *testing.T) {
	path := filepath.Join(t.TempDir(), "captcha.db")
	open := func() *sharedStore {
		db, err := gorm.Open(sqlite.Open(path))
		require.NoError(t, err)
		pool, err := db.DB()
		require.NoError(t, err)
		t.Cleanup(func() { _ = pool.Close() })
		require.NoError(t, db.AutoMigrate(&securitystate.Entry{}))
		return &sharedStore{db: func() *gorm.DB { return db }, secret: []byte("test-secret")}
	}
	first, second := open(), open()
	for _, s := range []*sharedStore{first, second} {
		require.NoError(t, s.UpsertSite(store.Site{SiteKey: "site", SecretHash: []byte("hash")}))
	}
	a, ok := first.GetSite("site")
	require.True(t, ok)
	b, ok := second.GetSite("site")
	require.True(t, ok)
	require.Equal(t, a.JWTSecret, b.JWTSecret)
	now := time.Now().Truncate(time.Millisecond)
	ends := now.Add(time.Minute)
	accepted, err := first.TryMarkChallengeSigUsed("signature", ends, now)
	require.NoError(t, err)
	require.True(t, accepted)
	accepted, err = second.TryMarkChallengeSigUsed("signature", ends, now)
	require.NoError(t, err)
	require.False(t, accepted)
	require.NoError(t, first.StoreRedeemToken("site", "token", ends))
	found, expired, err := second.ConsumeRedeemToken("site", "token", now)
	require.NoError(t, err)
	require.True(t, found)
	require.False(t, expired)
	found, _, err = first.ConsumeRedeemToken("site", "token", now)
	require.NoError(t, err)
	require.False(t, found)
	accepted, _, err = first.AllowRateLimit("captcha", "ip", 1, time.Minute, now)
	require.NoError(t, err)
	require.True(t, accepted)
	accepted, _, err = second.AllowRateLimit("captcha", "ip", 1, time.Minute, now)
	require.NoError(t, err)
	require.False(t, accepted)
}
