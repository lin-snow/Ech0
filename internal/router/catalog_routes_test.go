// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	openapiCatalog "github.com/lin-snow/ech0/internal/openapi/catalog"
)

// Routes that deliberately stay outside the catalog, with the reason.
var catalogExclusions = []struct {
	prefix string
	reason string
}{
	{"/api/docs", "OpenAPI UI"},
	{"/api/openapi.json", "OpenAPI spec"},
	{"/api/openapi.yaml", "OpenAPI spec"},
	{"/api/openapi-3.0.json", "OpenAPI spec"},
	{"/api/openapi-3.0.yaml", "OpenAPI spec"},
	{"/api/_catalog", "the catalog itself"},
	{"/api/files/", "static upload serving, addressed by URL not by operation"},
	{"/api/mcp", "JSON-RPC, tools are described by the MCP server itself"},
	{"/api/cap/", "third-party captcha handler"},
	{"/api/login", "the CLI signs in locally instead of shelling out to a command"},
	{"/api/auth/", "cookie/token plumbing plus device authorization, owned by `ech0 auth`"},
	{"/api/passkey/", "WebAuthn ceremonies need a browser"},
	{"/api/chat", "Ech0 Copilot: CLI support is intentionally deferred"},
}

func excludedFromCatalog(path string) bool {
	for _, exclusion := range catalogExclusions {
		if strings.HasPrefix(path, exclusion.prefix) {
			return true
		}
	}
	return false
}

func ginRouteKey(method, path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
			segments[index] = "{" + segment[1:] + "}"
		}
	}
	return method + " " + strings.Join(segments, "/")
}

// A hand-written raw operation is only safe if it cannot drift from the router:
// every catalogued operation must exist, and no /api or /ws route may escape
// the catalog unless it is in the exclusion list above.
func TestCatalogMatchesRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	handlers := buildTestHandlers()
	engine := gin.New()
	SetupRouter(engine, handlers, buildTestMWDeps())

	api := setupHumaAPI(gin.New())
	registerOperations(api, handlers, nil)
	doc, _, _, err := openapiCatalog.Build(api.OpenAPI())
	if err != nil {
		t.Fatalf("catalog.Build: %v", err)
	}

	routes := make(map[string]bool, len(engine.Routes()))
	for _, route := range engine.Routes() {
		routes[ginRouteKey(route.Method, route.Path)] = true
	}

	catalogued := make(map[string]string, len(doc.Ops))
	for id, op := range doc.Ops {
		mount := op.Mount
		if mount == "" {
			mount = openapiCatalog.MountAPI
		}
		key := ginRouteKey(strings.ToUpper(op.Method), "/"+mount+op.Path)
		catalogued[key] = id
		if !routes[key] {
			t.Errorf("catalog op %q (%s %s) has no route", id, op.Method, key)
		}
	}

	for _, route := range engine.Routes() {
		if route.Method == http.MethodHead || route.Method == http.MethodOptions {
			continue
		}
		if !strings.HasPrefix(route.Path, "/api/") && !strings.HasPrefix(route.Path, "/ws/") {
			continue
		}
		if excludedFromCatalog(route.Path) {
			continue
		}
		key := ginRouteKey(route.Method, route.Path)
		if _, ok := catalogued[key]; !ok {
			t.Errorf("route %s is neither in the catalog nor excluded", key)
		}
	}
}

func TestCatalogExclusionsAreDocumented(t *testing.T) {
	if len(catalogExclusions) == 0 {
		t.Fatal("catalog exclusions missing")
	}
	for _, exclusion := range catalogExclusions {
		if exclusion.reason == "" {
			t.Errorf("exclusion %q has no reason", exclusion.prefix)
		}
	}
}
