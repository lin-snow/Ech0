// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request, downloadFile } from '../request'

export function fetchHelloEch0() {
  return request<App.Api.Ech0.HelloEch0>({
    url: '/hello',
    method: 'GET',
  })
}

export function fetchDownloadExport(format?: ExportFormat) {
  return downloadFile({
    url: format ? `/migration/export/download?format=${format}` : '/migration/export/download',
    method: 'GET',
  })
}

export type CheckUpdateResult = {
  current_version: string
  latest_version: string
  has_update: boolean
}

export function fetchCheckUpdate() {
  return request<CheckUpdateResult>({
    url: '/system/check-update',
    method: 'GET',
    silentError: true,
  })
}

export function fetchGetWebsiteTitle(websiteURL: string) {
  return request<string>({
    url: `/website/title?website_url=${encodeURIComponent(websiteURL)}`,
    method: 'GET',
  })
}

export type MigrationSourceType = 'ech0' | 'memos' | 'capsule'

export interface StartMigrationPayload {
  source_type: MigrationSourceType
  source_payload: Record<string, unknown>
}

export interface MigrationStatusPayload extends StartMigrationPayload {
  version: number
  status: 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  phase?: string
  error_message: string
  started_at?: number
  updated_at?: number
  finished_at?: number
}

export function fetchStartMigration(data: StartMigrationPayload) {
  return request({
    url: '/migration/start',
    method: 'POST',
    data,
  })
}

export function fetchGetMigrationStatus() {
  return request<MigrationStatusPayload>({
    url: '/migration/status',
    method: 'GET',
  })
}

export function fetchCancelMigration() {
  return request<MigrationStatusPayload>({
    url: '/migration/cancel',
    method: 'POST',
  })
}

export function fetchCleanupMigration() {
  return request({
    url: '/migration/cleanup',
    method: 'POST',
  })
}

export type ExportFormat = 'snapshot' | 'capsule'

export interface ExportStatusPayload {
  version: number
  status: 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  phase?: string
  error_message: string
  file_name?: string
  size?: number
  format?: ExportFormat
  started_at?: number
  updated_at?: number
  finished_at?: number
}

export function fetchStartExport(params?: { format?: ExportFormat; include_private?: boolean }) {
  return request<ExportStatusPayload>({
    url: '/migration/export',
    method: 'POST',
    data: params,
  })
}

export function fetchGetExportStatus() {
  return request<ExportStatusPayload>({
    url: '/migration/export/status',
    method: 'GET',
  })
}

export function fetchCancelExport() {
  return request<ExportStatusPayload>({
    url: '/migration/export/cancel',
    method: 'POST',
  })
}

export interface UploadMigrationSourceZipResponse {
  source_type: MigrationSourceType
  tmp_dir: string
  source_payload: Record<string, unknown>
}

export function fetchUploadMigrationSourceZip(
  sourceType: UploadMigrationSourceZipResponse['source_type'],
  file: File,
) {
  const formData = new FormData()
  formData.append('source_type', sourceType)
  formData.append('file', file)
  return request<UploadMigrationSourceZipResponse>({
    url: '/migration/upload',
    method: 'POST',
    timeout: 30 * 60 * 1000,
    data: formData,
  })
}
