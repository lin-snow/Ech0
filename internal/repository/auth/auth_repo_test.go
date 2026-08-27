// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/cache"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newAuthRepo(t *testing.T) (*AuthRepository, *gorm.DB, cache.ICache[string, any]) {
	t.Helper()
	db := helpers.NewTestDB(t)
	c := helpers.NewTestCache()
	return NewAuthRepository(func() *gorm.DB { return db }, c), db, c
}

func insertUser(t *testing.T, db *gorm.DB, u userModel.User) userModel.User {
	t.Helper()
	require.NoError(t, db.Create(&u).Error)
	return u
}

func countIdentities(t *testing.T, db *gorm.DB, userID string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&userModel.UserExternalIdentity{}).Where("user_id = ?", userID).Count(&n).Error)
	return n
}

func TestAuthRepository_GetUserByUsername(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	insertUser(t, db, helpers.NewUser(func(u *userModel.User) {
		u.ID = "u-name"
		u.Username = "alice"
	}))

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetUserByUsername(context.Background(), "alice")
		require.NoError(t, err)
		assert.Equal(t, "u-name", got.ID)
		assert.Equal(t, "alice", got.Username)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetUserByUsername(context.Background(), "nobody")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestAuthRepository_GetUserByID(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	insertUser(t, db, helpers.NewUser(func(u *userModel.User) {
		u.ID = "u-id"
		u.Username = "bob"
	}))

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetUserByID(context.Background(), "u-id")
		require.NoError(t, err)
		assert.Equal(t, "bob", got.Username)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetUserByID(context.Background(), "missing")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestAuthRepository_BindOAuth_CreateOAuth2(t *testing.T) {
	repo, db, _ := newAuthRepo(t)

	err := repo.BindOAuth(context.Background(), "u-1", "github", "sub-1", "https://ignored", string(authModel.AuthTypeOAuth2))
	require.NoError(t, err)

	var got userModel.UserExternalIdentity
	require.NoError(t, db.Where("user_id = ? AND provider = ?", "u-1", "github").First(&got).Error)
	assert.Equal(t, "sub-1", got.Subject)
	assert.Equal(t, "", got.Issuer, "oauth2 identity must store empty issuer")
	assert.Equal(t, string(authModel.AuthTypeOAuth2), got.Protocol)
	assert.NotEmpty(t, got.ID, "BeforeCreate should fill uuid")
}

func TestAuthRepository_BindOAuth_CreateOIDC(t *testing.T) {
	repo, db, _ := newAuthRepo(t)

	err := repo.BindOAuth(context.Background(), "u-2", "keycloak", "sub-2", "  https://idp.example.com  ", string(authModel.AuthTypeOIDC))
	require.NoError(t, err)

	var got userModel.UserExternalIdentity
	require.NoError(t, db.Where("user_id = ? AND provider = ?", "u-2", "keycloak").First(&got).Error)
	assert.Equal(t, "sub-2", got.Subject)
	assert.Equal(t, "https://idp.example.com", got.Issuer)
	assert.Equal(t, string(authModel.AuthTypeOIDC), got.Protocol)
}

func TestAuthRepository_BindOAuth_UpsertUpdatesSubject(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.BindOAuth(ctx, "u-3", "github", "old-sub", "", string(authModel.AuthTypeOAuth2)))
	require.NoError(t, repo.BindOAuth(ctx, "u-3", "github", "new-sub", "", string(authModel.AuthTypeOAuth2)))

	assert.Equal(t, int64(1), countIdentities(t, db, "u-3"))

	var got userModel.UserExternalIdentity
	require.NoError(t, db.Where("user_id = ?", "u-3").First(&got).Error)
	assert.Equal(t, "new-sub", got.Subject)
}

func TestAuthRepository_BindOAuth_OAuth2AndOIDCAreDistinct(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.BindOAuth(ctx, "u-4", "google", "sub-oauth2", "", string(authModel.AuthTypeOAuth2)))
	require.NoError(t, repo.BindOAuth(ctx, "u-4", "google", "sub-oidc", "https://accounts.google.com", string(authModel.AuthTypeOIDC)))

	assert.Equal(t, int64(2), countIdentities(t, db, "u-4"))
}

func TestAuthRepository_GetUserByOAuthID(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	ctx := context.Background()
	insertUser(t, db, helpers.NewUser(func(u *userModel.User) {
		u.ID = "owner-oauth2"
		u.Username = "oauth2user"
	}))
	require.NoError(t, repo.BindOAuth(ctx, "owner-oauth2", "github", "gh-123", "", string(authModel.AuthTypeOAuth2)))

	t.Run("resolves user by oauth2 binding", func(t *testing.T) {
		got, err := repo.GetUserByOAuthID(ctx, "github", "gh-123")
		require.NoError(t, err)
		assert.Equal(t, "owner-oauth2", got.ID)
		assert.Equal(t, "oauth2user", got.Username)
	})

	t.Run("does not match oidc binding", func(t *testing.T) {
		require.NoError(t, repo.BindOAuth(ctx, "owner-oauth2", "keycloak", "kc-1", "https://idp", string(authModel.AuthTypeOIDC)))
		_, err := repo.GetUserByOAuthID(ctx, "keycloak", "kc-1")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("unknown subject", func(t *testing.T) {
		_, err := repo.GetUserByOAuthID(ctx, "github", "nope")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestAuthRepository_GetUserByOIDC(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	ctx := context.Background()
	insertUser(t, db, helpers.NewUser(func(u *userModel.User) {
		u.ID = "owner-oidc"
		u.Username = "oidcuser"
	}))
	require.NoError(t, repo.BindOAuth(ctx, "owner-oidc", "keycloak", "kc-sub", "https://idp.example.com", string(authModel.AuthTypeOIDC)))

	t.Run("resolves user by issuer+subject", func(t *testing.T) {
		got, err := repo.GetUserByOIDC(ctx, "keycloak", "kc-sub", "https://idp.example.com")
		require.NoError(t, err)
		assert.Equal(t, "owner-oidc", got.ID)
	})

	t.Run("wrong issuer misses", func(t *testing.T) {
		_, err := repo.GetUserByOIDC(ctx, "keycloak", "kc-sub", "https://other.example.com")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestAuthRepository_GetOAuthInfo(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.BindOAuth(ctx, "u-info", "github", "gh-sub", "", string(authModel.AuthTypeOAuth2)))

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetOAuthInfo("u-info", "github")
		require.NoError(t, err)
		assert.Equal(t, "gh-sub", got.Subject)
		assert.Equal(t, string(authModel.AuthTypeOAuth2), got.Protocol)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetOAuthInfo("u-info", "gitlab")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestAuthRepository_GetOAuthOIDCInfo(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.BindOAuth(ctx, "u-oidc-info", "keycloak", "kc-sub", "https://idp.example.com", string(authModel.AuthTypeOIDC)))

	t.Run("found by issuer", func(t *testing.T) {
		got, err := repo.GetOAuthOIDCInfo("u-oidc-info", "keycloak", "https://idp.example.com")
		require.NoError(t, err)
		assert.Equal(t, "kc-sub", got.Subject)
		assert.Equal(t, string(authModel.AuthTypeOIDC), got.Protocol)
	})

	t.Run("wrong issuer not found", func(t *testing.T) {
		_, err := repo.GetOAuthOIDCInfo("u-oidc-info", "keycloak", "https://wrong")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func newPasskey(opts ...func(*authModel.Passkey)) *authModel.Passkey {
	p := &authModel.Passkey{
		ID:             "pk-1",
		UserID:         "u-pk",
		CredentialID:   "cred-1",
		CredentialJSON: `{"id":"cred-1"}`,
		PublicKey:      "pubkey-b64",
		SignCount:      3,
		LastUsedAt:     100,
		DeviceName:     "MacBook",
		AAGUID:         "aaguid-xyz",
		CreatedAt:      10,
		UpdatedAt:      20,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func TestAuthRepository_CreatePasskey_RoundTrip(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()

	in := newPasskey()
	require.NoError(t, repo.CreatePasskey(ctx, in))

	got, err := repo.GetPasskeyByCredentialID("cred-1")
	require.NoError(t, err)
	assert.Equal(t, in.ID, got.ID)
	assert.Equal(t, in.UserID, got.UserID)
	assert.Equal(t, in.CredentialID, got.CredentialID)
	assert.Equal(t, in.CredentialJSON, got.CredentialJSON)
	assert.Equal(t, in.PublicKey, got.PublicKey)
	assert.Equal(t, in.SignCount, got.SignCount)
	assert.Equal(t, in.LastUsedAt, got.LastUsedAt)
	assert.Equal(t, in.DeviceName, got.DeviceName)
	assert.Equal(t, in.AAGUID, got.AAGUID)
}

func TestAuthRepository_GetPasskeyByCredentialID_NotFound(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	_, err := repo.GetPasskeyByCredentialID("does-not-exist")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestAuthRepository_ListPasskeysByUserID(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.CreatePasskey(ctx, newPasskey(func(p *authModel.Passkey) {
		p.ID = "pk-a"
		p.CredentialID = "cred-a"
	})))
	require.NoError(t, repo.CreatePasskey(ctx, newPasskey(func(p *authModel.Passkey) {
		p.ID = "pk-b"
		p.CredentialID = "cred-b"
	})))
	require.NoError(t, repo.CreatePasskey(ctx, newPasskey(func(p *authModel.Passkey) {
		p.ID = "pk-other"
		p.UserID = "u-other"
		p.CredentialID = "cred-other"
	})))

	list, err := repo.ListPasskeysByUserID("u-pk")
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "pk-b", list[0].ID)
	assert.Equal(t, "pk-a", list[1].ID)
}

func TestAuthRepository_ListPasskeysByUserID_Empty(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	list, err := repo.ListPasskeysByUserID("nobody")
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestAuthRepository_UpdatePasskeyUsage(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.CreatePasskey(ctx, newPasskey()))

	require.NoError(t, repo.UpdatePasskeyUsage(ctx, "pk-1", 42, 9999))

	got, err := repo.GetPasskeyByCredentialID("cred-1")
	require.NoError(t, err)
	assert.Equal(t, uint32(42), got.SignCount)
	assert.Equal(t, int64(9999), got.LastUsedAt)
}

func TestAuthRepository_UpdatePasskeyDeviceName(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.CreatePasskey(ctx, newPasskey()))

	t.Run("owner can rename", func(t *testing.T) {
		require.NoError(t, repo.UpdatePasskeyDeviceName(ctx, "u-pk", "pk-1", "iPhone"))
		got, err := repo.GetPasskeyByCredentialID("cred-1")
		require.NoError(t, err)
		assert.Equal(t, "iPhone", got.DeviceName)
	})

	t.Run("non-owner is a no-op", func(t *testing.T) {
		require.NoError(t, repo.UpdatePasskeyDeviceName(ctx, "u-intruder", "pk-1", "hacked"))
		got, err := repo.GetPasskeyByCredentialID("cred-1")
		require.NoError(t, err)
		assert.Equal(t, "iPhone", got.DeviceName)
	})
}

func TestAuthRepository_DeletePasskeyByID(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.CreatePasskey(ctx, newPasskey()))

	t.Run("non-owner cannot delete", func(t *testing.T) {
		require.NoError(t, repo.DeletePasskeyByID(ctx, "u-intruder", "pk-1"))
		_, err := repo.GetPasskeyByCredentialID("cred-1")
		require.NoError(t, err, "passkey should survive a foreign-user delete")
	})

	t.Run("owner deletes", func(t *testing.T) {
		require.NoError(t, repo.DeletePasskeyByID(ctx, "u-pk", "pk-1"))
		_, err := repo.GetPasskeyByCredentialID("cred-1")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestAuthRepository_RevokeToken(t *testing.T) {
	repo, _, _ := newAuthRepo(t)

	t.Run("revoked token is reported", func(t *testing.T) {
		repo.RevokeToken("jti-1", time.Minute)
		assert.True(t, repo.IsTokenRevoked("jti-1"))
	})

	t.Run("unknown jti is not revoked", func(t *testing.T) {
		assert.False(t, repo.IsTokenRevoked("jti-unknown"))
	})

	t.Run("empty jti never revokes or reports", func(t *testing.T) {
		repo.RevokeToken("", time.Minute)
		assert.False(t, repo.IsTokenRevoked(""))
	})

	t.Run("non-positive ttl is a no-op", func(t *testing.T) {
		repo.RevokeToken("jti-zero", 0)
		assert.False(t, repo.IsTokenRevoked("jti-zero"))
	})
}

func TestAuthRepository_OAuthCode_SingleUse(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	pair := &authModel.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}

	repo.StoreOAuthCode("code-1", pair, time.Minute)

	got, err := repo.GetAndDeleteOAuthCode("code-1")
	require.NoError(t, err)
	assert.Equal(t, "at", got.AccessToken)
	assert.Equal(t, "rt", got.RefreshToken)

	_, err = repo.GetAndDeleteOAuthCode("code-1")
	require.EqualError(t, err, commonModel.EXCHANGE_CODE_INVALID)
}

func TestAuthRepository_GetAndDeleteOAuthCode_Errors(t *testing.T) {
	repo, _, c := newAuthRepo(t)

	t.Run("empty code", func(t *testing.T) {
		_, err := repo.GetAndDeleteOAuthCode("")
		require.EqualError(t, err, commonModel.EXCHANGE_CODE_INVALID)
	})

	t.Run("missing code", func(t *testing.T) {
		_, err := repo.GetAndDeleteOAuthCode("ghost")
		require.EqualError(t, err, commonModel.EXCHANGE_CODE_INVALID)
	})

	t.Run("type mismatch is rejected and key consumed", func(t *testing.T) {
		c.SetWithTTL(oauthCodePrefix+"weird", "not-a-pair", 1, time.Minute)
		_, err := repo.GetAndDeleteOAuthCode("weird")
		require.EqualError(t, err, commonModel.EXCHANGE_CODE_INVALID)
		_, found, _ := c.Get(oauthCodePrefix + "weird")
		assert.False(t, found)
	})
}

func TestAuthRepository_StoreOAuthCode_Guards(t *testing.T) {
	repo, _, _ := newAuthRepo(t)
	pair := &authModel.TokenPair{AccessToken: "at"}

	cases := []struct {
		name string
		code string
		pair *authModel.TokenPair
		ttl  time.Duration
	}{
		{"empty code", "", pair, time.Minute},
		{"nil pair", "c", nil, time.Minute},
		{"non-positive ttl", "c", pair, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo.StoreOAuthCode(tc.code, tc.pair, tc.ttl)
			key := tc.code
			if key == "" {
				key = "c"
			}
			_, err := repo.GetAndDeleteOAuthCode(key)
			require.EqualError(t, err, commonModel.EXCHANGE_CODE_INVALID)
		})
	}
}

func TestAuthRepository_PasskeySession(t *testing.T) {
	repo, _, _ := newAuthRepo(t)

	t.Run("miss returns ErrRecordNotFound", func(t *testing.T) {
		_, err := repo.CacheGetPasskeySession("sess-missing")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("set then get", func(t *testing.T) {
		repo.CacheSetPasskeySession("sess-1", "payload", time.Minute)
		got, err := repo.CacheGetPasskeySession("sess-1")
		require.NoError(t, err)
		assert.Equal(t, "payload", got)
	})

	t.Run("delete removes it", func(t *testing.T) {
		repo.CacheSetPasskeySession("sess-2", 42, time.Minute)
		repo.CacheDeletePasskeySession("sess-2")
		_, err := repo.CacheGetPasskeySession("sess-2")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}
