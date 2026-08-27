// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Init {
      type Status = {
        initialized: boolean
        owner_exists: boolean
      }
    }

    namespace Connect {
      type Connect = {
        server_name: string
        server_url: string
        logo: string
        total_echos: number
        today_echos: number
        sys_username: string
        version: string
      }

      type Connected = {
        id: string
        connect_url: string
      }

      type ConnectedHealth = {
        id: string
        connect_url: string
        status: 'online' | 'offline'
        version: string
      }
    }
  }
}
