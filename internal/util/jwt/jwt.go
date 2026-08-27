// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

func CreateClaims(user userModel.User) jwt.Claims {
	leeway := time.Second * 60
	now := time.Now().UTC()
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		Type:     authModel.TokenTypeSession,
		Issuer:   config.Config().Auth.Jwt.Issuer,
		Subject:  user.Username,
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
	leeway := time.Second * 60
	now := time.Now().UTC()
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		Type:     authModel.TokenTypeRefresh,
		Issuer:   config.Config().Auth.Jwt.Issuer,
		Subject:  user.Username,
		Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
		ExpiresAt: jwt.NewNumericDate(
			now.Add(time.Duration(config.Config().Auth.Jwt.RefreshExpires) * time.Second),
		),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-leeway)),
		ID:        cryptoUtil.GenerateRandomString(16),
	}

	return claims
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
		Userid:    user.ID,
		Username:  user.Username,
		Type:      authModel.TokenTypeAccess,
		Scopes:    scopes,
		Issuer:    config.Config().Auth.Jwt.Issuer,
		Subject:   user.Username,
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

func ParseToken(tokenString string) (*authModel.MyClaims, error) {
	claims, err := parseTokenRaw(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != authModel.TokenTypeSession && claims.Type != authModel.TokenTypeAccess {
		return nil, errors.New("invalid token typ")
	}
	return claims, nil
}

func ParseRefreshToken(tokenString string) (*authModel.MyClaims, error) {
	claims, err := parseTokenRaw(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != authModel.TokenTypeRefresh {
		return nil, errors.New("invalid token typ: expected refresh")
	}
	return claims, nil
}

func parseTokenRaw(tokenString string) (*authModel.MyClaims, error) {
	claims := &authModel.MyClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return config.Config().Security.JWTSecret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*authModel.MyClaims); ok {
		return claims, nil
	}

	logUtil.Warn("parse token claims type mismatch", slog.String("module", "jwt"))
	return nil, errors.New("unknown claims type, cannot proceed")
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
			return "", fmt.Errorf("oauth state 缺少 %s", key)
		}
		s, ok := v.(string)
		if !ok || s == "" {
			return "", fmt.Errorf("oauth state %s 非法", key)
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
		return nil, errors.New("oauth state 缺少 exp")
	}
	expFloat, ok := expRaw.(float64)
	if !ok {
		return nil, errors.New("oauth state exp 非法")
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
		return nil, errors.New("id_token 为空")
	}
	if issuer == "" {
		return nil, errors.New("OIDC issuer 为空")
	}
	if jwksURL == "" {
		return nil, errors.New("JWKS URL 为空")
	}
	if clientID == "" {
		return nil, errors.New("OIDC client_id 为空")
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
			return nil, errors.New("id_token 缺少 nonce 声明")
		}
		if idTokenObj.Nonce != expectedNonce {
			return nil, errors.New("id_token nonce 不匹配")
		}
		claims["nonce"] = idTokenObj.Nonce
	}

	subVal, ok := claims["sub"]
	if !ok {
		return nil, errors.New("id_token 缺少 sub 声明")
	}

	switch v := subVal.(type) {
	case string:
		if v == "" {
			return nil, errors.New("id_token sub 为空")
		}
		claims["sub"] = v
	default:
		subStr := fmt.Sprint(v)
		if subStr == "" || subStr == "<nil>" {
			return nil, errors.New("id_token sub 无法转换为字符串")
		}
		claims["sub"] = subStr
	}

	return claims, nil
}
