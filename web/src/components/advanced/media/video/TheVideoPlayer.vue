<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div v-if="items.length" class="video-media">
    <div
      v-for="(item, idx) in items"
      :key="item.id || idx"
      class="video-media__poster"
      :style="aspectStyle(item, idx)"
      @mouseenter="onEnter(idx, $event)"
      @mousemove="onMove(idx, $event)"
      @mouseleave="onLeave(idx, $event)"
    >
      <span
        class="video-media__skeleton"
        :class="{ 'is-loaded': isLoaded(idx) }"
        aria-hidden="true"
      >
        <span class="video-media__spinner">
          <i v-for="seg in spinnerSegments" :key="seg" :style="{ '--seg': seg - 1 }"></i>
        </span>
      </span>
      <video
        class="video-media__frame"
        :class="{ 'is-loaded': isLoaded(idx) }"
        :src="posterSrc(item.src)"
        :aria-label="t('videoPlayer.play')"
        preload="metadata"
        muted
        playsinline
        tabindex="-1"
        @loadedmetadata="onMeta(idx, $event)"
        @loadeddata="onLoaded(idx)"
        @click="onClickVideo(idx, $event)"
        @touchstart.passive="onTouchStart(idx, $event)"
        @touchmove.passive="onTouchMove(idx, $event)"
        @touchend="onTouchEnd(idx, $event)"
        @touchcancel="onTouchEnd(idx, $event)"
        @contextmenu.prevent
        @play="onPlay(idx)"
        @pause="onPause(idx)"
        @ended="onPause(idx)"
        @timeupdate="onTimeUpdate(idx, $event)"
      ></video>
      <span class="video-media__tag" :class="{ 'is-hidden': tagHidden(idx) }">
        <Pause v-if="pausedByIndex[idx]" color="currentColor" />
        <Video v-else color="currentColor" />
        {{ tagText(idx) }}
      </span>
      <button
        type="button"
        class="video-media__fullscreen"
        :class="{ 'is-visible': controlsVisible(idx) }"
        @click.stop="open(idx, $event)"
      >
        <Full color="currentColor" />
        {{ t('videoPlayer.fullscreen') }}
      </button>
    </div>

    <TheVideoLightbox
      :visible="activeIndex !== null"
      :src="activeSrc"
      :start-time="resumeTime"
      @close="close"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getFileUrl, getHubFileUrl } from '@/utils/other'
import { formatMediaTime } from '../shared/time'
import Video from '@/components/icons/video.vue'
import Pause from '@/components/icons/pause.vue'
import Full from '@/components/icons/full.vue'
import TheVideoLightbox from './TheVideoLightbox.vue'

type VideoMeta = { duration: number; width: number; height: number }

const props = withDefaults(
  defineProps<{
    files?: App.Api.Ech0.FileObject[]
    baseUrl?: string
  }>(),
  { files: () => [] },
)

const { t } = useI18n()

const spinnerSegments = Array.from({ length: 12 }, (_, index) => index + 1)

const items = computed(() =>
  (props.files || []).map((file) => ({
    id: file.id,
    src: props.baseUrl ? getHubFileUrl(file, props.baseUrl) : getFileUrl(file),
    width: file.width,
    height: file.height,
  })),
)

const metaByIndex = ref<Record<number, VideoMeta>>({})

const loadedByIndex = ref<Record<number, boolean>>({})

function onLoaded(idx: number) {
  if (loadedByIndex.value[idx]) return
  loadedByIndex.value = { ...loadedByIndex.value, [idx]: true }
}

function isLoaded(idx: number) {
  return Boolean(loadedByIndex.value[idx])
}

const playingByIndex = ref<Record<number, boolean>>({})
const currentTimeByIndex = ref<Record<number, number>>({})
const pausedByIndex = ref<Record<number, boolean>>({})

const controlsShownByIndex = ref<Record<number, boolean>>({})

const CONTROLS_IDLE_HIDE = 2000
const hideTimerByIndex: Record<number, ReturnType<typeof setTimeout>> = {}

const MOVE_ACTIVATE_DISTANCE = 8
let lastPointer: { x: number; y: number } | null = null

const activeIndex = ref<number | null>(null)
const activeSrc = computed(() =>
  activeIndex.value !== null ? (items.value[activeIndex.value]?.src ?? '') : '',
)
const resumeTime = ref(0)

function posterSrc(src: string) {
  return src ? `${src}#t=0.1` : ''
}

function onMeta(idx: number, event: Event) {
  const el = event.target as HTMLVideoElement
  metaByIndex.value = {
    ...metaByIndex.value,
    [idx]: {
      duration: Number.isFinite(el.duration) ? el.duration : 0,
      width: el.videoWidth,
      height: el.videoHeight,
    },
  }
}

function prefersReducedMotion() {
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
}

const HOVER_PLAY_DELAY = 300
let hoverTimer: ReturnType<typeof setTimeout> | null = null

function clearHoverTimer() {
  if (hoverTimer !== null) {
    clearTimeout(hoverTimer)
    hoverTimer = null
  }
}

function clearHideTimer(idx: number) {
  const timer = hideTimerByIndex[idx]
  if (timer) {
    clearTimeout(timer)
    delete hideTimerByIndex[idx]
  }
}

function showControls(idx: number) {
  if (!controlsShownByIndex.value[idx]) controlsShownByIndex.value[idx] = true
  clearHideTimer(idx)
  if (playingByIndex.value[idx]) {
    hideTimerByIndex[idx] = setTimeout(() => {
      delete hideTimerByIndex[idx]
      controlsShownByIndex.value[idx] = false
    }, CONTROLS_IDLE_HIDE)
  }
}

function hideControls(idx: number) {
  clearHideTimer(idx)
  controlsShownByIndex.value[idx] = false
}

function pointerMovedEnough(event: MouseEvent) {
  const point = { x: event.clientX, y: event.clientY }
  if (!lastPointer) {
    lastPointer = point
    return true
  }
  if (Math.hypot(point.x - lastPointer.x, point.y - lastPointer.y) < MOVE_ACTIVATE_DISTANCE) {
    return false
  }
  lastPointer = point
  return true
}

const LONG_PRESS_DELAY = 350
const LONG_PRESS_MOVE_TOLERANCE = 10
let longPressTimer: ReturnType<typeof setTimeout> | null = null
let longPressStart: { x: number; y: number } | null = null
const longPressActiveByIndex: Record<number, boolean> = {}
const suppressClickByIndex: Record<number, boolean> = {}
let lastTouchAt = 0

function clearLongPressTimer() {
  if (longPressTimer !== null) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
}

function onTouchStart(idx: number, event: TouchEvent) {
  lastTouchAt = Date.now()
  suppressClickByIndex[idx] = false
  const touch = event.touches[0]
  longPressStart = touch ? { x: touch.clientX, y: touch.clientY } : null
  const video = event.currentTarget as HTMLVideoElement
  clearLongPressTimer()
  longPressTimer = setTimeout(() => {
    longPressTimer = null
    if (!video.paused) return
    video.muted = true
    longPressActiveByIndex[idx] = true
    showControls(idx)
    void video.play().catch(() => {})
  }, LONG_PRESS_DELAY)
}

function onTouchMove(idx: number, event: TouchEvent) {
  if (longPressTimer === null || !longPressStart) return
  const touch = event.touches[0]
  if (!touch) return
  const moved = Math.hypot(touch.clientX - longPressStart.x, touch.clientY - longPressStart.y)
  if (moved > LONG_PRESS_MOVE_TOLERANCE) clearLongPressTimer()
}

function onTouchEnd(idx: number, event: TouchEvent) {
  lastTouchAt = Date.now()
  clearLongPressTimer()
  longPressStart = null
  if (!longPressActiveByIndex[idx]) return
  longPressActiveByIndex[idx] = false
  suppressClickByIndex[idx] = true
  const video = event.currentTarget as HTMLVideoElement
  video.pause()
  video.currentTime = 0
  video.muted = true
  hideControls(idx)
}

function onEnter(idx: number, event: MouseEvent) {
  lastPointer = { x: event.clientX, y: event.clientY }
  showControls(idx)
  if (Date.now() - lastTouchAt < 800) return
  if (prefersReducedMotion()) return
  if (pausedByIndex.value[idx]) return
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video || !video.paused) return
  clearHoverTimer()
  hoverTimer = setTimeout(() => {
    hoverTimer = null
    video.muted = true
    void video.play().catch(() => {})
  }, HOVER_PLAY_DELAY)
}

function onMove(idx: number, event: MouseEvent) {
  if (pointerMovedEnough(event)) showControls(idx)
}

function onLeave(idx: number, event: MouseEvent) {
  lastPointer = null
  hideControls(idx)
  clearHoverTimer()
  const video = (event.currentTarget as HTMLElement).querySelector('video')
  if (!video) return
  if (!video.muted) return
  video.pause()
  video.currentTime = 0
  video.muted = true
}

function onClickVideo(idx: number, event: MouseEvent) {
  if (suppressClickByIndex[idx]) {
    suppressClickByIndex[idx] = false
    return
  }
  clearHoverTimer()
  const video = event.currentTarget as HTMLVideoElement
  if (video.paused || video.muted) {
    pausedByIndex.value[idx] = false
    video.muted = false
    void video.play().catch(() => {})
  } else {
    video.pause()
    pausedByIndex.value[idx] = true
  }
  showControls(idx)
}

function onPlay(idx: number) {
  playingByIndex.value[idx] = true
  pausedByIndex.value[idx] = false
  showControls(idx)
}

function onPause(idx: number) {
  playingByIndex.value[idx] = false
  clearHideTimer(idx)
}

function onTimeUpdate(idx: number, event: Event) {
  currentTimeByIndex.value[idx] = (event.target as HTMLVideoElement).currentTime
}

onBeforeUnmount(() => {
  clearHoverTimer()
  clearLongPressTimer()
  Object.values(hideTimerByIndex).forEach((timer) => clearTimeout(timer))
})

function aspectStyle(item: { width?: number; height?: number }, idx: number) {
  const meta = metaByIndex.value[idx]
  const width = item.width || meta?.width
  const height = item.height || meta?.height
  return { aspectRatio: width && height ? `${width} / ${height}` : '16 / 9' }
}

function tagText(idx: number) {
  if (pausedByIndex.value[idx]) return t('videoPlayer.paused')
  const duration = metaByIndex.value[idx]?.duration
  if (playingByIndex.value[idx] && duration) {
    const remaining = Math.max(0, duration - (currentTimeByIndex.value[idx] ?? 0))
    return `-${formatMediaTime(remaining)}`
  }
  return t('videoPlayer.label')
}

function tagHidden(idx: number) {
  return Boolean(playingByIndex.value[idx]) && !controlsShownByIndex.value[idx]
}

function controlsVisible(idx: number) {
  return Boolean(controlsShownByIndex.value[idx]) || Boolean(pausedByIndex.value[idx])
}

function open(idx: number, event?: MouseEvent) {
  clearHoverTimer()
  const trigger = event?.currentTarget as HTMLElement | undefined
  const video = trigger?.closest('.video-media__poster')?.querySelector('video')
  resumeTime.value = video ? video.currentTime : 0
  if (video) {
    video.pause()
    video.currentTime = 0
    video.muted = true
  }
  pausedByIndex.value[idx] = false
  hideControls(idx)
  activeIndex.value = idx
}

function close() {
  activeIndex.value = null
}
</script>

<style scoped>
.video-media {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.video-media__poster {
  position: relative;
  display: block;
  width: 100%;
  max-height: 70vh;
  padding: 0;
  overflow: hidden;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: #000;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

.video-media__poster:hover,
.video-media__poster:focus-within {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-soft);
}

.video-media__frame {
  position: relative;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: pointer;
  opacity: 0;

  -webkit-touch-callout: none;
  user-select: none;
  transition: opacity 0.3s ease;
}

.video-media__frame.is-loaded {
  opacity: 1;
}

.video-media__skeleton {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  background:
    linear-gradient(rgb(0 0 0 / 66%), rgb(0 0 0 / 76%)),
    radial-gradient(
      130% 130% at 50% 25%,
      var(--color-accent-soft) 0%,
      var(--color-bg-muted) 55%,
      var(--color-bg-canvas) 100%
    );
}

.video-media__skeleton.is-loaded {
  display: none;
}

.video-media__spinner {
  position: relative;
  width: 26px;
  height: 26px;
  color: rgb(255 255 255 / 82%);
}

.video-media__spinner i {
  position: absolute;
  top: 0;
  left: 50%;
  width: 2.4px;
  height: 6.4px;
  border-radius: 999px;
  background: currentColor;
  transform-origin: center 13px;
  transform: translateX(-50%) rotate(calc(var(--seg) * 30deg));
  animation: video-spinner-fade 1.1s linear infinite;
  animation-delay: calc(var(--seg) * -0.0916s);
}

@keyframes video-spinner-fade {
  0% {
    opacity: 1;
  }

  100% {
    opacity: 0.12;
  }
}

.video-media__tag {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.22rem;
  padding: 0.14rem 0.45rem;
  font-size: 0.72rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
  line-height: 1.4;
  color: #fff;
  background: rgb(0 0 0 / 42%);
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(6px);
  pointer-events: none;

  opacity: 1;
  transition: opacity 0.2s ease;
}

.video-media__tag.is-hidden {
  opacity: 0;
}

.video-media__poster:focus-within .video-media__tag.is-hidden {
  opacity: 1;
}

.video-media__tag svg {
  width: 0.9rem;
  height: 0.9rem;
}

.video-media__fullscreen {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.22rem;
  padding: 0.14rem 0.45rem;
  font-family: inherit;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  line-height: 1.4;
  color: #fff;
  background: rgb(0 0 0 / 42%);
  border: 1px solid rgb(255 255 255 / 20%);
  border-radius: var(--radius-sm);
  backdrop-filter: blur(6px);
  cursor: pointer;

  opacity: 0;
  transition:
    opacity 0.2s ease,
    background 0.15s ease,
    transform 0.15s ease;
}

.video-media__fullscreen:hover {
  background: rgb(0 0 0 / 58%);
}

.video-media__fullscreen:active {
  transform: scale(0.96);
}

.video-media__fullscreen:focus-visible {
  outline: none;
  opacity: 1;
  box-shadow: 0 0 0 3px var(--color-focus-ring);
}

.video-media__fullscreen svg {
  width: 0.9rem;
  height: 0.9rem;
}

.video-media__poster:focus-within .video-media__fullscreen,
.video-media__fullscreen.is-visible {
  opacity: 1;
}

@media (prefers-reduced-motion: reduce) {
  .video-media__poster,
  .video-media__frame,
  .video-media__fullscreen,
  .video-media__tag {
    transition: none;
  }

  .video-media__spinner i {
    animation: none;
    opacity: 0.5;
  }
}
</style>
