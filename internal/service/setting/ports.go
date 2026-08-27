// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"time"

	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	fileService "github.com/lin-snow/ech0/internal/service/file"
)

type Service interface {
	GetSetting(setting *model.SystemSetting) error
	UpdateSetting(ctx context.Context, newSetting *model.SystemSettingDto) error
	BootstrapDefaultLocale(ctx context.Context, locale string) error
	GetS3Setting(ctx context.Context, setting *model.S3Setting) error
	UpdateS3Setting(ctx context.Context, newSetting *model.S3SettingDto) error
	TestS3Connection(ctx context.Context, newSetting *model.S3SettingDto) error
	GetOAuth2Setting(ctx context.Context, setting *model.OAuth2Setting) error
	UpdateOAuth2Setting(ctx context.Context, newSetting *model.OAuth2SettingDto) error
	GetOAuth2Status(status *model.OAuth2Status) error
	GetPasskeySetting(ctx context.Context, setting *model.PasskeySetting) error
	UpdatePasskeySetting(ctx context.Context, newSetting *model.PasskeySettingDto) error
	GetPasskeyStatus(status *model.PasskeyStatus) error
	GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	DeleteWebhook(ctx context.Context, id string) error
	UpdateWebhook(ctx context.Context, id string, newWebhook *model.WebhookDto) error
	CreateWebhook(ctx context.Context, newWebhook *model.WebhookDto) error
	TestWebhook(ctx context.Context, id string) error
	ListAccessTokens(ctx context.Context) ([]model.AccessTokenSetting, error)
	CreateAccessToken(ctx context.Context, newToken *model.AccessTokenSettingDto) (string, error)
	DeleteAccessToken(ctx context.Context, id string) error
	GetSnapshotScheduleSetting(setting *model.SnapshotSchedule) error
	UpdateSnapshotScheduleSetting(ctx context.Context, newSetting *model.SnapshotScheduleDto) error
	GetAgentInfo(setting *model.AgentSetting) error
	GetAgentSettings(ctx context.Context, setting *model.AgentSetting) error
	UpdateAgentSettings(ctx context.Context, newSetting *model.AgentSettingDto) error
	TestAgentConnection(ctx context.Context, newSetting *model.AgentSettingDto) error
	GetEmbeddingSetting(ctx context.Context) (model.EmbeddingSetting, error)
	UpdateEmbeddingSetting(ctx context.Context, dto model.EmbeddingSettingDto) error
}

type (
	CommonService = commonService.Service
	FileService   = fileService.Service
)

type SettingRepository interface {
	ListAccessTokens(ctx context.Context, userID string) ([]model.AccessTokenSetting, error)
	CreateAccessToken(ctx context.Context, token *model.AccessTokenSetting) error
	GetAccessTokenByID(ctx context.Context, id string) (model.AccessTokenSetting, error)
	DeleteAccessTokenByID(ctx context.Context, id string) error
}

type TokenRevoker interface {
	RevokeToken(jti string, remainTTL time.Duration)
}

type WebhookRepository interface {
	GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	GetWebhookByID(ctx context.Context, id string) (*webhookModel.Webhook, error)
	CreateWebhook(ctx context.Context, webhook *webhookModel.Webhook) error
	UpdateWebhookByID(ctx context.Context, id string, webhook *webhookModel.Webhook) error
	UpdateWebhookDeliveryStatus(ctx context.Context, id string, status string, lastTrigger int64) error
	DeleteWebhookByID(ctx context.Context, id string) error
}
