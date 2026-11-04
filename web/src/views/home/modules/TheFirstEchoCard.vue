<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <TheSketchFrame class="mx-auto mb-3 w-full max-w-[34rem]">
    <section class="px-4 py-3.5" aria-live="polite">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <h2 class="text-[13px] font-semibold text-[var(--color-text-muted)]">
            {{ t('onboarding.firstEchoTitle') }}
          </h2>
          <p class="mt-1 text-pretty text-[11px] leading-4 text-[var(--color-text-secondary)]">
            {{ t('onboarding.firstEchoDesc') }}
          </p>
        </div>
        <button
          type="button"
          class="shrink-0 text-[11px] text-[var(--color-text-secondary)] underline-offset-2 hover:text-[var(--color-text-primary)] hover:underline"
          @click="emit('dismiss')"
        >
          {{ t('onboarding.dismiss') }}
        </button>
      </div>
    </section>
  </TheSketchFrame>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import TheSketchFrame from '@/components/common/TheSketchFrame.vue'

const emit = defineEmits<{
  (e: 'dismiss'): void
}>()

const { t } = useI18n()
</script>

<style scoped></style>
