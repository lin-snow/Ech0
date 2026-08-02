// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { localStg } from '@/utils/storage'
import { useAuthStore } from '@/stores/auth'
import { i18n } from '@/locales'

/**
 * 全站请求公共头的唯一真相源：Authorization + X-Locale + X-Timezone。
 * ofetch 的 onRequest 拦截器与 SSE 传输（service/request/sse.ts）都调它，
 * 杜绝公共头逻辑在多处手写漂移。
 */
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

/**
 * 静态站（`ech0 build` 产物）开关：index.html 里注入 `window.__ECH0_STATIC__ = true`。
 * 判定必须极轻且不 import adapter，保证非静态构建下 adapter 不进主 chunk。
 */
export function isStaticMode(): boolean {
  return typeof window !== 'undefined' && window.__ECH0_STATIC__ === true
}

export const getApiUrl = () => {
  const baseUrl = import.meta.env.VITE_SERVICE_BASE_URL
  const resolvedBaseUrl = baseUrl.replace(/\/+$/, '') // 正则去除末尾的斜杠

  // 检查是否使用正向代理
  if (import.meta.env.VITE_PROXY === 'YES') {
    // BaseURL + ProxyURL
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

// src/utils/ws.ts
export function getWsUrl(path: string) {
  // 取出基础地址
  const baseUrl = import.meta.env.VITE_SERVICE_BASE_URL

  // 根据当前协议选择 ws 或 wss
  const wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:'

  // 如果是相对路径（生产环境配置为 "/"），自动拼上当前域名
  if (baseUrl === '/' || baseUrl.startsWith('/')) {
    return `${wsProtocol}//${location.host}${path}`
  }

  // 否则使用开发环境配置的完整 baseUrl
  const httpUrl = new URL(baseUrl)
  return `${wsProtocol}//${httpUrl.host}${path}`
}
