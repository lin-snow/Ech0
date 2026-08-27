// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { ApiResult } from '../types/echo'
import { timeoutSignal } from '../utils/fetchTimeout'

export interface HealthzData {
  status: string
  version: string
}

const MIN_VERSION = '4.4.0'
const HEALTHZ_TIMEOUT_MS = 5000

function parseSemver(s: string): [number, number, number] | null {
  const m = /^(\d+)\.(\d+)\.(\d+)/.exec(s.trim())
  if (!m) return null
  return [Number(m[1]), Number(m[2]), Number(m[3])]
}

export function isVersionAtLeast(current: string, minimum: string): boolean {
  const a = parseSemver(current)
  const b = parseSemver(minimum)
  if (!a || !b) return false
  for (let i = 0; i < 3; i++) {
    if (a[i]! > b[i]!) return true
    if (a[i]! < b[i]!) return false
  }
  return true
}

export function meetsHubMinVersion(version: string): boolean {
  return isVersionAtLeast(version, MIN_VERSION)
}

export async function fetchHealthz(
  instanceUrl: string,
  signal?: AbortSignal,
): Promise<{ ok: true; version: string; status: string } | { ok: false; message: string }> {
  try {
    const res = await fetch(`${instanceUrl}/healthz`, {
      method: 'GET',
      credentials: 'omit',
      signal: timeoutSignal(signal, HEALTHZ_TIMEOUT_MS),
    })
    if (!res.ok) {
      return { ok: false, message: `HTTP ${res.status}` }
    }
    const json: unknown = await res.json()
    if (typeof json !== 'object' || json === null) {
      return { ok: false, message: 'invalid JSON' }
    }
    const r = json as ApiResult<HealthzData>
    if (r.code !== 1 || !r.data) {
      return { ok: false, message: r.msg || 'healthz failed' }
    }
    const { status, version } = r.data
    if (typeof version !== 'string' || !version.trim()) {
      return { ok: false, message: 'missing version' }
    }
    return { ok: true, version: version.trim(), status: typeof status === 'string' ? status : 'ok' }
  } catch (e) {
    const message = e instanceof Error ? e.message : String(e)
    return { ok: false, message }
  }
}

export function getMinSupportedVersion(): string {
  return MIN_VERSION
}
