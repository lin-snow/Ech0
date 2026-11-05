// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { ApiResult, EchoPost, EchoQueryPage } from '../types/echo'
import { timeoutSignal } from '../utils/fetchTimeout'
import { timeValueToUnixSeconds } from '../utils/timeValue'

const ECHO_QUERY_TIMEOUT_MS = 8000

function normalizeEchoPostTimes(item: EchoPost): EchoPost {
  return {
    ...item,
    created_at: timeValueToUnixSeconds(item.created_at),
    tags: item.tags?.map((t) => ({
      ...t,
      ...(t.created_at != null
        ? { created_at: timeValueToUnixSeconds(t.created_at) }
        : {}),
    })),
  }
}

export interface EchoQueryBody {
  page: number
  pageSize: number
  search: string
  tagIds: string[]
  sortBy: string
  sortOrder: string
}

const DEFAULT_QUERY: EchoQueryBody = {
  page: 1,
  pageSize: 10,
  search: '',
  tagIds: [],
  sortBy: '',
  sortOrder: 'desc',
}

function buildQueryBody(merged: EchoQueryBody): Record<string, unknown> {
  const body: Record<string, unknown> = {
    page: merged.page,
    pageSize: merged.pageSize,
    search: merged.search,
    tagIds: merged.tagIds,
    sortOrder: merged.sortOrder,
  }
  // Instances validate `sortBy` against an enum and answer 422 for an empty
  // string, so "default sort" has to be expressed by omitting the field.
  if (merged.sortBy) {
    body.sortBy = merged.sortBy
  }
  return body
}

export async function queryInstancePage(
  instanceUrl: string,
  body: Partial<EchoQueryBody> = {},
  signal?: AbortSignal,
): Promise<EchoPost[]> {
  const merged: EchoQueryBody = { ...DEFAULT_QUERY, ...body }
  // The hub is a read-only view of remote instances: it sends no identity and
  // no custom headers, so every request is a plain JSON POST that any version
  // accepts without a CORS allowlist entry.
  const res = await fetch(`${instanceUrl}/api/echo/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(buildQueryBody(merged)),
    credentials: 'omit',
    signal: timeoutSignal(signal, ECHO_QUERY_TIMEOUT_MS),
  })

  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }

  const json: unknown = await res.json()
  if (typeof json !== 'object' || json === null) {
    throw new Error('invalid JSON')
  }
  const r = json as ApiResult<EchoQueryPage>
  if (r.code !== 1 || !r.data || !Array.isArray(r.data.items)) {
    throw new Error(r.msg || 'query failed')
  }

  return r.data.items.map(normalizeEchoPostTimes)
}
