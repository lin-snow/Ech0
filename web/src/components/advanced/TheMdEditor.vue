<template>
  <MdEditor
    class="h-auto sm:min-h-[8rem] md:min-h-[13rem]"
    v-model="content"
    :id="initEditor.id"
    :theme="initEditor.theme"
    :language="initEditor.language"
    :show-code-row-number="initEditor.showCodeRowNumber"
    :preview-theme="initEditor.previewTheme"
    :code-theme="initEditor.codeTheme"
    :code-style-reverse="initEditor.codeStyleReverse"
    :no-img-zoom-in="initEditor.noImgZoomIn"
    :code-foldable="initEditor.codeFoldable"
    :auto-fold-threshold="initEditor.autoFoldThreshold"
    :toolbars="initEditor.toolbars"
    :no-prettier="initEditor.noPrettier"
    :tab-width="initEditor.tabWidth"
    :placeholder="initEditor.placeholder"
    :preview="initEditor.preview"
    :scroll-auto="initEditor.scrollAuto"
    :show-toolbar-name="initEditor.showToolbarName"
    :footers="initEditor.footers"
    :no-upload-img="initEditor.noUploadImg"
  />
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { MdEditor } from 'md-editor-v3'
import type { ToolbarNames } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const content = defineModel<string>()
const initEditor = reactive({
  class: 'theMdEditor',
  theme: 'light' as 'light' | 'dark',
  language: 'zh_CN',
  id: 'theMdEditor',
  showCodeRowNumber: false,
  previewTheme: 'github',
  codeTheme: 'atom',
  codeStyleReverse: false,
  noImgZoomIn: false,
  codeFoldable: true,
  autoFoldThreshold: 15,
  preview: false,
  toolbars: ['=', 'previewOnly', 'pageFullscreen'] as ToolbarNames[],
  noPrettier: false,
  tabWidth: 2,
  placeholder: '一吐为快~',
  scrollAuto: true,
  showToolbarName: false,
  footers: [],
  noUploadImg: true,
})
</script>

<style scoped lang="css">
#theMdEditor {
  height: 10rem;
}

:deep(.md-editor-toolbar-item) {
  color: #939393 !important;
}

:deep(ul li) {
  list-style-type: disc;
}
:deep(ul li li) {
  list-style-type: circle;
}
:deep(ul li li li) {
  list-style-type: square;
}
:deep(ol li) {
  list-style-type: decimal;
}
</style>
