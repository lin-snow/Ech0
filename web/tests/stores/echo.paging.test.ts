// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useEchoStore } from '@/stores/echo'
import { fetchQueryEchos } from '@/service/api'
import { TIMELINE_PAGE_ITEMS } from '@/utils/timeline'

vi.mock('@/service/api', () => ({
  fetchQueryEchos: vi.fn(),
  fetchGetTags: vi.fn(),
  fetchGetEchoById: vi.fn(),
  fetchCreateTag: vi.fn(),
}))

const queryMock = vi.mocked(fetchQueryEchos)

type EchoPageResponse = App.Api.Response<App.Api.Ech0.PaginationResult>

const echo = (id: string): App.Api.Ech0.Echo =>
  ({ id, content: id, created_at: 1 }) as App.Api.Ech0.Echo

const page = (ids: string[], total: number, nextCursor: string | null): EchoPageResponse => ({
  code: 1,
  msg: '',
  data: { items: ids.map(echo), total, nextCursor },
})

const failure: EchoPageResponse = {
  code: 0,
  msg: 'boom',
  data: null,
} as unknown as EchoPageResponse

const range = (prefix: string, count: number): string[] =>
  Array.from({ length: count }, (_, i) => `${prefix}${i}`)

const idsOf = (items: App.Api.Ech0.Echo[]): string[] => items.map((item) => item.id)

describe('echoStore paging', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    queryMock.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('stops auto-loading at the page boundary and offers the next page', async () => {
    queryMock.mockResolvedValueOnce(page(range('a', 60), 500, 'c60'))
    const store = useEchoStore()
    await store.refreshEchos(60)

    store.pageSize = 60
    queryMock.mockResolvedValueOnce(page(range('b', 40), 500, 'c100'))
    await store.loadMore()

    // The chunk is capped at the room the page has left.
    expect(queryMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ pageSize: 40, cursor: 'c60' }),
    )
    expect(store.echoList).toHaveLength(TIMELINE_PAGE_ITEMS)
    expect(store.canLoadMore).toBe(false)
    expect(store.hasNextPage).toBe(true)
  })

  it('opens the next page at the boundary cursor and steps back to the stored start', async () => {
    queryMock.mockResolvedValueOnce(page(range('a', 100), 500, 'c100'))
    const store = useEchoStore()
    await store.refreshEchos(100)

    queryMock.mockResolvedValueOnce(page(range('b', 3), 500, null))
    await store.nextPage()

    expect(queryMock).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 'c100' }))
    expect(store.pageIndex).toBe(1)
    expect(idsOf(store.echoList)).toEqual(['b0', 'b1', 'b2'])
    expect(store.hasPrevPage).toBe(true)

    queryMock.mockResolvedValueOnce(page(range('a', 7), 500, 'a7'))
    await store.prevPage()

    expect(queryMock).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: '' }))
    expect(store.pageIndex).toBe(0)
    expect(store.echoList[0]?.id).toBe('a0')
  })

  it('walks the cursor chain to a deep page, one full page per request', async () => {
    queryMock
      .mockResolvedValueOnce(page(range('p0-', 100), 300, 'c100'))
      .mockResolvedValueOnce(page(range('p1-', 100), 300, 'c200'))
      .mockResolvedValueOnce(page(range('p2-', 100), 300, 'c300'))
    const store = useEchoStore()

    await store.openAt(2, TIMELINE_PAGE_ITEMS)

    expect(queryMock.mock.calls.map(([params]) => [params.cursor, params.pageSize])).toEqual([
      ['', TIMELINE_PAGE_ITEMS],
      ['c100', TIMELINE_PAGE_ITEMS],
      ['c200', TIMELINE_PAGE_ITEMS],
    ])
    expect(store.pageIndex).toBe(2)
    expect(store.echoList[0]?.id).toBe('p2-0')
    expect(store.nextCursor).toBe('c300')
  })

  it('reopens a restored page from its stored cursor chain in one request', async () => {
    queryMock.mockResolvedValueOnce(page(range('p2-', 100), 300, 'c300'))
    const store = useEchoStore()

    await store.openFromChain(2, ['', 'c100', 'c200'], TIMELINE_PAGE_ITEMS)

    expect(queryMock).toHaveBeenCalledTimes(1)
    expect(queryMock.mock.calls.map(([params]) => [params.cursor, params.pageSize])).toEqual([
      ['c200', TIMELINE_PAGE_ITEMS],
    ])
    expect(store.pageIndex).toBe(2)
    expect(store.hasNextPage).toBe(true)
  })

  it('walks from the last stored cursor when the chain is short', async () => {
    queryMock
      .mockResolvedValueOnce(page(range('p1-', 100), 300, 'c200'))
      .mockResolvedValueOnce(page(range('p2-', 100), 300, 'c300'))
    const store = useEchoStore()

    await store.openFromChain(2, ['', 'c100'], TIMELINE_PAGE_ITEMS)

    expect(queryMock.mock.calls.map(([params]) => [params.cursor, params.pageSize])).toEqual([
      ['c100', TIMELINE_PAGE_ITEMS],
      ['c200', TIMELINE_PAGE_ITEMS],
    ])
    expect(store.pageIndex).toBe(2)
  })

  it('lands on the last reachable page when the walk runs out of feed', async () => {
    queryMock
      .mockResolvedValueOnce(page(range('p0-', 3), 3, null))
      .mockResolvedValueOnce(page(range('p0-', 3), 3, null))
    const store = useEchoStore()

    await store.openAt(5, TIMELINE_PAGE_ITEMS)

    expect(store.pageIndex).toBe(0)
    expect(store.echoList).toHaveLength(3)
    expect(store.loadError).toBeNull()
  })

  it('retries a failed load-more on backoff before surfacing the error', async () => {
    vi.useFakeTimers()
    queryMock.mockResolvedValueOnce(page(range('a', 2), 9, 'c2'))
    const store = useEchoStore()
    await store.refreshEchos(2)

    queryMock
      .mockResolvedValueOnce(failure)
      .mockResolvedValueOnce(failure)
      .mockResolvedValueOnce(failure)
    await store.loadMore()

    // The loading state is held through the backoff, so the sentinel cannot
    // start another request meanwhile.
    expect(store.loadError).toBeNull()
    expect(store.isLoadingMore).toBe(true)
    expect(store.canLoadMore).toBe(false)

    await vi.advanceTimersByTimeAsync(1000)
    expect(queryMock).toHaveBeenCalledTimes(3)
    expect(store.loadError).toBeNull()

    await vi.advanceTimersByTimeAsync(3000)
    expect(queryMock).toHaveBeenCalledTimes(4)
    expect(store.loadError).toBe('more')
    expect(store.isLoadingMore).toBe(false)
  })

  it('keeps appending when an automatic retry succeeds', async () => {
    vi.useFakeTimers()
    queryMock.mockResolvedValueOnce(page(range('a', 2), 9, 'c2'))
    const store = useEchoStore()
    await store.refreshEchos(2)

    queryMock.mockResolvedValueOnce(failure).mockResolvedValueOnce(page(['b0', 'b1'], 9, 'c4'))
    await store.loadMore()
    await vi.advanceTimersByTimeAsync(1000)

    expect(store.loadError).toBeNull()
    expect(idsOf(store.echoList)).toEqual(['a0', 'a1', 'b0', 'b1'])
    expect(store.nextCursor).toBe('c4')
  })

  it('retries a failed page switch at the page it was switching to', async () => {
    queryMock.mockResolvedValueOnce(page(range('a', TIMELINE_PAGE_ITEMS), 300, 'c100'))
    const store = useEchoStore()
    await store.refreshEchos(TIMELINE_PAGE_ITEMS)

    queryMock.mockResolvedValueOnce(failure)
    await store.nextPage()
    expect(store.loadError).toBe('first')
    expect(store.pageIndex).toBe(0)

    queryMock.mockResolvedValueOnce(page(range('b', 3), 300, null))
    await store.retryLoad()

    expect(queryMock).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 'c100' }))
    expect(store.pageIndex).toBe(1)
    expect(store.loadError).toBeNull()
  })
})
