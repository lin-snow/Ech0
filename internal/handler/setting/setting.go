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
	WebhookIDInput      struct {
		ID string `path:"id" format:"uuid" doc:"Webhook ID (UUID)"`
	}
	WebhookIDBodyInput struct {
		ID   string `path:"id" format:"uuid" doc:"Webhook ID (UUID)"`
		Body model.WebhookDto
	}
	AccessTokenInput      struct{ Body model.AccessTokenSettingDto }
	SnapshotScheduleInput struct{ Body model.SnapshotScheduleDto }
	AgentSettingInput     struct{ Body model.AgentSettingDto }
	EmbeddingSettingInput struct{ Body model.EmbeddingSettingDto }
)

type (
	SystemSettingOutput  struct{ Body model.SystemSetting }
	OAuth2StatusOutput   struct{ Body model.OAuth2Status }
	PasskeyStatusOutput  struct{ Body model.PasskeyStatus }
	AgentSettingOutput   struct{ Body model.AgentSetting }
	S3SettingOutput      struct{ Body model.S3Setting }
	OAuth2SettingOutput  struct{ Body model.OAuth2Setting }
	PasskeySettingOutput struct{ Body model.PasskeySetting }
	WebhookListOutput    struct {
		Body commonModel.List[webhookModel.Webhook]
	}
	SnapshotScheduleOutput struct{ Body model.SnapshotSchedule }
	EmbeddingSettingOutput struct{ Body model.EmbeddingSetting }
	AccessTokenListOutput  struct {
		Body commonModel.List[model.AccessTokenSetting]
	}
	StringOutput struct{ Body string }
	EmptyOutput  = struct{}
)

func (h *SettingHandler) GetSettings(_ context.Context, _ *EmptyInput) (*SystemSettingOutput, error) {
	var settings model.SystemSetting
	if err := h.settingService.GetSetting(&settings); err != nil {
		return nil, err
	}
	return &SystemSettingOutput{Body: settings}, nil
}

func (h *SettingHandler) GetOAuth2Status(_ context.Context, _ *EmptyInput) (*OAuth2StatusOutput, error) {
	var status model.OAuth2Status
	if err := h.settingService.GetOAuth2Status(&status); err != nil {
		return nil, err
	}
	return &OAuth2StatusOutput{Body: status}, nil
}

func (h *SettingHandler) GetPasskeyStatus(_ context.Context, _ *EmptyInput) (*PasskeyStatusOutput, error) {
	var status model.PasskeyStatus
	if err := h.settingService.GetPasskeyStatus(&status); err != nil {
		return nil, err
	}
	return &PasskeyStatusOutput{Body: status}, nil
}

func (h *SettingHandler) GetAgentInfo(_ context.Context, _ *EmptyInput) (*AgentSettingOutput, error) {
	var settings model.AgentSetting
	if err := h.settingService.GetAgentInfo(&settings); err != nil {
		return nil, err
	}
	settings.ApiKey = ""
	settings.Prompt = ""
	settings.BaseURL = ""
	return &AgentSettingOutput{Body: settings}, nil
}

func (h *SettingHandler) UpdateSettings(ctx context.Context, in *UpdateSettingsInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateSetting(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetS3Settings(ctx context.Context, _ *EmptyInput) (*S3SettingOutput, error) {
	var s3Setting model.S3Setting
	if err := h.settingService.GetS3Setting(ctx, &s3Setting); err != nil {
		return nil, err
	}
	return &S3SettingOutput{Body: s3Setting}, nil
}

func (h *SettingHandler) UpdateS3Settings(ctx context.Context, in *S3SettingInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateS3Setting(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) TestS3Connection(ctx context.Context, in *S3SettingInput) (*EmptyOutput, error) {
	if err := h.settingService.TestS3Connection(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetOAuth2Settings(ctx context.Context, _ *EmptyInput) (*OAuth2SettingOutput, error) {
	var oauthSetting model.OAuth2Setting
	if err := h.settingService.GetOAuth2Setting(ctx, &oauthSetting); err != nil {
		return nil, err
	}
	return &OAuth2SettingOutput{Body: oauthSetting}, nil
}

func (h *SettingHandler) UpdateOAuth2Settings(ctx context.Context, in *OAuth2SettingInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateOAuth2Setting(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetPasskeySettings(ctx context.Context, _ *EmptyInput) (*PasskeySettingOutput, error) {
	var passkeySetting model.PasskeySetting
	if err := h.settingService.GetPasskeySetting(ctx, &passkeySetting); err != nil {
		return nil, err
	}
	return &PasskeySettingOutput{Body: passkeySetting}, nil
}

func (h *SettingHandler) UpdatePasskeySettings(ctx context.Context, in *PasskeySettingInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdatePasskeySetting(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetWebhook(ctx context.Context, _ *EmptyInput) (*WebhookListOutput, error) {
	result, err := h.settingService.GetAllWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	return &WebhookListOutput{Body: commonModel.List[webhookModel.Webhook]{Items: result}}, nil
}

func (h *SettingHandler) CreateWebhook(ctx context.Context, in *WebhookInput) (*EmptyOutput, error) {
	if err := h.settingService.CreateWebhook(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) UpdateWebhook(ctx context.Context, in *WebhookIDBodyInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateWebhook(ctx, in.ID, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) DeleteWebhook(ctx context.Context, in *WebhookIDInput) (*EmptyOutput, error) {
	if err := h.settingService.DeleteWebhook(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) TestWebhook(ctx context.Context, in *WebhookIDInput) (*EmptyOutput, error) {
	if err := h.settingService.TestWebhook(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetSnapshotScheduleSetting(_ context.Context, _ *EmptyInput) (*SnapshotScheduleOutput, error) {
	var snapshotSchedule model.SnapshotSchedule
	if err := h.settingService.GetSnapshotScheduleSetting(&snapshotSchedule); err != nil {
		return nil, err
	}
	return &SnapshotScheduleOutput{Body: snapshotSchedule}, nil
}

func (h *SettingHandler) UpdateSnapshotScheduleSetting(ctx context.Context, in *SnapshotScheduleInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateSnapshotScheduleSetting(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetAgentSettings(ctx context.Context, _ *EmptyInput) (*AgentSettingOutput, error) {
	var settings model.AgentSetting
	if err := h.settingService.GetAgentSettings(ctx, &settings); err != nil {
		return nil, err
	}
	return &AgentSettingOutput{Body: settings}, nil
}

func (h *SettingHandler) UpdateAgentSettings(ctx context.Context, in *AgentSettingInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateAgentSettings(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) TestAgentConnection(ctx context.Context, in *AgentSettingInput) (*EmptyOutput, error) {
	if err := h.settingService.TestAgentConnection(ctx, &in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) GetEmbeddingSettings(ctx context.Context, _ *EmptyInput) (*EmbeddingSettingOutput, error) {
	setting, err := h.settingService.GetEmbeddingSetting(ctx)
	if err != nil {
		return nil, err
	}
	return &EmbeddingSettingOutput{Body: setting}, nil
}

func (h *SettingHandler) UpdateEmbeddingSettings(ctx context.Context, in *EmbeddingSettingInput) (*EmptyOutput, error) {
	if err := h.settingService.UpdateEmbeddingSetting(ctx, in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *SettingHandler) ListAccessTokens(ctx context.Context, _ *EmptyInput) (*AccessTokenListOutput, error) {
	result, err := h.settingService.ListAccessTokens(ctx)
	if err != nil {
		return nil, err
	}
	return &AccessTokenListOutput{Body: commonModel.List[model.AccessTokenSetting]{Items: result}}, nil
}

func (h *SettingHandler) CreateAccessToken(ctx context.Context, in *AccessTokenInput) (*StringOutput, error) {
	createdToken, err := h.settingService.CreateAccessToken(ctx, &in.Body)
	if err != nil {
		return nil, err
	}
	return &StringOutput{Body: createdToken}, nil
}

func (h *SettingHandler) DeleteAccessToken(ctx context.Context, in *IDInput) (*EmptyOutput, error) {
	if err := h.settingService.DeleteAccessToken(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}
