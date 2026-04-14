import { FILE_STORAGE_TYPE } from '@/constants/file'
import {
  fetchCreateExternalFile,
  fetchDeleteFile,
  fetchGetFileById,
  fetchGetPresignedUrl,
  fetchUpdateFileMeta,
  fetchUploadFile,
} from '@/service/api'
import { getApiUrl } from '@/service/request/shared'
import { useAuthStore } from '@/stores/auth'
import type {
  ExternalFileInput,
  FileEntity,
  FileUploadInput,
  PresignInput,
  PresignResult,
  UpdateFileMetaInput,
} from '../types'

function normalizeFileEntity(dto: App.Api.File.FileDto): FileEntity {
  return {
    id: String(dto.id || ''),
    key: dto.key || '',
    url: dto.url || '',
    contentType: dto.content_type,
    category: dto.category,
    storageType: dto.storage_type,
    size: dto.size,
    width: dto.width,
    height: dto.height,
  }
}

function ensureSuccess<T>(res: App.Api.Response<T>, fallbackMsg: string): T {
  if (res.code !== 1) {
    throw new Error(res.msg || fallbackMsg)
  }
  return res.data
}

export async function uploadFile(input: FileUploadInput): Promise<FileEntity> {
  const res = await fetchUploadFile(
    input.file,
    input.storageType || FILE_STORAGE_TYPE.LOCAL,
    input.category,
  )
  const data = ensureSuccess(res, '文件上传失败')
  return normalizeFileEntity(data)
}

export async function createExternalFile(input: ExternalFileInput): Promise<FileEntity> {
  const res = await fetchCreateExternalFile({
    url: input.url,
    category: input.category,
    content_type: input.contentType,
    width: input.width,
    height: input.height,
    name: input.name,
  })
  const data = ensureSuccess(res, '外链文件注册失败')
  return normalizeFileEntity(data)
}

export async function getFileById(id: string): Promise<FileEntity> {
  const res = await fetchGetFileById(id)
  const data = ensureSuccess(res, '文件查询失败')
  return normalizeFileEntity(data)
}

export async function deleteFileById(id: string): Promise<void> {
  const res = await fetchDeleteFile({ id })
  ensureSuccess(res, '文件删除失败')
}

export async function getPresign(input: PresignInput): Promise<PresignResult> {
  const res = await fetchGetPresignedUrl(
    input.fileName,
    input.contentType,
    input.storageType || FILE_STORAGE_TYPE.OBJECT,
  )
  return ensureSuccess(res, '预签名 URL 获取失败')
}

export async function updateFileMeta(input: UpdateFileMetaInput): Promise<FileEntity> {
  const res = await fetchUpdateFileMeta(input.id, {
    size: input.size,
    width: input.width,
    height: input.height,
    content_type: input.contentType,
  })
  const data = ensureSuccess(res, '文件元数据回填失败')
  return normalizeFileEntity(data)
}

export function buildStreamUrl(fileId: string, t = Date.now()): string {
  const token = useAuthStore().accessToken
  const tokenQuery = token ? `&token=${encodeURIComponent(token)}` : ''
  return `${getApiUrl()}/file/${fileId}/stream?t=${t}${tokenQuery}`
}

export const fileApiAdapter = {
  uploadFile,
  createExternalFile,
  getFileById,
  deleteFileById,
  getPresign,
  updateFileMeta,
  buildStreamUrl,
}

export type FileApiAdapter = typeof fileApiAdapter
