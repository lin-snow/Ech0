<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="frame" class="sketch-frame">
    <svg
      v-if="size.width > 0 && size.height > 0"
      class="sketch-frame__stroke"
      :viewBox="`0 0 ${size.width} ${size.height}`"
      preserveAspectRatio="none"
      aria-hidden="true"
      focusable="false"
    >
      <path :d="path" :stroke-dasharray="dashPattern" />
    </svg>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useElementSize } from '@vueuse/core'

/**
 * Hand-drawn frame: the outline is a wobbling dashed stroke sampled from the
 * real box size, so dashes stay even on any aspect ratio while the edge still
 * reads as pen rather than as a CSS border.
 */
const props = withDefaults(
  defineProps<{
    /** Changes the wobble and dash rhythm without changing the box. */
    seed?: number
    radius?: number
  }>(),
  { seed: 7, radius: 12 },
)

const STEP = 11
const WOBBLE = 1
const DASH_PAIRS = 14

/** Deterministic PRNG: the same box always draws the same stroke. */
const createRandom = (seed: number) => {
  let state = seed >>> 0
  return () => {
    state = (state + 0x6d2b79f5) >>> 0
    let t = Math.imul(state ^ (state >>> 15), 1 | state)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const frame = ref<HTMLDivElement | null>(null)
const box = useElementSize(frame)
const size = computed(() => ({
  width: Math.round(box.width.value),
  height: Math.round(box.height.value),
}))

type Point = { x: number; y: number }

const outlinePoints = (
  width: number,
  height: number,
  radius: number,
  random: () => number,
): Point[] => {
  const points: Point[] = []
  const push = (x: number, y: number, nx: number, ny: number) => {
    const offset = (random() * 2 - 1) * WOBBLE
    points.push({ x: x + nx * offset, y: y + ny * offset })
  }

  const r = Math.max(0, Math.min(radius, width / 2, height / 2))
  const edgeSteps = (length: number) => Math.max(1, Math.round(length / STEP))
  const cornerSteps = 3

  const line = (fromX: number, fromY: number, toX: number, toY: number, nx: number, ny: number) => {
    const steps = edgeSteps(Math.hypot(toX - fromX, toY - fromY))
    for (let i = 1; i <= steps; i++) {
      const t = i / steps
      push(fromX + (toX - fromX) * t, fromY + (toY - fromY) * t, nx, ny)
    }
  }

  const corner = (cx: number, cy: number, fromRad: number) => {
    for (let i = 1; i <= cornerSteps; i++) {
      const angle = fromRad + (Math.PI / 2) * (i / cornerSteps)
      const cos = Math.cos(angle)
      const sin = Math.sin(angle)
      push(cx + r * cos, cy + r * sin, cos, sin)
    }
  }

  push(r, 0, 0, -1)
  line(r, 0, width - r, 0, 0, -1)
  corner(width - r, r, -Math.PI / 2)
  line(width, r, width, height - r, 1, 0)
  corner(width - r, height - r, 0)
  line(width - r, height, r, height, 0, 1)
  corner(r, height - r, Math.PI / 2)
  line(0, height - r, 0, r, -1, 0)
  corner(r, r, Math.PI)

  return points
}

const path = computed(() => {
  const { width, height } = size.value
  if (!width || !height) return ''
  const points = outlinePoints(width, height, props.radius, createRandom(props.seed))
  const [first, ...rest] = points
  if (!first) return ''
  const move = `M${first.x.toFixed(1)} ${first.y.toFixed(1)}`
  const lines = rest.map((point) => `L${point.x.toFixed(1)} ${point.y.toFixed(1)}`)
  return `${move}${lines.join('')}Z`
})

const dashPattern = computed(() => {
  const random = createRandom(props.seed * 31 + 11)
  const parts: string[] = []
  for (let i = 0; i < DASH_PAIRS; i++) {
    parts.push((4.5 + random() * 5).toFixed(1), (3 + random() * 3.5).toFixed(1))
  }
  return parts.join(' ')
})
</script>

<style scoped>
.sketch-frame {
  position: relative;
}

.sketch-frame__stroke {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: visible;
  pointer-events: none;
}

.sketch-frame__stroke path {
  fill: none;
  stroke: var(--color-border-strong);
  stroke-width: 1.25;
  stroke-linecap: round;
  stroke-linejoin: round;
  opacity: 0.85;
}
</style>
