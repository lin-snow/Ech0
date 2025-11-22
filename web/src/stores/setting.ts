import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  fetchGetSettings,
  fetchGetStatus,
  fetchGetCommentSettings,
  fetchGetS3Settings,
  fetchGetOAuth2Settings,
  fetchGetAllWebhooks,
  fetchListAccessTokens,
  fetchGetFediverseSettings,
  fetchGetBackupScheduleSetting,
  fetchHelloEch0,
} from '@/service/api'
import { localStg } from '@/utils/storage'
import { theToast } from '@/utils/toast'
import router from '@/router'
import { CommentProvider, S3Provider, OAuth2Provider } from '@/enums/enums'

export const useSettingStore = defineStore('settingStore', () => {
  /**
   * State
   */
  const isSystemReady = ref<boolean>(false)
  const SystemSetting = ref<App.Api.Setting.SystemSetting>({
    site_title: import.meta.env.VITE_APP_TITLE,
    server_name: import.meta.env.VITE_APP_NAME,
    server_url: '',
    allow_register: true,
    ICP_number: '',
    meting_api: '',
    custom_css: '',
    custom_js: '',
    logo: '',
  })
  const CommentSetting = ref<App.Api.Setting.CommentSetting>({
    enable_comment: false,
    provider: CommentProvider.TWIKOO,
    comment_api: '',
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
  })
  const Webhooks = ref<App.Api.Setting.Webhook[]>([])
  const AccessTokens = ref<App.Api.Setting.AccessToken[]>([])
  const FediverseSetting = ref<App.Api.Setting.FediverseSetting>({
    enable: false,
    server_url: '',
  })
  const BackupSchedule = ref<App.Api.Setting.BackupSchedule>({
    enable: false,
    cron_expression: '0 2 * * 0',
  })
  const hello = ref<App.Api.Ech0.HelloEch0>()
  const loading = ref<boolean>(true)

  /**
   * Actions
   */
  const getSystemReady = async () => {
    // 检查localStorage中是否有系统状态
    const systemStatus = localStg.getItem<boolean>('systemStatus')
    if (systemStatus !== null) {
      // 如果有，直接使用localStorage中的值
      isSystemReady.value = systemStatus
    } else {
      // 如果没有，默认设置为false
      isSystemReady.value = false
    }

    // 检查系统是否准备好
    if (!isSystemReady.value) {
      // 如果系统未准备好，调用接口获取系统状态
      const res = await fetchGetStatus()
      if (res.code === 666) {
        isSystemReady.value = false
        theToast.info(res.msg)
        // 跳转到注册页面
        router.push({ name: 'auth' })
      } else {
        isSystemReady.value = true
        console.log('系统已准备好')
      }

      // 保存系统状态到localStorage
      localStg.setItem('systemStatus', isSystemReady.value)
    }
  }

  const getSystemSetting = async () => {
    await fetchGetSettings().then((res) => {
      if (res.code === 1) {
        SystemSetting.value = res.data
        loading.value = false
      }
    })
  }

  const getCommentSetting = async () => {
    fetchGetCommentSettings().then((res) => {
      if (res.code === 1) {
        CommentSetting.value = res.data
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
    const res = await fetchGetAllWebhooks()
    if (res.code === 1) {
      if (res.data) {
        Webhooks.value = res.data
      } else {
        Webhooks.value = []
      }
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

  const getFediverseSetting = async () => {
    const res = await fetchGetFediverseSettings()
    if (res.code === 1) {
      FediverseSetting.value = res.data
    }
  }

  const getBackupSchedule = async () => {
    const res = await fetchGetBackupScheduleSetting()
    if (res.code === 1) {
      BackupSchedule.value = res.data
    }
  }

  const getHelloEch0 = async () => {
    const res = await fetchHelloEch0()
    if (res.code === 1) {
      hello.value = res.data
    }
  }

  const setSystemReady = (status: boolean) => {
    isSystemReady.value = status
  }

  const init = () => {
    if (!isSystemReady.value) {
      getSystemReady()
    }
    getSystemSetting()
    getCommentSetting()
    getS3Setting()
    getHelloEch0()
  }

  return {
    isSystemReady,
    SystemSetting,
    CommentSetting,
    S3Setting,
    OAuth2Setting,
    Webhooks,
    AccessTokens,
    FediverseSetting,
    BackupSchedule,
    hello,
    loading,

    getAllAccessTokens,
    getSystemReady,
    getSystemSetting,
    getCommentSetting,
    getS3Setting,
    getOAuth2Setting,
    getAllWebhooks,
    getFediverseSetting,
    setSystemReady,
    getHelloEch0,
    getBackupSchedule,
    init,
  }
})
