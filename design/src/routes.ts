// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type PageId = 'foundations' | 'components' | 'patterns' | 'ux'
export type RoutePage = PageId | 'home'

export type Route = {
  page: RoutePage
  anchor?: string
}
