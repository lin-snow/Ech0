// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Embedding {
      type EmbeddingSetting = {
        enable: boolean
        model: string
        api_key: string
        base_url: string
        dim: number
        batch_size: number
      }

      type EmbeddingSettingDto = EmbeddingSetting

      type ReindexResult = {
        total: number
        indexed: number
        skipped: number
        failed: number
      }

      type ReindexStatus = {
        status: 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
        phase?: string
        error?: string
        payload?: ReindexResult
        started_at?: number
        finished_at?: number
      }
    }
  }
}
