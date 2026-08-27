// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request, requestWithDirectUrl } from '../request'

export function fetchGetConnectList() {
  return request<App.Api.Connect.Connected[]>({
    url: '/connect/list',
    method: 'GET',
  })
}

export function fetchGetConnect(connectUrl: string, silentError = false) {
  return requestWithDirectUrl<App.Api.Connect.Connect>({
    dirrectUrl: `${connectUrl}/api/connect`,
    url: '/',
    method: 'GET',
    silentError,
  })
}

export function fetchGetAllConnectInfo() {
  return request<App.Api.Connect.Connect[]>({
    url: '/connects/info',
    method: 'GET',
  })
}

export function fetchGetConnectsHealth() {
  return request<App.Api.Connect.ConnectedHealth[]>({
    url: '/connects/health',
    method: 'GET',
  })
}

export function fetchAddConnect(connectUrl: string) {
  return request<App.Api.Connect.Connected>({
    url: '/connects',
    method: 'POST',
    data: {
      connect_url: connectUrl,
    },
  })
}

export function fetchDeleteConnect(id: string) {
  return request<App.Api.Connect.Connected>({
    url: `/connects/${id}`,
    method: 'DELETE',
  })
}
