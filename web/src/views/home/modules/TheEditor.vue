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
      <TheEditorMedia />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEditorMedia from './TheEditor/TheEditorMedia.vue'
import TheEditorButtons from './TheEditor/TheEditorButtons.vue'

import { theToast } from '@/utils/toast'
import { defineAsyncComponent, onMounted, watch } from 'vue'
import { useEchoStore, useEditorStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Mode, ExtensionType, ImageLayout } from '@/enums/enums'
import { getEchoFiles } from '@/utils/echo'
import { useI18n } from 'vue-i18n'

const TheMdEditor = defineAsyncComponent(() => import('@/components/advanced/md/TheMdEditor.vue'))
const TheModePanel = defineAsyncComponent(() => import('./TheEditor/TheModePanel.vue'))
const TheMediaEditor = defineAsyncComponent(() => import('./TheEditor/TheMediaEditor.vue'))
const TheExtensionEditor = defineAsyncComponent(() => import('./TheEditor/TheExtensionEditor.vue'))
const TheTagsManager = defineAsyncComponent(() => import('./TheEditor/TheTagsManager.vue'))

const echoStore = useEchoStore()
const editorStore = useEditorStore()
const { echoToUpdate } = storeToRefs(echoStore)
const {
  ShowEditor,
  currentMode,
  isUpdateMode,
  echoToAdd,
  videoURL,
  extensionToAdd,
  websiteToAdd,
  locationToAdd,
  tweetToAdd,
  tagToAdd,
  currentExtensionType,
} = storeToRefs(editorStore)
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

const fillEditorFromEchoToUpdate = () => {
  editorStore.clearEditor()
  echoToAdd.value.content = echoToUpdate.value?.content || ''

  const existingMedia = getEchoFiles(echoToUpdate.value)
  if (existingMedia.length > 0) {
    editorStore.setFilesToAdd(
      existingMedia.map((f) => ({
        id: String(f.id || ''),
        url: f.url || '',
        storage_type: f.storage_type || 'local',
        category: f.category,
        content_type: f.content_type,
        key: f.key || '',
        size: f.size,
        width: f.width,
        height: f.height,
      })),
    )
  } else {
    editorStore.setFilesToAdd([])
  }

  if (echoToUpdate.value?.extension) {
    currentExtensionType.value = echoToUpdate.value.extension.type as ExtensionType
    extensionToAdd.value.extension_type = echoToUpdate.value.extension.type
    switch (echoToUpdate.value.extension.type) {
      case ExtensionType.MUSIC:
        extensionToAdd.value.extension = echoToUpdate.value.extension.payload.url || ''
        break
      case ExtensionType.VIDEO:
        extensionToAdd.value.extension = echoToUpdate.value.extension.payload.videoId || ''
        videoURL.value = echoToUpdate.value.extension.payload.videoId || ''
        break
      case ExtensionType.GITHUBPROJ:
        extensionToAdd.value.extension = echoToUpdate.value.extension.payload.repoUrl || ''
        break
      case ExtensionType.WEBSITE:
        websiteToAdd.value.title = echoToUpdate.value.extension.payload.title || ''
        websiteToAdd.value.site = echoToUpdate.value.extension.payload.site || ''
        break
      case ExtensionType.LOCATION: {
        const payload = echoToUpdate.value.extension.payload
        locationToAdd.value = {
          latitude: typeof payload.latitude === 'number' ? payload.latitude : null,
          longitude: typeof payload.longitude === 'number' ? payload.longitude : null,
          placeholder: payload.placeholder || '',
        }
        extensionToAdd.value.extension =
          typeof payload.latitude === 'number' && typeof payload.longitude === 'number'
            ? `${payload.latitude},${payload.longitude}`
            : ''
        break
      }
      case ExtensionType.TWEET: {
        const payload = echoToUpdate.value.extension.payload
        tweetToAdd.value = {
          url: payload.url || '',
          username: payload.username || '',
          statusId: payload.statusId || '',
        }
        extensionToAdd.value.extension = payload.url || ''
        break
      }
    }
  }

  const tags = echoToUpdate.value?.tags
  tagToAdd.value = Array.isArray(tags)
    ? tags.map((tag) => tag?.name).filter((name): name is string => !!name)
    : []
  echoToAdd.value.private = echoToUpdate.value?.private || false
  echoToAdd.value.layout = echoToUpdate.value?.layout || ImageLayout.WATERFALL
  window.scrollTo({ top: 0, behavior: 'smooth' })
  theToast.info(String(t('editor.enteredUpdateMode')))
}

watch(
  () => isUpdateMode.value,
  (newVal) => {
    if (newVal) {
      fillEditorFromEchoToUpdate()
    } else {
      echoToUpdate.value = null
    }
  },
)

onMounted(() => {
  if (isUpdateMode.value) {
    fillEditorFromEchoToUpdate()
  }
})
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
