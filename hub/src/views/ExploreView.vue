<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch, nextTick } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import { RouterLink } from 'vue-router'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import HubEchoCard from '../components/HubEchoCard.vue'
import HubToast from '../components/HubToast.vue'
import { useHubMergeFeed } from '../composables/useHubMergeFeed'
import { loadHubConfig } from '../services/loadHubConfig'
import { probeInstances } from '../services/probeInstances'
import { fetchInstancesConnectBundle } from '../services/connectApi'
import type { HubInstance } from '../types/hub'
import { localStg } from '../utils/storage'
import '../styles/hub-shell.css'

const EXPLORE_LAYOUT_KEY = 'hub_explore_layout'

function readExploreLayout(): 'list' | 'masonry' {
  const v = localStg.getItem<string>(EXPLORE_LAYOUT_KEY)
  return v === 'masonry' ? 'masonry' : 'list'
}

const { t } = useI18n()

const loadingHub = ref(true)
const loadingProbe = ref(false)
const hubError = ref<string | null>(null)
const instances = ref<HubInstance[]>([])

const eligibleInstances = ref<HubInstance[]>([])

const {
  echoList,
  isPreparing,
  isLoading,
  hasMore,
  hasTriedInitialLoad,
  reset,
  setInstanceLogos,
  prepareInstances,
  loadEchoListPage,
} = useHubMergeFeed()

const feedError = ref<string | null>(null)
const toastMessage = ref<string | null>(null)

const exploreLayout = ref<'list' | 'masonry'>(readExploreLayout())

const masonryViewportMd = useMediaQuery('(min-width: 768px)')

const feedLayoutIsMasonry = computed(
  () => exploreLayout.value === 'masonry' && masonryViewportMd.value,
)

const layoutToggleListActive = computed(
  () => exploreLayout.value === 'list' || !masonryViewportMd.value,
)
const layoutToggleMasonryActive = computed(
  () => exploreLayout.value === 'masonry' && masonryViewportMd.value,
)

watch(exploreLayout, (mode) => {
  localStg.setItem(EXPLORE_LAYOUT_KEY, mode)
  nextTick(() => {
    scheduleLayout()
  })
})

watch(feedLayoutIsMasonry, () => {
  nextTick(() => scheduleLayout())
})

const sentinelRef = ref<HTMLElement | null>(null)
const showBackTop = ref(false)
let sentinelObserver: IntersectionObserver | null = null

const updateShowBackTop = () => {
  showBackTop.value = window.scrollY > 300
}

const onSentinelVisible = () => {
  if (isLoading.value || isPreparing.value || !hasMore.value) return
  void loadEchoListPage()
}

const setupSentinelObserver = () => {
  sentinelObserver?.disconnect()
  sentinelObserver = null
  const el = sentinelRef.value
  if (!el) return
  sentinelObserver = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) onSentinelVisible()
    },
    { root: null, rootMargin: '0px 0px 300px 0px' },
  )
  sentinelObserver.observe(el)
}

onMounted(async () => {
  window.addEventListener('scroll', updateShowBackTop, { passive: true })

  loadingHub.value = true
  hubError.value = null
  try {
    instances.value = await loadHubConfig()
  } catch (e) {
    hubError.value = e instanceof Error ? e.message : String(e)
    loadingHub.value = false
    return
  }
  loadingHub.value = false

  if (instances.value.length === 0) return

  loadingProbe.value = true
  eligibleInstances.value = []
  try {
    const probed = await probeInstances(instances.value)
    eligibleInstances.value = probed.eligible
  } catch (e) {
    hubError.value = e instanceof Error ? e.message : String(e)
    loadingProbe.value = false
    return
  }
  loadingProbe.value = false

  if (eligibleInstances.value.length === 0) return

  feedError.value = null
  reset()
  try {
    const { logos } = await fetchInstancesConnectBundle(eligibleInstances.value)
    setInstanceLogos(logos)
    await prepareInstances(eligibleInstances.value)
    await loadEchoListPage()

    const n = eligibleInstances.value.length
    if (n > 0) {
      toastMessage.value = `Loaded ${n} instance${n === 1 ? '' : 's'}`
    }
  } catch (e) {
    feedError.value = e instanceof Error ? e.message : String(e)
  }

  await nextTick()
  scheduleLayout()
  setupSentinelObserver()
})

watch(isLoading, (loading) => {
  if (loading || !hasMore.value) return
  nextTick(() => onSentinelVisible())
})

watch(
  echoList,
  () => {
    nextTick(() => setupSentinelObserver())
  },
  { flush: 'post' },
)

function scheduleLayout() {
  updateShowBackTop()
}

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updateShowBackTop)
  sentinelObserver?.disconnect()
  sentinelObserver = null
})
</script>

<template>
  <div class="hub-shell min-h-screen w-full">
    <div class="hub-explore-page-dimmed min-h-screen w-full">
      <nav
        class="hub-explore-mono mx-auto flex w-full items-center justify-between gap-3 px-5 pt-8 pb-2 transition-[max-width] duration-200"
        :class="
          feedLayoutIsMasonry
            ? 'max-w-[min(100%,90rem)]'
            : 'max-w-[min(100%,34rem)]'
        "
      >
        <RouterLink
          to="/"
          class="text-sm text-[var(--color-text-muted)] no-underline transition hover:text-[var(--color-text-primary)]"
        >
          ← Home
        </RouterLink>
        <div
          class="inline-flex shrink-0 items-center gap-0.5 rounded border border-[var(--color-border-subtle)] p-0.5"
          role="group"
          :aria-label="t('hub.layoutSwitch')"
        >
          <button
            type="button"
            class="hub-explore-layout-btn inline-flex h-8 w-8 items-center justify-center rounded transition"
            :class="
              layoutToggleListActive
                ? 'bg-[var(--color-bg-elevated)] text-[var(--color-text-primary)]'
                : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
            "
            :aria-pressed="layoutToggleListActive"
            :title="t('hub.layoutList')"
            @click="exploreLayout = 'list'"
          >
            <span class="sr-only">{{ t('hub.layoutList') }}</span>
            <svg
              class="h-4 w-4"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
              />
            </svg>
          </button>
          <button
            type="button"
            class="hub-explore-layout-btn inline-flex h-8 w-8 items-center justify-center rounded transition"
            :class="
              layoutToggleMasonryActive
                ? 'bg-[var(--color-bg-elevated)] text-[var(--color-text-primary)]'
                : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]'
            "
            :aria-pressed="layoutToggleMasonryActive"
            :title="t('hub.layoutMasonry')"
            @click="exploreLayout = 'masonry'"
          >
            <span class="sr-only">{{ t('hub.layoutMasonry') }}</span>
            <svg
              class="h-4 w-4"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M3.75 4.5h6.75v6.75H3.75V4.5Zm9.75 0h6.75v6.75h-6.75V4.5Zm-9.75 9h6.75v6.75H3.75v-6.75Zm9.75 0h6.75v6.75h-6.75v-6.75Z"
              />
            </svg>
          </button>
        </div>
      </nav>

      <div
        class="mx-auto flex w-full justify-center px-5 pb-10 pt-2 transition-[max-width] duration-200"
        :class="
          feedLayoutIsMasonry
            ? 'max-w-[min(100%,90rem)]'
            : 'max-w-[min(100%,34rem)]'
        "
      >
      <div class="w-full text-[var(--color-text-muted)]">
        <section v-if="loadingHub" class="hub-explore-mono my-8">
          <TheLoadingIndicator label="Loading hub.json…" />
        </section>

        <section
          v-else-if="hubError"
          class="hub-explore-mono my-8 text-center text-[var(--color-danger)]"
        >
          <strong>Couldn’t load the instance list</strong>
          <p class="mt-1">{{ hubError }}</p>
        </section>

        <section
          v-else-if="!instances.length"
          class="hub-explore-mono my-8 text-center text-[var(--color-text-secondary)]"
        >
          No instances configured. Add them in <code>hub/public/hub.json</code> or open a GitHub
          issue to list yours.
        </section>

        <template v-else>
          <section v-if="loadingProbe" class="hub-explore-mono my-8">
            <TheLoadingIndicator label="Checking instances…" />
          </section>

          <section v-else id="explore-feed" class="scroll-mt-8">
            <section v-if="isPreparing" class="hub-explore-mono my-8">
              <TheLoadingIndicator :label="t('hub.loading')" />
            </section>

            <section
              v-else-if="feedError"
              class="hub-explore-mono my-8 text-center text-[var(--color-danger)]"
            >
              {{ feedError }}
            </section>

            <section
              v-else-if="echoList.length === 0 && hasTriedInitialLoad && !isPreparing"
              class="hub-explore-mono my-8"
            >
              <p class="text-center text-[var(--color-text-secondary)]">
                {{ t('hub.emptyConnectHint') }}
              </p>
            </section>

            <div
              v-else-if="echoList.length > 0 && !isPreparing"
              class="hub-explore-feed"
              :class="{ 'hub-explore-feed--masonry': feedLayoutIsMasonry }"
            >
              <div
                v-for="item in echoList"
                :key="item.virtual_key"
                class="hub-explore-echo-row hub-item-wrap py-2"
                :class="
                  feedLayoutIsMasonry
                    ? 'hub-item-wrap--masonry block w-full'
                    : 'flex items-center justify-center'
                "
              >
                <HubEchoCard
                  :echo="item"
                  :variant="feedLayoutIsMasonry ? 'masonry' : 'default'"
                />
              </div>
            </div>

            <div v-if="echoList.length > 0 && !hasMore" class="hub-explore-mono my-8">
              <p class="text-center text-[var(--color-text-secondary)]">
                {{ t('hub.noMoreData') }}
              </p>
            </div>

            <div v-if="isLoading && echoList.length > 0" class="hub-explore-mono my-6">
              <TheLoadingIndicator :label="t('hub.loading')" />
            </div>

            <div ref="sentinelRef" class="h-1" />
          </section>
        </template>
      </div>
    </div>
    </div>

    <div
      v-show="showBackTop"
      class="fixed bottom-6 z-50 transition-all duration-500"
      style="right: max(1rem, env(safe-area-inset-right, 0px))"
    >
      <TheBackTop class="w-8 h-8 p-1" />
    </div>
    <HubToast :message="toastMessage" />
  </div>
</template>

<style scoped>
.hub-item-wrap {
  contain: layout;
}

.hub-explore-feed--masonry {
  column-count: 2;
  column-gap: 1rem;
}

@media (min-width: 768px) {
  .hub-explore-feed--masonry {
    column-count: 3;
  }
}

@media (min-width: 1280px) {
  .hub-explore-feed--masonry {
    column-count: 4;
  }
}

.hub-explore-feed--masonry .hub-item-wrap--masonry {
  contain: unset;
  break-inside: avoid;
  page-break-inside: avoid;
  display: inline-block;
  width: 100%;
  vertical-align: top;
  box-sizing: border-box;
}
</style>
