<template>
  <div>
    <h2 class="text-[var(--text-color-500)] font-bold my-2">插入图片（支持直链、本地、S3存储）</h2>
    <div v-if="!ImageUploading" class="flex items-center gap-2 mb-3">
      <div class="flex items-center gap-2">
        <span class="text-[var(--text-color-500)]">选择添加方式：</span>
        <!-- 直链 -->
        <BaseButton
          :icon="Url"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetImageSource(ImageSource.URL)"
          title="插入图片链接"
        />
        <!-- 上传本地 -->
        <BaseButton
          :icon="Upload"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetImageSource(ImageSource.LOCAL)"
          title="上传本地图片"
        />
        <!-- S3 存储 -->
        <BaseButton
          v-if="S3Setting.enable"
          :icon="Bucket"
          class="w-7 h-7 sm:w-7 sm:h-7 rounded-md"
          @click="handleSetImageSource(ImageSource.S3)"
          title="S3存储图片"
        />
      </div>
    </div>

    <!-- 布局方式选择 -->
    <div class="mb-2 flex items-center gap-2">
      <span class="text-[var(--text-color-500)]">选择布局方式：</span>
      <BaseSelect
        v-model="echoToAdd.layout"
        :options="layoutOptions"
        class="w-32 h-7"
        placeholder="请选择布局方式"
      />
    </div>

    <!-- 智能压缩 -->
    <div v-if="imageToAdd.image_source !== ImageSource.URL" class="mb-3 flex items-center">
      <span class="text-[var(--text-color-500)]">智能压缩：</span>
      <BaseSwitch v-model="enableCompressor" />
    </div>

    <!-- 当前上传方式与状态 -->
    <div class="text-[var(--text-color-300)] text-sm mb-2">
      当前上传方式为
      <span class="font-bold">
        {{
          imageToAdd.image_source === ImageSource.URL
            ? '直链'
            : imageToAdd.image_source === ImageSource.LOCAL
              ? '本地存储'
              : 'S3存储'
        }}</span
      >
      {{ !ImageUploading ? '' : '，正在上传中...' }}
    </div>

    <div class="my-1">
      <!-- 图片上传 -->
      <TheUppy
        v-if="imageToAdd.image_source !== ImageSource.URL"
        :TheImageSource="imageToAdd.image_source"
        :EnableCompressor="enableCompressor"
      />

      <!-- 图片直链 -->
      <div v-if="imageToAdd.image_source === ImageSource.URL" class="flex items-center gap-2">
        <BaseInput
          v-model="imageToAdd.image_url"
          class="rounded-lg h-auto flex-1"
          placeholder="请输入图片链接..."
        />
        <BaseButton
          v-if="imageToAdd.image_url != ''"
          :icon="Addmore"
          class="w-8 h-8 sm:w-8 sm:h-8 rounded-md shrink-0"
          @click="editorStore.handleAddMoreImage"
          title="添加更多图片"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useEditorStore, useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { ImageSource, ImageLayout } from '@/enums/enums'
import Url from '@/components/icons/url.vue'
import Upload from '@/components/icons/upload.vue'
import Bucket from '@/components/icons/bucket.vue'
import Addmore from '@/components/icons/addmore.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import TheUppy from '@/components/advanced/TheUppy.vue'
import { localStg } from '@/utils/storage'

const editorStore = useEditorStore()
const { imageToAdd, ImageUploading, echoToAdd } = storeToRefs(editorStore)
const settingStore = useSettingStore()
const { S3Setting } = storeToRefs(settingStore)
const enableCompressor = ref<boolean>(false)

const handleSetImageSource = (source: ImageSource) => {
  imageToAdd.value.image_source = source

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
