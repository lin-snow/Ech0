// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { QueueClient, send } from '@vercel/queue'

const queue = new QueueClient()

export default queue.handleNodeCallback(async (message) => {
  const origin = process.env.ECH0_SERVICE_URL
  const secret = process.env.CRON_SECRET
  if (!origin || !secret) throw new Error('Missing maintenance worker configuration')

  const headers = { authorization: `Bearer ${secret}`, 'content-type': 'application/json' }
  const response = await fetch(new URL('/api/maintenance', origin), {
    method: 'POST', headers, body: JSON.stringify(message), redirect: 'error',
    signal: AbortSignal.timeout(65000),
  })
  if (!response.ok) throw new Error(`Maintenance returned HTTP ${response.status}`)
  if (response.status !== 204) {
    const result = await response.json()
    for (const next of Array.isArray(result) ? result : [result]) {
      const scheduled = Boolean(next.snapshot_due)
      const key = next.cleanup ? `cleanup-${next.cleanup}-${next.revision}` : scheduled ? `schedule-${next.schedule}-${next.snapshot_due}` : `snapshot-${next.run}-${next.revision}`
      await send('ech0-maintenance', next, {
        idempotencyKey: !scheduled && message.due ? `${key}-reconcile-${message.due}` : key,
        delaySeconds: scheduled ? Math.max(0, next.snapshot_due - Math.floor(Date.now() / 1000)) : 0,
        retentionSeconds: 3 * 86400,
      })
    }
  }
})
