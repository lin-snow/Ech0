// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
)

func CreateClaims(user userModel.User) jwt.Claims {
	leeway := time.Second * 60
	now := time.Now().UTC()
	claims := authModel.MyClaims{
		Type:     authModel.TokenTypeSession,
		Issuer:   config.Config().Auth.Jwt.Issuer,
		Subject:  user.ID,
		Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
		ExpiresAt: jwt.NewNumericDate(
			now.Add(time.Duration(config.Config().Auth.Jwt.Expires) * time.Second),
		),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-leeway)),
		ID:        cryptoUtil.GenerateRandomString(16),
	}

	return claims
}

func CreateRefreshClaims(user userModel.User) jwt.Claims {
	return buildRefreshClaims(user, cryptoUtil.GenerateRandomString(16))
}

// RenewRefreshClaims re-issues refresh claims for an existing session, keeping its
// JTI so revocation still targets the same session, while sliding ExpiresAt forward.
func RenewRefreshClaims(user userModel.User, jti string) jwt.Claims {
	if jti == "" {
		jti = cryptoUtil.GenerateRandomString(16)
	}
	return buildRefreshClaims(user, jti)
}

func buildRefreshClaims(user userModel.User, jti string) authModel.MyClaims {
	leeway := time.Second * 60
	now := time.Now().UTC()
	return authModel.MyClaims{
		Type:     authModel.TokenTypeRefresh,
		Issuer:   config.Config().Auth.Jwt.Issuer,
		Subject:  user.ID,
		Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
		ExpiresAt: jwt.NewNumericDate(
			now.Add(time.Duration(config.Config().Auth.Jwt.RefreshExpires) * time.Second),
		),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-leeway)),
		ID:        jti,
	}
}

func CreateClaimsWithExpiry(user userModel.User, expiry int64) jwt.Claims {
	return CreateAccessClaimsWithExpiry(user, expiry, nil, "", "")
}

func CreateAccessClaimsWithExpiry(
	user userModel.User,
	expiry int64,
	scopes []string,
	audience string,
	jti string,
) jwt.Claims {
	leeway := time.Second * 60
	audiences := jwt.ClaimStrings{config.Config().Auth.Jwt.Audience}
	if audience != "" {
		audiences = jwt.ClaimStrings{audience}
	}
	const neverExpiryFallback = int64(100 * 365 * 24 * 3600)
	if expiry <= 0 {
		expiry = neverExpiryFallback
	}

	claims := authModel.MyClaims{
		Type:      authModel.TokenTypeAccess,
		Scopes:    scopes,
		Issuer:    config.Config().Auth.Jwt.Issuer,
		Subject:   user.ID,
		Audience:  audiences,
		ID:        jti,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiry) * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		NotBefore: jwt.NewNumericDate(time.Now().UTC().Add(-leeway)),
	}

	return claims
}

func GenerateToken(claim jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(config.Config().Security.JWTSecret)
}

// ParseSessionToken accepts only a session token minted by this instance for
// its configured audience.
func ParseSessionToken(tokenString string) (*authModel.MyClaims, error) {
	return parseToken(tokenString, authModel.TokenTypeSession, true)
}

// ParseAccessToken accepts only an access token. Its audience names the
// resource it was minted for, so the audience check belongs to the endpoint
// consuming the token; this parser only requires that one is present.
func ParseAccessToken(tokenString string) (*authModel.MyClaims, error) {
	claims, err := parseToken(tokenString, authModel.TokenTypeAccess, false)
	if err != nil {
		return nil, err
	}
	if len(claims.Audience) == 0 {
		return nil, errors.New("access token carries no audience")
	}
	return claims, nil
}

func ParseRefreshToken(tokenString string) (*authModel.MyClaims, error) {
	return parseToken(tokenString, authModel.TokenTypeRefresh, true)
}

// parseToken is the single verification path: HS256 only (no algorithm
// confusion), issuer pinned, expiry required, and the instance audience checked
// for every token type except access tokens (which carry their own).
func parseToken(tokenString, typ string, instanceAudience bool) (*authModel.MyClaims, error) {
	cfg := config.Config().Auth.Jwt
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30 * time.Second),
	}
	if instanceAudience {
		options = append(options, jwt.WithAudience(cfg.Audience))
	}

	claims := &authModel.MyClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(*jwt.Token) (any, error) {
			return config.Config().Security.JWTSecret, nil
		},
		options...,
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}
	if claims.Type != typ {
		return nil, fmt.Errorf("unexpected token type %q", claims.Type)
	}
	if claims.Subject == "" {
		return nil, errors.New("token has no subject")
	}
	return claims, nil
}

func GenerateOAuthState(
	action string,
	userID string,
	redirect, provider string,
) (string, string, error) {
	now := time.Now().UTC()
	expiration := now.Add(10 * time.Minute)

	nonce := cryptoUtil.GenerateRandomString(16)

	claims := jwt.MapClaims{
		"action":   action,
		"user_id":  userID,
		"nonce":    nonce,
		"redirect": redirect,
		"exp":      expiration.Unix(),
		"iat":      now.Unix(),
		"provider": provider,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		return "", "", err
	}

	return state, nonce, nil
}

func ParseOAuthState(stateStr string) (*authModel.OAuthState, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(stateStr, claims, func(token *jwt.Token) (any, error) {
		return config.Config().Security.JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}

	getStringClaim := func(key string) (string, error) {
		v, ok := claims[key]
		if !ok {
			return "", fmt.Errorf("oauth state is missing %s", key)
		}
		s, ok := v.(string)
		if !ok || s == "" {
			return "", fmt.Errorf("oauth state %s is invalid", key)
		}
		return s, nil
	}

	action, err := getStringClaim("action")
	if err != nil {
		return nil, err
	}
	nonce, err := getStringClaim("nonce")
	if err != nil {
		return nil, err
	}
	redirect, err := getStringClaim("redirect")
	if err != nil {
		return nil, err
	}
	provider, err := getStringClaim("provider")
	if err != nil {
		return nil, err
	}

	expRaw, ok := claims["exp"]
	if !ok {
		return nil, errors.New("oauth state is missing exp")
	}
	expFloat, ok := expRaw.(float64)
	if !ok {
		return nil, errors.New("oauth state exp is invalid")
	}

	return &authModel.OAuthState{
		Action:   action,
		UserID:   fmt.Sprint(claims["user_id"]),
		Nonce:    nonce,
		Redirect: redirect,
		Exp:      int64(expFloat),
		Provider: provider,
	}, nil
}

func ParseAndVerifyIDToken(idToken, issuer, jwksURL, clientID, expectedNonce string) (jwt.MapClaims, error) {
	if idToken == "" {
		return nil, errors.New("id_token is empty")
	}
	if issuer == "" {
		return nil, errors.New("OIDC issuer is empty")
	}
	if jwksURL == "" {
		return nil, errors.New("JWKS url is empty")
	}
	if clientID == "" {
		return nil, errors.New("OIDC client_id is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	remoteKeySet := oidc.NewRemoteKeySet(ctx, jwksURL)
	verifier := oidc.NewVerifier(issuer, remoteKeySet, &oidc.Config{ClientID: clientID})
	idTokenObj, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{}
	if err := idTokenObj.Claims(&claims); err != nil {
		return nil, err
	}

	if expectedNonce != "" {
		if idTokenObj.Nonce == "" {
			return nil, errors.New("id_token is missing the nonce claim")
		}
		if idTokenObj.Nonce != expectedNonce {
			return nil, errors.New("id_token nonce mismatch")
		}
		claims["nonce"] = idTokenObj.Nonce
	}

	subVal, ok := claims["sub"]
	if !ok {
		return nil, errors.New("id_token is missing the sub claim")
	}

	switch v := subVal.(type) {
	case string:
		if v == "" {
			return nil, errors.New("id_token sub is empty")
		}
		claims["sub"] = v
	default:
		subStr := fmt.Sprint(v)
		if subStr == "" || subStr == "<nil>" {
			return nil, errors.New("id_token sub is not a string")
		}
		claims["sub"] = subStr
	}

	return claims, nil
}
