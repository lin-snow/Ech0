// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ofetch } from 'ofetch'
import { getInitReadyStatus, buildCommonHeaders, isStaticMode } from './shared'
import { useAuthStore } from '@/stores/auth'
import { theToast } from '@/utils/toast'
import { i18n } from '@/locales'

interface RequestOptions {
  dirrectUrl?: string
  dirrectUrlAndData?: string
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  timeout?: number
  silentError?: boolean
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data?: any
}

const ofetchInstance = ofetch.create({
  baseURL: import.meta.env.VITE_SERVICE_BASE_URL,
  timeout: 20000,
  ignoreResponseError: true,

  onRequest({ options }) {
    const isDirectUrl = options.headers.get('X-Direct-URL')
    if (!isDirectUrl) {
      // 公共头单一真相源（与 SSE 传输共用），见 service/request/shared.ts
      for (const [key, value] of Object.entries(buildCommonHeaders())) {
        options.headers.set(key, value)
      }
    }

    options.headers.delete('X-Direct-URL')
  },
  onResponseError: async ({ response }) => {
    let data
    try {
      data = await response.json()
    } catch {
      data = { code: 0, msg: String(i18n.global.t('common.requestFailed')), data: null }
    }

    response._data = data

    return data
  },
})

export const request = async <T>(requestOptions: RequestOptions): Promise<App.Api.Response<T>> => {
  // 静态站（ech0 build 产物）：请求由本地 dataset 应答，跳过 token 重试与错误 toast。
  // 动态 import 保证 adapter 只在静态模式下加载，非静态构建里它不进主 chunk。
  if (isStaticMode()) {
    const { handleStaticRequest } = await import('./static-adapter')
    return handleStaticRequest<T>(requestOptions.url, requestOptions.method, requestOptions.data)
  }

  const isSystemReady = getInitReadyStatus()

  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    requestOptions.url = `${proxyUrl}${requestOptions.url}`
  }

  const doRequest = () =>
    ofetchInstance<App.Api.Response<T>>(requestOptions.url, {
      method: requestOptions.method,
      body: requestOptions.data,
      timeout: requestOptions.timeout,
      credentials: 'include',
    })

  let res = await doRequest()

  const isTokenError =
    res.error_code === 'TOKEN_MISSING' ||
    res.error_code === 'TOKEN_INVALID' ||
    res.error_code === 'TOKEN_PARSE_ERROR' ||
    res.error_code === 'TOKEN_REVOKED'

  if (isTokenError) {
    const refreshed = await useAuthStore().silentRefresh()
    if (refreshed) {
      res = await doRequest()
    }
  }

  if (res.code !== 1 && !requestOptions.silentError) {
    if (isSystemReady) {
      const translated =
        res.message_key && i18n.global.te(res.message_key)
          ? i18n.global.t(res.message_key, (res.message_params || {}) as Record<string, unknown>)
          : res.msg
      theToast.error(
        translated ? String(translated) : String(i18n.global.t('common.requestFailed')),
      )
    }
  }

  return res
}

export const requestWithDirectUrl = async <T>(
  requestOptions: RequestOptions,
): Promise<App.Api.Response<T>> => {
  const isSystemReady = getInitReadyStatus()

  return ofetchInstance<App.Api.Response<T>>(
    requestOptions.dirrectUrl ? requestOptions.dirrectUrl : '',
    {
      method: requestOptions.method,
      body: requestOptions.data,
      timeout: requestOptions.timeout,
      credentials: 'omit',
    },
  ).then((res) => {
    if (res.code !== 1 && !requestOptions.silentError) {
      if (isSystemReady) {
        const translated =
          res.message_key && i18n.global.te(res.message_key)
            ? i18n.global.t(res.message_key, (res.message_params || {}) as Record<string, unknown>)
            : res.msg
        theToast.error(
          translated ? String(translated) : String(i18n.global.t('common.requestFailed')),
        )
      }
    }

    return res
  })
}

export const requestWithDirectUrlAndData = async <T>(
  requestOptions: RequestOptions,
): Promise<T> => {
  return ofetchInstance<T>(
    requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
    {
      method: requestOptions.method,
      body: requestOptions.data,
      timeout: requestOptions.timeout,
      credentials: 'omit',
      headers: {
        'X-Direct-URL': requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
      },
    },
  ).then((res) => {
    return res
  })
}

export const downloadFile = async (requestOptions: RequestOptions): Promise<Blob> => {
  if (isStaticMode()) {
    throw new Error('File download is not available in static mode')
  }

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
    timeout: requestOptions.timeout,
    credentials: 'include',
  }).then((res) => {
    if (res instanceof Blob) {
      return res
    }
    const msg = String(i18n.global.t('common.requestFailed'))
    theToast.error(msg)
    throw new Error(msg)
  })
}
