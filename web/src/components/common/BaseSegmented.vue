<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="seg" role="tablist">
    <button
      v-for="opt in options"
      :key="opt.value"
      type="button"
      role="tab"
      :aria-selected="modelValue === opt.value"
      class="seg__btn"
      :class="{ 'seg__btn--active': modelValue === opt.value }"
      @click="emit('update:modelValue', opt.value)"
    >
      {{ opt.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue: string
  options: { label: string; value: string }[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()
</script>

<style scoped>
.seg {
  display: inline-flex;
  gap: 0.25rem;
  padding: 0.25rem;
  margin-bottom: 1rem;
  background: var(--color-bg-muted);
  border-radius: var(--radius-md);
}

.seg__btn {
  padding: 0.35rem 1.1rem;
  font-size: 0.85rem;
  font-weight: 600;
  line-height: 1.2;
  color: var(--color-text-secondary);
  border-radius: var(--radius-sm);
  transition:
    background 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.seg__btn:hover:not(.seg__btn--active) {
  color: var(--color-text-primary);
}

.seg__btn--active {
  color: var(--color-text-primary);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

@media (width < 640px) {
  .seg {
    display: flex;
    width: 100%;
  }

  .seg__btn {
    flex: 1;
    text-align: center;
    padding: 0.4rem 0.5rem;
  }
}
</style>
