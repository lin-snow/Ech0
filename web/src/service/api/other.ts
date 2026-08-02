// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request, downloadFile } from '../request'

// Hello Ech0
export function fetchHelloEch0() {
  return request<App.Api.Ech0.HelloEch0>({
    url: '/hello',
    method: 'GET',
  })
}

// 导出产物（下载）- 使用专门的下载函数；format 缺省时后端按 snapshot 处理，故不拼 query。
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
    // 失败提示由控制台自行展示，避免与 request 全局错误 Toast 重复
    silentError: true,
  })
}

// 获取网站标题
export function fetchGetWebsiteTitle(websiteURL: string) {
  return request<string>({
    url: `/website/title?website_url=${encodeURIComponent(websiteURL)}`,
    method: 'GET',
  })
}

// 迁移来源:ech0 快照、memos 导出,以及 Ech0 胶囊(内容交换格式)。
export type MigrationSourceType = 'ech0' | 'memos' | 'capsule'

export interface StartMigrationPayload {
  source_type: MigrationSourceType
  // 胶囊导入时这里额外携带 include_private,决定是否一并写入胶囊内的私密 echo。
  source_payload: Record<string, unknown>
}

export interface MigrationStatusPayload extends StartMigrationPayload {
  version: number
  status: 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  // 后端 GlobalMigrationStateDTO 一直回传细粒度阶段(extracting/loading/reporting/completed),
  // 之前前端契约漏了它,补上以驱动 job 进度步进器。
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

// 导出（手动快照异步出口）：与导入对称，统一收敛到 Migrator 域，走 export 作业（job.Manager）。
// 两种产物语义完全不同：snapshot 是整个 data/ 的 zip（含账号与凭据，唯一可灾难恢复）；
// capsule 只含内容（人类可读，可分享/搬家/编静态站），二者不可互换。
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
    // 上传后端还会执行解压，耗时可能显著高于默认请求超时。
    timeout: 30 * 60 * 1000,
    data: formData,
  })
}
