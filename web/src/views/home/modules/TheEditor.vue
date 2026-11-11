<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div v-if="ShowEditor" class="editor-shell">
    <div class="editor-shell__inner">
      <div class="editor-shell__body">
        <TheMdEditor v-if="currentMode === Mode.ECH0" class="rounded-[var(--radius-xs)]" />
        <TheMediaEditor v-if="currentMode === Mode.Media" />
        <TheModePanel v-if="currentMode === Mode.Panel" />
        <TheExtensionEditor v-if="currentMode === Mode.EXTEN" />
        <TheTagsManager v-if="currentMode === Mode.TagManage" />
      </div>

      <TheEditorButtons />
      <TheEditorPreview />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEditorPreview from './TheEditor/TheEditorPreview.vue'
import TheEditorButtons from './TheEditor/TheEditorButtons.vue'

import { theToast } from '@/utils/toast'
import { defineAsyncComponent, watch } from 'vue'
import { useEditorStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Mode } from '@/enums/enums'
import { useI18n } from 'vue-i18n'

const TheMdEditor = defineAsyncComponent(() => import('@/components/advanced/md/TheMdEditor.vue'))
const TheModePanel = defineAsyncComponent(() => import('./TheEditor/TheModePanel.vue'))
const TheMediaEditor = defineAsyncComponent(() => import('./TheEditor/TheMediaEditor.vue'))
const TheExtensionEditor = defineAsyncComponent(() => import('./TheEditor/TheExtensionEditor.vue'))
const TheTagsManager = defineAsyncComponent(() => import('./TheEditor/TheTagsManager.vue'))

const editorStore = useEditorStore()
const { ShowEditor, currentMode, videoURL, extensionToAdd } = storeToRefs(editorStore)
const { t } = useI18n()

watch(
  () => videoURL.value,
  (newVal) => {
    if (newVal.length > 0) {
      const bvRegex = /(BV[0-9A-Za-z]{10})/
      const ytRegex =
        /(?:https?:\/\/(?:www\.)?)?(?:youtu\.be\/|youtube\.com\/(?:(?:watch)?\?(?:.*&)?v(?:i)?=|(?:embed|shorts|live)\/))([\w-]+)/
      let match = newVal.match(bvRegex)
      if (match) {
        extensionToAdd.value.extension = match[0]
      } else {
        match = newVal.match(ytRegex)
        if (match) {
          extensionToAdd.value.extension = match[1] ?? ''
        } else {
          theToast.error(String(t('editor.videoShareLinkInvalid')))
        }
      }
    }
  },
)
</script>

<style scoped>
.editor-shell {
  position: relative;
  z-index: 24;
  isolation: isolate;
  margin: 0 auto;
  border-radius: var(--radius-xs);
  border: 0;
  background: transparent;
  box-shadow: none;
}

.editor-shell__inner {
  width: 100%;
  padding: 0.2rem;
}

.editor-shell__body {
  margin-bottom: 0.15rem;
  border-radius: var(--radius-xs);
  padding: 0.15rem;
}

@media (width >= 640px) {
  .editor-shell__inner {
    padding: 0.25rem;
  }

  .editor-shell__body {
    padding: 0.2rem;
  }
}
</style>
