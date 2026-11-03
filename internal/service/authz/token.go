// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/url"
	"strings"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authzModel "github.com/lin-snow/ech0/internal/model/authz"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// Exchange redeems an authorization code. PKCE is verified before the code is
// consumed, and consumption happens before minting so a replay can never yield
// a second token.
func (s *AuthzService) Exchange(
	ctx context.Context,
	origin string,
	form url.Values,
) (authzModel.TokenResponse, *authzModel.Error) {
	if grant := form.Get("grant_type"); grant != "authorization_code" {
		return authzModel.TokenResponse{}, authzModel.ErrUnsupportedGrantType(
			"only authorization_code is supported")
	}

	code := strings.TrimSpace(form.Get("code"))
	clientID := strings.TrimSpace(form.Get("client_id"))
	redirectURI := strings.TrimSpace(form.Get("redirect_uri"))
	verifier := strings.TrimSpace(form.Get("code_verifier"))
	if code == "" || clientID == "" || redirectURI == "" || verifier == "" {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidRequest(
			"code, client_id, redirect_uri and code_verifier are required")
	}
	if err := validateCodeVerifier(verifier); err != nil {
		return authzModel.TokenResponse{}, err
	}

	claims, err := s.parseCode(code)
	if err != nil {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("authorization code is invalid or expired")
	}

	issuer := s.issuer(ctx, origin)
	if claims.Issuer != issuer {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("authorization code was issued by another server")
	}
	if claims.ClientID != clientID {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("client_id does not match the authorization code")
	}
	if claims.RedirectURI != redirectURI {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("redirect_uri does not match the authorization code")
	}
	if claims.Resource != s.resourceURI(ctx, origin) {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("authorization code is bound to another resource")
	}
	if resource := strings.TrimRight(strings.TrimSpace(form.Get("resource")), "/"); resource != "" &&
		resource != claims.Resource {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("resource does not match the authorization code")
	}
	if !verifyPKCE(verifier, claims.CodeChallenge) {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("PKCE verification failed")
	}
	if !s.consumeCode(claims.ID) {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("authorization code has already been used")
	}

	token, err := s.tokens.IssueAccessTokenForUser(ctx, claims.UserID, &settingModel.AccessTokenSettingDto{
		Name:     tokenName(claims.ClientName),
		Expiry:   settingModel.ONE_MONTH_EXPIRY,
		Scopes:   claims.Scopes,
		Audience: authModel.AudienceMCPRemote,
	})
	if err != nil {
		return authzModel.TokenResponse{}, authzModel.ErrInvalidGrant("failed to issue access token")
	}

	return authzModel.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(authzModel.AccessTokenTTL.Seconds()),
		Scope:       strings.Join(claims.Scopes, " "),
	}, nil
}

func validateCodeVerifier(verifier string) *authzModel.Error {
	if len(verifier) < 43 || len(verifier) > 128 {
		return authzModel.ErrInvalidGrant("code_verifier must be 43-128 characters")
	}
	for _, r := range verifier {
		if !isBase64URLRune(r) || r == '=' {
			return authzModel.ErrInvalidGrant("code_verifier must be base64url")
		}
	}
	return nil
}

func verifyPKCE(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

func tokenName(clientName string) string {
	name := "mcp:" + strings.TrimSpace(clientName)
	if name == "mcp:" {
		name = "mcp:client"
	}
	if runes := []rune(name); len(runes) > 120 {
		name = string(runes[:120])
	}
	return name
}
