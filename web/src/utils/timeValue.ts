// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export function timeValueToMs(raw: unknown): number {
  if (raw == null) return 0
  if (typeof raw === 'number') {
    if (!Number.isFinite(raw)) return 0
    return raw < 1e12 ? raw * 1000 : raw
  }
  if (typeof raw === 'string') {
    const s = raw.trim()
    if (!s) return 0
    if (/^-?\d+(\.\d+)?$/.test(s)) {
      const n = Number(s)
      if (!Number.isFinite(n)) return 0
      return n < 1e12 ? n * 1000 : n
    }
    const ms = Date.parse(s)
    return Number.isNaN(ms) ? 0 : ms
  }
  return 0
}

export function timeValueToUnixSeconds(raw: unknown): number {
  const ms = timeValueToMs(raw)
  if (ms <= 0) return 0
  return Math.floor(ms / 1000)
}
