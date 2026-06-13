<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ExtensionCardShell
    v-if="musicInfo && musicInfo.server === MusicProvider.APPLE && musicInfo.id"
    size="wide"
    :header-label="t('extensionCard.music')"
  >
    <template #header-icon><Music /></template>
    <div class="music-player-wrap apple-player-wrap">
      <ExtensionCardSkeleton v-if="!isAppleReady" :min-height="175" />
      <iframe
        allow="autoplay *; encrypted-media *; fullscreen *; clipboard-write"
        frameborder="0"
        height="175"
        class="apple-frame"
        :class="{ 'is-ready': isAppleReady }"
        sandbox="allow-forms allow-popups allow-same-origin allow-scripts allow-storage-access-by-user-activation allow-top-navigation-by-user-activation"
        :src="`https://embed.music.apple.com/cn/${musicInfo.type}/${musicInfo.id}`"
        @load="isAppleReady = true"
      >
      </iframe>
    </div>
  </ExtensionCardShell>

  <ExtensionCardShell
    v-else-if="musicInfo && musicInfo.server !== MusicProvider.APPLE && !loadFailed"
    size="wide"
    padding="compact"
    :header-label="t('extensionCard.music')"
  >
    <template #header-icon><Music /></template>
    <div class="music-player-wrap">
      <ExtensionCardSkeleton v-if="isLoading || !track" :min-height="76" />
      <section
        v-else
        class="music-player"
        :class="{ 'is-playing': isPlaybackActive }"
        :aria-label="track.name || t('extensionCard.musicPlayer')"
        @mouseenter="isHovered = true"
        @mouseleave="isHovered = false"
      >
        <audio
          ref="audioRef"
          :src="track.url"
          preload="metadata"
          @canplay="isAudioLoading = false"
          @durationchange="syncAudioState"
          @ended="handleEnded"
          @error="handleAudioError"
          @loadedmetadata="handleLoadedMetadata"
          @loadstart="isAudioLoading = isPlaying"
          @pause="isPlaying = false"
          @playing="handlePlaying"
          @ratechange="updateMediaPositionState"
          @timeupdate="syncAudioState"
        ></audio>

        <div class="cover-wrap">
          <img
            v-if="track.cover"
            class="cover"
            :src="track.cover"
            :alt="t('extensionCard.musicCoverAlt', { title: track.name })"
            width="52"
            height="52"
            loading="lazy"
            referrerpolicy="no-referrer"
          />
          <Music v-else class="cover-placeholder" />
        </div>

        <div class="player-main">
          <div class="display" aria-live="polite">
            <Transition name="display-swap" mode="out-in">
              <div v-if="displayTrackInfo" key="track" class="track-info">
                <strong :title="track.name">{{ track.name || t('extensionCard.music') }}</strong>
                <span :title="track.artist">{{
                  track.artist || t('extensionCard.musicArtistUnknown')
                }}</span>
              </div>
              <div v-else key="lyric" class="lyric-window">
                <Transition name="lyric-scroll" mode="out-in">
                  <p
                    v-if="currentLyric"
                    :key="currentLyric.time"
                    class="lyric"
                    :title="currentLyric.text"
                  >
                    <template v-if="currentLyric.words.length">
                      <span
                        v-for="word in currentLyric.words"
                        :key="`${word.time}-${word.text}`"
                        :class="{ active: word.time <= currentTime }"
                        >{{ word.text }}</span
                      >
                      <span v-if="currentLyricTranslation" class="lyric-translation">
                        {{ currentLyricTranslation }}
                      </span>
                    </template>
                    <template v-else>{{ currentLyric.text }}</template>
                  </p>
                </Transition>
              </div>
            </Transition>
          </div>

          <div class="timeline">
            <span>{{ formatTime(currentTime) }}</span>
            <input
              class="progress"
              type="range"
              min="0"
              :max="duration || 0"
              step="0.1"
              :value="currentTime"
              :disabled="duration <= 0"
              :aria-label="t('extensionCard.musicProgress')"
              :style="{ '--music-progress': `${progressPercent}%` }"
              @input="handleSeek"
            />
            <span>{{ formatTime(duration) }}</span>
          </div>
        </div>

        <button
          type="button"
          class="play-button"
          :aria-label="
            isPlaybackActive ? t('extensionCard.musicPause') : t('extensionCard.musicPlay')
          "
          :title="isPlaybackActive ? t('extensionCard.musicPause') : t('extensionCard.musicPlay')"
          :disabled="audioFailed"
          @click="togglePlayback"
        >
          <span v-if="isAudioLoading" class="loading-dot" aria-hidden="true"></span>
          <Pause v-else-if="isPlaying" color="currentColor" />
          <Play v-else color="currentColor" />
        </button>
      </section>
    </div>
  </ExtensionCardShell>

  <div v-else class="extension-card-invalid">
    <Music class="invalid-icon" />
    <div class="invalid-copy">
      <p class="invalid-title">{{ t('extensionCard.musicUnavailable') }}</p>
      <p class="invalid-subtitle">{{ t('extensionCard.musicUnavailableHint') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import Music from '@/components/icons/music.vue'
import Pause from '@/components/icons/pause.vue'
import Play from '@/components/icons/play.vue'
import { ExtensionType, MusicProvider } from '@/enums/enums'
import { useSettingStore } from '@/stores'
import { parseMusicURL } from '@/utils/other'
import {
  normalizeMetingTrack,
  parseLyrics,
  type LyricLine,
  type MusicTrack,
} from '@/utils/musicPlayer'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'
import ExtensionCardSkeleton from '../shared/ExtensionCardSkeleton.vue'

const TRACK_INFO_DURATION = 4000
const PLAYER_EVENT = 'ech0:music-play'
const DEFAULT_VOLUME = 0.7

const { SystemSetting, loading: settingsLoading } = storeToRefs(useSettingStore())
const { t } = useI18n()

const props = defineProps<{
  echo: {
    extension?: App.Api.Ech0.EchoExtension | null
  }
}>()

const playerToken = {}
const audioRef = ref<HTMLAudioElement | null>(null)
const track = ref<MusicTrack | null>(null)
const lyrics = ref<LyricLine[]>([])
const currentTime = ref(0)
const duration = ref(0)
const isLoading = ref(false)
const isAudioLoading = ref(false)
const isPlaying = ref(false)
const isHovered = ref(false)
const showTrackInfo = ref(true)
const loadFailed = ref(false)
const audioFailed = ref(false)
const isAppleReady = ref(false)
let requestController: AbortController | null = null
let lyricController: AbortController | null = null
let trackInfoTimer: ReturnType<typeof setTimeout> | null = null

const musicURL = computed(() => {
  if (props.echo.extension?.type !== ExtensionType.MUSIC) return ''
  return props.echo.extension.payload.url
})

const musicInfo = computed(() => parseMusicURL(musicURL.value))

const metingAPI = computed(() => {
  if (!settingsLoading.value && SystemSetting.value?.meting_api?.length) {
    return SystemSetting.value.meting_api
  }
  return 'https://meting.soopy.cn/api'
})

const musicSourceKey = computed(() => {
  const info = musicInfo.value
  return `${info?.server ?? ''}|${info?.type ?? ''}|${info?.id ?? ''}|${metingAPI.value}`
})

const currentLyric = computed(() => {
  for (let index = lyrics.value.length - 1; index >= 0; index -= 1) {
    if (lyrics.value[index]!.time <= currentTime.value) return lyrics.value[index]
  }
  return null
})
const currentLyricTranslation = computed(() => {
  const text = currentLyric.value?.text
  if (!text?.includes('\n')) return ''
  return text.split('\n').slice(1).join('\n')
})

const displayTrackInfo = computed(
  () =>
    !isPlaying.value ||
    isHovered.value ||
    showTrackInfo.value ||
    !lyrics.value.length ||
    !currentLyric.value,
)
const isPlaybackActive = computed(() => isPlaying.value || isAudioLoading.value)
const progressPercent = computed(() =>
  duration.value > 0 ? Math.min(100, (currentTime.value / duration.value) * 100) : 0,
)

function buildMetingURL() {
  const info = musicInfo.value
  if (!info || info.server === MusicProvider.APPLE) return ''

  const params: Record<string, string> = {
    server: info.server,
    type: info.type,
    id: info.id,
    auth: '',
    r: String(Math.random()),
  }
  const base = metingAPI.value.trim()

  if (/:(server|type|id|auth|r)\b/.test(base)) {
    return Object.entries(params).reduce(
      (url, [key, value]) => url.replaceAll(`:${key}`, encodeURIComponent(value)),
      base,
    )
  }

  const separator = base.includes('?') ? '&' : '?'
  return `${base}${separator}${new URLSearchParams(params).toString()}`
}

async function loadTrack() {
  requestController?.abort()
  lyricController?.abort()
  pause()
  track.value = null
  lyrics.value = []
  currentTime.value = 0
  duration.value = 0
  loadFailed.value = false
  audioFailed.value = false
  isAppleReady.value = false

  const info = musicInfo.value
  if (!info || info.server === MusicProvider.APPLE) return

  const controller = new AbortController()
  requestController = controller
  isLoading.value = true
  try {
    const response = await fetch(buildMetingURL(), {
      headers: { Accept: 'application/json' },
      signal: controller.signal,
    })
    if (!response.ok) throw new Error(`Meting API returned ${response.status}`)

    const payload = await response.json()
    if (requestController !== controller) return

    const resolvedTrack = normalizeMetingTrack(payload)
    if (!resolvedTrack) throw new Error('Meting API returned no playable track')
    track.value = resolvedTrack
    await nextTick()
    applyVolume()
    revealTrackInfo()
    void loadLyrics(resolvedTrack.lrc)
  } catch (error) {
    if (!isAbortError(error)) loadFailed.value = true
  } finally {
    if (requestController === controller) {
      requestController = null
      isLoading.value = false
    }
  }
}

async function loadLyrics(url: string) {
  lyricController?.abort()
  lyrics.value = []
  if (!url) return

  const controller = new AbortController()
  lyricController = controller
  try {
    const response = await fetch(url, { signal: controller.signal })
    if (!response.ok) return
    const source = await response.text()
    if (lyricController !== controller) return
    lyrics.value = parseLyrics(source)
  } catch (error) {
    if (!isAbortError(error)) lyrics.value = []
  } finally {
    if (lyricController === controller) lyricController = null
  }
}

function isAbortError(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const value = error as { name?: string; cause?: unknown }
  return value.name === 'AbortError' || isAbortError(value.cause)
}

function applyVolume() {
  if (audioRef.value) audioRef.value.volume = DEFAULT_VOLUME
}

async function play() {
  const audio = audioRef.value
  if (!audio || !track.value || audioFailed.value) return

  window.dispatchEvent(new CustomEvent(PLAYER_EVENT, { detail: playerToken }))
  isAudioLoading.value = true
  try {
    await audio.play()
  } catch (error) {
    isAudioLoading.value = false
    if (
      !(error instanceof DOMException && ['AbortError', 'NotAllowedError'].includes(error.name))
    ) {
      audioFailed.value = true
    }
  }
}

function pause() {
  audioRef.value?.pause()
  isAudioLoading.value = false
  isPlaying.value = false
}

function togglePlayback() {
  if (isPlaybackActive.value) pause()
  else void play()
}

function handlePlaying() {
  isAudioLoading.value = false
  isPlaying.value = true
  revealTrackInfo()
  setupMediaSession()
  updateMediaPlaybackState()
}

function handleEnded() {
  pause()
  currentTime.value = 0
  if (audioRef.value) audioRef.value.currentTime = 0
  updateMediaPlaybackState()
  updateMediaPositionState()
}

function handleAudioError() {
  pause()
  audioFailed.value = true
}

function handleLoadedMetadata() {
  applyVolume()
  syncAudioState()
}

function syncAudioState() {
  const audio = audioRef.value
  if (!audio) return
  currentTime.value = Number.isFinite(audio.currentTime) ? audio.currentTime : 0
  duration.value = Number.isFinite(audio.duration) ? audio.duration : 0
  updateMediaPositionState()
}

function handleSeek(event: Event) {
  const audio = audioRef.value
  if (!audio) return
  const position = Number((event.currentTarget as HTMLInputElement).value)
  if (!Number.isFinite(position)) return
  audio.currentTime = position
  currentTime.value = position
  updateMediaPositionState()
}

function revealTrackInfo() {
  showTrackInfo.value = true
  if (trackInfoTimer) clearTimeout(trackInfoTimer)
  trackInfoTimer = setTimeout(() => {
    showTrackInfo.value = false
  }, TRACK_INFO_DURATION)
}

function formatTime(value: number) {
  if (!Number.isFinite(value) || value < 0) return '00:00'
  const minutes = Math.floor(value / 60)
  const seconds = Math.floor(value % 60)
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

function setupMediaSession() {
  if (!('mediaSession' in navigator) || !('MediaMetadata' in window) || !track.value) return

  navigator.mediaSession.metadata = new MediaMetadata({
    title: track.value.name,
    artist: track.value.artist,
    artwork: track.value.cover ? [{ src: track.value.cover }] : [],
  })

  const handlers: Partial<Record<MediaSessionAction, MediaSessionActionHandler | null>> = {
    play: () => void play(),
    pause,
    stop: () => {
      pause()
      if (audioRef.value) audioRef.value.currentTime = 0
      syncAudioState()
    },
    seekbackward: (details) => seekBy(-(details.seekOffset ?? 10)),
    seekforward: (details) => seekBy(details.seekOffset ?? 10),
    seekto: (details) => {
      if (details.seekTime !== undefined) seekTo(details.seekTime, details.fastSeek)
    },
  }

  for (const [action, handler] of Object.entries(handlers)) {
    try {
      navigator.mediaSession.setActionHandler(action as MediaSessionAction, handler)
    } catch {
      // Browsers may expose only part of the Media Session API.
    }
  }
}

function seekBy(offset: number) {
  seekTo(currentTime.value + offset)
}

function seekTo(value: number, fastSeek = false) {
  const audio = audioRef.value
  if (!audio || duration.value <= 0) return
  const position = Math.min(Math.max(value, 0), duration.value)
  if (fastSeek && 'fastSeek' in audio) audio.fastSeek(position)
  else audio.currentTime = position
  currentTime.value = position
  updateMediaPositionState()
}

function updateMediaPlaybackState() {
  if ('mediaSession' in navigator) {
    navigator.mediaSession.playbackState = isPlaying.value ? 'playing' : 'paused'
  }
}

function updateMediaPositionState() {
  if (!('mediaSession' in navigator) || duration.value <= 0) return
  try {
    navigator.mediaSession.setPositionState({
      duration: duration.value,
      playbackRate: audioRef.value?.playbackRate ?? 1,
      position: Math.min(currentTime.value, duration.value),
    })
  } catch {
    // Metadata and duration can briefly be out of sync while audio loads.
  }
}

function handleOtherPlayer(event: Event) {
  if ((event as CustomEvent<object>).detail !== playerToken) pause()
}

watch(musicSourceKey, loadTrack, { immediate: true })
watch(isPlaying, updateMediaPlaybackState)

onMounted(() => {
  window.addEventListener(PLAYER_EVENT, handleOtherPlayer)
})

onBeforeUnmount(() => {
  requestController?.abort()
  lyricController?.abort()
  if (trackInfoTimer) clearTimeout(trackInfoTimer)
  window.removeEventListener(PLAYER_EVENT, handleOtherPlayer)
  pause()
})
</script>

<style scoped>
.music-player-wrap {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 76px;
}

.music-player-wrap > * {
  width: 100%;
}

.apple-player-wrap {
  min-height: 175px;
}

.apple-player-wrap > :deep(.extension-skeleton) {
  position: absolute;
  inset: 0;
  z-index: 1;
}

.apple-frame {
  display: block;
  width: 100%;
  min-height: 175px;
  overflow: hidden;
  border-radius: inherit;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.apple-frame.is-ready {
  opacity: 1;
}

.music-player {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
  padding: 0.55rem;
  border-radius: calc(var(--radius-md) - 0.15rem);
  background: var(--color-bg-surface);
}

.music-player audio {
  display: none;
}

.cover-wrap {
  display: grid;
  flex: 0 0 3.25rem;
  width: 3.25rem;
  height: 3.25rem;
  place-items: center;
  overflow: hidden;
  border-radius: calc(var(--radius-md) - 0.2rem);
  background: var(--color-bg-muted);
  color: var(--color-text-muted);
}

.cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 1.45rem;
  height: 1.45rem;
  opacity: 0.75;
}

.player-main {
  flex: 1 1 auto;
  min-width: 0;
}

.display {
  min-height: 2.35rem;
  overflow: hidden;
  line-height: 1.25;
}

.track-info {
  display: grid;
  min-width: 0;
}

.track-info > strong,
.track-info > span {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.track-info > strong {
  font-size: 0.9rem;
  font-weight: 650;
  color: var(--color-text-primary);
}

.track-info > span {
  margin-top: 0.1rem;
  font-size: 0.72rem;
  color: var(--color-text-muted);
}

.lyric-window {
  max-height: 2.35rem;
  overflow: hidden;
}

.lyric {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  font-size: 0.8rem;
  font-weight: 600;
  line-height: 1.4;
  color: var(--color-text-secondary);
  white-space: pre-line;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.lyric > span {
  transition:
    color 0.15s ease,
    opacity 0.15s ease;
}

.lyric > span:not(.active, .lyric-translation) {
  opacity: 0.42;
}

.lyric > span.active {
  color: var(--color-accent);
}

.lyric > .lyric-translation {
  display: block;
  overflow: hidden;
  font-size: 0.72rem;
  font-weight: 500;
  color: var(--color-text-muted);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.timeline {
  display: grid;
  grid-template-columns: auto minmax(3rem, 1fr) auto;
  align-items: center;
  gap: 0.45rem;
  color: var(--color-text-muted);
  font-size: 0.65rem;
  font-variant-numeric: tabular-nums;
}

.progress {
  width: 100%;
  height: 0.9rem;
  margin: 0;
  appearance: none;
  cursor: pointer;
  background: transparent;
}

.progress::-webkit-slider-runnable-track {
  height: 0.2rem;
  border-radius: 999px;
  background: linear-gradient(
    to right,
    var(--color-accent) var(--music-progress),
    var(--color-border-subtle) var(--music-progress)
  );
}

.progress::-webkit-slider-thumb {
  width: 0.7rem;
  height: 0.7rem;
  margin-top: -0.25rem;
  appearance: none;
  border: 2px solid var(--color-bg-surface);
  border-radius: 50%;
  background: var(--color-accent);
  box-shadow: var(--shadow-sm);
}

.progress::-moz-range-track {
  height: 0.2rem;
  border-radius: 999px;
  background: var(--color-border-subtle);
}

.progress::-moz-range-progress {
  height: 0.2rem;
  border-radius: 999px;
  background: var(--color-accent);
}

.progress::-moz-range-thumb {
  width: 0.7rem;
  height: 0.7rem;
  border: 2px solid var(--color-bg-surface);
  border-radius: 50%;
  background: var(--color-accent);
  box-shadow: var(--shadow-sm);
}

.progress:disabled {
  cursor: default;
  opacity: 0.55;
}

.play-button {
  display: grid;
  flex: 0 0 2.25rem;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0;
  place-items: center;
  border: 1px solid var(--color-border-subtle);
  border-radius: 50%;
  background: var(--color-bg-muted);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition:
    color 0.15s ease,
    background-color 0.15s ease,
    border-color 0.15s ease,
    transform 0.1s ease;
}

.play-button:hover:not(:disabled) {
  color: var(--color-accent);
  border-color: color-mix(in srgb, var(--color-accent) 45%, var(--color-border-subtle));
  background: var(--color-accent-soft);
}

.play-button:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

.play-button:active:not(:disabled) {
  transform: scale(0.94);
}

.play-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.play-button :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.loading-dot {
  width: 0.95rem;
  height: 0.95rem;
  border: 2px solid var(--color-border-subtle);
  border-top-color: var(--color-accent);
  border-radius: 50%;
  animation: music-loading 0.8s linear infinite;
}

.extension-card-invalid {
  display: flex;
  align-items: center;
  width: 100%;
  gap: 0.6rem;
  padding: 0.75rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
  color: var(--color-text-muted);
}

.invalid-icon {
  flex: 0 0 1rem;
  width: 1rem;
  height: 1rem;
}

.invalid-copy {
  min-width: 0;
}

.invalid-title {
  margin: 0;
  font-size: 0.86rem;
  line-height: 1.35;
  color: var(--color-text-secondary);
}

.invalid-subtitle {
  margin: 0.12rem 0 0;
  font-size: 0.74rem;
  line-height: 1.35;
}

.display-swap-enter-active,
.display-swap-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.display-swap-enter-from {
  opacity: 0;
  transform: translateY(0.3rem);
}

.display-swap-leave-to {
  opacity: 0;
  transform: translateY(-0.3rem);
}

.lyric-scroll-enter-active,
.lyric-scroll-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.lyric-scroll-enter-from {
  opacity: 0;
  transform: translateY(100%);
}

.lyric-scroll-leave-to {
  opacity: 0;
  transform: translateY(-100%);
}

@keyframes music-loading {
  to {
    transform: rotate(360deg);
  }
}

@media (width <= 420px) {
  .music-player {
    gap: 0.55rem;
  }

  .cover-wrap {
    flex-basis: 2.8rem;
    width: 2.8rem;
    height: 2.8rem;
  }

  .timeline {
    gap: 0.3rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .apple-frame,
  .play-button,
  .display-swap-enter-active,
  .display-swap-leave-active,
  .lyric-scroll-enter-active,
  .lyric-scroll-leave-active {
    transition: none;
  }

  .loading-dot {
    animation: none;
  }
}
</style>
