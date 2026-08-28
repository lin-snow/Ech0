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

export function answerChatAsk(askId: string, answers: App.Api.Chat.ChatAskAnswer[]) {
  return request({
    url: `/chat/answer`,
    method: 'POST',
    data: { ask_id: askId, answers },
  })
}

/**
 * Reads a key off a value already proven to be an object. The result is
 * `unknown`, so every caller still has to check what it got.
 */
const field = (source: object, key: string): unknown => (source as Record<string, unknown>)[key]

const str = (source: object, key: string): string | undefined => {
  const value = field(source, key)
  return typeof value === 'string' ? value : undefined
}

/**
 * An ask blocks the turn until the reader answers it, so a malformed payload
 * would strand the run behind a picker nobody can act on. Everything optional
 * is dropped rather than trusted: the fields are the model's own words and only
 * ever reach the DOM as text.
 */
function parseAsk(data: unknown): App.Api.Chat.ChatAsk | null {
  if (data === null || typeof data !== 'object') return null
  const askId = str(data, 'ask_id')
  const rawQuestions = field(data, 'questions')
  if (askId === undefined || askId.length === 0 || !Array.isArray(rawQuestions)) return null

  const questions: App.Api.Chat.ChatAskQuestion[] = []
  for (const raw of rawQuestions) {
    if (raw === null || typeof raw !== 'object') continue
    const id = str(raw, 'id')
    const text = str(raw, 'text')
    if (id === undefined || text === undefined) continue

    const rawOptions = field(raw, 'options')
    const options: App.Api.Chat.ChatAskOption[] = []
    if (Array.isArray(rawOptions)) {
      for (const opt of rawOptions) {
        if (opt === null || typeof opt !== 'object') continue
        const label = str(opt, 'label')
        if (label === undefined) continue
        options.push({ label, description: str(opt, 'description') })
      }
    }
    const recommended = field(raw, 'recommended')
    questions.push({
      id,
      text,
      header: str(raw, 'header'),
      detail: str(raw, 'detail'),
      options: options.length > 0 ? options : undefined,
      multi: field(raw, 'multi') === true,
      recommended:
        typeof recommended === 'number' && recommended >= 0 && recommended < options.length
          ? recommended
          : undefined,
    })
  }

  return questions.length > 0 ? { ask_id: askId, questions } : null
}

interface ChatStreamHandlers {
  onSearching?: (query: string) => void
  onSources?: (sources: App.Api.Chat.ChatSource[]) => void
  onCoverage?: (coverage: App.Api.Chat.ChatCoverage) => void
  onReasoning?: (text: string) => void
  onReasoningDone?: (durationMs: number) => void
  onDelta?: (text: string) => void
  onAsk?: (ask: App.Api.Chat.ChatAsk) => void
  onAskClosed?: (askId: string) => void
  onAskMalformed?: () => void
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
        case 'ask': {
          const ask = parseAsk(data)
          if (ask) handlers.onAsk?.(ask)
          // A round nobody can draw leaves the run parked behind a picker that
          // is not on screen, and it stays parked until its budget runs out.
          // Said out loud rather than dropped: silence here looks exactly like
          // the assistant having stopped for no reason.
          else handlers.onAskMalformed?.()
          break
        }
        case 'ask_closed': {
          const askId = data !== null && typeof data === 'object' ? str(data, 'ask_id') : undefined
          if (askId) handlers.onAskClosed?.(askId)
          break
        }
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
