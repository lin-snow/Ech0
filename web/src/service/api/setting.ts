// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetSettings() {
  return request<App.Api.Setting.SystemSetting>({
    url: '/settings',
    method: 'GET',
  })
}

export function fetchUpdateSettings(systemSetting: App.Api.Setting.SystemSetting) {
  return request({
    url: '/settings',
    method: 'PUT',
    data: systemSetting,
  })
}

export function fetchGetS3Settings() {
  return request<App.Api.Setting.S3Setting>({
    url: '/s3/settings',
    method: 'GET',
  })
}

export function fetchUpdateS3Settings(s3Setting: App.Api.Setting.S3Setting) {
  return request({
    url: '/s3/settings',
    method: 'PUT',
    data: s3Setting,
  })
}

export function fetchTestS3Connection(s3Setting: App.Api.Setting.S3Setting) {
  return request({
    url: '/s3/settings/test',
    method: 'POST',
    data: s3Setting,
    silentError: true,
  })
}

export function fetchGetOAuth2Settings() {
  return request<App.Api.Setting.OAuth2Setting>({
    url: '/oauth2/settings',
    method: 'GET',
  })
}

export function fetchUpdateOAuth2Settings(oauth2Setting: App.Api.Setting.OAuth2Setting) {
  return request({
    url: '/oauth2/settings',
    method: 'PUT',
    data: oauth2Setting,
  })
}

export function fetchGetOAuth2Status() {
  return request<App.Api.Setting.OAuth2Status>({
    url: '/oauth2/status',
    method: 'GET',
  })
}

export function fetchGetPasskeyStatus() {
  return request<App.Api.Setting.PasskeyStatus>({
    url: '/passkey/status',
    method: 'GET',
  })
}

export function fetchGetPasskeySettings() {
  return request<App.Api.Setting.PasskeySetting>({
    url: '/passkey/settings',
    method: 'GET',
  })
}

export function fetchUpdatePasskeySettings(passkeySetting: App.Api.Setting.PasskeySetting) {
  return request({
    url: '/passkey/settings',
    method: 'PUT',
    data: passkeySetting,
  })
}

export function fetchGetOAuthInfo(provider?: string) {
  return request<App.Api.Setting.OAuthInfo>({
    url: '/oauth/info?' + (provider ? `provider=${encodeURIComponent(provider)}` : ''),
    method: 'GET',
  })
}

export function fetchGetAllWebhooks() {
  return request<App.Api.Setting.Webhook[]>({
    url: '/webhook',
    method: 'GET',
  })
}

export function fetchCreateWebhook(webhook: App.Api.Setting.WebhookDto) {
  return request({
    url: '/webhook',
    method: 'POST',
    data: webhook,
  })
}

export function fetchUpdateWebhook(webhookId: string, webhook: App.Api.Setting.WebhookDto) {
  return request({
    url: `/webhook/${webhookId}`,
    method: 'PUT',
    data: webhook,
  })
}

export function fetchDeleteWebhook(webhookId: string) {
  return request({
    url: `/webhook/${webhookId}`,
    method: 'DELETE',
  })
}

export function fetchTestWebhook(webhookId: string) {
  return request({
    url: `/webhook/${webhookId}/test`,
    method: 'POST',
  })
}

export function fetchListAccessTokens() {
  return request<App.Api.Setting.AccessToken[]>({
    url: '/access-tokens',
    method: 'GET',
  })
}

export function fetchCreateAccessToken(dto: App.Api.Setting.AccessTokenDto) {
  return request<string>({
    url: '/access-tokens',
    method: 'POST',
    data: dto,
  })
}

export function fetchDeleteAccessToken(tokenId: string) {
  return request({
    url: `/access-tokens/${tokenId}`,
    method: 'DELETE',
  })
}

export function fetchGetSnapshotScheduleSetting() {
  return request<App.Api.Setting.SnapshotSchedule>({
    url: '/snapshot/schedule',
    method: 'GET',
  })
}

export function fetchUpdateSnapshotScheduleSetting(
  snapshotSchedule: App.Api.Setting.SnapshotScheduleDto,
) {
  return request({
    url: '/snapshot/schedule',
    method: 'POST',
    data: snapshotSchedule,
  })
}

export function fetchGetAgentInfo() {
  return request<App.Api.Setting.AgentSetting>({
    url: '/agent/info',
    method: 'GET',
  })
}

export function fetchGetAgentSettings() {
  return request<App.Api.Setting.AgentSetting>({
    url: '/agent/settings',
    method: 'GET',
  })
}

export function fetchUpdateAgentSettings(agentSetting: App.Api.Setting.AgentSettingDto) {
  return request({
    url: '/agent/settings',
    method: 'PUT',
    data: agentSetting,
  })
}

export function fetchTestAgentConnection(agentSetting: App.Api.Setting.AgentSettingDto) {
  return request({
    url: '/agent/settings/test',
    method: 'POST',
    data: agentSetting,
    silentError: true,
  })
}
