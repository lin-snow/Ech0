// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const SIZE_UNITS = ['B', 'KB', 'MB', 'GB'] as const

export function formatBytes(bytes: number | undefined | null): string {
  if (bytes == null || !Number.isFinite(bytes) || bytes < 0) return '—'
  if (bytes === 0) return '0 B'
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < SIZE_UNITS.length - 1) {
    value /= 1024
    unit++
  }
  const fixed = value >= 100 || unit === 0 ? value.toFixed(0) : value.toFixed(1)
  return `${fixed} ${SIZE_UNITS[unit]}`
}
