<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    ref="listEl"
    class="echo-virtual-list"
    :class="{ 'is-restoring': restoring }"
    :style="{ height: `${totalSize}px` }"
  >
    <div
      v-for="row in virtualItems"
      :key="String(row.key)"
      :data-index="row.index"
      :data-echo-row="row.index"
      :data-echo-id="items[row.index]?.id ?? ''"
      :ref="measureRow"
      class="echo-virtual-row"
      :style="{ transform: `translateY(${row.start - scrollMargin}px)` }"
    >
      <TheEchoCard v-if="items[row.index]" :echo="items[row.index]!" :index="row.index" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Ref } from 'vue'
import { useVirtualizer, useWindowVirtualizer } from '@tanstack/vue-virtual'
import type { VirtualItem, Virtualizer } from '@tanstack/vue-virtual'
import TheEchoCard from '@/components/advanced/echo/cards/TheEchoCard.vue'
import { HOME_TIMELINE_ANCHOR_KEY } from '@/constants/timeline'
import { parseTimelineAnchor } from '@/utils/timelineAnchor'
import type { TimelineAnchor } from '@/utils/timelineAnchor'

const props = defineProps<{
  items: App.Api.Ech0.Echo[]
  scrollTarget: HTMLElement | null
  useWindow: boolean
}>()

const initialAnchor = parseTimelineAnchor(sessionStorage.getItem(HOME_TIMELINE_ANCHOR_KEY))

const emit = defineEmits<{
  (e: 'anchor-change', anchor: TimelineAnchor | null): void
}>()

const listEl = ref<HTMLElement | null>(null)
const scrollMargin = ref(0)

// Measured row heights survive reloads. Without them a restore into a
// virtualized, unmeasured region can only be estimated, and the target row
// may never be rendered.
const HEIGHT_CACHE_KEY = 'home:timeline:heights:v1'
const HEIGHT_CACHE_LIMIT = 1000

type HeightCache = { width: number; sizes: Record<string, number> }

const readHeightCache = (): HeightCache => {
  const empty = (): HeightCache => ({ width: window.innerWidth, sizes: {} })
  try {
    const raw = sessionStorage.getItem(HEIGHT_CACHE_KEY)
    if (!raw) return empty()
    const parsed = JSON.parse(raw) as Partial<HeightCache>
    if (
      typeof parsed.width !== 'number' ||
      typeof parsed.sizes !== 'object' ||
      parsed.sizes === null ||
      Math.abs(parsed.width - window.innerWidth) > 2
    ) {
      return empty()
    }
    return { width: parsed.width, sizes: parsed.sizes as Record<string, number> }
  } catch {
    return empty()
  }
}

const heightCache = readHeightCache()
let heightPersistTimer: number | null = null

const persistHeights = () => {
  if (heightPersistTimer !== null) return
  heightPersistTimer = window.setTimeout(() => {
    heightPersistTimer = null
    try {
      sessionStorage.setItem(HEIGHT_CACHE_KEY, JSON.stringify(heightCache))
    } catch {
      // Storage quota or private mode: measured heights are an optimization.
    }
  }, 300)
}

const rememberHeight = (id: string, size: number) => {
  if (!id || size <= 0 || heightCache.sizes[id] === size) return
  const keys = Object.keys(heightCache.sizes)
  if (keys.length >= HEIGHT_CACHE_LIMIT) {
    for (const key of keys.slice(0, keys.length - HEIGHT_CACHE_LIMIT + 1)) {
      delete heightCache.sizes[key]
    }
  }
  heightCache.sizes[id] = size
  persistHeights()
}

const seededMeasurements = computed(() =>
  props.items.flatMap((item, index) => {
    const size = heightCache.sizes[item.id]
    return size
      ? [{ index, key: item.id, size, start: 0, end: size, lane: 0 }]
      : []
  }),
)
let disposed = false
let restored = initialAnchor === null
const restoring = ref(false)
// Offset written by the last programmatic scroll. Scroll events during a
// restore that disagree with this value are the user's own scrolling.
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

// ---------------------------------------------------------------------------
// Viewport anchoring
//
// Rows are absolutely positioned from measured heights. Any late measurement
// (first mount, deferred measurement flush at scroll end, async content) would
// move the rows under the pointer. Remember the row at the viewport top and
// restore its on-screen offset in the same microtask the virtualizer commits
// the new transforms, so the visible content never moves.
// ---------------------------------------------------------------------------

let anchor: { id: string; top: number } | null = null

const rowElement = (index: number): HTMLElement | null =>
  listEl.value?.querySelector<HTMLElement>(`[data-index="${index}"]`) ?? null

const captureAnchor = () => {
  if (!restored) return
  const list = listEl.value
  if (!list) return
  const top = viewportTop()
  for (const row of list.querySelectorAll<HTMLElement>('[data-index]')) {
    const rect = row.getBoundingClientRect()
    if (rect.bottom <= top) continue
    const id = (row as HTMLElement).dataset.echoId
    if (!id) return
    anchor = { id, top: Math.round(rect.top - top) }
    return
  }
}

const syncAnchor = () => {
  const current = anchor
  if (!restored || !current) return
  const index = props.items.findIndex((item) => item.id === current.id)
  if (index < 0) return
  const row = rowElement(index)
  if (!row) return
  const delta = Math.round(row.getBoundingClientRect().top - viewportTop() - current.top)
  if (Math.abs(delta) < 1) return
  writeScroll(readScroll() + delta)
}

const handleVirtualizerChange = () => {
  void nextTick(syncAnchor)
}

// Rows are re-measured on mount; the estimate only fills unmeasured rows
// (never visited, therefore uncached). Using the average of previously
// measured heights keeps those guesses close to reality, which shrinks the
// range corrections the viewport anchor has to make while scrolling.
const estimateHeight = (() => {
  const sizes = heightCache.sizes
  const values = Object.values(sizes)
  if (!values.length) return 240
  const mean = values.reduce((sum, size) => sum + size, 0) / values.length
  return Math.max(48, Math.min(480, Math.round(mean)))
})()

const virtualizerOptions = computed(() => ({
  count: props.items.length,
  estimateSize: () => estimateHeight,
  overscan: 4,
  scrollMargin: scrollMargin.value,
  getItemKey: (index: number) => props.items[index]?.id ?? index,
  initialMeasurementsCache: seededMeasurements.value,
  // Compensation is owned by the viewport anchor; a second implicit
  // adjustment from the virtualizer would compound the correction.
  shouldAdjustScrollPositionOnItemSizeChange: () => false,
  onChange: handleVirtualizerChange,
}))

const elementVirtualizerOptions = computed(() => ({
  ...virtualizerOptions.value,
  getScrollElement: () => props.scrollTarget,
}))

const virtualizer = (props.useWindow
  ? useWindowVirtualizer(virtualizerOptions)
  : useVirtualizer<HTMLElement, HTMLElement>(elementVirtualizerOptions)) as unknown as Ref<
  Virtualizer<Element | Window, HTMLElement>
>

const virtualItems = computed<VirtualItem[]>(() => virtualizer.value.getVirtualItems())
const totalSize = computed(() => virtualizer.value.getTotalSize())

const measureRow = (el: unknown) => {
  if (el instanceof HTMLElement) {
    virtualizer.value.measureElement(el)
    rememberHeight(el.dataset.echoId ?? '', Math.round(el.getBoundingClientRect().height))
    return
  }
  virtualizer.value.measureElement(null)
}

const updateScrollMargin = () => {
  const list = listEl.value
  if (!list) return
  if (props.useWindow) {
    scrollMargin.value = Math.max(0, Math.round(list.getBoundingClientRect().top + window.scrollY))
    return
  }
  const root = props.scrollTarget
  if (!root) return
  const listTop = list.getBoundingClientRect().top
  const rootTop = root.getBoundingClientRect().top
  scrollMargin.value = Math.max(0, Math.round(listTop - rootTop + root.scrollTop))
}

// ---------------------------------------------------------------------------
// Session anchor restore
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

const scheduleRestoreStep = () => {
  if (restoreFrame !== null || restored || disposed) return
  restoreFrame = window.requestAnimationFrame(() => {
    restoreFrame = null
    stepRestore()
  })
}

// Restoring into a virtual list happens before the target row exists: the
// initial scroll is clamped by the not-yet-measured content height, then the
// range catches up and the target offset changes with the measurements. Each
// frame recomputes the target and follows it until the row is rendered, then
// pins it to the exact saved offset with a real DOM measurement.
const stepRestore = () => {
  if (disposed || restored) return
  const target = initialAnchor
  // Negative offsets are stale anchors saved above the list; the default
  // (stay at top) is the correct position for them.
  if (!target || target.offset < 0) {
    finishRestore()
    return
  }

  const index = props.items.findIndex((item) => item.id === target.id)
  if (index < 0) {
    if (restoreAttempts < 20) {
      restoreAttempts += 1
      scheduleRestoreStep()
    } else {
      finishRestore()
    }
    return
  }

  updateScrollMargin()
  const offsetInfo = virtualizer.value.getOffsetForIndex(index, 'start')
  if (!offsetInfo) {
    if (restoreAttempts < 20) {
      restoreAttempts += 1
      scheduleRestoreStep()
    } else {
      finishRestore()
    }
    return
  }

  const row = rowElement(index)
  if (!row) {
    writeScroll(Math.max(0, offsetInfo[0] + target.offset))
    if (restoreAttempts < 20) {
      restoreAttempts += 1
      scheduleRestoreStep()
    } else {
      finishRestore()
    }
    return
  }

  const desiredTop = -target.offset
  const delta = Math.round(row.getBoundingClientRect().top - viewportTop() - desiredTop)
  if (Math.abs(delta) < 1) {
    finishRestore()
    return
  }

  writeScroll(readScroll() + delta)
  if (restoreAttempts < 20) {
    restoreAttempts += 1
    scheduleRestoreStep()
  } else {
    finishRestore()
  }
}

const onRestoreScroll = () => {
  if (restored || programmaticScroll === null) return
  // Only our own writes move the scroller during a restore; anything else
  // (scrollbar drag, momentum, script elsewhere) is user intent.
  if (Math.abs(readScroll() - programmaticScroll) > 4) {
    markUserScroll()
  }
}

const beginRestore = (): boolean => {
  if (!initialAnchor || restored) return true
  const index = props.items.findIndex((item) => item.id === initialAnchor?.id)
  if (index < 0) return false
  restoring.value = true
  stepRestore()
  return true
}

// ---------------------------------------------------------------------------
// Anchor persistence
// ---------------------------------------------------------------------------

const currentAnchor = (): TimelineAnchor | null => {
  const rows = virtualizer.value.getVirtualItems()
  if (!rows.length) return null
  const scrollOffset = props.useWindow ? window.scrollY : (props.scrollTarget?.scrollTop ?? 0)
  // Above the first row (header area, page top) there is nothing to restore:
  // persisting a negative offset would make the next restore scroll down.
  if (scrollOffset < rows[0].start) return null
  for (const row of rows) {
    if (row.start + row.size <= scrollOffset) continue
    const item = props.items[row.index]
    if (!item) continue
    return { id: item.id, offset: Math.round(scrollOffset - row.start) }
  }
  return null
}

let anchorTimer: number | null = null
const scheduleAnchor = () => {
  if (!restored || anchorTimer !== null) return
  anchorTimer = window.setTimeout(() => {
    anchorTimer = null
    // null = above the list; the parent clears the stored anchor so a
    // refresh after scrolling back up stays at the top.
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

const handleResize = () => {
  if (Math.abs(heightCache.width - window.innerWidth) > 2) {
    heightCache.width = window.innerWidth
    heightCache.sizes = {}
    persistHeights()
  }
  scheduleMarginUpdate()
  bindScroll()
}

let marginRaf: number | null = null
const scheduleMarginUpdate = () => {
  if (marginRaf !== null) return
  marginRaf = window.requestAnimationFrame(() => {
    marginRaf = null
    updateScrollMargin()
  })
}

onMounted(() => {
  updateScrollMargin()
  bindScroll()
  beginRestore()
  window.addEventListener('resize', handleResize, { passive: true })
  window.addEventListener('pagehide', emitAnchorNow)
  window.addEventListener('wheel', markUserScroll, { passive: true })
  window.addEventListener('touchstart', markUserScroll, { passive: true })
  window.addEventListener('keydown', onKeyScroll)
})

watch(
  () => props.items.length,
  () => {
    if (!restored) beginRestore()
  },
)

watch(
  () => props.scrollTarget,
  () => {
    nextTick(() => {
      scheduleMarginUpdate()
      bindScroll()
    })
  },
)

onBeforeUnmount(() => {
  disposed = true
  emitAnchorNow()
  unbindScroll()
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('pagehide', emitAnchorNow)
  window.removeEventListener('wheel', markUserScroll)
  window.removeEventListener('touchstart', markUserScroll)
  window.removeEventListener('keydown', onKeyScroll)
  if (anchorTimer !== null) {
    window.clearTimeout(anchorTimer)
    anchorTimer = null
  }
  if (marginRaf !== null) {
    window.cancelAnimationFrame(marginRaf)
    marginRaf = null
  }
  if (restoreFrame !== null) {
    window.cancelAnimationFrame(restoreFrame)
    restoreFrame = null
  }
  if (heightPersistTimer !== null) {
    window.clearTimeout(heightPersistTimer)
    heightPersistTimer = null
  }
})
</script>

<style scoped>
.echo-virtual-list {
  position: relative;
  width: 100%;

  /* The viewport anchor compensates measurement changes; browser scroll
     anchoring would add a second, conflicting correction. */
  overflow-anchor: none;
}

.echo-virtual-list.is-restoring {
  pointer-events: none;
}

.echo-virtual-row {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}
</style>
