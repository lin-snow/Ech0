<template>
  <PanelCard>
    <!-- 用户设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-gray-600 font-bold text-lg">用户中心</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="editMode" @click="handleUpdateUser" title="编辑">
            <Saveupdate class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
          </button>
          <button @click="editMode = !editMode" title="编辑">
            <Edit v-if="!editMode" class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            <Close v-else class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 头像 -->
      <div class="flex justify-start items-center">
        <img
          :src="
            !user?.avatar || user?.avatar.length === 0
              ? '/favicon.svg'
              : user.avatar.startsWith('http')
                ? user.avatar
                : `${API_URL}${user.avatar}`
          "
          alt="头像"
          class="w-12 h-12 rounded-full ml-2 mr-9 ring-1 ring-gray-200 shadow-sm"
        />
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

      <!-- 用户名 -->
      <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
        <h2 class="font-semibold w-30">用户名:</h2>
        <span v-if="!editMode">{{ user?.username }}</span>
        <BaseInput
          v-else
          v-model="userInfo.username"
          type="text"
          placeholder="请输入用户名"
          class="w-36 py-1!"
        />
      </div>

      <!-- 密码 -->
      <div class="flex flex-row items-center justify-start text-stone-500 gap-2 h-10">
        <h2 class="font-semibold w-30">密码:</h2>
        <span v-if="!editMode">******</span>
        <BaseInput
          v-else
          v-model="userInfo.password"
          type="password"
          placeholder="请输入密码"
          class="w-36 py-1!"
          autocomplete="off"
        />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'

import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { fetchGetCurrentUser, fetchUpdateUser, fetchUploadImage } from '@/service/api'
import { theToast } from '@/utils/toast'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/user'
import { getApiUrl } from '@/service/request/shared'

const userStore = useUserStore()
const { refreshCurrentUser } = userStore
const { user } = storeToRefs(userStore)
const userInfo = ref<App.Api.User.UserInfo>({
  username: '',
  password: '',
  is_admin: false,
  avatar: '',
})

const editMode = ref<boolean>(false)
const API_URL = getApiUrl()

const handleUpdateUser = async () => {
  await fetchUpdateUser(userInfo.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
        editMode.value = false
      }
    })
    .finally(() => {
      // 重新获取设置
      refreshCurrentUser()
    })
    .catch((err) => {
      console.error(err)
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
    const res = await theToast.promise(fetchUploadImage(file), {
      loading: '头像上传中...',
      success: '头像上传成功！',
      error: '上传失败，请稍后再试',
    })

    // 只需处理成功结果即可，失败的 toast 已由 request() 自动处理
    if (res.code === 1) {
      userInfo.value.avatar = res.data
      if (user.value) user.value.avatar = res.data
    }
  } catch (err) {
    console.error('上传异常', err)
    // 注意：这里只有抛出异常时才会进入，正常 res.code ≠ 1 是不会进来的
  } finally {
    target.value = ''
  }
}

onMounted(() => {
  fetchGetCurrentUser().then((res) => {
    if (res.code === 1) {
      userInfo.value.username = res.data.username
      userInfo.value.password = res.data.password || ''
      userInfo.value.avatar = res.data.avatar || ''
      userInfo.value.is_admin = res.data.is_admin
    }
  })
})
</script>

<style scoped></style>
