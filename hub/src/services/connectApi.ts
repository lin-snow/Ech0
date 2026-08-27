// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { HubInstance } from '../types/hub'
import type { ApiResult } from '../types/echo'
import { timeoutSignal } from '../utils/fetchTimeout'
import { normalizeHubInstanceUrl } from '../utils/hubUrl'
import { HUB_FAN_OUT_LIMIT, pMapLimit } from '../utils/pMapLimit'

const CONNECT_TIMEOUT_MS = 6000

export async function fetchInstanceConnect(
  instanceUrl: string,
  signal?: AbortSignal,
): Promise<App.Api.Connect.Connect | null> {
  const base = normalizeHubInstanceUrl(instanceUrl)
  const res = await fetch(`${base}/api/connect`, {
    credentials: 'omit',
    signal: timeoutSignal(signal, CONNECT_TIMEOUT_MS),
  })
  if (!res.ok) return null
  const json: unknown = await res.json()
  if (typeof json !== 'object' || json === null) return null
  const r = json as ApiResult<App.Api.Connect.Connect>
  if (r.code !== 1 || !r.data) return null
  return r.data
}

export interface InstanceConnectSummary {
  urlKey: string
  id: string
  serverName: string
  username: string
  rawLogo: string
  todayEchos: number
}

interface InstanceConnectFetched {
  urlKey: string
  rawLogo: string
  summary: InstanceConnectSummary
}

export async function fetchInstancesConnectBundle(
  instances: HubInstance[],
  signal?: AbortSignal,
): Promise<{ logos: Map<string, string>; summaries: InstanceConnectSummary[] }> {
  const settled = await pMapLimit<HubInstance, InstanceConnectFetched | null>(
    instances,
    HUB_FAN_OUT_LIMIT,
    async (inst) => {
      const urlKey = normalizeHubInstanceUrl(inst.url)
      const data = await fetchInstanceConnect(urlKey, signal)
      if (!data) return null
      const rawLogo = data.logo?.trim() ?? ''
      return {
        urlKey,
        rawLogo,
        summary: {
          urlKey,
          id: inst.id,
          serverName: data.server_name?.trim() || inst.id,
          username: data.sys_username?.trim() ?? '',
          rawLogo,
          todayEchos: typeof data.today_echos === 'number' ? data.today_echos : 0,
        },
      }
    },
  )

  const logos = new Map<string, string>()
  const summaries: InstanceConnectSummary[] = []
  for (const r of settled) {
    if (r.status !== 'fulfilled' || !r.value) continue
    const { urlKey, rawLogo, summary } = r.value
    if (rawLogo) logos.set(urlKey, rawLogo)
    summaries.push(summary)
  }

  summaries.sort((a, b) => a.serverName.localeCompare(b.serverName, 'en'))
  return { logos, summaries }
}
