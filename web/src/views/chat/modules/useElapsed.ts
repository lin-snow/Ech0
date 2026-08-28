// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { onBeforeUnmount, ref, watch, type Ref } from 'vue'

const TICK_MS = 100

/**
 * Wall-clock since `active` last turned true, at 100ms resolution. Restarts on
 * every rising edge and freezes on the falling one, so a settled run keeps the
 * number it ended on until the caller swaps in the server-reported duration.
 */
export function useElapsed(active: Ref<boolean | undefined>): Ref<number> {
  const elapsedMs = ref<number>(0)
  let timer = 0
  let startedAt = 0

  const stop = () => {
    if (timer) {
      clearInterval(timer)
      timer = 0
    }
  }

  watch(
    active,
    (now) => {
      stop()
      if (now !== true) return
      startedAt = Date.now()
      elapsedMs.value = 0
      timer = window.setInterval(() => {
        elapsedMs.value = Date.now() - startedAt
      }, TICK_MS)
    },
    { immediate: true },
  )

  onBeforeUnmount(stop)

  return elapsedMs
}

/**
 * `4.2s` / `1m 04.2s`. One decimal at a fixed width so a ticking clock rendered
 * with tabular figures never shifts the text beside it.
 */
export function formatDuration(ms: number): string {
  const total = Math.max(0, ms) / 1000
  if (total < 60) return `${total.toFixed(1)}s`
  const minutes = Math.floor(total / 60)
  const seconds = total - minutes * 60
  return `${minutes}m ${seconds.toFixed(1).padStart(4, '0')}s`
}
