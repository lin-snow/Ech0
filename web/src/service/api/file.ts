// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { downloadFile, request } from '../request'

function uploadForm(dto: App.Api.File.UploadFileDto | App.Api.File.RegisterExternalFileDto) {
  const formData = new FormData()
  if ('file' in dto) {
    formData.append('file', dto.file)
    if (dto.profile) formData.append('profile', dto.profile)
  }
  if ('url' in dto) {
    formData.append('external_url', dto.url)
    if (dto.content_type) formData.append('content_type', dto.content_type)
  }
  if (dto.name) formData.append('name', dto.name)
  if ('width' in dto && dto.width) formData.append('width', String(dto.width))
  if ('height' in dto && dto.height) formData.append('height', String(dto.height))
  return formData
}

export function fetchUploadFile(dto: App.Api.File.UploadFileDto) {
  return request<App.Api.File.UploadResult>({
    url: `/files`,
    method: 'POST',
    data: uploadForm(dto),
  })
}

export function fetchRegisterExternalFile(dto: App.Api.File.RegisterExternalFileDto) {
  return request<App.Api.File.UploadResult>({
    url: `/files`,
    method: 'POST',
    data: uploadForm(dto),
  })
}

export function fetchDeleteFile(file: App.Api.File.FileDeleteDto) {
  const searchParams = new URLSearchParams()
  if (file.force) searchParams.set('force', 'true')
  const query = searchParams.toString()
  return request({
    url: `/files/${file.id}${query ? `?${query}` : ''}`,
    method: 'DELETE',
  })
}

export function fetchGetFileById(id: string) {
  return request<App.Api.File.FileDto>({
    url: `/files/${id}`,
    method: 'GET',
  })
}

export function fetchRenameFile(id: string, dto: App.Api.File.RenameFileDto) {
  return request<App.Api.File.FileDto>({
    url: `/files/${id}`,
    method: 'PATCH',
    data: dto,
  })
}

export function fetchListFiles(query: App.Api.File.FileListQuery) {
  const searchParams = new URLSearchParams({
    page: String(query.page),
    pageSize: String(query.pageSize),
    search: query.search || '',
  })
  if (query.kind) {
    searchParams.set('kind', query.kind)
  }
  if (query.access) {
    searchParams.set('access', query.access)
  }
  return request<App.Api.File.FileListResult>({
    url: `/files?${searchParams.toString()}`,
    method: 'GET',
  })
}

export function fetchDownloadFileById(id: string) {
  return downloadFile({
    url: `/files/${id}/content`,
    method: 'GET',
  })
}
