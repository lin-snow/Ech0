<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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

const canLookup = computed(() => userCode.value.trim().length >= 8 && !busy.value)

async function lookup() {
  if (!canLookup.value) return
  busy.value = true
  try {
    const res = await fetchDevicePending(userCode.value.trim())
    if (res.code === 1 && res.data) {
      pending.value = res.data
      return
    }
    pending.value = null
    theToast.error(t('devicePage.lookupFailed'))
  } finally {
    busy.value = false
  }
}

async function approve() {
  if (!pending.value || busy.value) return
  busy.value = true
  try {
    const res = await fetchDeviceApprove(userCode.value.trim())
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

onMounted(() => {
  if (userCode.value.trim() !== '') void lookup()
})
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
        <div class="flex justify-end gap-2">
          <BaseButton :loading="busy" :disabled="!canLookup" @click="lookup">
            {{ t('devicePage.lookup') }}
          </BaseButton>
          <BaseButton v-if="pending" :loading="busy" @click="approve">
            {{ t('devicePage.approve') }}
          </BaseButton>
        </div>

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
