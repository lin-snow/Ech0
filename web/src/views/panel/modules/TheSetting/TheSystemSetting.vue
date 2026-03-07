<template>
  <PanelCard>
    <!-- 系统设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">系统设置</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="editMode" @click="handleUpdateSystemSetting" title="编辑">
            <Saveupdate class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
          <button @click="editMode = !editMode" title="编辑">
            <Edit
              v-if="!editMode"
              class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>
      <!-- 服务器&站点图标 -->
      <div class="flex justify-start items-center mb-4">
        <div class="w-28 sm:w-23">
          <img
            :src="
              (!SystemSetting?.server_logo || SystemSetting?.server_logo.length === 0) &&
              SystemSetting?.server_logo !== 'Ech0.svg'
                ? '/Ech0.svg'
                : `${API_URL}${SystemSetting?.server_logo}`
            "
            alt="头像"
            class="w-12 h-12 rounded-full ml-2 mr-9 ring-1 ring-gray-200 shadow-sm"
          />
        </div>
        <div>
          <!-- 点击上传头像 -->
          <input
            id="file-input"
            class="hidden"
            type="file"
            accept="image/*"
            ref="fileInput"
            @change="handleUploadImage"
          />
          <BaseButton
            v-if="editMode"
            class="rounded-md text-center w-auto text-align-center h-8 md:ml-5"
            @click="handTriggerUpload"
          >
            更改
          </BaseButton>
        </div>
      </div>

      <!-- 站点标题 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 mb-1"
      >
        <h2 class="font-semibold w-26 shrink-0">站点标题:</h2>
        <span v-if="!editMode">{{
          SystemSetting?.site_title.length === 0 ? '暂无' : SystemSetting.site_title
        }}</span>
        <BaseInput
          v-else
          v-model="SystemSetting.site_title"
          type="text"
          placeholder="请输入站点标题"
          class="w-full py-1!"
        />
      </div>
      <!-- 服务名称 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 mb-1"
      >
        <h2 class="font-semibold w-26 shrink-0">服务名称:</h2>
        <span v-if="!editMode">{{
          SystemSetting?.server_name.length === 0 ? '暂无' : SystemSetting.server_name
        }}</span>
        <BaseInput
          v-else
          v-model="SystemSetting.server_name"
          type="text"
          placeholder="请输入服务名称"
          class="w-full py-1!"
        />
      </div>
      <!-- 服务地址 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 mb-1"
      >
        <h2 class="font-semibold w-26 shrink-0">服务地址:</h2>
        <span v-if="!editMode">{{
          SystemSetting?.server_name.length === 0 ? '暂无' : SystemSetting.server_url
        }}</span>
        <BaseInput
          v-else
          v-model="SystemSetting.server_url"
          type="text"
          placeholder="请输入服务地址,带http(s)"
          class="w-full py-1!"
        />
      </div>
      <!-- ICP备案号 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 mb-1"
      >
        <h2 class="font-semibold w-26 shrink-0">ICP备案:</h2>
        <span
          v-if="!editMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="SystemSetting.ICP_number"
          style="vertical-align: middle"
        >
          {{ SystemSetting.ICP_number.length === 0 ? '暂无' : SystemSetting.ICP_number }}
        </span>
        <BaseInput
          v-else
          v-model="SystemSetting.ICP_number"
          type="text"
          placeholder="请输入ICP备案号"
          class="w-full py-1!"
        />
      </div>
      <!-- Meting API -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 mb-1"
      >
        <h2 class="font-semibold w-26 shrink-0">MetingAPI:</h2>
        <span
          v-if="!editMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="SystemSetting.meting_api"
          style="vertical-align: middle"
        >
          {{ SystemSetting.meting_api.length === 0 ? '暂无' : SystemSetting.meting_api }}
        </span>
        <BaseInput
          v-else
          v-model="SystemSetting.meting_api"
          type="text"
          placeholder="Meting API地址,带http(s)"
          class="w-full py-1!"
        />
      </div>
      <!-- 自定义 CSS -->
      <div class="flex flex-row justify-start text-[var(--text-color-next-500)] gap-2 mb-1">
        <h2 class="font-semibold w-26 shrink-0">自定义 CSS:</h2>
        <span
          v-if="!editMode"
          class="truncate max-w-full inline-block align-middle"
          :title="SystemSetting.custom_css"
          style="vertical-align: middle"
          >{{ SystemSetting?.custom_css?.length === 0 ? '暂无' : '******' }}</span
        >
        <BaseTextArea
          v-else
          v-model="SystemSetting.custom_css"
          type="text"
          placeholder="请输入自定义 CSS"
          class="w-full py-1!"
        />
      </div>
      <!-- 自定义 Script -->
      <div class="flex flex-row justify-start text-[var(--text-color-next-500)] gap-2 mb-1">
        <h2 class="font-semibold w-26 shrink-0">自定义 JS:</h2>
        <span
          v-if="!editMode"
          class="truncate max-w-full inline-block align-middle"
          :title="SystemSetting.custom_js"
          style="vertical-align: middle"
          >{{ SystemSetting?.custom_js?.length === 0 ? '暂无' : '******' }}</span
        >
        <BaseTextArea
          v-else
          v-model="SystemSetting.custom_js"
          type="text"
          placeholder="请输入自定义 Script"
          class="w-full py-1!"
        />
      </div>
      <!-- 允许注册 -->
      <div class="flex flex-row items-center justify-start text-[var(--text-color-next-500)]">
        <h2 class="font-semibold w-26 shrink-0">允许注册:</h2>
        <BaseSwitch v-model="SystemSetting.allow_register" :disabled="!editMode" />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { fetchUpdateSettings, fetchUploadFile } from '@/service/api'
import { ImageSource } from '@/enums/enums'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { getApiUrl } from '@/service/request/shared'

const settingStore = useSettingStore()
const { getSystemSetting } = settingStore
const { SystemSetting } = storeToRefs(settingStore)

const editMode = ref<boolean>(false)
const API_URL = getApiUrl()

const handleUpdateSystemSetting = async () => {
  await fetchUpdateSettings(settingStore.SystemSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      editMode.value = false
      // 重新获取设置
      getSystemSetting()
    })
}

const fileInput = ref<HTMLInputElement | null>(null)
const handTriggerUpload = () => {
  if (fileInput.value) {
    fileInput.value.click()
  }
}
const handleUploadImage = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  try {
    const res = await theToast.promise(fetchUploadFile(file, ImageSource.LOCAL, 'image'), {
      loading: '服务器 Logo 上传中...',
      success: '服务器 Logo 上传成功！',
      error: '上传失败，请稍后再试',
    })

    // 只需处理成功结果即可，失败的 toast 已由 request() 自动处理
    if (res.code === 1 && res.data.url) {
      SystemSetting.value.server_logo = res.data.url
    } else {
      SystemSetting.value.server_logo = '/Ech0.svg'
    }
  } catch (err) {
    console.error('上传异常', err)
    // 注意：这里只有抛出异常时才会进入，正常 res.code ≠ 1 是不会进来的
  } finally {
    target.value = ''
  }
}

onMounted(() => {
  getSystemSetting()
})
</script>

<style scoped></style>
