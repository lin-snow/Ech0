// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
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
    home_layout: 'single',
    allow_register: true,
    default_locale: 'zh-CN',
    ICP_number: '',
    footer_content: '',
    footer_link: '',
    meting_api: '',
    custom_css: '',
    custom_js: '',
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
      if (res.code === 1) {
        SystemSetting.value = res.data
        loading.value = false
      }
    })
  }

  const getS3Setting = async () => {
    fetchGetS3Settings().then((res) => {
      if (res.code === 1) {
        S3Setting.value = res.data
      }
    })
  }

  const getOAuth2Setting = async () => {
    const res = await fetchGetOAuth2Settings()
    if (res.code === 1) {
      OAuth2Setting.value = res.data
    }
    return OAuth2Setting.value
  }

  const getAllWebhooks = async () => {
    webhooksLoading.value = true
    webhooksError.value = ''
    try {
      const res = await fetchGetAllWebhooks()
      if (res.code === 1) {
        Webhooks.value = res.data || []
        return
      }
      webhooksError.value = res.msg || 'Failed to load webhooks'
    } catch (error) {
      webhooksError.value = error instanceof Error ? error.message : 'Failed to load webhooks'
    } finally {
      webhooksLoading.value = false
    }
  }

  const getAllAccessTokens = async () => {
    const res = await fetchListAccessTokens()
    if (res.code === 1) {
      if (res.data) {
        AccessTokens.value = res.data
      } else {
        AccessTokens.value = []
      }
    }
  }

  const getSnapshotSchedule = async () => {
    const res = await fetchGetSnapshotScheduleSetting()
    if (res.code === 1) {
      SnapshotSchedule.value = res.data
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
      if (res.code === 1) {
        applyExportState(res.data)
        if (isExportTerminal(res.data.status)) {
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
    if (res.code === 1 && res.data) {
      applyExportState(res.data)
      if (!res.data.format) snapshotFormat.value = format
      snapshotPolling.value = true
      scheduleSnapshotPoll()
    }
    return res
  }

  const restoreSnapshotTask = async () => {
    const res = await fetchGetExportStatus()
    if (res.code === 1 && (res.data.status === 'pending' || res.data.status === 'running')) {
      applyExportState(res.data)
      snapshotPolling.value = true
      scheduleSnapshotPoll()
    }
  }

  const getHelloEch0 = async () => {
    const res = await fetchHelloEch0()
    if (res.code === 1) {
      hello.value = res.data
    }
  }

  const getAgentSetting = async () => {
    const res = await fetchGetAgentSettings()
    if (res.code === 1) {
      AgentSetting.value = res.data
    }
  }

  const getAgentInfo = async () => {
    const res = await fetchGetAgentInfo()
    if (res.code === 1) {
      AgentSetting.value.enable = res.data.enable
      AgentSetting.value.model = res.data.model
      AgentSetting.value.protocol = res.data.protocol
    }
  }

  const init = async () => {
    await getSystemSetting()
    if (userStore.isLogin && userStore.user?.is_admin) {
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
