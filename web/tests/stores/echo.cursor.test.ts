// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useEchoStore } from '@/stores/echo'
import { fetchQueryEchos } from '@/service/api'

vi.mock('@/service/api', () => ({
  fetchQueryEchos: vi.fn(),
  fetchGetTags: vi.fn(),
  fetchGetEchoById: vi.fn(),
  fetchCreateTag: vi.fn(),
}))

const queryMock = vi.mocked(fetchQueryEchos)

const echo = (id: string): App.Api.Ech0.Echo =>
  ({ id, content: id, created_at: 1 }) as App.Api.Ech0.Echo

type EchoPageResponse = App.Api.Response<App.Api.Ech0.PaginationResult>

const page = (ids: string[], total: number, nextCursor: string | null): EchoPageResponse => ({
  code: 1,
  msg: '',
  data: { items: ids.map(echo), total, nextCursor },
})

describe('echoStore cursor stream', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    queryMock.mockReset()
  })

  it('starts from the first cursor page and keeps the next position', async () => {
    queryMock.mockResolvedValueOnce(page(['a', 'b'], 2, 'cursor-1'))
    const store = useEchoStore()

    await store.refreshEchos()

    expect(queryMock).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 7, cursor: '' }),
    )
    expect(store.echoList.map((item) => item.id)).toEqual(['a', 'b'])
    expect(store.nextCursor).toBe('cursor-1')
    expect(store.canLoadMore).toBe(true)
    expect(store.isTimelineCurrent).toBe(true)
  })

  it('appends the next page, drops duplicates, and ends on a null cursor', async () => {
    queryMock.mockResolvedValueOnce(page(['a', 'b'], 3, 'cursor-1'))
    const store = useEchoStore()
    await store.refreshEchos()

    queryMock.mockResolvedValueOnce(page(['b', 'c'], 3, null))
    await store.loadMore()

    expect(queryMock).toHaveBeenLastCalledWith(expect.objectContaining({ cursor: 'cursor-1' }))
    expect(store.echoList.map((item) => item.id)).toEqual(['a', 'b', 'c'])
    expect(store.echoIndexMap.get('c')).toBe(2)
    expect(store.nextCursor).toBeNull()
    expect(store.canLoadMore).toBe(false)
  })

  it('discards a response superseded by a newer query', async () => {
    const staleGate = Promise.withResolvers<EchoPageResponse>()
    queryMock.mockImplementationOnce(() => staleGate.promise)
    const store = useEchoStore()
    const stale = store.refreshEchos()

    queryMock.mockResolvedValueOnce(page(['fresh'], 1, null))
    await store.refreshEchos()

    staleGate.resolve(page(['stale'], 1, null))
    await stale

    expect(store.echoList.map((item) => item.id)).toEqual(['fresh'])
  })

  it('marks the loaded timeline stale as soon as the filter set changes', async () => {
    queryMock.mockResolvedValueOnce(page(['a'], 1, null))
    const store = useEchoStore()
    await store.refreshEchos()
    expect(store.isTimelineCurrent).toBe(true)

    store.searchValue = 'golang'
    expect(store.isTimelineCurrent).toBe(false)
  })

  it('retries a failed first page', async () => {
    queryMock.mockResolvedValueOnce({ code: 0, msg: 'boom', data: null })
    const store = useEchoStore()
    await store.refreshEchos()
    expect(store.loadError).toBe('first')

    queryMock.mockResolvedValueOnce(page(['a'], 1, null))
    await store.retryLoad()

    expect(store.loadError).toBeNull()
    expect(store.echoList.map((item) => item.id)).toEqual(['a'])
  })

  it('retries a failed deep first page at the depth it was opened with', async () => {
    queryMock.mockResolvedValueOnce({ code: 0, msg: 'boom', data: null })
    const store = useEchoStore()
    await store.refreshEchos(21)
    expect(store.loadError).toBe('first')

    queryMock.mockResolvedValueOnce(page(['a'], 21, null))
    await store.retryLoad()

    expect(queryMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1, pageSize: 21, cursor: '' }),
    )
    expect(store.loadError).toBeNull()
  })

  it('removes a deleted echo locally without resetting the stream', async () => {
    queryMock.mockResolvedValueOnce(page(['a', 'b'], 2, 'cursor-1'))
    const store = useEchoStore()
    await store.refreshEchos()

    store.removeEcho('a')

    expect(store.echoList.map((item) => item.id)).toEqual(['b'])
    expect(store.total).toBe(1)
    expect(store.echoIndexMap.get('b')).toBe(0)
    expect(store.nextCursor).toBe('cursor-1')
  })
})
