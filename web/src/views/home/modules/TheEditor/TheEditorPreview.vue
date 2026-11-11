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
          v-if="previewOptions.length > 1"
          v-model="activePreview"
          :options="previewOptions"
          compact
        />
      </div>

      <article
        v-if="activePreview === 'content' && hasContentPreview"
        data-test="content-preview"
        class="markdown-preview-card preview-content-card max-h-[28rem]"
      >
        <div class="preview-card-body">
          <MarkdownRenderer :content="editorStore.echoToAdd.content" />
        </div>
      </article>

      <TheEditorMedia v-else-if="activePreview === 'media' && hasMediaPreview" />

      <div v-else-if="activePreview === 'extension' && hasExtensionPreview" class="relative">
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
const MarkdownRenderer = defineAsyncComponent(
  () => import('@/editor/components/MarkdownRenderer.vue'),
)

const { t } = useI18n()
const { openConfirm } = useBaseDialog()
const editorStore = useEditorStore()

type PreviewTab = 'content' | 'media' | 'extension'

const hasContentPreview = computed(() => !!editorStore.echoToAdd.content?.trim())
const hasMediaPreview = computed(() => editorStore.filesToAdd.length > 0)
const hasExtensionPreview = computed(() => editorStore.extensionPreview !== null)
const hasAnyPreview = computed(
  () => hasContentPreview.value || hasMediaPreview.value || hasExtensionPreview.value,
)

const preferredPreview = (): PreviewTab => {
  if (editorStore.currentMode === Mode.EXTEN && hasExtensionPreview.value) return 'extension'
  if (editorStore.currentMode === Mode.Media && hasMediaPreview.value) return 'media'
  if (hasMediaPreview.value) return 'media'
  if (hasExtensionPreview.value) return 'extension'
  return 'content'
}

const activePreview = ref<PreviewTab>(preferredPreview())

// Keep the switch on a usable tab: new files/extensions take focus, a single
// available preview is forced, and manual choices survive until the draft or
// the editing mode changes.
watch(
  [
    () => hasContentPreview.value,
    () => hasMediaPreview.value,
    () => hasExtensionPreview.value,
    () => editorStore.currentMode,
  ],
  ([content, media, extension, mode], previous) => {
    const available: PreviewTab[] = []
    if (content) available.push('content')
    if (media) available.push('media')
    if (extension) available.push('extension')
    if (available.length === 0) return
    if (available.length === 1) {
      activePreview.value = available[0]!
      return
    }

    const previousSources = previous && previous.length > 0 ? previous : null
    const [, prevMedia, prevExtension] = previousSources ?? []
    if (previousSources) {
      if (media && !prevMedia) {
        activePreview.value = 'media'
        return
      }
      if (extension && !prevExtension) {
        activePreview.value = 'extension'
        return
      }
    }
    if (previousSources && mode !== previousSources[3]) {
      activePreview.value = preferredPreview()
      return
    }
    if (!available.includes(activePreview.value)) {
      activePreview.value = preferredPreview()
    }
  },
  { immediate: true },
)

const previewOptions = computed(() => {
  const options: { label: string; value: string }[] = []
  if (hasContentPreview.value) {
    options.push({ label: String(t('editor.previewContent')), value: 'content' })
  }
  if (hasMediaPreview.value) {
    options.push({ label: String(t('editor.previewFiles')), value: 'media' })
  }
  if (hasExtensionPreview.value) {
    options.push({ label: String(t('editor.previewExtension')), value: 'extension' })
  }
  return options
})

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

<style scoped>
/* The dashed preview frame is the only container here; a second shadow on the
   text card makes it read as a floating dialog. */
.preview-content-card {
  box-shadow: none;
}
</style>
