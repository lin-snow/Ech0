// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

type DedupeKey = 'id' | 'url'

export type FileSelectorOptions<T extends { category?: unknown }> = {
  categories?: App.Api.File.Category[]
  predicate?: (file: T, index: number, files: T[]) => boolean
  dedupeBy?: DedupeKey
}

function normalize(value: unknown): string {
  return String(value || '').toLowerCase()
}

function passCategory(category: unknown, categories?: App.Api.File.Category[]): boolean {
  if (!categories || categories.length === 0) return true
  const allowed = new Set(categories.map((item) => normalize(item)))
  return allowed.has(normalize(category))
}

function dedupeFiles<T>(files: T[], dedupeBy?: DedupeKey): T[] {
  if (!dedupeBy) return files
  const seen = new Set<string>()
  const result: T[] = []
  for (const file of files as Array<T & Record<string, unknown>>) {
    const key = String(file?.[dedupeBy] || '')
    if (!key) {
      result.push(file)
      continue
    }
    if (seen.has(key)) continue
    seen.add(key)
    result.push(file)
  }
  return result
}

export function filterFiles<T extends { category?: unknown }>(
  files: T[] = [],
  options: FileSelectorOptions<T> = {},
): T[] {
  const filtered = files.filter((file, index, arr) => {
    if (!passCategory(file.category, options.categories)) return false
    if (options.predicate && !options.predicate(file, index, arr)) return false
    return true
  })
  return dedupeFiles(filtered, options.dedupeBy)
}

export const selectFiles = filterFiles
