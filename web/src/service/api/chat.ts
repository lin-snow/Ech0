// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '@/service/request'
import { sseStream } from '@/service/request/sse'

export function getChatSession() {
  return request<App.Api.Chat.ChatMessage[]>({
    url: `/chat/session`,
    method: 'GET',
  })
}

export function clearChatSession() {
  return request({
    url: `/chat/session`,
    method: 'DELETE',
  })
}

interface ChatStreamHandlers {
  onSearching?: (query: string) => void
  onSources?: (sources: App.Api.Chat.ChatSource[]) => void
  onCoverage?: (coverage: App.Api.Chat.ChatCoverage) => void
  onReasoning?: (text: string) => void
  onReasoningDone?: (durationMs: number) => void
  onDelta?: (text: string) => void
  onError?: (message: string) => void
  onDone?: () => void
}

export function chatStream(question: string, handlers: ChatStreamHandlers): () => void {
  let done = false
  const finish = () => {
    if (done) return
    done = true
    handlers.onDone?.()
  }

  return sseStream({
    path: '/chat',
    body: { question },
    onEvent: (event, data) => {
      switch (event) {
        case 'searching':
          handlers.onSearching?.((data as { name: string; query: string }).query)
          break
        case 'sources':
          handlers.onSources?.(data as App.Api.Chat.ChatSource[])
          break
        case 'coverage':
          handlers.onCoverage?.(data as App.Api.Chat.ChatCoverage)
          break
        case 'reasoning':
          handlers.onReasoning?.((data as { text: string }).text)
          break
        case 'reasoning_done':
          handlers.onReasoningDone?.((data as { duration_ms: number }).duration_ms)
          break
        case 'delta':
          handlers.onDelta?.((data as { text: string }).text)
          break
        case 'error':
          handlers.onError?.((data as { message: string }).message)
          break
        case 'done':
          finish()
          break
      }
    },
    onError: (message) => handlers.onError?.(message),
    onClose: finish,
  })
}
