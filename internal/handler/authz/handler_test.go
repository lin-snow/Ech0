// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authzModel "github.com/lin-snow/ech0/internal/model/authz"
	authzService "github.com/lin-snow/ech0/internal/service/authz"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type fakeService struct {
	exchange    authzModel.TokenResponse
	exchangeErr *authzModel.Error
	registered  authzModel.ClientMetadata
	consent     authzModel.ConsentContext
}

func (f fakeService) ProtectedResourceMetadata(context.Context, string) authzModel.ProtectedResourceMetadata {
	return authzModel.ProtectedResourceMetadata{Resource: "https://ech0.test/mcp"}
}

func (f fakeService) AuthorizationServerMetadata(context.Context, string) authzModel.AuthorizationServerMetadata {
	return authzModel.AuthorizationServerMetadata{Issuer: "https://ech0.test"}
}

func (f fakeService) Consent(context.Context, string, url.Values, string) (authzModel.ConsentContext, *authzModel.Error) {
	return f.consent, nil
}

func (f fakeService) Approve(context.Context, string, url.Values, string) (string, *authzModel.Error) {
	return "http://127.0.0.1:3000/callback?code=abc", nil
}

func (f fakeService) Deny(context.Context, string, url.Values) (string, *authzModel.Error) {
	return "http://127.0.0.1:3000/callback?error=access_denied", nil
}

func (f fakeService) RegisterClient(context.Context, authzModel.ClientMetadata) (authzModel.ClientMetadata, *authzModel.Error) {
	return f.registered, nil
}

func (f fakeService) Exchange(context.Context, string, url.Values) (authzModel.TokenResponse, *authzModel.Error) {
	return f.exchange, f.exchangeErr
}

func testRouter(service authzService.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewAuthzHandler(service)
	engine.GET("/.well-known/oauth-protected-resource", handler.ProtectedResource())
	engine.GET("/.well-known/oauth-authorization-server", handler.AuthorizationServerMetadata())
	engine.POST("/oauth/token", handler.Token())
	engine.POST("/oauth/register", handler.Register())
	engine.GET("/consent", handler.Consent())
	engine.POST("/authorize", handler.Approve())
	return engine
}

func TestTokenEndpointReturnsProtocolJSON(t *testing.T) {
	engine := testRouter(fakeService{exchange: authzModel.TokenResponse{
		AccessToken: "token", TokenType: "Bearer", ExpiresIn: 3600, Scope: "echo:read",
	}})

	request := httptest.NewRequest(http.MethodPost, "/oauth/token",
		strings.NewReader("grant_type=authorization_code&code=abc"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"access_token":"token"`) {
		t.Errorf("body = %s", body)
	}
}

func TestTokenEndpointMapsOAuthError(t *testing.T) {
	engine := testRouter(fakeService{exchangeErr: authzModel.ErrUnsupportedGrantType("nope")})
	request := httptest.NewRequest(http.MethodPost, "/oauth/token",
		strings.NewReader("grant_type=password"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"error":"unsupported_grant_type"`) {
		t.Errorf("body = %s", body)
	}
}

func TestRegisterEndpointReturns201(t *testing.T) {
	engine := testRouter(fakeService{registered: authzModel.ClientMetadata{
		ClientID: "ech0-test", ClientName: "Test", RedirectURIs: []string{"http://127.0.0.1:3000/cb"},
	}})
	request := httptest.NewRequest(http.MethodPost, "/oauth/register",
		strings.NewReader(`{"client_name":"Test","redirect_uris":["http://127.0.0.1:3000/cb"]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"client_id":"ech0-test"`) {
		t.Errorf("body = %s", body)
	}
}

func TestConsentRejectsAccessToken(t *testing.T) {
	engine := testRouter(fakeService{})
	request := httptest.NewRequest(http.MethodGet, "/consent", nil)
	accessViewer := viewer.NewUserViewerWithToken(
		"user-1", authModel.TokenTypeAccess, []string{authModel.ScopeEchoRead}, nil, "jti")
	request = request.WithContext(viewer.WithContext(request.Context(), accessViewer))
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
}

func TestConsentReturnsEnvelopeForSession(t *testing.T) {
	engine := testRouter(fakeService{consent: authzModel.ConsentContext{
		ClientName: "Test Client", RedirectURI: "http://127.0.0.1:3000/cb",
		Scopes: []string{"echo:read"}, Resource: "https://ech0.test/mcp",
	}})
	request := httptest.NewRequest(http.MethodGet, "/consent", nil)
	sessionViewer := viewer.NewUserViewerWithToken(
		"user-1", authModel.TokenTypeSession, nil, nil, "jti")
	request = request.WithContext(viewer.WithContext(request.Context(), sessionViewer))
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"client_name":"Test Client"`) {
		t.Errorf("body = %s", body)
	}
}
