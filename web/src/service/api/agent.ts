// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetRecent() {
  return request<string>({
    url: '/agent/recent',
    method: 'GET',
  })
}
