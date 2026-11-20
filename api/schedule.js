// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { timingSafeEqual } from 'node:crypto'
import { send } from '@vercel/queue'

// Reconcile the current hour and cover the entire next daily cron window.
// Overlap is intentional. A repeated delivery uses the original cutoff.
export async function scheduleHours(now, publish) {
  const seconds = Math.floor(now / 1000)
  const currentHour = seconds - seconds % 3600
  for (let hour = 0; hour <= 25; hour++) {
    const due = currentHour + hour * 3600
    await publish('ech0-maintenance', { due }, {
      delaySeconds: Math.max(0, due - seconds),
      retentionSeconds: 3 * 86400,
      idempotencyKey: `hour-${due}`,
    })
  }
}

export default async function handler(req, res) {
  const secret = process.env.CRON_SECRET
  const header = Buffer.from(req.headers.authorization ?? '')
  const expected = Buffer.from(`Bearer ${secret}`)
  if (!secret || header.length !== expected.length || !timingSafeEqual(header, expected)) {
    return res.status(401).end()
  }
  if (req.method !== 'GET') return res.status(405).end()
  await scheduleHours(Date.now(), send)
  res.setHeader('Cache-Control', 'no-store')
  res.status(204).end()
}
