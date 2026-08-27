<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div v-if="src" class="audio-player">
    <audio
      ref="audioEl"
      class="audio-player__native"
      preload="metadata"
      :src="src"
      @loadedmetadata="onLoadedMetadata"
      @durationchange="onDurationChange"
      @timeupdate="onTimeUpdate"
      @play="isPlaying = true"
      @pause="isPlaying = false"
      @ended="onEnded"
    >
      {{ t('audioPlayer.unsupported') }}
    </audio>

    <div class="audio-player__head">
      <span class="audio-player__tag">
        <Music color="currentColor" />
        {{ t('audioPlayer.label') }}
      </span>
      <span class="audio-player__time">
        {{ formatMediaTime(currentTime) }} <span class="audio-player__time-sep">/</span>
        {{ formatMediaTime(duration) }}
      </span>
    </div>

    <div class="audio-player__controls">
      <button
        type="button"
        class="audio-player__toggle"
        :aria-label="isPlaying ? t('audioPlayer.pause') : t('audioPlayer.play')"
        @click="toggle"
      >
        <Pause v-if="isPlaying" color="currentColor" />
        <Play v-else color="currentColor" />
      </button>

      <input
        class="audio-player__range"
        type="range"
        min="0"
        max="100"
        step="0.1"
        :value="progress"
        :style="{ '--progress': `${progress}%` }"
        :aria-label="t('audioPlayer.seek')"
        @input="onScrub"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getFileUrl, getHubFileUrl } from '@/utils/other'
import { formatMediaTime } from '../shared/time'
import Play from '@/components/icons/play.vue'
import Pause from '@/components/icons/pause.vue'
import Music from '@/components/icons/music.vue'

const props = withDefaults(
  defineProps<{
    files?: App.Api.Ech0.FileObject[]
    baseUrl?: string
  }>(),
  { files: () => [] },
)

const { t } = useI18n()

const src = computed(() => {
  const file = (props.files || [])[0]
  if (!file) return ''
  return props.baseUrl ? getHubFileUrl(file, props.baseUrl) : getFileUrl(file)
})

const audioEl = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)

const progress = computed(() =>
  duration.value > 0 ? Math.min(100, (currentTime.value / duration.value) * 100) : 0,
)

function toggle() {
  const el = audioEl.value
  if (!el) return
  if (el.paused) {
    void el.play().catch(() => {})
  } else {
    el.pause()
  }
}

function onScrub(event: Event) {
  const el = audioEl.value
  const value = Number((event.target as HTMLInputElement).value)
  if (!el || !Number.isFinite(duration.value) || duration.value <= 0) return
  const next = (value / 100) * duration.value
  el.currentTime = next
  currentTime.value = next
}

function onLoadedMetadata() {
  duration.value = audioEl.value?.duration ?? 0
}

function onDurationChange() {
  duration.value = audioEl.value?.duration ?? 0
}

function onTimeUpdate() {
  currentTime.value = audioEl.value?.currentTime ?? 0
}

function onEnded() {
  isPlaying.value = false
  currentTime.value = 0
}
</script>

<style scoped>
.audio-player {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
  padding: 0.65rem 0.8rem;
  border-radius: var(--radius-md);

  border: 1px solid color-mix(in srgb, var(--color-border-subtle) 55%, transparent);
  background: color-mix(in srgb, var(--color-bg-muted) 55%, transparent);
  backdrop-filter: blur(12px) saturate(1.1);
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.audio-player:hover {
  border-color: var(--color-border-subtle);
  box-shadow: var(--shadow-soft);
}

.audio-player__native {
  display: none;
}

.audio-player__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.audio-player__tag {
  display: inline-flex;
  align-items: center;
  gap: 0.26rem;
  padding: 0.1rem 0.42rem;
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1.5;
  color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent-soft) 50%, transparent);
  border-radius: var(--radius-sm);
}

.audio-player__tag svg {
  width: 0.95rem;
  height: 0.95rem;
}

.audio-player__time {
  flex-shrink: 0;
  font-family: var(--font-family-mono);
  font-size: 0.7rem;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-muted);
  white-space: nowrap;
}

.audio-player__time-sep {
  opacity: 0.5;
}

.audio-player__controls {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.audio-player__toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 2rem;
  height: 2rem;
  padding: 0;
  font-size: 1rem;
  color: var(--color-accent);

  background: color-mix(in srgb, var(--color-accent-soft) 65%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-accent) 16%, transparent);
  border-radius: 50%;
  cursor: pointer;
  transition:
    background 0.15s ease,
    border-color 0.15s ease;
}

.audio-player__toggle:hover {
  background: color-mix(in srgb, var(--color-accent-soft) 88%, transparent);
  border-color: color-mix(in srgb, var(--color-accent) 30%, transparent);
}

.audio-player__toggle:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px var(--color-focus-ring);
}

.audio-player__range {
  flex: 1 1 auto;
  min-width: 3rem;
  height: 5px;
  margin: 0;
  padding: 0;
  border-radius: 999px;
  cursor: pointer;
  appearance: none;
  background: linear-gradient(
    to right,
    color-mix(in srgb, var(--color-accent) 80%, transparent) var(--progress, 0%),
    var(--color-border-subtle) var(--progress, 0%)
  );
}

.audio-player__range::-webkit-slider-thumb {
  appearance: none;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--color-accent);
  border: 2px solid var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

.audio-player__range::-moz-range-thumb {
  width: 12px;
  height: 12px;
  border: 2px solid var(--color-bg-surface);
  border-radius: 50%;
  background: var(--color-accent);
  box-shadow: var(--shadow-sm);
}

.audio-player__range:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px var(--color-focus-ring);
}
</style>
