<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="route-progress"
    :class="{ 'route-progress--visible': visible }"
    role="progressbar"
    aria-hidden="true"
  >
    <div class="route-progress__bar" :style="{ width: progress + '%' }"></div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const visible = ref(false)
const progress = ref(0)

let showTimer: number | null = null
let trickleTimer: number | null = null
let hideTimer: number | null = null

const SHOW_DELAY = 120

const clearTimer = (id: number | null) => {
  if (id !== null) window.clearTimeout(id)
}

const stopTrickle = () => {
  if (trickleTimer !== null) {
    window.clearInterval(trickleTimer)
    trickleTimer = null
  }
}

const start = () => {
  clearTimer(hideTimer)
  hideTimer = null
  stopTrickle()

  progress.value = 0
  visible.value = false
  clearTimer(showTimer)
  showTimer = window.setTimeout(() => {
    visible.value = true
    progress.value = 12
    trickleTimer = window.setInterval(() => {
      const remaining = 90 - progress.value
      if (remaining <= 0) return
      progress.value += Math.max(0.4, remaining * 0.08)
    }, 220)
  }, SHOW_DELAY)
}

const finish = () => {
  clearTimer(showTimer)
  showTimer = null
  stopTrickle()

  if (!visible.value) {
    progress.value = 0
    return
  }

  progress.value = 100
  hideTimer = window.setTimeout(() => {
    visible.value = false
    hideTimer = window.setTimeout(() => {
      progress.value = 0
      hideTimer = null
    }, 200)
  }, 120)
}

const removeBefore = router.beforeEach(() => {
  start()
  return true
})
const removeAfter = router.afterEach(() => {
  finish()
})
const removeError = router.onError(() => {
  finish()
})

onBeforeUnmount(() => {
  removeBefore()
  removeAfter()
  removeError()
  clearTimer(showTimer)
  clearTimer(hideTimer)
  stopTrickle()
})
</script>

<style scoped>
.route-progress {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  z-index: 9999;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.route-progress--visible {
  opacity: 1;
}

.route-progress__bar {
  height: 100%;
  width: 0;
  background: var(--color-accent, #e07020);
  box-shadow: 0 0 6px var(--color-accent, #e07020);
  transition:
    width 0.2s ease,
    opacity 0.2s ease;
}
</style>
