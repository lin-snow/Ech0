// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ref, watch, onBeforeUnmount, type Ref } from 'vue'

const WORD_INTERVAL_MS = 200

const isSpace = (ch: string): boolean => ch === ' ' || ch === '\n' || ch === '\t' || ch === '\r'

const lastWhitespaceIndex = (s: string): number => {
  for (let i = s.length - 1; i >= 0; i -= 1) {
    if (isSpace(s[i])) return i
  }
  return -1
}

const advanceWords = (s: string, from: number, n: number, limit: number): number => {
  let i = from
  for (let w = 0; w < n && i < limit; w += 1) {
    while (i < limit && isSpace(s[i])) i += 1
    while (i < limit && !isSpace(s[i])) i += 1
  }
  return i
}

const DANGLING_MARKER = /^[ \t]*([-*+]|\d{1,9}[.)]|#{1,6}|>)[ \t]*$/
const tailIsDanglingMarker = (s: string, idx: number): boolean => {
  const lineStart = s.lastIndexOf('\n', idx - 1) + 1
  return DANGLING_MARKER.test(s.slice(lineStart, idx))
}
const markerLineStart = (s: string, idx: number): number => s.lastIndexOf('\n', idx - 1) + 1

export function useSmoothReveal(source: Ref<string>, streaming: Ref<boolean>): Ref<string> {
  const displayed = ref('')
  let raf = 0
  let last = 0
  let acc = 0

  const tick = (now: number) => {
    raf = 0
    const target = source.value
    const limit = streaming.value ? lastWhitespaceIndex(target) + 1 : target.length
    if (displayed.value.length >= limit) {
      last = 0
      acc = 0
      return
    }
    if (!last) last = now
    acc += Math.min(now - last, 100)
    last = now

    let idx = displayed.value.length
    while (acc >= WORD_INTERVAL_MS && idx < limit) {
      acc -= WORD_INTERVAL_MS
      idx = advanceWords(target, idx, 1, limit)
      let guard = 0
      while (idx < limit && tailIsDanglingMarker(target, idx) && guard < 8) {
        idx = advanceWords(target, idx, 1, limit)
        guard += 1
      }
    }
    if (acc > WORD_INTERVAL_MS) acc = WORD_INTERVAL_MS
    if (tailIsDanglingMarker(target, idx)) idx = markerLineStart(target, idx)
    if (idx > displayed.value.length) displayed.value = target.slice(0, idx)
    schedule()
  }

  const schedule = () => {
    if (!raf && typeof requestAnimationFrame === 'function') {
      raf = requestAnimationFrame(tick)
    }
  }

  watch(
    source,
    (s) => {
      if (!s.startsWith(displayed.value)) displayed.value = ''
      schedule()
    },
    { immediate: true },
  )

  watch(streaming, () => schedule())

  onBeforeUnmount(() => {
    if (raf) cancelAnimationFrame(raf)
  })

  return displayed
}
