<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <section class="no-permission" role="status" aria-live="polite">
    <span class="no-permission__icon" aria-hidden="true">
      <svg
        viewBox="0 0 24 24"
        width="1em"
        height="1em"
        fill="none"
        stroke="currentColor"
        stroke-width="1.6"
      >
        <rect x="4.5" y="10.5" width="15" height="9.5" rx="2.2" />
        <path d="M8 10.5V7.8a4 4 0 0 1 8 0v2.7" />
        <circle cx="12" cy="15.4" r="1.1" fill="currentColor" stroke="none" />
      </svg>
    </span>
    <h2 class="no-permission__title">{{ t('common.noPermission.title') }}</h2>
    <p class="no-permission__description">{{ t('common.noPermission.description') }}</p>
    <div v-if="$slots.default" class="no-permission__actions">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<style scoped>
.no-permission {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 4rem 1.5rem;
  text-align: center;
}

.no-permission__icon {
  display: grid;
  place-items: center;
  width: 3rem;
  height: 3rem;
  border-radius: 9999px;
  background: var(--color-bg-muted);
  color: var(--color-text-muted);
  font-size: 1.5rem;
}

.no-permission__title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.no-permission__description {
  margin: 0;
  max-width: 28rem;
  font-size: 0.875rem;
  line-height: 1.6;
  color: var(--color-text-secondary);
}

.no-permission__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
  margin-top: 0.5rem;
}
</style>
