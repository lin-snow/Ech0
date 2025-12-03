<template>
  <div class="model-viewer-container rounded-lg overflow-hidden relative">
    <!-- 右上角标签和下载按钮 -->
    <div class="absolute top-2 right-2 z-10 flex items-center gap-2">
      <!-- 模型类型标签 -->
      <span class="px-2 py-1 text-xs font-medium rounded bg-black/50 text-white backdrop-blur-sm">
        {{ modelType }}
      </span>
      <!-- 下载按钮 -->
      <button
        @click="handleDownload"
        class="p-1.5 rounded bg-black/50 text-white backdrop-blur-sm hover:bg-black/70 transition-colors"
        title="下载模型"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="7 10 12 15 17 10" />
          <line x1="12" y1="15" x2="12" y2="3" />
        </svg>
      </button>
    </div>
    
    <model-viewer
      :src="modelUrl"
      :alt="alt"
      auto-rotate
      camera-controls
      shadow-intensity="0"
      exposure="1"
      class="w-full h-64 sm:h-80"
      :poster="poster"
      loading="lazy"
      reveal="auto"
    >
    </model-viewer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelSrc: string
  alt?: string
  poster?: string
}>()

const envURL = import.meta.env.VITE_SERVICE_BASE_URL as string
const backendURL = envURL.endsWith('/') ? envURL.slice(0, -1) : envURL

// 处理模型URL，如果是相对路径则拼接后端地址
const modelUrl = computed(() => {
  if (props.modelSrc.startsWith('http://') || props.modelSrc.startsWith('https://')) {
    return props.modelSrc
  }
  return `${backendURL}/api${props.modelSrc}`
})

const alt = computed(() => props.alt || '3D Model')

// 获取模型类型（从文件扩展名）
const modelType = computed(() => {
  const src = props.modelSrc.toLowerCase()
  if (src.endsWith('.glb')) return 'GLB'
  if (src.endsWith('.gltf')) return 'GLTF'
  return '3D'
})

// 下载模型
const handleDownload = () => {
  const link = document.createElement('a')
  link.href = modelUrl.value
  // 从URL中提取文件名
  const fileName = props.modelSrc.split('/').pop() || `model.${modelType.value.toLowerCase()}`
  link.download = fileName
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<style scoped>
model-viewer {
  --poster-color: transparent;
  background-color: transparent;
}

.model-viewer-container {
  background-color: transparent;
}
</style>
