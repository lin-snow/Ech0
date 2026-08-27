// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/kvstore"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testRPID   = "example.com"
	testOrigin = "https://example.com"
)

func TestUserHandleRoundTrip(t *testing.T) {
	for _, id := range []string{"u-1", "", "01HXY-uuid-7"} {
		handle := makeUserHandle(id)
		assert.Equal(t, id, userIDFromHandle(handle))
	}
}

func TestNewNonce(t *testing.T) {
	n1, err := newNonce()
	require.NoError(t, err)
	assert.NotEmpty(t, n1)

	decoded, derr := base64.RawURLEncoding.DecodeString(n1)
	require.NoError(t, derr)
	assert.Len(t, decoded, 32)

	n2, err := newNonce()
	require.NoError(t, err)
	assert.NotEqual(t, n1, n2, "两次 nonce 必须不同")
}

func TestPasskeySessionKeyBuilders(t *testing.T) {
	assert.Equal(t, passkeyRegKey+":abc", getPasskeyRegisterSessionKey("abc"))
	assert.Equal(t, passkeyLoginKey+":abc", getPasskeyLoginSessionKey("abc"))
}

func TestBindingPermissionError(t *testing.T) {
	cases := []struct {
		provider string
		want     string
	}{
		{string(commonModel.OAuth2GITHUB), commonModel.NO_PERMISSION_BINDING_GITHUB},
		{string(commonModel.OAuth2GOOGLE), commonModel.NO_PERMISSION_BINDING_GOOGLE},
		{string(commonModel.OAuth2QQ), commonModel.NO_PERMISSION_BINDING_QQ},
		{string(commonModel.OAuth2CUSTOM), commonModel.NO_PERMISSION_BINDING_CUSTOM},
		{"unknown", commonModel.NO_PERMISSION_DENIED},
	}
	for _, tc := range cases {
		t.Run(tc.provider, func(t *testing.T) {
			require.EqualError(t, bindingPermissionError(tc.provider), tc.want)
		})
	}
}

func TestPasskeyRegisterBegin_Success(t *testing.T) {
	ctx := helpers.CtxAsUser("u-1")
	svc, repo, _, _ := newSvc(t, kvstore.NewMemory())

	repo.EXPECT().
		GetUserByID(mock.Anything, "u-1").
		Return(userModel.User{ID: "u-1", Username: "alice"}, nil).
		Once()
	repo.EXPECT().
		ListPasskeysByUserID("u-1").
		Return(nil, nil).
		Once()

	var cachedKey string
	var cachedVal any
	repo.EXPECT().
		CacheSetPasskeySession(mock.Anything, mock.Anything, passkeySessionTTL).
		Run(func(key string, val any, _ time.Duration) {
			cachedKey = key
			cachedVal = val
		}).
		Once()

	resp, err := svc.PasskeyRegisterBegin(ctx, testRPID, testOrigin, "")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Nonce)
	require.NotNil(t, resp.PublicKey)

	assert.Equal(t, getPasskeyRegisterSessionKey(resp.Nonce), cachedKey)
	sess, ok := cachedVal.(passkeySessionCache)
	require.True(t, ok)
	assert.Equal(t, testOrigin, sess.Origin)
	assert.Equal(t, "Passkey", sess.DeviceName)
}

func TestPasskeyRegisterBegin_UserLookupError(t *testing.T) {
	ctx := helpers.CtxAsUser("u-err")
	svc, repo, _, _ := newSvc(t, kvstore.NewMemory())

	lookupErr := errors.New("user not found")
	repo.EXPECT().
		GetUserByID(mock.Anything, "u-err").
		Return(userModel.User{}, lookupErr).
		Once()

	resp, err := svc.PasskeyRegisterBegin(ctx, testRPID, testOrigin, "My Phone")
	require.ErrorIs(t, err, lookupErr)
	assert.Empty(t, resp.Nonce)
}

func TestPasskeyLoginBegin_Success(t *testing.T) {
	svc, repo, _, _ := newSvc(t, kvstore.NewMemory())

	var cachedKey string
	repo.EXPECT().
		CacheSetPasskeySession(mock.Anything, mock.Anything, passkeySessionTTL).
		Run(func(key string, _ any, _ time.Duration) { cachedKey = key }).
		Once()

	resp, err := svc.PasskeyLoginBegin(testRPID, testOrigin)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Nonce)
	require.NotNil(t, resp.PublicKey)
	assert.Equal(t, getPasskeyLoginSessionKey(resp.Nonce), cachedKey)
}

func TestListPasskeys_MapsToDTO(t *testing.T) {
	ctx := helpers.CtxAsUser("u-1")
	svc, repo, _, _ := newSvc(t, kvstore.NewMemory())

	repo.EXPECT().
		ListPasskeysByUserID("u-1").
		Return([]authModel.Passkey{
			{ID: "pk-1", DeviceName: "Phone", AAGUID: "aaguid-1", LastUsedAt: 100, CreatedAt: 50},
			{ID: "pk-2", DeviceName: "Laptop", AAGUID: "aaguid-2", LastUsedAt: 200, CreatedAt: 60},
		}, nil).
		Once()

	devs, err := svc.ListPasskeys(ctx)
	require.NoError(t, err)
	require.Len(t, devs, 2)
	assert.Equal(t, "pk-1", devs[0].ID)
	assert.Equal(t, "Phone", devs[0].DeviceName)
	assert.Equal(t, "aaguid-2", devs[1].AAGUID)
	assert.Equal(t, int64(200), devs[1].LastUsedAt)
}

func TestListPasskeys_Error(t *testing.T) {
	ctx := helpers.CtxAsUser("u-1")
	svc, repo, _, _ := newSvc(t, kvstore.NewMemory())

	listErr := errors.New("db down")
	repo.EXPECT().ListPasskeysByUserID("u-1").Return(nil, listErr).Once()

	devs, err := svc.ListPasskeys(ctx)
	require.ErrorIs(t, err, listErr)
	assert.Nil(t, devs)
}

func TestDeletePasskey(t *testing.T) {
	t.Run("success runs delete inside tx", func(t *testing.T) {
		ctx := helpers.CtxAsUser("u-1")
		svc, repo, _, tx := newSvc(t, kvstore.NewMemory())
		runsTxInline(tx)
		repo.EXPECT().
			DeletePasskeyByID(mock.Anything, "u-1", "pk-1").
			Return(nil).
			Once()

		require.NoError(t, svc.DeletePasskey(ctx, "pk-1"))
	})

	t.Run("repo error propagates", func(t *testing.T) {
		ctx := helpers.CtxAsUser("u-1")
		svc, repo, _, tx := newSvc(t, kvstore.NewMemory())
		runsTxInline(tx)
		delErr := errors.New("not owned")
		repo.EXPECT().
			DeletePasskeyByID(mock.Anything, "u-1", "pk-1").
			Return(delErr).
			Once()

		require.ErrorIs(t, svc.DeletePasskey(ctx, "pk-1"), delErr)
	})
}

func TestUpdatePasskeyDeviceName(t *testing.T) {
	t.Run("blank name rejected before tx", func(t *testing.T) {
		ctx := helpers.CtxAsUser("u-1")
		svc, _, _, _ := newSvc(t, kvstore.NewMemory())
		err := svc.UpdatePasskeyDeviceName(ctx, "pk-1", "   ")
		require.EqualError(t, err, commonModel.INVALID_PARAMS_BODY)
	})

	t.Run("valid name updates inside tx", func(t *testing.T) {
		ctx := helpers.CtxAsUser("u-1")
		svc, repo, _, tx := newSvc(t, kvstore.NewMemory())
		runsTxInline(tx)
		repo.EXPECT().
			UpdatePasskeyDeviceName(mock.Anything, "u-1", "pk-1", "New Name").
			Return(nil).
			Once()

		require.NoError(t, svc.UpdatePasskeyDeviceName(ctx, "pk-1", "New Name"))
	})

	t.Run("repo error propagates", func(t *testing.T) {
		ctx := helpers.CtxAsUser("u-1")
		svc, repo, _, tx := newSvc(t, kvstore.NewMemory())
		runsTxInline(tx)
		updErr := errors.New("update failed")
		repo.EXPECT().
			UpdatePasskeyDeviceName(mock.Anything, "u-1", "pk-1", "Name").
			Return(updErr).
			Once()

		require.ErrorIs(t, svc.UpdatePasskeyDeviceName(ctx, "pk-1", "Name"), updErr)
	})
}
