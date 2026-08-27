<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <TransitionRoot :show="isOpen" as="template">
    <Dialog @close="handleDialogClose" class="relative z-5000">
      <TransitionChild
        enter="duration-300 ease-out"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="duration-200 ease-in"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
      </TransitionChild>

      <div class="fixed inset-0 flex items-center justify-center p-4">
        <TransitionChild
          enter="duration-300 ease-out"
          enter-from="opacity-0 scale-95"
          enter-to="opacity-100 scale-100"
          leave="duration-200 ease-in"
          leave-from="opacity-100 scale-100"
          leave-to="opacity-0 scale-95"
        >
          <DialogPanel
            class="w-full max-w-sm rounded-[var(--radius-lg)] bg-[var(--dialog-bg-color)] p-6 shadow-[var(--shadow-md)] ring-1 ring-inset ring-[var(--color-border-subtle)]"
          >
            <DialogTitle class="text-base font-semibold text-[var(--dialog-title-color)]">
              {{ title }}
            </DialogTitle>
            <DialogDescription class="mt-2 text-sm text-[var(--dialog-text-color)] leading-relaxed">
              {{ description }}
            </DialogDescription>

            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="cancel"
                class="cursor-pointer px-3 py-2 rounded-[var(--radius-md)] bg-[var(--dialog-cancel-btn-bg-color)] shadow-[var(--shadow-sm)] ring-1 ring-inset ring-[var(--color-border-subtle)] text-[var(--dialog-btn-text-color)] hover:text-[var(--dialog-hover-text-color)]"
              >
                {{ t('commonUi.cancel') }}
              </button>
              <button
                @click="confirm"
                class="cursor-pointer px-3 py-2 rounded-[var(--radius-md)] bg-[var(--dialog-confirm-btn-bg-color)] text-[var(--dialog-confirm-text-color)] shadow-[var(--shadow-sm)] hover:opacity-90"
              >
                {{ t('commonUi.confirm') }}
              </button>
            </div>
          </DialogPanel>
        </TransitionChild>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Dialog,
  DialogPanel,
  DialogTitle,
  DialogDescription,
  TransitionChild,
  TransitionRoot,
} from '@headlessui/vue'

defineProps({
  title: String,
  description: String,
})

const emit = defineEmits(['confirm', 'cancel'])
const { t } = useI18n()

const isOpen = ref(false)

function open() {
  isOpen.value = true
}

function close() {
  isOpen.value = false
}

function handleDialogClose() {
  emit('cancel')
  close()
}

function confirm() {
  emit('confirm')
  close()
}

function cancel() {
  emit('cancel')
  close()
}

defineExpose({ open })
</script>
