// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { FILE_CATEGORY } from '@/constants/file'

export type FileCategory = (typeof FILE_CATEGORY)[keyof typeof FILE_CATEGORY]

export type FileEntity = {
  id: string
  name?: string
  url: string
  contentType?: string
  category?: FileCategory
  kind?: App.Api.File.Kind
  access?: App.Api.File.Access
  size?: number
  width?: number
  height?: number
  createdAt?: number
}

export type FileUploadInput = {
  file: File
  profile?: App.Api.File.Profile
}

export type ExternalFileInput = {
  url: string
  name?: string
  contentType?: string
  width?: number
  height?: number
}

export type FileAttachment = {
  id?: string
  name?: string
  url: string
  category?: FileCategory
  content_type?: string
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
