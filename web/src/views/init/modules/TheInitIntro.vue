<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <section class="mb-3 text-center">
    <p class="text-[10px] tracking-[0.16em] uppercase text-[var(--color-text-secondary)] mb-2">
      {{ t('init.introKicker') }}
    </p>
    <h1
      class="text-5xl italic font-bold text-[var(--color-text-muted)] mb-2 font-serif leading-none"
    >
      Ech0
    </h1>
    <h2 class="text-sm font-semibold text-[var(--color-text-muted)] mb-1">
      {{ t('init.introHeading') }}
    </h2>
    <p class="text-xs text-[var(--color-text-secondary)] leading-5">
      {{ t('init.introDescription') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<style scoped></style>
