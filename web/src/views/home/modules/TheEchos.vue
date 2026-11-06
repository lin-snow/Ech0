<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="mx-auto mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6"
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
      :layout-key="resolvedLayoutKey"
      :aria-busy="echoStore.isLoadingMore"
      @anchor-change="persistAnchor"
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
        <p>{{ t('homeFeed.loadMoreFailed') }}</p>
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
import { HOME_TIMELINE_ANCHOR_KEY, HOME_TIMELINE_DEPTH_KEY } from '@/constants/timeline'
import { serializeTimelineAnchor, type TimelineAnchor } from '@/utils/timelineAnchor'
import { isScrollableElement } from '@/utils/scroll'
import { localStg } from '@/utils/storage'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

const props = defineProps<{
  scrollTarget?: HTMLElement | null
  compact?: boolean
  layoutKey?: string
}>()

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()
const { SystemSetting } = storeToRefs(settingStore)
const { isFilteringMode, hasActiveQuery, searchingMode } = storeToRefs(echoStore)
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
  if (echoStore.loadError === 'more') return t('homeFeed.loadMoreFailed')
  if (echoStore.echoList.length > 0 && echoStore.nextCursor === null) {
    return t('homeFeed.allLoaded')
  }
  return ''
})

const persistAnchor = (anchor: TimelineAnchor) => {
  sessionStorage.setItem(HOME_TIMELINE_ANCHOR_KEY, serializeTimelineAnchor(anchor))
}

const useWindowScroll = ref(true)
const scrollModeReady = ref(false)
const syncScrollMode = () => {
  useWindowScroll.value = !isScrollableElement(props.scrollTarget ?? null)
}

const resolvedLayoutKey = computed(() =>
  [
    props.layoutKey ?? '',
    showChecklist.value,
    showFirstEchoCard.value,
    searchingMode.value,
    isFilteringMode.value,
  ].join('|'),
)

const sentinel = ref<HTMLElement | null>(null)
useInfiniteScroll({
  sentinel,
  scrollTarget: () => props.scrollTarget ?? null,
  canLoad: () => echoStore.canLoadMore,
  load: () => echoStore.loadMore(),
  items: computed(() => echoStore.echoList),
})

const parseLegacyPage = (raw: unknown): number => {
  const value = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isFinite(value) || value < 1) return 1
  return Math.floor(value)
}

const storedDepth = (): number => {
  const value = Number(sessionStorage.getItem(HOME_TIMELINE_DEPTH_KEY))
  return Number.isFinite(value) && value > 1 ? Math.floor(value) : 1
}

watch(
  () => echoStore.echoList.length,
  (length) => {
    const depth = Math.max(1, Math.ceil(length / echoStore.pageSize))
    sessionStorage.setItem(HOME_TIMELINE_DEPTH_KEY, String(depth))
  },
)

watch(isFilteringMode, () => {
  echoStore.refreshEchos()
})

watch(
  () => props.scrollTarget,
  () => syncScrollMode(),
)

onMounted(async () => {
  const legacyPage = parseLegacyPage(route.query.page)
  if (route.query.page !== undefined) {
    const nextQuery = { ...route.query }
    delete nextQuery.page
    router.replace({ query: nextQuery })
  }

  syncScrollMode()
  scrollModeReady.value = true
  window.addEventListener('resize', syncScrollMode, { passive: true })

  // A refresh started elsewhere (publishing returns to this tab) already owns
  // the request; wait for it instead of firing a duplicate.
  if (echoStore.isLoading && echoStore.hasFetched) return
  if (echoStore.isTimelineCurrent) return

  const depth = Math.max(storedDepth(), legacyPage)
  await echoStore.refreshEchos(depth > 1 ? depth * echoStore.pageSize : undefined)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', syncScrollMode)
})
</script>

<style scoped>
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
