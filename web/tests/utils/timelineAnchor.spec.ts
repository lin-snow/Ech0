// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import { parseTimelineAnchor, serializeTimelineAnchor } from '@/utils/timelineAnchor'

describe('timelineAnchor', () => {
  it('round-trips an anchor', () => {
    const anchor = { id: 'echo-1', offset: 42.6 }
    expect(parseTimelineAnchor(serializeTimelineAnchor(anchor))).toEqual({
      id: 'echo-1',
      offset: 43,
    })
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
