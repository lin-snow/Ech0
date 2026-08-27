// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, type Ref, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { theToast } from '@/utils/toast'
import { formatBytes } from '@/utils/file'
import { getImageSize } from '@/utils/image'
import { compressImage, inferFileExtFromType } from './compress'
import { getPresign, updateFileMeta } from './api/adapter'
import { globalFileRegistry } from './registry/file-registry'
import { httpUpload, UPLOAD_KIND } from './upload'

export const UPLOAD_STATUS = {
  PENDING: 'pending',
  COMPRESSING: 'compressing',
  UPLOADING: 'uploading',
  SUCCESS: 'success',
  ERROR: 'error',
  CANCELLED: 'cancelled',
} as const

export type UploadStatus = (typeof UPLOAD_STATUS)[keyof typeof UPLOAD_STATUS]

const ACTIVE_STATUSES = [
  UPLOAD_STATUS.PENDING,
  UPLOAD_STATUS.COMPRESSING,
  UPLOAD_STATUS.UPLOADING,
] as const satisfies readonly UploadStatus[]

function isActiveStatus(status: UploadStatus): boolean {
  return (ACTIVE_STATUSES as readonly UploadStatus[]).includes(status)
}

function pickField(obj: Record<string, unknown>, keys: readonly string[]): string {
  for (const k of keys) {
    const v = obj[k]
    if (v != null && v !== '') return String(v)
  }
  return ''
}

export interface QueueItem {
  id: string
  file: File
  originalFile: File
  originalSize: number
  compressionAttempted?: boolean
  status: UploadStatus
  progress: number
  error?: string
  preview: string
  result?: App.Api.Ech0.FileToAdd
  abort?: AbortController
  delivered?: boolean
}

export interface UseUploadOptions {
  storageType: Ref<App.Api.File.StorageType>
  enableCompressor: Ref<boolean>
  category?: App.Api.File.Category
  allowedTypes?: string[]
  maxFiles?: number
  maxFileSize?: number
  concurrency?: number
  onAllComplete: (files: App.Api.Ech0.FileToAdd[]) => void
}

function extractUploadPayload(response: unknown): Record<string, unknown> {
  const resp = (response || {}) as Record<string, unknown>
  const data = resp.data
  if (data && typeof data === 'object') return data as Record<string, unknown>
  return resp
}

function mimeMatches(rule: string, fileType: string): boolean {
  if (rule === fileType) return true
  if (rule.endsWith('/*')) {
    return fileType.startsWith(rule.slice(0, -1))
  }
  return false
}

export function useUpload(opts: UseUploadOptions) {
  const items = ref<QueueItem[]>([])

  const userStore = useUserStore()
  const authStore = useAuthStore()
  const { t } = useI18n()

  const envURL = (import.meta.env.VITE_SERVICE_BASE_URL as string) || ''
  const backendURL = envURL.endsWith('/') ? envURL.slice(0, -1) : envURL

  const category = opts.category ?? FILE_CATEGORY.IMAGE
  const allowedTypes = opts.allowedTypes?.length ? opts.allowedTypes : ['image/*']
  const maxFiles = opts.maxFiles ?? 6
  const maxFileSize = opts.maxFileSize
  const concurrency = opts.concurrency ?? 3

  let inFlight = 0

  const isUploading = computed(() => items.value.some((i) => isActiveStatus(i.status)))
  const totalActive = computed(() => items.value.filter((i) => isActiveStatus(i.status)).length)

  function genId(file: File): string {
    return `${file.name}|${file.size}|${file.lastModified}`
  }

  function isAllowed(file: File): boolean {
    return allowedTypes.some((rule) => mimeMatches(rule, file.type))
  }

  function liveCount(): number {
    return items.value.filter(
      (i) => i.status !== UPLOAD_STATUS.CANCELLED && i.status !== UPLOAD_STATUS.ERROR,
    ).length
  }

  function addFiles(files: File[]) {
    if (!userStore.isLogin) {
      theToast.error(String(t('uploader.loginRequired')))
      return
    }

    const existingIds = new Set(items.value.map((i) => i.id))
    let remaining = Math.max(0, maxFiles - liveCount())
    let rejected = false
    const oversized: string[] = []

    for (const file of files) {
      if (remaining <= 0) {
        rejected = true
        break
      }
      if (!isAllowed(file)) continue
      if (maxFileSize != null && file.size > maxFileSize) {
        oversized.push(file.name)
        continue
      }
      const id = genId(file)
      if (existingIds.has(id)) continue
      existingIds.add(id)

      items.value.push({
        id,
        file,
        originalFile: file,
        originalSize: file.size,
        status: UPLOAD_STATUS.PENDING,
        progress: 0,
        preview: URL.createObjectURL(file),
      })
      remaining--
    }

    if (rejected) {
      theToast.info(String(t('uploader.maxFilesReached', { max: maxFiles })))
    }
    if (oversized.length > 0 && maxFileSize != null) {
      theToast.error(
        String(
          t('uploader.fileTooLarge', {
            name: oversized[0] + (oversized.length > 1 ? ` (+${oversized.length - 1})` : ''),
            max: formatBytes(maxFileSize),
          }),
        ),
      )
    }

    pump()
  }

  function pump() {
    while (inFlight < concurrency) {
      const next = items.value.find((i) => i.status === UPLOAD_STATUS.PENDING)
      if (!next) break
      void processItem(next)
    }

    const stillWorking = items.value.some((i) => isActiveStatus(i.status))
    if (!stillWorking && inFlight === 0) {
      const undelivered = items.value.filter(
        (i) => i.status === UPLOAD_STATUS.SUCCESS && i.result && !i.delivered,
      )
      if (undelivered.length > 0) {
        for (const item of undelivered) item.delivered = true
        opts.onAllComplete(undelivered.map((i) => i.result!))
      }
    }
  }

  async function processItem(item: QueueItem) {
    inFlight++
    const ctl = new AbortController()
    item.abort = ctl

    try {
      let working = item.file
      if (opts.enableCompressor.value && working.type.startsWith('image/')) {
        item.status = UPLOAD_STATUS.COMPRESSING
        item.compressionAttempted = true
        try {
          working = await compressImage(working)
          item.file = working
        } catch (err) {
          console.warn('[uploader] compression failed, using original file:', err)
        }
      }

      if (ctl.signal.aborted) {
        item.status = UPLOAD_STATUS.CANCELLED
        return
      }

      item.status = UPLOAD_STATUS.UPLOADING
      item.progress = 0

      const result =
        opts.storageType.value === FILE_STORAGE_TYPE.OBJECT
          ? await uploadToS3(working, item, ctl.signal)
          : await uploadToLocal(working, item, ctl.signal)

      item.result = result
      item.status = UPLOAD_STATUS.SUCCESS
      item.progress = 100

      globalFileRegistry.upsert({
        id: result.id || '',
        key: result.key,
        url: result.url,
        category: result.category,
        contentType: result.content_type,
        storageType: result.storage_type,
        size: result.size,
        width: result.width,
        height: result.height,
      })
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        item.status = UPLOAD_STATUS.CANCELLED
      } else {
        item.status = UPLOAD_STATUS.ERROR
        const msg =
          (err instanceof Error ? err.message : String(err)) || String(t('uploader.uploadError'))
        item.error = msg
        theToast.error(msg)
      }
    } finally {
      item.abort = undefined
      inFlight--
      pump()
    }
  }

  async function uploadToLocal(
    file: File,
    item: QueueItem,
    signal: AbortSignal,
  ): Promise<App.Api.Ech0.FileToAdd> {
    const res = await httpUpload(
      file,
      {
        kind: UPLOAD_KIND.LOCAL,
        endpoint: `${backendURL}/api/files/upload`,
        authHeader: authStore.authHeader,
        fields: {
          category,
          storage_type: FILE_STORAGE_TYPE.LOCAL,
        },
      },
      {
        signal,
        onProgress: (loaded, total) => {
          if (total > 0) item.progress = Math.round((loaded / total) * 100)
        },
      },
    )

    const payload = extractUploadPayload(res.responseBody)
    const fileId = pickField(payload, ['id', 'file_id', 'ID'])
    const fileKey = pickField(payload, ['key', 'object_key'])
    const fileUrl = pickField(payload, ['url', 'file_url', 'access_url'])
    if (!fileId || !fileUrl) {
      throw new Error(String(t('uploader.missingFileIdentifier')))
    }
    return {
      id: fileId,
      url: fileUrl,
      storage_type: FILE_STORAGE_TYPE.LOCAL,
      key: fileKey,
      content_type: typeof payload.content_type === 'string' ? payload.content_type : file.type,
      size: typeof payload.size === 'number' ? payload.size : file.size,
      width: typeof payload.width === 'number' ? payload.width : undefined,
      height: typeof payload.height === 'number' ? payload.height : undefined,
      category,
    }
  }

  async function uploadToS3(
    file: File,
    item: QueueItem,
    signal: AbortSignal,
  ): Promise<App.Api.Ech0.FileToAdd> {
    const contentType = file.type || 'application/octet-stream'
    const rawName = String(file.name || '').trim()
    const fileName = rawName || `upload_${Date.now()}${inferFileExtFromType(contentType)}`

    const presign = await getPresign({
      fileName,
      contentType,
      storageType: FILE_STORAGE_TYPE.OBJECT,
    })

    if (signal.aborted) throw new DOMException('Aborted', 'AbortError')

    await httpUpload(
      file,
      {
        kind: UPLOAD_KIND.S3,
        presignUrl: presign.presign_url,
        contentType,
      },
      {
        signal,
        onProgress: (loaded, total) => {
          if (total > 0) item.progress = Math.round((loaded / total) * 100)
        },
      },
    )

    let width: number | undefined
    let height: number | undefined
    if (category === FILE_CATEGORY.IMAGE) {
      try {
        const dim = await getImageSize(item.preview)
        width = dim.width
        height = dim.height
      } catch {}
    }

    let resolvedSize = file.size
    let resolvedWidth = width
    let resolvedHeight = height
    let resolvedContentType: string | undefined = contentType
    try {
      const updated = await updateFileMeta({
        id: presign.id,
        size: file.size,
        width,
        height,
        contentType,
      })
      resolvedSize = updated.size ?? resolvedSize
      resolvedWidth = updated.width
      resolvedHeight = updated.height
      resolvedContentType = updated.contentType ?? resolvedContentType
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(t('uploader.fileMetaFillbackFailed'))
      theToast.error(msg)
    }

    return {
      id: presign.id,
      url: presign.file_url,
      storage_type: FILE_STORAGE_TYPE.OBJECT,
      key: presign.key,
      content_type: resolvedContentType,
      size: resolvedSize,
      width: resolvedWidth,
      height: resolvedHeight,
      category,
    }
  }

  function retry(id: string) {
    const item = items.value.find((i) => i.id === id)
    if (!item) return
    if (item.status !== UPLOAD_STATUS.ERROR && item.status !== UPLOAD_STATUS.CANCELLED) return
    item.status = UPLOAD_STATUS.PENDING
    item.error = undefined
    item.progress = 0
    item.file = item.originalFile
    item.compressionAttempted = false
    item.delivered = false
    pump()
  }

  function moveItem(fromId: string, toId: string) {
    if (fromId === toId) return
    const from = items.value.findIndex((i) => i.id === fromId)
    const to = items.value.findIndex((i) => i.id === toId)
    if (from < 0 || to < 0) return
    const [moved] = items.value.splice(from, 1)
    items.value.splice(to, 0, moved)
  }

  function remove(id: string) {
    const idx = items.value.findIndex((i) => i.id === id)
    if (idx < 0) return
    const item = items.value[idx]
    if (item.abort) item.abort.abort()
    URL.revokeObjectURL(item.preview)
    items.value.splice(idx, 1)
    pump()
  }

  function cancelAll() {
    for (const item of items.value) {
      if (item.abort) item.abort.abort()
      URL.revokeObjectURL(item.preview)
    }
    items.value = []
  }

  function reset() {
    cancelAll()
  }

  return {
    items,
    isUploading,
    totalActive,
    addFiles,
    retry,
    remove,
    moveItem,
    cancelAll,
    reset,
  }
}
