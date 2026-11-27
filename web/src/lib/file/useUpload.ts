// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, type Ref, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { FILE_CATEGORY } from '@/constants/file'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import { theToast } from '@/utils/toast'
import { formatBytes } from '@/utils/file'
import { getImageSize } from '@/utils/image'
import { getFileToAddUrl } from '@/utils/other'
import { compressImage } from './compress'
import { globalFileRegistry } from './registry/file-registry'
import { httpUpload } from './upload'

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
  /** Row is an attachment that already exists server-side (update mode/draft). */
  hydrated?: boolean
}

export interface UseUploadOptions {
  enableCompressor: Ref<boolean>
  profile?: App.Api.File.Profile
  category?: App.Api.File.Category
  allowedTypes?: string[]
  maxFiles?: number
  maxFileSize?: number
  concurrency?: number
  onAllComplete: (files: App.Api.Ech0.FileToAdd[]) => void
}

function mimeMatches(rule: string, fileType: string): boolean {
  if (rule === fileType) return true
  if (rule.endsWith('/*')) {
    return fileType.startsWith(rule.slice(0, -1))
  }
  return false
}

async function measureImage(file: File): Promise<{ width?: number; height?: number }> {
  if (!file.type.startsWith('image/')) return {}
  const objectUrl = URL.createObjectURL(file)
  try {
    const size = await getImageSize(objectUrl)
    return { width: size.width, height: size.height }
  } catch {
    return {}
  } finally {
    URL.revokeObjectURL(objectUrl)
  }
}

export function useUpload(opts: UseUploadOptions) {
  const items = ref<QueueItem[]>([])

  const userStore = useUserStore()
  const authStore = useAuthStore()
  const { t } = useI18n()

  const envURL = (import.meta.env.VITE_SERVICE_BASE_URL as string) || ''
  const backendURL = envURL.endsWith('/') ? envURL.slice(0, -1) : envURL

  const profile = opts.profile ?? 'attachment'
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

  function hydratedName(file: App.Api.Ech0.FileToAdd): string {
    if (file.name) return file.name
    try {
      const urlName = decodeURIComponent(new URL(file.url).pathname.split('/').pop() || '')
      if (urlName) return urlName
    } catch {}
    return String(file.id || 'file')
  }

  /**
   * Shows attachments that already belong to the echo (update mode, restored
   * draft) in the same list as new uploads, so both can be reordered together.
   */
  function hydrate(files: App.Api.Ech0.FileToAdd[]) {
    const known = new Set(
      items.value.map((item) => item.result?.id).filter((id): id is string => !!id),
    )

    for (const file of files) {
      if (!file.id || known.has(file.id)) continue
      const viewFile = new File([], hydratedName(file), {
        type: file.content_type || '',
      })
      items.value.push({
        id: `existing:${file.id}`,
        file: viewFile,
        originalFile: viewFile,
        originalSize: file.size ?? 0,
        status: UPLOAD_STATUS.SUCCESS,
        progress: 100,
        preview: getFileToAddUrl(file),
        result: file,
        delivered: true,
        hydrated: true,
      })
      known.add(file.id)
    }
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

      const result = await uploadMultipart(working, item, ctl.signal)

      item.result = result
      item.status = UPLOAD_STATUS.SUCCESS
      item.progress = 100

      globalFileRegistry.upsert({
        id: result.id || '',
        name: result.name,
        url: result.url,
        category: result.category,
        contentType: result.content_type,
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

  async function uploadMultipart(
    file: File,
    item: QueueItem,
    signal: AbortSignal,
  ): Promise<App.Api.Ech0.FileToAdd> {
    const dimensions = await measureImage(file)
    const fields: Record<string, string> = {
      profile,
      name: file.name,
    }
    if (dimensions.width) fields.width = String(dimensions.width)
    if (dimensions.height) fields.height = String(dimensions.height)

    const res = await httpUpload(
      file,
      {
        endpoint: `${backendURL}/api/files`,
        authHeader: authStore.authHeader,
        fields,
      },
      {
        signal,
        onProgress: (loaded, total) => {
          if (total > 0) item.progress = Math.round((loaded / total) * 100)
        },
      },
    )

    // The upload endpoint answers the bare create result: `{ file, upload? }`.
    const body: unknown = res.responseBody
    if (!body || typeof body !== 'object' || !('file' in body)) {
      throw new Error(String(t('uploader.missingFileIdentifier')))
    }
    const rawFile = body.file
    if (!rawFile || typeof rawFile !== 'object') {
      throw new Error(String(t('uploader.missingFileIdentifier')))
    }
    const payload = rawFile as Record<string, unknown>
    const fileId = pickField(payload, ['id'])
    const fileUrl = pickField(payload, ['url'])
    if (!fileId || !fileUrl) {
      throw new Error(String(t('uploader.missingFileIdentifier')))
    }
    const payloadCategory = pickField(payload, ['category'])
    return {
      id: fileId,
      url: fileUrl,
      name: typeof payload.name === 'string' ? payload.name : file.name,
      content_type: typeof payload.content_type === 'string' ? payload.content_type : file.type,
      size: typeof payload.size === 'number' ? payload.size : file.size,
      width: typeof payload.width === 'number' ? payload.width : dimensions.width,
      height: typeof payload.height === 'number' ? payload.height : dimensions.height,
      category: payloadCategory
        ? (payloadCategory as App.Api.File.Category)
        : (category as App.Api.File.Category),
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
    hydrate,
    retry,
    remove,
    moveItem,
    cancelAll,
    reset,
  }
}
