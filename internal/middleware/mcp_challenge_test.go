// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResourceMetadataChallenge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET(
		"/mcp",
		RequireAuthWithChallenge(nil, ResourceMetadataChallenge("/.well-known/oauth-protected-resource/mcp")),
		func(ctx *gin.Context) { ctx.Status(http.StatusOK) },
	)

	request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	request.Host = "ech0.example"
	request.Header.Set("X-Forwarded-Proto", "https")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	want := `Bearer resource_metadata="https://ech0.example/.well-known/oauth-protected-resource/mcp"`
	if got := recorder.Header().Get("WWW-Authenticate"); got != want {
		t.Errorf("WWW-Authenticate = %q, want %q", got, want)
	}
}
