// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Chat {
      type ChatSource = {
        echo_id: string
        content: string
        username: string
        echo_created: number
        distance: number
        files?: ChatSourceFile[]
        extension?: Ech0.EchoExtension
      }

      type ChatSourceFile = {
        id: string
        key?: string
        storage_type: File.StorageType
        url: string
        content_type?: string
        category?: File.Category
        size?: number
        width?: number
        height?: number
      }

      type ChatCoverage = {
        total: number
        returned: number
        buckets: number
        truncated: boolean
      }

      type ChatAskOption = {
        label: string
        description?: string
      }

      type ChatAskQuestion = {
        id: string
        text: string
        header?: string
        detail?: string
        options?: ChatAskOption[]
        multi?: boolean
        /** Index into `options`. A hint the model leaves; never a default. */
        recommended?: number
      }

      type ChatAsk = {
        ask_id: string
        questions: ChatAskQuestion[]
      }

      type ChatAskAnswer = {
        question_id: string
        selected?: string[]
        custom?: string
      }

      type ChatAskExchange = {
        questions: ChatAskQuestion[]
        answers: ChatAskAnswer[]
      }

      type ChatMessage = {
        role: 'user' | 'assistant'
        content: string
        sources?: ChatSource[]
        searches?: string[]
        coverage?: ChatCoverage
        failed?: boolean
        reasoning?: string
        reasoning_ms?: number
        reasoningActive?: boolean
        asks?: ChatAskExchange[]
        /** Transient, client-only: the round currently blocking the run. */
        pendingAsk?: ChatAsk
      }

      type StreamEvent =
        | { type: 'searching'; data: { name: string; query: string } }
        | { type: 'sources'; data: ChatSource[] }
        | { type: 'coverage'; data: ChatCoverage }
        | { type: 'reasoning'; data: { text: string } }
        | { type: 'reasoning_done'; data: { duration_ms: number } }
        | { type: 'delta'; data: { text: string } }
        | { type: 'error'; data: { message: string } }
        | { type: 'ask'; data: ChatAsk }
        | { type: 'ask_closed'; data: { ask_id: string } }
        | { type: 'done'; data: { done: boolean } }
    }
  }
}
