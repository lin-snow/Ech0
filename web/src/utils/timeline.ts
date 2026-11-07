// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const HOME_TIMELINE_ANCHOR_KEY = 'home:timeline:anchor'
export const HOME_TIMELINE_COUNT_KEY = 'home:timeline:count'

/** Band below the timeline viewport where the next page starts prefetching. */
export const TIMELINE_LOAD_AHEAD_PX = 600

/** Placeholder row height the virtualizer assumes before a card is measured. */
export const TIMELINE_ROW_ESTIMATE_PX = 240

/**
 * Conservative row height for sizing a request. Text-only cards render around
 * 100-130px, so assuming less than that would under-fill a viewport.
 */
const TIMELINE_FILL_ROW_PX = 160

/** Baseline rows per request; also the unit of legacy `?page=N` links. */
export const TIMELINE_BASE_PAGE_SIZE = 7

/** Page-size bounds; the server clamps to the same range. */
const TIMELINE_PAGE_MIN = TIMELINE_BASE_PAGE_SIZE
const TIMELINE_PAGE_MAX = 100

/**
 * Rows to request so one response covers the viewport plus the prefetch band.
 *
 * A page fixed at a handful of rows keeps the sentinel inside the band on a
 * tall screen: the feed then chases page after page nonstop, toggling the
 * loading indicator until the content finally outgrows the viewport. Sizing
 * the page from the viewport fills the screen in a single request, so the
 * automatic chain never starts.
 */
export function timelinePageSize(viewportPx: number): number {
  const target = Math.max(0, viewportPx) + TIMELINE_LOAD_AHEAD_PX
  const rows = Math.ceil(target / TIMELINE_FILL_ROW_PX) + 2
  return Math.min(TIMELINE_PAGE_MAX, Math.max(TIMELINE_PAGE_MIN, rows))
}

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
