<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { fetchDeviceApprove, fetchDevicePending } from '@/service/api/device'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const userCode = ref(typeof route.query.code === 'string' ? route.query.code : '')
const pending = ref<App.Api.Device.PendingInfo | null>(null)
const approved = ref(false)
const busy = ref(false)
const lookupFailed = ref(false)

/** Codes are 8 letters/digits; the hyphen is decoration, so it is added for the API call. */
const normalizedCode = computed(() => userCode.value.toUpperCase().replace(/[^A-Z0-9]/g, ''))
const codeReady = computed(() => normalizedCode.value.length === 8)
const formattedCode = computed(() =>
  codeReady.value ? `${normalizedCode.value.slice(0, 4)}-${normalizedCode.value.slice(4)}` : '',
)

let lastLookup = ''
let debounce: ReturnType<typeof setTimeout> | undefined

async function lookup(code: string) {
  lastLookup = code
  busy.value = true
  try {
    const res = await fetchDevicePending(code)
    // A newer keystroke may have superseded this request; only apply matching results.
    if (code !== formattedCode.value) return
    if (res.code === 1 && res.data) {
      pending.value = res.data
      return
    }
    lookupFailed.value = true
  } finally {
    busy.value = false
  }
}

function scheduleLookup() {
  clearTimeout(debounce)
  pending.value = null
  lookupFailed.value = false
  if (!codeReady.value) {
    lastLookup = ''
    return
  }
  const code = formattedCode.value
  if (code === lastLookup) return
  debounce = setTimeout(() => void lookup(code), 300)
}

async function approve() {
  if (!pending.value || busy.value) return
  busy.value = true
  try {
    const res = await fetchDeviceApprove(formattedCode.value)
    if (res.code === 1) {
      approved.value = true
      pending.value = null
      return
    }
    theToast.error(t('devicePage.approveFailed'))
  } finally {
    busy.value = false
  }
}

// Typing a complete code is the intent to look it up — there is no separate submit step.
watch(normalizedCode, scheduleLookup)

onMounted(() => {
  if (codeReady.value) void lookup(formattedCode.value)
})

onBeforeUnmount(() => clearTimeout(debounce))
</script>

<template>
  <div class="flex justify-center items-center min-h-screen px-4">
    <div class="w-[min(88vw,28rem)]">
      <h1
        class="text-5xl italic font-bold text-center text-[var(--color-text-muted)] mb-2 font-serif"
      >
        Ech0
      </h1>
      <h2 class="text-lg font-bold text-center text-[var(--color-text-muted)] mb-1">
        {{ t('devicePage.title') }}
      </h2>
      <p class="text-sm text-center text-[var(--color-text-secondary)] mb-6">
        {{ t('devicePage.subtitle') }}
      </p>

      <div v-if="approved" class="rounded-[var(--radius-md)] border border-[var(--input-border-color)] p-4 text-center">
        <p class="text-[var(--color-text-primary)] mb-4">{{ t('devicePage.approved') }}</p>
        <BaseButton @click="router.push({ name: 'home' })">
          {{ t('devicePage.backHome') }}
        </BaseButton>
      </div>

      <template v-else>
        <BaseInput
          v-model="userCode"
          :label="t('devicePage.codeLabel')"
          :placeholder="t('devicePage.codePlaceholder')"
          class="mb-4"
          autocomplete="one-time-code"
        />
        <div v-if="pending" class="flex justify-end gap-2">
          <BaseButton :loading="busy" @click="approve">
            {{ t('devicePage.approve') }}
          </BaseButton>
        </div>
        <p v-if="lookupFailed" class="mt-2 text-sm text-[var(--color-danger)]">
          {{ t('devicePage.lookupFailed') }}
        </p>

        <div
          v-if="pending"
          class="mt-6 rounded-[var(--radius-md)] border border-[var(--input-border-color)] p-4 space-y-3"
        >
          <p class="text-sm text-[var(--color-text-primary)]">
            <span class="text-[var(--color-text-secondary)]">{{ t('devicePage.client') }}</span>
            {{ pending.client_name }}
          </p>
          <div>
            <p class="text-sm text-[var(--color-text-secondary)] mb-1">{{ t('devicePage.scopes') }}</p>
            <ul class="flex flex-wrap gap-1">
              <li
                v-for="scope in pending.scopes"
                :key="scope"
                class="text-xs px-2 py-0.5 rounded-[var(--radius-sm)] border border-[var(--input-border-color)] text-[var(--color-text-primary)]"
              >
                {{ scope }}
              </li>
            </ul>
          </div>
          <p class="text-xs text-[var(--color-text-muted)]">
            {{ t('devicePage.pending', { seconds: pending.expires_in }) }}
          </p>
        </div>
      </template>
    </div>
  </div>
</template>
