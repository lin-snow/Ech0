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
      }

      type StreamEvent =
        | { type: 'searching'; data: { name: string; query: string } }
        | { type: 'sources'; data: ChatSource[] }
        | { type: 'coverage'; data: ChatCoverage }
        | { type: 'reasoning'; data: { text: string } }
        | { type: 'reasoning_done'; data: { duration_ms: number } }
        | { type: 'delta'; data: { text: string } }
        | { type: 'error'; data: { message: string } }
        | { type: 'done'; data: { done: boolean } }
    }
  }
}
