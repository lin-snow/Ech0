<template>
  <div
    class="w-full max-w-sm bg-[var(--echo-detail-bg-color)] h-auto p-5 shadow rounded-lg mx-auto"
  >
    <!-- 顶部Logo 和 用户名 -->
    <div class="flex flex-row items-center gap-2 mt-2 mb-4">
      <!-- <div class="text-xl">👾</div> -->
      <div>
        <img
          :src="logo"
          alt="logo"
          class="w-10 h-10 sm:w-12 sm:h-12 rounded-full ring-1 ring-gray-200 shadow-sm object-cover"
        />
      </div>
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--text-color-700)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            {{ SystemSetting.server_name }}
          </h2>

          <div>
            <Verified class="text-sky-500 w-5 h-5" />
          </div>
        </div>
        <span class="text-[var(--echo-detail-username-color)] font-serif"
          >@ {{ echo.username }}
        </span>
      </div>
    </div>

    <!-- 图片 && 内容 -->
    <div>
      <div class="py-4">
        <!-- 根据布局决定文字与图片顺序 -->
        <!-- grid 和 horizontal 时，文字在图片上；其他布局（waterfall/carousel/null/undefined）文字在图片下 -->
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID || props.echo.layout === ImageLayout.HORIZONTAL
          "
        >
          <!-- 文字在上 -->
          <div class="mb-3">
            <MdPreview
              :id="previewOptions.proviewId"
              :modelValue="props.echo.content"
              :theme="theme"
              :show-code-row-number="previewOptions.showCodeRowNumber"
              :preview-theme="previewOptions.previewTheme"
              :code-theme="previewOptions.codeTheme"
              :code-style-reverse="previewOptions.codeStyleReverse"
              :no-img-zoom-in="previewOptions.noImgZoomIn"
              :code-foldable="previewOptions.codeFoldable"
              :auto-fold-threshold="previewOptions.autoFoldThreshold"
            />
          </div>

          <TheImageGallery :images="props.echo.images" :layout="props.echo.layout" />
        </template>

        <template v-else>
          <!-- 图片在上，文字在下 -->
          <TheImageGallery :images="props.echo.images" :layout="props.echo.layout" />

          <div class="mt-3">
            <MdPreview
              :id="previewOptions.proviewId"
              :modelValue="props.echo.content"
              :theme="theme"
              :show-code-row-number="previewOptions.showCodeRowNumber"
              :preview-theme="previewOptions.previewTheme"
              :code-theme="previewOptions.codeTheme"
              :code-style-reverse="previewOptions.codeStyleReverse"
              :no-img-zoom-in="previewOptions.noImgZoomIn"
              :code-foldable="previewOptions.codeFoldable"
              :auto-fold-threshold="previewOptions.autoFoldThreshold"
            />
          </div>
        </template>

        <!-- 扩展内容 -->
        <div v-if="props.echo.extension" class="my-4">
          <div v-if="props.echo.extension_type === ExtensionType.MUSIC">
            <TheAPlayerCard :echo="props.echo" />
          </div>
          <div v-if="props.echo.extension_type === ExtensionType.VIDEO">
            <TheVideoCard :videoId="props.echo.extension" class="px-2 mx-auto hover:shadow-md" />
          </div>
          <TheGithubCard
            v-if="props.echo.extension_type === ExtensionType.GITHUBPROJ"
            :GithubURL="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
          <TheWebsiteCard
            v-if="props.echo.extension_type === ExtensionType.WEBSITE"
            :website="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
          <TheModel3DCard
            v-if="props.echo.extension_type === ExtensionType.MODEL3D"
            :modelSrc="props.echo.extension"
            class="px-2 mx-auto hover:shadow-md"
          />
        </div>
      </div>
    </div>

    <!-- 日期时间 && 操作按钮 -->
    <div class="flex justify-between items-center">
      <!-- 日期时间 -->
      <div class="flex justify-start items-center h-auto">
        <div class="flex justify-start text-sm text-[var(--echo-detail-datetime-color)] mr-1">
          {{ formatDate(props.echo.created_at) }}
        </div>
        <!-- 标签 -->
        <div class="text-sm text-[var(--text-color-300)] w-18 truncate text-nowrap">
          <span>{{ props.echo.tags ? `#${props.echo.tags[0]?.name}` : '' }}</span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div ref="menuRef" class="relative flex items-center justify-center gap-2 h-auto">
        <!-- 分享 -->
        <div class="flex items-center justify-end" title="分享">
          <button
            @click="handleShareEcho(props.echo.id)"
            title="分享"
            :class="[
              'transform transition-transform duration-150',
              isShareAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <Share class="w-4 h-4" />
          </button>
        </div>

        <!-- 点赞 -->
        <div class="flex items-center justify-end" title="点赞">
          <div class="flex items-center gap-1">
            <!-- 点赞按钮   -->
            <button
              @click="handleLikeEcho(props.echo.id)"
              title="点赞"
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </button>

            <!-- 点赞数量   -->
            <span class="text-sm text-[var(--text-color-400)]">
              <!-- 如果点赞数不超过99，则显示数字，否则显示99+ -->
              {{ props.echo.fav_count > 99 ? '99+' : props.echo.fav_count }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheGithubCard from './TheGithubCard.vue'
import TheVideoCard from './TheVideoCard.vue'
import Verified from '../icons/verified.vue'
import GrayLike from '../icons/graylike.vue'
import Share from '../icons/share.vue'
import TheAPlayerCard from './TheAPlayerCard.vue'
import TheWebsiteCard from './TheWebsiteCard.vue'
import TheImageGallery from './TheImageGallery.vue'
import TheModel3DCard from './TheModel3DCard.vue'
import 'md-editor-v3/lib/preview.css'
import { MdPreview } from 'md-editor-v3'
import { onMounted, computed, ref } from 'vue'
import { fetchLikeEcho } from '@/service/api'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { storeToRefs } from 'pinia'
import { fetchGetStatus } from '@/service/api'
import { useSettingStore } from '@/stores/setting'
import { getApiUrl } from '@/service/request/shared'
import { ExtensionType, ImageLayout } from '@/enums/enums'
import { formatDate } from '@/utils/other'
import { useThemeStore } from '@/stores/theme'

const emit = defineEmits(['updateLikeCount'])

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()
const themeStore = useThemeStore()

const theme = computed(() => (themeStore.theme === 'light' ? 'light' : 'dark'))
const previewOptions = {
  proviewId: 'preview-only',
  showCodeRowNumber: false,
  previewTheme: 'github',
  codeTheme: 'atom',
  codeStyleReverse: true,
  noImgZoomIn: false,
  codeFoldable: true,
  autoFoldThreshold: 15,
}

const isLikeAnimating = ref(false)
const isShareAnimating = ref(false)

const LIKE_LIST_KEY = 'likedEchoIds'
const likedEchoIds: number[] = localStg.getItem(LIKE_LIST_KEY) || []
const hasLikedEcho = (echoId: number): boolean => {
  return likedEchoIds.includes(echoId)
}
const handleLikeEcho = (echoId: number) => {
  isLikeAnimating.value = true
  setTimeout(() => {
    isLikeAnimating.value = false
  }, 250) // 对应 duration-250

  // 检查LocalStorage中是否已经点赞过
  if (hasLikedEcho(echoId)) {
    theToast.info('你已经点赞过了,感谢你的喜欢！')
    return
  }

  fetchLikeEcho(echoId).then((res) => {
    if (res.code === 1) {
      likedEchoIds.push(echoId)
      localStg.setItem(LIKE_LIST_KEY, likedEchoIds)
      // 发送更新事件
      emit('updateLikeCount', echoId)
      theToast.info('点赞成功！')
    }
  })
}

const handleShareEcho = (echoId: number) => {
  isShareAnimating.value = true
  setTimeout(() => {
    isShareAnimating.value = false
  }, 250) // 对应 duration-250

  const shareUrl = `${window.location.origin}/echo/${echoId}\n ———— 来自 Ech0 分享`
  navigator.clipboard.writeText(shareUrl).then(() => {
    theToast.info('已复制到剪贴板！')
  })
}

const settingStore = useSettingStore()

const { SystemSetting } = storeToRefs(settingStore)

const apiUrl = getApiUrl()
const logo = ref<string>('/favicon.svg')

onMounted(() => {
  fetchGetStatus().then((res) => {
    if (res.code === 1) {
      const theLogo = res.data.logo
      if (theLogo && theLogo !== '') {
        logo.value = `${apiUrl}${theLogo}`
      }
    }
  })
})
</script>

<style scoped lang="css">
#preview-only {
  background-color: inherit;
}

.md-editor {
  font-family: var(--font-sans);
  /* font-family: 'LXGW WenKai Screen'; */
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
