<template>
  <div class="p-1 my-4">
    <!-- 扩展附加内容 -->
    <div class="mb-1">
      <h2 class="text-gray-500 font-bold mb-1">扩展附加内容</h2>
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
      </div>
    </div>
    <!-- 模式切换 -->
    <div class="mb-1">
      <h2 class="text-gray-500 font-bold mb-1">模式切换</h2>
      <div class="flex flex-row items-center gap-2">
        <!-- 打开Todo模式 -->
        <BaseButton :icon="Todo" @click="emit('switchTodo')" class="w-7 h-7 rounded-md" />
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

const enum ExtensionType {
  MUSIC = 'MUSIC',
  VIDEO = 'VIDEO',
  GITHUBPROJ = 'GITHUBPROJ',
  WEBSITE = 'WEBSITE',
}

const emit = defineEmits([
  'switchTodo',
  'switchExtension',
  'switchPlayMusic',
  'addMusic',
  'addVideo',
  'addGithubproj',
  'addWeblink',
])

const handleAddExtension = (extensiontype: ExtensionType) => {
  emit('switchExtension', extensiontype)
}

const handlePlayMusic = () => {
  emit('switchPlayMusic')
}
</script>

<style scoped></style>
