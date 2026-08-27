// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { timeValueToMs } from './timeValue'

export function formatHubDate(ts: number | string): string {
  const ms = timeValueToMs(ts)
  const d = new Date(ms)
  if (Number.isNaN(d.getTime())) return String(ts)
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(d)
}
