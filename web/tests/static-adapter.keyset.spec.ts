// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeAll, describe, expect, it, vi } from 'vitest'

import type * as StaticAdapterModule from '@/service/request/static-adapter'
import type { StaticDataset } from '@/service/request/static-adapter'

type StaticAdapter = typeof StaticAdapterModule

const SECOND = 1_770_000_000

const echo = (id: string, createdAt: number): App.Api.Ech0.Echo => ({
  id,
  content: `content ${id}`,
  username: 'owner',
  layout: 'waterfall',
  private: false,
  user_id: 'u1',
  reactions: [],
  created_at: createdAt,
  echo_files: [],
  tags: [],
})

const baseDataset = {
  schema_version: 1,
  generated_at: SECOND,
  base_url: '/blog/',
  init_status: { initialized: true, owner_exists: true },
  settings: { site_title: 'Ech0' },
  hello: { hello: 'hi' },
  agent: { enable: false },
  tags: [],
  heatmap: [],
  comments: [],
  comment_form: { enable_comment: false },
  connects: [],
  connect: { server_name: 'Ech0' },
}

// Every case needs its own module instance because the dataset is memoized per
// module; a static import cannot be re-evaluated, so the test reloads it.
async function loadAdapter(echos: App.Api.Ech0.Echo[]): Promise<StaticAdapter> {
  vi.resetModules()
  const payload = { ...baseDataset, echos } as unknown as StaticDataset
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => ({ ok: true, status: 200, json: async () => payload })),
  )
  return await import('@/service/request/static-adapter')
}

const query = async (
  adapter: StaticAdapter,
  body: Partial<App.Api.Ech0.EchoQueryParams>,
): Promise<App.Api.Ech0.PaginationResult> => {
  const res = await adapter.handleStaticRequest<App.Api.Ech0.PaginationResult>(
    '/api/echo/query',
    'POST',
    body,
  )
  expect(res.code).toBe(1)
  return res.data
}

const ids = (page: App.Api.Ech0.PaginationResult): string[] => page.items.map((item) => item.id)

beforeAll(() => {
  window.__ECH0_STATIC_BASE__ = '/blog/'
})

describe('static adapter keyset ordering', () => {
  it('同一秒多条按 id DESC 排序，pageSize=2 时整组不重复、cursor 前进', async () => {
    const adapter = await loadAdapter([
      echo('e-a', SECOND),
      echo('e-b', SECOND),
      echo('e-c', SECOND),
      echo('e-d', SECOND),
    ])

    const first = await query(adapter, { page: 1, pageSize: 2, cursor: '' })
    expect(ids(first)).toEqual(['e-d', 'e-c'])

    const second = await query(adapter, { page: 1, pageSize: 2, cursor: first.nextCursor ?? '' })
    expect(ids(second)).toEqual(['e-b', 'e-a'])
    expect([...ids(first), ...ids(second)]).toEqual(['e-d', 'e-c', 'e-b', 'e-a'])
  })

  it('pageSize=1 时同秒整组逐条返回，不跳过中间行', async () => {
    const adapter = await loadAdapter([
      echo('e-a', SECOND),
      echo('e-b', SECOND),
      echo('e-c', SECOND),
    ])

    const seen: string[] = []
    let cursor = ''
    for (let page = 0; page < 5; page++) {
      const res = await query(adapter, { page: 1, pageSize: 1, cursor })
      seen.push(...ids(res))
      if (res.nextCursor === null) break
      cursor = res.nextCursor
    }

    expect(seen).toEqual(['e-c', 'e-b', 'e-a'])
  })

  it('asc 时 tiebreak 同为 ASC，跨页无重复无缺口', async () => {
    const adapter = await loadAdapter([
      echo('e-c', SECOND),
      echo('e-a', SECOND),
      echo('e-b', SECOND + 1),
      echo('e-d', SECOND - 1),
    ])

    const first = await query(adapter, { page: 1, pageSize: 2, cursor: '', sortOrder: 'asc' })
    expect(ids(first)).toEqual(['e-d', 'e-a'])

    const second = await query(adapter, {
      page: 1,
      pageSize: 2,
      cursor: first.nextCursor ?? '',
      sortOrder: 'asc',
    })
    expect(ids(second)).toEqual(['e-c', 'e-b'])
    expect([...ids(first), ...ids(second)]).toEqual(['e-d', 'e-a', 'e-c', 'e-b'])
  })

  it('cursor 指向的 Echo 已被删除时按位置续接，而不是返回空页', async () => {
    const before = await loadAdapter([
      echo('e-d', SECOND),
      echo('e-c', SECOND),
      echo('e-b', SECOND),
      echo('e-a', SECOND),
    ])
    const page = await query(before, { page: 1, pageSize: 2, cursor: '' })
    expect(ids(page)).toEqual(['e-d', 'e-c'])

    // e-c vanishes between requests; the boundary must still be positional.
    const after = await loadAdapter([echo('e-d', SECOND), echo('e-b', SECOND), echo('e-a', SECOND)])
    const next = await query(after, { page: 1, pageSize: 2, cursor: page.nextCursor ?? '' })
    expect(ids(next)).toEqual(['e-b', 'e-a'])
  })
})
