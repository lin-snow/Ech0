<template>
  <div class="w-full px-2">
    <div class="rounded-md shadow-sm ring-1 ring-gray-200 ring-inset bg-white p-4 mb-2">
      <!-- 设置 -->
      <div>
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
                : `${API_URL}${user?.avatar}`
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
              class="rounded-md text-center w-auto text-align-center h-8"
              @click="handTriggerUpload"
            >
              更改
            </BaseButton>
          </div>
        </div>

        <!-- 用户名 -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30">用户名:</h2>
          <span v-if="!editMode">{{ user?.username }}</span>
          <BaseInput
            v-else
            v-model="userInfo.username"
            type="text"
            placeholder="请输入用户名"
            class="w-36 !py-1"
          />
        </div>

        <!-- 密码 -->
        <div class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10">
          <h2 class="font-semibold w-30">密码:</h2>
          <span v-if="!editMode">******</span>
          <BaseInput
            v-else
            v-model="userInfo.password"
            type="password"
            placeholder="请输入密码"
            class="w-36 !py-1"
            autocomplete="off"
          />
        </div>
      </div>
    </div>

    <div class="rounded-md shadow-sm ring-1 ring-gray-200 ring-inset bg-white p-4">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-gray-600 font-bold text-lg">用户管理</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button @click="userEditMode = !userEditMode" title="编辑">
            <Edit v-if="!userEditMode" class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
            <Close v-else class="w-5 h-5 text-gray-400 hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 用户列表 -->
      <div v-if="allusers.length === 0">
        <h2 class="text-gray-500 font-semibold text-center">暂无其它用户</h2>
      </div>
      <div v-else>
        <div
          v-for="user in allusers"
          :key="user.id"
          class="flex flex-row items-center justify-start text-gray-500 gap-2 h-10"
        >
          <h2 class="font-semibold w-30">{{ user.username }}</h2>
          <BaseSwitch
            v-model="user.is_admin"
            :disabled="!userEditMode"
            class="w-14"
            @click="handleUpdateUserPermission(user.id)"
          />
          <BaseButton
            :icon="Deluser"
            class="rounded-md text-center w-auto text-align-center h-8"
            :disabled="!userEditMode"
            @click="handleDeleteUser(user.id)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Deluser from '@/components/icons/deluser.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import {
  fetchGetCurrentUser,
  fetchUpdateUser,
  fetchUploadImage,
  fetchGetAllUsers,
  fetchUpdateUserPermission,
  fetchDeleteUser,
} from '@/service/api'
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
const allusers = ref<App.Api.User.User[]>([])
const editMode = ref<boolean>(false)
const userEditMode = ref<boolean>(false)
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

const handleDeleteUser = async (userId: number) => {
  if (confirm('确定要删除该用户吗？')) {
    fetchDeleteUser(userId).then((res) => {
      if (res.code === 1) {
        getAllUsers()
      }
    })
  }
}

const handleUpdateUserPermission = async (userId: number) => {
  fetchUpdateUserPermission(userId)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      // 重新获取设置
      getAllUsers()
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
  if (target.files && target.files.length > 0) {
    const file = target.files[0]
    fetchUploadImage(file).then((res) => {
      if (res.code === 1) {
        userInfo.value.avatar = res.data
        if (user.value) user.value.avatar = res.data
        theToast.success('头像上传成功！')
      }
    })
  }
}

const getAllUsers = async () => {
  await fetchGetAllUsers().then((res) => {
    if (res.code === 1) {
      allusers.value = res.data
    }
  })
}

onMounted(async () => {
  await fetchGetCurrentUser().then((res) => {
    if (res.code === 1) {
      userInfo.value.username = res.data.username
      userInfo.value.password = res.data.password || ''
      userInfo.value.avatar = res.data.avatar || ''
      userInfo.value.is_admin = res.data.is_admin
    }
  })

  await getAllUsers()
})
</script>
