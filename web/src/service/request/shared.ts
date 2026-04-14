import { localStg } from '@/utils/storage'

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
