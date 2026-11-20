// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { QueueClient } from '@vercel/queue'

const queue = new QueueClient()

export default queue.handleNodeCallback(async ({ due }) => {
  const host = process.env.VERCEL_URL
  const secret = process.env.CRON_SECRET
  if (!host || !secret) throw new Error('Missing maintenance worker configuration')

  const headers = { authorization: `Bearer ${secret}`, 'content-type': 'application/json' }
  const protection = process.env.VERCEL_AUTOMATION_BYPASS_SECRET
  if (protection) headers['x-vercel-protection-bypass'] = protection
  const response = await fetch(`https://${host}/api/maintenance`, {
    method: 'POST', headers, body: JSON.stringify({ due }), redirect: 'error',
    signal: AbortSignal.timeout(30000),
  })
  if (!response.ok) throw new Error(`Maintenance returned HTTP ${response.status}`)
})
