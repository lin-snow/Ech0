import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchGetSettings, fetchGetStatus } from '@/service/api'
import { localStg } from '@/utils/storage'
import { theToast } from '@/utils/toast'
import { useRouter } from 'vue-router'

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
  })
  const loading = ref<boolean>(true)
  const router = useRouter()

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
      await fetchGetStatus()
        .then((res) => {
          if (res.code === 666) {
            isSystemReady.value = false
            theToast.info(res.msg)
            // 跳转到登录页面
            router.push({ name: 'auth' })
          } else {
            isSystemReady.value = true
            console.log('系统已准备好')
          }
        })
        .finally(() => {
          // 保存系统状态到localStorage
          localStg.setItem('systemStatus', isSystemReady.value)
        })
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

  if (!isSystemReady.value) {
    getSystemSetting()
  }

  const setSystemReady = (status: boolean) => {
    isSystemReady.value = status
  }

  return { isSystemReady, SystemSetting, loading, getSystemReady, getSystemSetting, setSystemReady }
})
