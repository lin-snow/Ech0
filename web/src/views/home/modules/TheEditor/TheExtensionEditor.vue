<template>
  <div>
    <!-- 音乐分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.MUSIC">
      <h2 class="text-gray-500 font-bold mb-1">音乐分享</h2>
      <p class="text-gray-400 text-sm">支持网易云/QQ音乐/Apple Music</p>
      <p class="text-gray-400 text-sm mb-1">注意：不支持VIP歌曲，建议使用自建API</p>
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
        class="mt-1 text-gray-400 text-md"
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
      <div class="text-gray-500 font-bold mb-1">视频分享（支持Bilibili、YouTube）</div>
      <div class="text-gray-400 mb-1">粘贴自动提取ID</div>
      <BaseInput
        v-model="editorStore.videoURL"
        class="rounded-lg h-auto w-full my-2"
        placeholder="B站/YouTube链接..."
      />
      <div class="text-gray-500 my-1">Video ID：{{ editorStore.extensionToAdd.extension }}</div>
    </div>
    <!-- Github项目分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.GITHUBPROJ">
      <div class="text-gray-500 font-bold mb-1">Github项目分享</div>
      <BaseInput
        v-model="editorStore.extensionToAdd.extension"
        class="rounded-lg h-auto w-full"
        placeholder="https://github.com/username/repo"
      />
    </div>
    <!-- 网站链接分享 -->
    <div v-if="editorStore.currentExtensionType === ExtensionType.WEBSITE">
      <div class="text-gray-500 font-bold mb-1">网站链接分享</div>
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
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { ExtensionType } from '@/enums/enums'
import { parseMusicURL, extractAndCleanMusicURL } from '@/utils/other' // 导入新函数
import { useEditorStore } from '@/stores/editor'
import { ref, watch } from 'vue' // 从 vue 导入 watch
import { fetchGetWebsiteTitle } from '@/service/api'
import { theToast } from '@/utils/toast'

const editorStore = useEditorStore()
const isFetchingWebsiteTitle = ref(false)

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
