<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    ref="listEl"
    class="echo-virtual-list"
    :style="{ height: `${totalSize}px` }"
    :aria-busy="busy"
  >
    <div
      v-for="row in virtualItems"
      :key="String(row.key)"
      :data-index="row.index"
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
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import type { Ref } from 'vue'
import { useVirtualizer, useWindowVirtualizer } from '@tanstack/vue-virtual'
import type { VirtualItem, Virtualizer } from '@tanstack/vue-virtual'
import TheEchoCard from '@/components/advanced/echo/cards/TheEchoCard.vue'
import {
  HOME_TIMELINE_ANCHOR_KEY,
  parseTimelineAnchor,
  serializeTimelineAnchor,
} from '@/utils/timeline'

const props = defineProps<{
  items: App.Api.Ech0.Echo[]
  scrollTarget: HTMLElement | null
  useWindow: boolean
  busy?: boolean
}>()

const listEl = ref<HTMLElement | null>(null)
const scrollMargin = ref(0)

const readScroll = (): number =>
  props.useWindow ? window.scrollY : (props.scrollTarget?.scrollTop ?? 0)

// The virtualizer positions rows from the list's own top; scrollMargin maps
// that origin onto the page (window) or the scroll container's content.
const updateScrollMargin = () => {
  const list = listEl.value
  if (!list) return
  if (props.useWindow) {
    scrollMargin.value = Math.max(0, Math.round(list.getBoundingClientRect().top + window.scrollY))
    return
  }
  const root = props.scrollTarget
  if (!root) return
  scrollMargin.value = Math.max(
    0,
    Math.round(
      list.getBoundingClientRect().top - root.getBoundingClientRect().top + root.scrollTop,
    ),
  )
}

const virtualizerOptions = computed(() => ({
  count: props.items.length,
  estimateSize: () => 240,
  overscan: 4,
  scrollMargin: scrollMargin.value,
  getItemKey: (index: number) => props.items[index]?.id ?? index,
}))

const elementVirtualizerOptions = computed(() => ({
  ...virtualizerOptions.value,
  getScrollElement: () => props.scrollTarget,
}))

// The mode is fixed for this instance; the parent remounts the list via :key
// when the page crosses the layout breakpoint.
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

// ---------------------------------------------------------------------------
// Refresh restore
//
// One anchor (first visible Echo + how far it is scrolled past the top) is
// kept in sessionStorage. Restoring is a single offset write: rows above the
// anchor are still estimates, so the anchor row lands exactly where it was,
// and the virtualizer's default size-change compensation keeps it there while
// neighbouring rows are measured.
// ---------------------------------------------------------------------------

const initialAnchor = parseTimelineAnchor(sessionStorage.getItem(HOME_TIMELINE_ANCHOR_KEY))

const restoreAnchor = () => {
  const anchor = initialAnchor
  if (!anchor) return
  const index = props.items.findIndex((item) => item.id === anchor.id)
  const offset = index >= 0 ? virtualizer.value.getOffsetForIndex(index, 'start') : null
  if (!offset) {
    sessionStorage.removeItem(HOME_TIMELINE_ANCHOR_KEY)
    return
  }
  virtualizer.value.scrollToOffset(offset[0] + anchor.offset)
}

// ---------------------------------------------------------------------------
// Anchor persistence (debounced to one write per scroll gesture)
// ---------------------------------------------------------------------------

let anchorTimer: number | null = null

const persistAnchor = () => {
  const rows = virtualizer.value.getVirtualItems()
  if (!rows.length) return
  const scroll = readScroll()
  // Above the list there is no row to anchor; keep the previous anchor so a
  // layout/mode switch that momentarily resets the scroller cannot erase it.
  if (scroll < rows[0]!.start) return
  const row = rows.find((candidate) => candidate.end > scroll) ?? rows[rows.length - 1]!
  const item = props.items[row.index]
  if (!item) return
  sessionStorage.setItem(
    HOME_TIMELINE_ANCHOR_KEY,
    serializeTimelineAnchor({ id: item.id, offset: Math.max(0, Math.round(scroll - row.start)) }),
  )
}

const schedulePersist = () => {
  if (anchorTimer !== null) return
  anchorTimer = window.setTimeout(() => {
    anchorTimer = null
    persistAnchor()
  }, 150)
}

const flushPersist = () => {
  if (anchorTimer !== null) {
    window.clearTimeout(anchorTimer)
    anchorTimer = null
  }
  persistAnchor()
}

let scrollBoundTo: HTMLElement | Window | null = null
const bindScroll = () => {
  scrollBoundTo?.removeEventListener('scroll', schedulePersist)
  scrollBoundTo = props.useWindow ? window : props.scrollTarget
  scrollBoundTo?.addEventListener('scroll', schedulePersist, { passive: true })
}

let resizeTimer: number | null = null
const onResize = () => {
  if (resizeTimer !== null) window.clearTimeout(resizeTimer)
  resizeTimer = window.setTimeout(() => {
    resizeTimer = null
    updateScrollMargin()
    // Card heights depend on width; discard stale measurements.
    virtualizer.value.measure()
    bindScroll()
  }, 150)
}

onMounted(async () => {
  updateScrollMargin()
  bindScroll()
  window.addEventListener('pagehide', flushPersist)
  window.addEventListener('resize', onResize, { passive: true })
  await nextTick()
  restoreAnchor()
})

onBeforeUnmount(() => {
  flushPersist()
  scrollBoundTo?.removeEventListener('scroll', schedulePersist)
  window.removeEventListener('pagehide', flushPersist)
  window.removeEventListener('resize', onResize)
  if (resizeTimer !== null) window.clearTimeout(resizeTimer)
  if (anchorTimer !== null) window.clearTimeout(anchorTimer)
})
</script>

<style scoped>
.echo-virtual-list {
  position: relative;
  width: 100%;

  /* The virtualizer compensates for measurement changes itself; browser
     scroll anchoring would add a second, conflicting correction. */
  overflow-anchor: none;
}

.echo-virtual-row {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}
</style>
