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
} from '@/service/api/oauthserver'
import { ACCESS_SCOPE_GROUPS } from '@/constants/accessScopes'
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
/** Load failure replaces the card; action failure keeps it on screen. */
const loadError = ref('')
const actionError = ref('')

const canApprove = computed(() => user.value?.is_admin === true)
const requestIncomplete = computed(
  () => !params.value.client_id || !params.value.redirect_uri || !params.value.code_challenge,
)

const scopeCount = computed(() => consent.value?.scopes.length ?? 0)

type ScopeGroupView = {
  key: string
  label: string
  items: { value: string; label: string }[]
}

const scopeGroups = computed<ScopeGroupView[]>(() => {
  const requested = new Set(consent.value?.scopes ?? [])
  const known = new Set<string>()
  const groups: ScopeGroupView[] = []

  for (const group of ACCESS_SCOPE_GROUPS) {
    const items = group.items
      .filter((item) => requested.has(item.value))
      .map((item) => {
        known.add(item.value)
        return { value: item.value, label: String(t(item.labelKey)) }
      })
    if (items.length > 0) {
      groups.push({ key: group.labelKey, label: String(t(group.labelKey)), items })
    }
  }

  const other = [...requested]
    .filter((scope) => !known.has(scope))
    .map((scope) => ({ value: scope, label: scope }))
  if (other.length > 0) {
    groups.push({ key: 'other', label: String(t('oauthConsent.otherScopes')), items: other })
  }

  return groups
})

async function load() {
  if (requestIncomplete.value) {
    loadError.value = t('oauthConsent.invalidRequest')
    loading.value = false
    return
  }
  try {
    const res = await fetchOAuthConsent(params.value)
    if (res.code === 1 && res.data) {
      consent.value = res.data
    } else {
      loadError.value = res.msg || t('oauthConsent.loadFailed')
    }
  } catch (error) {
    loadError.value =
      error instanceof Error && error.message ? error.message : t('oauthConsent.loadFailed')
  } finally {
    loading.value = false
  }
}

async function decide(decision: 'approve' | 'deny') {
  if (submitting.value || requestIncomplete.value) return
  submitting.value = true
  actionError.value = ''
  try {
    const res = await submitOAuthConsent({ ...params.value, decision })
    if (res.code === 1 && res.data?.redirect_url) {
      window.location.assign(res.data.redirect_url)
      return
    }
    actionError.value = res.msg || t('oauthConsent.failed')
  } catch (error) {
    actionError.value =
      error instanceof Error && error.message ? error.message : t('oauthConsent.failed')
  } finally {
    // A failed submit must leave both buttons usable for a retry.
    submitting.value = false
  }
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

      <p v-if="loadError" class="mt-6 text-xs text-[var(--color-danger)] break-words">
        {{ loadError }}
      </p>

      <section v-if="loading" class="mt-10 text-center text-sm text-[var(--color-text-muted)]">
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
          <div class="mb-2 flex items-baseline justify-between gap-2">
            <p class="text-sm text-[var(--color-text-muted)]">{{ t('oauthConsent.scopes') }}</p>
            <p class="text-[11px] text-[var(--color-text-muted)]">
              {{
                t('oauthConsent.scopeSummary', {
                  count: scopeCount,
                  groups: scopeGroups.length,
                })
              }}
            </p>
          </div>
          <div class="space-y-1.5">
            <details
              v-for="(group, index) in scopeGroups"
              :key="group.key"
              class="scope-group rounded-[var(--radius-md)] border border-[var(--color-border-subtle)] overflow-hidden"
              :open="index === 0"
            >
              <summary
                class="flex cursor-pointer list-none items-center justify-between gap-3 px-3 py-2.5 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-bg-muted)]/60"
              >
                <span class="flex min-w-0 items-center gap-2">
                  <span class="truncate font-medium">{{ group.label }}</span>
                  <span
                    class="shrink-0 rounded-full bg-[var(--color-bg-muted)] px-1.5 py-0.5 text-[11px] text-[var(--color-text-muted)]"
                  >
                    {{ group.items.length }}
                  </span>
                </span>
                <svg
                  class="scope-group__chevron h-4 w-4 shrink-0 text-[var(--color-text-muted)]"
                  viewBox="0 0 20 20"
                  fill="none"
                  aria-hidden="true"
                >
                  <path
                    d="M5 7.5 10 12.5 15 7.5"
                    stroke="currentColor"
                    stroke-width="1.75"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </summary>
              <ul class="flex flex-wrap gap-1.5 px-3 pb-3">
                <li
                  v-for="item in group.items"
                  :key="item.value"
                  :title="item.value"
                  class="text-xs px-2 py-0.5 rounded-full bg-[var(--color-accent-soft)] text-[var(--color-text-primary)]"
                >
                  {{ item.label }}
                </li>
              </ul>
            </details>
          </div>
        </div>

        <p class="text-xs text-[var(--color-text-muted)] break-all">
          {{ t('oauthConsent.resource') }}:
          <span class="font-medium text-[var(--color-text-primary)]">
            {{ consent.resource_name || consent.resource }}
          </span>
          <span v-if="consent.resource_name" class="mt-0.5 block">{{ consent.resource }}</span>
        </p>

        <p v-if="actionError" class="text-xs text-[var(--color-danger)] break-words">
          {{ actionError }}
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
