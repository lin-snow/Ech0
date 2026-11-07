// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const HOME_TIMELINE_ANCHOR_KEY = 'home:timeline:anchor'
export const HOME_TIMELINE_DEPTH_KEY = 'home:timeline:depth'

/** Identifies the first Echo visible at the top of the timeline viewport. */
export type TimelineAnchor = {
  id: string
  /** Pixels of the card already scrolled past the viewport top. */
  offset: number
  /** Filter signature the anchor was captured under; a mismatch invalidates it. */
  signature?: string
}

export function parseTimelineAnchor(raw: string | null): TimelineAnchor | null {
  if (!raw) return null
  try {
    const value: unknown = JSON.parse(raw)
    if (typeof value !== 'object' || value === null) return null
    const { id, offset, signature } = value as {
      id?: unknown
      offset?: unknown
      signature?: unknown
    }
    if (typeof id !== 'string' || id.trim() === '') return null
    if (typeof offset !== 'number' || !Number.isFinite(offset)) return null
    if (typeof signature !== 'undefined' && typeof signature !== 'string') return null
    return {
      id,
      offset: Math.max(0, Math.round(offset)),
      ...(typeof signature === 'string' ? { signature } : {}),
    }
  } catch {
    return null
  }
}

export function serializeTimelineAnchor(anchor: TimelineAnchor): string {
  return JSON.stringify({
    id: anchor.id,
    offset: Math.round(anchor.offset),
    ...(anchor.signature ? { signature: anchor.signature } : {}),
  })
}
