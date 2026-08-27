// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useWebSocket } from '@vueuse/core'
import { reactive, watch } from 'vue'
import { i18n } from '@/locales'
import { useAuthStore } from '@/stores/auth'

type ReconnectOptions = {
  retries?: number
  delay?: number
  onFailed?: () => void
}

type HeartbeatOptions = {
  message?: string
  responseMessage?: string
  interval?: number
  pongTimeout?: number
}

interface WSOptions {
  url: string
  autoReconnect?: boolean | ReconnectOptions
  heartbeat?: boolean | HeartbeatOptions
  protocols?: string[]
}

type Callback<T> = (payload: T) => void

export function useOWebSocket<T = unknown>(options: WSOptions) {
  const { url, autoReconnect = true, heartbeat = true, protocols } = options

  const token = useAuthStore().accessToken

  const wsUrl = token ? `${url}?token=${token}` : url

  const { status, data, send, open, close, ws } = useWebSocket(wsUrl, {
    autoReconnect,
    heartbeat,
    protocols,
    immediate: false,
  })

  const listeners = reactive<Record<string, Callback<T>[]>>({})

  watch(data, (msg) => {
    if (!msg) return

    try {
      const parsed = JSON.parse(msg as string) as T
      listeners['default']?.forEach((cb) => cb(parsed))
    } catch {
      console.warn(i18n.global.t('websocket.invalidMessage'), msg)
    }
  })

  const sendMessage = (payload: unknown) => {
    send(JSON.stringify(payload))
  }

  const onMessage = (cb: Callback<T>, type: string = 'default') => {
    if (!listeners[type]) listeners[type] = []
    listeners[type].push(cb)
  }

  return { status, sendMessage, onMessage, open, close, ws }
}
