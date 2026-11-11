<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <TheSketchFrame v-if="hasAnyPreview" class="mx-auto my-7 w-11/12">
    <section class="px-3 py-2.5">
      <div class="mb-5 flex flex-wrap items-center justify-between gap-2">
        <span class="text-[11px] font-semibold tracking-wide text-[var(--color-text-muted)]">
          {{ t('editor.preview') }}
        </span>
        <BaseSegmented
          v-if="hasMediaPreview && hasExtensionPreview"
          v-model="activePreview"
          :options="previewOptions"
          compact
        />
      </div>

      <TheEditorMedia v-if="activePreview === 'media' && hasMediaPreview" />

      <div v-else-if="hasExtensionPreview" class="relative">
        <button
          type="button"
          class="absolute -top-2.5 -right-2.5 z-[2] w-7 h-7 flex items-center justify-center rounded-full bg-[var(--color-bg-surface)] text-[var(--color-text-secondary)] border border-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] transition-colors hover:bg-[var(--color-danger)] hover:text-white hover:border-[var(--color-danger)]"
          v-tooltip="t('editor.removeExtension')"
          :aria-label="t('editor.removeExtension')"
          @click="handleRemoveExtension"
        >
          <Close color="currentColor" class="w-4 h-4" />
        </button>
        <TheExtensionRenderer
          :key="previewKey"
          :echo="{ extension: editorStore.extensionPreview }"
        />
      </div>
    </section>
  </TheSketchFrame>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Mode } from '@/enums/enums'
import { useEditorStore } from '@/stores'
import { useBaseDialog } from '@/composables/useBaseDialog'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheSketchFrame from '@/components/common/TheSketchFrame.vue'
import Close from '@/components/icons/close.vue'
import TheEditorMedia from './TheEditorMedia.vue'

const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

const { t } = useI18n()
const { openConfirm } = useBaseDialog()
const editorStore = useEditorStore()

const hasMediaPreview = computed(() => editorStore.filesToAdd.length > 0)
const hasExtensionPreview = computed(() => editorStore.extensionPreview !== null)
const hasAnyPreview = computed(() => hasMediaPreview.value || hasExtensionPreview.value)

const activePreview = ref<string>(editorStore.currentMode === Mode.EXTEN ? 'extension' : 'media')

// Keep the switch on a usable tab: force the only available preview, and when
// both exist prefer the one for the mode the user is editing.
watch(
  [
    () => editorStore.filesToAdd.length > 0,
    () => editorStore.extensionPreview !== null,
    () => editorStore.currentMode,
  ],
  () => {
    if (!hasMediaPreview.value && hasExtensionPreview.value) {
      activePreview.value = 'extension'
      return
    }
    if (hasMediaPreview.value && !hasExtensionPreview.value) {
      activePreview.value = 'media'
      return
    }
    if (!hasMediaPreview.value) return
    activePreview.value = editorStore.currentMode === Mode.EXTEN ? 'extension' : 'media'
  },
  { immediate: true },
)

const previewOptions = computed(() => [
  { label: String(t('editor.previewFiles')), value: 'media' },
  { label: String(t('editor.previewExtension')), value: 'extension' },
])

const previewKey = computed(() => {
  const extension = editorStore.extensionPreview
  return extension ? `${extension.type}:${JSON.stringify(extension.payload)}` : ''
})

const handleRemoveExtension = () => {
  openConfirm({
    title: String(t('editor.removeExtensionConfirmTitle')),
    description: '',
    onConfirm: async () => {
      editorStore.removeExtension()
      if (editorStore.isUpdateMode) {
        await editorStore.handleAddOrUpdateEcho(true)
      }
    },
  })
}
</script>
