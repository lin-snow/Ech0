// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { isSafari } from '@/utils/other'

const DEFAULT_CONVERT_SIZE = 5 * 1024 * 1024
const DEFAULT_QUALITY = 0.6
const DEFAULT_CONVERT_TYPES = ['image/jpeg', 'image/png', 'image/webp']

export interface CompressOptions {
  quality?: number
  convertSize?: number
  outputMimeType?: string
  convertTypes?: string[]
}

export function inferFileExtFromType(contentType: string): string {
  const normalized = String(contentType || '').toLowerCase()
  if (normalized.includes('png')) return '.png'
  if (normalized.includes('webp')) return '.webp'
  if (normalized.includes('gif')) return '.gif'
  if (normalized.includes('bmp')) return '.bmp'
  if (normalized.includes('avif')) return '.avif'
  if (normalized.includes('jpeg') || normalized.includes('jpg')) return '.jpg'
  return '.bin'
}

function stripExt(name: string): string {
  const dot = name.lastIndexOf('.')
  return dot >= 0 ? name.slice(0, dot) : name
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('image load failed'))
    img.src = src
  })
}

async function reencodeViaCanvas(file: File, mime: string, quality: number): Promise<Blob> {
  const url = URL.createObjectURL(file)
  try {
    const img = await loadImage(url)
    const canvas = document.createElement('canvas')
    canvas.width = img.naturalWidth || img.width
    canvas.height = img.naturalHeight || img.height
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('canvas 2d context unavailable')
    ctx.drawImage(img, 0, 0)
    return await new Promise<Blob>((resolve, reject) => {
      canvas.toBlob(
        (blob) => (blob ? resolve(blob) : reject(new Error('canvas.toBlob returned null'))),
        mime,
        quality,
      )
    })
  } finally {
    URL.revokeObjectURL(url)
  }
}

export async function compressImage(file: File, opts: CompressOptions = {}): Promise<File> {
  const quality = opts.quality ?? DEFAULT_QUALITY
  const convertSize = opts.convertSize ?? DEFAULT_CONVERT_SIZE
  const convertTypes = opts.convertTypes ?? DEFAULT_CONVERT_TYPES
  const outputMimeType = opts.outputMimeType ?? (isSafari() ? 'image/jpeg' : 'image/webp')

  if (!file.type || !convertTypes.includes(file.type)) {
    return file
  }

  const targetMime = file.size >= convertSize ? outputMimeType : file.type
  const blob = await reencodeViaCanvas(file, targetMime, quality)

  if (blob.size >= file.size && targetMime === file.type) {
    return file
  }

  const baseName = stripExt(file.name)
  const newName = baseName + inferFileExtFromType(targetMime)
  return new File([blob], newName, {
    type: targetMime,
    lastModified: file.lastModified,
  })
}
