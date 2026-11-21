// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/ratelimit"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func perform(t *testing.T, r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func assertRateLimited(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header is missing")
	}

	var body struct {
		ErrorCode  string `json:"error_code"`
		MessageKey string `json:"message_key"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.ErrorCode != commonModel.ErrCodeRateLimited {
		t.Fatalf("error_code = %q, want %q", body.ErrorCode, commonModel.ErrCodeRateLimited)
	}
	if body.MessageKey != commonModel.MsgKeyAuthRateLimited {
		t.Fatalf("message_key = %q, want %q", body.MessageKey, commonModel.MsgKeyAuthRateLimited)
	}
}

func TestRateLimitBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(2, 2))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := range 2 {
		if rec := perform(t, r, http.MethodGet, "/test"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
	assertRateLimited(t, perform(t, r, http.MethodGet, "/test"))
}

func TestBudgetSharesAnonymousAddressBudget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := ratelimit.NewMemoryLimiter(ratelimit.Policy{
		ratelimit.ScopeAnonymous: {Limit: 2, Window: time.Minute},
	})

	r := gin.New()
	r.Use(ClientIP(), Budget(limiter))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := range 2 {
		if rec := perform(t, r, http.MethodGet, "/test"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
	assertRateLimited(t, perform(t, r, http.MethodGet, "/test"))
}

func TestBudgetTracksAuthenticatedUsersSeparately(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := ratelimit.NewMemoryLimiter(ratelimit.Policy{
		ratelimit.ScopeAnonymous:     {Limit: 1, Window: time.Minute},
		ratelimit.ScopeAuthenticated: {Limit: 2, Window: time.Minute},
	})

	r := gin.New()
	r.Use(ClientIP())
	r.Use(func(c *gin.Context) {
		user := c.GetHeader("X-Test-User")
		if user != "" {
			ctx := viewer.WithContext(c.Request.Context(), viewer.NewUserViewer(user))
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	})
	r.Use(Budget(limiter))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Anonymous traffic is limited to one request per address.
	if rec := perform(t, r, http.MethodGet, "/test"); rec.Code != http.StatusOK {
		t.Fatalf("anonymous first request: status = %d", rec.Code)
	}
	assertRateLimited(t, perform(t, r, http.MethodGet, "/test"))

	// Signed-in users draw from their own per-user budget.
	for i := range 2 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Test-User", "u-1")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("authenticated request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}

	// A different user is unaffected by the first one's exhaustion.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Test-User", "u-2")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("second user: status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestBudgetIgnoresForwardedForFromPublicPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := ratelimit.NewMemoryLimiter(ratelimit.Policy{
		ratelimit.ScopeAnonymous: {Limit: 1, Window: time.Minute},
	})

	r := gin.New()
	r.Use(ClientIP(), Budget(limiter))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	first := httptest.NewRequest(http.MethodGet, "/test", nil)
	first.RemoteAddr = "203.0.113.10:1234"
	first.Header.Set("X-Forwarded-For", "198.51.100.1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, first)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request: status = %d", rec.Code)
	}

	second := httptest.NewRequest(http.MethodGet, "/test", nil)
	second.RemoteAddr = "203.0.113.10:5678"
	second.Header.Set("X-Forwarded-For", "198.51.100.2")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, second)
	assertRateLimited(t, rec)
}

func TestScopeLimitChargesCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := ratelimit.NewMemoryLimiter(ratelimit.Policy{
		ratelimit.ScopeSignUpIP: {Limit: 1, Window: time.Hour},
	})

	r := gin.New()
	r.Use(ClientIP(), ScopeLimit(limiter, ratelimit.ScopeSignUpIP, 1))
	r.POST("/register", func(c *gin.Context) { c.Status(http.StatusCreated) })

	if rec := perform(t, r, http.MethodPost, "/register"); rec.Code != http.StatusCreated {
		t.Fatalf("first registration: status = %d", rec.Code)
	}
	assertRateLimited(t, perform(t, r, http.MethodPost, "/register"))
}

func TestNilLimiterDisablesBudget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ClientIP(), Budget(nil), ScopeLimit(nil, ratelimit.ScopeSignUpIP, 1))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	for range 5 {
		if rec := perform(t, r, http.MethodGet, "/test"); rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	}
}
