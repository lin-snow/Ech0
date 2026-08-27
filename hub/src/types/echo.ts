// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export interface EchoTag {
  id: string
  name: string
  created_at?: number | string
  usage_count?: number
}

export interface EchoPost {
  id: string
  content: string
  username?: string
  created_at: number | string
  fav_count?: number
  tags?: EchoTag[]
  echo_files?: App.Api.Ech0.EchoFile[]
  layout?: string
  extension?: App.Api.Ech0.EchoExtension | null
  private?: boolean
  user_id?: string
}

export interface HubPostMeta {
  instanceId: string
  instanceUrl: string
}

export type EchoPostWithHub = EchoPost & { _hub: HubPostMeta }

export interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

export interface EchoQueryPage {
  total: number
  items: EchoPost[]
}
