<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="home-sidebar-nav">
    <div class="home-sidebar-nav__mobile-row">
      <nav class="home-sidebar-nav__nav" aria-label="Primary">
        <RouterLink
          v-for="item in visibleItems"
          :key="item.id"
          :to="item.to"
          class="home-sidebar-nav__link"
          :class="{ 'home-sidebar-nav__link--active': isItemActive(item) }"
        >
          {{ t(item.labelKey) }}
        </RouterLink>
      </nav>
      <button
        v-if="showSearchTrigger"
        type="button"
        class="home-sidebar-nav__search-trigger"
        :aria-expanded="searchOpenState"
        :aria-label="t('homeTop.searchTitle')"
        @click="handleSearchClick"
      >
        <Search class="home-sidebar-nav__search-icon" />
      </button>
      <button
        v-if="chatAvailable"
        type="button"
        class="home-sidebar-nav__chat-trigger"
        :aria-label="t('chatLauncher.title')"
        @click="emit('openChat')"
      >
        <Chat class="home-sidebar-nav__chat-icon" />
      </button>
    </div>
    <div v-if="showMobileFilter" class="home-sidebar-nav__mobile-filter">
      <TheFilter @open-palette="emit('openPalette')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEchoStore, useSettingStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import Search from '@/components/icons/search.vue'
import Chat from '@/components/icons/chat.vue'
import TheFilter from './TheFilter.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const echoStore = useEchoStore()
const settingStore = useSettingStore()
const { isLogin } = storeToRefs(userStore)
const { searchingMode, isFilteringMode } = storeToRefs(echoStore)
const { AgentSetting } = storeToRefs(settingStore)
const chatAvailable = computed(() => isLogin.value && AgentSetting.value.enable)
const props = defineProps<{
  mobileSearchOpen?: boolean
}>()
const emit = defineEmits<{
  (event: 'update:mobileSearchOpen', value: boolean): void
  (event: 'openPalette'): void
  (event: 'openChat'): void
}>()
const localSearchOpen = ref(false)
const searchOpenState = computed({
  get: () => props.mobileSearchOpen ?? localSearchOpen.value,
  set: (value: boolean) => {
    if (props.mobileSearchOpen === undefined) {
      localSearchOpen.value = value
      return
    }
    emit('update:mobileSearchOpen', value)
  },
})
const currentHomeTab = computed<'home' | 'publish' | 'status' | 'tags' | 'hub'>(() => {
  if (route.query.tab === 'publish') return 'publish'
  if (route.query.tab === 'status') return 'status'
  if (route.query.tab === 'tags') return 'tags'
  if (route.query.tab === 'hub') return 'hub'
  return 'home'
})
const showSearchTrigger = computed(() => route.name === 'home' && currentHomeTab.value === 'home')
const showMobileFilter = computed(
  () =>
    showSearchTrigger.value &&
    (searchOpenState.value || searchingMode.value || isFilteringMode.value),
)

const items = [
  {
    id: 'home',
    to: { name: 'home' },
    labelKey: 'homeSidebar.home',
    kind: 'homeTab',
  },
  { id: 'panel', to: { name: 'panel' }, labelKey: 'homeSidebar.panel', kind: 'route' },
  {
    id: 'publish',
    to: { name: 'home', query: { tab: 'publish' } },
    labelKey: 'homeSidebar.publish',
    kind: 'homeTab',
  },
  {
    id: 'status',
    to: { name: 'home', query: { tab: 'status' } },
    labelKey: 'homeSidebar.status',
    kind: 'homeTab',
  },
  {
    id: 'tags',
    to: { name: 'home', query: { tab: 'tags' } },
    labelKey: 'homeSidebar.tags',
    kind: 'homeTab',
  },
  {
    id: 'hub',
    to: { name: 'home', query: { tab: 'hub' } },
    labelKey: 'homeSidebar.plaza',
    kind: 'homeTab',
  },
] as const

const visibleItems = computed(() =>
  items.filter((item) => {
    if (item.id === 'publish' || item.id === 'panel') return isLogin.value
    return true
  }),
)

const isItemActive = (item: (typeof items)[number]) => {
  if (item.kind === 'homeTab') {
    const tab = currentHomeTab.value
    const itemTab = 'query' in item.to ? item.to.query.tab : 'home'
    return route.name === 'home' && tab === itemTab
  }
  return route.name === item.to.name
}

const handleSearchClick = async () => {
  if (route.name !== 'home' || currentHomeTab.value !== 'home') {
    await router.push({ name: 'home' })
    searchOpenState.value = true
    return
  }
  searchOpenState.value = !searchOpenState.value
}

watch(showSearchTrigger, (visible) => {
  if (!visible) searchOpenState.value = false
})
</script>

<style scoped>
.home-sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.home-sidebar-nav__mobile-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.home-sidebar-nav__nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
  flex: 1 1 auto;
}

.home-sidebar-nav__link {
  display: block;
  padding: 0.4rem 0.7rem;
  border-radius: var(--radius-xs);
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition:
    color 0.2s,
    background 0.2s;
}

.home-sidebar-nav__link:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.home-sidebar-nav__link.home-sidebar-nav__link--active {
  color: var(--color-text-primary);
  background: var(--nav-link-active-bg);
  box-shadow: var(--shadow-soft);
}

.home-sidebar-nav__link.home-sidebar-nav__link--active:hover {
  color: var(--color-text-primary);
  background: var(--nav-link-active-bg);
}

.home-sidebar-nav__search-trigger,
.home-sidebar-nav__chat-trigger {
  display: none;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.home-sidebar-nav__mobile-filter {
  display: none;
}

.home-sidebar-nav__search-icon,
.home-sidebar-nav__chat-icon {
  width: 1rem;
  height: 1rem;
}

:deep(.home-sidebar-nav__search-icon path),
:deep(.home-sidebar-nav__chat-icon path) {
  fill: currentColor;
}

@media (width <= 819.98px) {
  .home-sidebar-nav__nav {
    flex-direction: row;
    gap: 0.2rem;
    overflow-x: auto;
    white-space: nowrap;
  }

  .home-sidebar-nav__nav::-webkit-scrollbar {
    display: none;
  }

  .home-sidebar-nav__link {
    flex: 0 0 auto;
    padding: 0.3rem 0.5rem;
    border-radius: var(--radius-xs);
    font-size: 0.9rem;
    line-height: 1.2;
  }

  .home-sidebar-nav__search-trigger,
  .home-sidebar-nav__chat-trigger {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    padding: 0.2rem;
    border-radius: 0.375rem;
    color: var(--color-text-muted);
    transition:
      color 0.2s,
      background 0.2s;
  }

  .home-sidebar-nav__search-trigger:hover,
  .home-sidebar-nav__chat-trigger:hover {
    color: var(--color-text-secondary);
    background: var(--color-bg-muted);
  }

  .home-sidebar-nav__mobile-filter {
    display: block;
    margin: 0;
  }
}
</style>
