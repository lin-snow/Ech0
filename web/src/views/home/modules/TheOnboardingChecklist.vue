<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <TheSketchFrame v-if="visible" class="mx-auto my-4 w-full max-w-[34rem]">
    <section class="px-4 py-3.5" aria-live="polite">
      <header class="mb-2.5 flex items-baseline justify-between gap-3">
        <h2 class="text-[13px] font-semibold text-[var(--color-text-muted)]">
          {{ t('onboarding.title') }}
        </h2>
        <div class="flex items-baseline gap-2.5">
          <span class="text-[11px] tabular-nums text-[var(--color-text-secondary)]">
            {{ doneCount }}/{{ steps.length }}
          </span>
          <button
            type="button"
            class="text-[11px] text-[var(--color-text-secondary)] underline-offset-2 hover:text-[var(--color-text-primary)] hover:underline"
            @click="dismiss"
          >
            {{ t('onboarding.hide') }}
          </button>
        </div>
      </header>

      <ul class="space-y-0.5">
        <li
          v-for="step in steps"
          :key="step.id"
          class="flex items-start gap-2.5 rounded-md px-1 py-1.5"
        >
          <span
            class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full"
            :class="
              step.done
                ? 'bg-[var(--color-success,#16a34a)]'
                : 'border border-[var(--color-border-strong)]'
            "
            aria-hidden="true"
          >
            <svg
              v-if="step.done"
              viewBox="0 0 24 24"
              class="h-2.5 w-2.5 text-white"
              fill="none"
              stroke="currentColor"
              stroke-width="3.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M4.5 12.5 9.5 17.5 19.5 6.5" />
            </svg>
          </span>

          <div class="min-w-0 flex-1">
            <div class="flex min-h-5 items-center justify-between gap-2">
              <p
                class="min-w-0 text-xs font-medium leading-5"
                :class="
                  step.done
                    ? 'text-[var(--color-text-secondary)]'
                    : 'text-[var(--color-text-primary)]'
                "
              >
                {{ step.title }}
              </p>

              <span
                v-if="step.done"
                class="shrink-0 text-[11px] leading-none text-[var(--color-text-secondary)]"
              >
                {{ t('onboarding.done') }}
              </span>
              <button
                v-else
                type="button"
                class="h-5 shrink-0 whitespace-nowrap rounded-full border border-[var(--color-border-strong)] px-2 text-[11px] leading-none text-[var(--color-text-muted)] transition-colors hover:border-[var(--color-accent)] hover:text-[var(--color-accent)]"
                @click="openStep(step.id)"
              >
                {{ step.action }}
              </button>
            </div>

            <p
              v-if="!step.done"
              class="text-pretty text-[11px] leading-4 text-[var(--color-text-secondary)]"
            >
              {{ step.description }}
            </p>
          </div>
        </li>
      </ul>
    </section>
  </TheSketchFrame>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useEchoStore, useSettingStore } from '@/stores'
import {
  resolveSiteOnboardingSteps,
  type SiteOnboardingStepId,
} from '@/composables/useSiteOnboarding'
import { useSiteChecklist } from '@/composables/useSiteChecklist'
import TheSketchFrame from '@/components/common/TheSketchFrame.vue'

const router = useRouter()
const { t } = useI18n()
const echoStore = useEchoStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)
const { visible, dismiss } = useSiteChecklist()

type StepCopy = { title: string; description: string; action: string }

const stepCopy = computed<Record<SiteOnboardingStepId, StepCopy>>(() => ({
  publish: {
    title: String(t('onboarding.publishTitle')),
    description: String(t('onboarding.publishDesc')),
    action: String(t('onboarding.actionPublish')),
  },
  siteTitle: {
    title: String(t('onboarding.siteTitleTitle')),
    description: String(t('onboarding.siteTitleDesc')),
    action: String(t('onboarding.actionSiteTitle')),
  },
  siteUrl: {
    title: String(t('onboarding.siteURLTitle')),
    description: String(t('onboarding.siteURLDesc')),
    action: String(t('onboarding.actionSiteURL')),
  },
}))

const steps = computed(() =>
  resolveSiteOnboardingSteps({
    echoTotal: echoStore.total,
    siteTitle: SystemSetting.value.site_title ?? '',
    serverURL: SystemSetting.value.server_url ?? '',
  }).map((step) => ({
    id: step.id,
    done: step.done,
    title: stepCopy.value[step.id].title,
    description: stepCopy.value[step.id].description,
    action: stepCopy.value[step.id].action,
  })),
)

const doneCount = computed(() => steps.value.filter((step) => step.done).length)

const openStep = (id: SiteOnboardingStepId) => {
  if (id === 'publish') {
    void router.push({ name: 'home', query: { tab: 'publish' } })
    return
  }
  void router.push({ name: 'panel-setting' })
}
</script>

<style scoped></style>
