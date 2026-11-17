<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <span
    class="loading-indicator"
    :class="[{ 'loading-indicator--center': center }, `loading-indicator--${size}`]"
    role="status"
    aria-live="polite"
  >
    <span class="loading-wordmark" aria-hidden="true">
      <span v-for="(letter, index) in letters" :key="index" :style="{ '--index': index }">
        {{ letter }}
      </span>
    </span>
    <span v-if="label" class="loading-label">{{ label }}</span>
    <span v-else class="sr-only">Loading</span>
  </span>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    size?: 'sm' | 'md' | 'lg'
    label?: string
    center?: boolean
  }>(),
  {
    size: 'md',
    label: '',
    center: true,
  },
)

// The startup loader (index.html) hops the letters of the wordmark, one glyph
// at a time. This is that same motion shrunk down to fit a button or a list
// row — the letters must stay the letters, never become dots.
const letters = ['E', 'c', 'h', '0']
</script>

<style scoped>
.loading-indicator {
  /*
   * Consumers may set --spinner-size / --spinner-color on this element (the
   * component root keeps its own scope attribute, so a plain class works).
   * Both are read through a fallback instead of being redeclared here, so the
   * consumer's value always wins regardless of stylesheet order.
   */
  --loading-size: var(--spinner-size, 16px);
  --loading-color: var(--spinner-color, var(--color-text-muted));
  --label-size: 0.95rem;
  --hop-cycle: 1.4s;
  --hop-stagger: 0.09s;

  display: inline-flex;
  flex-direction: column;
  align-items: center;
  gap: 0.42rem;
  color: var(--loading-color);
  vertical-align: middle;
}

.loading-indicator--center {
  display: flex;
  width: fit-content;
  margin-inline: auto;
}

.loading-indicator--sm {
  --loading-size: var(--spinner-size, 12px);
  --label-size: 0.85rem;
}

.loading-indicator--lg {
  --loading-size: var(--spinner-size, 22px);
  --label-size: 1.05rem;
}

.loading-wordmark {
  display: inline-flex;
  align-items: baseline;
  font-family: var(--font-family-display);
  font-size: var(--loading-size);
  line-height: 1;
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.loading-wordmark span {
  display: inline-block;
  transform-origin: 50% 100%;
  animation: loading-letter-hop var(--hop-cycle) ease-in-out infinite;
  animation-delay: calc(var(--index) * var(--hop-stagger));
}

.loading-wordmark span:nth-child(odd) {
  --tilt: -4deg;
}

.loading-wordmark span:nth-child(even) {
  --tilt: 4deg;
}

.loading-label {
  font-size: var(--label-size);
  line-height: 1.2;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  border: 0;
  white-space: nowrap;
}

/* Squash, air, land, settle — the startup loader's letter hop, verbatim. */
@keyframes loading-letter-hop {
  0%,
  40%,
  100% {
    transform: translateY(0) scale(1) rotate(0deg);
  }

  10% {
    transform: translateY(0.04em) scale(1.08, 0.92) rotate(0deg);
  }

  22% {
    transform: translateY(-0.3em) scale(0.96, 1.05) rotate(var(--tilt, 0deg));
  }

  32% {
    transform: translateY(0) scale(1.06, 0.94) rotate(0deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-wordmark span {
    animation: none;
  }
}
</style>
