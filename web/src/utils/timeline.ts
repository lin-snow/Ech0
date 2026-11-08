// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const HOME_TIMELINE_ANCHOR_KEY = 'home:timeline:anchor'
export const HOME_TIMELINE_PAGE_KEY = 'home:timeline:page'
export const HOME_TIMELINE_CURSORS_KEY = 'home:timeline:cursors'

/** Band below the timeline viewport where the next chunk starts prefetching. */
export const TIMELINE_LOAD_AHEAD_PX = 600

/**
 * Echos per page. A page fills up in chunks through infinite scroll; once full,
 * the reader steps to the next page instead of appending without bound, so the
 * rendered DOM stays small without virtualization.
 */
export const TIMELINE_PAGE_ITEMS = 100

/**
 * Deepest page a reload restores by walking the cursor chain. Deeper positions
 * reopen the top of the feed.
 */
export const TIMELINE_MAX_RESTORE_PAGE = 10

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

/**
 * Rows the next chunk of a page may request: the requested size, capped so the
 * chunk never crosses the page boundary. Zero means the page is already full.
 */
export function timelineChunkSize(requested: number, loadedInPage: number): number {
  const remaining = Math.max(0, TIMELINE_PAGE_ITEMS - loadedInPage)
  return Math.min(Math.max(1, Math.trunc(requested) || 1), remaining)
}

/** Page index that holds the row at `rows` (1-based count) of a feed. */
export function timelinePageForRows(rows: number): number {
  return Math.max(0, Math.ceil(rows / TIMELINE_PAGE_ITEMS) - 1)
}

/** Parses the stored page index; unusable values and depths past the restore cap reopen the top. */
export function parseTimelinePage(raw: string | null): number {
  const value = Number(raw)
  if (!Number.isInteger(value) || value < 1 || value > TIMELINE_MAX_RESTORE_PAGE) return 0
  return value
}

/** Cursor tokens are opaque; this only bounds what the backend already caps. */
const CURSOR_TOKEN_MAX_LENGTH = 512

/**
 * A page's opening cursors, captured under the filter signature that produced
 * them. Restoring the chain skips walking every page before the saved one;
 * the signature keeps a filtered chain from seeding another query.
 */
export type TimelineCursorChain = {
  signature: string
  page: number
  starts: string[]
}

export function parseTimelineCursorChain(raw: string | null): TimelineCursorChain | null {
  if (!raw) return null
  try {
    const value: unknown = JSON.parse(raw)
    if (typeof value !== 'object' || value === null) return null
    const { signature, page, starts } = value as {
      signature?: unknown
      page?: unknown
      starts?: unknown
    }
    if (typeof signature !== 'string') return null
    if (
      typeof page !== 'number' ||
      !Number.isInteger(page) ||
      page < 0 ||
      page > TIMELINE_MAX_RESTORE_PAGE
    ) {
      return null
    }
    if (!Array.isArray(starts) || starts.length <= page) return null
    if (starts.length > TIMELINE_MAX_RESTORE_PAGE + 1) return null
    const chain: string[] = []
    for (let index = 0; index < starts.length; index += 1) {
      const start = starts[index]
      if (typeof start !== 'string') return null
      if (index === 0 ? start !== '' : start === '') return null
      if (start.length > CURSOR_TOKEN_MAX_LENGTH) return null
      chain.push(start)
    }
    return { signature, page, starts: chain }
  } catch {
    return null
  }
}

export function serializeTimelineCursorChain(chain: TimelineCursorChain): string {
  return JSON.stringify({
    signature: chain.signature,
    page: chain.page,
    starts: chain.starts,
  })
}

/** Returns the stored chain only when it belongs to the query being opened. */
export function restoreTimelineChain(
  raw: string | null,
  signature: string,
): TimelineCursorChain | null {
  const chain = parseTimelineCursorChain(raw)
  return chain && chain.signature === signature ? chain : null
}

/** How a mount should reopen the feed: a saved chain, a legacy page, or the top. */
export type TimelineRestorePlan =
  | { mode: 'chain'; page: number; starts: string[] }
  | { mode: 'page'; page: number }
  | { mode: 'top' }

export function resolveTimelineRestore(
  rawChain: string | null,
  signature: string,
  legacyPage: number,
): TimelineRestorePlan {
  const chain = restoreTimelineChain(rawChain, signature)
  if (chain) {
    // Page 0 needs no chain: the normal refresh already opens the top.
    return chain.page > 0
      ? { mode: 'chain', page: chain.page, starts: chain.starts }
      : { mode: 'top' }
  }
  // A chain was stored for another query; falling back to the legacy page
  // would reopen that page without the filters it belonged to.
  if (rawChain) return { mode: 'top' }
  if (legacyPage > 0) return { mode: 'page', page: legacyPage }
  return { mode: 'top' }
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
