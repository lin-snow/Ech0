// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { downloadFile, request } from '../request'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'

export function fetchUploadFile(
  file: File,
  storageType: App.Api.File.StorageType = FILE_STORAGE_TYPE.LOCAL,
  category: App.Api.File.Category = FILE_CATEGORY.IMAGE,
) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('storage_type', storageType)
  formData.append('category', category)

  return request<App.Api.File.FileDto>({
    url: `/files/upload`,
    method: 'POST',
    data: formData,
  })
}

export function fetchCreateExternalFile(dto: App.Api.File.CreateExternalFileDto) {
  return request<App.Api.File.FileDto>({
    url: `/files/external`,
    method: 'POST',
    data: dto,
  })
}

export function fetchDeleteFile(file: App.Api.File.FileDeleteDto) {
  return request({
    url: `/file/${file.id}`,
    method: 'DELETE',
  })
}

export function fetchGetFileById(id: string) {
  return request<App.Api.File.FileDto>({
    url: `/file/${id}`,
    method: 'GET',
  })
}

export function fetchUpdateFileMeta(id: string, dto: App.Api.File.UpdateFileMetaDto) {
  return request<App.Api.File.FileDto>({
    url: `/file/${id}/meta`,
    method: 'PUT',
    data: dto,
  })
}

export function fetchGetPresignedUrl(
  fileName: string,
  contentType?: string,
  storageType: App.Api.File.StorageType = FILE_STORAGE_TYPE.OBJECT,
) {
  return request<App.Api.Ech0.PresignResult>({
    url: `/files/presign`,
    method: 'PUT',
    data: {
      file_name: fileName,
      content_type: contentType,
      storage_type: storageType,
    },
  })
}

export function fetchListFiles(query: App.Api.File.FileListQuery) {
  const searchParams = new URLSearchParams({
    page: String(query.page),
    pageSize: String(query.pageSize),
    search: query.search || '',
  })
  if (query.storage_type) {
    searchParams.set('storage_type', query.storage_type)
  }
  return request<App.Api.File.FileListResult>({
    url: `/files?${searchParams.toString()}`,
    method: 'GET',
  })
}

export function fetchFileTree(query: App.Api.File.FileTreeQuery) {
  const searchParams = new URLSearchParams({
    storage_type: query.storage_type,
  })
  if (query.prefix) {
    searchParams.set('prefix', query.prefix)
  }
  return request<App.Api.File.FileTreeResult>({
    url: `/file/tree?${searchParams.toString()}`,
    method: 'GET',
  })
}

export function fetchDownloadFileById(id: string) {
  return downloadFile({
    url: `/file/${id}/stream`,
    method: 'GET',
  })
}

export function fetchDownloadFileByPath(query: App.Api.File.FilePathStreamQuery) {
  const searchParams = new URLSearchParams({
    storage_type: query.storage_type,
    path: query.path,
  })
  if (query.name) {
    searchParams.set('name', query.name)
  }
  if (query.content_type) {
    searchParams.set('content_type', query.content_type)
  }
  return downloadFile({
    url: `/file/stream?${searchParams.toString()}`,
    method: 'GET',
  })
}
