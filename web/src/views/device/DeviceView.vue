<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { fetchDeviceApprove, fetchDevicePending } from '@/service/api/device'
import { ACCESS_SCOPE_LABEL_KEYS } from '@/constants/accessScopes'
import { useSettingStore, useUserStore } from '@/stores'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const settingStore = useSettingStore()
const { user } = storeToRefs(userStore)
const { SystemSetting } = storeToRefs(settingStore)

/** Codes are 8 letters/digits; the hyphen is decoration the server strips anyway. */
function formatCode(raw: string) {
  const chars = raw
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
    .slice(0, 8)
  return chars.length > 4 ? `${chars.slice(0, 4)}-${chars.slice(4)}` : chars
}

const userCode = ref(formatCode(typeof route.query.code === 'string' ? route.query.code : ''))
const pending = ref<App.Api.Device.PendingInfo | null>(null)
const approved = ref(false)
const approving = ref(false)
const lookupFailed = ref(false)
const expired = ref(false)
const remaining = ref(0)

const siteName = computed(() => String(SystemSetting.value?.server_name || 'Ech0'))
const username = computed(() => user.value?.username || '')
/** Approving mints an access token, which the backend restricts to admins. */
const canApprove = computed(() => user.value?.is_admin === true)

const normalizedCode = computed(() => userCode.value.replace(/[^A-Z0-9]/g, ''))
const codeReady = computed(() => normalizedCode.value.length === 8)
const formattedCode = computed(() =>
  codeReady.value ? `${normalizedCode.value.slice(0, 4)}-${normalizedCode.value.slice(4)}` : '',
)
const remainingText = computed(
  () => `${Math.floor(remaining.value / 60)}:${String(remaining.value % 60).padStart(2, '0')}`,
)

let debounce: ReturnType<typeof setTimeout> | undefined
let ticker: ReturnType<typeof setInterval> | undefined
let deadline = 0

function stopTicker() {
  if (ticker === undefined) return
  clearInterval(ticker)
  ticker = undefined
}

function tick() {
  remaining.value = Math.max(0, Math.ceil((deadline - Date.now()) / 1000))
  if (remaining.value > 0) return
  stopTicker()
  if (!pending.value) return
  pending.value = null
  expired.value = true
}

function startTicker(seconds: number) {
  stopTicker()
  deadline = Date.now() + seconds * 1000
  remaining.value = Math.max(0, seconds)
  if (remaining.value > 0) ticker = setInterval(tick, 1000)
}

function onCodeInput(value: string) {
  userCode.value = formatCode(value)
}

async function lookup(code: string) {
  const res = await fetchDevicePending(code)
  // A newer keystroke may have superseded this request; only apply matching results.
  if (code !== formattedCode.value) return
  if (res.code === 1 && res.data) {
    if (res.data.expires_in <= 0) {
      expired.value = true
      return
    }
    startTicker(res.data.expires_in)
    pending.value = res.data
    return
  }
  lookupFailed.value = true
}

function scheduleLookup() {
  clearTimeout(debounce)
  stopTicker()
  pending.value = null
  lookupFailed.value = false
  expired.value = false
  if (!codeReady.value) return
  const code = formattedCode.value
  debounce = setTimeout(() => void lookup(code), 300)
}

async function approve() {
  if (!pending.value || approving.value) return
  approving.value = true
  try {
    const res = await fetchDeviceApprove(formattedCode.value)
    if (res.code === 1) {
      approved.value = true
      pending.value = null
      stopTicker()
      return
    }
    theToast.error(t('devicePage.approveFailed'))
  } finally {
    approving.value = false
  }
}

/** Logout keeps the code in the URL so the next account lands back on the approval. */
async function switchAccount() {
  const target = formattedCode.value ? `/device?code=${formattedCode.value}` : '/device'
  await userStore.logout()
  await router.push({ name: 'auth', query: { redirect: target } })
}

function scopeLabel(scope: string) {
  const key = ACCESS_SCOPE_LABEL_KEYS[scope]
  return key ? String(t(key)) : scope
}

// Typing a complete code is the intent to look it up — there is no separate submit step.
watch(normalizedCode, scheduleLookup)

onMounted(() => {
  if (codeReady.value) void lookup(formattedCode.value)
})

onBeforeUnmount(() => {
  clearTimeout(debounce)
  stopTicker()
})
</script>

<template>
  <main class="flex justify-center items-center min-h-screen px-4 py-10">
    <div class="w-[min(92vw,25rem)]">
      <header class="flex flex-col items-center text-center">
        <span class="text-lg italic font-bold font-serif text-[var(--color-text-muted)]">
          {{ siteName }}
        </span>
        <p class="mt-1.5 flex items-center gap-1.5 text-xs">
          <span class="text-[var(--color-text-muted)]">{{ t('devicePage.signedInAs') }}</span>
          <span class="font-medium text-[var(--color-text-secondary)]">{{ username }}</span>
          <span class="text-[var(--color-border-strong)]">·</span>
          <button
            type="button"
            class="cursor-pointer text-[var(--color-accent)] hover:opacity-80"
            @click="switchAccount"
          >
            {{ t('devicePage.switchAccount') }}
          </button>
        </p>
      </header>

      <div v-if="approved" class="mt-10 text-center">
        <p class="text-sm text-[var(--color-text-primary)]">{{ t('devicePage.approved') }}</p>
        <BaseButton class="mt-4 px-3 py-1.5 text-sm" @click="router.push({ name: 'home' })">
          {{ t('devicePage.backHome') }}
        </BaseButton>
      </div>

      <div v-else class="mt-9">
        <BaseInput
          :model-value="userCode"
          :label="t('devicePage.codeLabel')"
          :placeholder="t('devicePage.codePlaceholder')"
          class="text-center font-mono tracking-[0.25em] uppercase"
          autocomplete="one-time-code"
          spellcheck="false"
          @update:model-value="onCodeInput"
        />
        <p v-if="lookupFailed" class="mt-2 text-xs text-[var(--color-danger)]">
          {{ t('devicePage.lookupFailed') }}
        </p>
        <p v-else-if="expired" class="mt-2 text-xs text-[var(--color-danger)]">
          {{ t('devicePage.expired') }}
        </p>
        <p v-else class="mt-2 text-xs text-[var(--color-text-muted)]">
          {{ t('devicePage.codeHelp') }}
        </p>

        <section v-if="pending" class="mt-6 space-y-4">
          <div class="flex items-baseline gap-1.5 text-sm">
            <span class="shrink-0 text-[var(--color-text-muted)]">{{ t('devicePage.client') }}</span>
            <span class="font-medium text-[var(--color-text-primary)] break-all">
              {{ pending.client_name }}
            </span>
          </div>

          <div>
            <p class="text-sm text-[var(--color-text-muted)] mb-2">{{ t('devicePage.scopes') }}</p>
            <ul class="flex flex-wrap gap-1.5">
              <li
                v-for="scope in pending.scopes"
                :key="scope"
                :title="scope"
                class="text-xs px-2 py-0.5 rounded-full bg-[var(--color-accent-soft)] text-[var(--color-text-primary)]"
              >
                {{ scopeLabel(scope) }}
              </li>
            </ul>
          </div>

          <div class="flex items-center justify-between text-xs">
            <span class="text-[var(--color-text-muted)]">{{ t('devicePage.expiresIn') }}</span>
            <span class="font-mono tabular-nums text-[var(--color-text-secondary)]">
              {{ remainingText }}
            </span>
          </div>

          <button
            type="button"
            :disabled="approving || !canApprove"
            class="w-full cursor-pointer px-4 py-2.5 rounded-[var(--radius-md)] bg-[var(--color-accent)] text-[var(--dialog-confirm-text-color)] text-sm font-semibold transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            @click="approve"
          >
            {{ approving ? t('devicePage.approving') : t('devicePage.approve') }}
          </button>
          <p v-if="!canApprove" class="text-xs text-[var(--color-danger)]">
            {{ t('devicePage.adminRequired') }}
          </p>
        </section>

        <p class="mt-8 text-xs leading-relaxed text-[var(--color-text-muted)]">
          {{ t('devicePage.securityHint') }}
        </p>
      </div>
    </div>
  </main>
</template>
