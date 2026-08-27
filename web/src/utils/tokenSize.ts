// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const K = 1000
const M = 1000 * 1000

export function parseTokenSize(input: string): number {
  const s = (input ?? '').trim().toLowerCase()
  if (s === '') return 0
  const match = s.match(/^(\d+(?:\.\d+)?)\s*([km])?$/)
  if (!match) return 0
  const value = Number.parseFloat(match[1])
  if (!Number.isFinite(value)) return 0
  const unit = match[2]
  const tokens = unit === 'm' ? value * M : unit === 'k' ? value * K : value
  return Math.max(0, Math.round(tokens))
}

export function formatTokenSize(tokens: number): string {
  if (!tokens || tokens <= 0) return ''
  if (tokens % M === 0) return `${tokens / M}m`
  if (tokens % K === 0) return `${tokens / K}k`
  return String(tokens)
}
