<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div v-if="hasAnyPreview">
    <div v-if="hasMediaPreview && hasExtensionPreview" class="flex justify-center">
      <BaseSegmented v-model="activePreview" :options="previewOptions" />
    </div>

    <TheEditorMedia v-if="activePreview === 'media' && hasMediaPreview" />
    <TheSketchFrame v-else-if="hasExtensionPreview" class="w-11/12 mx-auto my-7">
      <section class="px-3 py-2.5">
        <TheExtensionRenderer
          :key="previewKey"
          :echo="{ extension: editorStore.extensionPreview }"
        />
      </section>
    </TheSketchFrame>
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Mode } from '@/enums/enums'
import { useEditorStore } from '@/stores'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheSketchFrame from '@/components/common/TheSketchFrame.vue'
import TheEditorMedia from './TheEditorMedia.vue'

const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

const { t } = useI18n()
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
</script>
