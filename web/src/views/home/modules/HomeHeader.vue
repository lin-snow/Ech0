<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="home-header">
    <div class="home-header__brand">
      <button
        type="button"
        class="home-header__logo-wrap"
        :aria-label="t('homeSidebar.home')"
        @click="handleGoExplore"
      >
        <img :src="logo" alt="" loading="lazy" class="home-header__logo" />
      </button>
      <h1 class="home-header__title">
        <span>{{ typedTitle }}</span>
        <span v-if="isTypingTitle" class="home-header__cursor" aria-hidden="true"></span>
      </h1>
    </div>

    <div class="home-header__actions">
      <div class="home-header__links">
        <a
          :href="rssHref"
          v-tooltip="t('homeTop.rssTitle')"
          :aria-label="t('homeTop.rssTitle')"
          class="home-header__link-icon"
        >
          <Rss class="w-4 h-4" />
        </a>
        <button
          type="button"
          v-tooltip="t('zenMode.tooltip')"
          :aria-label="t('zenMode.tooltip')"
          class="home-header__link-icon"
          @click="handleGoZen"
        >
          <Zen class="w-4 h-4" />
        </button>
        <button
          type="button"
          v-tooltip="themeToggleTooltip"
          :aria-label="t('homeNav.themeToggleTitle', { mode: nextThemeModeLabel })"
          class="home-header__link-icon"
          @click="handleThemeToggle"
        >
          <component :is="themeIcon" class="w-4 h-4" />
        </button>
        <TheLocaleToggle />
        <template v-if="!isStatic">
          <button
            v-if="!isLogin"
            type="button"
            v-tooltip="t('authPage.login')"
            :title="t('authPage.login')"
            :aria-label="t('authPage.login')"
            class="home-header__link-icon"
            @click="handleGoLogin"
          >
            <Auth class="block w-4 h-4" />
          </button>
          <button
            v-else
            type="button"
            v-tooltip="t('panelPage.logout')"
            :title="t('panelPage.logout')"
            :aria-label="t('panelPage.logout')"
            class="home-header__link-icon"
            @click="handleLogout"
          >
            <Signoff class="block w-4 h-4" />
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import LightIcon from '@/components/icons/light.vue'
import DarkIcon from '@/components/icons/dark.vue'
import TreeIcon from '@/components/icons/tree.vue'
import SystemIcon from '@/components/icons/system.vue'
import Rss from '@/components/icons/rss.vue'
import Auth from '@/components/icons/auth.vue'
import Signoff from '@/components/icons/signoff.vue'
import Zen from '@/components/icons/zen.vue'
import TheLocaleToggle from '@/components/common/TheLocaleToggle.vue'
import { storeToRefs } from 'pinia'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingStore, useUserStore, useThemeStore } from '@/stores'
import { resolveAvatarUrl, isStaticMode, staticBase } from '@/service/request/shared'
import { useRouter } from 'vue-router'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

const settingStore = useSettingStore()
const userStore = useUserStore()
const themeStore = useThemeStore()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)
const { t } = useI18n()
const router = useRouter()
const { openConfirm } = useBaseDialog()

const isStatic = isStaticMode()
const rssHref = computed(() => (isStatic ? staticBase() + 'rss.xml' : '/rss'))

const logo = computed(() => {
  if (isLogin.value && user.value?.avatar) {
    return resolveAvatarUrl(user.value.avatar)
  }
  return resolveAvatarUrl(SystemSetting.value?.server_logo)
})
const fullTitle = computed(() => String(SystemSetting.value?.server_name ?? ''))
const typedTitle = ref('')
const isTypingTitle = ref(false)
const CURSOR_INTRO_DELAY_MS = 2000
const TITLE_TYPING_INTERVAL_MS = 85
let introDelayTimer: ReturnType<typeof setTimeout> | null = null
let typingTimer: ReturnType<typeof setTimeout> | null = null

const clearTypingTimers = () => {
  if (introDelayTimer) {
    clearTimeout(introDelayTimer)
    introDelayTimer = null
  }
  if (typingTimer) {
    clearTimeout(typingTimer)
    typingTimer = null
  }
}

const runTypingEffect = () => {
  clearTypingTimers()
  typedTitle.value = ''
  isTypingTitle.value = true

  const nextTitle = fullTitle.value
  if (!nextTitle) {
    isTypingTitle.value = false
    return
  }

  let index = 0
  const typeNext = () => {
    index += 1
    typedTitle.value = nextTitle.slice(0, index)

    if (index < nextTitle.length) {
      typingTimer = setTimeout(typeNext, TITLE_TYPING_INTERVAL_MS)
      return
    }

    isTypingTitle.value = false
  }

  introDelayTimer = setTimeout(typeNext, CURSOR_INTRO_DELAY_MS)
}

const nextThemeMode = computed(() => {
  if (themeStore.mode === 'system') return 'light'
  if (themeStore.mode === 'light') return 'sunny'
  if (themeStore.mode === 'sunny') return 'dark'
  return 'system'
})

const themeIcon = computed(() => {
  // 按钮图标反映当前生效模式：跟随系统 -> 系统图标
  if (themeStore.mode === 'system') return SystemIcon
  if (themeStore.mode === 'light') return LightIcon
  if (themeStore.mode === 'dark') return DarkIcon
  return TreeIcon
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

const handleThemeToggle = async () => {
  await themeStore.toggleTheme()
}

const handleGoExplore = async () => {
  await router.push({ name: 'home' })
}

const handleGoZen = async () => {
  await router.push({ name: 'zen' })
}

const handleGoLogin = async () => {
  await router.push({ name: 'auth' })
}

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

watch(fullTitle, (nextTitle, prevTitle) => {
  if (nextTitle !== prevTitle) {
    runTypingEffect()
  }
})

onMounted(() => {
  runTypingEffect()
})

onBeforeUnmount(() => {
  clearTypingTimers()
})
</script>

<style scoped>
.home-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 0.25rem;
}

.home-header__brand {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  min-width: 0;
}

.home-header__logo-wrap {
  flex-shrink: 0;
  border: none;
  background: transparent;
  padding: 0;
  cursor: pointer;
}

.home-header__logo {
  width: 2.125rem;
  height: 2.125rem;
  border-radius: 9999px;
  object-fit: cover;
  border: 2px solid var(--color-bg-surface);
  box-shadow:
    0 0 0 1px var(--color-border-subtle),
    0 1px 2px rgb(0 0 0 / 6%);
}

.home-header__title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.home-header__cursor {
  display: inline-block;
  width: 0.42em;
  height: 0.9em;
  margin-left: 0.12em;
  border-radius: 0.08em;
  background: currentcolor;
  vertical-align: -0.08em;
  animation: home-header-cursor-blink 0.95s steps(1, end) infinite;
}

@keyframes home-header-cursor-blink {
  0%,
  45% {
    opacity: 1;
  }

  46%,
  100% {
    opacity: 0;
  }
}

@media (width >= 640px) {
  .home-header__title {
    font-size: 1.25rem;
  }
}

.home-header__actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.home-header__links {
  display: flex;
  align-items: center;
}

.home-header__link-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.35rem;
  color: var(--color-text-muted);
  border-radius: 0.375rem;
  transition:
    color 0.2s,
    background 0.2s;
}

.home-header__link-icon:hover {
  color: var(--color-text-secondary);
  background: var(--color-border-subtle);
}
</style>
