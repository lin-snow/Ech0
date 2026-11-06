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
import type { TimelineAnchor } from '@/utils/timelineAnchor'

const props = defineProps<{
  items: App.Api.Ech0.Echo[]
  scrollTarget: HTMLElement | null
  useWindow: boolean
  layoutKey: string
  restoreAnchor: TimelineAnchor | null
}>()

const emit = defineEmits<{
  (e: 'anchor-change', anchor: TimelineAnchor): void
}>()

const listEl = ref<HTMLElement | null>(null)
const scrollMargin = ref(0)

const virtualizerOptions = computed(() => ({
  count: props.items.length,
  estimateSize: () => 280,
  overscan: 6,
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

const viewportTop = (): number => {
  if (props.useWindow) return 0
  return props.scrollTarget?.getBoundingClientRect().top ?? 0
}

const viewportBottom = (): number => {
  if (props.useWindow) return window.innerHeight
  return props.scrollTarget?.getBoundingClientRect().bottom ?? window.innerHeight
}

const currentAnchor = (): TimelineAnchor | null => {
  const list = listEl.value
  if (!list) return null
  const top = viewportTop()
  const bottom = viewportBottom()
  for (const row of list.querySelectorAll<HTMLElement>('[data-echo-row]')) {
    const rect = row.getBoundingClientRect()
    if (rect.top >= bottom) break
    if (rect.bottom <= top) continue
    const index = Number(row.dataset.index)
    const item = props.items[index]
    if (!item) continue
    return { id: item.id, offset: Math.round(top - rect.top) }
  }
  return null
}

let anchorTimer: number | null = null
const scheduleAnchor = () => {
  if (anchorTimer !== null) return
  anchorTimer = window.setTimeout(() => {
    anchorTimer = null
    updateScrollMargin()
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
  const anchor = props.restoreAnchor
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
}

.echo-virtual-row {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}
</style>
