// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/lin-snow/ech0/internal/test/mocks/connectmock"
	"github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/lin-snow/ech0/internal/test/mocks/txmock"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func passthroughTx(t *testing.T) *txmock.MockTransactor {
	t.Helper()
	tx := txmock.NewMockTransactor(t)
	tx.EXPECT().
		Run(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).
		Once()
	return tx
}

func adminCommon(t *testing.T, userID string) *commonmock.MockService {
	t.Helper()
	cs := commonmock.NewMockService(t)
	cs.EXPECT().
		CommonGetUserByUserId(mock.Anything, userID).
		Return(userModel.User{IsAdmin: true}, nil).
		Once()
	return cs
}

func TestAddConnect_NonAdminDenied(t *testing.T) {
	const userID = "u-1"
	tx := passthroughTx(t)
	cs := commonmock.NewMockService(t)
	cs.EXPECT().
		CommonGetUserByUserId(mock.Anything, userID).
		Return(userModel.User{IsAdmin: false}, nil).
		Once()
	repo := connectmock.NewMockRepository(t)

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: "https://example.com"})

	require.Error(t, err)
	assert.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
}

func TestAddConnect_UserLookupErrorPropagates(t *testing.T) {
	const userID = "u-1"
	wantErr := errors.New("db down")
	tx := passthroughTx(t)
	cs := commonmock.NewMockService(t)
	cs.EXPECT().
		CommonGetUserByUserId(mock.Anything, userID).
		Return(userModel.User{}, wantErr).
		Once()
	repo := connectmock.NewMockRepository(t)

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: "https://example.com"})

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestAddConnect_EmptyURLRejected(t *testing.T) {
	const userID = "u-1"
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: ""})

	require.Error(t, err)
	assert.EqualError(t, err, commonModel.INVALID_CONNECTION_URL)
}

func TestAddConnect_SSRFPrevalidation(t *testing.T) {
	const userID = "u-1"
	cases := []struct {
		name string
		url  string
	}{
		{"loopback ipv4", "http://127.0.0.1"},
		{"localhost alias", "http://localhost"},
		{"localhost suffix", "http://api.localhost"},
		{"cloud metadata link-local", "http://169.254.169.254"},
		{"private 192.168", "http://192.168.1.10"},
		{"private 10.x", "http://10.0.0.5"},
		{"loopback ipv6", "http://[::1]"},
		{"docker host alias", "http://host.docker.internal"},
		{"non-http scheme", "ftp://example.com"},
		{"embedded userinfo", "http://user:pass@example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx := passthroughTx(t)
			cs := adminCommon(t, userID)
			repo := connectmock.NewMockRepository(t)

			svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
			err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: tc.url})

			require.Error(t, err)
			assert.EqualError(t, err, commonModel.INVALID_CONNECTION_URL)
		})
	}
}

func TestAddConnect_DuplicateRejected(t *testing.T) {
	const userID = "u-1"
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().
		GetAllConnects(mock.Anything).
		Return([]model.Connected{{ID: "x", ConnectURL: "https://example.com"}}, nil).
		Once()

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: "https://example.com"})

	require.Error(t, err)
	assert.EqualError(t, err, commonModel.CONNECT_HAS_EXISTS)
}

func TestAddConnect_GetAllConnectsErrorPropagates(t *testing.T) {
	const userID = "u-1"
	wantErr := errors.New("list failed")
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().
		GetAllConnects(mock.Anything).
		Return(nil, wantErr).
		Once()

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: "https://example.com"})

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestAddConnect_Success_TrimsURLBeforePersist(t *testing.T) {
	const userID = "u-1"
	tx := passthroughTx(t)
	cs := adminCommon(t, userID)
	repo := connectmock.NewMockRepository(t)
	repo.EXPECT().
		GetAllConnects(mock.Anything).
		Return([]model.Connected{}, nil).
		Once()
	repo.EXPECT().
		CreateConnect(mock.Anything, mock.MatchedBy(func(c *model.Connected) bool {
			return c != nil && c.ConnectURL == "https://example.com"
		})).
		Return(nil).
		Once()

	svc := connectService.NewConnectService(tx, repo, nil, cs, nil)
	err := svc.AddConnect(helpers.CtxAsUser(userID), model.Connected{ConnectURL: "  https://example.com/  "})

	require.NoError(t, err)
}

func systemSettingJSON(t *testing.T, serverURL, serverLogo string) string {
	t.Helper()
	raw, err := json.Marshal(settingModel.SystemSetting{
		ServerName: "My Ech0",
		ServerURL:  serverURL,
		ServerLogo: serverLogo,
	})
	require.NoError(t, err)
	return string(raw)
}

func TestGetConnect_LogoDerivation(t *testing.T) {
	cases := []struct {
		name       string
		serverURL  string
		serverLogo string
		wantLogo   string
	}{
		{"empty logo falls back to default", "https://ech0.app", "", "https://ech0.app/Ech0.svg"},
		{"Ech0.svg sentinel falls back to default", "https://ech0.app", "Ech0.svg", "https://ech0.app/Ech0.svg"},
		{"/Ech0.svg sentinel falls back to default", "https://ech0.app", "/Ech0.svg", "https://ech0.app/Ech0.svg"},
		{"trailing slash on server url is trimmed", "https://ech0.app/", "", "https://ech0.app/Ech0.svg"},
		{"absolute https logo is kept verbatim", "https://ech0.app", "https://cdn.example.com/logo.png", "https://cdn.example.com/logo.png"},
		{"absolute http logo is kept verbatim", "https://ech0.app", "http://cdn.example.com/logo.png", "http://cdn.example.com/logo.png"},
		{"root-relative logo joins server url", "https://ech0.app", "/static/logo.png", "https://ech0.app/static/logo.png"},
		{"relative logo joins server url with slash", "https://ech0.app/", "logo.png", "https://ech0.app/logo.png"},
		{"surrounding whitespace is trimmed", "https://ech0.app", "  /static/logo.png  ", "https://ech0.app/static/logo.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kv := kvmock.NewMockStore(t)
			kv.EXPECT().
				Get(mock.Anything, commonModel.SystemSettingsKey).
				Return(systemSettingJSON(t, tc.serverURL, tc.serverLogo), nil).
				Once()

			cs := commonmock.NewMockService(t)
			cs.EXPECT().GetOwner().Return(userModel.User{Username: "owner"}, nil).Once()

			echoRepo := connectmock.NewMockEchoRepository(t)
			echoRepo.EXPECT().GetTodayEchos(true, "UTC").Return([]echoModel.Echo{}).Once()
			echoRepo.EXPECT().GetEchosByPage(1, 1, "", true).Return(nil, int64(0)).Once()

			svc := connectService.NewConnectService(nil, nil, echoRepo, cs, kv)
			got, err := svc.GetConnect()

			require.NoError(t, err)
			assert.Equal(t, tc.wantLogo, got.Logo)
		})
	}
}

func TestGetConnect_PopulatesMetrics(t *testing.T) {
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().
		Get(mock.Anything, commonModel.SystemSettingsKey).
		Return(systemSettingJSON(t, "https://ech0.app", ""), nil).
		Once()

	cs := commonmock.NewMockService(t)
	cs.EXPECT().GetOwner().Return(userModel.User{Username: "alice"}, nil).Once()

	echoRepo := connectmock.NewMockEchoRepository(t)
	echoRepo.EXPECT().GetTodayEchos(true, "UTC").Return(make([]echoModel.Echo, 3)).Once()
	echoRepo.EXPECT().GetEchosByPage(1, 1, "", true).Return(nil, int64(42)).Once()

	svc := connectService.NewConnectService(nil, nil, echoRepo, cs, kv)
	got, err := svc.GetConnect()

	require.NoError(t, err)
	assert.Equal(t, "My Ech0", got.ServerName)
	assert.Equal(t, "https://ech0.app", got.ServerURL)
	assert.Equal(t, "alice", got.SysUsername)
	assert.Equal(t, 3, got.TodayEchos)
	assert.Equal(t, 42, got.TotalEchos)
	assert.Equal(t, versionPkg.Version, got.Version)
}

func TestGetConnect_SettingErrorShortCircuits(t *testing.T) {
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().
		Get(mock.Anything, commonModel.SystemSettingsKey).
		Return("", errors.New("kv unavailable")).
		Once()

	cs := commonmock.NewMockService(t)
	echoRepo := connectmock.NewMockEchoRepository(t)

	svc := connectService.NewConnectService(nil, nil, echoRepo, cs, kv)
	_, err := svc.GetConnect()

	require.Error(t, err)
}

func TestGetConnect_OwnerErrorShortCircuits(t *testing.T) {
	wantErr := errors.New("no owner")
	kv := kvmock.NewMockStore(t)
	kv.EXPECT().
		Get(mock.Anything, commonModel.SystemSettingsKey).
		Return(systemSettingJSON(t, "https://ech0.app", ""), nil).
		Once()

	cs := commonmock.NewMockService(t)
	cs.EXPECT().GetOwner().Return(userModel.User{}, wantErr).Once()

	echoRepo := connectmock.NewMockEchoRepository(t)

	svc := connectService.NewConnectService(nil, nil, echoRepo, cs, kv)
	_, err := svc.GetConnect()

	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}
