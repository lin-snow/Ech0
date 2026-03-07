import { request } from '../request'

// 获取系统设置
export function fetchGetSettings() {
  return request<App.Api.Setting.SystemSetting>({
    url: '/settings',
    method: 'GET',
  })
}

// 获取评论设置
export function fetchGetCommentSettings() {
  return request<App.Api.Setting.CommentSetting>({
    url: '/comment/settings',
    method: 'GET',
  })
}

// 更新系统设置
export function fetchUpdateSettings(systemSetting: App.Api.Setting.SystemSetting) {
  return request({
    url: '/settings',
    method: 'PUT',
    data: systemSetting,
  })
}

// 更新评论设置
export function fetchUpdateCommentSettings(commentSetting: App.Api.Setting.CommentSetting) {
  return request({
    url: '/comment/settings',
    method: 'PUT',
    data: commentSetting,
  })
}

// 获取 S3 存储设置
export function fetchGetS3Settings() {
  return request<App.Api.Setting.S3Setting>({
    url: '/s3/settings',
    method: 'GET',
  })
}

// 更新 S3 存储设置
export function fetchUpdateS3Settings(s3Setting: App.Api.Setting.S3Setting) {
  return request({
    url: '/s3/settings',
    method: 'PUT',
    data: s3Setting,
  })
}

// 获取 OAuth2 设置
export function fetchGetOAuth2Settings() {
  return request<App.Api.Setting.OAuth2Setting>({
    url: '/oauth2/settings',
    method: 'GET',
  })
}

// 更新 OAuth2 设置
export function fetchUpdateOAuth2Settings(oauth2Setting: App.Api.Setting.OAuth2Setting) {
  return request({
    url: '/oauth2/settings',
    method: 'PUT',
    data: oauth2Setting,
  })
}

// 获取 OAuth2 状态
export function fetchGetOAuth2Status() {
  return request<App.Api.Setting.OAuth2Status>({
    url: '/oauth2/status',
    method: 'GET',
  })
}

// 获取 OAuth2 绑定信息
export function fetchGetOAuthInfo(provider?: string) {
  return request<App.Api.Setting.OAuthInfo>({
    url: '/oauth/info?' + (provider ? `provider=${encodeURIComponent(provider)}` : ''),
    method: 'GET',
  })
}

// 获取 Webhook 列表
export function fetchGetAllWebhooks() {
  return request<App.Api.Setting.Webhook[]>({
    url: '/webhook',
    method: 'GET',
  })
}

// 创建 Webhook
export function fetchCreateWebhook(webhook: App.Api.Setting.WebhookDto) {
  return request({
    url: '/webhook',
    method: 'POST',
    data: webhook,
  })
}

// 更新 Webhook
export function fetchUpdateWebhook(webhookId: number, webhook: App.Api.Setting.WebhookDto) {
  return request({
    url: `/webhook/${webhookId}`,
    method: 'PUT',
    data: webhook,
  })
}

// 删除 Webhook
export function fetchDeleteWebhook(webhookId: number) {
  return request({
    url: `/webhook/${webhookId}`,
    method: 'DELETE',
  })
}

// 列出访问令牌
export function fetchListAccessTokens() {
  return request<App.Api.Setting.AccessToken[]>({
    url: '/access-tokens',
    method: 'GET',
  })
}

// 创建访问令牌
export function fetchCreateAccessToken(dto: App.Api.Setting.AccessTokenDto) {
  return request<string>({
    url: '/access-tokens',
    method: 'POST',
    data: dto,
  })
}

// 删除访问令牌
export function fetchDeleteAccessToken(tokenId: number) {
  return request({
    url: `/access-tokens/${tokenId}`,
    method: 'DELETE',
  })
}

// 获取备份计划
export function fetchGetBackupScheduleSetting() {
  return request<App.Api.Setting.BackupSchedule>({
    url: '/backup/schedule',
    method: 'GET',
  })
}

// 更新备份计划
export function fetchUpdateBackupScheduleSetting(
  backupSchedule: App.Api.Setting.BackupScheduleDto,
) {
  return request({
    url: '/backup/schedule',
    method: 'POST',
    data: backupSchedule,
  })
}

// 获取LLM Agent信息(无需鉴权)
export function fetchGetAgentInfo() {
  return request<App.Api.Setting.AgentSetting>({
    url: '/agent/info',
    method: 'GET',
  })
}

// 获取LLM Agent设置
export function fetchGetAgentSettings() {
  return request<App.Api.Setting.AgentSetting>({
    url: '/agent/settings',
    method: 'GET',
  })
}

// 更新LLM Agent设置
export function fetchUpdateAgentSettings(agentSetting: App.Api.Setting.AgentSettingDto) {
  return request({
    url: '/agent/settings',
    method: 'PUT',
    data: agentSetting,
  })
}
