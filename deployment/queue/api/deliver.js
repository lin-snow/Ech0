// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { QueueClient } from '@vercel/queue'

const queue = new QueueClient()

export default queue.handleNodeCallback(async ({ due }) => {
  const origin = process.env.ECH0_SERVICE_URL
  const secret = process.env.CRON_SECRET
  if (!origin || !secret) throw new Error('Missing maintenance worker configuration')

  const headers = { authorization: `Bearer ${secret}`, 'content-type': 'application/json' }
  const response = await fetch(new URL('/api/maintenance', origin), {
    method: 'POST', headers, body: JSON.stringify({ due }), redirect: 'error',
    signal: AbortSignal.timeout(30000),
  })
  if (!response.ok) throw new Error(`Maintenance returned HTTP ${response.status}`)
})
