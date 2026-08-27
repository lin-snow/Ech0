// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request } from '../request'

export function fetchGetEmbeddingSettings() {
  return request<App.Api.Embedding.EmbeddingSetting>({
    url: '/embedding/settings',
    method: 'GET',
  })
}

export function fetchUpdateEmbeddingSettings(data: App.Api.Embedding.EmbeddingSettingDto) {
  return request<null>({
    url: '/embedding/settings',
    method: 'PUT',
    data,
  })
}

export function fetchReindexEmbeddings() {
  return request<App.Api.Embedding.ReindexStatus>({
    url: '/embedding/reindex',
    method: 'POST',
  })
}

export function fetchReindexStatus() {
  return request<App.Api.Embedding.ReindexStatus>({
    url: '/embedding/reindex/status',
    method: 'GET',
  })
}

export function fetchCancelReindex() {
  return request<App.Api.Embedding.ReindexStatus>({
    url: '/embedding/reindex/cancel',
    method: 'POST',
  })
}
