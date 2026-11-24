// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import {
  fetchDeleteFile,
  fetchGetFileById,
  fetchRegisterExternalFile,
  fetchRenameFile,
  fetchUploadFile,
} from '@/service/api'
import { getApiUrl } from '@/service/request/shared'
import { useAuthStore } from '@/stores/auth'
import type { ExternalFileInput, FileEntity, FileUploadInput } from '../types'

function normalizeFileEntity(dto: App.Api.File.FileDto): FileEntity {
  return {
    id: String(dto.id || ''),
    name: dto.name,
    url: dto.url || '',
    contentType: dto.content_type,
    category: dto.category,
    kind: dto.kind,
    access: dto.access,
    size: dto.size,
    width: dto.width,
    height: dto.height,
    createdAt: dto.created_at,
  }
}

// The transport answers `null` on failure (it has already toasted the problem),
// so a missing body is the only failure signal a caller can see.
function ensureSuccess<T>(res: T | null, fallbackMsg: string): T {
  if (!res) {
    throw new Error(fallbackMsg)
  }
  return res
}

// A void endpoint answers 204, which the transport surfaces as `undefined`.
function ensureDone(res: unknown, fallbackMsg: string): void {
  if (res === null) {
    throw new Error(fallbackMsg)
  }
}

export async function uploadFile(input: FileUploadInput): Promise<FileEntity> {
  const res = await fetchUploadFile({
    file: input.file,
    profile: input.profile,
  })
  const data = ensureSuccess(res, '文件上传失败')
  return normalizeFileEntity(data.file)
}

export async function createExternalFile(input: ExternalFileInput): Promise<FileEntity> {
  const res = await fetchRegisterExternalFile({
    url: input.url,
    name: input.name,
    content_type: input.contentType,
    width: input.width,
    height: input.height,
  })
  const data = ensureSuccess(res, '外链文件注册失败')
  return normalizeFileEntity(data.file)
}

export async function getFileById(id: string): Promise<FileEntity> {
  const res = await fetchGetFileById(id)
  const data = ensureSuccess(res, '文件查询失败')
  return normalizeFileEntity(data)
}

export async function deleteFileById(id: string, force = false): Promise<void> {
  const res = await fetchDeleteFile({ id, force })
  ensureDone(res, '文件删除失败')
}

export async function renameFile(id: string, name: string): Promise<FileEntity> {
  const res = await fetchRenameFile(id, { name })
  const data = ensureSuccess(res, '文件重命名失败')
  return normalizeFileEntity(data)
}

export function buildStreamUrl(fileId: string, t = Date.now(), token?: string): string {
  const resolved = token ?? useAuthStore().accessToken
  const tokenQuery = resolved ? `&token=${encodeURIComponent(resolved)}` : ''
  return `${getApiUrl()}/files/${fileId}/content?t=${t}${tokenQuery}`
}

export const fileApiAdapter = {
  uploadFile,
  createExternalFile,
  getFileById,
  deleteFileById,
  renameFile,
  buildStreamUrl,
}

export type FileApiAdapter = typeof fileApiAdapter
