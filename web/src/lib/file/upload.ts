// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const UPLOAD_KIND = {
  LOCAL: 'local',
  S3: 's3',
} as const

export type UploadKind = (typeof UPLOAD_KIND)[keyof typeof UPLOAD_KIND]

export type UploadTarget =
  | {
      kind: typeof UPLOAD_KIND.LOCAL
      endpoint: string
      authHeader: string
      fields: Record<string, string>
    }
  | {
      kind: typeof UPLOAD_KIND.S3
      presignUrl: string
      contentType: string
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

function parseResponseBody(text: string): unknown {
  if (!text) return null
  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

function extractErrorMessage(body: unknown): string | null {
  if (body && typeof body === 'object' && 'msg' in body) {
    const msg = (body as { msg?: unknown }).msg
    if (typeof msg === 'string' && msg.trim()) return msg
  }
  return null
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

    if (target.kind === UPLOAD_KIND.LOCAL) {
      xhr.open('POST', target.endpoint, true)
      if (target.authHeader) {
        xhr.setRequestHeader('Authorization', target.authHeader)
      }
    } else {
      xhr.open('PUT', target.presignUrl, true)
      xhr.setRequestHeader('Content-Type', target.contentType)
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
      const body = parseResponseBody(xhr.responseText)
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve({ responseBody: body })
      } else {
        const msg = extractErrorMessage(body) || `Upload failed (${xhr.status})`
        reject(new UploadError(xhr.status, body, msg))
      }
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

    if (target.kind === UPLOAD_KIND.LOCAL) {
      const fd = new FormData()
      fd.append('file', file)
      for (const [k, v] of Object.entries(target.fields)) {
        fd.append(k, v)
      }
      xhr.send(fd)
    } else {
      xhr.send(file)
    }
  })
}
