// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    type Response<T> = {
      code: number
      msg: string
      error_code?: string
      message_key?: string
      message_params?: Record<string, unknown>
      data: T
    }
  }
}
