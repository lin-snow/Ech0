// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"errors"
	"net/url"
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildOAuthAuthorizeURL(t *testing.T) {
	svc := &AuthService{}

	t.Run("github carries standard authorize params", func(t *testing.T) {
		setting := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
		raw := svc.buildOAuthAuthorizeURL(&setting, string(commonModel.OAuth2GITHUB), "state-gh", "")
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.Equal(t, "idp.example.com", u.Host)
		q := u.Query()
		assert.Equal(t, setting.ClientID, q.Get("client_id"))
		assert.Equal(t, setting.RedirectURI, q.Get("redirect_uri"))
		assert.Equal(t, "code", q.Get("response_type"))
		assert.Equal(t, "state-gh", q.Get("state"))
		assert.Contains(t, q.Get("scope"), "read:user")
	})

	t.Run("google forces offline access and consent", func(t *testing.T) {
		setting := fullOAuth2Setting(string(commonModel.OAuth2GOOGLE))
		raw := svc.buildOAuthAuthorizeURL(&setting, string(commonModel.OAuth2GOOGLE), "state-g", "")
		u, err := url.Parse(raw)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "offline", q.Get("access_type"))
		assert.Equal(t, "consent", q.Get("prompt"))
		assert.Equal(t, "state-g", q.Get("state"))
	})

	t.Run("qq builds manual query with display=pc", func(t *testing.T) {
		setting := fullOAuth2Setting(string(commonModel.OAuth2QQ))
		raw := svc.buildOAuthAuthorizeURL(&setting, string(commonModel.OAuth2QQ), "state-qq", "")
		u, err := url.Parse(raw)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "code", q.Get("response_type"))
		assert.Equal(t, setting.ClientID, q.Get("client_id"))
		assert.Equal(t, setting.RedirectURI, q.Get("redirect_uri"))
		assert.Equal(t, "state-qq", q.Get("state"))
		assert.Equal(t, "pc", q.Get("display"))
		assert.Equal(t, "read:user", q.Get("scope"))
	})

	t.Run("custom non-oidc omits nonce param", func(t *testing.T) {
		setting := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
		setting.IsOIDC = false
		raw := svc.buildOAuthAuthorizeURL(&setting, string(commonModel.OAuth2CUSTOM), "state-c", "nonce-ignored")
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.Equal(t, "state-c", u.Query().Get("state"))
		assert.Empty(t, u.Query().Get("nonce"))
	})

	t.Run("custom oidc appends nonce param", func(t *testing.T) {
		setting := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
		setting.IsOIDC = true
		raw := svc.buildOAuthAuthorizeURL(&setting, string(commonModel.OAuth2CUSTOM), "state-c", "nonce-123")
		u, err := url.Parse(raw)
		require.NoError(t, err)
		assert.Equal(t, "nonce-123", u.Query().Get("nonce"))
	})

	t.Run("unknown provider returns empty string", func(t *testing.T) {
		setting := fullOAuth2Setting("github")
		assert.Empty(t, svc.buildOAuthAuthorizeURL(&setting, "unknown-provider", "s", ""))
	})
}

func TestGetOAuthLoginURL(t *testing.T) {
	helpers.SetJWTSecret(t, "login-url-secret")

	t.Run("provider not configured returns OAUTH2_NOT_CONFIGURED", func(t *testing.T) {
		svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
		out, err := svc.GetOAuthLoginURL(string(commonModel.OAuth2GOOGLE), "")
		require.EqualError(t, err, commonModel.OAUTH2_NOT_CONFIGURED)
		assert.Empty(t, out)
	})

	t.Run("invalid client redirect is rejected before state issuance", func(t *testing.T) {
		svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
		out, err := svc.GetOAuthLoginURL(string(commonModel.OAuth2GITHUB), "/relative/path")
		require.EqualError(t, err, commonModel.INVALID_PARAMS)
		assert.Empty(t, out)
	})

	t.Run("success builds an authorize url with state", func(t *testing.T) {
		svc, _, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
		out, err := svc.GetOAuthLoginURL(string(commonModel.OAuth2GITHUB), "")
		require.NoError(t, err)
		u, perr := url.Parse(out)
		require.NoError(t, perr)
		assert.Equal(t, "idp.example.com", u.Host)
		assert.NotEmpty(t, u.Query().Get("state"))
		assert.Equal(t, "code", u.Query().Get("response_type"))
	})
}

func TestGetOAuthInfo_NonAdminRejected(t *testing.T) {
	ctx := helpers.CtxAsUser("u-1")
	svc, repo, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
	repo.EXPECT().
		GetUserByID(mock.Anything, "u-1").
		Return(userModel.User{ID: "u-1", IsAdmin: false}, nil).
		Once()

	out, err := svc.GetOAuthInfo(ctx, string(commonModel.OAuth2GITHUB))
	require.EqualError(t, err, commonModel.NO_PERMISSION_BINDING_GITHUB)
	assert.Equal(t, userModel.OAuthInfoDto{}, out)
}

func TestGetOAuthInfo_GetUserError(t *testing.T) {
	ctx := helpers.CtxAsUser("u-1")
	svc, repo, _, _ := newSvc(t, seedOAuth2KV(t, fullOAuth2Setting(string(commonModel.OAuth2GITHUB))))
	lookupErr := errors.New("user lookup failed")
	repo.EXPECT().
		GetUserByID(mock.Anything, "u-1").
		Return(userModel.User{}, lookupErr).
		Once()

	out, err := svc.GetOAuthInfo(ctx, string(commonModel.OAuth2GITHUB))
	require.ErrorIs(t, err, lookupErr)
	assert.Equal(t, userModel.OAuthInfoDto{}, out)
}

func TestGetOAuthInfo_OAuth2Branch(t *testing.T) {
	ctx := helpers.CtxAsUser("admin-1")
	setting := fullOAuth2Setting(string(commonModel.OAuth2GITHUB))
	setting.IsOIDC = false
	svc, repo, _, _ := newSvc(t, seedOAuth2KV(t, setting))

	repo.EXPECT().
		GetUserByID(mock.Anything, "admin-1").
		Return(userModel.User{ID: "admin-1", IsAdmin: true}, nil).
		Once()
	repo.EXPECT().
		GetOAuthInfo("admin-1", string(commonModel.OAuth2GITHUB)).
		Return(userModel.UserExternalIdentity{
			UserID:   "admin-1",
			Provider: string(commonModel.OAuth2GITHUB),
			Subject:  "ext-42",
		}, nil).
		Once()

	out, err := svc.GetOAuthInfo(ctx, string(commonModel.OAuth2GITHUB))
	require.NoError(t, err)
	assert.Equal(t, string(commonModel.OAuth2GITHUB), out.Provider)
	assert.Equal(t, "admin-1", out.UserID)
	assert.Equal(t, "ext-42", out.OAuthID)
	assert.Equal(t, string(authModel.AuthTypeOAuth2), out.AuthType)
}

func TestGetOAuthInfo_OIDCBranch(t *testing.T) {
	ctx := helpers.CtxAsUser("admin-1")
	setting := fullOAuth2Setting(string(commonModel.OAuth2CUSTOM))
	setting.IsOIDC = true
	setting.Issuer = "https://idp.example.com"
	svc, repo, _, _ := newSvc(t, seedOAuth2KV(t, setting))

	repo.EXPECT().
		GetUserByID(mock.Anything, "admin-1").
		Return(userModel.User{ID: "admin-1", IsAdmin: true}, nil).
		Once()
	repo.EXPECT().
		GetOAuthOIDCInfo("admin-1", string(commonModel.OAuth2CUSTOM), "https://idp.example.com").
		Return(userModel.UserExternalIdentity{
			UserID:   "admin-1",
			Provider: string(commonModel.OAuth2CUSTOM),
			Subject:  "sub-oidc",
			Issuer:   "https://idp.example.com",
		}, nil).
		Once()

	out, err := svc.GetOAuthInfo(ctx, string(commonModel.OAuth2CUSTOM))
	require.NoError(t, err)
	assert.Equal(t, "sub-oidc", out.OAuthID)
	assert.Equal(t, "https://idp.example.com", out.Issuer)
	assert.Equal(t, string(authModel.AuthTypeOIDC), out.AuthType)
}
