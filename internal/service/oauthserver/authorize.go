// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"
	"net/url"
	"strings"

	"github.com/lin-snow/ech0/internal/authz"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

const maxStateLength = 1024

// AuthorizationRequest validates an /oauth/authorize query. The checks are
// ordered so that client identity and redirect_uri are trusted before any
// response is derived from them.
func (s *OAuthServerService) AuthorizationRequest(
	ctx context.Context,
	origin string,
	query url.Values,
) (oauthserverModel.AuthorizationRequest, *oauthserverModel.Error) {
	clientID := strings.TrimSpace(query.Get("client_id"))
	redirectURI := strings.TrimSpace(query.Get("redirect_uri"))
	if redirectURI == "" {
		return oauthserverModel.AuthorizationRequest{}, oauthserverModel.ErrInvalidRequest("redirect_uri is required")
	}

	meta, oauthErr := s.resolveClient(ctx, clientID)
	if oauthErr != nil {
		return oauthserverModel.AuthorizationRequest{}, oauthErr
	}
	if !redirectURIRegistered(meta.RedirectURIs, redirectURI) {
		return oauthserverModel.AuthorizationRequest{}, oauthserverModel.ErrInvalidRequest(
			"redirect_uri is not registered for this client")
	}

	if responseType := query.Get("response_type"); responseType != "code" {
		return oauthserverModel.AuthorizationRequest{}, oauthserverModel.ErrUnsupportedResponseType(
			"response_type must be code")
	}

	challenge := strings.TrimSpace(query.Get("code_challenge"))
	if err := validateCodeChallenge(challenge); err != nil {
		return oauthserverModel.AuthorizationRequest{}, err
	}
	switch strings.ToUpper(strings.TrimSpace(query.Get("code_challenge_method"))) {
	case "S256":
	default:
		return oauthserverModel.AuthorizationRequest{}, oauthserverModel.ErrInvalidRequest(
			"code_challenge_method must be S256")
	}

	scopes, err := parseScopes(query.Get("scope"), s.resource().DefaultScopes)
	if err != nil {
		return oauthserverModel.AuthorizationRequest{}, err
	}

	resource := strings.TrimRight(strings.TrimSpace(query.Get("resource")), "/")
	expectedResource := s.resourceURI(ctx, origin)
	if resource == "" {
		resource = expectedResource
	} else if resource != expectedResource {
		return oauthserverModel.AuthorizationRequest{}, oauthserverModel.ErrInvalidRequest(
			"resource does not identify this MCP server")
	}

	state := query.Get("state")
	if len(state) > maxStateLength {
		return oauthserverModel.AuthorizationRequest{}, oauthserverModel.ErrInvalidRequest("state is too long")
	}

	clientName := strings.TrimSpace(meta.ClientName)
	if clientName == "" {
		clientName = clientID
	}

	return oauthserverModel.AuthorizationRequest{
		ClientID:            clientID,
		ClientName:          clientName,
		ClientURI:           meta.ClientURI,
		RedirectURI:         redirectURI,
		ResponseType:        "code",
		Scopes:              scopes,
		State:               state,
		Resource:            resource,
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	}, nil
}

// Consent validates the request and checks that the approving user may mint
// tokens. Ech0 restricts access-token issuance to administrators, so the same
// rule applies to OAuth consent.
func (s *OAuthServerService) Consent(
	ctx context.Context,
	origin string,
	query url.Values,
	userID string,
) (oauthserverModel.ConsentContext, *oauthserverModel.Error) {
	request, oauthErr := s.AuthorizationRequest(ctx, origin, query)
	if oauthErr != nil {
		return oauthserverModel.ConsentContext{}, oauthErr
	}
	if _, oauthErr := s.consentUser(ctx, userID); oauthErr != nil {
		return oauthserverModel.ConsentContext{}, oauthErr
	}
	return oauthserverModel.ConsentContext{
		ClientName:   request.ClientName,
		ClientURI:    request.ClientURI,
		RedirectURI:  request.RedirectURI,
		Scopes:       request.Scopes,
		Resource:     request.Resource,
		ResourceName: s.resource().Name,
		State:        request.State,
	}, nil
}

// Approve issues a single-use authorization code and returns the redirect the
// browser must follow. The code is bound to the client, redirect URI, scopes,
// resource and PKCE challenge.
func (s *OAuthServerService) Approve(
	ctx context.Context,
	origin string,
	query url.Values,
	userID string,
) (string, *oauthserverModel.Error) {
	request, oauthErr := s.AuthorizationRequest(ctx, origin, query)
	if oauthErr != nil {
		return "", oauthErr
	}
	user, oauthErr := s.consentUser(ctx, userID)
	if oauthErr != nil {
		// The request is fully validated, so RFC 6749 §4.1.2.1 wants the error
		// delivered to the client's redirect_uri rather than as local JSON.
		return s.authorizationErrorRedirect(ctx, origin, request, oauthErr), nil
	}
	code, oauthErr := s.signCode(ctx, origin, request, user)
	if oauthErr != nil {
		return "", oauthErr
	}
	return redirectWithValues(request.RedirectURI, url.Values{
		"code":  {code},
		"state": {request.State},
		"iss":   {s.issuer(ctx, origin)},
	}), nil
}

// Deny returns the RFC 6749 access_denied redirect for a validated request.
func (s *OAuthServerService) Deny(
	ctx context.Context,
	origin string,
	query url.Values,
) (string, *oauthserverModel.Error) {
	request, oauthErr := s.AuthorizationRequest(ctx, origin, query)
	if oauthErr != nil {
		return "", oauthErr
	}
	return s.authorizationErrorRedirect(ctx, origin, request,
		oauthserverModel.ErrAccessDenied("the user declined the authorization request")), nil
}

// authorizationErrorRedirect renders a post-validation OAuth error as a
// redirect to the client, reusing the deny path. It is only called after
// AuthorizationRequest has trusted client_id and redirect_uri.
func (s *OAuthServerService) authorizationErrorRedirect(
	ctx context.Context,
	origin string,
	request oauthserverModel.AuthorizationRequest,
	oauthErr *oauthserverModel.Error,
) string {
	return redirectWithValues(request.RedirectURI, url.Values{
		"error":             {oauthErr.Code},
		"error_description": {oauthErr.Description},
		"state":             {request.State},
		"iss":               {s.issuer(ctx, origin)},
	})
}

// consentUser enforces the population allowed to mint MCP tokens: Ech0
// restricts access-token issuance to administrators, so consent inherits the
// same rule.
func (s *OAuthServerService) consentUser(
	ctx context.Context,
	userID string,
) (userModel.User, *oauthserverModel.Error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return userModel.User{}, oauthserverModel.ErrAccessDenied("approving user not found")
	}
	if !user.IsAdmin {
		return userModel.User{}, oauthserverModel.ErrAccessDenied(
			"only administrators can authorize MCP clients")
	}
	return user, nil
}

func parseScopes(raw string, defaults []string) ([]string, *oauthserverModel.Error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return append([]string(nil), defaults...), nil
	}
	// fields may repeat a scope, so allow up to the vocabulary size; the
	// dedupe below makes the resulting grant at most ScopeCount entries.
	if len(fields) > authz.ScopeCount() {
		return nil, oauthserverModel.ErrInvalidScope("too many scopes requested")
	}
	seen := make(map[string]struct{}, len(fields))
	scopes := make([]string, 0, len(fields))
	for _, scope := range fields {
		if !authz.IsValidScope(scope) {
			return nil, oauthserverModel.ErrInvalidScope("unknown scope: " + scope)
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	return scopes, nil
}

func redirectURIRegistered(registered []string, requested string) bool {
	for _, candidate := range registered {
		if redirectURIMatches(candidate, requested) {
			return true
		}
	}
	return false
}

// redirectURIMatches normally requires an exact match, but RFC 8252 §7.3 lets
// a native client pick an ephemeral loopback port at request time: for two
// http loopback URLs the port is ignored and the path (plus query) must match.
func redirectURIMatches(registered, requested string) bool {
	if registered == requested {
		return true
	}
	want, err := url.Parse(registered)
	if err != nil {
		return false
	}
	got, err := url.Parse(requested)
	if err != nil {
		return false
	}
	if want.Scheme != "http" || got.Scheme != "http" {
		return false
	}
	if !isLoopbackHost(want.Hostname()) || !isLoopbackHost(got.Hostname()) {
		return false
	}
	if !strings.EqualFold(want.Hostname(), got.Hostname()) {
		return false
	}
	return want.Path == got.Path && want.RawQuery == got.RawQuery &&
		want.Fragment == "" && got.Fragment == ""
}

func validateCodeChallenge(challenge string) *oauthserverModel.Error {
	if len(challenge) < 43 || len(challenge) > 128 {
		return oauthserverModel.ErrInvalidRequest("code_challenge must be 43-128 characters")
	}
	for _, r := range challenge {
		if !isBase64URLRune(r) {
			return oauthserverModel.ErrInvalidRequest("code_challenge must be base64url")
		}
	}
	return nil
}

func isBase64URLRune(r rune) bool {
	return r == '-' || r == '_' || r == '=' ||
		(r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func redirectWithValues(rawURI string, values url.Values) string {
	parsed, err := url.Parse(rawURI)
	if err != nil {
		return rawURI
	}
	query := parsed.Query()
	for key, list := range values {
		if len(list) == 0 || list[0] == "" {
			continue
		}
		query.Set(key, list[0])
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
