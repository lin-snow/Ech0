// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export function timeoutSignal(outer: AbortSignal | undefined, timeoutMs: number): AbortSignal {
  const t = AbortSignal.timeout(timeoutMs)
  return outer ? AbortSignal.any([outer, t]) : t
}
