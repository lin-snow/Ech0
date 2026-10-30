// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Device {
      type PendingInfo = {
        client_name: string
        scopes: string[]
        expires_in: number
      }
    }
  }
}
