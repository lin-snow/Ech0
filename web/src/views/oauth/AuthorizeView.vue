<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import {
  fetchOAuthConsent,
  submitOAuthConsent,
  type OAuthAuthorizeParams,
  type OAuthConsentContext,
} from '@/service/api/authz'
import { ACCESS_SCOPE_LABEL_KEYS } from '@/constants/accessScopes'
import { useUserStore } from '@/stores'

const { t } = useI18n()
const route = useRoute()
const userStore = useUserStore()
const { user } = storeToRefs(userStore)

function queryString(key: string): string {
  const value = route.query[key]
  return typeof value === 'string' ? value : ''
}

const params = computed<OAuthAuthorizeParams>(() => ({
  client_id: queryString('client_id'),
  redirect_uri: queryString('redirect_uri'),
  response_type: queryString('response_type') || 'code',
  scope: queryString('scope'),
  state: queryString('state'),
  resource: queryString('resource'),
  code_challenge: queryString('code_challenge'),
  code_challenge_method: queryString('code_challenge_method') || 'S256',
}))

const consent = ref<OAuthConsentContext | null>(null)
const loading = ref(true)
const submitting = ref(false)
const error = ref('')

const canApprove = computed(() => user.value?.is_admin === true)
const requestIncomplete = computed(
  () => !params.value.client_id || !params.value.redirect_uri || !params.value.code_challenge,
)

function scopeLabel(scope: string) {
  const key = ACCESS_SCOPE_LABEL_KEYS[scope]
  return key ? String(t(key)) : scope
}

async function load() {
  if (requestIncomplete.value) {
    error.value = t('oauthConsent.invalidRequest')
    loading.value = false
    return
  }
  const res = await fetchOAuthConsent(params.value)
  if (res.code === 1 && res.data) {
    consent.value = res.data
  } else {
    error.value = res.msg || t('oauthConsent.loadFailed')
  }
  loading.value = false
}

async function decide(decision: 'approve' | 'deny') {
  if (submitting.value || requestIncomplete.value) return
  submitting.value = true
  error.value = ''
  const res = await submitOAuthConsent({ ...params.value, decision })
  if (res.code === 1 && res.data?.redirect_url) {
    window.location.assign(res.data.redirect_url)
    return
  }
  error.value = res.msg || t('oauthConsent.failed')
  submitting.value = false
}

onMounted(load)
</script>

<template>
  <main class="flex justify-center items-center min-h-screen px-4 py-10">
    <div class="w-[min(92vw,28rem)]">
      <header class="flex flex-col items-center text-center gap-1">
        <h1 class="text-lg font-semibold text-[var(--color-text-primary)]">
          {{ t('oauthConsent.title') }}
        </h1>
        <p class="text-xs text-[var(--color-text-muted)]">{{ t('oauthConsent.subtitle') }}</p>
      </header>

      <p v-if="error" class="mt-6 text-xs text-[var(--color-danger)] break-words">
        {{ error }}
      </p>

      <section v-else-if="loading" class="mt-10 text-center text-sm text-[var(--color-text-muted)]">
        {{ t('oauthConsent.loading') }}
      </section>

      <section v-else-if="consent" class="mt-6 space-y-5">
        <div class="rounded-[var(--radius-md)] bg-[var(--color-bg-muted)] px-3 py-2.5">
          <p class="text-[11px] uppercase tracking-wide text-[var(--color-text-muted)]">
            {{ t('oauthConsent.client') }}
          </p>
          <p class="mt-0.5 text-sm font-medium text-[var(--color-text-primary)] break-all">
            {{ consent.client_name }}
          </p>
          <p
            v-if="consent.client_uri"
            class="mt-0.5 text-xs text-[var(--color-text-muted)] break-all"
          >
            {{ consent.client_uri }}
          </p>
        </div>

        <div>
          <p class="mb-2 text-sm text-[var(--color-text-muted)]">{{ t('oauthConsent.scopes') }}</p>
          <ul class="flex flex-wrap gap-1.5">
            <li
              v-for="scope in consent.scopes"
              :key="scope"
              :title="scope"
              class="text-xs px-2 py-0.5 rounded-full bg-[var(--color-accent-soft)] text-[var(--color-text-primary)]"
            >
              {{ scopeLabel(scope) }}
            </li>
          </ul>
        </div>

        <p class="text-xs text-[var(--color-text-muted)] break-all">
          {{ t('oauthConsent.resource') }}: {{ consent.resource }}
        </p>

        <div class="flex gap-2">
          <button
            type="button"
            :disabled="submitting"
            class="flex-1 cursor-pointer px-4 py-2.5 rounded-[var(--radius-md)] border border-[var(--color-border-subtle)] text-sm font-medium text-[var(--color-text-primary)] transition-opacity hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-60"
            @click="decide('deny')"
          >
            {{ t('oauthConsent.deny') }}
          </button>
          <button
            type="button"
            :disabled="submitting || !canApprove"
            class="flex-1 cursor-pointer px-4 py-2.5 rounded-[var(--radius-md)] bg-[var(--color-accent)] text-[var(--dialog-confirm-text-color)] text-sm font-semibold transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            @click="decide('approve')"
          >
            {{ submitting ? t('oauthConsent.approving') : t('oauthConsent.approve') }}
          </button>
        </div>

        <p v-if="!canApprove" class="text-xs text-[var(--color-danger)]">
          {{ t('oauthConsent.adminRequired') }}
        </p>
        <p class="text-xs leading-relaxed text-[var(--color-text-muted)]">
          {{ t('oauthConsent.securityHint') }}
        </p>
      </section>
    </div>
  </main>
</template>
