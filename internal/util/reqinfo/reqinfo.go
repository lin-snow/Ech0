// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package reqinfo derives the externally visible scheme, host and origin of an
// HTTP request. Reverse proxies are expected to forward X-Forwarded-Proto and
// X-Forwarded-Host; both may carry a comma-separated chain, in which case the
// first value is the client-facing one. Every origin consumer (discovery
// metadata, OAuth issuer, MCP challenges, same-origin checks, sitemap) reads
// the request through this package so those views cannot drift.
package reqinfo

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	forwardedProtoHeader = "X-Forwarded-Proto"
	forwardedHostHeader  = "X-Forwarded-Host"
)

// Scheme returns the scheme a client used. TLS termination wins over a
// forwarded header; unknown forwarded values are ignored rather than echoed
// into URLs and headers.
func Scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	switch strings.ToLower(firstValue(r.Header.Get(forwardedProtoHeader))) {
	case "https":
		return "https"
	case "http":
		return "http"
	default:
		return "http"
	}
}

// Host returns the client-facing host: X-Forwarded-Host when a proxy sets it,
// otherwise the Host header.
func Host(r *http.Request) string {
	if host := firstValue(r.Header.Get(forwardedHostHeader)); host != "" {
		return host
	}
	return strings.TrimSpace(r.Host)
}

// Origin returns scheme://host without a trailing slash.
func Origin(r *http.Request) string {
	return Scheme(r) + "://" + Host(r)
}

// IsSameOrigin reports whether an Origin header value names the same origin as
// the request, comparing the client-facing host and scheme.
func IsSameOrigin(origin string, r *http.Request) bool {
	trimmed := strings.TrimSuffix(strings.TrimSpace(origin), "/")
	if trimmed == "" {
		return false
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Scheme, Scheme(r)) && strings.EqualFold(parsed.Host, Host(r))
}

func firstValue(raw string) string {
	return strings.TrimSpace(strings.Split(raw, ",")[0])
}
