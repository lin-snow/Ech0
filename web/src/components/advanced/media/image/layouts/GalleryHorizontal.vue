<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-[88%] mx-auto mb-4">
    <div class="horizontal-gallery">
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

      <button
        v-if="overflows"
        class="gallery-arrow gallery-arrow--prev"
        type="button"
        :disabled="!canScrollPrev"
        :aria-label="t('imageGallery.prevImage')"
        @click.stop="scrollByPage(-1)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="m14.5 6-6 6 6 6"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>

      <button
        v-if="overflows"
        class="gallery-arrow gallery-arrow--next"
        type="button"
        :disabled="!canScrollNext"
        :aria-label="t('imageGallery.nextImage')"
        @click.stop="scrollByPage(1)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path
            d="m9.5 6 6 6-6 6"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </button>

      <div v-if="overflows && images.length > 1" class="gallery-segments">
        <div class="gallery-segments__track" role="group" @pointerdown="onSegmentsPointerDown">
          <button
            v-for="(image, idx) in images"
            :key="`segment-${getImageKey(image, idx)}`"
            class="gallery-segment"
            :class="{ 'gallery-segment--active': idx === activeIndex }"
            type="button"
            :aria-label="t('imageGallery.goToImage', { index: idx + 1 })"
            :aria-current="idx === activeIndex ? 'step' : undefined"
            @click="scrollToIndex(idx)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryHorizontalProps } from './types'

const props = defineProps<GalleryHorizontalProps>()
const { t } = useI18n()

const scrollRef = ref<HTMLDivElement | null>(null)
const scrollWidth = ref(0)
const clientWidth = ref(0)
const scrollLeft = ref(0)

const overflows = computed(() => scrollWidth.value - clientWidth.value > 1)
const canScrollPrev = computed(() => scrollLeft.value > 1)
const canScrollNext = computed(() => scrollWidth.value - clientWidth.value - scrollLeft.value > 1)

const activeIndex = computed(() => {
  const max = scrollWidth.value - clientWidth.value
  const count = props.images.length
  if (max <= 0 || count <= 1) return 0
  const progress = Math.min(1, Math.max(0, scrollLeft.value / max))
  return Math.round(progress * (count - 1))
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

const scrollByPage = (direction: 1 | -1) => {
  const el = scrollRef.value
  if (!el) return
  const step = Math.max(120, el.clientWidth * 0.85)
  el.scrollBy({ left: direction * step, behavior: 'smooth' })
}

const scrollToIndex = (index: number, behavior: ScrollBehavior = 'smooth') => {
  const el = scrollRef.value
  const count = props.images.length
  if (!el || count <= 1) return
  const max = el.scrollWidth - el.clientWidth
  if (max <= 0) return
  const clamped = Math.min(count - 1, Math.max(0, index))
  el.scrollTo({ left: (max * clamped) / (count - 1), behavior })
}

const onSegmentsPointerDown = (event: PointerEvent) => {
  const track = event.currentTarget as HTMLElement | null
  const count = props.images.length
  if (!track || count <= 1 || !overflows.value) return
  const rect = track.getBoundingClientRect()
  if (rect.width <= 0) return

  const seek = (clientX: number) => {
    const ratio = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width))
    scrollToIndex(Math.round(ratio * (count - 1)), 'auto')
  }

  track.setPointerCapture?.(event.pointerId)
  seek(event.clientX)

  const onMove = (e: PointerEvent) => seek(e.clientX)
  const onUp = () => {
    track.releasePointerCapture?.(event.pointerId)
    window.removeEventListener('pointermove', onMove)
    window.removeEventListener('pointerup', onUp)
    window.removeEventListener('pointercancel', onUp)
  }
  window.addEventListener('pointermove', onMove)
  window.addEventListener('pointerup', onUp)
  window.addEventListener('pointercancel', onUp)
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
</script>

<style scoped>
.horizontal-gallery {
  position: relative;
  width: 100%;
}

.horizontal-scroll-container {
  width: 100%;
  overflow: auto hidden;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
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

.gallery-segments {
  display: flex;
  justify-content: center;
  margin-top: 6px;
}

.gallery-segments__track {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  touch-action: none;
  user-select: none;
}

.gallery-segment {
  position: relative;
  width: 18px;
  height: 14px;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.gallery-segment::after {
  content: '';
  position: absolute;
  inset: 5px 0;
  border-radius: 9999px;
  background: var(--color-gallery-scrollbar);
  transition: background 0.18s ease;
}

.gallery-segment:hover::after {
  background: color-mix(in oklab, var(--color-text-muted) 55%, transparent);
}

.gallery-segment--active::after,
.gallery-segment--active:hover::after {
  background: color-mix(in oklab, var(--color-text-primary) 72%, transparent);
}

.gallery-segment:focus-visible {
  border-radius: 9999px;
  outline: 2px solid var(--color-accent);
  outline-offset: 1px;
}

.gallery-arrow {
  position: absolute;
  top: 50%;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  padding: 0;
  border: 1px solid var(--color-border-subtle);
  border-radius: 9999px;
  background: color-mix(in oklab, var(--color-bg-surface) 82%, transparent);
  color: var(--color-text-secondary);
  box-shadow: 0 2px 10px rgb(0 0 0 / 12%);
  backdrop-filter: blur(8px);
  cursor: pointer;
  opacity: 1;
  transform: translateY(-50%);
  transition:
    opacity 0.18s ease,
    transform 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.gallery-arrow svg {
  width: 20px;
  height: 20px;
}

.gallery-arrow--prev {
  left: 6px;
}

.gallery-arrow--next {
  right: 6px;
}

.gallery-arrow:hover {
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
}

.gallery-arrow:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

/* Pointer devices: reveal the arrows on hover / keyboard focus. */
@media (any-hover: hover) {
  .horizontal-gallery .gallery-arrow {
    opacity: 0;
    transform: translateY(-50%) scale(0.92);
  }

  .horizontal-gallery:hover .gallery-arrow,
  .horizontal-gallery .gallery-arrow:focus-visible {
    opacity: 1;
    transform: translateY(-50%) scale(1);
  }
}

.horizontal-gallery .gallery-arrow:disabled {
  pointer-events: none;
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .gallery-arrow {
    transition: none;
  }
}
</style>
