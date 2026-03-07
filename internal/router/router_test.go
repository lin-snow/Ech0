package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	inboxHandler "github.com/lin-snow/ech0/internal/handler/inbox"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	todoHandler "github.com/lin-snow/ech0/internal/handler/todo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
)

func TestSetupRouter_RegistersKeyRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers())

	expectRoutes := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/swagger/*any"},
		{method: http.MethodPost, path: "/api/login"},
		{method: http.MethodPost, path: "/api/echo"},
		{method: http.MethodGet, path: "/api/status"},
		{method: http.MethodGet, path: "/api/settings"},
		{method: http.MethodGet, path: "/api/agent/recent"},
		{method: http.MethodGet, path: "/ws/dashboard/metrics"},
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
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers())

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
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

func buildTestHandlers() *handler.Bundle {
	return handler.NewBundle(
		webHandler.NewWebHandler(),
		userHandler.NewUserHandler(nil),
		echoHandler.NewEchoHandler(nil),
		commonHandler.NewCommonHandler(nil),
		settingHandler.NewSettingHandler(nil),
		inboxHandler.NewInboxHandler(nil),
		todoHandler.NewTodoHandler(nil),
		connectHandler.NewConnectHandler(nil),
		backupHandler.NewBackupHandler(nil),
		dashboardHandler.NewDashboardHandler(nil),
		agentHandler.NewAgentHandler(nil),
	)
}
