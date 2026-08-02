// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * 静态站请求 adapter。
 *
 * `ech0 build` 把胶囊烘焙成 `dataset.json` + 内嵌 SPA 产物，并在 index.html 里注入
 * `window.__ECH0_STATIC__ = true`。同一份 bundle 在静态托管上运行时，`service/request`
 * 会**动态 import** 本文件，把网络请求换成纯浏览器内的 dataset 查询引擎：零后端、零 Node。
 *
 * 非静态构建下本文件不会被求值（动态 import 让它进独立 chunk），生产路径零运行时开销。
 * 契约（路由表 / 查询引擎语义 / dataset 形状）见 docs/dev/capsule/spec.md §10。
 */

declare global {
  interface Window {
    /** 静态站开关，由 `ech0 build` 注入 index.html */
    __ECH0_STATIC__?: boolean
    /** 静态站部署基址，保证以 `/` 开头且以 `/` 结尾 */
    __ECH0_STATIC_BASE__?: string
  }
}

/** `dataset.json` 的形状，字段名与后端 JSON 契约逐字一致；时间一律 Unix 秒 */
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

/** 查询引擎入参，对齐后端 EchoQueryDto（internal/service/echo/echo.go） */
type EchoQueryBody = Partial<App.Api.Ech0.EchoQueryParams>

/** dataset 只拉一次，并发请求共享同一个 Promise */
let datasetPromise: Promise<StaticDataset> | null = null

const DEFAULT_PAGE_SIZE = 10
const MAX_PAGE_SIZE = 100

const ok = <T>(data: unknown): App.Api.Response<T> => ({ code: 1, msg: '', data: data as T })

const unavailable = <T>(msg = 'Not available in static mode'): App.Api.Response<T> => ({
  code: 0,
  msg,
  data: null as T,
})

/** 拉取并缓存 dataset；失败时清空缓存，允许后续请求重试 */
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

/**
 * URL 归一：剥掉 hash/query、绝对地址只取 pathname、容忍 `/api` 前缀在与不在两种写法、
 * 去掉尾斜杠。同时把 query 参数解出来（`limit` / `echo_id` 要用）。
 */
function normalizeUrl(rawUrl: string): { path: string; query: URLSearchParams } {
  const withoutHash = String(rawUrl ?? '').split('#')[0]
  const queryIndex = withoutHash.indexOf('?')
  let path = queryIndex >= 0 ? withoutHash.slice(0, queryIndex) : withoutHash
  const query = new URLSearchParams(queryIndex >= 0 ? withoutHash.slice(queryIndex + 1) : '')

  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(path)) {
    try {
      path = new URL(path).pathname
    } catch {
      // 解析失败就按原样当作路径处理
    }
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

/** query 里的 limit，缺省 fallback，夹到 [1, 100] */
function readLimit(query: URLSearchParams, fallback: number): number {
  const raw = Number.parseInt(query.get('limit') ?? '', 10)
  const limit = Number.isFinite(raw) ? raw : fallback
  return Math.max(1, Math.min(limit, MAX_PAGE_SIZE))
}

/** Echo 的 created_at 允许 number（Unix 秒）或 ISO 字符串，统一折算成 Unix 秒 */
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

/** 静态站冻结展示：私密 Echo 一律不下发 */
function publicEchos(dataset: StaticDataset): App.Api.Ech0.Echo[] {
  return (dataset.echos ?? []).filter((echo) => echo.private !== true)
}

/** 查询引擎：过滤 → 排序 → 分页，`total` 是过滤后（分页前）的总数 */
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

/**
 * 同月同日发布的 Echo。`sameYear` 为真时限定今天（`/echo/today`），
 * 否则是历年同月同日（`/echo/onthisday`）。时区一律取浏览器本地时区。
 */
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

/** 路由表：未命中的一律返回 code 0（index.ts 在静态模式下不弹 toast） */
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
      return ok<T>(candidates[Math.floor(Math.random() * candidates.length)])
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
      // 静态站不主动探测远端实例
      return ok<T>([])
    default:
      break
  }

  // 命中详情：GET /echo/{id}
  const echoMatch = method === 'GET' ? /^\/echo\/([^/]+)$/.exec(path) : null
  if (echoMatch) {
    const echo = publicEchos(dataset).find((item) => item.id === echoMatch[1])
    return echo ? ok<T>(echo) : unavailable<T>('Echo not found')
  }

  return unavailable<T>()
}

/**
 * 静态模式下的请求入口：把 (url, method, body) 解释成对 dataset 的一次查询。
 * 任何写操作（点赞 / 评论 / 登录刷新）都返回 code 0，SPA 会优雅降级成只读展示。
 */
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
