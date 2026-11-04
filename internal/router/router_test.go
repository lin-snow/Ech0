// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/handler"
	authHandler "github.com/lin-snow/ech0/internal/handler/auth"
	commentHandler "github.com/lin-snow/ech0/internal/handler/comment"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	copilotHandler "github.com/lin-snow/ech0/internal/handler/copilot"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	embeddingHandler "github.com/lin-snow/ech0/internal/handler/embedding"
	fileHandler "github.com/lin-snow/ech0/internal/handler/file"
	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	migratorHandler "github.com/lin-snow/ech0/internal/handler/migrator"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
	"github.com/lin-snow/ech0/internal/mcp"
	"github.com/lin-snow/ech0/internal/middleware"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/internal/visitor"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSetupRouter_RegistersKeyRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	expectRoutes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/docs"},
		{method: http.MethodGet, path: "/api/openapi.json"},
		{method: http.MethodPost, path: "/api/login"},
		{method: http.MethodPost, path: "/api/auth/device/code"},
		{method: http.MethodPost, path: "/api/auth/device/token"},
		{method: http.MethodGet, path: "/api/auth/device/pending"},
		{method: http.MethodPost, path: "/api/auth/device/approve"},
		{method: http.MethodPost, path: "/api/echo"},
		{method: http.MethodGet, path: "/api/init/status"},
		{method: http.MethodGet, path: "/api/settings"},
		{method: http.MethodGet, path: "/api/agent/recent"},
		{method: http.MethodPost, path: "/api/connects"},
		{method: http.MethodDelete, path: "/api/connects/:id"},
		{method: http.MethodGet, path: "/api/connects/health"},
		{method: http.MethodGet, path: "/api/system/logs"},
		{method: http.MethodGet, path: "/api/system/logs/stream"},
		{method: http.MethodGet, path: "/ws/system/logs"},
	}

	routes := engine.Routes()
	for _, expected := range expectRoutes {
		if !containsRoute(routes, expected.method, expected.path) {
			t.Fatalf("expected route missing: %s %s", expected.method, expected.path)
		}
	}
}

func TestSetupRouter_EnvelopeContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode envelope failed: %v (body=%s)", err, rec.Body.String())
	}
	if body.Code != 1 {
		t.Fatalf("expected success code=1, got %d", body.Code)
	}
	if body.Msg == "" {
		t.Fatalf("expected non-empty msg, got empty")
	}
	if body.Data["hello"] == nil {
		t.Fatalf("expected data.hello carried through envelope, got data=%v", body.Data)
	}
}

func TestSetupRouter_AuthGroupProtected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSetupRouter_PublicEchoRoutesAllowAnonymous(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	for _, path := range []string{
		"/api/echo/today",
		"/api/echo/hot",
		"/api/echo/random",
		"/api/echo/onthisday",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected %s to be anonymously reachable, got 401", path)
		}
	}
}

func TestSetupRouter_AllUsersRouteProtected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSetupRouter_AccessTokenWithoutRequiredScopeGetsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()

	api := engine.Group("/api")
	api.Use(middleware.NoCache(), middleware.RequireAuth(nil))
	api.PUT(
		"/settings",
		middleware.RequireScopes(authz.ScopeAdminSite),
		func(ctx *gin.Context) { ctx.Status(http.StatusOK) },
	)

	user := userModel.User{ID: "u-1", Username: "scope-user"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeEchoRead},
			authz.AudiencePublic,
			"jti-read-only",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestSetupRouter_AccessTokenWithScopePasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()

	api := engine.Group("/api")
	api.Use(middleware.NoCache(), middleware.RequireAuth(nil))
	api.PUT(
		"/settings",
		middleware.RequireScopes(authz.ScopeAdminSite),
		func(ctx *gin.Context) { ctx.Status(http.StatusOK) },
	)

	user := userModel.User{ID: "u-2", Username: "scope-admin"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeAdminSite},
			authz.AudiencePublic,
			"jti-admin",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestSetupRouter_DeleteRoutesRequireDeleteScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	user := userModel.User{ID: "u-delete-1", Username: "delete-scope"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeEchoWrite, authz.ScopeFileWrite, authz.ScopeConnectWrite},
			authz.AudienceCLI,
			"jti-write-only",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	paths := []string{
		"/api/echo/0190f3f4-0000-7000-8000-000000000001",
		"/api/tag/0190f3f4-0000-7000-8000-000000000002",
		"/api/file/0190f3f4-0000-7000-8000-000000000003",
		"/api/connects/0190f3f4-0000-7000-8000-000000000004",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodDelete, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s status = %d, want %d for a write-only token", path, rec.Code, http.StatusForbidden)
		}
	}
}

func TestSetupRouter_AccessTokenCannotMintTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	user := userModel.User{ID: "u-token-1", Username: "token-scope"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			authz.AllScopes(),
			authz.AudienceCLI,
			"jti-token-admin",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/access-tokens", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("access token minted another token: status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestSetupRouter_IntegrationCommentRouteExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	if !containsRoute(engine.Routes(), http.MethodPost, "/api/comments/integration") {
		t.Fatal("expected route POST /api/comments/integration to be registered")
	}
}

func TestSetupRouter_IntegrationCommentRejectsNoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	req := httptest.NewRequest(http.MethodPost, "/api/comments/integration", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSetupRouter_IntegrationCommentRejectsWrongScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	user := userModel.User{ID: "u-integ-1", Username: "integ-user"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeEchoRead},
			authz.AudienceIntegration,
			"jti-wrong-scope",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/comments/integration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestSetupRouter_IntegrationCommentRejectsWrongAudience(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	user := userModel.User{ID: "u-integ-2", Username: "integ-user-2"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeCommentWrite},
			authz.AudiencePublic,
			"jti-wrong-aud",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/comments/integration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestSetupRouter_IntegrationCommentAcceptsCLIAudience(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	user := userModel.User{ID: "u-integ-3", Username: "integ-user-3"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeCommentWrite},
			authz.AudienceCLI,
			"jti-cli-aud",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/comments/integration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	// The audience + scope gate passes; the empty body fails later in validation.
	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Fatalf("expected CLI audience to pass the gate, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func containsRoute(routes []gin.RouteInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}

	return false
}

func buildTestMWDeps() *middleware.Deps {
	return middleware.NewDeps(nil)
}

func buildTestHandlers() *handler.Bundle {
	return handler.NewBundle(
		webHandler.NewWebHandler(visitor.NewTracker()),
		userHandler.NewUserHandler(nil),
		authHandler.NewAuthHandler(nil, nil),
		echoHandler.NewEchoHandler(nil),
		fileHandler.NewFileHandler(nil),
		commentHandler.NewCommentHandler(nil),
		initHandler.NewInitHandler(nil),
		commonHandler.NewCommonHandler(nil),
		settingHandler.NewSettingHandler(nil),
		connectHandler.NewConnectHandler(nil),
		migratorHandler.NewMigrationHandler(nil),
		dashboardHandler.NewDashboardHandler(nil),
		copilotHandler.NewCopilotHandler(nil, nil),
		embeddingHandler.NewEmbeddingHandler(nil),
		mcp.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil),
		nil,
	)
}

func initTestDatabase(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
	database.SetDB(db)
}

func TestSetupRouter_MCPAudienceMismatchReturnsChallenge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	user := userModel.User{ID: "u-mcp", Username: "mcp-user"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authz.ScopeMCPRead},
			authz.AudienceCLI,
			"jti-mcp-audience",
		),
	)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	req.Host = "ech0.example"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}
	challenge := rec.Header().Get("WWW-Authenticate")
	if !strings.Contains(challenge, `error="invalid_token"`) {
		t.Fatalf("challenge missing invalid_token: %q", challenge)
	}
	if !strings.Contains(challenge, `resource_metadata="https://ech0.example/.well-known/oauth-protected-resource/mcp"`) {
		t.Fatalf("challenge missing metadata URL: %q", challenge)
	}
}
