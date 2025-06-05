import { localStg } from '@/utils/storage'

export const getAuthToken = () => {
  const token = localStg.getItem<string>('token')
  return token ? `Bearer ${token}` : ''
}

export const saveAuthToken = (token: string) => {
  if (token) {
    localStg.setItem('token', token)
  }
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

export const getSystemReadyStatus = () => {
  // 检查localStorage中是否有系统状态
  const systemStatus = localStg.getItem<boolean>('systemStatus')
  if (systemStatus !== null) {
    // 如果有，直接使用localStorage中的值
    return systemStatus
  } else {
    // 如果没有，默认设置为false
    return false
  }
}
