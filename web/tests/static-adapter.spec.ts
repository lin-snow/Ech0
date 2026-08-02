// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { beforeAll, describe, expect, it, vi } from 'vitest'
import type { Mock } from 'vitest'

import { handleStaticRequest } from '@/service/request/static-adapter'
import type { StaticDataset } from '@/service/request/static-adapter'

const echo = (id: string, content: string, createdAt: number, favCount = 0): App.Api.Ech0.Echo => ({
  id,
  content,
  username: 'owner',
  layout: 'waterfall',
  private: false,
  user_id: 'u1',
  fav_count: favCount,
  created_at: createdAt,
  echo_files: [],
  tags: [],
})

const comment = (id: string, echoId: string): App.Api.Comment.CommentItem => ({
  id,
  echo_id: echoId,
  parent_id: null,
  user_id: undefined,
  nickname: 'guest',
  email: '',
  website: '',
  content: `comment ${id}`,
  status: 'approved',
  hot: false,
  source: 'guest',
  created_at: 1_770_000_000,
  updated_at: 1_770_000_000,
})

const dataset = {
  schema_version: 1,
  generated_at: 1_770_000_000,
  base_url: '/blog/',
  init_status: { initialized: true, owner_exists: true },
  settings: { site_title: 'Ech0' },
  hello: { hello: 'hi' },
  agent: { enable: false },
  echos: [
    echo('e5', 'hello world', 1_770_000_500, 3),
    echo('e4', 'HELLO again', 1_770_000_400, 9),
    echo('e3', 'nothing here', 1_770_000_300, 1),
    echo('e2', 'say hello softly', 1_770_000_200, 5),
    echo('e1', 'unrelated note', 1_770_000_100, 0),
  ],
  tags: [{ id: 't1', name: 'tag', usage_count: 1, created_at: 1_770_000_000 }],
  heatmap: [{ date: '2026-08-02', count: 1 }],
  comments: [comment('c1', 'e1'), comment('c2', 'e2'), comment('c3', 'e1')],
  comment_form: { enable_comment: false },
  connects: [],
  connect: { server_name: 'Ech0' },
} as unknown as StaticDataset

let fetchMock: Mock

beforeAll(() => {
  // dataset 只拉一次，缓存在 module-level Promise 里，所以整个文件共享同一份桩数据
  window.__ECH0_STATIC_BASE__ = '/blog/'
  fetchMock = vi.fn(async () => ({ ok: true, status: 200, json: async () => dataset }))
  vi.stubGlobal('fetch', fetchMock)
})

describe('handleStaticRequest', () => {
  it('从注入基址拉取 dataset.json 且只拉一次', async () => {
    await handleStaticRequest('/init/status', 'GET', undefined)
    await handleStaticRequest('/api/tags', 'GET', undefined)

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock.mock.calls[0][0]).toBe('/blog/dataset.json')
  })

  it('POST /api/echo/query 分页与 search 过滤正确，total 为过滤后总数', async () => {
    const page1 = await handleStaticRequest<App.Api.Ech0.PaginationResult>(
      '/api/echo/query',
      'POST',
      { page: 1, pageSize: 2, search: 'hello' },
    )

    expect(page1.code).toBe(1)
    // 大小写不敏感命中 e5 / e4 / e2，按 created_at 降序
    expect(page1.data.total).toBe(3)
    expect(page1.data.items.map((item) => item.id)).toEqual(['e5', 'e4'])

    const page2 = await handleStaticRequest<App.Api.Ech0.PaginationResult>(
      '/api/echo/query',
      'POST',
      { page: 2, pageSize: 2, search: 'HeLLo' },
    )
    expect(page2.data.total).toBe(3)
    expect(page2.data.items.map((item) => item.id)).toEqual(['e2'])

    const unfiltered = await handleStaticRequest<App.Api.Ech0.PaginationResult>(
      '/api/echo/query',
      'POST',
      { page: 1, pageSize: 0 },
    )
    // pageSize < 1 回落到 10，total 是全部公开 Echo
    expect(unfiltered.data.total).toBe(5)
    expect(unfiltered.data.items).toHaveLength(5)
  })

  it('GET /api/echo/{id} 命中返回 Echo，未命中返回 code 0', async () => {
    const hit = await handleStaticRequest<App.Api.Ech0.Echo>('/api/echo/e3', 'GET', undefined)
    expect(hit.code).toBe(1)
    expect(hit.data.id).toBe('e3')

    const miss = await handleStaticRequest<App.Api.Ech0.Echo | null>(
      '/api/echo/nope',
      'GET',
      undefined,
    )
    expect(miss.code).toBe(0)
    expect(miss.data).toBeNull()
  })

  it('GET /api/comments?echo_id=x 只返回该 echo 的评论', async () => {
    const res = await handleStaticRequest<App.Api.Comment.CommentItem[]>(
      '/api/comments?echo_id=e1&_t=123',
      'GET',
      undefined,
    )

    expect(res.code).toBe(1)
    expect(res.data.map((item) => item.id)).toEqual(['c1', 'c3'])
  })

  it('PUT /api/echo/like/{id} 返回 code 0（冻结展示）', async () => {
    const res = await handleStaticRequest('/api/echo/like/e1', 'PUT', undefined)

    expect(res.code).toBe(0)
    expect(res.data).toBeNull()
  })

  it('未知路径返回 code 0', async () => {
    const res = await handleStaticRequest('/api/panel/comments?page=1', 'GET', undefined)

    expect(res.code).toBe(0)
    expect(res.msg).toBe('Not available in static mode')
    expect(res.data).toBeNull()
  })
})
