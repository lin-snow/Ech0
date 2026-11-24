// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetCurrentUser() {
  return request<App.Api.User.User>({
    url: '/user',
    method: 'GET',
  })
}

export function fetchUpdateUser(user: App.Api.User.UserInfo) {
  return request({
    url: '/user',
    method: 'PUT',
    data: user,
  })
}

export function fetchGetAllUsers() {
  return request<App.Api.User.User[]>({
    url: '/users',
    method: 'GET',
  })
}

export function fetchUpdateUserRole(id: string, role: 'admin' | 'member') {
  return request({
    url: `/user/role/${id}`,
    method: 'PUT',
    data: { role },
  })
}

export function fetchTransferOwnership(id: string) {
  return request({
    url: `/user/owner/${id}`,
    method: 'PUT',
  })
}

export function fetchUpdateUserStatus(id: string, status: 'ACTIVE' | 'DISABLED') {
  return request({
    url: `/user/status/${id}`,
    method: 'PUT',
    data: { status },
  })
}

export function fetchDeleteUser(id: string) {
  return request({
    url: `/user/${id}`,
    method: 'DELETE',
  })
}

export function fetchBindOAuth2(provider: string, redirect_uri: string) {
  return request<string>({
    url: `/oauth/${provider}/bind`,
    method: 'POST',
    data: {
      redirect_uri: redirect_uri,
    },
  })
}
