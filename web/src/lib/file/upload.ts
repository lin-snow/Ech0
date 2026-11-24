// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { problemMessage } from '@/service/request/problem'

export interface UploadTarget {
  endpoint: string
  authHeader: string
  fields: Record<string, string>
}

export interface UploadHooks {
  onProgress?: (loaded: number, total: number) => void
  signal?: AbortSignal
}

export interface UploadResult {
  responseBody: unknown
}

export class UploadError extends Error {
  constructor(
    public readonly status: number,
    public readonly body: unknown,
    message: string,
  ) {
    super(message)
    this.name = 'UploadError'
  }
}

export function httpUpload(
  file: File,
  target: UploadTarget,
  hooks: UploadHooks = {},
): Promise<UploadResult> {
  return new Promise((resolve, reject) => {
    if (hooks.signal?.aborted) {
      reject(new DOMException('Aborted', 'AbortError'))
      return
    }

    const xhr = new XMLHttpRequest()
    xhr.open('POST', target.endpoint, true)
    if (target.authHeader) {
      xhr.setRequestHeader('Authorization', target.authHeader)
    }

    if (hooks.onProgress) {
      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) hooks.onProgress!(e.loaded, e.total)
      }
    }

    const onAbort = () => xhr.abort()
    if (hooks.signal) hooks.signal.addEventListener('abort', onAbort)
    const cleanup = () => {
      if (hooks.signal) hooks.signal.removeEventListener('abort', onAbort)
    }

    xhr.onload = () => {
      cleanup()
      let body: unknown = null
      if (xhr.responseText) {
        try {
          body = JSON.parse(xhr.responseText)
        } catch {
          body = xhr.responseText
        }
      }
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve({ responseBody: body })
        return
      }
      const problem: App.Api.Problem =
        body && typeof body === 'object' && 'status' in body
          ? (body as App.Api.Problem)
          : { status: xhr.status }
      reject(new UploadError(xhr.status, body, problemMessage(problem)))
    }
    xhr.onerror = () => {
      cleanup()
      reject(new UploadError(0, null, 'Network error during upload'))
    }
    xhr.ontimeout = () => {
      cleanup()
      reject(new UploadError(0, null, 'Upload timed out'))
    }
    xhr.onabort = () => {
      cleanup()
      reject(new DOMException('Aborted', 'AbortError'))
    }

    const fd = new FormData()
    fd.append('file', file)
    for (const [k, v] of Object.entries(target.fields)) {
      fd.append(k, v)
    }
    xhr.send(fd)
  })
}
