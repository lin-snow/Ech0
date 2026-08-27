// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func registerSetting(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	adminSettings := secured(revoker, authModel.ScopeAdminSettings)
	adminToken := secured(revoker, authModel.ScopeAdminToken)

	route(api, public(), huma.Operation{
		OperationID: "settings-get",
		Method:      http.MethodGet,
		Path:        "/settings",
		Summary:     "获取系统全局设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetSettings)

	route(api, public(), huma.Operation{
		OperationID: "oauth2-status",
		Method:      http.MethodGet,
		Path:        "/oauth2/status",
		Summary:     "获取 OAuth2 状态",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetOAuth2Status)

	route(api, public(), huma.Operation{
		OperationID: "passkey-status",
		Method:      http.MethodGet,
		Path:        "/passkey/status",
		Summary:     "获取 Passkey 状态",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetPasskeyStatus)

	route(api, public(), huma.Operation{
		OperationID: "agent-info",
		Method:      http.MethodGet,
		Path:        "/agent/info",
		Summary:     "获取 Agent 公开信息",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetAgentInfo)

	route(api, adminSettings, huma.Operation{
		OperationID: "settings-update",
		Method:      http.MethodPut,
		Path:        "/settings",
		Summary:     "更新系统全局设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateSettings)

	route(api, adminSettings, huma.Operation{
		OperationID: "s3-get",
		Method:      http.MethodGet,
		Path:        "/s3/settings",
		Summary:     "获取 S3 存储设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetS3Settings)

	route(api, adminSettings, huma.Operation{
		OperationID: "s3-update",
		Method:      http.MethodPut,
		Path:        "/s3/settings",
		Summary:     "更新 S3 存储设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateS3Settings)

	route(api, adminSettings, huma.Operation{
		OperationID: "s3-test",
		Method:      http.MethodPost,
		Path:        "/s3/settings/test",
		Summary:     "测试 S3 存储连接",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.TestS3Connection)

	route(api, adminSettings, huma.Operation{
		OperationID: "oauth2-get",
		Method:      http.MethodGet,
		Path:        "/oauth2/settings",
		Summary:     "获取 OAuth2 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetOAuth2Settings)

	route(api, adminSettings, huma.Operation{
		OperationID: "oauth2-update",
		Method:      http.MethodPut,
		Path:        "/oauth2/settings",
		Summary:     "更新 OAuth2 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateOAuth2Settings)

	route(api, adminSettings, huma.Operation{
		OperationID: "passkey-get",
		Method:      http.MethodGet,
		Path:        "/passkey/settings",
		Summary:     "获取 Passkey 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetPasskeySettings)

	route(api, adminSettings, huma.Operation{
		OperationID: "passkey-update",
		Method:      http.MethodPut,
		Path:        "/passkey/settings",
		Summary:     "更新 Passkey 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdatePasskeySettings)

	route(api, adminSettings, huma.Operation{
		OperationID: "webhook-list",
		Method:      http.MethodGet,
		Path:        "/webhook",
		Summary:     "获取所有 Webhook",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetWebhook)

	route(api, adminSettings, huma.Operation{
		OperationID: "webhook-create",
		Method:      http.MethodPost,
		Path:        "/webhook",
		Summary:     "创建 Webhook",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.CreateWebhook)

	route(api, adminSettings, huma.Operation{
		OperationID: "webhook-update",
		Method:      http.MethodPut,
		Path:        "/webhook/{id}",
		Summary:     "更新 Webhook",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateWebhook)

	route(api, adminSettings, huma.Operation{
		OperationID: "webhook-delete",
		Method:      http.MethodDelete,
		Path:        "/webhook/{id}",
		Summary:     "删除 Webhook",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.DeleteWebhook)

	route(api, adminSettings, huma.Operation{
		OperationID: "webhook-test",
		Method:      http.MethodPost,
		Path:        "/webhook/{id}/test",
		Summary:     "测试 Webhook",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.TestWebhook)

	route(api, adminSettings, huma.Operation{
		OperationID: "snapshot-schedule-get",
		Method:      http.MethodGet,
		Path:        "/snapshot/schedule",
		Summary:     "获取定时快照计划",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetSnapshotScheduleSetting)

	route(api, adminSettings, huma.Operation{
		OperationID: "snapshot-schedule-update",
		Method:      http.MethodPost,
		Path:        "/snapshot/schedule",
		Summary:     "设置定时快照计划",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateSnapshotScheduleSetting)

	route(api, adminSettings, huma.Operation{
		OperationID: "agent-settings-get",
		Method:      http.MethodGet,
		Path:        "/agent/settings",
		Summary:     "获取 Agent 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetAgentSettings)

	route(api, adminSettings, huma.Operation{
		OperationID: "agent-settings-update",
		Method:      http.MethodPut,
		Path:        "/agent/settings",
		Summary:     "更新 Agent 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateAgentSettings)

	route(api, adminSettings, huma.Operation{
		OperationID: "agent-settings-test",
		Method:      http.MethodPost,
		Path:        "/agent/settings/test",
		Summary:     "测试 Copilot 连接",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.TestAgentConnection)

	route(api, adminSettings, huma.Operation{
		OperationID: "embedding-settings-get",
		Method:      http.MethodGet,
		Path:        "/embedding/settings",
		Summary:     "获取 Embedding 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.GetEmbeddingSettings)

	route(api, adminSettings, huma.Operation{
		OperationID: "embedding-settings-update",
		Method:      http.MethodPut,
		Path:        "/embedding/settings",
		Summary:     "更新 Embedding 设置",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.UpdateEmbeddingSettings)

	route(api, adminToken, huma.Operation{
		OperationID: "access-token-list",
		Method:      http.MethodGet,
		Path:        "/access-tokens",
		Summary:     "列出访问令牌",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.ListAccessTokens)

	route(api, adminToken, huma.Operation{
		OperationID: "access-token-create",
		Method:      http.MethodPost,
		Path:        "/access-tokens",
		Summary:     "创建访问令牌",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.CreateAccessToken)

	route(api, adminToken, huma.Operation{
		OperationID: "access-token-delete",
		Method:      http.MethodDelete,
		Path:        "/access-tokens/{id}",
		Summary:     "删除访问令牌",
		Tags:        []string{"Setting"},
	}, h.SettingHandler.DeleteAccessToken)
}
