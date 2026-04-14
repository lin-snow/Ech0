package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/handler"
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	authHandler "github.com/lin-snow/ech0/internal/handler/auth"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commentHandler "github.com/lin-snow/ech0/internal/handler/comment"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	fileHandler "github.com/lin-snow/ech0/internal/handler/file"
	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	migrationHandler "github.com/lin-snow/ech0/internal/handler/migration"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
	"github.com/lin-snow/ech0/internal/mcp"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
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
		{method: http.MethodGet, path: "/swagger/*any"},
		{method: http.MethodPost, path: "/api/login"},
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
	api.Use(middleware.NoCache(), middleware.JWTAuthMiddleware(nil))
	api.PUT(
		"/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		func(ctx *gin.Context) { ctx.Status(http.StatusOK) },
	)

	user := userModel.User{ID: "u-1", Username: "scope-user"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authModel.ScopeEchoRead},
			authModel.AudiencePublic,
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
	api.Use(middleware.NoCache(), middleware.JWTAuthMiddleware(nil))
	api.PUT(
		"/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		func(ctx *gin.Context) { ctx.Status(http.StatusOK) },
	)

	user := userModel.User{ID: "u-2", Username: "scope-admin"}
	token, err := jwtUtil.GenerateToken(
		jwtUtil.CreateAccessClaimsWithExpiry(
			user,
			int64(time.Hour),
			[]string{authModel.ScopeAdminSettings},
			authModel.AudiencePublic,
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
			[]string{authModel.ScopeEchoRead},
			authModel.AudienceIntegration,
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
			[]string{authModel.ScopeCommentWrite},
			authModel.AudiencePublic,
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
		backupHandler.NewBackupHandler(nil),
		migrationHandler.NewMigrationHandler(nil),
		dashboardHandler.NewDashboardHandler(nil),
		agentHandler.NewAgentHandler(nil),
		mcp.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil),
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
