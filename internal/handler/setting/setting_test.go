// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	settingmock "github.com/lin-snow/ech0/internal/test/mocks/settingmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func assertStatusErr(t *testing.T, err error, wantStatus int) {
	t.Helper()
	require.Error(t, err)
	var se *commonModel.StatusError
	require.ErrorAs(t, err, &se)
	assert.Equal(t, wantStatus, se.Status)
}

func internalErr() error {
	return commonModel.Fail(http.StatusInternalServerError, "boom")
}

func TestSettingHandler_GetSettings(t *testing.T) {
	t.Run("success fills body", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().
			GetSetting(mock.Anything).
			Run(func(s *settingModel.SystemSetting) { s.SiteTitle = "MyEch0" }).
			Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetSettings(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "MyEch0", out.Body.SiteTitle)
	})

	t.Run("error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetSetting(mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetSettings(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestSettingHandler_GetOAuth2Status(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().
			GetOAuth2Status(mock.Anything).
			Run(func(s *settingModel.OAuth2Status) { s.Enabled = true; s.Provider = "github" }).
			Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetOAuth2Status(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, out.Body.Enabled)
		assert.Equal(t, "github", out.Body.Provider)
	})

	t.Run("error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetOAuth2Status(mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetOAuth2Status(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_GetPasskeyStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().
			GetPasskeyStatus(mock.Anything).
			Run(func(s *settingModel.PasskeyStatus) { s.PasskeyReady = true }).
			Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetPasskeyStatus(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, out.Body.PasskeyReady)
	})

	t.Run("error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetPasskeyStatus(mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetPasskeyStatus(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_GetAgentInfo_Scrubs(t *testing.T) {
	t.Run("success scrubs sensitive fields", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().
			GetAgentInfo(mock.Anything).
			Run(func(s *settingModel.AgentSetting) {
				s.Enable = true
				s.Protocol = "openai"
				s.Model = "gpt-4o"
				s.ApiKey = "sk-secret"
				s.Prompt = "internal prompt"
				s.BaseURL = "https://llm.internal"
			}).
			Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetAgentInfo(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, out.Body.Enable)
		assert.Equal(t, "openai", out.Body.Protocol)
		assert.Equal(t, "gpt-4o", out.Body.Model)
		assert.Empty(t, out.Body.ApiKey, "ApiKey 必须脱敏")
		assert.Empty(t, out.Body.Prompt, "Prompt 必须脱敏")
		assert.Empty(t, out.Body.BaseURL, "BaseURL 必须脱敏")
	})

	t.Run("error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetAgentInfo(mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetAgentInfo(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_UpdateSettings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateSetting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateSettings(context.Background(), &settingHandler.UpdateSettingsInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateSetting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateSettings(context.Background(), &settingHandler.UpdateSettingsInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestSettingHandler_S3(t *testing.T) {
	t.Run("get success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetS3Setting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetS3Settings(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("get error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetS3Setting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetS3Settings(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("update success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateS3Setting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateS3Settings(context.Background(), &settingHandler.S3SettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateS3Setting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdateS3Settings(context.Background(), &settingHandler.S3SettingInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("test connection success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().TestS3Connection(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.TestS3Connection(context.Background(), &settingHandler.S3SettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("test connection error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().TestS3Connection(mock.Anything, mock.Anything).
			Return(commonModel.Fail(http.StatusBadRequest, commonModel.S3_CONFIG_ERROR)).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.TestS3Connection(context.Background(), &settingHandler.S3SettingInput{})

		assertStatusErr(t, err, http.StatusBadRequest)
	})
}

func TestSettingHandler_OAuth2Settings(t *testing.T) {
	t.Run("get success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetOAuth2Setting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetOAuth2Settings(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("get error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetOAuth2Setting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetOAuth2Settings(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("update success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateOAuth2Setting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateOAuth2Settings(context.Background(), &settingHandler.OAuth2SettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateOAuth2Setting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdateOAuth2Settings(context.Background(), &settingHandler.OAuth2SettingInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_PasskeySettings(t *testing.T) {
	t.Run("get success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetPasskeySetting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetPasskeySettings(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("get error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetPasskeySetting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetPasskeySettings(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("update success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdatePasskeySetting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdatePasskeySettings(context.Background(), &settingHandler.PasskeySettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdatePasskeySetting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdatePasskeySettings(context.Background(), &settingHandler.PasskeySettingInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_Webhooks(t *testing.T) {
	t.Run("list success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		want := []webhookModel.Webhook{{ID: "w1", Name: "hook", URL: "https://h.example"}}
		svc.EXPECT().GetAllWebhooks(mock.Anything).Return(want, nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetWebhook(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body.Items)
	})

	t.Run("list error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetAllWebhooks(mock.Anything).Return(nil, internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetWebhook(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})

	t.Run("create success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().CreateWebhook(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.CreateWebhook(context.Background(), &settingHandler.WebhookInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("create error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().CreateWebhook(mock.Anything, mock.Anything).
			Return(commonModel.Fail(http.StatusBadRequest, commonModel.INVALID_WEBHOOK_URL)).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.CreateWebhook(context.Background(), &settingHandler.WebhookInput{})

		assertStatusErr(t, err, http.StatusBadRequest)
	})

	t.Run("update success passes id through", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateWebhook(mock.Anything, "w-9", mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateWebhook(context.Background(), &settingHandler.WebhookIDBodyInput{ID: "w-9"})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateWebhook(mock.Anything, mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdateWebhook(context.Background(), &settingHandler.WebhookIDBodyInput{ID: "x"})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("delete success passes id through", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().DeleteWebhook(mock.Anything, "w-3").Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.DeleteWebhook(context.Background(), &settingHandler.WebhookIDInput{ID: "w-3"})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("delete error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().DeleteWebhook(mock.Anything, mock.Anything).Return(errors.New("nope")).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.DeleteWebhook(context.Background(), &settingHandler.WebhookIDInput{ID: "x"})

		require.Error(t, err)
	})

	t.Run("test success passes id through", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().TestWebhook(mock.Anything, "w-7").Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.TestWebhook(context.Background(), &settingHandler.WebhookIDInput{ID: "w-7"})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("test error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().TestWebhook(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.TestWebhook(context.Background(), &settingHandler.WebhookIDInput{ID: "x"})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_SnapshotSchedule(t *testing.T) {
	t.Run("get success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().
			GetSnapshotScheduleSetting(mock.Anything).
			Run(func(s *settingModel.SnapshotSchedule) { s.Enable = true; s.CronExpression = "0 0 * * *" }).
			Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetSnapshotScheduleSetting(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.True(t, out.Body.Enable)
		assert.Equal(t, "0 0 * * *", out.Body.CronExpression)
	})

	t.Run("get error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetSnapshotScheduleSetting(mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetSnapshotScheduleSetting(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("update success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateSnapshotScheduleSetting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateSnapshotScheduleSetting(context.Background(), &settingHandler.SnapshotScheduleInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error invalid cron", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateSnapshotScheduleSetting(mock.Anything, mock.Anything).
			Return(commonModel.Fail(http.StatusBadRequest, commonModel.INVALID_CRON_EXPRESSION)).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdateSnapshotScheduleSetting(context.Background(), &settingHandler.SnapshotScheduleInput{})

		assertStatusErr(t, err, http.StatusBadRequest)
	})
}

func TestSettingHandler_AgentSettings(t *testing.T) {
	t.Run("get success keeps sensitive fields", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().
			GetAgentSettings(mock.Anything, mock.Anything).
			Run(func(_ context.Context, s *settingModel.AgentSetting) { s.ApiKey = "sk-admin"; s.Model = "gpt-4o" }).
			Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetAgentSettings(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "sk-admin", out.Body.ApiKey)
		assert.Equal(t, "gpt-4o", out.Body.Model)
	})

	t.Run("get error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetAgentSettings(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.GetAgentSettings(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("update success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateAgentSettings(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateAgentSettings(context.Background(), &settingHandler.AgentSettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateAgentSettings(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdateAgentSettings(context.Background(), &settingHandler.AgentSettingInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})

	t.Run("test connection success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().TestAgentConnection(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.TestAgentConnection(context.Background(), &settingHandler.AgentSettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("test connection error missing api key", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().TestAgentConnection(mock.Anything, mock.Anything).
			Return(commonModel.Fail(http.StatusBadRequest, commonModel.AGENT_API_KEY_MISSING)).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.TestAgentConnection(context.Background(), &settingHandler.AgentSettingInput{})

		assertStatusErr(t, err, http.StatusBadRequest)
	})
}

func TestSettingHandler_EmbeddingSettings(t *testing.T) {
	t.Run("get success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		want := settingModel.EmbeddingSetting{Enable: true, Model: "text-embedding-3-small", Dim: 1536}
		svc.EXPECT().GetEmbeddingSetting(mock.Anything).Return(want, nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetEmbeddingSettings(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body)
	})

	t.Run("get error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().GetEmbeddingSetting(mock.Anything).Return(settingModel.EmbeddingSetting{}, internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.GetEmbeddingSettings(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})

	t.Run("update success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateEmbeddingSetting(mock.Anything, mock.Anything).Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.UpdateEmbeddingSettings(context.Background(), &settingHandler.EmbeddingSettingInput{})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("update error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().UpdateEmbeddingSetting(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.UpdateEmbeddingSettings(context.Background(), &settingHandler.EmbeddingSettingInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}

func TestSettingHandler_AccessTokens(t *testing.T) {
	t.Run("list success", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		want := []settingModel.AccessTokenSetting{{ID: "t1", Name: "ci", TokenType: "access"}}
		svc.EXPECT().ListAccessTokens(mock.Anything).Return(want, nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.ListAccessTokens(context.Background(), &settingHandler.EmptyInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body.Items)
	})

	t.Run("list error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().ListAccessTokens(mock.Anything).Return(nil, internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.ListAccessTokens(context.Background(), &settingHandler.EmptyInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})

	t.Run("create returns plaintext token", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().CreateAccessToken(mock.Anything, mock.Anything).Return("plaintext-token-xyz", nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.CreateAccessToken(context.Background(), &settingHandler.AccessTokenInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "plaintext-token-xyz", out.Body)
	})

	t.Run("create error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().CreateAccessToken(mock.Anything, mock.Anything).Return("", internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.CreateAccessToken(context.Background(), &settingHandler.AccessTokenInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})

	t.Run("delete success passes id through", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().DeleteAccessToken(mock.Anything, "t-5").Return(nil).Once()

		h := settingHandler.NewSettingHandler(svc)
		out, err := h.DeleteAccessToken(context.Background(), &settingHandler.IDInput{ID: "t-5"})

		require.NoError(t, err)
		assert.Equal(t, &settingHandler.EmptyOutput{}, out)
	})

	t.Run("delete error", func(t *testing.T) {
		svc := settingmock.NewMockService(t)
		svc.EXPECT().DeleteAccessToken(mock.Anything, mock.Anything).Return(internalErr()).Once()

		h := settingHandler.NewSettingHandler(svc)
		_, err := h.DeleteAccessToken(context.Background(), &settingHandler.IDInput{ID: "x"})

		assertStatusErr(t, err, http.StatusInternalServerError)
	})
}
