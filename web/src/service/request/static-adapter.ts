// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare global {
  interface Window {
    __ECH0_STATIC__?: boolean
    __ECH0_STATIC_BASE__?: string
  }
}

export type StaticDataset = {
  schema_version: number
  generated_at: number
  base_url: string
  init_status: App.Api.Init.Status
  settings: App.Api.Setting.SystemSetting
  hello: App.Api.Ech0.HelloEch0
  agent: App.Api.Setting.AgentSetting
  echos: App.Api.Ech0.Echo[]
  tags: App.Api.Ech0.Tag[]
  heatmap: App.Api.Ech0.HeatMap
  comments: App.Api.Comment.CommentItem[]
  comment_form: App.Api.Comment.FormMeta
  connects: App.Api.Connect.Connected[]
  connect: App.Api.Connect.Connect
}

type EchoQueryBody = Partial<App.Api.Ech0.EchoQueryParams>

let datasetPromise: Promise<StaticDataset> | null = null

const DEFAULT_PAGE_SIZE = 10
const MAX_PAGE_SIZE = 100

const ok = <T>(data: unknown): App.Api.Response<T> => ({ code: 1, msg: '', data: data as T })

const unavailable = <T>(msg = 'Not available in static mode'): App.Api.Response<T> => ({
  code: 0,
  msg,
  data: null as T,
})

function loadDataset(): Promise<StaticDataset> {
  if (!datasetPromise) {
    const base = (typeof window !== 'undefined' && window.__ECH0_STATIC_BASE__) || '/'
    const url = `${base.endsWith('/') ? base : `${base}/`}dataset.json`
    datasetPromise = fetch(url, { credentials: 'same-origin' })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`dataset.json request failed with status ${response.status}`)
        }
        return response.json() as Promise<StaticDataset>
      })
      .catch((error) => {
        datasetPromise = null
        throw error
      })
  }
  return datasetPromise
}

function normalizeUrl(rawUrl: string): { path: string; query: URLSearchParams } {
  const withoutHash = String(rawUrl ?? '').split('#')[0]
  const queryIndex = withoutHash.indexOf('?')
  let path = queryIndex >= 0 ? withoutHash.slice(0, queryIndex) : withoutHash
  const query = new URLSearchParams(queryIndex >= 0 ? withoutHash.slice(queryIndex + 1) : '')

  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(path)) {
    try {
      path = new URL(path).pathname
    } catch {}
  }

  path = path.replace(/\/{2,}/g, '/')
  if (!path.startsWith('/')) {
    path = `/${path}`
  }
  if (path === '/api') {
    path = '/'
  } else if (path.startsWith('/api/')) {
    path = path.slice(4)
  }
  if (path.length > 1) {
    path = path.replace(/\/+$/, '')
  }

  return { path: path || '/', query }
}

function readLimit(query: URLSearchParams, fallback: number): number {
  const raw = Number.parseInt(query.get('limit') ?? '', 10)
  const limit = Number.isFinite(raw) ? raw : fallback
  return Math.max(1, Math.min(limit, MAX_PAGE_SIZE))
}

function toUnixSeconds(value: number | string | undefined): number {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : 0
  }
  if (typeof value === 'string' && value !== '') {
    const parsed = Date.parse(value)
    if (Number.isFinite(parsed)) {
      return Math.floor(parsed / 1000)
    }
  }
  return 0
}

function publicEchos(dataset: StaticDataset): App.Api.Ech0.Echo[] {
  return (dataset.echos ?? []).filter((echo) => echo.private !== true)
}

function queryEchos(dataset: StaticDataset, body: EchoQueryBody): App.Api.Ech0.PaginationResult {
  const search = (body.search ?? '').trim().toLowerCase()
  const tagIds = (body.tagIds ?? []).filter((id) => typeof id === 'string' && id !== '')
  const { dateFrom, dateTo } = body

  const filtered = publicEchos(dataset).filter((echo) => {
    if (search !== '' && !(echo.content ?? '').toLowerCase().includes(search)) {
      return false
    }
    if (tagIds.length > 0) {
      const echoTagIds = (echo.tags ?? []).map((tag) => tag.id)
      if (!tagIds.some((id) => echoTagIds.includes(id))) {
        return false
      }
    }
    if (typeof dateFrom === 'number' || typeof dateTo === 'number') {
      const createdAt = toUnixSeconds(echo.created_at)
      if (typeof dateFrom === 'number' && createdAt < dateFrom) {
        return false
      }
      if (typeof dateTo === 'number' && createdAt > dateTo) {
        return false
      }
    }
    return true
  })

  const byFavCount = body.sortBy === 'fav_count'
  const ascending = String(body.sortOrder ?? '').toLowerCase() === 'asc'
  const sorted = filtered.slice().sort((a, b) => {
    const delta = byFavCount
      ? (a.fav_count ?? 0) - (b.fav_count ?? 0)
      : toUnixSeconds(a.created_at) - toUnixSeconds(b.created_at)
    return ascending ? delta : -delta
  })

  const rawPageSize = Number(body.pageSize)
  let pageSize = Number.isFinite(rawPageSize) ? Math.trunc(rawPageSize) : DEFAULT_PAGE_SIZE
  if (pageSize < 1) {
    pageSize = DEFAULT_PAGE_SIZE
  } else if (pageSize > MAX_PAGE_SIZE) {
    pageSize = MAX_PAGE_SIZE
  }

  const rawPage = Number(body.page)
  const page = Number.isFinite(rawPage) && rawPage >= 1 ? Math.trunc(rawPage) : 1
  const start = (page - 1) * pageSize

  return { items: sorted.slice(start, start + pageSize), total: sorted.length }
}

function echosOnThisDay(dataset: StaticDataset, sameYear: boolean): App.Api.Ech0.Echo[] {
  const now = new Date()
  return publicEchos(dataset).filter((echo) => {
    const date = new Date(toUnixSeconds(echo.created_at) * 1000)
    return (
      date.getMonth() === now.getMonth() &&
      date.getDate() === now.getDate() &&
      (!sameYear || date.getFullYear() === now.getFullYear())
    )
  })
}

function route<T>(
  dataset: StaticDataset,
  method: string,
  path: string,
  query: URLSearchParams,
  body: unknown,
): App.Api.Response<T> {
  const key = `${method} ${path}`

  switch (key) {
    case 'GET /init/status':
      return ok<T>(dataset.init_status)
    case 'GET /settings':
      return ok<T>(dataset.settings)
    case 'GET /agent/info':
      return ok<T>(dataset.agent)
    case 'GET /hello':
      return ok<T>(dataset.hello)
    case 'GET /oauth2/status':
      return ok<T>({ enabled: false, provider: '', oauth_ready: false })
    case 'GET /passkey/status':
      return ok<T>({ passkey_ready: false })
    case 'POST /echo/query':
      return ok<T>(queryEchos(dataset, (body ?? {}) as EchoQueryBody))
    case 'GET /echo/today':
      return ok<T>(echosOnThisDay(dataset, true))
    case 'GET /echo/onthisday':
      return ok<T>(echosOnThisDay(dataset, false))
    case 'GET /echo/hot':
      return ok<T>(
        publicEchos(dataset)
          .slice()
          .sort((a, b) => (b.fav_count ?? 0) - (a.fav_count ?? 0))
          .slice(0, readLimit(query, 5)),
      )
    case 'GET /echo/random': {
      const candidates = publicEchos(dataset)
      if (candidates.length === 0) {
        return unavailable<T>()
      }
      const pick = new Uint32Array(1)
      crypto.getRandomValues(pick)
      return ok<T>(candidates[pick[0] % candidates.length])
    }
    case 'GET /tags':
      return ok<T>(dataset.tags ?? [])
    case 'GET /heatmap':
      return ok<T>(dataset.heatmap ?? [])
    case 'GET /comments/form':
      return ok<T>(dataset.comment_form)
    case 'GET /comments': {
      const echoId = query.get('echo_id') ?? ''
      return ok<T>((dataset.comments ?? []).filter((comment) => comment.echo_id === echoId))
    }
    case 'GET /comments/public':
      return ok<T>((dataset.comments ?? []).slice(0, readLimit(query, DEFAULT_PAGE_SIZE)))
    case 'GET /connect/list':
      return ok<T>(dataset.connects ?? [])
    case 'GET /connect':
      return ok<T>(dataset.connect)
    case 'GET /connects/info':
      return ok<T>([])
    default:
      break
  }

  const echoMatch = method === 'GET' ? /^\/echo\/([^/]+)$/.exec(path) : null
  if (echoMatch) {
    const echo = publicEchos(dataset).find((item) => item.id === echoMatch[1])
    return echo ? ok<T>(echo) : unavailable<T>('Echo not found')
  }

  return unavailable<T>()
}

export async function handleStaticRequest<T>(
  url: string,
  method: string,
  body: unknown,
): Promise<App.Api.Response<T>> {
  const { path, query } = normalizeUrl(url)
  const upperMethod = String(method ?? 'GET').toUpperCase()

  let dataset: StaticDataset
  try {
    dataset = await loadDataset()
  } catch (error) {
    console.error('[ech0-static] failed to load dataset.json:', error)
    return unavailable<T>('Static dataset unavailable')
  }

  return route<T>(dataset, upperMethod, path, query, body)
}
