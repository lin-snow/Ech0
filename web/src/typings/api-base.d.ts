// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    // RFC 9457 problem document. `status` is the only machine-readable signal a
    // client may branch on (plus `type` for the rare case two errors share a
    // status); `detail` is developer-facing English and must never be parsed.
    type Problem = {
      type?: string
      title?: string
      status: number
      detail?: string
      errors?: {
        location?: string
        message: string
        value?: unknown
      }[]
    }
  }
}
