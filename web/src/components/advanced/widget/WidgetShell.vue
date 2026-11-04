<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="widget-shell" :class="{ 'widget-shell--fluid': fluid }">
    <div class="widget-head">
      <span v-if="$slots.icon" class="widget-icon" aria-hidden="true">
        <slot name="icon" />
      </span>
      <h2 class="widget-title">{{ title }}</h2>
      <span v-if="meta" class="widget-meta" :title="metaTitle">{{ meta }}</span>
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    meta?: string
    metaTitle?: string
    fluid?: boolean
  }>(),
  {
    meta: undefined,
    metaTitle: undefined,
    fluid: false,
  },
)
</script>

<style scoped>
.widget-shell {
  width: 100%;
  max-width: 19rem;
  margin-inline: auto;
  padding: 0.875rem;
}

.widget-shell--fluid {
  max-width: none;
}

.widget-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 1.5rem;
  margin-bottom: 0.625rem;
}

.widget-icon {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.125rem;
  height: 1.125rem;
  color: var(--color-text-muted);
}

.widget-icon :slotted(svg) {
  display: block;
  width: 1.125rem;
  height: 1.125rem;
}

.widget-title {
  margin: 0;
  font-family: var(--font-family-display);
  font-size: 1.125rem;
  font-weight: 600;
  line-height: 1.2;
  letter-spacing: -0.02em;
  color: var(--color-text-primary);
}

.widget-meta {
  margin-left: auto;
  font-size: 0.6875rem;
  font-weight: 500;
  line-height: 1.2;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
</style>
