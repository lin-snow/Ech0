<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="editor-media-panel">
    <h2 class="text-[var(--color-text-secondary)] font-bold my-2">
      {{ t('editor.mediaSectionTitle') }}
    </h2>

    <div class="mb-3 flex items-center gap-2">
      <span class="shrink-0 whitespace-nowrap text-[var(--color-text-secondary)]">{{
        t('editor.mediaTypeLabel')
      }}</span>
      <BaseSegmented
        class="media-type-seg"
        :modelValue="effectiveCategory"
        :options="mediaTypeOptions"
        @update:modelValue="onSelectCategory"
      />
    </div>

    <div v-if="effectiveCategory === FILE_CATEGORY.IMAGE" class="mb-2 flex items-center gap-2">
      <span class="text-[var(--color-text-secondary)]">{{ t('editor.imageLayout') }}</span>
      <BaseSelect
        v-model="echoToAdd.layout"
        :options="layoutOptions"
        class="w-36 h-7"
        :placeholder="t('editor.imageLayoutPlaceholder')"
      />
    </div>

    <div v-if="effectiveCategory === FILE_CATEGORY.IMAGE" class="mb-3 flex items-center">
      <span class="text-[var(--color-text-secondary)]">{{ t('editor.imageSmartCompress') }}</span>
      <BaseSwitch v-model="enableCompressor" />
    </div>

    <div class="my-1">
      <TheUploader
        v-if="!singleClipFull"
        :key="effectiveCategory"
        :fileCategory="effectiveCategory"
        :allowedFileTypes="acceptedTypes"
        :EnableCompressor="enableCompressor && effectiveCategory === FILE_CATEGORY.IMAGE"
        :maxFileSize="maxFileSize"
        :maxFiles="maxFiles"
      />

      <div v-else class="text-[var(--color-text-muted)] text-sm px-1 py-2">
        {{ t('editor.singleMediaLimit') }}
      </div>

      <div class="flex items-center gap-2">
        <BaseInput
          v-model="fileToAdd.url"
          class="rounded-lg h-auto flex-1"
          :placeholder="t('editor.imageUrlPlaceholder')"
        />
        <BaseButton
          v-if="fileToAdd.url != '' && !singleClipFull"
          :icon="Addmore"
          class="w-8 h-8 sm:w-8 sm:h-8 rounded-md shrink-0"
          @click="editorStore.handleAddMoreFile"
          :tooltip="t('editor.addMoreImages')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useEditorStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ImageLayout } from '@/enums/enums'
import { FILE_CATEGORY, type FileCategory } from '@/constants/file'
import Addmore from '@/components/icons/addmore.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import TheUploader from '@/components/advanced/TheUploader.vue'
import { useI18n } from 'vue-i18n'

const IMAGE_MAX_FILE_SIZE = 20 * 1024 * 1024
const AUDIO_MAX_FILE_SIZE = 20 * 1024 * 1024
const VIDEO_MAX_FILE_SIZE = 64 * 1024 * 1024

const editorStore = useEditorStore()
const { fileToAdd, filesToAdd, echoToAdd, effectiveCategory } = storeToRefs(editorStore)

const singleClipFull = computed(
  () =>
    (effectiveCategory.value === FILE_CATEGORY.AUDIO ||
      effectiveCategory.value === FILE_CATEGORY.VIDEO) &&
    filesToAdd.value.length > 0,
)
const enableCompressor = ref<boolean>(false)
const { t } = useI18n()

const mediaTypeOptions = computed<{ label: string; value: FileCategory }[]>(() => [
  { label: String(t('editor.mediaTypeImage')), value: FILE_CATEGORY.IMAGE },
  { label: String(t('editor.mediaTypeAudio')), value: FILE_CATEGORY.AUDIO },
  { label: String(t('editor.mediaTypeVideo')), value: FILE_CATEGORY.VIDEO },
])
const onSelectCategory = (v: string) => editorStore.setSelectedCategory(v as FileCategory)

const acceptedTypes = computed<string[]>(() => {
  switch (effectiveCategory.value) {
    case FILE_CATEGORY.AUDIO:
      return ['audio/*']
    case FILE_CATEGORY.VIDEO:
      return ['video/mp4']
    default:
      return ['image/*']
  }
})
const maxFileSize = computed<number>(() => {
  switch (effectiveCategory.value) {
    case FILE_CATEGORY.AUDIO:
      return AUDIO_MAX_FILE_SIZE
    case FILE_CATEGORY.VIDEO:
      return VIDEO_MAX_FILE_SIZE
    default:
      return IMAGE_MAX_FILE_SIZE
  }
})

const IMAGE_MAX_FILES = 9
const maxFiles = computed<number>(() =>
  effectiveCategory.value === FILE_CATEGORY.IMAGE ? IMAGE_MAX_FILES : 1,
)

const layoutOptions = computed(() => [
  { label: String(t('editor.layoutWaterfall')), value: ImageLayout.WATERFALL },
  { label: String(t('editor.layoutGrid')), value: ImageLayout.GRID },
  { label: String(t('editor.layoutCarousel')), value: ImageLayout.CAROUSEL },
  { label: String(t('editor.layoutHorizontal')), value: ImageLayout.HORIZONTAL },
  { label: String(t('editor.layoutStack')), value: ImageLayout.STACK },
])
</script>

<style scoped>
.editor-media-panel {
  margin: 0.75rem 0;
  padding: 0.75rem;
  border: 1px dashed var(--color-border-strong);
  border-radius: var(--radius-xs);
}

.editor-media-panel .media-type-seg {
  margin-bottom: 0;
  padding: 0.15rem;
}

.editor-media-panel .media-type-seg :deep(.seg__btn) {
  padding: 0.25rem 0.7rem;
  font-size: 0.8rem;
}
</style>
