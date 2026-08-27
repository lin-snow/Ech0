// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace User {
      type User = {
        id: string
        username: string
        email?: string
        password?: string
        is_admin: boolean
        is_owner?: boolean
        avatar?: string
        locale: string
      }

      type UserInfo = {
        username: string
        password: string
        email?: string
        is_admin: boolean
        is_owner?: boolean
        avatar: string
        avatar_file_id?: string
        locale: string
      }
    }
  }
}
