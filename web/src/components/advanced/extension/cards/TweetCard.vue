<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ExtensionCardShell :header-label="t('extensionCard.tweet')">
    <template #header-icon><XBrand /></template>
    <template #header-actions>
      <a
        :href="payload.url"
        target="_blank"
        rel="noopener noreferrer"
        class="tweet-card__jump"
        :aria-label="t('extensionCard.jump')"
      >
        <span class="tweet-card__jump-text">{{ t('extensionCard.jump') }}</span>
        <LinkIcon class="tweet-card__jump-icon" />
      </a>
    </template>
    <div ref="rootRef" class="tweet-card">
      <div v-if="state !== 'fallback'" class="tweet-card__stage">
        <ExtensionCardSkeleton
          v-if="state === 'loading'"
          :min-height="120"
          class="tweet-card__loader"
        />
        <blockquote
          ref="quoteRef"
          class="twitter-tweet tweet-card__quote"
          :data-theme="widgetTheme"
          data-dnt="true"
          data-conversation="none"
        >
          <a :href="payload.url"></a>
        </blockquote>
      </div>

      <a
        v-else
        :href="payload.url"
        target="_blank"
        rel="noopener noreferrer"
        class="tweet-card__fallback"
      >
        <span class="tweet-card__fallback-icon">
          <XBrand />
        </span>
        <div class="tweet-card__fallback-meta">
          <span class="tweet-card__fallback-handle">@{{ payload.username }}</span>
          <span class="tweet-card__fallback-action">{{ t('extensionCard.tweetView') }}</span>
        </div>
      </a>
    </div>
  </ExtensionCardShell>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { useThemeStore } from '@/stores/theme'
import { loadExternalScript } from '@/utils/loadExternalAsset'
import XBrand from '@/components/icons/tweet.vue'
import LinkIcon from '@/components/icons/link.vue'
import ExtensionCardShell from '../shared/ExtensionCardShell.vue'
import ExtensionCardSkeleton from '../shared/ExtensionCardSkeleton.vue'

const WIDGETS_SRC = 'https://platform.twitter.com/widgets.js'
const RENDER_TIMEOUT_MS = 6_000

type RenderedEvent = { target: HTMLElement }
type TwitterWidgets = {
  load: (el?: HTMLElement) => void
}
type TwitterEvents = {
  bind: (event: string, handler: (e: RenderedEvent) => void) => void
  unbind: (event: string, handler: (e: RenderedEvent) => void) => void
}
type TwitterGlobal = {
  widgets?: TwitterWidgets
  events?: TwitterEvents
}
type LoadState = 'loading' | 'rendered' | 'fallback'

const { t } = useI18n()
const themeStore = useThemeStore()
const { theme } = storeToRefs(themeStore)

defineProps<{
  payload: { url: string; username: string; statusId: string }
}>()

const rootRef = ref<HTMLElement | null>(null)
const quoteRef = ref<HTMLElement | null>(null)
const state = ref<LoadState>('loading')
const widgetTheme = computed<'dark' | 'light'>(() => (theme.value === 'dark' ? 'dark' : 'light'))

let observer: IntersectionObserver | null = null
let renderTimer: ReturnType<typeof setTimeout> | null = null
let hasTriggered = false

const clearRenderTimer = () => {
  if (renderTimer) {
    clearTimeout(renderTimer)
    renderTimer = null
  }
}

const triggerFallback = () => {
  clearRenderTimer()
  state.value = 'fallback'
}

const handleRendered = (event: RenderedEvent) => {
  if (state.value !== 'loading') return
  if (!rootRef.value?.contains(event.target)) return
  clearRenderTimer()
  state.value = 'rendered'
}

const armRenderTimeout = () => {
  clearRenderTimer()
  renderTimer = setTimeout(() => {
    if (state.value === 'loading' && quoteRef.value?.isConnected) {
      triggerFallback()
    }
  }, RENDER_TIMEOUT_MS)
}

const renderTweet = async () => {
  if (hasTriggered) return
  hasTriggered = true

  try {
    await loadExternalScript(WIDGETS_SRC, { timeoutMs: RENDER_TIMEOUT_MS })
  } catch {
    triggerFallback()
    return
  }

  const twttr = (window as Window & { twttr?: TwitterGlobal }).twttr
  if (!twttr?.widgets || !rootRef.value) {
    triggerFallback()
    return
  }

  twttr.events?.bind('rendered', handleRendered)
  armRenderTimeout()
  try {
    twttr.widgets.load(rootRef.value)
  } catch {
    triggerFallback()
  }
}

onMounted(() => {
  if (!rootRef.value) return

  if (typeof IntersectionObserver === 'undefined') {
    void renderTweet()
    return
  }

  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          observer?.disconnect()
          observer = null
          void renderTweet()
          break
        }
      }
    },
    { rootMargin: '200px 0px' },
  )
  observer.observe(rootRef.value)
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
  clearRenderTimer()
  const twttr = (window as Window & { twttr?: TwitterGlobal }).twttr
  twttr?.events?.unbind('rendered', handleRendered)
})

watch(widgetTheme, () => {
  if (!hasTriggered || state.value === 'fallback') return
  const twttr = (window as Window & { twttr?: TwitterGlobal }).twttr
  if (twttr?.widgets && rootRef.value && quoteRef.value?.isConnected) {
    state.value = 'loading'
    armRenderTimeout()
    try {
      twttr.widgets.load(rootRef.value)
    } catch {
      triggerFallback()
    }
  }
})
</script>

<style scoped>
.tweet-card {
  min-width: 0;
}

.tweet-card__jump {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  font-size: 0.78rem;
  color: var(--color-text-muted);
  text-decoration: none;
  border-radius: var(--radius-sm);
  padding: 0.1rem 0.3rem;
  transition:
    color 0.15s ease,
    background 0.15s ease;
}

.tweet-card__jump:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-surface);
}

.tweet-card__jump:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px var(--color-focus-ring);
}

.tweet-card__jump-text {
  font-weight: 600;
  letter-spacing: 0.01em;
}

.tweet-card__jump-icon {
  width: 0.85rem;
  height: 0.85rem;
}

.tweet-card__stage {
  position: relative;
}

.tweet-card__loader {
  position: absolute;
  inset: 0;
  z-index: 1;
  background: var(--color-bg-surface);
  border-bottom-left-radius: var(--radius-md);
  border-bottom-right-radius: var(--radius-md);
}

.tweet-card__quote {
  margin: 0;
  padding: 0;
}

.tweet-card :deep(.twitter-tweet),
.tweet-card :deep(iframe.twitter-tweet),
.tweet-card :deep(twitter-widget) {
  margin: 0 !important;
  width: 100% !important;
  border-bottom-left-radius: var(--radius-md) !important;
  border-bottom-right-radius: var(--radius-md) !important;
  overflow: hidden !important;
}

.tweet-card__fallback {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  text-decoration: none;
  transition: background 0.15s ease;
}

.tweet-card__fallback:hover {
  background: var(--color-bg-muted);
}

.tweet-card__fallback:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 1px var(--color-focus-ring),
    0 0 0 4px var(--card-focus-ring-outer);
}

.tweet-card__fallback-icon {
  flex-shrink: 0;
  width: 2.1rem;
  height: 2.1rem;
  border-radius: 9999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-primary);
  font-size: 0.85rem;
}

.tweet-card__fallback-meta {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tweet-card__fallback-handle {
  color: var(--color-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
  line-height: 1.3;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tweet-card__fallback-action {
  margin-top: 0.15rem;
  color: var(--color-text-muted);
  font-size: 0.78rem;
  font-family: var(--font-family-mono);
}
</style>
