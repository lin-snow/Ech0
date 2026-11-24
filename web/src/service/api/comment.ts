// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { reportProblemDetail, request } from '../request'

export function fetchGetCommentFormMeta() {
  return request<App.Api.Comment.FormMeta>({
    url: '/comments/form',
    method: 'GET',
  })
}

export function fetchGetComments(echoId: string, cacheBust = true) {
  const cacheSuffix = cacheBust ? `&_t=${Date.now()}` : ''
  return request<App.Api.Comment.CommentItem[]>({
    url: `/comments?echo_id=${encodeURIComponent(echoId)}${cacheSuffix}`,
    method: 'GET',
  })
}

export function fetchGetPublicComments(limit = 30) {
  const safeLimit = Math.max(1, Math.min(limit, 100))
  return request<App.Api.Comment.CommentItem[]>({
    url: `/comments/public?limit=${safeLimit}`,
    method: 'GET',
  })
}

export function fetchCreateComment(payload: App.Api.Comment.CreateCommentDto) {
  return request<App.Api.Comment.CreateCommentResult>({
    url: '/comments',
    method: 'POST',
    data: payload,
  })
}

export function fetchGetPanelComments(params: App.Api.Comment.PanelListQuery) {
  const search = new URLSearchParams()
  search.set('page', String(params.page))
  search.set('page_size', String(params.page_size))
  if (params.keyword) search.set('keyword', params.keyword)
  if (params.status) search.set('status', params.status)
  if (params.echo_id) search.set('echo_id', params.echo_id)
  if (typeof params.hot === 'boolean') search.set('hot', String(params.hot))
  return request<App.Api.Comment.PanelPageResult>({
    url: `/panel/comments?${search.toString()}`,
    method: 'GET',
  })
}

export function fetchGetPanelCommentById(id: string) {
  return request<App.Api.Comment.CommentItem>({
    url: `/panel/comments/${id}`,
    method: 'GET',
  })
}

export function fetchUpdatePanelCommentStatus(id: string, status: App.Api.Comment.CommentStatus) {
  return request({
    url: `/panel/comments/${id}/status`,
    method: 'PATCH',
    data: { status },
  })
}

export function fetchUpdatePanelCommentHot(id: string, hot: boolean) {
  return request({
    url: `/panel/comments/${id}/hot`,
    method: 'PATCH',
    data: { hot },
  })
}

export function fetchDeletePanelComment(id: string) {
  return request({
    url: `/panel/comments/${id}`,
    method: 'DELETE',
  })
}

export function fetchBatchPanelComments(action: App.Api.Comment.BatchAction, ids: string[]) {
  return request({
    url: '/panel/comments/batch',
    method: 'POST',
    data: { action, ids },
  })
}

export function fetchGetCommentSystemSetting() {
  return request<App.Api.Comment.SystemSetting>({
    url: '/panel/comments/settings',
    method: 'GET',
  })
}

export function fetchUpdateCommentSystemSetting(setting: App.Api.Comment.SystemSetting) {
  return request({
    url: '/panel/comments/settings',
    method: 'PUT',
    data: setting,
  })
}

export function fetchTestCommentEmail(setting: App.Api.Comment.SystemSetting) {
  return request({
    url: '/panel/comments/settings/test-email',
    onProblem: reportProblemDetail,
    method: 'POST',
    data: { setting },
  })
}
