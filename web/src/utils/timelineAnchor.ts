// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** Identifies the first Echo visible at the top of the timeline viewport. */
export type TimelineAnchor = {
  id: string
  /** Pixels of the card already scrolled past the viewport top. */
  offset: number
}

export function parseTimelineAnchor(raw: string | null): TimelineAnchor | null {
  if (!raw) return null
  try {
    const value: unknown = JSON.parse(raw)
    if (typeof value !== 'object' || value === null) return null
    const { id, offset } = value as { id?: unknown; offset?: unknown }
    if (typeof id !== 'string' || id.trim() === '') return null
    if (typeof offset !== 'number' || !Number.isFinite(offset)) return null
    return { id, offset: Math.round(offset) }
  } catch {
    return null
  }
}

export function serializeTimelineAnchor(anchor: TimelineAnchor): string {
  return JSON.stringify({ id: anchor.id, offset: Math.round(anchor.offset) })
}
