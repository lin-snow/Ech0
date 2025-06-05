<template>
  <div
    v-if="musicUrl"
    class="w-full flex flex-col gap-2 p-4 bg-white rounded-lg ring-1 ring-gray-200 ring-inset mx-auto shadow-sm hover:shadow-md mb-2"
  >
    <p class="text-gray-600 font-bold text-lg flex items-center">
      <Album class="mr-1" /> 正在听的音乐：
    </p>
    <div class="flex items-center gap-4">
      <button
        @click="togglePlay"
        class="w-8 h-8 flex items-center justify-center rounded-full bg-red-100 shadow-sm hover:bg-red-200 text-white transition"
      >
        <span v-if="!isPlaying">
          <Play class="w-5 h-5" :color="'#ee5b5bd9'" />
        </span>
        <span v-else>
          <Pause class="w-6 h-6" :color="'#ee5b5bd9'" />
        </span>
      </button>

      <!-- 提示 -->
      <div v-if="isPlaying" class="text-gray-500">播放中...</div>
      <div v-else class="text-gray-500">暂停中...</div>
    </div>

    <audio
      ref="audioRef"
      :src="url"
      @play="isPlaying = true"
      @pause="isPlaying = false"
      preload="none"
    />
  </div>
</template>

<script setup lang="ts">
import { fetchGetMusic } from '@/service/api'
import { getApiUrl } from '@/service/request/shared'
import { onMounted, ref } from 'vue'
import Album from '../icons/album.vue'
import Pause from '../icons/pause.vue'
import Play from '../icons/play.vue'

const url = ref<string>('')
const musicUrl = ref<string>('')
const isPlaying = ref(false)
const audioRef = ref<HTMLAudioElement | null>(null)

function togglePlay() {
  if (!audioRef.value) return
  if (isPlaying.value) {
    audioRef.value.pause()
  } else {
    audioRef.value.play()
  }
}

const handleGetMusic = async () => {
  const res = await fetchGetMusic()
  musicUrl.value = res.data
  url.value = `${getApiUrl()}/playmusic?t=${Date.now()}` // 添加时间戳，绕过缓存
}

defineExpose({
  handleGetMusic,
})

onMounted(async () => {
  handleGetMusic()
})
</script>
