// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/kvstore"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	authmock "github.com/lin-snow/ech0/internal/test/mocks/authmock"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return string(raw)
}

func TestLogin(t *testing.T) {
	helpers.SetJWTSecret(t, "login-secret")

	const username = "alice"
	const plainPassword = "s3cr3t"
	const userID = "u-1"
	md5Hash := cryptoUtil.MD5Encrypt(plainPassword)
	bcryptHash, err := cryptoUtil.HashPassword(plainPassword)
	require.NoError(t, err)

	cases := []struct {
		name      string
		dto       authModel.LoginDto
		setupRepo func(repo *authmock.MockRepository)
		wantErr   string
	}{
		{
			name:      "empty username short-circuits before repo",
			dto:       authModel.LoginDto{Username: "", Password: plainPassword},
			setupRepo: func(*authmock.MockRepository) {},
			wantErr:   commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY,
		},
		{
			name:      "empty password short-circuits before repo",
			dto:       authModel.LoginDto{Username: username, Password: ""},
			setupRepo: func(*authmock.MockRepository) {},
			wantErr:   commonModel.USERNAME_OR_PASSWORD_NOT_BE_EMPTY,
		},
		{
			name: "user not found maps to USER_NOTFOUND",
			dto:  authModel.LoginDto{Username: username, Password: plainPassword},
			setupRepo: func(repo *authmock.MockRepository) {
				repo.EXPECT().
					GetUserByUsername(mock.Anything, username).
					Return(userModel.User{}, errors.New("record not found")).
					Once()
			},
			wantErr: commonModel.USER_NOTFOUND,
		},
		{
			name: "missing local auth row maps to PASSWORD_INCORRECT",
			dto:  authModel.LoginDto{Username: username, Password: plainPassword},
			setupRepo: func(repo *authmock.MockRepository) {
				repo.EXPECT().
					GetUserByUsername(mock.Anything, username).
					Return(userModel.User{ID: userID, Username: username}, nil).
					Once()
				repo.EXPECT().
					GetLocalAuthByUserID(mock.Anything, userID).
					Return(userModel.UserLocalAuth{}, gorm.ErrRecordNotFound).
					Once()
			},
			wantErr: commonModel.PASSWORD_INCORRECT,
		},
		{
			name: "local auth lookup db error maps to PASSWORD_INCORRECT",
			dto:  authModel.LoginDto{Username: username, Password: plainPassword},
			setupRepo: func(repo *authmock.MockRepository) {
				repo.EXPECT().
					GetUserByUsername(mock.Anything, username).
					Return(userModel.User{ID: userID, Username: username}, nil).
					Once()
				repo.EXPECT().
					GetLocalAuthByUserID(mock.Anything, userID).
					Return(userModel.UserLocalAuth{}, errors.New("db down")).
					Once()
			},
			wantErr: commonModel.PASSWORD_INCORRECT,
		},
		{
			name: "wrong password maps to PASSWORD_INCORRECT",
			dto:  authModel.LoginDto{Username: username, Password: plainPassword},
			setupRepo: func(repo *authmock.MockRepository) {
				repo.EXPECT().
					GetUserByUsername(mock.Anything, username).
					Return(userModel.User{ID: userID, Username: username}, nil).
					Once()
				repo.EXPECT().
					GetLocalAuthByUserID(mock.Anything, userID).
					Return(userModel.UserLocalAuth{
						UserID:       userID,
						PasswordHash: cryptoUtil.MD5Encrypt("another-password"),
						PasswordAlgo: cryptoUtil.AlgoMD5,
					}, nil).
					Once()
			},
			wantErr: commonModel.PASSWORD_INCORRECT,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _, _ := newSvc(t, kvstore.NewMemory())
			tc.setupRepo(repo)

			pair, err := svc.Login(&tc.dto)
			require.EqualError(t, err, tc.wantErr)
			assert.Nil(t, pair)
		})
	}

	t.Run("success with bcrypt hash issues token pair without upgrade", func(t *testing.T) {
		svc, repo, _, _ := newSvc(t, kvstore.NewMemory())
		repo.EXPECT().
			GetUserByUsername(mock.Anything, username).
			Return(userModel.User{ID: userID, Username: username}, nil).
			Once()
		repo.EXPECT().
			GetLocalAuthByUserID(mock.Anything, userID).
			Return(userModel.UserLocalAuth{
				UserID:       userID,
				PasswordHash: bcryptHash,
				PasswordAlgo: cryptoUtil.AlgoBcrypt,
			}, nil).
			Once()

		pair, err := svc.Login(&authModel.LoginDto{Username: username, Password: plainPassword})
		require.NoError(t, err)
		require.NotNil(t, pair)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		assert.Equal(t, config.Config().Auth.Jwt.Expires, pair.ExpiresIn)
	})

	t.Run("success with legacy md5 hash triggers lazy upgrade to bcrypt", func(t *testing.T) {
		svc, repo, _, _ := newSvc(t, kvstore.NewMemory())
		repo.EXPECT().
			GetUserByUsername(mock.Anything, username).
			Return(userModel.User{ID: userID, Username: username}, nil).
			Once()
		repo.EXPECT().
			GetLocalAuthByUserID(mock.Anything, userID).
			Return(userModel.UserLocalAuth{
				UserID:       userID,
				PasswordHash: md5Hash,
				PasswordAlgo: cryptoUtil.AlgoMD5,
			}, nil).
			Once()
		repo.EXPECT().
			UpdateLocalAuthPassword(mock.Anything, userID, mock.Anything, cryptoUtil.AlgoBcrypt).
			Return(nil).
			Once()

		pair, err := svc.Login(&authModel.LoginDto{Username: username, Password: plainPassword})
		require.NoError(t, err)
		require.NotNil(t, pair)
		assert.NotEmpty(t, pair.AccessToken)
	})
}

func TestRevokeToken_Delegates(t *testing.T) {
	svc, _, authRepo, _ := newSvc(t, kvstore.NewMemory())

	authRepo.EXPECT().RevokeToken("jti-1", 5*time.Minute).Once()
	svc.RevokeToken("jti-1", 5*time.Minute)
}

func TestIsTokenRevoked_Delegates(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: "revoked", want: true},
		{name: "not revoked", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, authRepo, _ := newSvc(t, kvstore.NewMemory())
			authRepo.EXPECT().IsTokenRevoked("jti-x").Return(tc.want).Once()
			assert.Equal(t, tc.want, svc.IsTokenRevoked("jti-x"))
		})
	}
}

func TestExchangeOAuthCode(t *testing.T) {
	t.Run("hit returns stored pair", func(t *testing.T) {
		svc, _, authRepo, _ := newSvc(t, kvstore.NewMemory())
		want := &authModel.TokenPair{AccessToken: "acc", RefreshToken: "ref", ExpiresIn: 900}
		authRepo.EXPECT().GetAndDeleteOAuthCode("code-1").Return(want, nil).Once()

		got, err := svc.ExchangeOAuthCode("code-1")
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("miss propagates error", func(t *testing.T) {
		svc, _, authRepo, _ := newSvc(t, kvstore.NewMemory())
		sentinel := errors.New("code not found or expired")
		authRepo.EXPECT().GetAndDeleteOAuthCode("missing").Return(nil, sentinel).Once()

		got, err := svc.ExchangeOAuthCode("missing")
		require.ErrorIs(t, err, sentinel)
		assert.Nil(t, got)
	})
}

func TestPasskeyBoundary(t *testing.T) {
	t.Run("returns configured rp id and origins", func(t *testing.T) {
		kv := kvstore.NewMemory()
		raw := mustJSON(t, settingModel.PasskeySetting{
			WebAuthnRPID:           "example.com",
			WebAuthnAllowedOrigins: []string{"https://example.com", "https://app.example.com"},
		})
		require.NoError(t, kv.Set(context.Background(), commonModel.PasskeySettingKey, raw))

		svc, _, _, _ := newSvc(t, kv)
		rpID, origins := svc.PasskeyBoundary(context.Background())
		assert.Equal(t, "example.com", rpID)
		assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, origins)
	})

	t.Run("corrupt setting falls back to empty boundary", func(t *testing.T) {
		kv := kvstore.NewMemory()
		require.NoError(t, kv.Set(context.Background(), commonModel.PasskeySettingKey, "not-valid-json"))

		svc, _, _, _ := newSvc(t, kv)
		rpID, origins := svc.PasskeyBoundary(context.Background())
		assert.Empty(t, rpID)
		assert.Nil(t, origins)
	})
}
