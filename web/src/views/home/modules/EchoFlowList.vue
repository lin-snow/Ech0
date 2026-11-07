<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="echo-flow-list" :class="{ 'is-restoring': restoring }">
    <div
      v-for="(item, index) in items"
      :key="item.id"
      :data-echo-id="item.id"
      class="echo-flow-row"
    >
      <TheEchoCard :echo="item" :index="index" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/echo/cards/TheEchoCard.vue'
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { HOME_TIMELINE_ANCHOR_KEY } from '@/constants/timeline'
import { parseTimelineAnchor } from '@/utils/timelineAnchor'
import type { TimelineAnchor } from '@/utils/timelineAnchor'

// Plain document-flow rendering. Every loaded echo is one static row: no
// estimates, no measurement pass, no scroll compensation, nothing that can
// flicker. Chosen for feeds that fit in memory (the common personal case);
// very large feeds fall back to the virtualized list.
const props = defineProps<{
  items: App.Api.Ech0.Echo[]
  scrollTarget: HTMLElement | null
  useWindow: boolean
}>()

const emit = defineEmits<{
  (e: 'anchor-change', anchor: TimelineAnchor | null): void
}>()

const initialAnchor = parseTimelineAnchor(sessionStorage.getItem(HOME_TIMELINE_ANCHOR_KEY))

let disposed = false
let restored = initialAnchor === null
const restoring = ref(false)
let programmaticScroll: number | null = null

const readScroll = (): number =>
  props.useWindow ? window.scrollY : (props.scrollTarget?.scrollTop ?? 0)

const writeScroll = (value: number) => {
  programmaticScroll = value
  if (props.useWindow) {
    window.scrollTo(0, value)
    return
  }
  const target = props.scrollTarget
  if (target) {
    target.scrollTop = value
  }
}

const viewportTop = (): number =>
  props.useWindow ? 0 : (props.scrollTarget?.getBoundingClientRect().top ?? 0)

const rowElement = (id: string): HTMLElement | null =>
  document.querySelector<HTMLElement>(`[data-echo-id="${id}"]`)

// ---------------------------------------------------------------------------
// Restore: rows are real DOM nodes, so the saved anchor is pinned with one
// exact measurement instead of an estimate/measure feedback loop.
// ---------------------------------------------------------------------------

let restoreFrame: number | null = null
let restoreAttempts = 0

const finishRestore = () => {
  restored = true
  restoring.value = false
  programmaticScroll = null
  if (restoreFrame !== null) {
    window.cancelAnimationFrame(restoreFrame)
    restoreFrame = null
  }
  captureAnchor()
}

const restoreStep = () => {
  if (disposed || restored) return
  const target = initialAnchor
  if (!target || target.offset < 0) {
    finishRestore()
    return
  }
  const row = rowElement(target.id)
  if (!row) {
    // The anchor echo is still outside the loaded pages. Stay put and wait
    // for its page instead of scrolling to an estimate.
    if (restoreAttempts < 60) {
      restoreAttempts += 1
      restoreFrame = window.requestAnimationFrame(restoreStep)
      return
    }
    finishRestore()
    return
  }
  const delta = row.getBoundingClientRect().top - viewportTop() + target.offset
  if (Math.abs(delta) < 1) {
    finishRestore()
    return
  }
  writeScroll(readScroll() + delta)
  if (restoreAttempts < 60) {
    restoreAttempts += 1
    restoreFrame = window.requestAnimationFrame(restoreStep)
    return
  }
  finishRestore()
}

const beginRestore = (): boolean => {
  if (!initialAnchor || restored) return true
  if (props.items.findIndex((item) => item.id === initialAnchor?.id) < 0) return false
  restoring.value = true
  restoreStep()
  return true
}

const onRestoreScroll = () => {
  if (restored || programmaticScroll === null) return
  // Any scroll we did not write ourselves is user intent; hand back control.
  if (Math.abs(readScroll() - programmaticScroll) > 4) {
    markUserScroll()
  }
}

// ---------------------------------------------------------------------------
// Anchor persistence
// ---------------------------------------------------------------------------

const captureAnchor = () => {
  // All rows are real nodes; the debounced currentAnchor() reads the DOM
  // directly, nothing to precompute here.
  scheduleAnchor()
}

const currentAnchor = (): TimelineAnchor | null => {
  const rows = document.querySelectorAll<HTMLElement>('.echo-flow-row')
  const first = rows[0]
  const scrollOffset = readScroll()
  if (first) {
    const start = first.getBoundingClientRect().top - viewportTop() + scrollOffset
    if (scrollOffset < start) return null
  }
  for (const row of rows) {
    const rowTop = row.getBoundingClientRect().top - viewportTop()
    if (rowTop + row.offsetHeight <= 0) continue
    const id = row.dataset.echoId
    if (!id) continue
    return { id, offset: Math.max(0, Math.round(-rowTop)) }
  }
  return null
}

let anchorTimer: number | null = null
const scheduleAnchor = () => {
  if (!restored || anchorTimer !== null) return
  anchorTimer = window.setTimeout(() => {
    anchorTimer = null
    // null = nothing useful to restore (e.g. scrolled above the list);
    // the parent clears the stored anchor so a refresh stays at the top.
    emit('anchor-change', currentAnchor())
  }, 120)
}

const emitAnchorNow = () => {
  if (anchorTimer !== null) {
    window.clearTimeout(anchorTimer)
    anchorTimer = null
  }
  if (!restored) return
  emit('anchor-change', currentAnchor())
}

const SCROLL_KEYS = new Set(['ArrowUp', 'ArrowDown', 'PageUp', 'PageDown', 'Home', 'End', ' '])
const markUserScroll = () => {
  if (!restored) finishRestore()
}
const onKeyScroll = (event: KeyboardEvent) => {
  if (SCROLL_KEYS.has(event.key)) markUserScroll()
}

let scrollBoundTo: HTMLElement | Window | null = null
const handleScroll = () => {
  onRestoreScroll()
  if (!restored) return
  captureAnchor()
  scheduleAnchor()
}
const unbindScroll = () => {
  scrollBoundTo?.removeEventListener('scroll', handleScroll)
  scrollBoundTo = null
}
const bindScroll = () => {
  unbindScroll()
  scrollBoundTo = props.useWindow ? window : props.scrollTarget
  scrollBoundTo?.addEventListener('scroll', handleScroll, { passive: true })
}

onMounted(() => {
  bindScroll()
  beginRestore()
  window.addEventListener('resize', bindScroll, { passive: true })
  window.addEventListener('pagehide', emitAnchorNow)
  window.addEventListener('wheel', markUserScroll, { passive: true })
  window.addEventListener('touchstart', markUserScroll, { passive: true })
  window.addEventListener('keydown', onKeyScroll)
})

watch(
  () => props.items.length,
  () => {
    if (!restored && restoreFrame === null) beginRestore()
  },
)

watch(
  () => props.scrollTarget,
  () => {
    nextTick(bindScroll)
  },
)

onBeforeUnmount(() => {
  disposed = true
  emitAnchorNow()
  unbindScroll()
  window.removeEventListener('resize', bindScroll)
  window.removeEventListener('pagehide', emitAnchorNow)
  window.removeEventListener('wheel', markUserScroll)
  window.removeEventListener('touchstart', markUserScroll)
  window.removeEventListener('keydown', onKeyScroll)
  if (anchorTimer !== null) {
    window.clearTimeout(anchorTimer)
    anchorTimer = null
  }
  if (restoreFrame !== null) {
    window.cancelAnimationFrame(restoreFrame)
    restoreFrame = null
  }
})
</script>

<style scoped>
.echo-flow-list {
  position: relative;
  width: 100%;
}

.echo-flow-list.is-restoring {
  pointer-events: none;
}
</style>
