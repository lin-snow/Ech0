<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="panel-page-wrap mt-4">
    <div class="panel-shell border p-4 mx-auto flex flex-col max-w-screen-lg w-full mb-12">
      <section class="panel-welcome mb-5 md:mb-2">
        <div class="panel-welcome-main">
          <h1 class="panel-welcome-username">
            {{ t('panelPage.welcomeBack', { username }) }} <span class="wave-hand">👋</span>
          </h1>
          <div class="panel-welcome-meta">
            <p class="panel-welcome-date">{{ dateText }}</p>
            <p class="panel-welcome-greeting">{{ greeting }}</p>
          </div>
          <p class="panel-welcome-tip">{{ t('panelPage.welcomeTip') }}</p>
        </div>
        <div class="panel-welcome-actions">
          <BaseButton :icon="BackHand" @click="router.push('/')" class="panel-home-btn" />
        </div>
      </section>

      <div class="md:hidden mb-6 px-2 flex justify-between items-center mb-3">
        <div class="w-1/2">
          <BaseSelect
            class="!focus:ring-0 h-9"
            v-model="selectedRoute"
            :options="routeOptions"
            :placeholder="t('panelPage.selectPage')"
            @change="handleRouteChange"
          />
        </div>

        <div class="flex gap-2 items-center">
          <BaseButton
            v-if="userStore.isLogin"
            :icon="Logout"
            @click="handleLogout"
            class="w-9 h-9 rounded-md"
            :tooltip="t('panelPage.logout')"
          >
          </BaseButton>
          <BaseButton
            v-else
            :icon="Auth"
            @click="router.push('/auth')"
            class="w-9 h-9 rounded-md"
            :tooltip="t('panelPage.loginOrSignup')"
          >
          </BaseButton>
        </div>
      </div>

      <div class="mx-auto flex my-4 w-full max-w-screen-lg rounded-md panel-main">
        <div class="hidden md:flex flex-col gap-2 w-52 pr-4 shrink-0 panel-nav">
          <BaseButton
            :icon="Dashboard"
            @click="router.push('/panel/dashboard')"
            :class="getButtonClasses('panel-dashboard')"
          >
            {{ t('panelPageNav.dashboard') }}
          </BaseButton>

          <BaseButton
            :icon="Setting"
            @click="router.push('/panel/setting')"
            :class="getButtonClasses('panel-setting')"
          >
            {{ t('panelPageNav.setting') }}
          </BaseButton>

          <BaseButton
            :icon="User"
            @click="router.push('/panel/user')"
            :class="getButtonClasses('panel-user')"
          >
            {{ t('panelPageNav.userCenter') }}
          </BaseButton>

          <BaseButton
            :icon="Storage"
            @click="router.push('/panel/storage')"
            :class="getButtonClasses('panel-storage')"
          >
            {{ t('panelPageNav.storage') }}
          </BaseButton>

          <BaseButton
            :icon="Data"
            @click="router.push('/panel/data-management')"
            :class="getButtonClasses('panel-data-management')"
          >
            {{ t('panelPageNav.dataManagement') }}
          </BaseButton>

          <BaseButton
            :icon="Extension"
            @click="router.push('/panel/comment')"
            :class="getButtonClasses('panel-comment')"
          >
            {{ t('panelPageNav.comment') }}
          </BaseButton>

          <BaseButton
            :icon="Sso"
            @click="router.push('/panel/sso')"
            :class="getButtonClasses('panel-sso')"
          >
            {{ t('panelPageNav.sso') }}
          </BaseButton>

          <BaseButton
            :icon="Extension"
            @click="router.push('/panel/extension')"
            :class="getButtonClasses('panel-extension')"
          >
            {{ t('panelPageNav.extension') }}
          </BaseButton>

          <BaseButton
            :icon="Others"
            @click="router.push('/panel/advance')"
            :class="getButtonClasses('panel-advance')"
          >
            {{ t('panelPageNav.advance') }}
          </BaseButton>

          <BaseButton
            :icon="Log"
            @click="router.push('/panel/system-log')"
            :class="getButtonClasses('panel-system-log')"
          >
            {{ t('panelPageNav.systemLog') }}
          </BaseButton>

          <div class="h-px bg-[var(--color-border-subtle)] mx-2" />

          <BaseButton :icon="Logout" @click="handleLogout" :class="getBottomButtonClasses()">
            {{ t('panelPage.logout') }}
          </BaseButton>

          <BaseButton :icon="Auth" @click="router.push('/auth')" :class="getBottomButtonClasses()">
            {{ t('panelPage.login') }}
          </BaseButton>
        </div>

        <div class="flex-1 min-w-0 panel-content">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import User from '@/components/icons/user.vue'
import Auth from '@/components/icons/auth.vue'
import BackHand from '@/components/icons/backhand.vue'
import Extension from '@/components/icons/extension.vue'
import Others from '@/components/icons/theothers.vue'
import Dashboard from '@/components/icons/dashboard.vue'
import Setting from '@/components/icons/setting.vue'
import Storage from '@/components/icons/storage.vue'
import Data from '@/components/icons/data.vue'
import Sso from '@/components/icons/sso.vue'
import Logout from '@/components/icons/logout.vue'
import Log from '@/components/icons/log.vue'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores'
import { useRouter, useRoute } from 'vue-router'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()
const { t, locale } = useI18n()

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

const currentRoute = computed(() => route.name as string)
const selectedRoute = ref(route.path)
const username = computed(() => userStore.user?.username || String(t('panelPage.friend')))
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return String(t('panelPage.greetingEarlyMorning'))
  if (hour < 12) return String(t('panelPage.greetingMorning'))
  if (hour < 18) return String(t('panelPage.greetingAfternoon'))
  return String(t('panelPage.greetingEvening'))
})
const dateText = computed(() => {
  return new Intl.DateTimeFormat(locale.value, {
    month: 'long',
    day: 'numeric',
    weekday: 'long',
  }).format(new Date())
})

const getButtonClasses = (routeName: string) => {
  const baseClasses =
    'w-full flex items-center gap-2 pl-3 py-1.5 rounded-[var(--radius-md)] transition-all duration-200 border-none !ring-0 justify-start bg-transparent whitespace-nowrap'
  const activeClasses =
    currentRoute.value === routeName
      ? 'text-[var(--color-text-primary)]! bg-[var(--nav-link-active-bg)]! shadow-[var(--shadow-soft)]!'
      : 'text-[var(--color-text-secondary)] !shadow-none hover:text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)]'

  return `${baseClasses} ${activeClasses}`
}

const getBottomButtonClasses = () => {
  return 'w-full flex items-center gap-2 pl-3 py-1.5 rounded-[var(--radius-md)] transition-all duration-200 border-none !shadow-none !ring-0 text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-muted)] justify-start bg-transparent whitespace-nowrap'
}

const routeOptions = computed(() => [
  { label: String(t('panelPageNav.dashboard')), value: '/panel/dashboard' },
  { label: String(t('panelPageNav.setting')), value: '/panel/setting' },
  { label: String(t('panelPageNav.userCenter')), value: '/panel/user' },
  { label: String(t('panelPageNav.storage')), value: '/panel/storage' },
  { label: String(t('panelPageNav.dataManagement')), value: '/panel/data-management' },
  { label: String(t('panelPageNav.comment')), value: '/panel/comment' },
  { label: String(t('panelPageNav.sso')), value: '/panel/sso' },
  { label: String(t('panelPageNav.extension')), value: '/panel/extension' },
  { label: String(t('panelPageNav.advance')), value: '/panel/advance' },
  { label: String(t('panelPageNav.systemLog')), value: '/panel/system-log' },
])

watch(
  () => route.path,
  (newPath) => {
    selectedRoute.value = newPath
  },
)

const handleRouteChange = () => {
  router.push(selectedRoute.value)
}

const handleLogout = () => {
  if (!userStore.isLogin) {
    theToast.info(String(t('panelPage.notLogin')))
    return
  }

  openConfirm({
    title: String(t('panelPage.logoutConfirmTitle')),
    description: '',
    onConfirm: () => {
      userStore.logout()
      router.push('/')
      theToast.success(String(t('panelPage.logoutSuccess')))
    },
  })
}
</script>

<style scoped>
.panel-page-wrap {
  min-height: calc(100vh - 3.5rem);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.6rem 0.5rem 0;
}

.panel-shell {
  border-color: var(--color-border-subtle);
  border-radius: var(--radius-lg);
  background: var(--color-bg-canvas);
}

.panel-nav {
  color: var(--color-text-secondary);
}

.panel-main {
  align-items: flex-start;
  gap: 0.25rem;
}

.panel-content {
  max-width: 53rem;
  width: 100%;
  padding-left: 0.2rem;
}

.panel-welcome {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.5rem 0.25rem 0.65rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.panel-welcome-main {
  min-width: 0;
  text-align: left;
}

.panel-welcome-username {
  margin: 0;
  font-size: clamp(1.45rem, 2.4vw, 1.9rem);
  line-height: 1.2;
  color: var(--color-text-primary);
  font-weight: 800;
  font-family: var(--font-family-display);
  letter-spacing: 0.01em;
}

.panel-welcome-greeting {
  font-size: 0.9rem;
  line-height: 1.2;
  color: var(--color-text-muted);
  font-weight: 600;
}

.panel-welcome-meta {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  margin-top: 0.35rem;
}

.panel-welcome-tip {
  margin-top: 0.3rem;
  color: var(--color-text-muted);
  font-size: 0.88rem;
}

.panel-welcome-date {
  display: inline-flex;
  align-items: center;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  white-space: nowrap;
  font-weight: 600;
}

.panel-welcome-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.5rem;
}

.panel-home-btn {
  border: 1px solid var(--color-border-subtle) !important;
  border-radius: 999px !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
  width: 2rem !important;
  height: 2rem !important;
  padding: 0 !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease;
}

.panel-home-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.wave-hand {
  display: inline-block;
  margin-left: 0.2rem;
  transform-origin: center;
  will-change: transform;
}

.wave-hand:hover {
  animation: hand-shake 620ms ease-in-out;
}

@keyframes hand-shake {
  0% {
    transform: rotate(0deg) scale(1);
  }

  15% {
    transform: rotate(16deg) scale(1.08);
  }

  30% {
    transform: rotate(-14deg) scale(1.08);
  }

  45% {
    transform: rotate(12deg) scale(1.06);
  }

  60% {
    transform: rotate(-10deg) scale(1.04);
  }

  75% {
    transform: rotate(7deg) scale(1.02);
  }

  100% {
    transform: rotate(0deg) scale(1);
  }
}

@media (width <= 768px) {
  .panel-page-wrap {
    min-height: auto;
    display: block;
    padding: 0.4rem 0.35rem 0;
  }

  .panel-welcome {
    align-items: flex-start;
    flex-direction: row;
    justify-content: space-between;
    gap: 0.6rem;
  }

  .panel-welcome-date {
    white-space: normal;
  }

  .panel-main {
    gap: 0;
  }

  .panel-content {
    max-width: 100%;
    padding-left: 0;
  }
}
</style>
