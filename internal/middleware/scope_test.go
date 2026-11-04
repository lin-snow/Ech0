// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/authz"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func TestRequireScopes_ReturnsScopeForbiddenCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeAccess,
				[]string{authz.ScopeEchoRead},
				[]string{authz.AudiencePublic},
				"jti-scope-test",
			),
		)
		c.Next()
	})
	r.GET("/protected", RequireScopes(authz.ScopeAdminSite), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseErrorCode(rec.Body.Bytes()); got != commonModel.ErrCodeScopeForbidden {
		t.Fatalf("expected error code %s, got %s", commonModel.ErrCodeScopeForbidden, got)
	}
}

func TestRequireScopes_ReturnsAudienceForbiddenCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeAccess,
				[]string{authz.ScopeAdminSite},
				[]string{"unknown-audience"},
				"jti-audience-test",
			),
		)
		c.Next()
	})
	r.GET("/protected", RequireScopes(authz.ScopeAdminSite), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseErrorCode(rec.Body.Bytes()); got != commonModel.ErrCodeAudienceForbidden {
		t.Fatalf("expected error code %s, got %s", commonModel.ErrCodeAudienceForbidden, got)
	}
}

func TestRequireScopes_ProfileReadCannotAccessProfileWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeAccess,
				[]string{authz.ScopeProfileRead},
				[]string{authz.AudiencePublic},
				"jti-profile-write-test",
			),
		)
		c.Next()
	})
	r.PUT("/user", RequireScopes(authz.ScopeProfileWrite), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/user", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseErrorCode(rec.Body.Bytes()); got != commonModel.ErrCodeScopeForbidden {
		t.Fatalf("expected error code %s, got %s", commonModel.ErrCodeScopeForbidden, got)
	}
}

func TestRequireScopes_ProfileWriteAllowsAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeAccess,
				[]string{authz.ScopeProfileWrite},
				[]string{authz.AudiencePublic},
				"jti-profile-write-ok",
			),
		)
		c.Next()
	})
	r.PUT("/user", RequireScopes(authz.ScopeProfileWrite), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/user", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireSession_RejectsAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeAccess,
				[]string{authz.ScopeAdminSite},
				[]string{authz.AudienceCLI},
				"jti-session-test",
			),
		)
		c.Next()
	})
	r.POST("/passkey", RequireSession(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/passkey", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseErrorCode(rec.Body.Bytes()); got != commonModel.ErrCodeScopeForbidden {
		t.Fatalf("expected error code %s, got %s", commonModel.ErrCodeScopeForbidden, got)
	}
}

func TestRequireSession_AllowsSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeSession,
				nil,
				nil,
				"jti-session-ok",
			),
		)
		c.Next()
	})
	r.POST("/passkey", RequireSession(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/passkey", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireAdminSession_RejectsNonAdminSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeSession,
				nil,
				nil,
				"jti-admin-session-deny",
			),
		)
		c.Next()
	})
	r.GET("/logs", RequireAdminSession(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseErrorCode(rec.Body.Bytes()); got != commonModel.ErrCodeScopeForbidden {
		t.Fatalf("expected error code %s, got %s", commonModel.ErrCodeScopeForbidden, got)
	}
}

func TestRequireAdminSession_AllowsAdminSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeSession,
				nil,
				nil,
				"jti-admin-session-allow",
			).WithAdmin(true),
		)
		c.Next()
	})
	r.GET("/logs", RequireAdminSession(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireAdminSession_LeavesAccessTokensToScopeChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		viewer.AttachToRequest(
			&c.Request,
			viewer.NewUserViewerWithToken(
				"user-1",
				authModel.TokenTypeAccess,
				[]string{authz.ScopeAdminOps},
				[]string{authz.AudienceCLI},
				"jti-admin-access",
			),
		)
		c.Next()
	})
	r.GET("/logs", RequireAdminSession(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func parseErrorCode(body []byte) string {
	var payload struct {
		ErrorCode string `json:"error_code"`
	}
	_ = json.Unmarshal(body, &payload)
	return payload.ErrorCode
}
