<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="base-textarea w-full">
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--color-text-secondary)] mb-1"
    >
      {{ label }}
    </label>

    <div class="relative">
      <textarea
        :id="id"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :rows="rows"
        :class="[
          'block w-full px-3 py-2 rounded-[var(--radius-md)] border border-[var(--input-border-color)] text-[var(--input-text-color)] focus:outline-none focus:ring-2 focus:ring-[var(--input-focus-ring-color)] transition duration-150 ease-in-out shadow-[var(--shadow-sm)] sm:text-sm',
          disabled
            ? 'bg-[var(--select-disabled-bg-color)] cursor-not-allowed opacity-70'
            : 'bg-[var(--textarea-bg-color)] hover:border-[var(--input-hover-border-color)]',
          customClass,
        ]"
        :maxlength="maxLength"
        @input="
          $emit('update:modelValue', $event.target && ($event.target as HTMLTextAreaElement).value)
        "
        v-bind="$attrs"
      ></textarea>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: string
  id?: string
  label?: string
  placeholder?: string
  rows?: number
  disabled?: boolean
  readonly?: boolean
  customClass?: string
  maxLength?: number
}>()

const customClass = props.customClass
const rows = props.rows || 3
</script>

<style scoped>
.base-textarea {
  display: flex;
  flex-direction: column;
}

textarea {
  resize: vertical;
  outline: none;
}
</style>
