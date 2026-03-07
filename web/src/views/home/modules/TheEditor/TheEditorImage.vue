<template>
  <!-- 图片预览 -->
  <div
    v-if="
      imagesToAdd &&
      imagesToAdd.length > 0 &&
      (currentMode === Mode.ECH0 || currentMode === Mode.Image)
    "
    class="relative rounded-lg shadow-lg w-5/6 mx-auto my-7"
  >
    <button
      @click="handleRemoveImage"
      class="absolute -top-3 -right-4 bg-red-100 hover:bg-red-300 text-[var(--text-color-600)] rounded-lg w-7 h-7 flex items-center justify-center shadow"
      title="移除图片"
    >
      <Close class="w-4 h-4" />
    </button>
    <div class="rounded-lg overflow-hidden">
      <template v-for="(img, idx) in imagesToAdd" :key="idx">
        <a
          :href="getImageToAddUrl(img)"
          data-fancybox="gallery"
          :data-thumb="getImageToAddUrl(img)"
          :class="{ hidden: idx !== imageIndex }"
        >
          <img
            :src="getImageToAddUrl(img)"
            alt="Image"
            class="max-w-full object-cover"
            loading="lazy"
          />
        </a>
      </template>
    </div>
  </div>
  <!-- 图片切换 -->
  <div v-if="imagesToAdd.length > 1" class="flex items-center justify-center">
    <button @click="imageIndex = Math.max(imageIndex - 1, 0)">
      <Prev class="w-7 h-7" />
    </button>
    <span class="text-[var(--text-color-500)] text-sm mx-2">
      {{ imageIndex + 1 }} / {{ imagesToAdd.length }}
    </span>
    <button @click="imageIndex = Math.min(imageIndex + 1, imagesToAdd.length - 1)">
      <Next class="w-7 h-7" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import Next from '@/components/icons/next.vue'
import Prev from '@/components/icons/prev.vue'
import Close from '@/components/icons/close.vue'
import { getImageToAddUrl } from '@/utils/other'
import { fetchDeleteFile } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useEchoStore, useEditorStore } from '@/stores'
import { Mode } from '@/enums/enums'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { ImageSource } from '@/enums/enums'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()

// const images = defineModel<App.Api.Ech0.ImageToAdd[]>('imagesToAdd', { required: true })

// const { currentMode } = defineProps<{
//   currentMode: Mode
// }>()

// const emit = defineEmits(['handleAddorUpdateEcho'])

const imageIndex = ref<number>(0) // 临时图片索引变量
const echoStore = useEchoStore()
const { echoToUpdate } = storeToRefs(echoStore)
const editorStore = useEditorStore()
const { imagesToAdd, currentMode, isUpdateMode } = storeToRefs(editorStore)

const handleRemoveImage = () => {
  if (
    imageIndex.value < 0 ||
    imageIndex.value >= imagesToAdd.value.length ||
    imagesToAdd.value.length === 0
  ) {
    theToast.error('当前图片索引无效，无法删除！')
    return
  }
  const index = imageIndex.value

  openConfirm({
    title: '确定要移除图片吗？',
    description: '',
    onConfirm: () => {
      const imageToDel: App.Api.Ech0.FileToDelete = {
        url: String(imagesToAdd.value[index]?.image_url),
        source: String(imagesToAdd.value[index]?.image_source),
        object_key: imagesToAdd.value[index]?.object_key,
      }

      if (imageToDel.source === ImageSource.LOCAL || imageToDel.source === ImageSource.S3) {
        fetchDeleteFile({
          url: imageToDel.url,
          source: imageToDel.source,
          object_key: imageToDel.object_key,
        }).then(() => {
          // 这里不管图片是否远程删除成功都强制删除图片
          // 从数组中删除图片
          imagesToAdd.value.splice(index, 1)

          // 如果删除成功且当前处于Echo更新模式，则需要立马执行更新（图片删除操作不可逆，需要立马更新确保后端数据同步）
          if (isUpdateMode.value && echoToUpdate.value) {
            editorStore.handleAddOrUpdateEcho(true)
          }
        })
      } else {
        imagesToAdd.value.splice(index, 1)
      }

      imageIndex.value = 0
    },
  })
}

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
})
</script>

<style scoped></style>
