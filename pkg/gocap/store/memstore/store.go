// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package memstore

import (
	"fmt"
	"sync"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/store"
)

type redeemEntry struct {
	expiresAt time.Time
}

type rateWindow struct {
	count     int
	expiresAt time.Time
}

type Store struct {
	mu sync.RWMutex

	sites map[string]store.Site

	usedChallengeSig map[string]time.Time
	redeemTokens     map[string]redeemEntry
	rateWindows      map[string]rateWindow

	stopCh chan struct{}
	once   sync.Once
}

type Options struct {
	GCInterval time.Duration
}

func New(opts Options) *Store {
	s := &Store{
		sites:            make(map[string]store.Site),
		usedChallengeSig: make(map[string]time.Time),
		redeemTokens:     make(map[string]redeemEntry),
		rateWindows:      make(map[string]rateWindow),
		stopCh:           make(chan struct{}),
	}

	interval := opts.GCInterval
	if interval <= 0 {
		interval = 2 * time.Second
	}
	s.startGC(interval)

	return s
}

func (s *Store) UpsertSite(siteCfg store.Site) error {
	if siteCfg.SiteKey == "" {
		return fmt.Errorf("site key is required")
	}
	if len(siteCfg.SecretHash) == 0 {
		return fmt.Errorf("secret hash is required")
	}
	if len(siteCfg.JWTSecret) == 0 {
		return fmt.Errorf("jwt secret is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	siteCfg.SecretHash = append([]byte(nil), siteCfg.SecretHash...)
	siteCfg.JWTSecret = append([]byte(nil), siteCfg.JWTSecret...)
	s.sites[siteCfg.SiteKey] = siteCfg
	return nil
}

func (s *Store) GetSite(siteKey string) (store.Site, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site, ok := s.sites[siteKey]
	if ok {
		site.SecretHash = append([]byte(nil), site.SecretHash...)
		site.JWTSecret = append([]byte(nil), site.JWTSecret...)
	}
	return site, ok
}

func (s *Store) DeleteSite(siteKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sites, siteKey)
	return nil
}

func (s *Store) TryMarkChallengeSigUsed(sig string, expiresAt time.Time, now time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if exp, ok := s.usedChallengeSig[sig]; ok {
		if exp.After(now) {
			return false, nil
		}
		delete(s.usedChallengeSig, sig)
	}
	s.usedChallengeSig[sig] = expiresAt
	return true, nil
}

func (s *Store) StoreRedeemToken(siteKey, token string, expiresAt time.Time) error {
	k := redeemKey(siteKey, token)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.redeemTokens[k] = redeemEntry{expiresAt: expiresAt}
	return nil
}

func (s *Store) ConsumeRedeemToken(siteKey, token string, now time.Time) (bool, bool, error) {
	k := redeemKey(siteKey, token)
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.redeemTokens[k]
	if !ok {
		return false, false, nil
	}
	delete(s.redeemTokens, k)
	if !entry.expiresAt.After(now) {
		return true, true, nil
	}
	return true, false, nil
}

func (s *Store) AllowRateLimit(scope, key string, max int, window time.Duration, now time.Time) (bool, int, error) {
	if max <= 0 || window <= 0 {
		return true, max, nil
	}

	windowMs := int64(window / time.Millisecond)
	if windowMs <= 0 {
		return true, max, nil
	}
	windowBucket := now.UnixMilli() / windowMs
	rlKey := rateKey(scope, key, windowMs, windowBucket)

	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.rateWindows[rlKey]
	if !ok {
		val = rateWindow{
			count:     0,
			expiresAt: time.UnixMilli((windowBucket+1)*windowMs + 1),
		}
	}
	val.count++
	s.rateWindows[rlKey] = val

	remaining := max - val.count
	if remaining < 0 {
		remaining = 0
	}
	return val.count <= max, remaining, nil
}

func (s *Store) Close() error {
	s.once.Do(func() {
		close(s.stopCh)
	})
	return nil
}

func redeemKey(siteKey, token string) string {
	return siteKey + ":" + token
}

func rateKey(scope, key string, windowMs, bucket int64) string {
	return fmt.Sprintf("%s:%s:%d:%d", scope, key, windowMs, bucket)
}
