// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/viewer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRequireAuth_RejectsTokenWithoutType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth(nil))
	r.GET("/api/user", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	legacyToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "legacy",
		"username": "legacy",
		"exp":      time.Now().UTC().Add(10 * time.Minute).Unix(),
	})
	tokenString, err := legacyToken.SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		t.Fatalf("sign legacy token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestOptionalAuth_AllowsAnonymousWithInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OptionalAuth(nil))
	r.GET("/api/echo/page", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/echo/page", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireAuth_RejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth(nil))
	r.GET("/api/s3/settings", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/s3/settings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAuth_RejectsAdminScopeTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth(nil))
	r.GET("/api/settings", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "admin-1", Username: "admin"}
	claims := jwtUtil.CreateAccessClaimsWithExpiry(
		user,
		int64(time.Hour),
		[]string{authz.ScopeAdminSite},
		authz.AudiencePublic,
		"jti-admin-query",
	)
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings?token="+tokenString, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseAuthDetail(rec.Body.Bytes()); got != apperr.ErrTokenQueryForbidden.Error() {
		t.Fatalf("expected the query-string rejection, got %q", got)
	}
}

func TestOptionalAuth_RejectsAdminScopeTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OptionalAuth(nil))
	r.GET("/api/echo/page", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "admin-1", Username: "admin"}
	claims := jwtUtil.CreateAccessClaimsWithExpiry(
		user,
		int64(time.Hour),
		[]string{authz.ScopeAdminSite},
		authz.AudiencePublic,
		"jti-admin-query-optional",
	)
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/echo/page?token="+tokenString, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseAuthDetail(rec.Body.Bytes()); got != apperr.ErrTokenQueryForbidden.Error() {
		t.Fatalf("expected the query-string rejection, got %q", got)
	}
}

func TestRequireAuth_AllowsSessionType(t *testing.T) {
	initMiddlewareTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth(nil))
	r.GET("/api/user", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "user-1", Username: "alice"}
	tokenString, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func initMiddlewareTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
	database.SetDB(db)
}

func parseAuthDetail(body []byte) string {
	var payload struct {
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(body, &payload)
	return payload.Detail
}

type accountStatusRevoker struct {
	state userModel.AccountState
	err   error
}

func (accountStatusRevoker) RevokeToken(string, time.Duration) {}
func (accountStatusRevoker) IsTokenRevoked(string) bool        { return false }
func (r accountStatusRevoker) AccountState(context.Context, string) (userModel.AccountState, error) {
	return r.state, r.err
}

var (
	activeAdminState = userModel.AccountState{
		Found: true, Role: userModel.RoleAdmin, Status: userModel.StatusActive,
	}
	disabledState = userModel.AccountState{
		Found: true, Role: userModel.RoleAdmin, Status: userModel.StatusDisabled,
	}
)

func TestRequireAuth_RejectsDisabledAccount(t *testing.T) {
	initMiddlewareTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth(accountStatusRevoker{state: disabledState}))
	r.GET("/api/user", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "user-1", Username: "alice"}
	tokenString, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseAuthDetail(rec.Body.Bytes()); got != apperr.ErrAccountDisabled.Error() {
		t.Fatalf("expected the disabled-account detail, got %q", got)
	}
}

func TestRequireAuth_AllowsActiveAccount(t *testing.T) {
	initMiddlewareTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth(accountStatusRevoker{state: activeAdminState}))
	r.GET("/api/user", func(c *gin.Context) {
		if got := viewer.MustFromContext(c.Request.Context()).Role(); got != userModel.RoleAdmin {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "user-1", Username: "alice"}
	tokenString, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestOptionalAuth_RejectsDisabledAccount(t *testing.T) {
	initMiddlewareTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OptionalAuth(accountStatusRevoker{state: disabledState}))
	r.POST("/api/comments", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "user-1", Username: "alice"}
	tokenString, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/comments", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseAuthDetail(rec.Body.Bytes()); got != apperr.ErrAccountDisabled.Error() {
		t.Fatalf("expected the disabled-account detail, got %q", got)
	}
}

func TestOptionalAuth_AttachesActiveAccount(t *testing.T) {
	initMiddlewareTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(OptionalAuth(accountStatusRevoker{state: activeAdminState}))
	r.POST("/api/comments", func(c *gin.Context) {
		if got := viewer.MustFromContext(c.Request.Context()).UserID(); got != "user-1" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "user-1", Username: "alice"}
	tokenString, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/comments", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
