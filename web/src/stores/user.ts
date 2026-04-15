import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { fetchLogin, fetchSignup, fetchGetCurrentUser, fetchExchangeCode } from '@/service/api'
import { localStg } from '@/utils/storage'
import { theToast } from '@/utils/toast'
import router from '@/router'
import { useAuthStore } from './auth'
import { useEchoStore } from './echo'
import { i18n, setI18nLocale } from '@/locales'

export const useUserStore = defineStore('userStore', () => {
  const authStore = useAuthStore()
  const user = ref<App.Api.User.User | null>(null)
  const isLogin = computed(() => !!user.value)
  const initialized = ref<boolean>(false)

  async function login(userInfo: App.Api.Auth.LoginParams) {
    const res = await fetchLogin(userInfo)
    if (res.code === 1 && res.data?.access_token) {
      authStore.setToken(res.data.access_token)

      await refreshCurrentUser()

      theToast.success(String(i18n.global.t('auth.loginSuccess')))

      const echoStore = useEchoStore()
      echoStore.clearEchos()

      router.push({ name: 'home' })
    }
  }

  async function loginWithTokenPair(data: App.Api.Auth.TokenPairResponse) {
    if (data?.access_token) {
      authStore.setToken(data.access_token)

      await refreshCurrentUser()

      theToast.success(String(i18n.global.t('auth.loginSuccess')))

      const echoStore = useEchoStore()
      echoStore.clearEchos()

      router.push({ name: 'home' })
    }
  }

  async function loginWithCode(code: string) {
    const res = await fetchExchangeCode(code)
    if (res.code === 1 && res.data?.access_token) {
      await loginWithTokenPair(res.data)
    }
  }

  async function signup(userInfo: App.Api.Auth.SignupParams) {
    return await fetchSignup(userInfo).then((res) => {
      if (res.code === 1) {
        theToast.success(String(i18n.global.t('auth.signupSuccess')))
        return true
      }
      return false
    })
  }

  async function logout() {
    await authStore.logout()

    user.value = null

    const echoStore = useEchoStore()
    echoStore.clearEchos()

    localStg.setItem('needLoginRedirect', true)
  }

  async function autoLogin() {
    // OAuth 回调带有 code 参数，将由 loginWithCode 独立完成登录，
    // 跳过 silentRefresh 避免用已清除的 cookie 触发无意义的 401。
    const url = new URL(window.location.href)
    if (url.pathname.endsWith('/auth') && url.searchParams.has('code')) {
      return
    }
    const ok = await authStore.silentRefresh()
    if (ok) {
      await refreshCurrentUser()
    }
  }

  async function refreshCurrentUser() {
    const res = await fetchGetCurrentUser()
    if (res.code === 1) {
      user.value = res.data
      if (res.data.locale) {
        await setI18nLocale(res.data.locale)
      }
    } else {
      user.value = null
      authStore.clearToken()
    }
  }

  const init = async () => {
    await autoLogin()
    initialized.value = true
  }

  return {
    initialized,
    user,
    isLogin,
    login,
    loginWithTokenPair,
    loginWithCode,
    signup,
    logout,
    autoLogin,
    refreshCurrentUser,
    init,
  }
})
