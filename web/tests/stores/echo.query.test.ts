// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useEchoStore } from '@/stores/echo'

vi.mock('@/service/api', () => ({
  fetchQueryEchos: vi.fn(),
  fetchGetTags: vi.fn(),
  fetchGetEchoById: vi.fn(),
  fetchCreateTag: vi.fn(),
}))

describe('echoStore.hasActiveQuery', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('is false on the unfiltered timeline', () => {
    expect(useEchoStore().hasActiveQuery).toBe(false)
  })

  it('tracks every narrowing input, not just tag filtering', () => {
    const store = useEchoStore()

    store.searchValue = 'missing-keyword'
    expect(store.hasActiveQuery).toBe(true)
    store.searchValue = ''
    expect(store.hasActiveQuery).toBe(false)

    store.isFilteringMode = true
    expect(store.hasActiveQuery).toBe(true)
    store.isFilteringMode = false

    store.dateFrom = Date.now()
    expect(store.hasActiveQuery).toBe(true)
    store.resetDateRange()

    store.selectedTagIds = ['tag-1']
    expect(store.hasActiveQuery).toBe(true)
    store.resetSelectedTags()

    store.visibilityFilter = 'private'
    expect(store.hasActiveQuery).toBe(true)
    store.resetVisibilityFilter()
    expect(store.hasActiveQuery).toBe(false)
  })
})
