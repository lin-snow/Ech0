// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// withViewer installs a principal in the request context without the auth
// middleware, so each test isolates the gate it names.
func withViewer(v viewer.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		viewer.AttachToRequest(&c.Request, v)
		c.Next()
	}
}

func TestRequireScopes_ReturnsScopeForbiddenCode(t *testing.T) {
	r := newTestRouter(withViewer(accessToken("user-1", "jti-scope-test", []string{authz.ScopeEchoRead})))
	r.GET("/protected", RequireScopes(authz.ScopeAdminSite), okHandler)

	rec := serve(r, http.MethodGet, "/protected")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrScopeForbidden.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrScopeForbidden.Error(), got)
	}
}

func TestRequireScopes_ReturnsAudienceForbiddenCode(t *testing.T) {
	r := newTestRouter(withViewer(viewer.NewAccess("user-1", "jti-audience-test", role.Member, []string{authz.ScopeAdminSite}, []string{"unknown-audience"})))
	r.GET("/protected", RequireScopes(authz.ScopeAdminSite), okHandler)

	rec := serve(r, http.MethodGet, "/protected")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrAudienceForbidden.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrAudienceForbidden.Error(), got)
	}
}

func TestRequireScopes_ProfileReadCannotAccessProfileWrite(t *testing.T) {
	r := newTestRouter(withViewer(accessToken("user-1", "jti-profile-write-test", []string{authz.ScopeProfileRead})))
	r.PUT("/user", RequireScopes(authz.ScopeProfileWrite), okHandler)

	rec := serve(r, http.MethodPut, "/user")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrScopeForbidden.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrScopeForbidden.Error(), got)
	}
}

func TestRequireScopes_ProfileWriteAllowsAccess(t *testing.T) {
	r := newTestRouter(withViewer(accessToken("user-1", "jti-profile-write-ok", []string{authz.ScopeProfileWrite})))
	r.PUT("/user", RequireScopes(authz.ScopeProfileWrite), okHandler)

	rec := serve(r, http.MethodPut, "/user")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// A session carries the scopes its role implies, so a member session reaches
// member endpoints without any bypass branch in the middleware.
func TestRequireScopes_SessionUsesRoleDerivedScopes(t *testing.T) {
	r := newTestRouter(withViewer(session("user-1", role.Member)))
	r.GET("/user", RequireScopes(authz.ScopeProfileRead), okHandler)

	rec := serve(r, http.MethodGet, "/user")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireScopes_SessionMemberDeniedAdminScope(t *testing.T) {
	r := newTestRouter(withViewer(session("user-1", role.Member)))
	r.GET("/settings", RequireScopes(authz.ScopeAdminSite), okHandler)

	rec := serve(r, http.MethodGet, "/settings")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireScopes_SessionAdminAdmittedByRole(t *testing.T) {
	r := newTestRouter(withViewer(session("user-1", role.Admin)))
	r.GET("/settings", RequireScopes(authz.ScopeAdminSite), okHandler)

	rec := serve(r, http.MethodGet, "/settings")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireScopes_RejectsAnonymous(t *testing.T) {
	r := newTestRouter(withViewer(viewer.Anonymous))
	r.GET("/user", RequireScopes(authz.ScopeProfileRead), okHandler)

	rec := serve(r, http.MethodGet, "/user")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrTokenInvalid.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrTokenInvalid.Error(), got)
	}
}

func TestRequireMinRank_RejectsMemberSession(t *testing.T) {
	r := newTestRouter(withViewer(session("user-1", role.Member)))
	r.GET("/logs", RequireMinRank(role.Admin), okHandler)

	rec := serve(r, http.MethodGet, "/logs")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrScopeForbidden.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrScopeForbidden.Error(), got)
	}
}

func TestRequireMinRank_AllowsAdminSession(t *testing.T) {
	r := newTestRouter(withViewer(session("user-1", role.Admin)))
	r.GET("/logs", RequireMinRank(role.Admin), okHandler)

	rec := serve(r, http.MethodGet, "/logs")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// Rank comes from the live account, not from the token's scopes: a token that
// carries an admin scope still fails when the account has been demoted.
func TestRequireMinRank_AccessTokenUsesLiveRole(t *testing.T) {
	token := accessToken("user-1", "jti-admin-access", []string{authz.ScopeAdminOps})

	r := newTestRouter(withViewer(token))
	r.GET("/logs", RequireMinRank(role.Admin), okHandler)
	if rec := serve(r, http.MethodGet, "/logs"); rec.Code != http.StatusForbidden {
		t.Fatalf("member token: expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	promoted := viewer.NewAccess("user-1", "jti-admin-access", role.Admin, []string{authz.ScopeAdminOps}, []string{authz.AudienceCLI})
	r2 := newTestRouter(withViewer(promoted))
	r2.GET("/logs", RequireMinRank(role.Admin), okHandler)
	if rec := serve(r2, http.MethodGet, "/logs"); rec.Code != http.StatusOK {
		t.Fatalf("admin token: expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireSession_RejectsAccessToken(t *testing.T) {
	r := newTestRouter(withViewer(accessToken("user-1", "jti-session-test", []string{authz.ScopeAdminSite})))
	r.POST("/passkey", RequireSession(), okHandler)

	rec := serve(r, http.MethodPost, "/passkey")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrScopeForbidden.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrScopeForbidden.Error(), got)
	}
}

func TestRequireSession_AllowsSession(t *testing.T) {
	r := newTestRouter(withViewer(session("user-1", role.Member)))
	r.POST("/passkey", RequireSession(), okHandler)

	rec := serve(r, http.MethodPost, "/passkey")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireAudienceWithChallenge_Returns401AndMetadata(t *testing.T) {
	r := newTestRouter(withViewer(viewer.NewAccess("user-1", "jti-audience-mismatch", role.Member, []string{authz.ScopeEchoRead}, []string{authz.AudienceCLI})))
	path := "/.well-known/oauth-protected-resource/mcp"
	r.GET("/mcp", RequireAudienceWithChallenge(InvalidTokenChallenge(path), authz.AudienceMCPRemote), okHandler)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Host = "ech0.example"
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	want := `Bearer error="invalid_token", resource_metadata="https://ech0.example` + path + `"`
	if got := rec.Header().Get("WWW-Authenticate"); got != want {
		t.Fatalf("WWW-Authenticate = %q, want %q", got, want)
	}
	if got := parseDetail(rec.Body.Bytes()); got != apperr.ErrTokenInvalid.Error() {
		t.Fatalf("expected detail %q, got %q", apperr.ErrTokenInvalid.Error(), got)
	}
}

func TestRequireAudienceWithChallenge_AllowsMatchingAudience(t *testing.T) {
	r := newTestRouter(withViewer(viewer.NewAccess("user-1", "jti-audience-match", role.Member, nil, []string{authz.AudienceMCPRemote})))
	r.GET("/mcp", RequireAudienceWithChallenge(InvalidTokenChallenge("/.well-known/oauth-protected-resource/mcp"), authz.AudienceMCPRemote), okHandler)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func accessToken(userID, jti string, scopes []string) viewer.Context {
	return viewer.NewAccess(userID, jti, role.Member, scopes, []string{authz.AudiencePublic})
}

func session(userID string, r role.Role) viewer.Context {
	return viewer.NewSession(userID, r, authz.ScopesForRole(r))
}

func newTestRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewares...)
	return r
}

func okHandler(c *gin.Context) { c.Status(http.StatusOK) }

func serve(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func parseDetail(body []byte) string {
	var payload struct {
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(body, &payload)
	return payload.Detail
}
