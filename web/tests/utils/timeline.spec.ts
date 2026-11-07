// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import {
  TIMELINE_BASE_PAGE_SIZE,
  TIMELINE_LOAD_AHEAD_PX,
  parseTimelineAnchor,
  serializeTimelineAnchor,
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
