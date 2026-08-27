// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { fetchHealthz, meetsHubMinVersion, getMinSupportedVersion } from './healthz'
import type { HubInstance } from '../types/hub'
import { HUB_FAN_OUT_LIMIT, pMapLimit } from '../utils/pMapLimit'

export interface ProbeFailure {
  instance: HubInstance
  reason: string
}

export interface ProbeResult {
  eligible: HubInstance[]
  probeFailures: ProbeFailure[]
}

type ProbeOutcome =
  | { kind: 'ok'; instance: HubInstance }
  | { kind: 'fail'; instance: HubInstance; reason: string }

export async function probeInstances(
  instances: HubInstance[],
  signal?: AbortSignal,
): Promise<ProbeResult> {
  const probeFailures: ProbeFailure[] = []
  const eligible: HubInstance[] = []

  const settled = await pMapLimit<HubInstance, ProbeOutcome>(
    instances,
    HUB_FAN_OUT_LIMIT,
    async (inst) => {
      const h = await fetchHealthz(inst.url, signal)
      if (!h.ok) {
        return { kind: 'fail', instance: inst, reason: h.message }
      }
      if (!meetsHubMinVersion(h.version)) {
        return {
          kind: 'fail',
          instance: inst,
          reason: `Version ${h.version} is below the Hub minimum (${getMinSupportedVersion()})`,
        }
      }
      return { kind: 'ok', instance: inst }
    },
  )

  for (let i = 0; i < settled.length; i++) {
    const r = settled[i]!
    if (r.status === 'rejected') {
      const inst = instances[i]!
      probeFailures.push({
        instance: inst,
        reason: r.reason instanceof Error ? r.reason.message : String(r.reason),
      })
      continue
    }
    const val = r.value
    if (val.kind === 'fail') {
      probeFailures.push({ instance: val.instance, reason: val.reason })
    } else {
      eligible.push(val.instance)
    }
  }

  return { eligible, probeFailures }
}
