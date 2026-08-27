// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	service "github.com/lin-snow/ech0/internal/service/setting"
)

type SettingHandler struct {
	settingService service.Service
}

func NewSettingHandler(settingService service.Service) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
	}
}

type (
	EmptyInput struct{}
	IDInput    struct {
		ID string `path:"id" format:"uuid" doc:"资源 ID（UUID）"`
	}
	UpdateSettingsInput struct{ Body model.SystemSettingDto }
	S3SettingInput      struct{ Body model.S3SettingDto }
	OAuth2SettingInput  struct{ Body model.OAuth2SettingDto }
	PasskeySettingInput struct{ Body model.PasskeySettingDto }
	WebhookInput        struct{ Body model.WebhookDto }
	WebhookIDBodyInput  struct {
		ID   string `path:"id" format:"uuid" doc:"Webhook ID（UUID）"`
		Body model.WebhookDto
	}
	AccessTokenInput      struct{ Body model.AccessTokenSettingDto }
	SnapshotScheduleInput struct{ Body model.SnapshotScheduleDto }
	AgentSettingInput     struct{ Body model.AgentSettingDto }
	EmbeddingSettingInput struct{ Body model.EmbeddingSettingDto }
)

type (
	SystemSettingOutput    = commonModel.Result[model.SystemSetting]
	OAuth2StatusOutput     = commonModel.Result[model.OAuth2Status]
	PasskeyStatusOutput    = commonModel.Result[model.PasskeyStatus]
	AgentSettingOutput     = commonModel.Result[model.AgentSetting]
	S3SettingOutput        = commonModel.Result[model.S3Setting]
	OAuth2SettingOutput    = commonModel.Result[model.OAuth2Setting]
	PasskeySettingOutput   = commonModel.Result[model.PasskeySetting]
	WebhookListOutput      = commonModel.Result[[]webhookModel.Webhook]
	SnapshotScheduleOutput = commonModel.Result[model.SnapshotSchedule]
	EmbeddingSettingOutput = commonModel.Result[model.EmbeddingSetting]
	AccessTokenListOutput  = commonModel.Result[[]model.AccessTokenSetting]
	StringOutput           = commonModel.Result[string]
	EmptyOutput            = commonModel.Result[any]
)

func (h *SettingHandler) GetSettings(ctx context.Context, _ *EmptyInput) (SystemSettingOutput, error) {
	var settings model.SystemSetting
	if err := h.settingService.GetSetting(&settings); err != nil {
		return SystemSettingOutput{}, err
	}
	return commonModel.OK(settings, commonModel.GET_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) GetOAuth2Status(ctx context.Context, _ *EmptyInput) (OAuth2StatusOutput, error) {
	var status model.OAuth2Status
	if err := h.settingService.GetOAuth2Status(&status); err != nil {
		return OAuth2StatusOutput{}, err
	}
	return commonModel.OK(status, commonModel.GET_OAUTH2_STATUS_SUCCESS), nil
}

func (h *SettingHandler) GetPasskeyStatus(ctx context.Context, _ *EmptyInput) (PasskeyStatusOutput, error) {
	var status model.PasskeyStatus
	if err := h.settingService.GetPasskeyStatus(&status); err != nil {
		return PasskeyStatusOutput{}, err
	}
	return commonModel.OK(status, commonModel.GET_PASSKEY_STATUS_SUCCESS), nil
}

func (h *SettingHandler) GetAgentInfo(ctx context.Context, _ *EmptyInput) (AgentSettingOutput, error) {
	var settings model.AgentSetting
	if err := h.settingService.GetAgentInfo(&settings); err != nil {
		return AgentSettingOutput{}, err
	}
	settings.ApiKey = ""
	settings.Prompt = ""
	settings.BaseURL = ""
	return commonModel.OK(settings, commonModel.GET_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdateSettings(ctx context.Context, in *UpdateSettingsInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateSetting(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) GetS3Settings(ctx context.Context, _ *EmptyInput) (S3SettingOutput, error) {
	var s3Setting model.S3Setting
	if err := h.settingService.GetS3Setting(ctx, &s3Setting); err != nil {
		return S3SettingOutput{}, err
	}
	return commonModel.OK(s3Setting, commonModel.GET_S3_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdateS3Settings(ctx context.Context, in *S3SettingInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateS3Setting(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_S3_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) TestS3Connection(ctx context.Context, in *S3SettingInput) (EmptyOutput, error) {
	if err := h.settingService.TestS3Connection(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.TEST_S3_CONNECTION_SUCCESS), nil
}

func (h *SettingHandler) GetOAuth2Settings(ctx context.Context, _ *EmptyInput) (OAuth2SettingOutput, error) {
	var oauthSetting model.OAuth2Setting
	if err := h.settingService.GetOAuth2Setting(ctx, &oauthSetting); err != nil {
		return OAuth2SettingOutput{}, err
	}
	return commonModel.OK(oauthSetting, commonModel.GET_OAUTH_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdateOAuth2Settings(ctx context.Context, in *OAuth2SettingInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateOAuth2Setting(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_OAUTH_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) GetPasskeySettings(ctx context.Context, _ *EmptyInput) (PasskeySettingOutput, error) {
	var passkeySetting model.PasskeySetting
	if err := h.settingService.GetPasskeySetting(ctx, &passkeySetting); err != nil {
		return PasskeySettingOutput{}, err
	}
	return commonModel.OK(passkeySetting, commonModel.GET_PASSKEY_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdatePasskeySettings(ctx context.Context, in *PasskeySettingInput) (EmptyOutput, error) {
	if err := h.settingService.UpdatePasskeySetting(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_PASSKEY_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) GetWebhook(ctx context.Context, _ *EmptyInput) (WebhookListOutput, error) {
	result, err := h.settingService.GetAllWebhooks(ctx)
	if err != nil {
		return WebhookListOutput{}, err
	}
	return commonModel.OK(result, commonModel.GET_WEBHOOK_SUCCESS), nil
}

func (h *SettingHandler) CreateWebhook(ctx context.Context, in *WebhookInput) (EmptyOutput, error) {
	if err := h.settingService.CreateWebhook(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.CREATE_WEBHOOK_SUCCESS), nil
}

func (h *SettingHandler) UpdateWebhook(ctx context.Context, in *WebhookIDBodyInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateWebhook(ctx, in.ID, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_WEBHOOK_SUCCESS), nil
}

func (h *SettingHandler) DeleteWebhook(ctx context.Context, in *IDInput) (EmptyOutput, error) {
	if err := h.settingService.DeleteWebhook(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_WEBHOOK_SUCCESS), nil
}

func (h *SettingHandler) TestWebhook(ctx context.Context, in *IDInput) (EmptyOutput, error) {
	if err := h.settingService.TestWebhook(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.TEST_WEBHOOK_SUCCESS), nil
}

func (h *SettingHandler) GetSnapshotScheduleSetting(ctx context.Context, _ *EmptyInput) (SnapshotScheduleOutput, error) {
	var snapshotSchedule model.SnapshotSchedule
	if err := h.settingService.GetSnapshotScheduleSetting(&snapshotSchedule); err != nil {
		return SnapshotScheduleOutput{}, err
	}
	return commonModel.OK(snapshotSchedule, commonModel.GET_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdateSnapshotScheduleSetting(ctx context.Context, in *SnapshotScheduleInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateSnapshotScheduleSetting(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.SCHEDULE_SNAPSHOT_SUCCESS), nil
}

func (h *SettingHandler) GetAgentSettings(ctx context.Context, _ *EmptyInput) (AgentSettingOutput, error) {
	var settings model.AgentSetting
	if err := h.settingService.GetAgentSettings(ctx, &settings); err != nil {
		return AgentSettingOutput{}, err
	}
	return commonModel.OK(settings, commonModel.GET_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdateAgentSettings(ctx context.Context, in *AgentSettingInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateAgentSettings(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) TestAgentConnection(ctx context.Context, in *AgentSettingInput) (EmptyOutput, error) {
	if err := h.settingService.TestAgentConnection(ctx, &in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.AGENT_TEST_CONNECTION_SUCCESS), nil
}

func (h *SettingHandler) GetEmbeddingSettings(ctx context.Context, _ *EmptyInput) (EmbeddingSettingOutput, error) {
	setting, err := h.settingService.GetEmbeddingSetting(ctx)
	if err != nil {
		return EmbeddingSettingOutput{}, err
	}
	return commonModel.OK(setting, commonModel.GET_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) UpdateEmbeddingSettings(ctx context.Context, in *EmbeddingSettingInput) (EmptyOutput, error) {
	if err := h.settingService.UpdateEmbeddingSetting(ctx, in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

func (h *SettingHandler) ListAccessTokens(ctx context.Context, _ *EmptyInput) (AccessTokenListOutput, error) {
	result, err := h.settingService.ListAccessTokens(ctx)
	if err != nil {
		return AccessTokenListOutput{}, err
	}
	return commonModel.OK(result, commonModel.LIST_ACCESS_TOKENS_SUCCESS), nil
}

func (h *SettingHandler) CreateAccessToken(ctx context.Context, in *AccessTokenInput) (StringOutput, error) {
	createdToken, err := h.settingService.CreateAccessToken(ctx, &in.Body)
	if err != nil {
		return StringOutput{}, err
	}
	return commonModel.OK(createdToken, commonModel.CREATE_ACCESS_TOKEN_SUCCESS), nil
}

func (h *SettingHandler) DeleteAccessToken(ctx context.Context, in *IDInput) (EmptyOutput, error) {
	if err := h.settingService.DeleteAccessToken(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_ACCESS_TOKEN_SUCCESS), nil
}
