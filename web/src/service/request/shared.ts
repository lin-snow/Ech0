// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { localStg } from '@/utils/storage'
import { useAuthStore } from '@/stores/auth'
import { i18n } from '@/locales'

export function buildCommonHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    'X-Timezone': Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
    'X-Locale': i18n.global.locale.value,
  }
  const authHeader = useAuthStore().authHeader
  if (authHeader) {
    headers['Authorization'] = authHeader
  }
  return headers
}

export function isStaticMode(): boolean {
  return typeof window !== 'undefined' && window.__ECH0_STATIC__ === true
}

export function staticBase(): string {
  if (typeof window === 'undefined') return '/'
  return window.__ECH0_STATIC_BASE__ || '/'
}

export const getApiUrl = () => {
  const baseUrl = import.meta.env.VITE_SERVICE_BASE_URL
  const resolvedBaseUrl = baseUrl.replace(/\/+$/, '')

  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    return `${resolvedBaseUrl}${proxyUrl}`
  }
  return resolvedBaseUrl
}

const getServiceBaseUrl = () => {
  const baseUrl = import.meta.env.VITE_SERVICE_BASE_URL
  return baseUrl.replace(/\/+$/, '')
}

export const resolveAvatarUrl = (rawUrl?: string, fallback = '/Ech0.svg') => {
  const value = (rawUrl || '').trim()
  if (!value || value === 'Ech0.svg' || value === '/Ech0.svg') {
    return fallback
  }

  if (/^https?:\/\//i.test(value)) {
    return value
  }

  if (value.startsWith('/api/')) {
    return `${getServiceBaseUrl()}${value}`
  }

  const apiUrl = getApiUrl().replace(/\/+$/, '')
  if (value.startsWith('/')) {
    return `${apiUrl}${value}`
  }
  return `${apiUrl}/${value}`
}

export const getInitReadyStatus = () => {
  const initStatus = localStg.getItem<boolean>('initialized')
  if (initStatus !== null) {
    return initStatus
  }
  return false
}

export function getWsUrl(path: string) {
  const baseUrl = import.meta.env.VITE_SERVICE_BASE_URL

  const wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:'

  if (baseUrl === '/' || baseUrl.startsWith('/')) {
    return `${wsProtocol}//${location.host}${path}`
  }

  const httpUrl = new URL(baseUrl)
  return `${wsProtocol}//${httpUrl.host}${path}`
}
