// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"
	"net/url"
	"strings"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authzModel "github.com/lin-snow/ech0/internal/model/authz"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

const maxStateLength = 1024

// AuthorizationRequest validates an /oauth/authorize query. The checks are
// ordered so that client identity and redirect_uri are trusted before any
// response is derived from them.
func (s *AuthzService) AuthorizationRequest(
	ctx context.Context,
	origin string,
	query url.Values,
) (authzModel.AuthorizationRequest, *authzModel.Error) {
	clientID := strings.TrimSpace(query.Get("client_id"))
	redirectURI := strings.TrimSpace(query.Get("redirect_uri"))
	if redirectURI == "" {
		return authzModel.AuthorizationRequest{}, authzModel.ErrInvalidRequest("redirect_uri is required")
	}

	meta, oauthErr := s.resolveClient(ctx, clientID)
	if oauthErr != nil {
		return authzModel.AuthorizationRequest{}, oauthErr
	}
	if !contains(meta.RedirectURIs, redirectURI) {
		return authzModel.AuthorizationRequest{}, authzModel.ErrInvalidRequest(
			"redirect_uri is not registered for this client")
	}

	if responseType := query.Get("response_type"); responseType != "code" {
		return authzModel.AuthorizationRequest{}, authzModel.ErrUnsupportedResponseType(
			"response_type must be code")
	}

	challenge := strings.TrimSpace(query.Get("code_challenge"))
	if err := validateCodeChallenge(challenge); err != nil {
		return authzModel.AuthorizationRequest{}, err
	}
	switch strings.ToUpper(strings.TrimSpace(query.Get("code_challenge_method"))) {
	case "S256":
	default:
		return authzModel.AuthorizationRequest{}, authzModel.ErrInvalidRequest(
			"code_challenge_method must be S256")
	}

	scopes, err := parseScopes(query.Get("scope"))
	if err != nil {
		return authzModel.AuthorizationRequest{}, err
	}

	resource := strings.TrimRight(strings.TrimSpace(query.Get("resource")), "/")
	expectedResource := s.resourceURI(ctx, origin)
	if resource == "" {
		resource = expectedResource
	} else if resource != expectedResource {
		return authzModel.AuthorizationRequest{}, authzModel.ErrInvalidRequest(
			"resource does not identify this MCP server")
	}

	state := query.Get("state")
	if len(state) > maxStateLength {
		return authzModel.AuthorizationRequest{}, authzModel.ErrInvalidRequest("state is too long")
	}

	clientName := strings.TrimSpace(meta.ClientName)
	if clientName == "" {
		clientName = clientID
	}

	return authzModel.AuthorizationRequest{
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
func (s *AuthzService) Consent(
	ctx context.Context,
	origin string,
	query url.Values,
	userID string,
) (authzModel.ConsentContext, *authzModel.Error) {
	request, _, oauthErr := s.consentRequest(ctx, origin, query, userID)
	if oauthErr != nil {
		return authzModel.ConsentContext{}, oauthErr
	}
	return authzModel.ConsentContext{
		ClientName:  request.ClientName,
		ClientURI:   request.ClientURI,
		RedirectURI: request.RedirectURI,
		Scopes:      request.Scopes,
		Resource:    request.Resource,
		State:       request.State,
	}, nil
}

// Approve issues a single-use authorization code and returns the redirect the
// browser must follow. The code is bound to the client, redirect URI, scopes,
// resource and PKCE challenge.
func (s *AuthzService) Approve(
	ctx context.Context,
	origin string,
	query url.Values,
	userID string,
) (string, *authzModel.Error) {
	request, user, oauthErr := s.consentRequest(ctx, origin, query, userID)
	if oauthErr != nil {
		return "", oauthErr
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
func (s *AuthzService) Deny(
	ctx context.Context,
	origin string,
	query url.Values,
) (string, *authzModel.Error) {
	request, oauthErr := s.AuthorizationRequest(ctx, origin, query)
	if oauthErr != nil {
		return "", oauthErr
	}
	return redirectWithValues(request.RedirectURI, url.Values{
		"error":             {"access_denied"},
		"error_description": {"the user declined the authorization request"},
		"state":             {request.State},
		"iss":               {s.issuer(ctx, origin)},
	}), nil
}

func (s *AuthzService) consentRequest(
	ctx context.Context,
	origin string,
	query url.Values,
	userID string,
) (authzModel.AuthorizationRequest, userModel.User, *authzModel.Error) {
	request, oauthErr := s.AuthorizationRequest(ctx, origin, query)
	if oauthErr != nil {
		return authzModel.AuthorizationRequest{}, userModel.User{}, oauthErr
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return authzModel.AuthorizationRequest{}, userModel.User{}, authzModel.ErrAccessDenied("approving user not found")
	}
	if !user.IsAdmin {
		return authzModel.AuthorizationRequest{}, userModel.User{}, authzModel.ErrAccessDenied(
			"only administrators can authorize MCP clients")
	}
	if authModel.HasAdminScope(request.Scopes) && !user.IsAdmin {
		return authzModel.AuthorizationRequest{}, userModel.User{}, authzModel.ErrInvalidScope(
			"admin scopes require an administrator")
	}
	return request, user, nil
}

func parseScopes(raw string) ([]string, *authzModel.Error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return append([]string(nil), authzModel.DefaultScopes...), nil
	}
	if len(fields) > 20 {
		return nil, authzModel.ErrInvalidScope("too many scopes requested")
	}
	seen := make(map[string]struct{}, len(fields))
	scopes := make([]string, 0, len(fields))
	for _, scope := range fields {
		if !authModel.IsValidScope(scope) {
			return nil, authzModel.ErrInvalidScope("unknown scope: " + scope)
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	return scopes, nil
}

func validateCodeChallenge(challenge string) *authzModel.Error {
	if len(challenge) < 43 || len(challenge) > 128 {
		return authzModel.ErrInvalidRequest("code_challenge must be 43-128 characters")
	}
	for _, r := range challenge {
		if !isBase64URLRune(r) {
			return authzModel.ErrInvalidRequest("code_challenge must be base64url")
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
