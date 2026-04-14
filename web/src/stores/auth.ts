import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ofetch } from 'ofetch'
import { getApiUrl } from '@/service/request/shared'

export const useAuthStore = defineStore('authStore', () => {
  const accessToken = ref('')
  const authHeader = computed(() => (accessToken.value ? `Bearer ${accessToken.value}` : ''))

  let refreshPromise: Promise<boolean> | null = null

  function setToken(token: string) {
    accessToken.value = token || ''
  }

  function clearToken() {
    accessToken.value = ''
  }

  // 静默刷新：仅通过 HttpOnly Cookie 刷新 access token，前端不接触 refresh token。
  async function silentRefresh(): Promise<boolean> {
    if (refreshPromise) return refreshPromise
    refreshPromise = (async () => {
      try {
        const res = await ofetch<App.Api.Response<App.Api.Auth.TokenPairResponse>>(
          `${getApiUrl()}/auth/refresh`,
          { method: 'POST', credentials: 'include', ignoreResponseError: true },
        )
        if (res.code === 1 && res.data?.access_token) {
          setToken(res.data.access_token)
          return true
        }
        clearToken()
        return false
      } catch {
        clearToken()
        return false
      } finally {
        refreshPromise = null
      }
    })()
    return refreshPromise
  }

  // 登出：后端清 Cookie + 记录黑名单，前端清 access token 内存状态。
  async function logout() {
    try {
      await ofetch(`${getApiUrl()}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
        ignoreResponseError: true,
        headers: authHeader.value ? { Authorization: authHeader.value } : {},
      })
    } finally {
      clearToken()
    }
  }

  return {
    accessToken,
    authHeader,
    setToken,
    clearToken,
    silentRefresh,
    logout,
  }
})
