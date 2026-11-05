<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="activity" :class="{ 'activity--active': active }">
    <component
      :is="expandable ? 'button' : 'div'"
      class="activity__row"
      :class="{ 'activity__row--static': !expandable }"
      :type="expandable ? 'button' : undefined"
      :aria-expanded="expandable ? open : undefined"
      :aria-label="expandable ? rowLabel : undefined"
      @click="onRowClick"
    >
      <span class="activity__glyph">
        <slot name="glyph" />
      </span>

      <span class="activity__label" role="status">
        <!-- Measured off-screen so the visible label can animate to the width of
             the next string; the row then glides between phases instead of
             snapping as the text changes. -->
        <span ref="measureEl" class="activity__label-measure" aria-hidden="true">{{ label }}</span>
        <span class="activity__label-clip" :style="clipStyle">
          <Transition name="activity-swap">
            <span
              :key="label"
              class="activity__label-text"
              :class="{ 'activity__label-text--shimmer': active }"
              >{{ label }}</span
            >
          </Transition>
        </span>
      </span>

      <span v-if="detail" class="activity__detail" :title="detail">{{ detail }}</span>
      <slot name="trailing" />

      <span v-if="clock" class="activity__clock">{{ clock }}</span>

      <svg
        v-if="expandable"
        class="activity__chevron"
        :class="{ 'activity__chevron--open': open }"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="m6 9 6 6 6-6" />
      </svg>
    </component>

    <div
      v-if="expandable"
      class="activity__collapse"
      :class="{ 'activity__collapse--open': open }"
      :inert="!open"
    >
      <div class="activity__clip">
        <div class="activity__body">
          <slot v-if="everOpened" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useElementSize } from '@vueuse/core'

const props = withDefaults(
  defineProps<{
    label: string
    /** The step is still running: the label shimmers and the body opens itself. */
    active?: boolean
    /** Duration beside the label, already formatted. */
    clock?: string
    /** A short trailing fact — the current query, a coverage note. */
    detail?: string
    /** No body worth revealing — the row degrades to plain text. */
    expandable?: boolean
  }>(),
  { expandable: true },
)

/**
 * `null` means "follow the run": open while it is working, closed once it has
 * settled. The first click pins a choice, and that choice then survives the run
 * ending — a trace someone opened on purpose stays open.
 */
const pinned = ref<boolean | null>(null)

const open = computed<boolean>(() => pinned.value ?? props.active === true)

const onRowClick = () => {
  if (props.expandable) pinned.value = !open.value
}

const measureEl = ref<HTMLElement | null>(null)
const { width: labelWidth } = useElementSize(measureEl)

const clipStyle = computed<Record<string, string> | undefined>(() =>
  labelWidth.value > 0 ? { width: `${Math.ceil(labelWidth.value)}px` } : undefined,
)

/**
 * The label is a live region, and live-region text is not folded into an
 * ancestor button's name, so the row names itself from its own parts.
 */
const rowLabel = computed<string>(() =>
  props.detail ? `${props.label}, ${props.detail}` : props.label,
)

/**
 * A trace nobody has opened costs nothing to have: the body renders on its first
 * reveal and then stays, so a transcript full of collapsed rows never pays to
 * render text no one asked for, and the close animation still has content to
 * animate.
 */
const everOpened = ref<boolean>(open.value)

watch(open, (now) => {
  if (now) everOpened.value = true
})

/**
 * Some providers buffer the whole thought and ship it in one burst, so the step
 * runs for tens of milliseconds and the row would open and close again before
 * anyone could read it. A step that was never on screen long enough to read
 * stays open instead of vanishing; the reader closes it when they are done.
 */
const MIN_READABLE_MS = 1500

let activeSince = 0

watch(
  () => props.active === true,
  (now) => {
    if (now) {
      activeSince = Date.now()
      return
    }
    if (activeSince === 0) return
    const shown = Date.now() - activeSince
    activeSince = 0
    if (shown < MIN_READABLE_MS && pinned.value === null) pinned.value = true
  },
  { immediate: true },
)
</script>

<style scoped>
.activity {
  width: 100%;
  min-width: 0;
}

.activity__row {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  max-width: 100%;
  margin-left: -0.4rem;
  padding: 0.18rem 0.5rem 0.18rem 0.4rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-muted);
  font-size: 0.78rem;
  line-height: 1.5;
  text-align: left;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease;
}

.activity__row:hover {
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
}

.activity__row--static,
.activity__row--static:hover {
  background: transparent;
  cursor: default;
}

.activity__glyph {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  font-size: 1.05rem;
  opacity: 0.85;
}

/* A row with no glyph loses the slot's gap too, so a glyph-less status still
   starts on the transcript's text edge. */
.activity__glyph:empty {
  display: none;
}

.activity--active .activity__glyph {
  color: var(--color-accent);
  opacity: 1;
}

/* The label track is measured off-screen and the clip animates to that width,
   so swapping "Thinking" for "Searching Echos" pushes the clock aside at the
   same pace instead of teleporting it. */
.activity__label {
  position: relative;
  display: flex;
  min-width: 0;
  max-width: 100%;
}

.activity__label-measure {
  position: absolute;
  top: 0;
  left: 0;
  visibility: hidden;
  white-space: nowrap;
  pointer-events: none;
  font-weight: inherit;
}

.activity__label-clip {
  position: relative;
  flex: 0 1 auto;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  transition: width 0.45s cubic-bezier(0.22, 1, 0.36, 1);
}

.activity__label-text {
  display: block;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* A gradient swept across the text says "still going" without moving any
   layout, which is what lets this sit inline with the answer below it. */
.activity__label-text--shimmer {
  background-image: linear-gradient(
    90deg,
    var(--color-text-muted) 35%,
    var(--color-text-primary) 50%,
    var(--color-text-muted) 65%
  );
  background-clip: text;
  background-size: 200% 100%;
  color: transparent;
  animation: activity-shimmer 1.5s linear infinite;
}

@keyframes activity-shimmer {
  from {
    background-position: 150% center;
  }

  to {
    background-position: -50% center;
  }
}

/* The old label lifts out as the new one settles in, both blurred at the edges
   of the move — the same swap the answer's own text uses. */
.activity-swap-enter-active {
  transition:
    opacity 0.25s ease,
    transform 0.5s cubic-bezier(0.22, 1, 0.36, 1),
    filter 0.3s ease;
}

.activity-swap-leave-active {
  position: absolute;
  inset: 0;
  transition:
    opacity 0.25s ease,
    transform 0.5s cubic-bezier(0.22, 1, 0.36, 1),
    filter 0.3s ease;
}

.activity-swap-enter-from {
  opacity: 0;
  transform: translateY(6px);
  filter: blur(3px);
}

.activity-swap-leave-to {
  opacity: 0;
  transform: translateY(-6px);
  filter: blur(3px);
}

.activity__detail {
  flex: 0 1 auto;
  min-width: 0;
  max-width: 14rem;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 0.74rem;
  color: var(--color-text-muted);
  opacity: 0.75;
}

.activity__clock {
  flex-shrink: 0;
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
  opacity: 0.7;
}

.activity__chevron {
  width: 0.8rem;
  height: 0.8rem;
  flex-shrink: 0;
  opacity: 0;
  transition:
    transform 0.3s cubic-bezier(0.23, 1, 0.32, 1),
    opacity 0.18s ease;
}

.activity__row:hover .activity__chevron,
.activity__chevron--open {
  opacity: 0.7;
}

.activity__chevron--open {
  transform: rotate(180deg);
}

/* `0fr` → `1fr` animates the row track itself, so the body slides open at its
   natural height without anyone measuring it. */
.activity__collapse {
  display: grid;
  grid-template-rows: 0fr;
  opacity: 0;
  transition:
    grid-template-rows 0.38s cubic-bezier(0.23, 1, 0.32, 1),
    opacity 0.26s ease;
}

.activity__collapse--open {
  grid-template-rows: 1fr;
  opacity: 1;
}

.activity__clip {
  overflow: hidden;
}

.activity__body {
  position: relative;
  margin: 0.3rem 0 0 0.25rem;
  padding-left: 0.85rem;
}

.activity__body::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 1px;
  background: var(--color-border-strong);
}

@media (prefers-reduced-motion: reduce) {
  .activity__label-text--shimmer {
    background-image: none;
    color: var(--color-text-secondary);
    animation: none;
  }

  .activity__label-clip,
  .activity__chevron,
  .activity__collapse,
  .activity-swap-enter-active,
  .activity-swap-leave-active {
    transition: none;
  }

  .activity-swap-enter-from,
  .activity-swap-leave-to {
    transform: none;
    filter: none;
  }
}
</style>
