<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="listEl" class="echo-virtual-list" :style="{ height: `${totalSize}px` }">
    <div
      v-for="row in virtualItems"
      :key="String(row.key)"
      :data-index="row.index"
      :data-echo-row="row.index"
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
  layoutKey: string
}>()

const initialAnchor = parseTimelineAnchor(sessionStorage.getItem(HOME_TIMELINE_ANCHOR_KEY))

const emit = defineEmits<{
  (e: 'anchor-change', anchor: TimelineAnchor): void
}>()

const listEl = ref<HTMLElement | null>(null)
const scrollMargin = ref(0)

const virtualizerOptions = computed(() => ({
  count: props.items.length,
  // The measured rows average 232px (105-1104px). A closer estimate keeps
  // the scrollbar and restore target stable before rows are measured.
  estimateSize: () => 240,
  overscan: 4,
  scrollMargin: scrollMargin.value,
  getItemKey: (index: number) => props.items[index]?.id ?? index,
}))

const elementVirtualizerOptions = computed(() => ({
  ...virtualizerOptions.value,
  getScrollElement: () => props.scrollTarget,
}))

// The list switches between an inner scroll container (desktop) and the window
// (mobile). The parent remounts it with a key when that mode changes, so the
// hook choice stays fixed for this instance.
const virtualizer = (props.useWindow
  ? useWindowVirtualizer(virtualizerOptions)
  : useVirtualizer<HTMLElement, HTMLElement>(elementVirtualizerOptions)) as unknown as Ref<
  Virtualizer<Element | Window, HTMLElement>
>

const virtualItems = computed<VirtualItem[]>(() => virtualizer.value.getVirtualItems())
const totalSize = computed(() => virtualizer.value.getTotalSize())

const measureRow = (el: unknown) => {
  virtualizer.value.measureElement(el instanceof HTMLElement ? el : null)
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

const currentAnchor = (): TimelineAnchor | null => {
  const rows = virtualizer.value.getVirtualItems()
  if (!rows.length) return null
  const scrollOffset = props.useWindow ? window.scrollY : (props.scrollTarget?.scrollTop ?? 0)
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
  if (anchorTimer !== null) return
  anchorTimer = window.setTimeout(() => {
    anchorTimer = null
    const anchor = currentAnchor()
    if (anchor) emit('anchor-change', anchor)
  }, 120)
}

const restoreTimers = new Set<number>()

const emitAnchorNow = () => {
  if (anchorTimer !== null) {
    window.clearTimeout(anchorTimer)
    anchorTimer = null
  }
  const anchor = currentAnchor()
  if (anchor) emit('anchor-change', anchor)
}

let scrollBoundTo: HTMLElement | Window | null = null
const unbindScroll = () => {
  scrollBoundTo?.removeEventListener('scroll', scheduleAnchor)
  scrollBoundTo = null
}

const bindScroll = () => {
  unbindScroll()
  scrollBoundTo = props.useWindow ? window : props.scrollTarget
  scrollBoundTo?.addEventListener('scroll', scheduleAnchor, { passive: true })
}

let disposed = false
const later = (run: () => void, delay: number) => {
  const id = window.setTimeout(() => {
    restoreTimers.delete(id)
    if (!disposed) run()
  }, delay)
  restoreTimers.add(id)
}

const applyRestore = (index: number, anchor: TimelineAnchor, attempt: number, waitAttempt = 0) => {
  if (disposed || !listEl.value) return
  const instance = virtualizer.value
  const offsetInfo = instance.getOffsetForIndex(index, 'start')
  const target = Math.max(0, (offsetInfo?.[0] ?? 0) + anchor.offset)
  const maxScroll = props.useWindow
    ? document.documentElement.scrollHeight - window.innerHeight
    : (props.scrollTarget?.scrollHeight ?? 0) - (props.scrollTarget?.clientHeight ?? 0)

  // On a fresh document the virtual list may not have committed its measured
  // height yet, making the scrollable range 0 and clamping the restore to the
  // top. Wait for the layout to catch up before trusting the offset.
  if (target > 0 && maxScroll <= 0 && waitAttempt < 20) {
    later(() => applyRestore(index, anchor, attempt, waitAttempt + 1), 16)
    return
  }

  if (!offsetInfo) return
  instance.scrollToOffset(target, { align: 'start' })

  // The first scroll happens before the window around the anchor has been
  // measured. Once those rows exist, the anchor's start can move; re-apply a
  // couple of times so the card lands at the saved offset instead of an
  // estimate-derived one.
  if (attempt < 2) {
    later(() => applyRestore(index, anchor, attempt + 1), 32)
  }
}

const restoreScroll = () => {
  const anchor = initialAnchor
  if (!anchor) return
  const index = props.items.findIndex((item) => item.id === anchor.id)
  if (index < 0) return

  nextTick(() => {
    requestAnimationFrame(() => {
      if (disposed) return
      updateScrollMargin()
      requestAnimationFrame(() => applyRestore(index, anchor, 0))
    })
  })
}

let resizeObserver: ResizeObserver | null = null
let marginRaf: number | null = null
const scheduleMarginUpdate = () => {
  if (marginRaf !== null) return
  marginRaf = window.requestAnimationFrame(() => {
    marginRaf = null
    updateScrollMargin()
  })
}

const handleResize = () => {
  scheduleMarginUpdate()
  bindScroll()
}

onMounted(() => {
  nextTick(() => {
    updateScrollMargin()
    bindScroll()
    restoreScroll()
  })

  const parent = listEl.value?.parentElement
  if (parent && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(scheduleMarginUpdate)
    resizeObserver.observe(parent)
  }
  window.addEventListener('resize', handleResize, { passive: true })
  window.addEventListener('pagehide', emitAnchorNow)
})

watch(
  () => props.layoutKey,
  () => nextTick(scheduleMarginUpdate),
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
  resizeObserver?.disconnect()
  resizeObserver = null
  if (marginRaf !== null) {
    window.cancelAnimationFrame(marginRaf)
    marginRaf = null
  }
  if (anchorTimer !== null) {
    window.clearTimeout(anchorTimer)
    anchorTimer = null
  }
  restoreTimers.forEach((id) => window.clearTimeout(id))
  restoreTimers.clear()
})
</script>

<style scoped>
.echo-virtual-list {
  position: relative;
  width: 100%;

  /* Keep the browser's scroll anchoring out of the virtual layout: rows are
     absolutely positioned and measured after mount, so anchoring to a row that
     is about to unmount causes visible jumps. */
  overflow-anchor: none;
}

.echo-virtual-row {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}
</style>
