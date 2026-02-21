// 封装ofetch

import { ofetch } from 'ofetch'
import { getAuthToken, getSystemReadyStatus } from './shared'
import { theToast } from '@/utils/toast'

interface RequestOptions {
  dirrectUrl?: string
  dirrectUrlAndData?: string
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data?: any
}

const ofetchInstance = ofetch.create({
  baseURL: import.meta.env.VITE_SERVICE_BASE_URL,
  timeout: 20000,
  ignoreResponseError: true, // 忽略响应错误，让响应拦截器处理

  // 请求拦截器
  onRequest({ options }) {
    const token = getAuthToken()
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'

    const isDirectUrl = options.headers.get('X-Direct-URL')
    if (token && token.length > 0 && !isDirectUrl) {
      options.headers.append('Authorization', token)
    }
    if (!isDirectUrl) {
      options.headers.set('X-Timezone', timezone)
    }

    // 清空请求头
    options.headers.delete('X-Direct-URL')
  },
  // 响应拦截器
  onResponseError: async ({ response }) => {
    let data
    try {
      data = await response.json()
    } catch {
      data = { code: 0, msg: `请求失败`, data: null }
    }

    response._data = data

    // 不再 throw，让后续 then() 也能拿到
    return data
  },
})

export const request = async <T>(requestOptions: RequestOptions): Promise<App.Api.Response<T>> => {
  // 检查系统是否已经准备好
  const isSystemReady = getSystemReadyStatus()

  // 检查是否使用正向代理
  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    requestOptions.url = `${proxyUrl}${requestOptions.url}`
  }

  return ofetchInstance<App.Api.Response<T>>(requestOptions.url, {
    method: requestOptions.method,
    body: requestOptions.data,
  }).then((res) => {
    if (res.code !== 1) {
      if (isSystemReady) {
        theToast.error(res.msg ? String(res.msg) : '请求失败')
      }
    }

    return res
  })
}

// 直接请求
export const requestWithDirectUrl = async <T>(
  requestOptions: RequestOptions,
): Promise<App.Api.Response<T>> => {
  // 检查系统是否已经准备好
  const isSystemReady = getSystemReadyStatus()

  return ofetchInstance<App.Api.Response<T>>(
    requestOptions.dirrectUrl ? requestOptions.dirrectUrl : '',
    {
      method: requestOptions.method,
      body: requestOptions.data,
    },
  ).then((res) => {
    if (res.code !== 1) {
      if (isSystemReady) {
        theToast.error(res.msg ? String(res.msg) : '请求失败')
      }
    }

    return res
  })
}

// 直接请求 && 直接传递数据
export const requestWithDirectUrlAndData = async <T>(
  requestOptions: RequestOptions,
): Promise<T> => {
  return ofetchInstance<T>(
    requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
    {
      method: requestOptions.method,
      body: requestOptions.data,
      headers: {
        'X-Direct-URL': requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
      },
    },
  ).then((res) => {
    return res
  })
}

export const downloadFile = async (requestOptions: RequestOptions): Promise<Blob> => {
  // 检查是否使用正向代理
  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    requestOptions.url = `${proxyUrl}${requestOptions.url}`
  }

  return ofetchInstance<Blob>(requestOptions.url, {
    method: requestOptions.method,
    body: requestOptions.data,
  }).then((res) => {
    if (res instanceof Blob) {
      return res
    } else {
      theToast.error('下载失败')
      throw new Error('下载失败')
    }
  })
}
