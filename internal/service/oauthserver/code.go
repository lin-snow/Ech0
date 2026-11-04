// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

func (s *OAuthServerService) signCode(
	ctx context.Context,
	origin string,
	request oauthserverModel.AuthorizationRequest,
	user userModel.User,
) (string, *oauthserverModel.Error) {
	now := s.now()
	issuer := s.issuer(ctx, origin)
	claims := oauthserverModel.CodeClaims{
		Type:          oauthserverModel.CodeType,
		UserID:        user.ID,
		Username:      user.Username,
		ClientID:      request.ClientID,
		ClientName:    request.ClientName,
		RedirectURI:   request.RedirectURI,
		Scopes:        request.Scopes,
		Resource:      request.Resource,
		CodeChallenge: request.CodeChallenge,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{issuer},
			ExpiresAt: jwt.NewNumericDate(now.Add(oauthserverModel.CodeTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Minute)),
			ID:        uuidUtil.NewV7(),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		return "", oauthserverModel.ErrInvalidRequest("failed to issue authorization code")
	}
	return signed, nil
}

func (s *OAuthServerService) parseCode(tokenString string) (*oauthserverModel.CodeClaims, error) {
	claims := &oauthserverModel.CodeClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return config.Config().Security.JWTSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid authorization code")
	}
	if claims.Type != oauthserverModel.CodeType {
		return nil, errors.New("not an authorization code")
	}
	if claims.ID == "" || claims.UserID == "" || claims.ClientID == "" || claims.RedirectURI == "" {
		return nil, errors.New("incomplete authorization code")
	}
	return claims, nil
}
