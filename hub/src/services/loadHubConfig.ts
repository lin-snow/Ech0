// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { HubConfig, HubInstance } from '../types/hub'

function stripTrailingSlash(u: string): string {
  return u.replace(/\/+$/, '')
}

function isNonEmptyString(s: unknown): s is string {
  return typeof s === 'string' && s.trim().length > 0
}

export async function loadHubConfig(signal?: AbortSignal): Promise<HubInstance[]> {
  const res = await fetch('/hub.json', { signal })
  if (!res.ok) {
    throw new Error(`hub.json: HTTP ${res.status}`)
  }
  const raw: unknown = await res.json()
  if (
    typeof raw !== 'object' ||
    raw === null ||
    !('instances' in raw) ||
    !Array.isArray((raw as HubConfig).instances)
  ) {
    throw new Error('hub.json: invalid shape (expected { instances: [...] })')
  }

  const out: HubInstance[] = []
  const seen = new Set<string>()
  for (const row of (raw as HubConfig).instances) {
    if (typeof row !== 'object' || row === null) continue
    const id = 'id' in row ? (row as { id: unknown }).id : undefined
    const url = 'url' in row ? (row as { url: unknown }).url : undefined
    if (!isNonEmptyString(id) || !isNonEmptyString(url)) continue
    try {
      const parsed = new URL(url.trim())
      if (parsed.protocol !== 'https:' && parsed.protocol !== 'http:') continue
      const normalized = stripTrailingSlash(url.trim())
      if (seen.has(normalized)) continue
      seen.add(normalized)
      out.push({ id: id.trim(), url: normalized })
    } catch {
      continue
    }
  }

  return out
}
