// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	kvmock "github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	settingmock "github.com/lin-snow/ech0/internal/test/mocks/settingmock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type deps struct {
	tx          *txmock.MockTransactor
	common      *commonmock.MockService
	kv          *kvmock.MockStore
	settingRepo *settingmock.MockSettingRepository
	webhookRepo *settingmock.MockWebhookRepository
	revoker     *settingmock.MockTokenRevoker
	bus         *busen.Bus
}

func newDeps(t *testing.T) *deps {
	t.Helper()
	return &deps{
		tx:          txmock.NewMockTransactor(t),
		common:      commonmock.NewMockService(t),
		kv:          kvmock.NewMockStore(t),
		settingRepo: settingmock.NewMockSettingRepository(t),
		webhookRepo: settingmock.NewMockWebhookRepository(t),
		revoker:     settingmock.NewMockTokenRevoker(t),
		bus:         busen.New(),
	}
}

func (d *deps) build() *settingService.SettingService {
	return settingService.NewSettingService(
		d.tx,
		d.common,
		nil,
		nil,
		d.kv,
		d.settingRepo,
		d.webhookRepo,
		nil,
		d.revoker,
		func() *busen.Bus { return d.bus },
	)
}

func runTxExec() func(context.Context, func(context.Context) error) error {
	return func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
}

const testUserID = "user-1"

func (d *deps) expectAdmin() {
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, mock.Anything).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
}

func TestSettingService_NonAdminDenied(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	calls := map[string]func(svc *settingService.SettingService) error{
		"GetAllWebhooks": func(svc *settingService.SettingService) error {
			_, err := svc.GetAllWebhooks(ctx)
			return err
		},
		"CreateWebhook": func(svc *settingService.SettingService) error {
			return svc.CreateWebhook(ctx, &settingModel.WebhookDto{Name: "n", URL: "https://e.example.com"})
		},
		"UpdateWebhook": func(svc *settingService.SettingService) error {
			return svc.UpdateWebhook(ctx, "id-1", &settingModel.WebhookDto{Name: "n", URL: "https://e.example.com"})
		},
		"DeleteWebhook": func(svc *settingService.SettingService) error {
			return svc.DeleteWebhook(ctx, "id-1")
		},
		"TestWebhook": func(svc *settingService.SettingService) error {
			return svc.TestWebhook(ctx, "id-1")
		},
		"ListAccessTokens": func(svc *settingService.SettingService) error {
			_, err := svc.ListAccessTokens(ctx)
			return err
		},
		"CreateAccessToken": func(svc *settingService.SettingService) error {
			_, err := svc.CreateAccessToken(ctx, &settingModel.AccessTokenSettingDto{Name: "t"})
			return err
		},
		"DeleteAccessToken": func(svc *settingService.SettingService) error {
			return svc.DeleteAccessToken(ctx, "id-1")
		},
		"UpdateS3Setting": func(svc *settingService.SettingService) error {
			return svc.UpdateS3Setting(ctx, &settingModel.S3SettingDto{})
		},
		"TestS3Connection": func(svc *settingService.SettingService) error {
			return svc.TestS3Connection(ctx, &settingModel.S3SettingDto{})
		},
		"GetOAuth2Setting": func(svc *settingService.SettingService) error {
			return svc.GetOAuth2Setting(ctx, &settingModel.OAuth2Setting{})
		},
		"UpdateOAuth2Setting": func(svc *settingService.SettingService) error {
			return svc.UpdateOAuth2Setting(ctx, &settingModel.OAuth2SettingDto{})
		},
		"GetAgentSettings": func(svc *settingService.SettingService) error {
			return svc.GetAgentSettings(ctx, &settingModel.AgentSetting{})
		},
		"UpdateAgentSettings": func(svc *settingService.SettingService) error {
			return svc.UpdateAgentSettings(ctx, &settingModel.AgentSettingDto{})
		},
		"TestAgentConnection": func(svc *settingService.SettingService) error {
			return svc.TestAgentConnection(ctx, &settingModel.AgentSettingDto{})
		},
		"GetPasskeySetting": func(svc *settingService.SettingService) error {
			return svc.GetPasskeySetting(ctx, &settingModel.PasskeySetting{})
		},
		"UpdatePasskeySetting": func(svc *settingService.SettingService) error {
			return svc.UpdatePasskeySetting(ctx, &settingModel.PasskeySettingDto{})
		},
		"UpdateSnapshotScheduleSetting": func(svc *settingService.SettingService) error {
			return svc.UpdateSnapshotScheduleSetting(ctx, &settingModel.SnapshotScheduleDto{})
		},
		"UpdateEmbeddingSetting": func(svc *settingService.SettingService) error {
			return svc.UpdateEmbeddingSetting(ctx, settingModel.EmbeddingSettingDto{})
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			d := newDeps(t)
			d.common.EXPECT().
				CommonGetUserByUserId(mock.Anything, mock.Anything).
				Return(helpers.NewUser(), nil).
				Once()

			err := call(d.build())
			require.Error(t, err)
			assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
		})
	}
}

func TestSettingService_PropagatesUserLookupError(t *testing.T) {
	d := newDeps(t)
	boom := errors.New("db down")
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, mock.Anything).
		Return(helpers.NewUser(), boom).
		Once()

	_, err := d.build().GetAllWebhooks(helpers.CtxAsUser(testUserID))
	require.ErrorIs(t, err, boom)
}

func TestCreateWebhook(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("empty url rejected before persistence", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		err := d.build().CreateWebhook(ctx, &settingModel.WebhookDto{Name: "hook", URL: ""})
		require.Error(t, err)
		assert.Equal(t, commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY, err.Error())
	})

	t.Run("unsafe url rejected", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		err := d.build().CreateWebhook(ctx, &settingModel.WebhookDto{Name: "hook", URL: "http://127.0.0.1/x"})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_WEBHOOK_URL, err.Error())
	})

	t.Run("valid webhook persisted via transaction", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.webhookRepo.EXPECT().
			CreateWebhook(mock.Anything, mock.MatchedBy(func(w *webhookModel.Webhook) bool {
				return w != nil && w.Name == "hook" && w.URL == "https://hooks.example.com/path" && w.Secret == "s3cr3t"
			})).
			Return(nil).
			Once()

		err := d.build().CreateWebhook(ctx, &settingModel.WebhookDto{
			Name:     "hook",
			URL:      "https://hooks.example.com/path",
			Secret:   "s3cr3t",
			IsActive: true,
		})
		require.NoError(t, err)
	})
}

func TestUpdateWebhook_Valid(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
	d.webhookRepo.EXPECT().
		UpdateWebhookByID(mock.Anything, "wh-9", mock.MatchedBy(func(w *webhookModel.Webhook) bool {
			return w != nil && w.URL == "https://hooks.example.com/u"
		})).
		Return(nil).
		Once()

	err := d.build().UpdateWebhook(helpers.CtxAsUser(testUserID), "wh-9", &settingModel.WebhookDto{
		Name: "n", URL: "https://hooks.example.com/u",
	})
	require.NoError(t, err)
}

func TestDeleteAccessToken_RevokesJTI(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("blacklists jti then deletes", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		future := time.Now().UTC().Add(3 * time.Hour).Unix()
		d.settingRepo.EXPECT().
			GetAccessTokenByID(mock.Anything, "tok-1").
			Return(settingModel.AccessTokenSetting{JTI: "jti-1", Expiry: &future}, nil).
			Once()
		d.revoker.EXPECT().
			RevokeToken("jti-1", mock.MatchedBy(func(ttl time.Duration) bool { return ttl > 0 })).
			Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.settingRepo.EXPECT().DeleteAccessTokenByID(mock.Anything, "tok-1").Return(nil).Once()

		require.NoError(t, d.build().DeleteAccessToken(ctx, "tok-1"))
	})

	t.Run("get failure does not block delete and skips revoke", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.settingRepo.EXPECT().
			GetAccessTokenByID(mock.Anything, "tok-2").
			Return(settingModel.AccessTokenSetting{}, errors.New("not found")).
			Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.settingRepo.EXPECT().DeleteAccessTokenByID(mock.Anything, "tok-2").Return(nil).Once()

		require.NoError(t, d.build().DeleteAccessToken(ctx, "tok-2"))
	})

	t.Run("empty jti skips revoke", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.settingRepo.EXPECT().
			GetAccessTokenByID(mock.Anything, "tok-3").
			Return(settingModel.AccessTokenSetting{JTI: ""}, nil).
			Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.settingRepo.EXPECT().DeleteAccessTokenByID(mock.Anything, "tok-3").Return(nil).Once()

		require.NoError(t, d.build().DeleteAccessToken(ctx, "tok-3"))
	})
}

func TestListAccessTokens(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("filters and purges expired tokens", func(t *testing.T) {
		d := newDeps(t)
		admin := helpers.NewUser(helpers.AsAdmin)
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(admin, nil).
			Once()

		past := time.Now().UTC().Add(-time.Hour).Unix()
		future := time.Now().UTC().Add(time.Hour).Unix()
		d.settingRepo.EXPECT().
			ListAccessTokens(mock.Anything, admin.ID).
			Return([]settingModel.AccessTokenSetting{
				{ID: "never", Expiry: nil},
				{ID: "future", Expiry: &future},
				{ID: "expired", Expiry: &past},
			}, nil).
			Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.settingRepo.EXPECT().DeleteAccessTokenByID(mock.Anything, "expired").Return(nil).Once()

		got, err := d.build().ListAccessTokens(ctx)
		require.NoError(t, err)
		require.Len(t, got, 2)
		ids := []string{got[0].ID, got[1].ID}
		assert.ElementsMatch(t, []string{"never", "future"}, ids)
	})

	t.Run("repository error yields empty slice and nil error", func(t *testing.T) {
		d := newDeps(t)
		admin := helpers.NewUser(helpers.AsAdmin)
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(admin, nil).
			Once()
		d.settingRepo.EXPECT().
			ListAccessTokens(mock.Anything, admin.ID).
			Return(nil, errors.New("boom")).
			Once()

		got, err := d.build().ListAccessTokens(ctx)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestCreateAccessToken_Success(t *testing.T) {
	helpers.SetJWTSecret(t, "unit-test-secret")
	d := newDeps(t)
	admin := helpers.NewUser(helpers.AsAdmin)
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, mock.Anything).
		Return(admin, nil).
		Once()
	d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
	d.settingRepo.EXPECT().
		CreateAccessToken(mock.Anything, mock.MatchedBy(func(tok *settingModel.AccessTokenSetting) bool {
			return tok != nil &&
				tok.Name == "cli-token" &&
				tok.UserID == admin.ID &&
				tok.TokenType == authModel.TokenTypeAccess &&
				tok.Audience == authModel.AudienceCLI &&
				tok.JTI != "" &&
				tok.Expiry != nil &&
				strings.Contains(tok.Scopes, authModel.ScopeEchoRead)
		})).
		Return(nil).
		Once()

	token, err := d.build().CreateAccessToken(helpers.CtxAsUser(testUserID), &settingModel.AccessTokenSettingDto{
		Name:     "cli-token",
		Expiry:   settingModel.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authModel.ScopeEchoRead, authModel.ScopeEchoRead},
		Audience: authModel.AudienceCLI,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestUpdateSnapshotScheduleSetting(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("invalid cron rejected", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		err := d.build().UpdateSnapshotScheduleSetting(ctx, &settingModel.SnapshotScheduleDto{
			Enable: true, CronExpression: "definitely not cron",
		})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_CRON_EXPRESSION, err.Error())
	})

	t.Run("valid cron persisted", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.kv.EXPECT().
			Set(mock.Anything, commonModel.SnapshotScheduleKey, mock.Anything).
			Return(nil).
			Once()

		err := d.build().UpdateSnapshotScheduleSetting(ctx, &settingModel.SnapshotScheduleDto{
			Enable: true, CronExpression: "0 2 * * 0",
		})
		require.NoError(t, err)
	})
}

func TestUpdateAgentSettings_NormalizesProtocol(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.kv.EXPECT().
		Set(mock.Anything, commonModel.AgentSettingKey, mock.MatchedBy(func(raw string) bool {
			return strings.Contains(raw, `"openai"`) && !strings.Contains(raw, "gemini")
		})).
		Return(nil).
		Once()

	err := d.build().UpdateAgentSettings(helpers.CtxAsUser(testUserID), &settingModel.AgentSettingDto{
		Protocol: "gemini", Model: "m", ApiKey: "k", BaseURL: "https://api.example.com/",
	})
	require.NoError(t, err)
}

func TestUpdateEmbeddingSetting_Persists(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.kv.EXPECT().
		Set(mock.Anything, commonModel.EmbeddingSettingKey, mock.Anything).
		Return(nil).
		Once()

	err := d.build().UpdateEmbeddingSetting(helpers.CtxAsUser(testUserID), settingModel.EmbeddingSettingDto{
		Enable: true, Model: " m ", ApiKey: " k ", BaseURL: " u ", Dim: 8, BatchSize: 4,
	})
	require.NoError(t, err)
}

func TestUpdateS3Setting_PersistsWhenStorageNil(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.kv.EXPECT().Get(mock.Anything, commonModel.S3SettingKey).Return("", nil).Once()
	d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
	d.kv.EXPECT().
		Set(mock.Anything, commonModel.S3SettingKey, mock.Anything).
		Return(nil).
		Once()

	err := d.build().UpdateS3Setting(helpers.CtxAsUser(testUserID), &settingModel.S3SettingDto{
		Enable: true, Provider: string(commonModel.AWS), Endpoint: "https://s3.example.com", BucketName: "b",
	})
	require.NoError(t, err)
}
