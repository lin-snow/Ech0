// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type clientConfig struct {
	timeout time.Duration
	guard   bool
}

type Option func(*clientConfig)

func Timeout(d time.Duration) Option {
	return func(c *clientConfig) { c.timeout = d }
}

func Guard() Option {
	return func(c *clientConfig) { c.guard = true }
}

func NewClient(opts ...Option) *http.Client {
	cfg := clientConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	client := &http.Client{
		Timeout:   cfg.timeout,
		Transport: &loggingRoundTripper{base: transport},
	}

	if cfg.guard {
		transport.DialContext = secureDialContext(cfg.timeout)
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxSafeRedirects {
				return errors.New("too many redirects")
			}
			return Validate(req.URL.String())
		}
	}

	return client
}

type loggingRoundTripper struct {
	base http.RoundTripper
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := l.base.RoundTrip(req)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		logUtil.Debug(
			"egress request failed",
			slog.String("module", "egress"),
			slog.String("method", req.Method),
			slog.String("host", req.URL.Host),
			slog.Int64("latency_ms", latencyMs),
			logUtil.Err(err),
		)
		return resp, err
	}
	logUtil.Debug(
		"egress request",
		slog.String("module", "egress"),
		slog.String("method", req.Method),
		slog.String("host", req.URL.Host),
		slog.Int("status", resp.StatusCode),
		slog.Int64("latency_ms", latencyMs),
	)
	return resp, err
}

func Retry(maxAttempts int, initialBackoff time.Duration, fn func() error) error {
	var err error
	delay := initialBackoff
	for i := range maxAttempts {
		if err = fn(); err == nil {
			return nil
		}
		if i < maxAttempts-1 {
			time.Sleep(delay)
			delay *= 2
		}
	}
	return err
}

type Header struct {
	Header  string
	Content string
}

func Fetch(url, method string, h Header, timeout ...time.Duration) ([]byte, error) {
	if err := Validate(url); err != nil {
		return nil, err
	}

	clientTimeout := 2 * time.Second
	if len(timeout) > 0 {
		clientTimeout = timeout[0]
	}

	client := NewClient(Guard(), Timeout(clientTimeout))

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	if h.Header != "" && h.Content != "" {
		req.Header.Set(h.Header, h.Content)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logUtil.Warn(
				"close response body failed",
				slog.String("module", "egress"),
				logUtil.Err(closeErr),
			)
		}
	}()

	return readBodyWithLimit(resp.Body, defaultSafeResponseBodyLimitBytes)
}
