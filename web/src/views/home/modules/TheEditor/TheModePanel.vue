<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="editor-mode-panel p-3 my-3 border border-dashed border-[var(--dash-line-color)] rounded-xs"
  >
    <div class="mb-1">
      <h2 class="text-[var(--color-text-muted)] font-bold mb-1">{{ t('editor.extraContent') }}</h2>
      <div class="flex flex-row items-center gap-2">
        <BaseButton
          :icon="Music"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addMusic')"
          @click="handleAddExtension(ExtensionType.MUSIC)"
        />
        <BaseButton
          :icon="Video"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addVideo')"
          @click="handleAddExtension(ExtensionType.VIDEO)"
        />
        <BaseButton
          :icon="Githubproj"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addGithubProject')"
          @click="handleAddExtension(ExtensionType.GITHUBPROJ)"
        />
        <BaseButton
          :icon="Weblink"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addWebsiteLink')"
          @click="handleAddExtension(ExtensionType.WEBSITE)"
        />
        <BaseButton
          :icon="MapPin"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addLocation')"
          @click="handleAddExtension(ExtensionType.LOCATION)"
        />
        <BaseButton
          :icon="XBrand"
          class="w-7 h-7 rounded-xs"
          :tooltip="t('editor.addTweet')"
          @click="handleAddExtension(ExtensionType.TWEET)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Weblink from '@/components/icons/weblink.vue'
import Music from '@/components/icons/music.vue'
import Video from '@/components/icons/video.vue'
import Githubproj from '@/components/icons/githubproj.vue'
import MapPin from '@/components/icons/mappin.vue'
import XBrand from '@/components/icons/tweet.vue'
import BaseButton from '@/components/common/BaseButton.vue'

import { Mode, ExtensionType } from '@/enums/enums'
import { useEditorStore } from '@/stores'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const { t } = useI18n()

const handleAddExtension = (extensiontype: ExtensionType) => {
  editorStore.currentMode = Mode.EXTEN
  editorStore.currentExtensionType = extensiontype
  editorStore.extensionToAdd.extension_type = extensiontype
}
</script>

<style scoped>
.editor-mode-panel {
  border-color: var(--md-editor-mini-border);

  --btn-bg-color: var(--md-editor-mini-btn-bg);
  --btn-ring-color: var(--md-editor-actions-ring-color);
  --btn-hover-bg-color: var(--md-editor-actions-hover-bg);
  --btn-hover-border-color: var(--md-editor-actions-hover-border);
}
</style>
