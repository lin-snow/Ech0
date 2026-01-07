<template>
  <div
    v-if="ShowEditor"
    class="bg-[var(--editor-bg-color)] ring-1 ring-[var(--ring-color)] ring-inset rounded-lg mx-auto shadow-xs hover:shadow-sm"
  >
    <div class="mx-auto w-full px-3 py-4">
      <!-- The Title && Nav -->
      <TheTitleAndNav />

      <!-- The Editor -->
      <div class="rounded-lg p-2 sm:p-3 mb-1">
        <!-- EchoMode : TheMdEditor -->
        <TheMdEditor v-if="currentMode === Mode.ECH0" class="rounded-lg" />

        <!-- ImageMode : TheImageEditor -->
        <TheImageEditor v-if="currentMode === Mode.Image" />

        <!-- TodoMode : TheTodoModeEditor -->
        <TheTodoModeEditor v-if="currentMode === Mode.TODO" />

        <!-- InboxMode : TheInboxModeEditor -->
        <TheInboxModeEditor v-if="currentMode === Mode.INBOX" />

        <!-- MusicMode : TheMusicModeEditor -->
        <TheMusicModeEditor v-if="currentMode === Mode.PlayMusic" />

        <!-- The Mode Panel -->
        <TheModePanel v-if="currentMode === Mode.Panel" />

        <!-- ExtensionMode: TheExtensionEditor -->
        <TheExtensionEditor v-if="currentMode === Mode.EXTEN" />

        <!-- TagManageMode: TheTagsManager -->
        <TheTagsManager v-if="currentMode === Mode.TagManage" />
      </div>

      <!-- Editor Buttons -->
      <TheEditorButtons />

      <!-- Editor Image -->
      <TheEditorImage />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheMdEditor from '@/components/advanced/TheMdEditor.vue'
import TheModePanel from './TheEditor/TheModePanel.vue'
import TheTitleAndNav from './TheEditor/TheTitleAndNav.vue'
import TheImageEditor from './TheEditor/TheImageEditor.vue'
import TheEditorImage from './TheEditor/TheEditorImage.vue'
import TheEditorButtons from './TheEditor/TheEditorButtons.vue'
import TheTodoModeEditor from './TheEditor/TheTodoModeEditor.vue'
import TheInboxModeEditor from './TheEditor/TheInboxModeEditor.vue'
import TheMusicModeEditor from './TheEditor/TheMusicModeEditor.vue'
import TheExtensionEditor from './TheEditor/TheExtensionEditor.vue'
import TheTagsManager from './TheEditor/TheTagsManager.vue'

import { theToast } from '@/utils/toast'
import { watch } from 'vue'
import { useEchoStore, useEditorStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Mode, ExtensionType, ImageLayout } from '@/enums/enums'

/* --------------- 与Pinia相关 ---------------- */
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
  imagesToAdd,
  websiteToAdd,
  tagToAdd,
  currentExtensionType,
} = storeToRefs(editorStore)

/* -------------------------------------------- */

/* ------------------ 与Watch相关的各种函数 -------------- */
// 监听用户输入
watch(
  () => videoURL.value,
  (newVal) => {
    if (newVal.length > 0) {
      const bvRegex = /(BV[0-9A-Za-z]{10})/
      const ytRegex =
        /(?:https?:\/\/(?:www\.)?)?(?:youtu\.be\/|youtube\.com\/(?:(?:watch)?\?(?:.*&)?v(?:i)?=|(?:embed)\/))([\w-]+)/
      let match = newVal.match(bvRegex)
      if (match) {
        extensionToAdd.value.extension = match[0] //bilibili
      } else {
        match = newVal.match(ytRegex)
        if (match) {
          extensionToAdd.value.extension = match[1] ?? '' //youtube
        } else {
          theToast.error('请输入正确的B站/YT分享链接！')
        }
      }
    }
  },
)

// 监听是否处于更新模式
watch(
  () => isUpdateMode.value,
  (newVal) => {
    if (newVal) {
      // 处于更新模式（将待更新的数据填充到编辑器）
      // 0. 清空编辑器
      editorStore.clearEditor()

      // 1. 填充本文
      echoToAdd.value.content = echoToUpdate.value?.content || ''

      // 2. 填充图片
      if (echoToUpdate.value?.images && echoToUpdate.value.images.length > 0) {
        imagesToAdd.value = echoToUpdate.value.images.map((img) => ({
          image_url: img.image_url || '',
          image_source: img.image_source || '',
          object_key: img.object_key || '',
        }))
      } else {
        imagesToAdd.value = []
      }

      // 3. 填充扩展
      if (echoToUpdate.value?.extension && echoToUpdate.value.extension_type) {
        currentExtensionType.value = echoToUpdate.value.extension_type as ExtensionType
        extensionToAdd.value.extension = echoToUpdate.value.extension
        extensionToAdd.value.extension_type = echoToUpdate.value.extension_type
        // 根据扩展类型填充
        switch (echoToUpdate.value.extension_type) {
          case ExtensionType.MUSIC:
            break

          case ExtensionType.VIDEO:
            videoURL.value = echoToUpdate.value.extension // 直接使用extension填充B站链接
            break

          case ExtensionType.GITHUBPROJ:
            break

          case ExtensionType.WEBSITE:
            // 反序列化网站链接
            const websiteData = JSON.parse(echoToUpdate.value.extension) as {
              title?: string
              site?: string
            }
            websiteToAdd.value.title = websiteData.title || ''
            websiteToAdd.value.site = websiteData.site || ''
            break
        }
      }

      // 4. 填充标签
      const tags = echoToUpdate.value?.tags
      tagToAdd.value = Array.isArray(tags) && tags.length > 0 ? (tags[0]?.name ?? '') : ''

      // 5. 填充私密状态 && 布局方式
      echoToAdd.value.private = echoToUpdate.value?.private || false
      echoToAdd.value.layout = echoToUpdate.value?.layout || ImageLayout.WATERFALL

      // 6. 回到页面顶部（触发BackToTop）
      window.scrollTo({ top: 0, behavior: 'smooth' })

      // 7. 弹出通知，提示可以编辑了
      theToast.info('已进入更新模式，请编辑内容后点击更新按钮！')
    } else {
      // 退出更新模式
      echoToUpdate.value = null
    }
  },
)
/* ------------------------------------------------------- */
</script>
