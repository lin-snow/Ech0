<template>
  <div>
    <h2 class="text-gray-500 font-bold my-2">插入图片/视频（支持直链、本地、S3存储）</h2>
    <div v-if="!MediaUploading" class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2">
        <span class="text-gray-500">选择添加方式：</span>
        <!-- 直链 -->
        <BaseButton
          :icon="Url"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetMediaSource(ImageSource.URL)"
          title="插入图片/视频链接"
        />
        <!-- 上传本地 -->
        <BaseButton
          :icon="Upload"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetMediaSource(ImageSource.LOCAL)"
          title="上传本地图片/视频"
        />
        <!-- S3 存储 -->
        <BaseButton
          v-if="S3Setting.enable"
          :icon="Bucket"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetMediaSource(ImageSource.S3)"
          title="S3存储图片/视频"
        />
      </div>
      <div>
        <BaseButton
          v-if="mediaToAdd.media_url != ''"
          :icon="Addmore"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="editorStore.handleAddMoreMedia"
          title="添加更多图片/视频"
        />
      </div>
    </div>

    <!-- 布局方式选择 -->
    <div class="mb-3 flex items-center gap-2">
      <span class="text-gray-500">布局方式：</span>
      <BaseSelect
        v-model="echoToAdd.layout"
        :options="layoutOptions"
        class="w-32 h-7"
        placeholder="请选择布局方式"
      />
    </div>

    <!-- 当前上传方式与状态 -->
    <div class="text-gray-300 text-sm mb-1">
      当前上传方式为
      <span class="font-bold">
        {{
          mediaToAdd.media_source === ImageSource.URL
            ? '直链'
            : mediaToAdd.media_source === ImageSource.LOCAL
              ? '本地存储'
              : 'S3存储'
        }}</span
      >
      {{ !MediaUploading ? '' : '，正在上传中...' }}
    </div>

    <div class="my-1">
      <!-- 媒体上传 -->
      <TheUppy
        v-if="mediaToAdd.media_source !== ImageSource.URL"
        :TheImageSource="mediaToAdd.media_source"
      />

      <!-- 媒体直链 -->
      <BaseInput
        v-if="mediaToAdd.media_source === ImageSource.URL"
        v-model="mediaToAdd.media_url"
        class="rounded-lg h-auto w-full"
        placeholder="请输入图片或视频链接..."
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useEditorStore } from '@/stores/editor'
import { useSettingStore } from '@/stores/setting'
import { storeToRefs } from 'pinia'
import { ImageSource, ImageLayout } from '@/enums/enums'
import Url from '@/components/icons/url.vue'
import Upload from '@/components/icons/upload.vue'
import Bucket from '@/components/icons/bucket.vue'
import Addmore from '@/components/icons/addmore.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import TheUppy from '@/components/advanced/TheUppy.vue'
import { localStg } from '@/utils/storage'

const editorStore = useEditorStore()
const { mediaToAdd, MediaUploading, echoToAdd } = storeToRefs(editorStore)
const settingStore = useSettingStore()
const { S3Setting } = storeToRefs(settingStore)

const handleSetMediaSource = (source: ImageSource) => {
  mediaToAdd.value.media_source = source

  // 记忆上传方式
  localStg.setItem('image_source', source)
}

// 布局选择
const layoutOptions = [
  { label: '瀑布流', value: ImageLayout.WATERFALL },
  { label: '九宫格', value: ImageLayout.GRID },
  { label: '单图轮播', value: ImageLayout.CAROUSEL },
  { label: '水平轮播', value: ImageLayout.HORIZONTAL },
]
</script>

<style scoped></style>
