<template>
  <div class="flex flex-row items-center justify-between px-2">
    <div class="flex flex-row items-center gap-2">
      <!-- ShowMore -->
      <div>
        <BaseButton
          :icon="currentMode === Mode.ECH0 ? Advance : Back"
          @click="handleChangeMode"
          :class="['w-8 h-8 sm:w-9 sm:h-9 rounded-md'].join(' ')"
          title="其它"
        />
      </div>
      <!-- Photo Upload -->
      <div v-if="currentMode === Mode.ECH0">
        <BaseButton
          :icon="ImageUpload"
          @click="handleAddImageMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="添加图片"
        />
      </div>
      <!-- Privacy Set -->
      <div v-if="currentMode === Mode.ECH0">
        <BaseButton
          :icon="echoToAdd.private ? Private : Public"
          @click="handlePrivate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="是否私密"
        />
      </div>
      <!-- Tag Add or Select -->
      <div v-if="currentMode === Mode.ECH0">
        <div
          class="flex items-center justify-between rounded-sm border border-[var(--tag-editor-border-color)] border-dashed px-1"
        >
          <span class="text-gray-300">#</span>
          <BaseCombobox
            :key="tagOptions.length"
            v-model="tagToAdd"
            :multiple="false"
            :options="tagOptions"
            placeholder="标签"
            class="rounded-sm border-none w-auto"
            input-class="w-16 h-7 text-gray-500"
          />
        </div>
      </div>
    </div>

    <div class="flex flex-row items-center gap-2">
      <!-- Clear -->
      <!-- <div>
            <BaseButton
              :icon="Clear"
              @click="handleClear"
              class="w-8 h-8 rounded-md"
              title="清空输入和图片"
            />
          </div> -->
      <!-- Publish -->
      <div
        v-if="
          currentMode !== Mode.Panel &&
          currentMode !== Mode.TagManage &&
          currentMode !== Mode.PlayMusic &&
          isUpdateMode === false
        "
      >
        <BaseButton
          :icon="Publish"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="发布Echo / Todo"
        />
      </div>
      <!-- Exit Update -->
      <div
        v-if="
          currentMode !== Mode.Panel &&
          currentMode !== Mode.TODO &&
          currentMode !== Mode.PlayMusic &&
          isUpdateMode === true
        "
      >
        <BaseButton
          :icon="ExitUpdate"
          @click="handleExitUpdateMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="退出更新模式"
        />
      </div>
      <!-- Update -->
      <div
        v-if="
          currentMode !== Mode.Panel &&
          currentMode !== Mode.TODO &&
          currentMode !== Mode.PlayMusic &&
          isUpdateMode === true
        "
      >
        <BaseButton
          :icon="Update"
          @click="handleAddorUpdate"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-md"
          title="更新Echo"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Advance from '@/components/icons/advance.vue'
import ImageUpload from '@/components/icons/image.vue'
import Public from '@/components/icons/public.vue'
import Private from '@/components/icons/private.vue'
import Publish from '@/components/icons/publish.vue'
import Update from '@/components/icons/update.vue'
import ExitUpdate from '@/components/icons/exitupdate.vue'
import Back from '@/components/icons/back.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCombobox from '@/components/common/BaseCombobox.vue'
import { ImageSource, Mode } from '@/enums/enums'
import { storeToRefs } from 'pinia'
import { useEditorStore } from '@/stores/editor'
import { useEchoStore } from '@/stores/echo'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'

const editorStore = useEditorStore()
const { currentMode, isUpdateMode, echoToAdd, mediaToAdd, tagToAdd } = storeToRefs(editorStore)
const echoStore = useEchoStore()
const { tagOptions } = storeToRefs(echoStore)

const handleAddorUpdate = () => {
  editorStore.handleAddOrUpdate()
}

const handleChangeMode = () => {
  editorStore.toggleMode()
}

const handleAddImageMode = () => {
  if (mediaToAdd.value.media_source === '') {
    mediaToAdd.value.media_source = ImageSource.LOCAL
  }

  // 检查localStg中是否有记忆的上传方式
  const rememberedSource = localStg.getItem<ImageSource>('image_source')
  if (rememberedSource) {
    mediaToAdd.value.media_source = rememberedSource
  }

  editorStore.setMode(Mode.Image)
}

const handlePrivate = () => {
  editorStore.togglePrivate()
  theToast.info('已切换为 ' + (echoToAdd.value.private ? '私密' : '公开') + ' 状态')
}

const handleExitUpdateMode = () => {
  editorStore.handleExitUpdateMode()
}
</script>

<style scoped></style>
