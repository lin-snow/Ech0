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

// 泛型 T 表示服务端返回的数据结构
type Callback<T> = (payload: T) => void

export function useOWebSocket<T = unknown>(options: WSOptions) {
  const { url, autoReconnect = true, heartbeat = true, protocols } = options

  const token = useAuthStore().accessToken

  // WebSocket URL 支持携带 token，常用方式是 query
  const wsUrl = token ? `${url}?token=${token}` : url

  const { status, data, send, open, close, ws } = useWebSocket(wsUrl, {
    autoReconnect,
    heartbeat,
    protocols,
    immediate: false, // 可手动 open
  })

  // 消息回调表
  const listeners = reactive<Record<string, Callback<T>[]>>({})

  // 监听消息
  watch(data, (msg) => {
    if (!msg) return

    // console.log('收到 WebSocket 消息:', msg)
    try {
      const parsed = JSON.parse(msg as string) as T
      // 默认回调 key = 'default'
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
