// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ofetch } from 'ofetch'
import { getApiUrl, isStaticMode } from '@/service/request/shared'

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

  async function silentRefresh(): Promise<boolean> {
    if (refreshPromise) return refreshPromise
    if (isStaticMode()) {
      clearToken()
      return false
    }
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
