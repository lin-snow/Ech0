// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace SystemLog {
      type Entry = {
        time: string
        level: string
        msg: string
        module?: string
        caller?: string
        error?: string
        raw?: string
        fields?: Record<string, unknown>
      }

      type QueryParams = {
        tail?: number
        level?: string
        keyword?: string
      }
    }

    namespace Dashboard {
      type VisitorDayStat = {
        date: string
        pv: number
        uv: number
      }
    }
  }
}
