// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { filterFiles, type FileSelectorOptions } from '@/lib/file/selectors/file-selectors'

type EchoLike = {
  id?: string
  layout?: string | null
  echo_files?: Array<{
    echo_id?: string
    file_id?: string
    file?: {
      id?: string
      name?: string
      url?: string
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
    return {
      id: String(file?.id || item.file_id || ''),
      echo_id: String(echo.id || item.echo_id || ''),
      url: String(file?.url || ''),
      name: file?.name,
      category: (file?.category as App.Api.File.Category | undefined) || undefined,
      content_type: file?.content_type,
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
