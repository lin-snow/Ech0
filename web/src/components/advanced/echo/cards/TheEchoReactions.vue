<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootRef" class="echo-reactions">
    <div class="echo-reactions-anchor" ref="anchorRef">
      <Transition name="reaction-pop">
        <div v-if="open" class="reaction-pop" :style="popStyle" @keydown="onMenuKeydown">
          <div
            ref="barRef"
            class="reaction-bar"
            role="menu"
            :aria-label="t('reaction.add')"
            aria-orientation="horizontal"
          >
            <span
              v-for="particle in particles"
              :key="particle.id"
              class="reaction-particle"
              :style="particleStyle(particle)"
              aria-hidden="true"
            >
              {{ particle.emoji }}
            </span>

            <button
              v-for="(emoji, index) in catalog"
              :key="emoji.key"
              :ref="(el) => setItemRef(el, index)"
              type="button"
              role="menuitem"
              class="reaction-item"
              :class="{ 'reaction-item--reacted': isReacted(emoji.key) }"
              :data-emoji="emoji.key"
              :tabindex="index === activeIndex ? 0 : -1"
              :title="label(emoji.key)"
              :aria-label="label(emoji.key)"
              :aria-pressed="isReacted(emoji.key)"
              :style="{ '--i': index }"
              @focus="activeIndex = index"
              @pointerdown.prevent="pickFromPointer($event, emoji.key)"
              @click="onItemClick($event, emoji.key)"
            >
              <span class="reaction-item-emoji">{{ emoji.emoji }}</span>
              <span v-if="countOf(emoji.key) > 0" class="reaction-item-count">
                {{ countOf(emoji.key) > 99 ? '99+' : countOf(emoji.key) }}
              </span>
            </button>

            <span class="reaction-tail" :style="{ left: `${placement.tailX}px` }" />
            <span
              class="reaction-tail reaction-tail--small"
              :style="{ left: `${placement.tailX + 6}px` }"
            />
          </div>
        </div>
      </Transition>

      <button
        ref="triggerRef"
        type="button"
        class="reaction-trigger"
        :class="{ 'reaction-trigger--with-count': totalCount > 0 && !open }"
        :disabled="disabled"
        :aria-expanded="open"
        aria-haspopup="true"
        :aria-label="open ? t('reaction.close') : t('reaction.add')"
        @pointerdown="onTriggerPointerDown"
        @click="onTriggerClick"
      >
        <svg
          v-if="open"
          class="reaction-trigger-icon"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M6 6l12 12M18 6L6 18"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
          />
        </svg>
        <template v-else-if="lead">
          <span class="reaction-trigger-emoji" aria-hidden="true">{{ lead.emoji }}</span>
          <span class="reaction-trigger-count">{{ totalCount > 99 ? '99+' : totalCount }}</span>
        </template>
        <svg
          v-else
          class="reaction-trigger-icon"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M21 12a9 9 0 1 1-9-9"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
          />
          <circle cx="8.9" cy="10" r="1.35" fill="currentColor" />
          <circle cx="15.1" cy="10" r="1.35" fill="currentColor" />
          <path
            d="M8 13.9a4.7 4.7 0 0 0 8 0"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
          />
          <path
            d="M19 2.5v5M21.5 5h-5"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
          />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { onClickOutside } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { fetchReactEcho, fetchReactionCatalog, fetchUnreactEcho } from '@/service/api'
import { isStaticMode } from '@/service/request/shared'
import { theToast } from '@/utils/toast'
import { getVisitorId } from '@/utils/visitor'

type ReactionCount = App.Api.Ech0.ReactionCount
type ReactionKey = App.Api.Ech0.ReactionKey

type ReactionEcho = {
  id: string
  reactions?: ReactionCount[]
}

type Particle = {
  id: number
  emoji: string
  originX: number
  originY: number
  x0: number
  drift: number
  rise: number
  scale: number
  tilt: number
  blur: number
  duration: number
  delay: number
}

const FALLBACK_CATALOG: App.Api.Ech0.ReactionCatalogItem[] = [
  { key: 'like', emoji: '👍' },
  { key: 'heart', emoji: '❤️' },
  { key: 'haha', emoji: '😂' },
  { key: 'wow', emoji: '😮' },
  { key: 'sad', emoji: '😢' },
  { key: 'cry', emoji: '😭' },
  { key: 'party', emoji: '🎉' },
]

const props = defineProps<{
  echo: ReactionEcho
  /** When set, reactions are sent to this remote instance instead of the local API. */
  baseUrl?: string
}>()

const { t } = useI18n()

const staticMode = isStaticMode()
const disabled = computed(() => staticMode)
const catalog = ref<App.Api.Ech0.ReactionCatalogItem[]>([...FALLBACK_CATALOG])
const items = ref<ReactionCount[]>((props.echo.reactions ?? []).map((item) => ({ ...item })))

const open = ref(false)
const activeIndex = ref(0)
const rootRef = ref<HTMLElement | null>(null)
const anchorRef = ref<HTMLElement | null>(null)
const barRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const itemRefs: HTMLElement[] = []
const particles = ref<Particle[]>([])
let particleSeq = 0
let justOpened = false

const placement = ref({ side: 'top' as 'top' | 'bottom', shift: 0, tailX: 0 })

const prefersReducedMotion = () =>
  typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches

const catalogMap = computed(
  () => new Map(catalog.value.map((item) => [item.key, item.emoji] as const)),
)

const catalogOrder = computed(
  () => new Map(catalog.value.map((item, index) => [item.key, index] as const)),
)

const totalCount = computed(() => items.value.reduce((sum, item) => sum + item.count, 0))

// The trigger mirrors the old like button: one emoji plus the total, always
// the same width, so a busy post can never push the meta row onto a new line.
const lead = computed<ReactionCount | null>(() => {
  const [first] = [...items.value].sort((a, b) => {
    if (b.count !== a.count) return b.count - a.count
    return (catalogOrder.value.get(a.key) ?? 0) - (catalogOrder.value.get(b.key) ?? 0)
  })
  return first ?? null
})

const popStyle = computed(() => ({
  [placement.value.side === 'top' ? 'bottom' : 'top']: 'calc(100% + 16px)',
  marginLeft: `${placement.value.shift}px`,
}))

watch(
  () => props.echo.reactions,
  (next) => {
    items.value = (next ?? []).map((item) => ({ ...item }))
  },
  { deep: true },
)

onClickOutside(rootRef, () => close())

const REACTION_LABELS: { key: ReactionKey; labelKey: string }[] = [
  { key: 'like', labelKey: 'reaction.like' },
  { key: 'heart', labelKey: 'reaction.heart' },
  { key: 'haha', labelKey: 'reaction.haha' },
  { key: 'wow', labelKey: 'reaction.wow' },
  { key: 'sad', labelKey: 'reaction.sad' },
  { key: 'cry', labelKey: 'reaction.cry' },
  { key: 'party', labelKey: 'reaction.party' },
]

const label = (key: ReactionKey): string => {
  const entry = REACTION_LABELS.find((item) => item.key === key)
  return entry ? String(t(entry.labelKey)) : key
}

const countOf = (key: ReactionKey): number =>
  items.value.find((item) => item.key === key)?.count ?? 0

const isReacted = (key: ReactionKey): boolean =>
  items.value.some((item) => item.key === key && item.reacted)

const setItemRef = (el: unknown, index: number) => {
  if (el instanceof HTMLElement) {
    itemRefs[index] = el
  }
}

function particleStyle(particle: Particle) {
  return {
    left: `${particle.originX}px`,
    top: `${particle.originY}px`,
    '--x0': `${particle.x0}px`,
    '--drift': `${particle.drift}px`,
    '--rise': particle.rise,
    '--scale': particle.scale,
    '--tilt': `${particle.tilt}deg`,
    '--blur': `${particle.blur}px`,
    animationDuration: `${particle.duration}s`,
    animationDelay: `${particle.delay}s`,
  }
}

async function openPicker() {
  if (disabled.value) {
    theToast.info(String(t('reaction.unavailable')))
    return
  }
  open.value = true
  await nextTick()
  placeBar()
  const reacted = catalog.value.findIndex((item) => isReacted(item.key))
  activeIndex.value = reacted >= 0 ? reacted : 0
  if (!prefersReducedMotion()) {
    itemRefs[activeIndex.value]?.focus({ preventScroll: true })
  }
}

function close() {
  if (!open.value) return
  open.value = false
  particles.value = []
}

function placeBar() {
  const trigger = triggerRef.value?.getBoundingClientRect()
  const bar = barRef.value
  if (!trigger || !bar || typeof window === 'undefined') return

  const gap = 16
  const edge = 8
  // offsetWidth/offsetHeight ignore the enter animation's scale, so the bar is
  // placed against its final size instead of a shrunken mid-flight one.
  const barWidth = bar.offsetWidth
  const barHeight = bar.offsetHeight
  const side = trigger.top - barHeight - gap < edge ? 'bottom' : 'top'
  const centered = trigger.left + trigger.width / 2 - barWidth / 2
  const overLeft = edge - centered
  const overRight = centered + barWidth - (window.innerWidth - edge)
  const shift = overLeft > 0 ? overLeft : overRight > 0 ? -overRight : 0

  placement.value = {
    side,
    shift,
    tailX: trigger.left + trigger.width / 2 - (centered + shift),
  }
}

function onTriggerPointerDown(event: PointerEvent) {
  if (disabled.value || event.button !== 0) return
  if (open.value) return

  openPicker()
  justOpened = true

  const onRelease = (release: PointerEvent) => {
    document.removeEventListener('pointerup', onRelease)
    const target = document.elementFromPoint(release.clientX, release.clientY)
    const picked = target?.closest<HTMLElement>('[data-emoji]')
    const key = picked?.dataset.emoji as ReactionKey | undefined
    if (picked && key) {
      react(key, picked.getBoundingClientRect())
    }
    // Releasing off the trigger fires no click, so nothing else clears the guard.
    if (!triggerRef.value?.contains(target)) {
      justOpened = false
    }
  }

  document.addEventListener('pointerup', onRelease)
}

function onTriggerClick() {
  if (justOpened) {
    justOpened = false
    return
  }
  if (open.value) {
    close()
  } else {
    void openPicker()
  }
}

function onItemClick(event: MouseEvent, key: ReactionKey) {
  // Pointer picks already fired on pointerdown; only keyboard clicks land here.
  if (event.detail !== 0) return
  react(key, (event.currentTarget as HTMLElement).getBoundingClientRect())
}

function pickFromPointer(event: PointerEvent, key: ReactionKey) {
  react(key, (event.currentTarget as HTMLElement).getBoundingClientRect())
}

function onMenuKeydown(event: KeyboardEvent) {
  const count = catalog.value.length
  if (count === 0) return

  if (event.key === 'ArrowRight') {
    activeIndex.value = (activeIndex.value + 1) % count
  } else if (event.key === 'ArrowLeft') {
    activeIndex.value = (activeIndex.value - 1 + count) % count
  } else if (event.key === 'Home') {
    activeIndex.value = 0
  } else if (event.key === 'End') {
    activeIndex.value = count - 1
  } else if (event.key === 'Escape') {
    event.preventDefault()
    close()
    triggerRef.value?.focus()
    return
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    const key = catalog.value[activeIndex.value]?.key
    if (key) react(key, itemRefs[activeIndex.value]?.getBoundingClientRect() ?? null)
    return
  } else {
    return
  }

  event.preventDefault()
  itemRefs[activeIndex.value]?.focus()
}

function react(key: ReactionKey, from: DOMRect | null) {
  emitBurst(key, from)
  request(key, !isReacted(key))
}

function emitBurst(key: ReactionKey, from: DOMRect | null) {
  const bar = barRef.value?.getBoundingClientRect()
  if (prefersReducedMotion() || !from || !bar) return

  const emoji = catalogMap.value.get(key) ?? ''
  const originX = from.left + from.width / 2 - bar.left
  const originY = from.top + from.height / 2 - bar.top
  const seed = particleSeq

  const next: Particle[] = Array.from({ length: 5 }, (_, i) => {
    const lane = Math.random() * 2 - 1
    const direction = lane < 0 ? -1 : 1
    return {
      id: seed + i + 1,
      emoji,
      originX,
      originY,
      x0: lane * 6,
      drift: lane * 78,
      rise: 450 * (0.86 + Math.random() * 0.14),
      scale: 0.78 + Math.random() * 0.27,
      tilt: (1 + Math.random() * 3) * direction,
      blur: (0.18 + Math.random() * 0.12) * 26,
      duration: 1.4 + Math.random() * 0.4,
      delay: i * 0.25,
    }
  })
  particleSeq += 5
  particles.value = [...particles.value, ...next].slice(-60)

  window.setTimeout(() => {
    const ids = new Set(next.map((item) => item.id))
    particles.value = particles.value.filter((item) => !ids.has(item.id))
  }, 2600)
}

async function request(key: ReactionKey, add: boolean) {
  if (staticMode) {
    theToast.info(String(t('reaction.unavailable')))
    return
  }

  const snapshot = items.value.map((item) => ({ ...item }))
  optimistic(key, add)

  try {
    const next = props.baseUrl
      ? await callRemote(key, add ? 'PUT' : 'DELETE')
      : add
        ? (await fetchReactEcho(props.echo.id, key)).data
        : (await fetchUnreactEcho(props.echo.id, key)).data
    items.value = (next ?? []).map((item) => ({ ...item }))
  } catch {
    items.value = snapshot
    if (props.baseUrl) {
      theToast.error(String(t('reaction.failed')))
    }
  }
}

async function callRemote(key: ReactionKey, method: 'PUT' | 'DELETE'): Promise<ReactionCount[]> {
  const response = await fetch(
    `${props.baseUrl}/api/echo/${encodeURIComponent(props.echo.id)}/reactions/${encodeURIComponent(key)}`,
    {
      method,
      credentials: 'omit',
      headers: { 'X-Ech0-Visitor': getVisitorId() },
    },
  )
  const payload = (await response.json()) as App.Api.Response<ReactionCount[]>
  if (!response.ok || payload.code !== 1) {
    throw new Error(payload.msg || 'reaction failed')
  }
  return payload.data ?? []
}

function optimistic(key: ReactionKey, add: boolean) {
  const existing = items.value.find((item) => item.key === key)

  if (add) {
    if (existing) {
      items.value = items.value.map((item) =>
        item.key === key ? { ...item, count: item.count + 1, reacted: true } : item,
      )
      return
    }
    items.value = [
      ...items.value,
      { key, emoji: catalogMap.value.get(key) ?? '', count: 1, reacted: true },
    ]
    return
  }

  if (!existing) return
  items.value =
    existing.count > 1
      ? items.value.map((item) =>
          item.key === key ? { ...item, count: item.count - 1, reacted: false } : item,
        )
      : items.value.filter((item) => item.key !== key)
}

void (async () => {
  if (props.baseUrl || staticMode) return
  const res = await fetchReactionCatalog()
  if (res.code === 1 && res.data?.length) {
    catalog.value = res.data
  }
})()
</script>

<style scoped>
.echo-reactions {
  display: inline-flex;
  align-items: center;
  flex: none;
}

.echo-reactions-anchor {
  position: relative;
  display: inline-flex;
}

.reaction-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.18rem;
  height: 1.55rem;
  min-width: 1.55rem;
  padding: 0;
  border: 1px solid var(--color-border-subtle);
  border-radius: 999px;
  background: var(--color-bg-surface);
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    border-color 150ms ease,
    color 150ms ease,
    background-color 150ms ease,
    transform 120ms ease;
}

.reaction-trigger--with-count {
  padding: 0 0.42rem;
}

.reaction-trigger:hover:not(:disabled) {
  border-color: var(--color-border-strong);
  color: var(--color-text-primary);
}

.reaction-trigger:active:not(:disabled) {
  transform: scale(0.94);
}

.reaction-trigger:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.reaction-trigger-icon {
  width: 0.95rem;
  height: 0.95rem;
}

.reaction-trigger-emoji {
  font-size: 0.86rem;
  line-height: 1;
}

.reaction-trigger-count {
  font-family: var(--font-family-mono);
  font-size: 0.68rem;
}

.reaction-pop {
  position: absolute;
  left: 50%;
  z-index: 40;
  transform: translateX(-50%);
  transform-origin: bottom center;
}

.reaction-pop.reaction-pop-enter-active {
  animation: reaction-pop 240ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

.reaction-pop.reaction-pop-leave-active {
  animation: reaction-pop-out 140ms ease-in forwards;
}

@keyframes reaction-pop {
  from {
    opacity: 0;
    transform: translate(-50%, 8px) scale(0.85);
  }

  to {
    opacity: 1;
    transform: translate(-50%, 0) scale(1);
  }
}

@keyframes reaction-pop-out {
  from {
    opacity: 1;
    transform: translate(-50%, 0) scale(1);
  }

  to {
    opacity: 0;
    transform: translate(-50%, 5px) scale(0.9);
  }
}

.reaction-bar {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.1rem;
  padding: 0.28rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 999px;
  background: var(--color-bg-surface);
  box-shadow: 0 12px 32px rgb(0 0 0 / 12%);
}

.reaction-tail {
  position: absolute;
  bottom: -0.28rem;
  width: 0.55rem;
  height: 0.55rem;
  margin-left: -0.275rem;
  border-right: 1px solid var(--color-border-subtle);
  border-bottom: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  transform: rotate(45deg);
}

.reaction-tail--small {
  bottom: -0.85rem;
  width: 0.3rem;
  height: 0.3rem;
  margin-left: -0.15rem;
  border-radius: 999px;
  border: none;
}

.reaction-item {
  position: relative;
  z-index: 1;
  display: grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  padding: 0;
  border: none;
  border-radius: 999px;
  background: transparent;
  line-height: 1;
  cursor: pointer;
  outline: none;
  animation: reaction-item-in 280ms cubic-bezier(0.34, 1.56, 0.64, 1) backwards;
  animation-delay: calc(40ms + var(--i) * 35ms);
  transition:
    transform 160ms cubic-bezier(0.34, 1.56, 0.64, 1),
    background-color 160ms ease;
}

.reaction-item:hover {
  background: var(--color-bg-muted);
  transform: translateY(-4px) scale(1.28);
}

.reaction-item:active {
  transform: scale(0.92);
}

.reaction-item--reacted {
  background: color-mix(in srgb, var(--color-accent) 12%, transparent);
}

.reaction-item:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 1px;
}

.reaction-item-emoji {
  font-size: 1.3rem;
  line-height: 1;
}

.reaction-item-count {
  position: absolute;
  right: -0.05rem;
  bottom: -0.02rem;
  min-width: 0.7rem;
  padding: 0 0.12rem;
  border-radius: 999px;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-family: var(--font-family-mono);
  font-size: 0.52rem;
  line-height: 0.85rem;
  text-align: center;
  box-shadow: 0 0 0 1px var(--color-border-subtle);
}

@keyframes reaction-item-in {
  from {
    opacity: 0;
    transform: scale(0.4);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}

.reaction-particle {
  position: absolute;
  z-index: 0;
  margin: -0.65rem 0 0 -0.65rem;
  font-size: 1.3rem;
  line-height: 1;
  pointer-events: none;
  opacity: 0;
  animation-name: reaction-particle-rise;
  animation-timing-function: cubic-bezier(0.4, 0.3, 0.5, 1);
  animation-fill-mode: forwards;
}

@keyframes reaction-particle-rise {
  0% {
    opacity: 0;
    transform: translate(var(--x0), 0) scale(0.6) rotate(0deg);
    filter: blur(0);
  }

  3% {
    opacity: 1;
  }

  55% {
    opacity: 0.9;
    filter: blur(calc(var(--blur) * 0.35));
  }

  100% {
    opacity: 0;
    transform: translate(calc(var(--x0) + var(--drift)), calc(var(--rise) * -1px))
      scale(var(--scale)) rotate(var(--tilt));
    filter: blur(var(--blur));
  }
}

@media (prefers-reduced-motion: reduce) {
  .reaction-pop.reaction-pop-enter-active,
  .reaction-pop.reaction-pop-leave-active,
  .reaction-item {
    animation: none;
  }

  .reaction-item,
  .reaction-trigger {
    transition: none;
  }

  .reaction-particle {
    display: none;
  }
}
</style>
