<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="base-input w-full">
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--color-text-secondary)] mb-1"
    >
      {{ label }}
    </label>

    <div class="flex items-center">
      <slot name="prefix" />
      <input
        :id="id"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :class="[
          'block px-3 py-2 rounded-[var(--radius-md)] border border-[var(--input-border-color)] text-[var(--input-text-color)] bg-[var(--input-bg-color)]! focus:outline-none focus:ring-2 focus:ring-[var(--input-focus-ring-color)] transition duration-150 ease-in-out shadow-[var(--shadow-sm)] sm:text-sm w-full',
          disabled
            ? 'cursor-not-allowed opacity-70 text-[var(--color-text-muted)]'
            : 'hover:border-[var(--input-hover-border-color)]',
          customClass,
        ]"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        v-bind="$attrs"
      />
      <slot name="suffix" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: string | number | null | undefined
  id?: string
  label?: string
  placeholder?: string
  type?: string
  disabled?: boolean
  readonly?: boolean
  class?: string
}>()

const customClass = props.class
const type = props.type || 'text'
</script>

<style scoped>
.base-input {
  display: flex;
  flex-direction: column;
}

input {
  outline: none;
}
</style>
