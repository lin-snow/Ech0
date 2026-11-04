// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetInitStatus() {
  return request<App.Api.Init.Status>({
    url: '/init/status',
    method: 'GET',
  })
}

export function fetchInitOwner(payload: App.Api.Init.InitOwnerParams) {
  return request({
    url: '/init/owner',
    method: 'POST',
    data: payload,
  })
}
