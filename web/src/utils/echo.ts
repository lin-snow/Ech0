// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { FILE_STORAGE_TYPE } from '@/constants/file'
import { filterFiles, type FileSelectorOptions } from '@/lib/file/selectors/file-selectors'

type EchoLike = {
  id?: string
  layout?: string | null
  echo_files?: Array<{
    echo_id?: string
    file_id?: string
    file?: {
      id?: string
      key?: string
      url?: string
      storage_type?: string
      category?: string
      content_type?: string
      size?: number
      width?: number
      height?: number
    }
  }> | null
}

export function getEchoFiles(echo?: EchoLike | null): App.Api.Ech0.FileObject[] {
  if (!echo) return []

  return (echo.echo_files || []).map((item) => {
    const file = item.file
    const storageType = normalizeStorageType(file?.storage_type)
    return {
      id: String(file?.id || item.file_id || ''),
      echo_id: String(echo.id || item.echo_id || ''),
      url: String(file?.url || ''),
      storage_type: storageType,
      category: (file?.category as App.Api.File.Category | undefined) || undefined,
      content_type: file?.content_type,
      key: String(file?.key || ''),
      size: file?.size,
      width: file?.width,
      height: file?.height,
    }
  })
}

export function getEchoFilesBy(
  echo?: EchoLike | null,
  options: FileSelectorOptions<App.Api.Ech0.FileObject> = {},
): App.Api.Ech0.FileObject[] {
  return filterFiles(getEchoFiles(echo), options)
}

export const getEchoImages = (echo?: EchoLike | null) =>
  getEchoFilesBy(echo, { categories: ['image'] })

const CONTENT_LEADING_LAYOUTS = ['grid', 'horizontal', 'stack']

export function isContentLeadingEcho(echo?: EchoLike | null): boolean {
  const layout = echo?.layout
  return !!layout && CONTENT_LEADING_LAYOUTS.includes(layout)
}

function normalizeStorageType(raw: unknown): App.Api.File.StorageType {
  const value = String(raw || '').toLowerCase()
  if (value === FILE_STORAGE_TYPE.OBJECT) return FILE_STORAGE_TYPE.OBJECT
  if (value === FILE_STORAGE_TYPE.EXTERNAL) return FILE_STORAGE_TYPE.EXTERNAL
  return FILE_STORAGE_TYPE.LOCAL
}

export function countWords(content: string | null | undefined): number {
  if (!content) return 0
  const stripped = content
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`[^`]*`/g, ' ')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[#>*_~`[\]()!\-]+/g, ' ')
    .trim()
  if (!stripped) return 0
  const cjk = stripped.match(/[一-鿿぀-ヿ가-힯]/g)?.length ?? 0
  const ascii = stripped
    .replace(/[一-鿿぀-ヿ가-힯]/g, ' ')
    .split(/\s+/)
    .filter(Boolean).length
  return cjk + ascii
}
