// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { isAdminRole } from '@/utils/role'
import { ref } from 'vue'
import {
  fetchGetSettings,
  fetchGetS3Settings,
  fetchGetOAuth2Settings,
  fetchGetAllWebhooks,
  fetchListAccessTokens,
  fetchGetSnapshotScheduleSetting,
  fetchStartExport,
  fetchGetExportStatus,
  fetchGetAgentSettings,
  fetchGetAgentInfo,
  fetchHelloEch0,
} from '@/service/api'
import type { ExportFormat, ExportStatusPayload } from '@/service/api'
import { listItems } from '@/utils/list'
import { S3Provider, OAuth2Provider, AgentProtocol } from '@/enums/enums'
import { useUserStore } from './user'

const SNAPSHOT_STATUS_POLL_INTERVAL_MS = 3000
type SnapshotUIStatus = 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'

export const useSettingStore = defineStore('settingStore', () => {
  const userStore = useUserStore()

  const SystemSetting = ref<App.Api.Setting.SystemSetting>({
    site_title: import.meta.env.VITE_APP_TITLE,
    server_logo: '/Ech0.svg',
    server_logo_file_id: '',
    server_name: import.meta.env.VITE_APP_NAME,
    server_url: '',
    allow_register: true,
    default_locale: 'zh-CN',
    ICP_number: '',
    footer_content: '',
    footer_link: '',
    meting_api: '',
    custom_css: '',
    custom_js: '',
    privacy_policy: '',
    show_privacy_footer: false,
    enable_agent_discovery: true,
  })
  const S3Setting = ref<App.Api.Setting.S3Setting>({
    enable: false,
    provider: S3Provider.AWS,
    endpoint: '',
    access_key: '',
    secret_key: '',
    bucket_name: '',
    region: '',
    use_ssl: false,
    cdn_url: '',
    path_prefix: '',
    public_read: true,
    use_path_style: false,
  })
  const OAuth2Setting = ref<App.Api.Setting.OAuth2Setting>({
    enable: false,
    provider: OAuth2Provider.GITHUB,
    client_id: '',
    client_secret: '',
    redirect_uri: '',
    scopes: [],
    auth_url: '',
    token_url: '',
    user_info_url: '',

    is_oidc: false,
    issuer: '',
    jwks_url: '',
    auth_redirect_allowed_return_urls: [],
    cors_allowed_origins: [],
  })
  const Webhooks = ref<App.Api.Setting.Webhook[]>([])
  const webhooksLoading = ref<boolean>(false)
  const webhooksError = ref<string>('')
  const AccessTokens = ref<App.Api.Setting.AccessToken[]>([])
  const SnapshotSchedule = ref<App.Api.Setting.SnapshotSchedule>({
    enable: false,
    cron_expression: '0 2 * * 0',
  })
  const AgentSetting = ref<App.Api.Setting.AgentSetting>({
    enable: false,
    protocol: AgentProtocol.OPENAI,
    model: '',
    api_key: '',
    prompt: '',
    base_url: '',
    multimodal: false,
    context_window: 0,
  })
  const hello = ref<App.Api.Ech0.HelloEch0>()
  const loading = ref<boolean>(true)
  const snapshotStatus = ref<SnapshotUIStatus>('idle')
  const snapshotError = ref<string>('')
  const snapshotPhase = ref<string>('')
  const snapshotFileName = ref<string>('')
  const snapshotSize = ref<number>(0)
  const exportFormat = ref<ExportFormat>('snapshot')
  const exportIncludePrivate = ref<boolean>(false)
  const snapshotFormat = ref<ExportFormat>('snapshot')
  const snapshotPolling = ref<boolean>(false)
  const snapshotPollTimer = ref<ReturnType<typeof setTimeout> | null>(null)
  const snapshotPollInFlight = ref<boolean>(false)

  const getSystemSetting = async () => {
    await fetchGetSettings().then((res) => {
      if (res) {
        SystemSetting.value = res
        loading.value = false
      }
    })
  }

  const getS3Setting = async () => {
    fetchGetS3Settings().then((res) => {
      if (res) {
        S3Setting.value = res
      }
    })
  }

  const getOAuth2Setting = async () => {
    const res = await fetchGetOAuth2Settings()
    if (res) {
      OAuth2Setting.value = res
    }
    return OAuth2Setting.value
  }

  const getAllWebhooks = async () => {
    webhooksLoading.value = true
    webhooksError.value = ''
    try {
      const res = await fetchGetAllWebhooks()
      if (res) {
        Webhooks.value = listItems<App.Api.Setting.Webhook>(res)
        return
      }
      webhooksError.value = 'Failed to load webhooks'
    } catch (error) {
      webhooksError.value = error instanceof Error ? error.message : 'Failed to load webhooks'
    } finally {
      webhooksLoading.value = false
    }
  }

  const getAllAccessTokens = async () => {
    const res = await fetchListAccessTokens()
    if (res) {
      AccessTokens.value = listItems<App.Api.Setting.AccessToken>(res)
    }
  }

  const getSnapshotSchedule = async () => {
    const res = await fetchGetSnapshotScheduleSetting()
    if (res) {
      SnapshotSchedule.value = res
    }
  }

  const stopSnapshotPolling = () => {
    snapshotPolling.value = false
    if (snapshotPollTimer.value) {
      clearTimeout(snapshotPollTimer.value)
      snapshotPollTimer.value = null
    }
  }

  const applyExportState = (data: ExportStatusPayload) => {
    snapshotStatus.value = data.status
    snapshotError.value = data.error_message || ''
    snapshotPhase.value = data.phase || ''
    snapshotFileName.value = data.file_name || ''
    snapshotSize.value = data.size || 0
    snapshotFormat.value = data.format || 'snapshot'
  }

  const isExportTerminal = (status: SnapshotUIStatus) =>
    status === 'idle' || status === 'success' || status === 'failed' || status === 'cancelled'

  const scheduleSnapshotPoll = () => {
    if (!snapshotPolling.value) return
    if (snapshotPollTimer.value) clearTimeout(snapshotPollTimer.value)
    snapshotPollTimer.value = setTimeout(() => {
      void pollSnapshotStatus()
    }, SNAPSHOT_STATUS_POLL_INTERVAL_MS)
  }

  const pollSnapshotStatus = async () => {
    if (snapshotPollInFlight.value) return
    snapshotPollInFlight.value = true
    try {
      const res = await fetchGetExportStatus()
      if (res) {
        applyExportState(res)
        if (isExportTerminal(res.status)) {
          stopSnapshotPolling()
          return
        }
      }
    } catch (error) {
      snapshotError.value = error instanceof Error ? error.message : 'Failed to poll export status'
    } finally {
      snapshotPollInFlight.value = false
    }
    scheduleSnapshotPoll()
  }

  const startSnapshotTask = async (
    format: ExportFormat = 'snapshot',
    includePrivate: boolean = false,
  ) => {
    if (snapshotStatus.value === 'pending' || snapshotStatus.value === 'running') return null
    const res = await fetchStartExport({ format, include_private: includePrivate })
    if (res) {
      applyExportState(res)
      if (!res.format) snapshotFormat.value = format
      snapshotPolling.value = true
      scheduleSnapshotPoll()
    }
    return res
  }

  const restoreSnapshotTask = async () => {
    const res = await fetchGetExportStatus()
    if (res && (res.status === 'pending' || res.status === 'running')) {
      applyExportState(res)
      snapshotPolling.value = true
      scheduleSnapshotPoll()
    }
  }

  const getHelloEch0 = async () => {
    const res = await fetchHelloEch0()
    if (res) {
      hello.value = res
    }
  }

  const getAgentSetting = async () => {
    const res = await fetchGetAgentSettings()
    if (res) {
      AgentSetting.value = res
    }
  }

  const getAgentInfo = async () => {
    const res = await fetchGetAgentInfo()
    if (res) {
      AgentSetting.value.enable = res.enable
      AgentSetting.value.model = res.model
      AgentSetting.value.protocol = res.protocol
    }
  }

  const init = async () => {
    await getSystemSetting()
    if (userStore.isLogin && isAdminRole(userStore.user?.role)) {
      getS3Setting()
    }
    getAgentInfo()
    getHelloEch0()
  }

  return {
    SystemSetting,
    S3Setting,
    OAuth2Setting,
    Webhooks,
    webhooksLoading,
    webhooksError,
    AccessTokens,
    SnapshotSchedule,
    AgentSetting,
    hello,
    loading,

    getAllAccessTokens,
    getSystemSetting,
    getS3Setting,
    getOAuth2Setting,
    getAllWebhooks,
    getHelloEch0,
    getSnapshotSchedule,
    startSnapshotTask,
    pollSnapshotStatus,
    restoreSnapshotTask,
    stopSnapshotPolling,
    getAgentSetting,
    getAgentInfo,
    init,
    snapshotStatus,
    snapshotError,
    snapshotPhase,
    snapshotFileName,
    snapshotSize,
    snapshotPolling,
    snapshotFormat,
    exportFormat,
    exportIncludePrivate,
  }
})
