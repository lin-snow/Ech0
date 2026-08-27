<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-[88%] mx-auto mb-4">
    <div ref="scrollRef" class="horizontal-scroll-container" @scroll="onScroll">
      <div class="horizontal-scroll-wrapper">
        <GalleryImageItem
          v-for="(image, idx) in images"
          :key="getImageKey(image, idx)"
          :image="image"
          :src="resolvedSrcs[idx] || ''"
          :alt="getAlt(idx)"
          :loaded="isLoaded(image, idx)"
          :priority="!!priority && idx === 0"
          button-class="horizontal-item rounded-lg shrink-0"
          frame-class="h-full"
          img-class="h-full w-auto object-contain"
          :frame-style="getHorizontalAspectStyle(image, isLoaded(image, idx))"
          @click="open(idx, $event)"
          @load="onItemLoad(image, idx)"
          @error="onItemLoad(image, idx)"
        />
      </div>
    </div>
    <div
      class="horizontal-scroll-bar"
      :class="{ 'horizontal-scroll-bar--idle': !overflows }"
      role="scrollbar"
      aria-orientation="horizontal"
      :aria-valuenow="Math.round(scrollProgress * 100)"
      aria-valuemin="0"
      aria-valuemax="100"
      @pointerdown="onTrackPointerDown"
    >
      <div class="horizontal-scroll-bar__thumb" :style="thumbStyle" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryHorizontalProps } from './types'

const props = defineProps<GalleryHorizontalProps>()

const scrollRef = ref<HTMLDivElement | null>(null)
const scrollWidth = ref(0)
const clientWidth = ref(0)
const scrollLeft = ref(0)

const overflows = computed(() => scrollWidth.value - clientWidth.value > 1)
const scrollProgress = computed(() => {
  const max = scrollWidth.value - clientWidth.value
  if (max <= 0) return 0
  return Math.min(1, Math.max(0, scrollLeft.value / max))
})

const thumbStyle = computed(() => {
  if (!overflows.value) {
    return { width: '100%', left: '0%' }
  }
  const ratio = clientWidth.value / scrollWidth.value
  const widthPct = Math.max(12, ratio * 100)
  const leftPct = (100 - widthPct) * scrollProgress.value
  return { width: `${widthPct}%`, left: `${leftPct}%` }
})

const updateMetrics = () => {
  const el = scrollRef.value
  if (!el) return
  scrollWidth.value = el.scrollWidth
  clientWidth.value = el.clientWidth
  scrollLeft.value = el.scrollLeft
}

const onScroll = () => {
  const el = scrollRef.value
  if (!el) return
  scrollLeft.value = el.scrollLeft
}

const onItemLoad = (image: App.Api.Ech0.FileObject, idx: number) => {
  props.markLoaded(image, idx)
  nextTick(updateMetrics)
}

let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  updateMetrics()
  if (typeof ResizeObserver !== 'undefined' && scrollRef.value) {
    resizeObserver = new ResizeObserver(() => updateMetrics())
    resizeObserver.observe(scrollRef.value)
  }
  window.addEventListener('resize', updateMetrics)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  window.removeEventListener('resize', updateMetrics)
})

watch(
  () => props.images,
  () => nextTick(updateMetrics),
  { deep: true },
)

const onTrackPointerDown = (event: PointerEvent) => {
  if (!overflows.value) return
  const el = scrollRef.value
  const target = event.currentTarget as HTMLElement | null
  if (!el || !target) return
  const rect = target.getBoundingClientRect()
  const max = scrollWidth.value - clientWidth.value
  if (max <= 0) return
  const ratio = clientWidth.value / scrollWidth.value
  const thumbWidthPx = Math.max(rect.width * ratio, rect.width * 0.12)
  const travel = Math.max(1, rect.width - thumbWidthPx)

  const thumbLeftPx = max <= 0 ? 0 : (scrollLeft.value / max) * travel
  const localX = event.clientX - rect.left
  const insideThumb = localX >= thumbLeftPx && localX <= thumbLeftPx + thumbWidthPx
  const grabOffset = insideThumb ? localX - thumbLeftPx : thumbWidthPx / 2

  let pendingX: number | null = null
  let rafId = 0
  const flush = () => {
    rafId = 0
    if (pendingX === null) return
    const x = pendingX - rect.left - grabOffset
    pendingX = null
    const pct = Math.min(1, Math.max(0, x / travel))
    el.scrollTo({ left: pct * max, behavior: 'instant' as ScrollBehavior })
  }
  const seek = (clientX: number) => {
    pendingX = clientX
    if (!rafId) rafId = requestAnimationFrame(flush)
  }

  target.setPointerCapture?.(event.pointerId)
  seek(event.clientX)

  const onMove = (e: PointerEvent) => seek(e.clientX)
  const onUp = () => {
    if (rafId) {
      cancelAnimationFrame(rafId)
      rafId = 0
      flush()
    }
    target.releasePointerCapture?.(event.pointerId)
    window.removeEventListener('pointermove', onMove)
    window.removeEventListener('pointerup', onUp)
    window.removeEventListener('pointercancel', onUp)
  }
  window.addEventListener('pointermove', onMove)
  window.addEventListener('pointerup', onUp)
  window.addEventListener('pointercancel', onUp)
}
</script>

<style scoped>
.horizontal-scroll-container {
  position: relative;
  width: 100%;
  overflow: auto hidden;
  scrollbar-width: none;
}

.horizontal-scroll-container::-webkit-scrollbar {
  display: none;
}

.horizontal-scroll-wrapper {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  align-items: center;
}

.horizontal-item {
  height: 200px;
  width: auto;
  overflow: hidden;
}

.horizontal-scroll-bar {
  position: relative;
  width: 100%;
  height: 6px;
  margin-top: 10px;
  border-radius: 9999px;
  background: color-mix(in oklab, var(--color-text-muted) 18%, transparent);
  cursor: pointer;
  touch-action: none;
}

.horizontal-scroll-bar--idle {
  cursor: default;
  opacity: 0.55;
}

.horizontal-scroll-bar__thumb {
  position: absolute;
  top: 0;
  bottom: 0;
  height: 100%;
  border-radius: 9999px;
  background: color-mix(in oklab, var(--color-text-muted) 70%, transparent);
  transition: background 0.15s ease;
  will-change: left, width;
}

.horizontal-scroll-bar:not(.horizontal-scroll-bar--idle):hover .horizontal-scroll-bar__thumb {
  background: var(--color-text-secondary);
}
</style>
