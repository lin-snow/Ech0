// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cookie

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func findSetCookie(t *testing.T, w *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, ck := range w.Result().Cookies() {
		if ck.Name == name {
			return ck
		}
	}
	t.Fatalf("Set-Cookie %q not found in response", name)
	return nil
}

func TestIsHTTPS(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(c *gin.Context)
		expect bool
	}{
		{
			name:   "plain HTTP without any signal",
			setup:  func(c *gin.Context) {},
			expect: false,
		},
		{
			name: "direct TLS connection",
			setup: func(c *gin.Context) {
				c.Request.TLS = &tls.ConnectionState{}
			},
			expect: true,
		},
		{
			name: "X-Forwarded-Proto https (case-insensitive)",
			setup: func(c *gin.Context) {
				c.Request.Header.Set("X-Forwarded-Proto", "HTTPS")
			},
			expect: true,
		},
		{
			name: "X-Forwarded-Proto http",
			setup: func(c *gin.Context) {
				c.Request.Header.Set("X-Forwarded-Proto", "http")
			},
			expect: false,
		},
		{
			name: "Origin https prefix",
			setup: func(c *gin.Context) {
				c.Request.Header.Set("Origin", "https://example.com")
			},
			expect: true,
		},
		{
			name: "Origin http prefix",
			setup: func(c *gin.Context) {
				c.Request.Header.Set("Origin", "http://example.com")
			},
			expect: false,
		},
		{
			name: "Referer https prefix",
			setup: func(c *gin.Context) {
				c.Request.Header.Set("Referer", "https://example.com/page")
			},
			expect: true,
		},
		{
			name: "Referer http prefix",
			setup: func(c *gin.Context) {
				c.Request.Header.Set("Referer", "http://example.com/page")
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestContext(t)
			tt.setup(c)
			if got := isHTTPS(c); got != tt.expect {
				t.Fatalf("isHTTPS() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestSetRefreshTokenCookie(t *testing.T) {
	t.Run("attributes on plain HTTP are not Secure", func(t *testing.T) {
		c, w := newTestContext(t)
		SetRefreshTokenCookie(c, "tok-123", 3600)

		ck := findSetCookie(t, w, RefreshTokenCookieName)
		if ck.Value != "tok-123" {
			t.Errorf("Value = %q, want %q", ck.Value, "tok-123")
		}
		if ck.Path != "/api/auth" {
			t.Errorf("Path = %q, want %q", ck.Path, "/api/auth")
		}
		if ck.MaxAge != 3600 {
			t.Errorf("MaxAge = %d, want %d", ck.MaxAge, 3600)
		}
		if !ck.HttpOnly {
			t.Errorf("HttpOnly = false, want true")
		}
		if ck.SameSite != http.SameSiteLaxMode {
			t.Errorf("SameSite = %v, want Lax(%v)", ck.SameSite, http.SameSiteLaxMode)
		}
		if ck.Secure {
			t.Errorf("Secure = true on plain HTTP, want false")
		}
	})

	t.Run("Secure set when request is HTTPS", func(t *testing.T) {
		c, w := newTestContext(t)
		c.Request.TLS = &tls.ConnectionState{}
		SetRefreshTokenCookie(c, "tok-secure", 7200)

		ck := findSetCookie(t, w, RefreshTokenCookieName)
		if !ck.Secure {
			t.Errorf("Secure = false over TLS, want true")
		}
		if ck.MaxAge != 7200 {
			t.Errorf("MaxAge = %d, want %d", ck.MaxAge, 7200)
		}
	})
}

func TestClearRefreshTokenCookie(t *testing.T) {
	c, w := newTestContext(t)
	ClearRefreshTokenCookie(c)

	ck := findSetCookie(t, w, RefreshTokenCookieName)
	if ck.Value != "" {
		t.Errorf("Value = %q, want empty", ck.Value)
	}
	if ck.Path != "/api/auth" {
		t.Errorf("Path = %q, want %q", ck.Path, "/api/auth")
	}
	if ck.MaxAge >= 0 {
		t.Errorf("MaxAge = %d, want negative (expire immediately)", ck.MaxAge)
	}
	if !ck.HttpOnly {
		t.Errorf("HttpOnly = false, want true")
	}
	if ck.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax(%v)", ck.SameSite, http.SameSiteLaxMode)
	}
}

func TestGetRefreshTokenFromCookie(t *testing.T) {
	t.Run("returns token when cookie present", func(t *testing.T) {
		c, _ := newTestContext(t)
		c.Request.AddCookie(&http.Cookie{Name: RefreshTokenCookieName, Value: "read-me"})

		got, err := GetRefreshTokenFromCookie(c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "read-me" {
			t.Errorf("token = %q, want %q", got, "read-me")
		}
	})

	t.Run("returns error when cookie missing", func(t *testing.T) {
		c, _ := newTestContext(t)
		got, err := GetRefreshTokenFromCookie(c)
		if err == nil {
			t.Fatalf("expected error for missing cookie, got nil (value=%q)", got)
		}
		if got != "" {
			t.Errorf("token = %q, want empty on error", got)
		}
	})

	t.Run("round-trip: Set then read back the value", func(t *testing.T) {
		cWrite, w := newTestContext(t)
		SetRefreshTokenCookie(cWrite, "round-trip-tok", 3600)

		cRead, _ := newTestContext(t)
		for _, ck := range w.Result().Cookies() {
			cRead.Request.AddCookie(ck)
		}

		got, err := GetRefreshTokenFromCookie(cRead)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "round-trip-tok" {
			t.Errorf("token = %q, want %q", got, "round-trip-tok")
		}
	})
}
