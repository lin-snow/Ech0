<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="base-combobox">
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--color-text-secondary)] mb-1"
    >
      {{ label }}
    </label>

    <Combobox
      :modelValue="internalValue"
      :by="by"
      :multiple="multiple"
      :nullable="true"
      @update:model-value="onSelect"
    >
      <div class="relative">
        <div
          :class="[
            'flex items-center px-3 py-2 rounded-[var(--radius-md)] bg-[var(--input-bg-color)] border border-[var(--combobox-border-color)] shadow-[var(--shadow-sm)] focus-within:ring-2 focus-within:ring-[var(--input-focus-ring-color)] transition duration-150 ease-in-out',
            wrapperClass,
          ]"
          @focusout="onBlurOutside"
          @focusin="onFocusInput"
          @mousedown="onFocusInput"
        >
          <ComboboxInput
            :displayValue="displayValue"
            :placeholder="placeholder"
            @focus="onFocusInput"
            @click="onFocusInput"
            @input="onInputChange"
            :class="[
              'flex-1 min-w-0 bg-transparent outline-none sm:text-sm text-[var(--input-text-color)]',
              inputClass,
            ]"
          />

          <slot name="suffix">
            <ComboboxButton
              class="ml-1 text-[var(--color-text-muted)]"
              aria-label="展开选项列表"
              v-tooltip="'展开选项列表'"
            >
            </ComboboxButton>
          </slot>
        </div>

        <Transition
          enter="transition ease-out duration-100"
          enter-from="opacity-0 translate-y-1"
          enter-to="opacity-100 translate-y-0"
          leave="transition ease-in duration-75"
          leave-from="opacity-100 translate-y-0"
          leave-to="opacity-0 translate-y-1"
        >
          <ComboboxOptions
            static
            v-if="dropdownOpen && (filteredOptions.length > 0 || allowCreate)"
            class="w-full absolute z-20 mt-1 max-h-44 overflow-y-auto overscroll-contain rounded-[var(--radius-md)] bg-[var(--combobox-bg-color)] py-1 text-sm shadow-[var(--shadow-md)] ring-1 ring-[var(--combobox-border-color)] focus:outline-none"
          >
            <ComboboxOption
              v-for="item in filteredOptions"
              :key="getOptionLabel(item) || String(item)"
              :value="item"
              @mousedown="isUserClicking = true"
              v-slot="{ active }"
            >
              <div
                :class="[
                  'w-full truncate cursor-pointer select-none px-3 py-1.5 text-sm',
                  active
                    ? 'bg-[var(--select-label-hover-bg-color)] text-[var(--combobox-hover-text-color)]'
                    : 'text-[var(--color-text-secondary)]',
                ]"
              >
                <slot name="option" :option="item">{{ getOptionLabel(item) }}</slot>
              </div>
            </ComboboxOption>
          </ComboboxOptions>
        </Transition>
      </div>
    </Combobox>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Combobox,
  ComboboxInput,
  ComboboxOptions,
  ComboboxOption,
  ComboboxButton,
} from '@headlessui/vue'

type ClassValue = string | string[] | Record<string, boolean | number | string>

const props = defineProps<{
  modelValue: string | object | null | (string | object)[]
  options: string[] | object[]
  label?: string
  id?: string
  placeholder?: string
  by?: string | ((a: object | string, b: object | string) => boolean)
  labelField?: string
  allowCreate?: boolean
  multiple?: boolean
  inputClass?: ClassValue
  wrapperClass?: ClassValue
}>()

const emit = defineEmits(['update:modelValue', 'create'])

const query = ref('')
const dropdownOpen = ref(false)
const internalValue = ref(props.modelValue)
const labelField = props.labelField || 'name'
const allowCreate = props.allowCreate ?? false
const multiple = props.multiple ?? false
const isUserClicking = ref(false)

watch(
  () => props.modelValue,
  (val) => {
    internalValue.value = val
  },
)
watch(internalValue, (val) => {
  emit('update:modelValue', val)
})

const onSelect = (val: object | string) => {
  const selectedLabel = getOptionLabel(val)

  if (isUserClicking.value) {
    internalValue.value = val
    query.value = selectedLabel
    dropdownOpen.value = false
    isUserClicking.value = false
    return
  }

  if (query.value && selectedLabel.toLowerCase() === query.value.toLowerCase()) {
    internalValue.value = val
    query.value = selectedLabel
    return
  }
}

const onInputChange = (e: Event) => {
  const value = ((e.target as HTMLInputElement | null)?.value ?? '').trim()
  query.value = value

  if (value === '') {
    internalValue.value = multiple ? [] : ''
    emit('update:modelValue', internalValue.value)
    dropdownOpen.value = true
    return
  }

  const matched = props.options.find(
    (option) => getOptionLabel(option).toLowerCase() === value.toLowerCase(),
  )
  if (matched) {
    internalValue.value = matched
    emit('update:modelValue', matched)
  } else {
    internalValue.value = value
    emit('create', value)
    emit('update:modelValue', internalValue.value)
  }

  dropdownOpen.value = true
}

const onFocusInput = () => {
  dropdownOpen.value = true
}

const onBlurOutside = (e: FocusEvent) => {
  const currentTarget = e.currentTarget as HTMLElement
  if (!currentTarget.contains(e.relatedTarget as Node)) {
    dropdownOpen.value = false
    if (query.value.trim() === '') {
      internalValue.value = multiple ? [] : ''
      emit('update:modelValue', internalValue.value)
    }
  }
}

const getOptionLabel = (option: unknown): string => {
  if (option == null) return ''
  if (typeof option === 'object' && !Array.isArray(option)) {
    const record = option as Record<string, unknown>
    const candidate = record[labelField]
    if (candidate != null) return String(candidate)
  }
  return String(option ?? '')
}

const normalizedQuery = computed(() => query.value.trim().toLowerCase())

const filteredOptions = computed(() => {
  if (!normalizedQuery.value) return props.options
  const lowerQuery = normalizedQuery.value
  return props.options.filter((option) => getOptionLabel(option).toLowerCase().includes(lowerQuery))
})

const displayValue = (item: unknown) => {
  if (Array.isArray(item)) return item.map((i) => getOptionLabel(i)).join(', ')
  return getOptionLabel(item)
}
</script>

<style scoped>
.base-combobox {
  display: flex;
  flex-direction: column;
}
</style>
