// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccessClaimsWithExpiry_AudienceOverride(t *testing.T) {
	user := userModel.User{ID: "u-aud", Username: "aud"}

	t.Run("explicit-audience-overrides", func(t *testing.T) {
		claimsAny := CreateAccessClaimsWithExpiry(user, 3600, []string{authModel.ScopeEchoRead}, authModel.AudienceCLI, "jti-aud")
		claims, ok := claimsAny.(authModel.MyClaims)
		require.True(t, ok, "unexpected claims type %T", claimsAny)

		assert.Equal(t, jwt.ClaimStrings{authModel.AudienceCLI}, claims.Audience)
		assert.Equal(t, authModel.TokenTypeAccess, claims.Type)
		assert.Equal(t, []string{authModel.ScopeEchoRead}, claims.Scopes)
		assert.Equal(t, "jti-aud", claims.ID)
	})

	t.Run("empty-audience-falls-back-to-config", func(t *testing.T) {
		want := config.Config().Auth.Jwt.Audience
		claimsAny := CreateAccessClaimsWithExpiry(user, 3600, nil, "", "jti-default")
		claims, ok := claimsAny.(authModel.MyClaims)
		require.True(t, ok, "unexpected claims type %T", claimsAny)

		assert.Equal(t, jwt.ClaimStrings{want}, claims.Audience)
	})
}

func TestCreateAccessClaimsWithExpiry_NegativeExpiryGetsFiniteExp(t *testing.T) {
	user := userModel.User{ID: "u-neg", Username: "neg"}
	claimsAny := CreateAccessClaimsWithExpiry(user, -5, []string{authModel.ScopeProfileRead}, authModel.AudienceCLI, "jti-neg")
	claims, ok := claimsAny.(authModel.MyClaims)
	require.True(t, ok, "unexpected claims type %T", claimsAny)

	require.NotNil(t, claims.ExpiresAt, "negative expiry must still set ExpiresAt")
	atLeast := time.Now().UTC().Add(50 * 365 * 24 * time.Hour)
	assert.True(t, claims.ExpiresAt.After(atLeast), "expected far-future ExpiresAt, got %v", claims.ExpiresAt.Time)
}

func TestParseRefreshToken_TypeEnforcement(t *testing.T) {
	helpers.SetJWTSecret(t, "jwt-gaps-test-secret-0123456789ab")
	user := userModel.User{ID: "u-r", Username: "r"}

	t.Run("rejects-access-typ", func(t *testing.T) {
		accessClaims := CreateAccessClaimsWithExpiry(user, 3600, nil, authModel.AudienceCLI, "jti-acc")
		tokenStr, err := GenerateToken(accessClaims)
		require.NoError(t, err)

		_, err = ParseRefreshToken(tokenStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected refresh")
	})

	t.Run("rejects-session-typ", func(t *testing.T) {
		sessionClaims := CreateClaims(user)
		tokenStr, err := GenerateToken(sessionClaims)
		require.NoError(t, err)

		_, err = ParseRefreshToken(tokenStr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected refresh")
	})

	t.Run("accepts-refresh-typ", func(t *testing.T) {
		refreshClaims := CreateRefreshClaims(user)
		tokenStr, err := GenerateToken(refreshClaims)
		require.NoError(t, err)

		claims, err := ParseRefreshToken(tokenStr)
		require.NoError(t, err)
		assert.Equal(t, authModel.TokenTypeRefresh, claims.Type)
		assert.Equal(t, user.ID, claims.Userid)
	})
}

func signOAuthState(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(config.Config().Security.JWTSecret)
	require.NoError(t, err)
	return s
}

func validStateClaims() jwt.MapClaims {
	now := time.Now().UTC()
	return jwt.MapClaims{
		"action":   "login",
		"user_id":  "u1",
		"nonce":    "nonce-1",
		"redirect": "https://example.com/auth",
		"provider": "github",
		"exp":      now.Add(10 * time.Minute).Unix(),
		"iat":      now.Unix(),
	}
}

func TestParseOAuthState_MissingOrInvalidFields(t *testing.T) {
	helpers.SetJWTSecret(t, "jwt-gaps-test-secret-0123456789ab")

	cases := []struct {
		name      string
		mutate    func(jwt.MapClaims)
		wantInErr string
	}{
		{
			name:      "missing-action",
			mutate:    func(c jwt.MapClaims) { delete(c, "action") },
			wantInErr: "action",
		},
		{
			name:      "missing-nonce",
			mutate:    func(c jwt.MapClaims) { delete(c, "nonce") },
			wantInErr: "nonce",
		},
		{
			name:      "missing-redirect",
			mutate:    func(c jwt.MapClaims) { delete(c, "redirect") },
			wantInErr: "redirect",
		},
		{
			name:      "missing-provider",
			mutate:    func(c jwt.MapClaims) { delete(c, "provider") },
			wantInErr: "provider",
		},
		{
			name:      "empty-redirect",
			mutate:    func(c jwt.MapClaims) { c["redirect"] = "" },
			wantInErr: "redirect",
		},
		{
			name:      "non-string-provider",
			mutate:    func(c jwt.MapClaims) { c["provider"] = 123 },
			wantInErr: "provider",
		},
		{
			name:      "missing-exp",
			mutate:    func(c jwt.MapClaims) { delete(c, "exp") },
			wantInErr: "exp",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := validStateClaims()
			tc.mutate(claims)
			state := signOAuthState(t, claims)

			parsed, err := ParseOAuthState(state)
			require.Error(t, err)
			assert.Nil(t, parsed)
			assert.Contains(t, err.Error(), tc.wantInErr)
		})
	}
}

func TestParseOAuthState_BadSignature(t *testing.T) {
	helpers.SetJWTSecret(t, "jwt-gaps-test-secret-0123456789ab")

	parsed, err := ParseOAuthState("not-a-valid-jwt-token")
	require.Error(t, err)
	assert.Nil(t, parsed)
}
