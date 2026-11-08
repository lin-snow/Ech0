// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import {
  TIMELINE_BASE_PAGE_SIZE,
  TIMELINE_LOAD_AHEAD_PX,
  TIMELINE_MAX_RESTORE_PAGE,
  TIMELINE_PAGE_ITEMS,
  parseTimelineAnchor,
  parseTimelineCursorChain,
  parseTimelinePage,
  resolveTimelineRestore,
  restoreTimelineChain,
  serializeTimelineAnchor,
  serializeTimelineCursorChain,
  timelineChunkSize,
  timelinePageForRows,
  timelinePageSize,
} from '@/utils/timeline'

describe('timelineAnchor', () => {
  it('round-trips an anchor', () => {
    const anchor = { id: 'echo-1', offset: 42.6 }
    expect(parseTimelineAnchor(serializeTimelineAnchor(anchor))).toEqual({
      id: 'echo-1',
      offset: 43,
    })
  })

  it('round-trips an anchor with a filter signature', () => {
    const anchor = { id: 'echo-1', offset: 0, signature: '["",[],null,null,"all"]' }
    expect(parseTimelineAnchor(serializeTimelineAnchor(anchor))).toEqual(anchor)
    expect(parseTimelineAnchor('{"id":"e1","offset":1,"signature":5}')).toBeNull()
  })

  it('rejects malformed payloads instead of restoring to a random place', () => {
    expect(parseTimelineAnchor(null)).toBeNull()
    expect(parseTimelineAnchor('')).toBeNull()
    expect(parseTimelineAnchor('not-json')).toBeNull()
    expect(parseTimelineAnchor('null')).toBeNull()
    expect(parseTimelineAnchor('[]')).toBeNull()
    expect(parseTimelineAnchor('{"id":"","offset":10}')).toBeNull()
    expect(parseTimelineAnchor('{"id":"e1","offset":"10"}')).toBeNull()
    expect(parseTimelineAnchor('{"id":"e1","offset":null}')).toBeNull()
  })
})

describe('timelinePageSize', () => {
  it('never drops below the baseline page size', () => {
    expect(timelinePageSize(0)).toBe(TIMELINE_BASE_PAGE_SIZE)
    expect(timelinePageSize(-200)).toBe(TIMELINE_BASE_PAGE_SIZE)
    expect(timelinePageSize(120)).toBe(TIMELINE_BASE_PAGE_SIZE)
  })

  it('grows with the viewport so one page covers it plus the prefetch band', () => {
    const short = timelinePageSize(700)
    const tall = timelinePageSize(2600)
    const huge = timelinePageSize(5000)

    expect(short).toBeGreaterThan(TIMELINE_BASE_PAGE_SIZE)
    expect(tall).toBeGreaterThan(short)
    expect(huge).toBeGreaterThan(tall)
    // The band is 600px; the added rows must account for at least that much.
    expect(tall - timelinePageSize(2600 - TIMELINE_LOAD_AHEAD_PX)).toBeGreaterThan(0)
  })

  it('stays inside the server page-size clamp', () => {
    expect(timelinePageSize(20000)).toBe(100)
  })
})

describe('timelineChunkSize', () => {
  it('caps a chunk at the page boundary', () => {
    expect(timelineChunkSize(40, 70)).toBe(30)
  })

  it('reports no room once the page is full', () => {
    expect(timelineChunkSize(7, TIMELINE_PAGE_ITEMS)).toBe(0)
  })

  it('passes an in-range request through unchanged', () => {
    expect(timelineChunkSize(7, 0)).toBe(7)
  })

  it('never requests less than one row while there is room', () => {
    expect(timelineChunkSize(0, 10)).toBe(1)
  })
})

describe('timelinePageForRows', () => {
  it('keeps the first page for depths inside it', () => {
    expect(timelinePageForRows(TIMELINE_BASE_PAGE_SIZE)).toBe(0)
    expect(timelinePageForRows(TIMELINE_PAGE_ITEMS)).toBe(0)
  })

  it('moves to the next page one row past the boundary', () => {
    expect(timelinePageForRows(TIMELINE_PAGE_ITEMS + 1)).toBe(1)
  })
})

describe('parseTimelinePage', () => {
  it('reopens the top for absent, malformed or fractional values', () => {
    expect(parseTimelinePage(null)).toBe(0)
    expect(parseTimelinePage('abc')).toBe(0)
    expect(parseTimelinePage('-1')).toBe(0)
    expect(parseTimelinePage('1.5')).toBe(0)
  })

  it('accepts depths up to the restore cap and rejects deeper ones', () => {
    expect(parseTimelinePage('3')).toBe(3)
    expect(parseTimelinePage(String(TIMELINE_MAX_RESTORE_PAGE))).toBe(TIMELINE_MAX_RESTORE_PAGE)
    expect(parseTimelinePage(String(TIMELINE_MAX_RESTORE_PAGE + 1))).toBe(0)
  })
})

describe('timelineCursorChain', () => {
  const chain = {
    signature: '["",[],null,null,"all"]',
    page: 2,
    starts: ['', 'c100', 'c200'],
  }

  it('round-trips a chain', () => {
    expect(parseTimelineCursorChain(serializeTimelineCursorChain(chain))).toEqual(chain)
  })

  it('only restores a chain captured under the same query', () => {
    const raw = serializeTimelineCursorChain(chain)
    expect(restoreTimelineChain(raw, chain.signature)).toEqual(chain)
    expect(restoreTimelineChain(raw, '["daily",[],null,null,"all"]')).toBeNull()
  })

  it('plans a one-request restore only for a matching chain', () => {
    const raw = serializeTimelineCursorChain(chain)
    expect(resolveTimelineRestore(raw, chain.signature, 3)).toEqual({
      mode: 'chain',
      page: 2,
      starts: chain.starts,
    })
    // A chain from another filter set must not fall back to a legacy page
    // number captured without those filters.
    expect(resolveTimelineRestore(raw, '["daily",[],null,null,"all"]', 3)).toEqual({ mode: 'top' })
  })

  it('refreshes the top for a saved first page and falls back to a legacy page index', () => {
    const firstPage = serializeTimelineCursorChain({ ...chain, page: 0, starts: [''] })
    expect(resolveTimelineRestore(firstPage, chain.signature, 3)).toEqual({ mode: 'top' })
    expect(resolveTimelineRestore(null, chain.signature, 3)).toEqual({ mode: 'page', page: 3 })
    expect(resolveTimelineRestore(null, chain.signature, 0)).toEqual({ mode: 'top' })
    expect(resolveTimelineRestore('not-json', chain.signature, 2)).toEqual({ mode: 'top' })
  })

  it('rejects malformed or out-of-range chains', () => {
    expect(parseTimelineCursorChain(null)).toBeNull()
    expect(parseTimelineCursorChain('')).toBeNull()
    expect(parseTimelineCursorChain('not-json')).toBeNull()
    expect(parseTimelineCursorChain('null')).toBeNull()
    expect(parseTimelineCursorChain('[]')).toBeNull()
    expect(
      parseTimelineCursorChain(JSON.stringify({ signature: '', page: -1, starts: [''] })),
    ).toBeNull()
    expect(
      parseTimelineCursorChain(
        JSON.stringify({ signature: '', page: TIMELINE_MAX_RESTORE_PAGE + 1, starts: [''] }),
      ),
    ).toBeNull()
    // The chain must already contain the cursor that opens the saved page.
    expect(
      parseTimelineCursorChain(JSON.stringify({ signature: '', page: 1, starts: [''] })),
    ).toBeNull()
    expect(
      parseTimelineCursorChain(JSON.stringify({ signature: '', page: 0, starts: ['x'] })),
    ).toBeNull()
    expect(
      parseTimelineCursorChain(JSON.stringify({ signature: '', page: 0, starts: ['', ''] })),
    ).toBeNull()
    expect(
      parseTimelineCursorChain(JSON.stringify({ signature: '', page: 0, starts: ['', 5] })),
    ).toBeNull()
    expect(
      parseTimelineCursorChain(
        JSON.stringify({ signature: '', page: 0, starts: ['', 'x'.repeat(513)] }),
      ),
    ).toBeNull()
  })
})
