<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <aside class="home-left-rail">
    <!-- 品牌区：logo + 站点名 -->
    <button
      type="button"
      class="home-left-rail__brand"
      :aria-label="t('homeSidebar.home')"
      @click="handleGoExplore"
    >
      <img :src="logo" alt="" loading="lazy" class="home-left-rail__logo" />
      <span class="home-left-rail__title">{{ siteTitle }}</span>
    </button>

    <!-- 主导航：纵向排列，图标 + 文字（Twitter 风格） -->
    <nav class="home-left-rail__nav" aria-label="Primary">
      <RouterLink
        v-for="item in visibleItems"
        :key="item.id"
        :to="item.to"
        class="home-left-rail__link"
        :class="{ 'home-left-rail__link--active': isItemActive(item) }"
      >
        <component :is="item.icon" class="home-left-rail__link-icon" />
        <span class="home-left-rail__link-label">{{ t(item.labelKey) }}</span>
      </RouterLink>
    </nav>

    <!-- 快捷操作：主题 / RSS / Zen / 登录登出 -->
    <div class="home-left-rail__actions">
      <button
        type="button"
        v-tooltip="themeToggleTooltip"
        :aria-label="t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel })"
        class="home-left-rail__action"
        @click="handleThemeToggle"
      >
        <component :is="themeIcon" class="home-left-rail__action-icon" />
        <span class="home-left-rail__action-label">{{ currentThemeModeLabel }}</span>
      </button>
      <a
        href="/rss"
        v-tooltip="t('homeTop.rssTitle')"
        :aria-label="t('homeTop.rssTitle')"
        class="home-left-rail__action"
      >
        <Rss class="home-left-rail__action-icon" />
        <span class="home-left-rail__action-label">{{ t('homeTop.rssTitle') }}</span>
      </a>
      <button
        type="button"
        v-tooltip="t('zenMode.tooltip')"
        :aria-label="t('zenMode.tooltip')"
        class="home-left-rail__action"
        @click="handleGoZen"
      >
        <Zen class="home-left-rail__action-icon" />
        <span class="home-left-rail__action-label">Zen</span>
      </button>
      <button
        v-if="!isLogin"
        type="button"
        v-tooltip="t('authPage.login')"
        :aria-label="t('authPage.login')"
        class="home-left-rail__action"
        @click="handleGoLogin"
      >
        <Auth class="home-left-rail__action-icon" />
        <span class="home-left-rail__action-label">{{ t('authPage.login') }}</span>
      </button>
      <button
        v-else
        type="button"
        v-tooltip="t('panelPage.logout')"
        :aria-label="t('panelPage.logout')"
        class="home-left-rail__action"
        @click="handleLogout"
      >
        <Signoff class="home-left-rail__action-icon" />
        <span class="home-left-rail__action-label">{{ t('panelPage.logout') }}</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { useSettingStore, useUserStore, useThemeStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

import HomeIcon from '@/components/icons/home.vue'
import HubIcon from '@/components/icons/hub.vue'
import StatusIcon from '@/components/icons/status.vue'
import TagsetIcon from '@/components/icons/tagsetting.vue'
import PublishIcon from '@/components/icons/publish.vue'
import PanelIcon from '@/components/icons/panel.vue'
import LightIcon from '@/components/icons/light.vue'
import DarkIcon from '@/components/icons/dark.vue'
import TreeIcon from '@/components/icons/tree.vue'
import SystemIcon from '@/components/icons/system.vue'
import Rss from '@/components/icons/rss.vue'
import Zen from '@/components/icons/zen.vue'
import Auth from '@/components/icons/auth.vue'
import Signoff from '@/components/icons/signoff.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const settingStore = useSettingStore()
const userStore = useUserStore()
const themeStore = useThemeStore()
const { openConfirm } = useBaseDialog()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)

const logo = computed(() => {
  if (isLogin.value && user.value?.avatar) return resolveAvatarUrl(user.value.avatar)
  return resolveAvatarUrl(SystemSetting.value?.server_logo)
})
const siteTitle = computed(() => String(SystemSetting.value?.server_name ?? 'Ech0'))

const currentHomeTab = computed<'home' | 'publish' | 'status' | 'tags' | 'hub'>(() => {
  if (route.query.tab === 'publish') return 'publish'
  if (route.query.tab === 'status') return 'status'
  if (route.query.tab === 'tags') return 'tags'
  if (route.query.tab === 'hub') return 'hub'
  return 'home'
})

const items = [
  {
    id: 'home',
    to: { name: 'home' },
    labelKey: 'homeSidebar.home',
    kind: 'homeTab',
    icon: HomeIcon,
  },
  {
    id: 'hub',
    to: { name: 'home', query: { tab: 'hub' } },
    labelKey: 'homeSidebar.plaza',
    kind: 'homeTab',
    icon: HubIcon,
  },
  {
    id: 'status',
    to: { name: 'home', query: { tab: 'status' } },
    labelKey: 'homeSidebar.status',
    kind: 'homeTab',
    icon: StatusIcon,
  },
  {
    id: 'tags',
    to: { name: 'home', query: { tab: 'tags' } },
    labelKey: 'homeSidebar.tags',
    kind: 'homeTab',
    icon: TagsetIcon,
  },
  {
    id: 'publish',
    to: { name: 'home', query: { tab: 'publish' } },
    labelKey: 'homeSidebar.publish',
    kind: 'homeTab',
    icon: PublishIcon,
  },
  {
    id: 'panel',
    to: { name: 'panel' },
    labelKey: 'homeSidebar.panel',
    kind: 'route',
    icon: PanelIcon,
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

const nextThemeMode = computed(() => {
  if (themeStore.mode === 'system') return 'light'
  if (themeStore.mode === 'light') return 'sunny'
  if (themeStore.mode === 'sunny') return 'dark'
  return 'system'
})
const themeIcon = computed(() => {
  // 按钮始终展示“当前模式”，跟随系统时显示系统图标
  if (themeStore.mode === 'system') return SystemIcon
  if (themeStore.mode === 'light') return LightIcon
  if (themeStore.mode === 'dark') return DarkIcon
  return TreeIcon
})
const currentThemeModeLabel = computed(() => {
  if (themeStore.mode === 'system') return String(t('homeNav.themeSystem'))
  if (themeStore.mode === 'light') return String(t('homeNav.themeLight'))
  if (themeStore.mode === 'dark') return String(t('homeNav.themeDark'))
  return String(t('homeNav.themeSunny'))
})
const nextThemeModeLabel = computed(() => {
  if (nextThemeMode.value === 'system') return String(t('homeNav.themeSystem'))
  if (nextThemeMode.value === 'light') return String(t('homeNav.themeLight'))
  if (nextThemeMode.value === 'dark') return String(t('homeNav.themeDark'))
  return String(t('homeNav.themeSunny'))
})
const themeToggleTooltip = computed(() => ({
  content: String(t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel.value })),
  triggers: ['hover'],
  hideTriggers: ['hover', 'click', 'touch'],
}))

const handleThemeToggle = () => themeStore.toggleTheme()
const handleGoExplore = () => router.push({ name: 'home' })
const handleGoZen = () => router.push({ name: 'zen' })
const handleGoLogin = () => router.push({ name: 'auth' })
const handleLogout = () => {
  if (!isLogin.value) return
  openConfirm({
    title: String(t('panelPage.logoutConfirmTitle')),
    description: '',
    onConfirm: async () => {
      await userStore.logout()
      await router.push({ name: 'home' })
      theToast.success(String(t('panelPage.logoutSuccess')))
    },
  })
}
</script>

<style scoped>
.home-left-rail {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.25rem 0.5rem 1.5rem;
  min-width: 0;
}

.home-left-rail__brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  text-align: left;
  min-width: 0;
  transition: background 0.15s ease;
}

.home-left-rail__brand:hover {
  background: var(--color-bg-muted);
}

.home-left-rail__logo {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 9999px;
  object-fit: cover;
  flex-shrink: 0;
  border: 2px solid var(--color-bg-surface);
  box-shadow: 0 0 0 1px var(--color-border-subtle);
}

.home-left-rail__title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.home-left-rail__nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.home-left-rail__link {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.625rem 0.875rem;
  border-radius: 9999px;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition:
    color 0.15s ease,
    background 0.15s ease;
}

.home-left-rail__link:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.home-left-rail__link--active,
.home-left-rail__link--active:hover {
  color: var(--color-text-primary);
  background: var(--nav-link-active-bg);
  font-weight: 700;
  box-shadow: var(--shadow-sm);
}

.home-left-rail__link-icon {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
  color: currentcolor;
}

.home-left-rail__link-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.home-left-rail__actions {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  margin-top: auto;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border-subtle);
}

.home-left-rail__action {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.5rem 0.875rem;
  border: 0;
  background: transparent;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-muted);
  text-decoration: none;
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease;
}

.home-left-rail__action:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.home-left-rail__action-icon {
  width: 1.05rem;
  height: 1.05rem;
  flex-shrink: 0;
}

.home-left-rail__action-label {
  white-space: nowrap;
}
</style>
