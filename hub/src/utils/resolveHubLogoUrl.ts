// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { normalizeHubInstanceUrl } from './hubUrl'

export function resolveHubInstanceLogo(rawUrl: string | undefined, instanceOrigin: string): string {
  const value = (rawUrl || '').trim()
  if (!value) return ''

  if (/^https?:\/\//i.test(value)) return value

  const base = normalizeHubInstanceUrl(instanceOrigin)
  if (value.startsWith('/')) return `${base}${value}`
  return `${base}/${value}`
}
