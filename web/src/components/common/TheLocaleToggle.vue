<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootEl" class="locale-toggle">
    <button
      ref="triggerEl"
      type="button"
      v-tooltip="t('homeNav.localeToggleTitle')"
      :aria-label="t('homeNav.localeToggleTitle')"
      :aria-expanded="open"
      aria-haspopup="menu"
      class="locale-toggle__trigger"
      @click="open = !open"
    >
      <LocaleIcon class="w-4 h-4" />
    </button>

    <div v-if="open" ref="menuEl" class="locale-toggle__menu" role="menu">
      <button
        v-for="item in options"
        :key="item.value"
        type="button"
        role="menuitemradio"
        :aria-checked="item.value === currentLocale"
        class="locale-toggle__item"
        :class="{ 'locale-toggle__item--active': item.value === currentLocale }"
        @click="select(item.value)"
      >
        <span class="locale-toggle__check" aria-hidden="true">{{
          item.value === currentLocale ? '✓' : ''
        }}</span>
        <span class="locale-toggle__label">{{ item.label }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { onClickOutside } from '@vueuse/core'
import { storeToRefs } from 'pinia'
import LocaleIcon from '@/components/icons/locale.vue'
import { setI18nLocale, LOCALE_OPTIONS, type AppLocale } from '@/locales'
import { useUserStore } from '@/stores'
import { fetchUpdateUser } from '@/service/api/user'
import { theToast } from '@/utils/toast'

const { t, locale } = useI18n()
const userStore = useUserStore()
const { user, isLogin } = storeToRefs(userStore)

const open = ref(false)
const rootEl = ref<HTMLElement | null>(null)
const triggerEl = ref<HTMLElement | null>(null)
const menuEl = ref<HTMLElement | null>(null)

const currentLocale = computed(() => locale.value)

const options = LOCALE_OPTIONS

onClickOutside(
  menuEl,
  () => {
    open.value = false
  },
  { ignore: [triggerEl] },
)

async function select(target: AppLocale) {
  open.value = false
  if (target === currentLocale.value) return
  await setI18nLocale(target)
  if (isLogin.value && user.value) {
    user.value.locale = target
    const payload: App.Api.User.UserInfo = {
      username: user.value.username,
      password: user.value.password ?? '',
      email: user.value.email,
      is_admin: user.value.is_admin,
      is_owner: user.value.is_owner,
      avatar: user.value.avatar ?? '',
      locale: target,
    }
    fetchUpdateUser(payload).catch(() => {
      theToast.error(String(t('homeNav.localeSyncFailed')))
    })
  }
}
</script>

<style scoped>
.locale-toggle {
  position: relative;
  display: inline-flex;
}

.locale-toggle__trigger {
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

.locale-toggle__trigger:hover {
  color: var(--color-text-secondary);
  background: var(--color-border-subtle);
}

.locale-toggle__menu {
  position: absolute;
  top: calc(100% + 0.375rem);
  right: 0;
  z-index: 50;
  min-width: 8.5rem;
  padding: 0.25rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.5rem;
  box-shadow:
    0 4px 6px -1px rgb(0 0 0 / 8%),
    0 2px 4px -2px rgb(0 0 0 / 6%);
}

.locale-toggle__item {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  width: 100%;
  padding: 0.4rem 0.5rem;
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-align: left;
  border-radius: 0.375rem;
  transition:
    color 0.15s,
    background 0.15s;
}

.locale-toggle__item:hover {
  color: var(--color-text-primary);
  background: var(--color-border-subtle);
}

.locale-toggle__item--active {
  color: var(--color-text-primary);
  font-weight: 600;
}

.locale-toggle__check {
  flex-shrink: 0;
  width: 0.85rem;
  text-align: center;
  color: var(--color-text-primary);
}

.locale-toggle__label {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
}
</style>
