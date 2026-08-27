// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { getApiUrl, buildCommonHeaders } from './shared'

export interface SSEStreamOptions {
  path: string
  body?: unknown
  method?: 'GET' | 'POST'
  onEvent: (name: string, data: unknown) => void
  onError?: (message: string) => void
  onClose?: () => void
}

export function sseStream(opts: SSEStreamOptions): () => void {
  const controller = new AbortController()

  const dispatch = (rawEvent: string) => {
    let eventName = 'message'
    const dataLines: string[] = []
    for (const line of rawEvent.split('\n')) {
      if (line.startsWith('event:')) {
        eventName = line.slice(6).trim()
      } else if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trim())
      }
    }
    if (dataLines.length === 0) return
    let payload: unknown
    try {
      payload = JSON.parse(dataLines.join('\n'))
    } catch {
      return
    }
    opts.onEvent(eventName, payload)
  }

  const run = async () => {
    let resp: Response
    try {
      resp = await fetch(`${getApiUrl()}${opts.path}`, {
        method: opts.method ?? 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...buildCommonHeaders(),
        },
        body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
        credentials: 'include',
        signal: controller.signal,
      })
    } catch (e) {
      if (!controller.signal.aborted) {
        opts.onError?.(String(e))
      }
      return
    }

    if (!resp.ok || !resp.body) {
      opts.onError?.(`HTTP ${resp.status}`)
      return
    }

    const reader = resp.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    try {
      for (;;) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        let sepIndex: number
        while ((sepIndex = buffer.indexOf('\n\n')) !== -1) {
          const rawEvent = buffer.slice(0, sepIndex)
          buffer = buffer.slice(sepIndex + 2)
          if (rawEvent.trim().length > 0 && !rawEvent.startsWith(':')) {
            dispatch(rawEvent)
          }
        }
      }
      opts.onClose?.()
    } catch (e) {
      if (!controller.signal.aborted) {
        opts.onError?.(String(e))
      }
    }
  }

  run()

  return () => controller.abort()
}
