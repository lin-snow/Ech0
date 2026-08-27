// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package captcha

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/pkg/gocap/cap"
)

const defaultSiteKey = "ech0-comment"

func SiteKey() string {
	siteKey := strings.TrimSpace(config.Config().Comment.CaptchaSiteKey)
	if siteKey == "" {
		return defaultSiteKey
	}
	return siteKey
}

func Secret() string {
	secret := strings.TrimSpace(config.Config().Comment.CaptchaSecret)
	if secret != "" {
		return secret
	}
	sum := sha256.Sum256(config.Config().Security.JWTSecret)
	return hex.EncodeToString(sum[:])
}

func APIEndpoint() string {
	return "/api/cap/" + SiteKey() + "/"
}

func APIEndpointWithBase(baseURL string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return APIEndpoint()
	}
	return strings.TrimRight(base, "/") + APIEndpoint()
}

var sharedEngine = sync.OnceValues(NewEngine)

func SiteVerify(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("captcha token missing")
	}
	engine, err := sharedEngine()
	if err != nil {
		return err
	}
	ok, err := engine.SiteVerify(SiteKey(), Secret(), token)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("captcha verify failed")
	}
	return nil
}

func NewEngine() (*cap.Engine, error) {
	cfg := config.Config().Comment
	opts := []cap.Option{
		cap.WithInMemoryStore(),
		cap.WithEnableCORS(cfg.CaptchaEnableCORS),
		cap.WithRateLimitOnRedeem(cfg.CaptchaLimitOnRedeem),
		cap.WithRateLimitOnSiteVerify(cfg.CaptchaLimitOnVerify),
	}
	if cfg.CaptchaChallengeTTL > 0 {
		opts = append(opts, cap.WithChallengeTTL(time.Duration(cfg.CaptchaChallengeTTL)*time.Second))
	}
	if cfg.CaptchaRedeemTTL > 0 {
		opts = append(opts, cap.WithRedeemTTL(time.Duration(cfg.CaptchaRedeemTTL)*time.Second))
	}
	if cfg.CaptchaGCInterval > 0 {
		opts = append(opts, cap.WithGCInterval(time.Duration(cfg.CaptchaGCInterval)*time.Second))
	}
	if cfg.CaptchaRateLimitMax > 0 && cfg.CaptchaRateLimitWin > 0 {
		opts = append(opts, cap.WithRateLimit(cfg.CaptchaRateLimitMax, time.Duration(cfg.CaptchaRateLimitWin)*time.Second))
	}
	if cfg.CaptchaRateLimitScope != "" {
		opts = append(opts, cap.WithRateLimitScope(cfg.CaptchaRateLimitScope))
	}
	if cfg.CaptchaIPHeader != "" {
		opts = append(opts, cap.WithIPHeader(cfg.CaptchaIPHeader))
	}
	if cfg.CaptchaMaxBodyBytes > 0 {
		opts = append(opts, cap.WithMaxBodyBytes(int64(cfg.CaptchaMaxBodyBytes)))
	}

	engine, err := cap.New(opts...)
	if err != nil {
		return nil, err
	}

	if err := engine.RegisterSite(cap.SiteRegistration{
		SiteKey:        SiteKey(),
		Secret:         Secret(),
		Difficulty:     cfg.CaptchaDifficulty,
		ChallengeCount: cfg.CaptchaChallengeCount,
		SaltSize:       cfg.CaptchaSaltSize,
	}); err != nil {
		_ = engine.Close()
		return nil, err
	}
	return engine, nil
}

func NewHTTPHandler(stripPrefix string) (http.Handler, error) {
	engine, err := sharedEngine()
	if err != nil {
		return nil, err
	}
	handler := engine.Handler()
	prefix := strings.TrimSpace(stripPrefix)
	if prefix == "" {
		return handler, nil
	}
	return http.StripPrefix(prefix, handler), nil
}
