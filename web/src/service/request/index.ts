// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { FetchError, ofetch } from 'ofetch'
import { ACCOUNT_DISABLED_TYPE, problemMessage } from './problem'
import { handleStaticRequest } from './static-adapter'
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
  // Diagnostics endpoints (connection tests, migration failures) surface the
  // server's detail verbatim instead of the status-derived text.
  onProblem?: (problem: App.Api.Problem) => void
  data?: object
}

// ofetch types bodies structurally; every caller passes a typed DTO.
const toBody = (data: object | undefined) => data as unknown as Record<string, unknown> | undefined

const ofetchInstance = ofetch.create({
  baseURL: import.meta.env.VITE_SERVICE_BASE_URL,
  timeout: 20000,

  onRequest({ options }) {
    const isDirectUrl = options.headers.get('X-Direct-URL')
    if (!isDirectUrl) {
      for (const [key, value] of Object.entries(buildCommonHeaders())) {
        options.headers.set(key, value)
      }
    }

    options.headers.delete('X-Direct-URL')
  },
})

// Raw third-party fetches (GitHub cards, peer instances) keep the non-throwing
// behaviour so their callers keep receiving whatever body came back.
const rawFetchInstance = ofetch.create({
  baseURL: import.meta.env.VITE_SERVICE_BASE_URL,
  timeout: 20000,
  ignoreResponseError: true,
})

// A 204 (or any empty body) resolves to `undefined`; callers only need
// "did it succeed". Hand back a truthy value so `if (res)` means success for
// every endpoint while `null` keeps meaning failure.
const settle = <T>(data: T | undefined): T => (data === undefined ? (true as unknown as T) : data)

function problemOf(error: unknown): App.Api.Problem {
  if (error instanceof FetchError) {
    const body = error.data
    if (body && typeof body === 'object' && typeof (body as App.Api.Problem).status === 'number') {
      return body as App.Api.Problem
    }
    return { status: error.response?.status ?? 0, detail: error.message }
  }
  return { status: 0 }
}

function notify(problem: App.Api.Problem, options: RequestOptions) {
  if (options.onProblem) {
    options.onProblem(problem)
    return
  }
  if (options.silentError === true || !getInitReadyStatus()) return
  theToast.error(problemMessage(problem))
}

function handleAccountDisabled(): boolean {
  if (!window.location.pathname.endsWith('/auth')) {
    useAuthStore().clearToken()
    localStorage.setItem('needLoginRedirect', 'true')
    theToast.error(String(i18n.global.t('auth.account_disabled')))
    window.location.assign(`${import.meta.env.BASE_URL}auth`)
    return true
  }
  return false
}

/** For diagnostics endpoints: the server's detail is the whole answer. */
export const reportProblemDetail = (problem: App.Api.Problem) => {
  theToast.error(problem.detail || problem.title || String(i18n.global.t('common.requestFailed')))
}

export const request = async <T>(requestOptions: RequestOptions): Promise<T | null> => {
  if (isStaticMode()) {
    return handleStaticRequest<T>(requestOptions.url, requestOptions.method, requestOptions.data)
  }

  if (import.meta.env.VITE_PROXY === 'YES') {
    const proxyUrl = import.meta.env.VITE_PROXY_URL
    if (!proxyUrl) {
      throw new Error('Proxy URL is not defined')
    }
    requestOptions.url = `${proxyUrl}${requestOptions.url}`
  }

  const doRequest = () =>
    ofetchInstance<T>(requestOptions.url, {
      method: requestOptions.method,
      body: toBody(requestOptions.data),
      timeout: requestOptions.timeout,
      credentials: 'include',
    })

  try {
    return settle(await doRequest())
  } catch (error) {
    let problem = problemOf(error)

    if (problem.status === 401) {
      const refreshed = await useAuthStore().silentRefresh()
      if (refreshed) {
        try {
          return settle(await doRequest())
        } catch (retryError) {
          problem = problemOf(retryError)
        }
      }
    }

    if (problem.type === ACCOUNT_DISABLED_TYPE) {
      handleAccountDisabled()
      return null
    }

    notify(problem, requestOptions)
    return null
  }
}

export const requestWithDirectUrl = async <T>(
  requestOptions: RequestOptions,
): Promise<T | null> => {
  try {
    return settle(
      await ofetchInstance<T>(requestOptions.dirrectUrl ? requestOptions.dirrectUrl : '', {
        method: requestOptions.method,
        body: toBody(requestOptions.data),
        timeout: requestOptions.timeout,
        credentials: 'omit',
      }),
    )
  } catch (error) {
    notify(problemOf(error), requestOptions)
    return null
  }
}

export const requestWithDirectUrlAndData = async <T>(
  requestOptions: RequestOptions,
): Promise<T> => {
  return rawFetchInstance<T>(
    requestOptions.dirrectUrlAndData ? requestOptions.dirrectUrlAndData : '',
    {
      method: requestOptions.method,
      body: toBody(requestOptions.data),
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

  try {
    return await ofetchInstance<Blob>(requestOptions.url, {
      method: requestOptions.method,
      body: toBody(requestOptions.data),
      timeout: requestOptions.timeout,
      credentials: 'include',
    })
  } catch (error) {
    notify(problemOf(error), requestOptions)
    throw new Error(problemMessage(problemOf(error)))
  }
}
