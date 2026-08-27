// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/test/helpers"
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func settingJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return string(raw)
}

func TestGetSetting(t *testing.T) {
	t.Run("returns stored value", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.SystemSettingsKey).
			Return(settingJSON(t, settingModel.SystemSetting{SiteTitle: "My Blog", ServerName: "srv"}), nil).
			Once()

		var s settingModel.SystemSetting
		require.NoError(t, d.build().GetSetting(&s))
		assert.Equal(t, "My Blog", s.SiteTitle)
		assert.Equal(t, "srv", s.ServerName)
	})

	t.Run("missing key falls back to default without error", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().Get(mock.Anything, commonModel.SystemSettingsKey).Return("", kvstore.ErrNotFound).Once()

		var s settingModel.SystemSetting
		require.NoError(t, d.build().GetSetting(&s))
	})

	t.Run("backend error propagates", func(t *testing.T) {
		d := newDeps(t)
		boom := errors.New("kv down")
		d.kv.EXPECT().Get(mock.Anything, commonModel.SystemSettingsKey).Return("", boom).Once()

		var s settingModel.SystemSetting
		require.ErrorIs(t, d.build().GetSetting(&s), boom)
	})
}

func TestGetAgentInfo(t *testing.T) {
	t.Run("returns stored value", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.AgentSettingKey).
			Return(settingJSON(t, settingModel.AgentSetting{Enable: true, Model: "gpt-x"}), nil).
			Once()

		var a settingModel.AgentSetting
		require.NoError(t, d.build().GetAgentInfo(&a))
		assert.True(t, a.Enable)
		assert.Equal(t, "gpt-x", a.Model)
	})

	t.Run("backend error propagates", func(t *testing.T) {
		d := newDeps(t)
		boom := errors.New("kv down")
		d.kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return("", boom).Once()

		var a settingModel.AgentSetting
		require.ErrorIs(t, d.build().GetAgentInfo(&a), boom)
	})
}

func TestGetOAuth2Status(t *testing.T) {
	t.Run("ready when both allowlists present", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.OAuth2SettingKey).
			Return(settingJSON(t, settingModel.OAuth2Setting{
				Enable:                        true,
				Provider:                      "github",
				AuthRedirectAllowedReturnURLs: []string{"https://app.example.com"},
				CORSAllowedOrigins:            []string{"https://app.example.com"},
			}), nil).
			Once()

		var st settingModel.OAuth2Status
		require.NoError(t, d.build().GetOAuth2Status(&st))
		assert.True(t, st.Enabled)
		assert.Equal(t, "github", st.Provider)
		assert.True(t, st.OAuthReady)
	})

	t.Run("not ready when allowlists empty", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.OAuth2SettingKey).
			Return(settingJSON(t, settingModel.OAuth2Setting{Enable: false, Provider: "github"}), nil).
			Once()

		var st settingModel.OAuth2Status
		require.NoError(t, d.build().GetOAuth2Status(&st))
		assert.False(t, st.OAuthReady)
	})

	t.Run("backend error propagates", func(t *testing.T) {
		d := newDeps(t)
		boom := errors.New("kv down")
		d.kv.EXPECT().Get(mock.Anything, commonModel.OAuth2SettingKey).Return("", boom).Once()

		var st settingModel.OAuth2Status
		require.ErrorIs(t, d.build().GetOAuth2Status(&st), boom)
	})
}

func TestGetPasskeyStatus(t *testing.T) {
	t.Run("ready when rpid and origins present", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.PasskeySettingKey).
			Return(settingJSON(t, settingModel.PasskeySetting{
				WebAuthnRPID:           "example.com",
				WebAuthnAllowedOrigins: []string{"https://example.com"},
			}), nil).
			Once()

		var st settingModel.PasskeyStatus
		require.NoError(t, d.build().GetPasskeyStatus(&st))
		assert.True(t, st.PasskeyReady)
	})

	t.Run("not ready when empty", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.PasskeySettingKey).
			Return(settingJSON(t, settingModel.PasskeySetting{}), nil).
			Once()

		var st settingModel.PasskeyStatus
		require.NoError(t, d.build().GetPasskeyStatus(&st))
		assert.False(t, st.PasskeyReady)
	})
}

func TestGetSnapshotScheduleSetting(t *testing.T) {
	d := newDeps(t)
	d.kv.EXPECT().
		Get(mock.Anything, commonModel.SnapshotScheduleKey).
		Return(settingJSON(t, settingModel.SnapshotSchedule{Enable: true, CronExpression: "0 3 * * *"}), nil).
		Once()

	var s settingModel.SnapshotSchedule
	require.NoError(t, d.build().GetSnapshotScheduleSetting(&s))
	assert.True(t, s.Enable)
	assert.Equal(t, "0 3 * * *", s.CronExpression)
}

func TestGetEmbeddingSetting(t *testing.T) {
	t.Run("returns stored value", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.EmbeddingSettingKey).
			Return(settingJSON(t, settingModel.EmbeddingSetting{Enable: true, Model: "emb-3", Dim: 1536}), nil).
			Once()

		got, err := d.build().GetEmbeddingSetting(helpers.CtxAnonymous())
		require.NoError(t, err)
		assert.True(t, got.Enable)
		assert.Equal(t, "emb-3", got.Model)
		assert.Equal(t, 1536, got.Dim)
	})

	t.Run("backend error propagates", func(t *testing.T) {
		d := newDeps(t)
		boom := errors.New("kv down")
		d.kv.EXPECT().Get(mock.Anything, commonModel.EmbeddingSettingKey).Return("", boom).Once()

		_, err := d.build().GetEmbeddingSetting(helpers.CtxAnonymous())
		require.ErrorIs(t, err, boom)
	})
}

func TestGetAdminGatedSettings(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("GetAgentSettings success", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.AgentSettingKey).
			Return(settingJSON(t, settingModel.AgentSetting{Enable: true, Model: "m"}), nil).
			Once()

		var a settingModel.AgentSetting
		require.NoError(t, d.build().GetAgentSettings(ctx, &a))
		assert.Equal(t, "m", a.Model)
	})

	t.Run("GetAgentSettings backend error", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		boom := errors.New("kv down")
		d.kv.EXPECT().Get(mock.Anything, commonModel.AgentSettingKey).Return("", boom).Once()

		var a settingModel.AgentSetting
		require.ErrorIs(t, d.build().GetAgentSettings(ctx, &a), boom)
	})

	t.Run("GetOAuth2Setting success", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.OAuth2SettingKey).
			Return(settingJSON(t, settingModel.OAuth2Setting{Enable: true, Provider: "github", ClientID: "cid"}), nil).
			Once()

		var o settingModel.OAuth2Setting
		require.NoError(t, d.build().GetOAuth2Setting(ctx, &o))
		assert.Equal(t, "cid", o.ClientID)
	})

	t.Run("GetPasskeySetting success", func(t *testing.T) {
		d := newDeps(t)
		d.expectAdmin()
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.PasskeySettingKey).
			Return(settingJSON(t, settingModel.PasskeySetting{WebAuthnRPID: "example.com"}), nil).
			Once()

		var p settingModel.PasskeySetting
		require.NoError(t, d.build().GetPasskeySetting(ctx, &p))
		assert.Equal(t, "example.com", p.WebAuthnRPID)
	})
}

func TestGetS3Setting(t *testing.T) {
	stored := func(t *testing.T) string {
		return settingJSON(t, settingModel.S3Setting{
			Enable:     true,
			Provider:   "aws",
			Endpoint:   "s3.example.com",
			AccessKey:  "AKID",
			SecretKey:  "SECRET",
			BucketName: "bkt",
		})
	}

	t.Run("anonymous gets masked secrets", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().Get(mock.Anything, commonModel.S3SettingKey).Return(stored(t), nil).Once()

		var s settingModel.S3Setting
		require.NoError(t, d.build().GetS3Setting(helpers.CtxAnonymous(), &s))
		assert.Equal(t, "******", s.AccessKey)
		assert.Equal(t, "******", s.SecretKey)
		assert.Equal(t, "******", s.BucketName)
		assert.Equal(t, "******", s.Endpoint)
	})

	t.Run("non-admin gets masked secrets", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().Get(mock.Anything, commonModel.S3SettingKey).Return(stored(t), nil).Once()
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(helpers.NewUser(), nil).
			Once()

		var s settingModel.S3Setting
		require.NoError(t, d.build().GetS3Setting(helpers.CtxAsUser(testUserID), &s))
		assert.Equal(t, "******", s.SecretKey)
	})

	t.Run("admin sees plaintext secrets", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().Get(mock.Anything, commonModel.S3SettingKey).Return(stored(t), nil).Once()
		d.expectAdmin()

		var s settingModel.S3Setting
		require.NoError(t, d.build().GetS3Setting(helpers.CtxAsUser(testUserID), &s))
		assert.Equal(t, "AKID", s.AccessKey)
		assert.Equal(t, "SECRET", s.SecretKey)
		assert.Equal(t, "bkt", s.BucketName)
	})

	t.Run("backend error propagates before masking", func(t *testing.T) {
		d := newDeps(t)
		boom := errors.New("kv down")
		d.kv.EXPECT().Get(mock.Anything, commonModel.S3SettingKey).Return("", boom).Once()

		var s settingModel.S3Setting
		require.ErrorIs(t, d.build().GetS3Setting(helpers.CtxAnonymous(), &s), boom)
	})

	t.Run("authenticated user lookup error propagates", func(t *testing.T) {
		d := newDeps(t)
		boom := errors.New("lookup failed")
		d.kv.EXPECT().Get(mock.Anything, commonModel.S3SettingKey).Return(stored(t), nil).Once()
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(helpers.NewUser(), boom).
			Once()

		var s settingModel.S3Setting
		require.ErrorIs(t, d.build().GetS3Setting(helpers.CtxAsUser(testUserID), &s), boom)
	})
}

func TestUpdateSetting(t *testing.T) {
	ctx := helpers.CtxAsUser(testUserID)

	t.Run("non-admin denied inside transaction", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().Get(mock.Anything, commonModel.SystemSettingsKey).Return("", kvstore.ErrNotFound).Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(helpers.NewUser(), nil).
			Once()

		err := d.build().UpdateSetting(ctx, &settingModel.SystemSettingDto{SiteTitle: "x"})
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("admin persists setting and derived server_url", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().Get(mock.Anything, commonModel.SystemSettingsKey).Return("", kvstore.ErrNotFound).Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		d.kv.EXPECT().Set(mock.Anything, commonModel.SystemSettingsKey, mock.Anything).Return(nil).Once()
		d.kv.EXPECT().
			Set(mock.Anything, commonModel.ServerURLKey, "https://my.example.com").
			Return(nil).
			Once()

		err := d.build().UpdateSetting(ctx, &settingModel.SystemSettingDto{
			SiteTitle: "x", ServerURL: "https://my.example.com/",
		})
		require.NoError(t, err)
	})

	t.Run("server logo change confirms temp file", func(t *testing.T) {
		d := newDeps(t)
		file := filemock.NewMockService(t)
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.SystemSettingsKey).
			Return(settingJSON(t, settingModel.SystemSetting{ServerLogo: "old-logo.png"}), nil).
			Once()
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, mock.Anything).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		d.kv.EXPECT().Set(mock.Anything, commonModel.SystemSettingsKey, mock.Anything).Return(nil).Once()
		d.kv.EXPECT().Set(mock.Anything, commonModel.ServerURLKey, mock.Anything).Return(nil).Once()
		file.EXPECT().ConfirmTempFiles(mock.Anything, []string{"logo-file-1"}).Return(nil).Once()

		svc := settingService.NewSettingService(
			d.tx, d.common, file, nil, d.kv, d.settingRepo, d.webhookRepo, nil, d.revoker,
			func() *busen.Bus { return d.bus },
		)
		err := svc.UpdateSetting(ctx, &settingModel.SystemSettingDto{
			ServerLogo:       "new-logo.png",
			ServerLogoFileID: "logo-file-1",
		})
		require.NoError(t, err)
	})
}

func TestBootstrapDefaultLocale(t *testing.T) {
	ctx := helpers.CtxAnonymous()

	t.Run("blank locale is a no-op", func(t *testing.T) {
		d := newDeps(t)
		require.NoError(t, d.build().BootstrapDefaultLocale(ctx, "   "))
	})

	t.Run("default locale is a no-op", func(t *testing.T) {
		d := newDeps(t)
		require.NoError(t, d.build().BootstrapDefaultLocale(ctx, string(commonModel.DefaultLocale)))
	})

	t.Run("non-default locale written when current is still default", func(t *testing.T) {
		d := newDeps(t)
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.kv.EXPECT().Get(mock.Anything, commonModel.SystemSettingsKey).Return("", kvstore.ErrNotFound).Once()
		d.kv.EXPECT().Set(mock.Anything, commonModel.SystemSettingsKey, mock.Anything).Return(nil).Once()

		require.NoError(t, d.build().BootstrapDefaultLocale(ctx, "en-US"))
	})

	t.Run("not overwritten when admin already customized locale", func(t *testing.T) {
		d := newDeps(t)
		d.tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTxExec()).Once()
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.SystemSettingsKey).
			Return(settingJSON(t, settingModel.SystemSetting{DefaultLocale: "en-US"}), nil).
			Once()

		require.NoError(t, d.build().BootstrapDefaultLocale(ctx, "en-US"))
	})
}
