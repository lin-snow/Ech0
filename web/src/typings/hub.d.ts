// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Hub {
      type HubItem = string | { id: string; connect_url: string }
      type HubList = HubItem[]
      type HubItemInfo = Connect.Connect
      type HubInfoList = HubItemInfo[]

      type Echo = {
        id: string
        content: string
        username: string
        echo_files?: Ech0.EchoFile[]
        tags?: Tag[]
        layout?: string
        private: boolean
        user_id: string
        extension?: Ech0.EchoExtension | null
        fav_count: number
        created_at: number | string
        createdTs: number
        virtual_key: string
        server_name: string
        server_url: string
        logo: string
      }
    }
  }
}
