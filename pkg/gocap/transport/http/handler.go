// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package caphttp

import (
	"net/http"
	"strings"
	"time"

	"github.com/lin-snow/ech0/pkg/gocap/core"
)

type Options struct {
	RateLimitMax    int
	RateLimitWindow time.Duration
	RateLimitScope  string
	RateLimitRedeem bool
	RateLimitVerify bool
	IPHeader        string
	EnableCORS      bool
	MaxBodyBytes    int64
}

type Handler struct {
	service *core.Service
	opts    Options
}

func NewHandler(service *core.Service, opts Options) http.Handler {
	if opts.RateLimitScope == "" {
		opts.RateLimitScope = "cap"
	}
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = 1 << 20
	}
	h := &Handler{
		service: service,
		opts:    opts,
	}
	return recoverMiddleware(h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.opts.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST, OPTIONS")
		writeJSON(w, http.StatusMethodNotAllowed, errorBody{
			Success: false,
			Code:    "method_not_allowed",
			Error:   "Method not allowed",
		})
		return
	}

	siteKey, action, ok := parsePath(r.URL.Path)
	if !ok {
		writeCoreError(w, core.NewNotFound("Not found"))
		return
	}

	if h.shouldRateLimit(action) && h.opts.RateLimitMax > 0 && h.opts.RateLimitWindow > 0 {
		ip := getClientIP(r, h.opts.IPHeader)
		allowed, _, err := h.service.Store.AllowRateLimit(
			h.opts.RateLimitScope,
			siteKey+":"+ip,
			h.opts.RateLimitMax,
			h.opts.RateLimitWindow,
			h.service.Now(),
		)
		if err != nil {
			writeCoreError(w, core.NewInternal("Rate limiter failure"))
			return
		}
		if !allowed {
			writeCoreError(w, core.NewRateLimit("Rate limit exceeded"))
			return
		}
	}

	switch action {
	case "challenge":
		resp, err := h.service.CreateChallenge(siteKey)
		if err != nil {
			writeCoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	case "redeem":
		var req core.RedeemRequest
		r.Body = http.MaxBytesReader(w, r.Body, h.opts.MaxBodyBytes)
		if err := decodeJSON(r, &req, decodeOptions{Strict: false}); err != nil {
			writeDecodeError(w, err)
			return
		}
		resp, err := h.service.Redeem(siteKey, req)
		if err != nil {
			writeCoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	case "siteverify":
		var req core.SiteVerifyRequest
		r.Body = http.MaxBytesReader(w, r.Body, h.opts.MaxBodyBytes)
		if err := decodeJSON(r, &req, decodeOptions{Strict: true}); err != nil {
			writeDecodeError(w, err)
			return
		}
		resp, err := h.service.SiteVerify(siteKey, req)
		if err != nil {
			writeCoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	default:
		writeCoreError(w, core.NewNotFound("Not found"))
	}
}

func parsePath(path string) (siteKey, action string, ok bool) {
	p := strings.Trim(path, "/")
	parts := strings.Split(p, "/")
	switch len(parts) {
	case 2:
		siteKey, action = parts[0], parts[1]
	case 3:
		if parts[0] != "cap" {
			return "", "", false
		}
		siteKey, action = parts[1], parts[2]
	default:
		return "", "", false
	}
	if siteKey == "" || action == "" {
		return "", "", false
	}
	switch action {
	case "challenge", "redeem", "siteverify":
		return siteKey, action, true
	default:
		return "", "", false
	}
}

func (h *Handler) shouldRateLimit(action string) bool {
	switch action {
	case "challenge":
		return true
	case "redeem":
		return h.opts.RateLimitRedeem
	case "siteverify":
		return h.opts.RateLimitVerify
	default:
		return false
	}
}
