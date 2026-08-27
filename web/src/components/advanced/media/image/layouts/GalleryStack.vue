<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="stack-root">
    <div class="stack-scroll">
      <div class="stack-track">
        <div v-for="(row, rowIdx) in imageRows" :key="rowIdx" class="stack-row">
          <div
            v-for="(cell, colIdx) in row"
            :key="getImageKey(cell.image, cell.idx)"
            class="stack-card"
            :class="{ 'stack-card--row-start': colIdx === 0 }"
            :style="cardStyle(cell.idx)"
          >
            <GalleryImageItem
              :image="cell.image"
              :src="resolvedSrcs[cell.idx] || ''"
              :alt="getAlt(cell.idx)"
              :loaded="isLoaded(cell.image, cell.idx)"
              :priority="!!priority && cell.idx === 0"
              button-class="stack-btn"
              frame-class="stack-frame"
              img-class="stack-img"
              @click="open(cell.idx, $event)"
              @load="markLoaded(cell.image, cell.idx)"
              @error="markLoaded(cell.image, cell.idx)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryStackProps } from './types'

const STACK_MAX_PER_ROW = 5

const props = defineProps<GalleryStackProps>()

const imageRows = computed(() => {
  const list = props.images
  const rows: { image: App.Api.Ech0.FileObject; idx: number }[][] = []
  for (let i = 0; i < list.length; i++) {
    const rowIndex = Math.floor(i / STACK_MAX_PER_ROW)
    if (!rows[rowIndex]) rows[rowIndex] = []
    rows[rowIndex].push({ image: list[i], idx: i })
  }
  return rows
})

function buildStackZIndices(count: number): number[] {
  if (count === 0) return []
  const order = Array.from({ length: count }, (_, i) => i)
  let state = ((count + 1) * 2654435761 + 1597334677) >>> 0
  for (let i = count - 1; i > 0; i--) {
    state = (Math.imul(state, 1664525) + 1013904223) >>> 0
    const j = state % (i + 1)
    const tmp = order[i]!
    order[i] = order[j]!
    order[j] = tmp
  }
  const z: number[] = new Array(count)
  for (let rank = 0; rank < count; rank++) {
    z[order[rank]!] = 10 + rank
  }
  return z
}

const stackZIndices = computed(() => buildStackZIndices(props.images.length))

function mix32(n: number): number {
  let x = (Math.imul(n ^ 0x243f6a88, 0x9e3779b1) + 0x517cc1b7) >>> 0
  x ^= x >>> 16
  x = Math.imul(x, 0x7feb352d) >>> 0
  x ^= x >>> 15
  x = Math.imul(x, 0x846ca68b) >>> 0
  x ^= x >>> 16
  return x >>> 0
}

function stackRotateDeg(idx: number): number {
  const u = mix32(idx) / 4294967296
  return Math.round(-10 + u * 20)
}

function stackOffsetY(idx: number): number {
  return Math.round(Math.sin(idx * 1.07) * 10)
}

function cardStyle(idx: number): Record<string, string> {
  const deg = stackRotateDeg(idx)
  const y = stackOffsetY(idx)
  const z = stackZIndices.value[idx] ?? 10
  return {
    '--stack-rot': `${deg}deg`,
    '--stack-y': `${y}px`,
    zIndex: String(z),
  }
}
</script>

<style scoped>
.stack-root {
  --stamp-outer: 60px;

  --stamp-white-border: 2px;

  --stamp-corner-radius: 2px;

  --stack-hover-scale: 1.25;

  --stack-overlap: 0.22;

  --stack-row-overlap: 0.6;
  --gallery-stack-frame-shadow: var(--gallery-stack-frame-shadow);
  --gallery-stack-frame-shadow-hover: var(--gallery-stack-frame-shadow-hover);

  width: 100%;
  max-width: 100%;
  margin-left: auto;
  margin-right: auto;
  margin-bottom: 1rem;
}

.stack-scroll {
  position: relative;
  width: 100%;
  max-width: 100%;
  overflow: auto visible;
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
  scrollbar-color: rgb(0 0 0 / 12%) transparent;
  box-sizing: border-box;
}

.stack-scroll::-webkit-scrollbar {
  height: 4px;
}

.stack-track {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  width: max-content;
  max-width: 100%;
  min-width: 100%;
  box-sizing: border-box;
  padding: 6px 0 12px;
}

.stack-row {
  display: flex;
  flex-flow: row nowrap;
  align-items: center;
  width: max-content;
  max-width: 100%;
  min-height: calc(
    var(--stamp-outer) * var(--stack-hover-scale) + var(--stamp-outer) * 0.28 + 0.75rem
  );
  padding: calc(var(--stamp-outer) * (var(--stack-hover-scale) - 1) * 0.38 + 8px) 4px
    calc(var(--stamp-outer) * (var(--stack-hover-scale) - 1) * 0.32 + 6px);
  box-sizing: border-box;
}

.stack-row + .stack-row {
  margin-top: calc(var(--stamp-outer) * -1 * var(--stack-row-overlap));
}

.stack-card {
  flex: 0 0 auto;
  width: var(--stamp-outer);
  margin-left: calc(var(--stamp-outer) * -1 * var(--stack-overlap));
  transform: rotate(var(--stack-rot, 0deg)) translateY(var(--stack-y, 0)) translateZ(0);
  transform-origin: center center;
  transition: transform 0.2s ease;

  backface-visibility: hidden;
}

.stack-card--row-start {
  margin-left: 0;
}

.stack-card:focus-within,
.stack-card:hover {
  z-index: 9999 !important;
  transform: rotate(var(--stack-rot, 0deg)) translateY(var(--stack-y, 0))
    scale(var(--stack-hover-scale)) translateZ(0);
}

.stack-btn {
  display: block;
  box-sizing: border-box;
  width: var(--stamp-outer);
  height: var(--stamp-outer);
  min-width: 0;
  min-height: 0;
  padding: 0;
  overflow: visible;
}

:deep(.gallery-image-frame.stack-frame) {
  width: var(--stamp-outer);
  height: var(--stamp-outer);
  min-width: 0;
  min-height: 0;
  padding: 0;
  box-sizing: border-box;
  background: transparent;
  border: none;
  border-radius: var(--stamp-corner-radius) !important;
  overflow: hidden;
  transform: translateZ(0);
  box-shadow: var(--gallery-stack-frame-shadow);
  transition: box-shadow 0.2s ease;
}

.stack-card:hover :deep(.gallery-image-frame.stack-frame),
.stack-card:focus-within :deep(.gallery-image-frame.stack-frame) {
  box-shadow: var(--gallery-stack-frame-shadow-hover);
}

:deep(.gallery-image-frame.stack-frame .image-skeleton) {
  border-radius: var(--stamp-corner-radius) !important;
}

:deep(.gallery-image-frame.stack-frame .echoimg.stack-img) {
  display: block;
  width: 100% !important;
  height: 100% !important;
  max-width: none !important;
  max-height: none !important;
  object-fit: cover;
  object-position: center;
  border-radius: var(--stamp-corner-radius) !important;
  border: var(--stamp-white-border) solid #fff;
  box-sizing: border-box;
  box-shadow: none !important;

  transform: translateZ(0);
}

@media (prefers-reduced-motion: reduce) {
  .stack-scroll {
    scroll-behavior: auto;
  }

  .stack-card {
    transform: translateY(var(--stack-y, 0)) translateZ(0);
    transition: none;
  }

  .stack-card:focus-within,
  .stack-card:hover {
    transform: translateY(var(--stack-y, 0)) translateZ(0);
  }

  :deep(.gallery-image-frame.stack-frame) {
    transition: none;
  }

  .stack-card:hover :deep(.gallery-image-frame.stack-frame),
  .stack-card:focus-within :deep(.gallery-image-frame.stack-frame) {
    box-shadow: var(--gallery-stack-frame-shadow);
  }
}
</style>
