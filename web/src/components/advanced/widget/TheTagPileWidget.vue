<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <WidgetShell :title="t('tagPileWidget.title')" :fluid="fluid">
    <div ref="stageRef" class="tag-pile-stage" :style="{ minHeight }">
      <div v-if="!hasTags" class="tag-pile-empty">
        {{ t('tagPileWidget.empty') }}
      </div>
      <span
        v-for="item in physicsItems"
        :key="item.key"
        :ref="(el) => setTagRef(item.key, el as Element | null)"
        class="tag-pill"
        :class="{ 'is-dragging': item.dragging }"
        :style="{
          left: `${item.x}px`,
          top: `${item.y}px`,
          transform: `translate(-50%, -50%) rotate(${item.angle.toFixed(2)}deg)`,
          backgroundColor: item.backgroundColor,
          color: item.color,
          zIndex: item.zIndex,
        }"
        @pointerdown="handlePointerDown($event, item.key)"
      >
        {{ item.label }}
      </span>
    </div>
  </WidgetShell>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import WidgetShell from './WidgetShell.vue'

interface Props {
  layoutSeed?: number | string
  minHeight?: string
  ech0Version?: string
  fluid?: boolean
}

type PhysicsTag = {
  key: string
  label: string
  x: number
  y: number
  vx: number
  vy: number
  angle: number
  spin: number
  width: number
  height: number
  dragging: boolean
  pointerId: number | null
  dragOffsetX: number
  dragOffsetY: number
  lastDragX: number
  lastDragY: number
  lastDragTs: number
  zIndex: number
  backgroundColor: string
  color: string
}

const props = withDefaults(defineProps<Props>(), {
  layoutSeed: 'ech0-tag-pile',
  minHeight: '8.85rem',
  ech0Version: '--',
  fluid: false,
})

const { t } = useI18n()

const stageRef = ref<HTMLElement | null>(null)
const physicsItems = ref<PhysicsTag[]>([])
const stageRect = ref({ width: 0, height: 0 })
const rafId = ref<number | null>(null)
const lastFrameTs = ref(0)
const tagRefs = new Map<string, HTMLElement>()
let resizeObserver: ResizeObserver | null = null

const GRAVITY = 0.23
const AIR_DAMPING = 0.992
const SPIN_DAMPING = 0.992
const BOUNCE = 0.35
const GROUND_FRICTION = 0.92

const palette = [
  {
    bg: '#ece7ff',
    color: '#3f3b59',
  },
  {
    bg: '#e4f4ff',
    color: '#2f4a5a',
  },
  {
    bg: '#e9f8ec',
    color: '#355341',
  },
  {
    bg: '#fff0e6',
    color: '#5a4334',
  },
  {
    bg: '#fff7d9',
    color: '#5e5230',
  },
  {
    bg: '#ffe7ef',
    color: '#5a3b49',
  },
]

const fixedTags = computed(() => [
  `Ech0 ${props.ech0Version}`,
  'Go',
  'Gin',
  'Gorm',
  'SQLite',
  'Vue',
  'Vite',
  'TypeScript',
  'Pinia',
  'Vue-I18n',
  'Pnpm',
])

const normalizedTags = computed(() =>
  fixedTags.value
    .map((tag) => tag.trim())
    .filter(Boolean)
    .slice(0, 24),
)
const hasTags = computed(() => normalizedTags.value.length > 0)

const normalizeSeed = (seed: number | string) => {
  if (typeof seed === 'number') return Math.abs(seed) || 1
  let value = 2166136261
  for (let i = 0; i < seed.length; i += 1) {
    value ^= seed.charCodeAt(i)
    value = Math.imul(value, 16777619)
  }
  return value >>> 0 || 1
}

const createRandom = (seed: number) => {
  let value = seed >>> 0
  return () => {
    value += 0x6d2b79f5
    let t = value
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const setTagRef = (key: string, el: Element | null) => {
  if (el instanceof HTMLElement) {
    tagRefs.set(key, el)
  } else {
    tagRefs.delete(key)
  }
}

const getLocalPoint = (event: PointerEvent) => {
  const rect = stageRef.value?.getBoundingClientRect()
  if (!rect) return null
  return {
    x: event.clientX - rect.left,
    y: event.clientY - rect.top,
  }
}

const updateStageRect = () => {
  const rect = stageRef.value?.getBoundingClientRect()
  if (!rect) return
  stageRect.value = { width: rect.width, height: rect.height }
}

const applyBounds = (item: PhysicsTag) => {
  const { width, height } = stageRect.value
  if (!width || !height) return
  const halfW = item.width / 2
  const halfH = item.height / 2

  if (item.x < halfW) {
    item.x = halfW
    item.vx = Math.abs(item.vx) * BOUNCE
  } else if (item.x > width - halfW) {
    item.x = width - halfW
    item.vx = -Math.abs(item.vx) * BOUNCE
  }

  if (item.y < halfH) {
    item.y = halfH
    item.vy = Math.abs(item.vy) * BOUNCE
  } else if (item.y > height - halfH) {
    item.y = height - halfH
    item.vy = -Math.abs(item.vy) * BOUNCE
    item.vx *= GROUND_FRICTION
    item.spin *= 0.88
    if (Math.abs(item.vy) < 0.12) item.vy = 0
    if (Math.abs(item.vx) < 0.04) item.vx = 0
  }
}

const resolveCollisions = () => {
  const items = physicsItems.value
  for (let pass = 0; pass < 2; pass += 1) {
    for (let i = 0; i < items.length; i += 1) {
      const a = items[i]
      for (let j = i + 1; j < items.length; j += 1) {
        const b = items[j]
        const dx = b.x - a.x
        const dy = b.y - a.y
        const overlapX = (a.width + b.width) / 2 - Math.abs(dx)
        const overlapY = (a.height + b.height) / 2 - Math.abs(dy)
        if (overlapX <= 0 || overlapY <= 0) continue

        if (overlapX < overlapY) {
          const push = overlapX / 2
          const direction = dx >= 0 ? 1 : -1
          if (!a.dragging) a.x -= push * direction
          if (!b.dragging) b.x += push * direction
          if (!a.dragging) a.vx *= 0.94
          if (!b.dragging) b.vx *= 0.94
        } else {
          const push = overlapY / 2
          const direction = dy >= 0 ? 1 : -1
          if (!a.dragging) a.y -= push * direction
          if (!b.dragging) b.y += push * direction
          if (!a.dragging) a.vy *= 0.9
          if (!b.dragging) b.vy *= 0.9
        }

        applyBounds(a)
        applyBounds(b)
      }
    }
  }
}

const tick = (timestamp: number) => {
  if (!lastFrameTs.value) lastFrameTs.value = timestamp
  const deltaMs = Math.min(32, Math.max(8, timestamp - lastFrameTs.value))
  const dt = deltaMs / 16.666
  lastFrameTs.value = timestamp

  for (const item of physicsItems.value) {
    if (item.dragging) continue
    item.vy += GRAVITY * dt
    item.x += item.vx * dt
    item.y += item.vy * dt
    item.angle += item.spin * dt

    item.vx *= Math.pow(AIR_DAMPING, dt)
    item.vy *= Math.pow(AIR_DAMPING, dt)
    item.spin *= Math.pow(SPIN_DAMPING, dt)

    applyBounds(item)
  }

  resolveCollisions()
  rafId.value = requestAnimationFrame(tick)
}

const startLoop = () => {
  if (rafId.value !== null) return
  lastFrameTs.value = 0
  rafId.value = requestAnimationFrame(tick)
}

const stopLoop = () => {
  if (rafId.value !== null) {
    cancelAnimationFrame(rafId.value)
    rafId.value = null
  }
}

const updateMeasuredSizes = () => {
  for (const item of physicsItems.value) {
    const el = tagRefs.get(item.key)
    if (!el) continue
    item.width = el.offsetWidth || item.width
    item.height = el.offsetHeight || item.height
    applyBounds(item)
  }
}

const resetPhysics = async () => {
  const tags = normalizedTags.value
  if (!hasTags.value || !stageRef.value) {
    physicsItems.value = []
    return
  }

  updateStageRect()
  const { width, height } = stageRect.value
  if (width <= 0 || height <= 0) return

  const seed = normalizeSeed(props.layoutSeed) ^ normalizeSeed(fixedTags.value.join('|'))
  const random = createRandom(seed)
  const initialItems: PhysicsTag[] = tags.map((label, index) => {
    const color = palette[Math.floor(random() * palette.length)] ?? palette[0]
    const estimatedWidth = Math.max(58, label.length * 7.6 + 28)
    const estimatedHeight = 28
    return {
      key: `${label}-${index}`,
      label: label.toUpperCase(),
      x: Math.max(estimatedWidth / 2, Math.min(width - estimatedWidth / 2, random() * width)),
      y: -random() * 120 - index * 14,
      vx: (random() - 0.5) * 1.6,
      vy: random() * 0.6,
      angle: (random() - 0.5) * 14,
      spin: (random() - 0.5) * 0.32,
      width: estimatedWidth,
      height: estimatedHeight,
      dragging: false,
      pointerId: null,
      dragOffsetX: 0,
      dragOffsetY: 0,
      lastDragX: 0,
      lastDragY: 0,
      lastDragTs: 0,
      zIndex: index + 1,
      backgroundColor: color.bg,
      color: color.color,
    }
  })
  physicsItems.value = initialItems
  await nextTick()
  updateMeasuredSizes()
  startLoop()
}

const handlePointerDown = (event: PointerEvent, key: string) => {
  const item = physicsItems.value.find((entry) => entry.key === key)
  if (!item) return
  const point = getLocalPoint(event)
  if (!point) return

  item.dragging = true
  item.pointerId = event.pointerId
  item.dragOffsetX = point.x - item.x
  item.dragOffsetY = point.y - item.y
  item.lastDragX = point.x
  item.lastDragY = point.y
  item.lastDragTs = performance.now()
  item.vx = 0
  item.vy = 0
  item.spin = 0
  item.zIndex = Math.max(...physicsItems.value.map((entry) => entry.zIndex), 0) + 1

  if (event.currentTarget instanceof HTMLElement) {
    event.currentTarget.setPointerCapture(event.pointerId)
  }
}

const handlePointerMove = (event: PointerEvent) => {
  const item = physicsItems.value.find(
    (entry) => entry.dragging && entry.pointerId === event.pointerId,
  )
  if (!item) return
  const point = getLocalPoint(event)
  if (!point) return
  const now = performance.now()
  const dt = Math.max(1, now - item.lastDragTs)

  item.x = point.x - item.dragOffsetX
  item.y = point.y - item.dragOffsetY
  applyBounds(item)
  item.vx = ((point.x - item.lastDragX) / dt) * 16
  item.vy = ((point.y - item.lastDragY) / dt) * 16
  item.lastDragX = point.x
  item.lastDragY = point.y
  item.lastDragTs = now
}

const releasePointer = (pointerId: number) => {
  const item = physicsItems.value.find((entry) => entry.dragging && entry.pointerId === pointerId)
  if (!item) return
  item.dragging = false
  item.pointerId = null
  item.spin = Math.max(-0.45, Math.min(0.45, item.vx * 0.06))
}

const handlePointerUp = (event: PointerEvent) => {
  releasePointer(event.pointerId)
}

const handleResize = () => {
  updateStageRect()
  for (const item of physicsItems.value) {
    applyBounds(item)
  }
}

const minHeight = computed(() => props.minHeight)

watch(
  () => normalizedTags.value.join('|'),
  async () => {
    await resetPhysics()
  },
)

watch(
  () => props.ech0Version,
  async () => {
    await resetPhysics()
  },
)

onMounted(async () => {
  await nextTick()
  updateStageRect()
  await resetPhysics()
  resizeObserver = new ResizeObserver(() => {
    handleResize()
    updateMeasuredSizes()
  })
  if (stageRef.value) {
    resizeObserver.observe(stageRef.value)
  }
  window.addEventListener('pointermove', handlePointerMove)
  window.addEventListener('pointerup', handlePointerUp)
  window.addEventListener('pointercancel', handlePointerUp)
})

onBeforeUnmount(() => {
  stopLoop()
  resizeObserver?.disconnect()
  resizeObserver = null
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', handlePointerUp)
  window.removeEventListener('pointercancel', handlePointerUp)
})
</script>

<style scoped>
.tag-pile-stage {
  position: relative;
  border-radius: 0.75rem;
  background: transparent;
  border: 1px dashed var(--color-border-subtle);
  overflow: hidden;
  padding: 0.25rem;
}

.tag-pile-empty {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  font-size: 0.78rem;
  letter-spacing: 0.02em;
}

.tag-pill {
  position: absolute;
  padding: 0.32rem 0.78rem;
  border-radius: 999px;
  border: 1px solid var(--tag-pill-border);
  box-shadow:
    0 1px 2px var(--tag-pill-shadow-outer),
    inset 0 1px 0 var(--tag-pill-shadow-inner);
  font-family: var(--font-family-mono);
  font-size: 0.68rem;
  line-height: 1;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  white-space: nowrap;
  user-select: none;
  touch-action: none;
  cursor: grab;
}

.tag-pill.is-dragging {
  cursor: grabbing;
}
</style>
