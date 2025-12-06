<template>
  <div>
    <!-- 音乐分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.MUSIC">
      <h2 class="text-[var(--text-color-500)] font-bold mb-1">音乐分享</h2>
      <p class="text-[var(--text-color-400)] text-sm">支持网易云/QQ音乐/Apple Music</p>
      <p class="text-[var(--text-color-400)] text-sm mb-1">注意：不支持VIP歌曲，建议使用自建API</p>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        placeholder="音乐链接..."
      />
      <div
        v-if="
          editorStore.extensionToAdd.extension.length > 0 &&
          editorStore.extensionToAdd.extension_type === ExtensionType.MUSIC
        "
        class="mt-1 text-[var(--text-color-400)] text-md"
      >
        解析结果：
        <span v-if="parseMusicURL(editorStore.extensionToAdd.extension)" class="text-green-400"
          >成功</span
        >
        <span v-else class="text-red-300">失败</span>
      </div>
    </div>
    <!-- Bilibili/YouTube视频分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.VIDEO">
      <div class="text-[var(--text-color-500)] font-bold mb-1">
        视频分享（支持Bilibili、YouTube）
      </div>
      <div class="text-[var(--text-color-400)] mb-1">粘贴自动提取ID</div>
      <BaseInput
        v-model="editorStore.videoURL"
        class="rounded-lg h-auto w-full my-2"
        placeholder="B站/YouTube链接..."
      />
      <div class="text-[var(--text-color-500)] my-1">
        Video ID：{{ editorStore.extensionToAdd.extension }}
      </div>
    </div>
    <!-- Github项目分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.GITHUBPROJ">
      <div class="text-[var(--text-color-500)] font-bold mb-1">Github项目分享</div>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        placeholder="https://github.com/username/repo"
      />
    </div>
    <!-- 网站链接分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.WEBSITE">
      <div class="text-[var(--text-color-500)] font-bold mb-1">网站链接分享</div>
      <!-- 网站标题 -->
      <BaseInput
        v-model="editorStore.websiteToAdd.title"
        class="rounded-lg h-auto w-full mb-2"
        placeholder="网站标题..."
      />
      <div class="flex items-center gap-2">
        <BaseInput
          v-model="editorStore.websiteToAdd.site"
          class="rounded-lg h-auto flex-1"
          placeholder="https://example.com"
        />
        <BaseButton
          class="rounded-lg px-3 py-2 text-sm whitespace-nowrap"
          :disabled="isFetchingWebsiteTitle"
          @click="handleFetchWebsiteTitle"
        >
          {{ isFetchingWebsiteTitle ? '获取中…' : '获取标题' }}
        </BaseButton>
      </div>
    </div>
    <!-- 3D模型分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.MODEL3D">
      <div class="text-[var(--text-color-500)] font-bold mb-1">3D模型分享</div>
      <p class="text-[var(--text-color-400)] text-sm mb-2">支持 GLB/GLTF 格式的3D模型文件</p>
      
      <!-- 上传区域 -->
      <div
        class="border-2 border-dashed border-[var(--dash-line-color)] rounded-lg p-4 text-center cursor-pointer hover:border-[var(--primary-color)] transition-colors"
        @click="triggerFileInput"
        @dragover.prevent
        @drop.prevent="handleFileDrop"
      >
        <input
          ref="fileInputRef"
          type="file"
          accept=".glb,.gltf"
          class="hidden"
          @change="handleFileSelect"
        />
        <div v-if="!isUploading && !editorStore.extensionToAdd.extension">
          <div class="text-[var(--text-color-400)] mb-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto mb-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <path d="M2 17l10 5 10-5" />
              <path d="M2 12l10 5 10-5" />
            </svg>
          </div>
          <p class="text-[var(--text-color-400)]">点击或拖拽上传3D模型</p>
          <p class="text-[var(--text-color-300)] text-xs mt-1">最大支持 50MB</p>
        </div>
        <div v-else-if="isUploading" class="text-[var(--text-color-400)]">
          <div class="animate-spin w-8 h-8 border-2 border-[var(--primary-color)] border-t-transparent rounded-full mx-auto mb-2"></div>
          <p>上传中...</p>
        </div>
        <div v-else class="text-green-500">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto mb-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 6L9 17l-5-5" />
          </svg>
          <p>上传成功！</p>
          <p class="text-[var(--text-color-400)] text-xs mt-1 truncate">{{ uploadedFileName }}</p>
        </div>
      </div>
      
      <!-- 或者直接输入URL -->
      <div class="mt-3">
        <p class="text-[var(--text-color-400)] text-sm mb-1">或直接输入模型URL：</p>
        <BaseInput
          v-model="editorStore.extensionToAdd.extension"
          class="rounded-lg h-auto w-full"
          placeholder="https://example.com/model.glb"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { ExtensionType } from '@/enums/enums'
import { parseMusicURL, extractAndCleanMusicURL } from '@/utils/other' // 导入新函数
import { useEditorStore } from '@/stores/editor'
import { ref, watch } from 'vue' // 从 vue 导入 watch
import { fetchGetWebsiteTitle, fetchUploadModel } from '@/service/api'
import { theToast } from '@/utils/toast'

const editorStore = useEditorStore()
const isFetchingWebsiteTitle = ref(false)
const isUploading = ref(false)
const uploadedFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

const triggerFileInput = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    await uploadModel(file)
  }
}

const handleFileDrop = async (event: DragEvent) => {
  const file = event.dataTransfer?.files?.[0]
  if (file) {
    await uploadModel(file)
  }
}

const uploadModel = async (file: File) => {
  // 检查文件类型
  const ext = file.name.toLowerCase().split('.').pop()
  if (ext !== 'glb' && ext !== 'gltf') {
    theToast.error('请上传 GLB 或 GLTF 格式的3D模型文件')
    return
  }

  // 检查文件大小 (50MB)
  if (file.size > 50 * 1024 * 1024) {
    theToast.error('文件大小不能超过 50MB')
    return
  }

  isUploading.value = true
  try {
    const res = await fetchUploadModel(file)
    if (res.code === 1 && res.data) {
      editorStore.extensionToAdd.extension = res.data
      editorStore.extensionToAdd.extension_type = ExtensionType.MODEL3D
      uploadedFileName.value = file.name
      theToast.success('3D模型上传成功！')
    } else {
      theToast.error(res.msg || '上传失败')
    }
  } catch (error) {
    console.error('Upload model error:', error)
    theToast.error('上传失败，请稍后重试')
  } finally {
    isUploading.value = false
  }
}

const handleFetchWebsiteTitle = async () => {
  const websiteURL = (editorStore.websiteToAdd.site || '').trim()
  if (!websiteURL) {
    theToast.warning('请先输入网站链接')
    return
  }

  isFetchingWebsiteTitle.value = true
  try {
    const res = await fetchGetWebsiteTitle(websiteURL)
    if (res.code === 1) {
      editorStore.websiteToAdd.title = res.data
      theToast.success('已获取网站标题')
    } else {
      theToast.error(res.msg || '获取网站标题失败')
    }
  } catch (error) {
    console.error('Failed to fetch website title', error)
    theToast.error('获取网站标题失败')
  } finally {
    isFetchingWebsiteTitle.value = false
  }
}

// 监听音乐链接输入框的变化
watch(
  () => editorStore.extensionToAdd.extension,
  (newValue) => {
    // 只在当前是音乐分享模式，并且输入框有内容时才执行
    if (editorStore.currentExtensionType !== ExtensionType.MUSIC || !newValue) {
      return
    }

    // 尝试提取并清理链接
    const cleanUrl = extractAndCleanMusicURL(newValue)

    // 如果成功提取到干净的链接，并且这个链接和当前输入框的内容不一样
    // （防止无限循环和重复赋值）
    if (cleanUrl && cleanUrl !== newValue) {
      // 更新输入框的值为干净的链接
      editorStore.extensionToAdd.extension = cleanUrl
    }
  },
)
</script>

<style scoped></style>
