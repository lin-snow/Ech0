<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="echo-feed mx-auto mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6"
    :class="compact ? 'pl-1 pr-0 max-w-full' : 'px-2 sm:px-4 md:px-6 max-w-full'"
  >
    <TheOnboardingChecklist v-if="showChecklist" />
    <TheFirstEchoCard v-if="showFirstEchoCard" @dismiss="dismissFirstEchoCard" />

    <EchoTimelineList
      v-if="scrollModeReady && echoStore.echoList.length > 0"
      :key="useWindowScroll ? 'window' : 'element'"
      :items="echoStore.echoList"
      :scroll-target="scrollTarget ?? null"
      :use-window="useWindowScroll"
      :busy="echoStore.isLoadingMore"
      :signature="echoStore.loadedSignature"
    />

    <Transition name="fade">
      <div v-if="echoStore.isLoading && echoStore.echoList.length === 0" class="my-5">
        <TheLoadingIndicator class="echos-toolbar" size="lg" :label="t('homeFeed.loading')" />
      </div>
    </Transition>

    <Transition name="fade">
      <div v-if="showEmptyState" class="mx-auto my-5 text-center echos-toolbar">
        <p class="text-xl text-[var(--color-text-muted)]">
          {{ echoStore.isFilteringMode ? t('homeFeed.noMoreFiltered') : t('homeFeed.noMore') }}
        </p>
      </div>
    </Transition>

    <Transition name="fade">
      <div v-if="echoStore.loadError === 'first'" class="echo-feed-status">
        <p>{{ t('homeFeed.loadFailed') }}</p>
        <button type="button" class="echo-feed-retry" @click="echoStore.retryLoad()">
          {{ t('homeFeed.retry') }}
        </button>
      </div>
    </Transition>

    <Transition name="fade">
      <TheLoadingIndicator
        v-if="echoStore.isLoadingMore"
        class="mx-auto my-5 echos-toolbar"
        size="md"
        :label="t('homeFeed.loadingMore')"
      />
    </Transition>

    <div v-if="echoStore.loadError === 'more'" class="echo-feed-status">
      <p>{{ t('homeFeed.loadMoreFailed') }}</p>
      <button type="button" class="echo-feed-retry" @click="echoStore.retryLoad()">
        {{ t('homeFeed.retry') }}
      </button>
    </div>

    <p
      v-if="
        !echoStore.isLoadingMore &&
        echoStore.loadError === null &&
        echoStore.echoList.length > 0 &&
        echoStore.nextCursor === null
      "
      class="echo-feed-end echos-toolbar"
    >
      {{ t('homeFeed.allLoaded') }}
    </p>

    <div ref="sentinel" class="h-px" aria-hidden="true" />
    <p class="feed-sr-only" aria-live="polite">{{ liveMessage }}</p>

    <div
      v-show="showBackTop"
      :style="backTopStyle"
      class="fixed bottom-6 z-50 transition-all duration-500"
    >
      <TheBackTop class="h-8 w-8 p-1" :target="scrollTarget ?? null" />
    </div>

    <div v-if="footerContent" class="mt-6 text-center">
      <a v-if="footerLink" :href="footerLink" target="_blank" rel="noopener noreferrer">
        <span class="text-[var(--color-text-muted)] text-sm">
          {{ footerContent }}
        </span>
      </a>
      <span v-else class="text-[var(--color-text-muted)] text-sm">
        {{ footerContent }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import EchoTimelineList from './EchoTimelineList.vue'
import TheFirstEchoCard from './TheFirstEchoCard.vue'
import TheOnboardingChecklist from './TheOnboardingChecklist.vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useEchoStore, useSettingStore, useUserStore } from '@/stores'
import { FIRST_ECHO_CARD_SEEN_KEY, shouldShowFirstEchoCard } from '@/composables/useSiteOnboarding'
import { useSiteChecklist } from '@/composables/useSiteChecklist'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import {
  HOME_TIMELINE_COUNT_KEY,
  TIMELINE_BASE_PAGE_SIZE,
  TIMELINE_LOAD_AHEAD_PX,
  timelinePageSize,
} from '@/utils/timeline'
import { localStg } from '@/utils/storage'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

const props = defineProps<{
  scrollTarget?: HTMLElement | null
  compact?: boolean
}>()

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()
const { SystemSetting } = storeToRefs(settingStore)
const { isFilteringMode, hasActiveQuery } = storeToRefs(echoStore)
const { isLogin } = storeToRefs(userStore)

const firstEchoSeen = ref(localStg.getItem<boolean>(FIRST_ECHO_CARD_SEEN_KEY) === true)

const { visible: showChecklist } = useSiteChecklist()

const showFirstEchoCard = computed(() => {
  const echo = echoStore.echoList?.[0]
  if (!echo) return false
  return shouldShowFirstEchoCard({
    isOwner: isLogin.value,
    filtering: hasActiveQuery.value,
    echoTotal: echoStore.total,
    firstEchoPrivate: echo.private === true,
    dismissed: firstEchoSeen.value,
  })
})

const showEmptyState = computed(
  () =>
    !showChecklist.value &&
    !echoStore.isLoading &&
    echoStore.loadError !== 'first' &&
    echoStore.total === 0,
)

const dismissFirstEchoCard = () => {
  firstEchoSeen.value = true
  localStg.setItem(FIRST_ECHO_CARD_SEEN_KEY, true)
}
const footerContent = computed(
  () => SystemSetting.value.footer_content || SystemSetting.value.ICP_number,
)
const footerLink = computed(() => SystemSetting.value.footer_link)

const liveMessage = computed(() => {
  if (echoStore.isLoadingMore) return t('homeFeed.loadingMore')
  if (echoStore.loadError === 'first') return t('homeFeed.loadFailed')
  if (echoStore.loadError === 'more') return t('homeFeed.loadMoreFailed')
  if (echoStore.echoList.length > 0 && echoStore.nextCursor === null) {
    return t('homeFeed.allLoaded')
  }
  return ''
})

const useWindowScroll = ref(true)
const scrollModeReady = ref(false)
const syncScrollMode = () => {
  const target = props.scrollTarget
  const overflow = target ? getComputedStyle(target).overflowY : ''
  useWindowScroll.value =
    !target || (overflow !== 'auto' && overflow !== 'scroll' && overflow !== 'overlay')
}

// Floating "back to top" control: the timeline only scrolls the window on the
// mobile layout, the inner column on desktop, so it follows that scroller and
// parks itself in the gutter right of the column (same as the hub feed).
const BACK_TOP_OFFSET = 300
const showBackTop = ref(false)
const backTopStyle = ref<Record<string, string>>({ right: '1.5rem' })

const readFeedScroll = () =>
  useWindowScroll.value ? window.scrollY : (props.scrollTarget?.scrollTop ?? 0)

const updateBackTop = () => {
  showBackTop.value = readFeedScroll() > BACK_TOP_OFFSET
  const column = props.scrollTarget
  if (useWindowScroll.value || !column) {
    backTopStyle.value = { right: '1.5rem' }
    return
  }
  const rect = column.getBoundingClientRect()
  backTopStyle.value = {
    left: `${Math.min(window.innerWidth - 56, Math.round(rect.right + 24))}px`,
  }
}

let backTopScroller: HTMLElement | Window | null = null
let backTopColumnObserver: ResizeObserver | null = null
const observeBackTopColumn = () => {
  backTopColumnObserver?.disconnect()
  backTopColumnObserver = null
  const column = props.scrollTarget
  if (!column || typeof ResizeObserver === 'undefined') return
  // The breakpoint moves the column and the scrollbar width changes with it;
  // a resize event alone reads the pre-layout rect on some engines.
  backTopColumnObserver = new ResizeObserver(() => updateBackTop())
  backTopColumnObserver.observe(column)
}

// The home shell animates in over a few frames (deferred chunks, the rail
// claiming its final width), so a single measurement at mount can be off by
// the distance the column still travels. Re-measure for a short window
// instead of waiting for the first user scroll to correct it.
const BACK_TOP_SETTLE_MS = 800
let settleFrame: number | null = null
let settleUntil = 0
const settleBackTop = () => {
  settleUntil = performance.now() + BACK_TOP_SETTLE_MS
  if (settleFrame !== null) return
  const step = () => {
    updateBackTop()
    settleFrame = performance.now() < settleUntil ? requestAnimationFrame(step) : null
  }
  settleFrame = requestAnimationFrame(step)
}

const bindBackTop = () => {
  backTopScroller?.removeEventListener('scroll', updateBackTop)
  backTopScroller = useWindowScroll.value ? window : (props.scrollTarget ?? window)
  backTopScroller.addEventListener('scroll', updateBackTop, { passive: true })
  observeBackTopColumn()
  updateBackTop()
  settleBackTop()
}

// A fixed handful of rows cannot fill a tall screen, so every appended page
// leaves the sentinel inside the prefetch band and the feed chases the next
// one immediately (a flickering loading indicator and a burst of cursor
// requests). Size each page from the viewport instead. On the mobile layout
// the column is not the scroller: its clientHeight is the content height.
const syncPageSize = () => {
  const viewport = useWindowScroll.value
    ? window.innerHeight
    : (props.scrollTarget?.clientHeight ?? window.innerHeight)
  echoStore.pageSize = timelinePageSize(viewport)
}

const handleViewportChange = () => {
  syncScrollMode()
  syncPageSize()
  bindBackTop()
}

const sentinel = ref<HTMLElement | null>(null)
useInfiniteScroll({
  sentinel,
  scrollTarget: () => props.scrollTarget ?? null,
  canLoad: () => echoStore.canLoadMore,
  load: () => echoStore.loadMore(),
  items: computed(() => echoStore.echoList),
  rootMargin: TIMELINE_LOAD_AHEAD_PX,
})

const parseLegacyPage = (raw: unknown): number => {
  const value = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isFinite(value) || value < 1) return 1
  return Math.floor(value)
}

// Rows already on screen when the timeline was left (tab switch, back
// navigation). Stored as a row count rather than a page count so a different
// page size on the next mount cannot change how deep the feed reopens.
const storedCount = (): number => {
  const value = Number(sessionStorage.getItem(HOME_TIMELINE_COUNT_KEY))
  return Number.isFinite(value) && value > 0 ? Math.floor(value) : 0
}

watch(
  () => echoStore.echoList.length,
  (length) => {
    sessionStorage.setItem(HOME_TIMELINE_COUNT_KEY, String(length))
  },
)

watch(isFilteringMode, () => {
  echoStore.refreshEchos()
})

watch(
  () => props.scrollTarget,
  () => {
    syncScrollMode()
    syncPageSize()
    bindBackTop()
  },
)

onMounted(async () => {
  const legacyPage = parseLegacyPage(route.query.page)
  if (route.query.page !== undefined) {
    const nextQuery = { ...route.query }
    delete nextQuery.page
    router.replace({ query: nextQuery })
  }

  syncScrollMode()
  syncPageSize()
  bindBackTop()
  scrollModeReady.value = true
  window.addEventListener('resize', handleViewportChange, { passive: true })

  // A refresh started elsewhere (publishing returns to this tab) already owns
  // the request; wait for it instead of firing a duplicate.
  if (echoStore.isLoading && echoStore.hasFetched) return
  if (echoStore.isTimelineCurrent) return

  // legacy page numbers were counted in baseline-size pages.
  const legacyCount = legacyPage > 1 ? legacyPage * TIMELINE_BASE_PAGE_SIZE : 0
  await echoStore.refreshEchos(Math.max(echoStore.pageSize, storedCount(), legacyCount))
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleViewportChange)
  backTopScroller?.removeEventListener('scroll', updateBackTop)
  backTopScroller = null
  backTopColumnObserver?.disconnect()
  backTopColumnObserver = null
  if (settleFrame !== null) cancelAnimationFrame(settleFrame)
  settleFrame = null
})
</script>

<style scoped>
/*
 * Chrome's scroll anchoring keeps the node nearest the viewport bottom fixed
 * while content above it grows, so appending a page moved the scroll offset
 * down by the same amount and left the sentinel inside the prefetch band —
 * the loader then asked for the next cursor page again, chaining pages while
 * the reader sat still. The feed owns its scroll position; the virtualizer
 * compensates for measured rows itself.
 */
.echo-feed {
  overflow-anchor: none;
}

.echos-toolbar {
  font-family: var(--font-family-display);
}

.echo-feed-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.6rem;
  margin: 1.5rem 0;
  font-family: var(--font-family-display);
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.echo-feed-retry {
  padding: 0.35rem 1.1rem;
  font-family: var(--font-family-display);
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  background: transparent;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  cursor: pointer;
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease;
}

.echo-feed-retry:hover {
  color: var(--color-text-primary);
  border-color: var(--color-text-secondary);
  background: var(--color-border-subtle);
}

.echo-feed-retry:focus-visible {
  outline: 2px solid var(--color-text-secondary);
  outline-offset: 2px;
}

.echo-feed-end {
  margin: 1.5rem 0 0.5rem;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  text-align: center;
}

.feed-sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
