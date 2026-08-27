// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cap

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"

	"github.com/lin-snow/ech0/pkg/gocap/core"
	"github.com/lin-snow/ech0/pkg/gocap/store"
	"github.com/lin-snow/ech0/pkg/gocap/store/memstore"
	caphttp "github.com/lin-snow/ech0/pkg/gocap/transport/http"
)

type Engine struct {
	store   store.Store
	service *core.Service
	handler http.Handler
	cfg     config
}

func New(opts ...Option) (*Engine, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	st := cfg.customStore
	if st == nil {
		st = memstore.New(memstore.Options{GCInterval: cfg.gcInterval})
	}

	service := core.NewService(st, core.ServiceOptions{
		ChallengeTTL: cfg.challengeTTL,
		RedeemTTL:    cfg.redeemTTL,
		RNG:          rand.Reader,
		SecretPepper: cfg.secretPepper,
	})

	handler := caphttp.NewHandler(service, caphttp.Options{
		RateLimitMax:    cfg.rateLimit.Max,
		RateLimitWindow: cfg.rateLimit.Window,
		RateLimitScope:  cfg.rateLimit.Scope,
		RateLimitRedeem: cfg.rateLimitOnRedeem,
		RateLimitVerify: cfg.rateLimitOnVerify,
		EnableCORS:      cfg.enableCORS,
		IPHeader:        cfg.ipHeader,
		MaxBodyBytes:    cfg.maxBodyBytes,
	})

	return &Engine{
		store:   st,
		service: service,
		handler: handler,
		cfg:     cfg,
	}, nil
}

func (e *Engine) Handler() http.Handler {
	return e.handler
}

func (e *Engine) SiteVerify(siteKey, secret, response string) (bool, error) {
	resp, err := e.service.SiteVerify(siteKey, core.SiteVerifyRequest{
		Secret:   secret,
		Response: response,
	})
	if err != nil {
		var domainErr *core.Error
		if errors.As(err, &domainErr) && domainErr.Code != core.ErrCodeInternal {
			return false, nil
		}
		return false, err
	}
	return resp.Success, nil
}

func (e *Engine) RegisterSite(site SiteRegistration) error {
	if site.SiteKey == "" {
		return fmt.Errorf("site key is required")
	}
	if site.Secret == "" {
		return fmt.Errorf("secret is required")
	}

	jwtSecret := make([]byte, 32)
	if _, err := rand.Read(jwtSecret); err != nil {
		return fmt.Errorf("generate jwt secret: %w", err)
	}

	challengeCount := site.ChallengeCount
	if challengeCount <= 0 {
		challengeCount = 80
	}
	if challengeCount > 500 {
		return fmt.Errorf("challenge count out of range")
	}

	difficulty := site.Difficulty
	if difficulty <= 0 {
		difficulty = 4
	}
	if difficulty > 8 {
		return fmt.Errorf("difficulty out of range")
	}

	saltSize := site.SaltSize
	if saltSize <= 0 {
		saltSize = 32
	}

	secretHash := core.HashSecret(site.Secret, e.cfg.secretPepper)
	return e.store.UpsertSite(store.Site{
		SiteKey:          site.SiteKey,
		SecretHash:       secretHash,
		JWTSecret:        jwtSecret,
		Difficulty:       difficulty,
		ChallengeCount:   challengeCount,
		SaltSize:         saltSize,
		BlockOnRateLimit: true,
	})
}

func (e *Engine) RemoveSite(siteKey string) error {
	return e.store.DeleteSite(siteKey)
}

func (e *Engine) Close() error {
	return e.store.Close()
}
