// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/authz"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type fakeClients struct {
	byID map[string]oauthserverModel.Client
}

func newFakeClients() *fakeClients {
	return &fakeClients{byID: make(map[string]oauthserverModel.Client)}
}

func (f *fakeClients) FindClient(_ context.Context, clientID string) (oauthserverModel.Client, error) {
	client, ok := f.byID[clientID]
	if !ok {
		return oauthserverModel.Client{}, errors.New("not found")
	}
	return client, nil
}

func (f *fakeClients) CreateClient(_ context.Context, client *oauthserverModel.Client) error {
	f.byID[client.ClientID] = *client
	return nil
}

func (f *fakeClients) CountClients(context.Context) (int64, error) {
	return int64(len(f.byID)), nil
}

type fakeUsers struct {
	user userModel.User
}

func (f fakeUsers) GetUserByID(_ context.Context, id string) (userModel.User, error) {
	if id != f.user.ID {
		return userModel.User{}, errors.New("not found")
	}
	return f.user, nil
}

type fakeIssuer struct {
	calls int
	token string
}

func (f *fakeIssuer) IssueForUser(
	_ context.Context,
	_ string,
	_ *settingModel.AccessTokenSettingDto,
) (string, error) {
	f.calls++
	return f.token, nil
}

func testService() (*OAuthServerService, *fakeClients, *fakeIssuer, userModel.User) {
	clients := newFakeClients()
	user := userModel.User{ID: "user-1", Username: "admin", IsAdmin: true}
	issuer := &fakeIssuer{token: "access-token"}
	service := NewOAuthServerService(clients, fakeUsers{user: user}, issuer)
	return service, clients, issuer, user
}

func pkcePair() (verifier, challenge string) {
	verifier = strings.Repeat("a", 64)
	sum := sha256.Sum256([]byte(verifier))
	return verifier, base64.RawURLEncoding.EncodeToString(sum[:])
}

func registerTestClient(t *testing.T, service *OAuthServerService) oauthserverModel.ClientMetadata {
	t.Helper()
	registered, oauthErr := service.RegisterClient(context.Background(), oauthserverModel.ClientMetadata{
		ClientName:      "Test Client",
		RedirectURIs:    []string{"http://127.0.0.1:3000/callback"},
		ApplicationType: "native",
	})
	if oauthErr != nil {
		t.Fatalf("register client: %v", oauthErr)
	}
	return registered
}

func authorizeQuery(clientID, redirectURI, verifierChallenge string) url.Values {
	return url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {"echo:read echo:write"},
		"state":                 {"xyz"},
		"resource":              {"https://ech0.test/mcp"},
		"code_challenge":        {verifierChallenge},
		"code_challenge_method": {"S256"},
	}
}

func TestMetadataAdvertisesBestPractices(t *testing.T) {
	service, _, _, _ := testService()
	ctx := context.Background()

	prm := service.ProtectedResourceMetadata(ctx, "https://ech0.test")
	if prm.Resource != "https://ech0.test/mcp" {
		t.Errorf("resource = %q", prm.Resource)
	}
	if len(prm.AuthorizationServers) != 1 || prm.AuthorizationServers[0] != "https://ech0.test" {
		t.Errorf("authorization_servers = %v", prm.AuthorizationServers)
	}
	if len(prm.ScopesSupported) != len(authz.AllScopes()) {
		t.Errorf("scopes_supported = %v", prm.ScopesSupported)
	}

	as := service.AuthorizationServerMetadata(ctx, "https://ech0.test")
	if !as.ClientIDMetadataDocumentSupported {
		t.Error("client_id_metadata_document_supported must be true")
	}
	if !as.AuthorizationResponseIssParameterSupported {
		t.Error("authorization_response_iss_parameter_supported must be true")
	}
	if len(as.CodeChallengeMethodsSupported) != 1 || as.CodeChallengeMethodsSupported[0] != "S256" {
		t.Errorf("code_challenge_methods_supported = %v", as.CodeChallengeMethodsSupported)
	}
	if as.TokenEndpoint != "https://ech0.test/oauth/token" {
		t.Errorf("token_endpoint = %q", as.TokenEndpoint)
	}
}

func TestRegisterClientDefaultsAndValidation(t *testing.T) {
	service, _, _, _ := testService()
	registered := registerTestClient(t, service)
	if registered.ClientID == "" || registered.ClientIDIssuedAt == 0 {
		t.Fatalf("registered metadata = %+v", registered)
	}
	if registered.TokenEndpointAuthMethod != "none" {
		t.Errorf("auth method = %q", registered.TokenEndpointAuthMethod)
	}
	if len(registered.GrantTypes) != 1 || registered.GrantTypes[0] != "authorization_code" {
		t.Errorf("grant types = %v", registered.GrantTypes)
	}

	cases := map[string]oauthserverModel.ClientMetadata{
		"missing redirect":  {},
		"fragment redirect": {RedirectURIs: []string{"http://127.0.0.1:3000/cb#frag"}},
		"remote http":       {RedirectURIs: []string{"http://example.com/cb"}},
		"confidential":      {RedirectURIs: []string{"http://127.0.0.1:3000/cb"}, TokenEndpointAuthMethod: "client_secret_basic"},
		"unsupported grant": {RedirectURIs: []string{"http://127.0.0.1:3000/cb"}, GrantTypes: []string{"client_credentials"}},
	}
	for name, meta := range cases {
		t.Run(name, func(t *testing.T) {
			if _, oauthErr := service.RegisterClient(context.Background(), meta); oauthErr == nil {
				t.Fatal("expected registration to be rejected")
			}
		})
	}
}

func TestAuthorizationRequestValidation(t *testing.T) {
	service, _, _, _ := testService()
	client := registerTestClient(t, service)
	_, challenge := pkcePair()
	base := authorizeQuery(client.ClientID, "http://127.0.0.1:3000/callback", challenge)

	if _, oauthErr := service.AuthorizationRequest(context.Background(), "https://ech0.test", base); oauthErr != nil {
		t.Fatalf("baseline request rejected: %v", oauthErr)
	}

	cases := map[string]func(url.Values){
		"unregistered redirect": func(q url.Values) { q.Set("redirect_uri", "http://127.0.0.1:3000/other") },
		"missing pkce":          func(q url.Values) { q.Del("code_challenge") },
		"plain pkce":            func(q url.Values) { q.Set("code_challenge_method", "plain") },
		"wrong resource":        func(q url.Values) { q.Set("resource", "https://evil.test/mcp") },
		"unknown scope":         func(q url.Values) { q.Set("scope", "echo:read root:everything") },
		"wrong response type":   func(q url.Values) { q.Set("response_type", "token") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			query := cloneValues(base)
			mutate(query)
			if _, oauthErr := service.AuthorizationRequest(context.Background(), "https://ech0.test", query); oauthErr == nil {
				t.Fatal("expected request to be rejected")
			}
		})
	}

	t.Run("default scopes", func(t *testing.T) {
		query := cloneValues(base)
		query.Del("scope")
		request, oauthErr := service.AuthorizationRequest(context.Background(), "https://ech0.test", query)
		if oauthErr != nil {
			t.Fatalf("unexpected error: %v", oauthErr)
		}
		if strings.Join(request.Scopes, " ") != strings.Join(oauthserverModel.DefaultScopes, " ") {
			t.Errorf("scopes = %v", request.Scopes)
		}
	})
}

func TestApproveAndExchange(t *testing.T) {
	service, clients, issuer, user := testService()
	client := registerTestClient(t, service)
	_ = clients
	verifier, challenge := pkcePair()
	query := authorizeQuery(client.ClientID, "http://127.0.0.1:3000/callback", challenge)

	redirect, oauthErr := service.Approve(context.Background(), "https://ech0.test", query, user.ID)
	if oauthErr != nil {
		t.Fatalf("approve: %v", oauthErr)
	}
	parsed, err := url.Parse(redirect)
	if err != nil {
		t.Fatalf("parse redirect: %v", err)
	}
	if parsed.Query().Get("state") != "xyz" {
		t.Errorf("state not echoed: %s", redirect)
	}
	if parsed.Query().Get("iss") != "https://ech0.test" {
		t.Errorf("iss missing: %s", redirect)
	}
	code := parsed.Query().Get("code")
	if code == "" {
		t.Fatalf("code missing: %s", redirect)
	}

	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {client.ClientID},
		"redirect_uri":  {"http://127.0.0.1:3000/callback"},
		"code_verifier": {verifier},
		"resource":      {"https://ech0.test/mcp"},
	}
	token, oauthErr := service.Exchange(context.Background(), "https://ech0.test", form)
	if oauthErr != nil {
		t.Fatalf("exchange: %v", oauthErr)
	}
	if token.AccessToken != "access-token" || issuer.calls != 1 {
		t.Fatalf("token = %+v, issuer calls = %d", token, issuer.calls)
	}
	if token.Scope != "echo:read echo:write" {
		t.Errorf("scope = %q", token.Scope)
	}

	if _, oauthErr := service.Exchange(context.Background(), "https://ech0.test", form); oauthErr == nil {
		t.Fatal("replayed code was accepted")
	}

	second := freshCode(t, service, client.ClientID, challenge, user.ID)
	badPKCE := cloneValues(form)
	badPKCE.Set("code", second)
	badPKCE.Set("code_verifier", strings.Repeat("b", 64))
	if _, oauthErr := service.Exchange(context.Background(), "https://ech0.test", badPKCE); oauthErr == nil {
		t.Fatal("wrong PKCE verifier was accepted")
	}

	third := freshCode(t, service, client.ClientID, challenge, user.ID)
	wrongClient := cloneValues(form)
	wrongClient.Set("code", third)
	wrongClient.Set("client_id", "another-client")
	if _, oauthErr := service.Exchange(context.Background(), "https://ech0.test", wrongClient); oauthErr == nil {
		t.Fatal("client mismatch was accepted")
	}

	fourth := freshCode(t, service, client.ClientID, challenge, user.ID)
	wrongRedirect := cloneValues(form)
	wrongRedirect.Set("code", fourth)
	wrongRedirect.Set("redirect_uri", "http://127.0.0.1:3000/other")
	if _, oauthErr := service.Exchange(context.Background(), "https://ech0.test", wrongRedirect); oauthErr == nil {
		t.Fatal("redirect mismatch was accepted")
	}
}

func freshCode(t *testing.T, service *OAuthServerService, clientID, challenge, userID string) string {
	t.Helper()
	redirect, oauthErr := service.Approve(
		context.Background(), "https://ech0.test",
		authorizeQuery(clientID, "http://127.0.0.1:3000/callback", challenge), userID)
	if oauthErr != nil {
		t.Fatalf("approve: %v", oauthErr)
	}
	parsed, _ := url.Parse(redirect)
	return parsed.Query().Get("code")
}

func TestConsentRequiresAdministrator(t *testing.T) {
	clients := newFakeClients()
	user := userModel.User{ID: "user-2", Username: "member", IsAdmin: false}
	service := NewOAuthServerService(clients, fakeUsers{user: user}, &fakeIssuer{})
	client := registerTestClient(t, service)
	_, challenge := pkcePair()
	query := authorizeQuery(client.ClientID, "http://127.0.0.1:3000/callback", challenge)

	if _, oauthErr := service.Consent(context.Background(), "https://ech0.test", query, user.ID); oauthErr == nil ||
		oauthErr.Code != "access_denied" {
		t.Fatalf("error = %v, want access_denied", oauthErr)
	}
}

func TestDenyRedirectsWithAccessDenied(t *testing.T) {
	service, _, _, _ := testService()
	client := registerTestClient(t, service)
	_, challenge := pkcePair()
	query := authorizeQuery(client.ClientID, "http://127.0.0.1:3000/callback", challenge)

	redirect, oauthErr := service.Deny(context.Background(), "https://ech0.test", query)
	if oauthErr != nil {
		t.Fatalf("deny: %v", oauthErr)
	}
	parsed, _ := url.Parse(redirect)
	if parsed.Query().Get("error") != "access_denied" {
		t.Errorf("redirect = %s", redirect)
	}
	if parsed.Query().Get("state") != "xyz" {
		t.Errorf("state missing: %s", redirect)
	}
}

func TestClientIDMetadataDocument(t *testing.T) {
	service, _, _, user := testService()
	_, challenge := pkcePair()
	verifier, _ := pkcePair()

	const clientID = "https://client.example.com/client.json"
	service.http = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, _ := json.Marshal(oauthserverModel.ClientMetadata{
			ClientID:                clientID,
			ClientName:              "CIMD Client",
			RedirectURIs:            []string{"http://127.0.0.1:3000/callback"},
			GrantTypes:              []string{"authorization_code"},
			ResponseTypes:           []string{"code"},
			TokenEndpointAuthMethod: "none",
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    req,
		}, nil
	})}

	query := authorizeQuery(clientID, "http://127.0.0.1:3000/callback", challenge)
	consent, oauthErr := service.Consent(context.Background(), "https://ech0.test", query, user.ID)
	if oauthErr != nil {
		t.Fatalf("consent: %v", oauthErr)
	}
	if consent.ClientName != "CIMD Client" {
		t.Errorf("client name = %q", consent.ClientName)
	}
	if consent.ResourceName != "Ech0 MCP" {
		t.Errorf("resource name = %q", consent.ResourceName)
	}

	redirect, oauthErr := service.Approve(context.Background(), "https://ech0.test", query, user.ID)
	if oauthErr != nil {
		t.Fatalf("approve: %v", oauthErr)
	}
	parsed, _ := url.Parse(redirect)
	token, oauthErr := service.Exchange(context.Background(), "https://ech0.test", url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {parsed.Query().Get("code")},
		"client_id":     {clientID},
		"redirect_uri":  {"http://127.0.0.1:3000/callback"},
		"code_verifier": {verifier},
	})
	if oauthErr != nil {
		t.Fatalf("exchange: %v", oauthErr)
	}
	if token.AccessToken == "" {
		t.Fatal("empty access token")
	}

	badQuery := cloneValues(query)
	badQuery.Set("client_id", "https://other.example.com/client.json")
	if _, oauthErr := service.Consent(context.Background(), "https://ech0.test", badQuery, user.ID); oauthErr == nil {
		t.Fatal("client_id mismatch was accepted")
	}
}

func cloneValues(values url.Values) url.Values {
	cloned := make(url.Values, len(values))
	for key, list := range values {
		cloned[key] = append([]string(nil), list...)
	}
	return cloned
}

func TestCacheLifetimeHonorsNoStore(t *testing.T) {
	cases := map[string]time.Duration{
		"no-store":       0,
		"max-age=120":    2 * time.Minute,
		"max-age=1":      time.Minute,
		"max-age=999999": time.Hour,
		"":               cimdCacheTTL,
	}
	for header, want := range cases {
		if got := cacheLifetime(header); got != want {
			t.Errorf("cacheLifetime(%q) = %v, want %v", header, got, want)
		}
	}
}

func TestTokenNameIsRuneSafe(t *testing.T) {
	name := tokenName("mcp", strings.Repeat("世", 200))
	if len([]rune(name)) != 120 {
		t.Fatalf("rune length = %d, want 120", len([]rune(name)))
	}
}
