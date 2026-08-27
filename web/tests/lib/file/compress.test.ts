// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { compressImage, inferFileExtFromType } from '../../../src/lib/file/compress'

vi.mock('@/utils/other', () => ({
  isSafari: vi.fn(() => false),
}))

let mockToBlobOutput: { mime: string; size: number } = { mime: 'image/webp', size: 100 }

beforeEach(() => {
  vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob://fake')
  vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})

  vi.stubGlobal(
    'Image',
    class {
      naturalWidth = 100
      naturalHeight = 100
      width = 100
      height = 100
      onload: (() => void) | null = null
      onerror: (() => void) | null = null
      set src(_v: string) {
        queueMicrotask(() => this.onload?.())
      }
    },
  )

  HTMLCanvasElement.prototype.getContext = vi.fn(
    () => ({ drawImage: vi.fn() }) as unknown as CanvasRenderingContext2D,
  )
  HTMLCanvasElement.prototype.toBlob = vi.fn(function (
    this: HTMLCanvasElement,
    cb: BlobCallback,
    mime?: string,
  ) {
    const t = mime || mockToBlobOutput.mime
    cb(new Blob([new Uint8Array(mockToBlobOutput.size)], { type: t }))
  })
})

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

function fakeFile(name: string, type: string, size: number, lastModified = 1700000000000): File {
  const bytes = new Uint8Array(size)
  return new File([bytes], name, { type, lastModified })
}

describe('inferFileExtFromType', () => {
  it('maps known mimes to extensions', () => {
    expect(inferFileExtFromType('image/png')).toBe('.png')
    expect(inferFileExtFromType('image/webp')).toBe('.webp')
    expect(inferFileExtFromType('image/jpeg')).toBe('.jpg')
    expect(inferFileExtFromType('image/jpg')).toBe('.jpg')
    expect(inferFileExtFromType('image/gif')).toBe('.gif')
    expect(inferFileExtFromType('image/avif')).toBe('.avif')
    expect(inferFileExtFromType('image/bmp')).toBe('.bmp')
  })
  it('falls back to .bin for unknown', () => {
    expect(inferFileExtFromType('application/octet-stream')).toBe('.bin')
    expect(inferFileExtFromType('')).toBe('.bin')
  })
})

describe('compressImage', () => {
  it('returns the original file for non-convertible types (e.g. gif)', async () => {
    const f = fakeFile('a.gif', 'image/gif', 1000)
    const out = await compressImage(f)
    expect(out).toBe(f)
  })

  it('keeps original mime when below convertSize, only re-encodes for quality', async () => {
    mockToBlobOutput = { mime: 'image/png', size: 500 }
    const f = fakeFile('photo.png', 'image/png', 1024)
    const out = await compressImage(f, { convertSize: 5 * 1024 * 1024 })
    expect(out.type).toBe('image/png')
    expect(out.name).toBe('photo.png')
    expect(out).not.toBe(f)
    expect(HTMLCanvasElement.prototype.toBlob).toHaveBeenCalled()
    const callArgs = (HTMLCanvasElement.prototype.toBlob as unknown as { mock: { calls: unknown[][] } })
      .mock.calls[0]
    expect(callArgs[1]).toBe('image/png')
  })

  it('converts to webp on non-Safari when over threshold and rewrites filename ext', async () => {
    mockToBlobOutput = { mime: 'image/webp', size: 500 }
    const f = fakeFile('photo.png', 'image/png', 6 * 1024 * 1024)
    const out = await compressImage(f, { convertSize: 5 * 1024 * 1024 })
    expect(out.type).toBe('image/webp')
    expect(out.name).toBe('photo.webp')
    expect(out.lastModified).toBe(f.lastModified)
  })

  it('converts to jpeg when isSafari is true', async () => {
    mockToBlobOutput = { mime: 'image/jpeg', size: 500 }
    const other = await import('@/utils/other')
    vi.mocked(other.isSafari).mockReturnValueOnce(true)
    const f = fakeFile('photo.png', 'image/png', 6 * 1024 * 1024)
    const out = await compressImage(f, { convertSize: 5 * 1024 * 1024 })
    expect(out.type).toBe('image/jpeg')
    expect(out.name).toBe('photo.jpg')
  })

  it('returns the original when same-mime re-encode produces a larger blob', async () => {
    mockToBlobOutput = { mime: 'image/png', size: 9999 }
    const f = fakeFile('photo.png', 'image/png', 1024)
    const out = await compressImage(f, { convertSize: 5 * 1024 * 1024 })
    expect(out).toBe(f)
  })

  it('still returns the converted file when format-converting (webp), even if not strictly smaller', async () => {
    mockToBlobOutput = { mime: 'image/webp', size: 9_999_999 }
    const f = fakeFile('big.png', 'image/png', 6 * 1024 * 1024)
    const out = await compressImage(f, { convertSize: 5 * 1024 * 1024 })
    expect(out.type).toBe('image/webp')
    expect(out.name).toBe('big.webp')
  })
})
