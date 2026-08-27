// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAllWebhooks_Success(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("returns repository webhooks", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.webhookRepo.EXPECT().
			GetAllWebhooks(mock.Anything).
			Return([]webhookModel.Webhook{{ID: "wh-1", Name: "hook"}}, nil).
			Once()

		got, err := d.build().GetAllWebhooks(ctx)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "wh-1", got[0].ID)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		boom := errors.New("list failed")
		d.webhookRepo.EXPECT().GetAllWebhooks(mock.Anything).Return(nil, boom).Once()

		_, err := d.build().GetAllWebhooks(ctx)
		require.ErrorIs(t, err, boom)
	})
}

func TestDeleteWebhook_Success(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
	d.webhookRepo.EXPECT().DeleteWebhookByID(mock.Anything, "wh-9").Return(nil).Once()

	require.NoError(t, d.build().DeleteWebhook(helpers.CtxAsUser(testUserID), "wh-9"))
}

func TestUpdateWebhook_Validation(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("empty url rejected", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		err := d.build().UpdateWebhook(ctx, "wh-1", &settingModel.WebhookDto{Name: "hook", URL: ""})
		require.Error(t, err)
		assert.Equal(t, commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY, err.Error())
	})

	t.Run("private network url rejected", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		err := d.build().UpdateWebhook(ctx, "wh-1", &settingModel.WebhookDto{Name: "hook", URL: "http://127.0.0.1/x"})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_WEBHOOK_URL, err.Error())
	})
}

func TestTestWebhook_ValidationBranches(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("get webhook error propagates", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		boom := errors.New("not found")
		d.webhookRepo.EXPECT().GetWebhookByID(mock.Anything, "wh-x").Return(nil, boom).Once()

		require.ErrorIs(t, d.build().TestWebhook(ctx, "wh-x"), boom)
	})

	t.Run("stored unsafe url rejected before send", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.webhookRepo.EXPECT().
			GetWebhookByID(mock.Anything, "wh-x").
			Return(&webhookModel.Webhook{ID: "wh-x", URL: "http://127.0.0.1/x"}, nil).
			Once()

		err := d.build().TestWebhook(ctx, "wh-x")
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_WEBHOOK_URL, err.Error())
	})
}

func TestUpdateOAuth2Setting_Success(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.kv.EXPECT().
		Set(mock.Anything, commonModel.OAuth2SettingKey, mock.Anything).
		Return(nil).
		Once()

	err := d.build().UpdateOAuth2Setting(helpers.CtxAsUser(testUserID), &settingModel.OAuth2SettingDto{
		Enable:                        true,
		Provider:                      "github",
		ClientID:                      "cid",
		ClientSecret:                  "secret",
		AuthURL:                       "https://github.com/login/oauth/authorize/",
		AuthRedirectAllowedReturnURLs: []string{"https://app.example.com/"},
		CORSAllowedOrigins:            []string{"https://app.example.com/"},
	})
	require.NoError(t, err)
}

func TestUpdatePasskeySetting_Success(t *testing.T) {
	d := newDeps(t)
	d.expectAdmin()
	d.kv.EXPECT().
		Set(mock.Anything, commonModel.PasskeySettingKey, mock.Anything).
		Return(nil).
		Once()

	err := d.build().UpdatePasskeySetting(helpers.CtxAsUser(testUserID), &settingModel.PasskeySettingDto{
		WebAuthnRPID:           "example.com",
		WebAuthnAllowedOrigins: []string{"https://example.com/"},
	})
	require.NoError(t, err)
}
