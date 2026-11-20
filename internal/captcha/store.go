// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package captcha

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/securitystate"
	"github.com/lin-snow/ech0/pkg/gocap/store"
	"gorm.io/gorm"
)

// Site configuration is rebuilt from environment on every cold start.
// Signing keys and all mutable protocol state must agree across instances.
type sharedStore struct {
	db     func() *gorm.DB
	secret []byte
	sites  sync.Map
}

func (s *sharedStore) UpsertSite(site store.Site) error {
	if site.SiteKey == "" || len(site.SecretHash) == 0 {
		return errors.New("invalid captcha site")
	}
	site.JWTSecret = securitystate.Digest(s.secret, "captcha-sign", site.SiteKey)
	s.sites.Store(site.SiteKey, site)
	return nil
}
func (s *sharedStore) GetSite(key string) (store.Site, bool) {
	value, ok := s.sites.Load(key)
	if !ok {
		return store.Site{}, false
	}
	return value.(store.Site), true
}
func (s *sharedStore) DeleteSite(key string) error { s.sites.Delete(key); return nil }
func (s *sharedStore) Close() error                { return nil }

func (s *sharedStore) connection() (*gorm.DB, context.CancelFunc) {
	// The upstream Store interface has no request context.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return s.db().WithContext(ctx), cancel
}
func (s *sharedStore) TryMarkChallengeSigUsed(sig string, expiry, now time.Time) (bool, error) {
	db, cancel := s.connection()
	defer cancel()
	return securitystate.Acquire(db, securitystate.Key(s.secret, "captcha-challenge", sig), expiry, now)
}
func (s *sharedStore) StoreRedeemToken(site, token string, expiry time.Time) error {
	db, cancel := s.connection()
	defer cancel()
	_, err := securitystate.Acquire(db, securitystate.Key(s.secret, "captcha-redeem", site, token), expiry, time.Now())
	return err
}
func (s *sharedStore) ConsumeRedeemToken(site, token string, now time.Time) (bool, bool, error) {
	db, cancel := s.connection()
	defer cancel()
	expiry, found, err := securitystate.Consume(db, securitystate.Key(s.secret, "captcha-redeem", site, token))
	return found, found && !expiry.After(now), err
}
func (s *sharedStore) AllowRateLimit(scope, key string, limit int, window time.Duration, now time.Time) (bool, int, error) {
	ms := window.Milliseconds()
	if limit <= 0 || ms <= 0 {
		return true, limit, nil
	}
	bucket := now.UnixMilli() / ms
	id := securitystate.Key(s.secret, "captcha-rate", scope, key, strconv.FormatInt(ms, 10), strconv.FormatInt(bucket, 10))
	db, cancel := s.connection()
	defer cancel()
	count, err := securitystate.Increment(db, id, time.UnixMilli((bucket+1)*ms+1))
	return err == nil && count <= int64(limit), max(limit-int(count), 0), err
}
