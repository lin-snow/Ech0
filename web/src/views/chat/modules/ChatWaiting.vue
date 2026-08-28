<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="waiting" role="status">
    <span class="waiting__grid" aria-hidden="true">
      <span
        v-for="(delay, i) in CELL_DELAYS"
        :key="i"
        class="waiting__cell"
        :style="{ animationDelay: `${delay}ms` }"
      />
    </span>
    <span class="waiting__label">{{ t('chatPanel.waiting') }}</span>
    <span class="waiting__clock">{{ formatDuration(elapsedMs) }}</span>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { formatDuration, useElapsed } from './useElapsed'

const { t } = useI18n()

/**
 * Per-cell offsets in ms. The middle row leads and the outer rows trail it, so
 * the 3×3 grid reads as a wave crossing the block rather than nine blinks.
 */
const CELL_DELAYS = [90, 180, 270, 0, 90, 180, 90, 180, 270] as const

// Mounted only while the answer is pending, so the clock runs for this element's
// whole life.
const elapsedMs = useElapsed(ref(true))
</script>

<style scoped>
.waiting {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.4rem 0;
}

.waiting__grid {
  display: grid;
  grid-template-columns: repeat(3, 4px);
  flex-shrink: 0;
  gap: 1.5px;
}

.waiting__cell {
  width: 4px;
  height: 4px;
  border-radius: 1px;
  background: var(--color-text-primary);
  opacity: 0.15;
  animation: waiting-pixel 650ms ease-in-out infinite;
}

@keyframes waiting-pixel {
  0%,
  100% {
    opacity: 0.15;
  }

  18%,
  42% {
    opacity: 1;
  }

  62% {
    opacity: 0.15;
  }
}

.waiting__label {
  background-image: linear-gradient(
    90deg,
    var(--color-text-muted) 35%,
    var(--color-text-primary) 50%,
    var(--color-text-muted) 65%
  );
  background-clip: text;
  background-size: 200% 100%;
  color: transparent;
  font-size: 0.82rem;
  line-height: 1.5;
  animation: waiting-shimmer 1.5s linear infinite;
}

@keyframes waiting-shimmer {
  from {
    background-position: 150% center;
  }

  to {
    background-position: -50% center;
  }
}

.waiting__clock {
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: 0.74rem;
  font-variant-numeric: tabular-nums;
  opacity: 0.7;
}

@media (prefers-reduced-motion: reduce) {
  .waiting__cell {
    opacity: 0.45;
    animation: none;
  }

  .waiting__label {
    background-image: none;
    color: var(--color-text-muted);
    animation: none;
  }
}
</style>
