<template>
  <div class="p-3 my-3 border border-dashed border-[var(--dash-line-color)] rounded-lg">
    <!-- 扩展附加内容 -->
    <div class="mb-1">
      <h2 class="text-[var(--text-color-400)] font-bold mb-1">扩展附加内容</h2>
      <div class="flex flex-row items-center gap-2">
        <!-- 添加音乐 -->
        <BaseButton
          :icon="Music"
          class="w-7 h-7 rounded-md"
          title="添加音乐"
          @click="handleAddExtension(ExtensionType.MUSIC)"
        />
        <!-- 添加视频 -->
        <BaseButton
          :icon="Video"
          class="w-7 h-7 rounded-md"
          title="添加视频"
          @click="handleAddExtension(ExtensionType.VIDEO)"
        />
        <!-- 添加Github项目 -->
        <BaseButton
          :icon="Githubproj"
          class="w-7 h-7 rounded-md"
          title="添加Github项目"
          @click="handleAddExtension(ExtensionType.GITHUBPROJ)"
        />
        <!-- 添加网站链接 -->
        <BaseButton
          :icon="Weblink"
          class="w-7 h-7 rounded-md"
          title="添加网站链接"
          @click="handleAddExtension(ExtensionType.WEBSITE)"
        />
        <!-- 添加3D模型 -->
        <BaseButton
          :icon="Model3D"
          class="w-7 h-7 rounded-md"
          title="添加3D模型"
          @click="handleAddExtension(ExtensionType.MODEL3D)"
        />
      </div>
    </div>

    <!-- 模式切换 -->
    <div class="mb-1">
      <h2 class="text-[var(--text-color-400)] font-bold mb-1">模式切换</h2>
      <div class="flex flex-row items-center gap-2">
        <!-- 打开Todo模式 -->
        <BaseButton :icon="Todo" @click="handleTodo" class="w-7 h-7 rounded-md" />
        <!-- 打开音乐播放模式 -->
        <BaseButton
          :icon="Audio"
          class="w-7 h-7 rounded-md"
          title="音乐播放"
          @click="handlePlayMusic"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Weblink from '@/components/icons/weblink.vue'
import Music from '@/components/icons/music.vue'
import Todo from '@/components/icons/todo.vue'
import Video from '@/components/icons/video.vue'
import Githubproj from '@/components/icons/githubproj.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Audio from '@/components/icons/audio.vue'
import Model3D from '@/components/icons/model3d.vue'

import { theToast } from '@/utils/toast'
import { Mode, ExtensionType } from '@/enums/enums'
import { useEditorStore } from '@/stores/editor'
import { useTodoStore } from '@/stores/todo'

const editorStore = useEditorStore()

const todoStore = useTodoStore()
const { setTodoMode } = todoStore

const handleAddExtension = (extensiontype: ExtensionType) => {
  editorStore.currentMode = Mode.EXTEN
  editorStore.currentExtensionType = extensiontype
  editorStore.extensionToAdd.extension_type = extensiontype
}

const handleTodo = () => {
  setTodoMode(true)
  editorStore.currentMode = Mode.TODO
  theToast.info('已切换到 Todo 模式')
}

const handlePlayMusic = () => {
  editorStore.setMode(Mode.PlayMusic)
  theToast.info('已切换到 音乐播放器 模式')
}
</script>

<style scoped></style>
