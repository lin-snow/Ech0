// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import assert from 'node:assert/strict'
import test from 'node:test'
import { scheduleHours } from '../../api/schedule.js'

test('daily reconciliation covers the latest possible following cron and deduplicates overlapping hours', async () => {
  const first = Date.UTC(2026, 10, 1)
  const second = first + (24 * 60 + 59) * 60_000
  const jobs = new Map()
  const publishAt = now => async (topic, { due }, options) => {
    assert.equal(topic, 'ech0-maintenance')
    assert.equal(due % 3600, 0)
    assert.ok(options.delaySeconds < options.retentionSeconds)
    assert.equal(options.delaySeconds, Math.max(0, due - now / 1000))
    jobs.set(options.idempotencyKey, due)
  }
  await scheduleHours(first, publishAt(first))
  await scheduleHours(second, publishAt(second))
  const hours = [...jobs.values()].sort((a, b) => a - b)
  assert.equal(hours.length, 50)
  assert.ok(hours.slice(1).every((due, i) => due - hours[i] === 3600))
})
