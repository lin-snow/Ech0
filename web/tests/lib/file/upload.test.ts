// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { problemMessage } from '@/service/request/problem'

import { httpUpload, UploadError } from '../../../src/lib/file/upload'

class FakeXHR {
  method?: string
  url?: string
  headers: Record<string, string> = {}
  body?: unknown
  status = 0
  responseText = ''
  upload: { onprogress: ((e: ProgressEvent) => void) | null } = { onprogress: null }
  onload: (() => void) | null = null
  onerror: (() => void) | null = null
  ontimeout: (() => void) | null = null
  onabort: (() => void) | null = null

  constructor() {
    // eslint-disable-next-line @typescript-eslint/no-this-alias
    last = this
  }

  open(method: string, url: string): void {
    this.method = method
    this.url = url
  }
  setRequestHeader(name: string, value: string): void {
    this.headers[name] = value
  }
  send(body?: unknown): void {
    this.body = body
  }
  abort(): void {
    this.onabort?.()
  }

  __succeed(status: number, responseText: string): void {
    this.status = status
    this.responseText = responseText
    this.onload?.()
  }
  __fail(): void {
    this.onerror?.()
  }
  __progress(loaded: number, total: number): void {
    this.upload.onprogress?.({ lengthComputable: true, loaded, total } as ProgressEvent)
  }
}

let last: FakeXHR

beforeEach(() => {
  vi.stubGlobal('XMLHttpRequest', FakeXHR as unknown as typeof XMLHttpRequest)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function fakeFile(): File {
  return new File([new Uint8Array(10)], 'a.png', { type: 'image/png' })
}

describe('httpUpload — local POST', () => {
  it('sends multipart with auth header and resolves with the parsed body', async () => {
    const file = fakeFile()
    const promise = httpUpload(file, {
      endpoint: '/api/files/upload',
      authHeader: 'Bearer xyz',
      fields: { category: 'image', storage_type: 'local' },
    })
    queueMicrotask(() =>
      last.__succeed(200, JSON.stringify({ id: 'F1', url: '/x' })),
    )
    const res = await promise
    expect(last.method).toBe('POST')
    expect(last.url).toBe('/api/files/upload')
    expect(last.headers['Authorization']).toBe('Bearer xyz')
    expect(last.body).toBeInstanceOf(FormData)
    const fd = last.body as FormData
    expect(fd.get('category')).toBe('image')
    expect(fd.get('storage_type')).toBe('local')
    expect(fd.get('file')).toBeInstanceOf(File)
    expect(res.responseBody).toEqual({ id: 'F1', url: '/x' })
  })

  it('forwards progress events', async () => {
    const onProgress = vi.fn()
    const promise = httpUpload(
      fakeFile(),
      { endpoint: '/up', authHeader: '', fields: {} },
      { onProgress },
    )
    queueMicrotask(() => {
      last.__progress(50, 100)
      last.__progress(100, 100)
      last.__succeed(200, '{}')
    })
    await promise
    expect(onProgress).toHaveBeenNthCalledWith(1, 50, 100)
    expect(onProgress).toHaveBeenNthCalledWith(2, 100, 100)
  })

  it('rejects with UploadError on non-2xx and picks the text from the status', async () => {
    const promise = httpUpload(fakeFile(), {
      endpoint: '/up',
      authHeader: '',
      fields: {},
    })
    queueMicrotask(() =>
      last.__succeed(
        400,
        JSON.stringify({
          title: 'Bad Request',
          status: 400,
          detail: 'file too big',
        }),
      ),
    )
    await expect(promise).rejects.toMatchObject({
      name: 'UploadError',
      status: 400,
      message: problemMessage({ status: 400 }),
    })
  })

  it('rejects on network error', async () => {
    const promise = httpUpload(fakeFile(), {
      endpoint: '/up',
      authHeader: '',
      fields: {},
    })
    queueMicrotask(() => last.__fail())
    await expect(promise).rejects.toBeInstanceOf(UploadError)
  })
})

describe('httpUpload — target without auth', () => {
  it('posts the file as multipart to the given endpoint and omits Authorization', async () => {
    const file = fakeFile()
    const promise = httpUpload(file, {
      endpoint: 'https://bucket.example.com/api/files?sig=abc',
      authHeader: '',
      fields: { profile: 'attachment', name: 'a.png' },
    })
    queueMicrotask(() => last.__succeed(200, ''))
    const res = await promise
    expect(last.method).toBe('POST')
    expect(last.url).toBe('https://bucket.example.com/api/files?sig=abc')
    expect(last.headers['Content-Type']).toBeUndefined()
    expect(last.headers['Authorization']).toBeUndefined()
    expect(last.body).toBeInstanceOf(FormData)
    const fd = last.body as FormData
    expect(fd.get('file')).toBe(file)
    expect(fd.get('profile')).toBe('attachment')
    expect(fd.get('name')).toBe('a.png')
    expect(res.responseBody).toBeNull()
  })
})

describe('httpUpload — abort', () => {
  it('rejects with AbortError when signal fires', async () => {
    const ctl = new AbortController()
    const promise = httpUpload(
      fakeFile(),
      { endpoint: '/up', authHeader: '', fields: {} },
      { signal: ctl.signal },
    )
    queueMicrotask(() => ctl.abort())
    await expect(promise).rejects.toMatchObject({ name: 'AbortError' })
  })

  it('rejects immediately if signal is already aborted', async () => {
    const ctl = new AbortController()
    ctl.abort()
    await expect(
      httpUpload(
        fakeFile(),
        { kind: 'local', endpoint: '/up', authHeader: '', fields: {} },
        { signal: ctl.signal },
      ),
    ).rejects.toMatchObject({ name: 'AbortError' })
  })
})
