// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'

export type FileCategory = (typeof FILE_CATEGORY)[keyof typeof FILE_CATEGORY]
export type FileStorageType = (typeof FILE_STORAGE_TYPE)[keyof typeof FILE_STORAGE_TYPE]

export type FileEntity = {
  id: string
  key?: string
  url: string
  contentType?: string
  category?: FileCategory
  storageType?: FileStorageType
  size?: number
  width?: number
  height?: number
}

export type FileUploadInput = {
  file: File
  category: FileCategory
  storageType?: FileStorageType
}

export type ExternalFileInput = {
  url: string
  category?: FileCategory
  contentType?: string
  width?: number
  height?: number
  name?: string
}

export type PresignInput = {
  category?: FileCategory
  size?: number
  fileName: string
  contentType?: string
  storageType?: FileStorageType
}

export type PresignResult = {
  id: string
  file_name: string
  content_type: string
  key: string
  presign_url: string
  require_create_only?: boolean
  file_url: string
}

export type UpdateFileMetaInput = {
  id: string
  size: number
  width?: number
  height?: number
  contentType?: string
}

export type FileAttachment = {
  id?: string
  key?: string
  url: string
  category?: FileCategory
  content_type?: string
  storage_type: FileStorageType
  size?: number
  width?: number
  height?: number
}

export type FileValidationRule = {
  requireId?: boolean
  maxCount?: number
}

export type QueueTaskStatus = 'queued' | 'running' | 'retrying' | 'success' | 'failed' | 'cancelled'

export type QueueTask = {
  id: string
  name: string
  status: QueueTaskStatus
  attempt: number
  progress: number
  error?: string
  result?: FileEntity
  input: FileUploadInput
}

export type QueueOptions = {
  concurrency?: number
  maxRetry?: number
}
