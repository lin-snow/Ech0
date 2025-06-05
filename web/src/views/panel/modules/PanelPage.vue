<template>
  <div class="px-4 pb-4 py-2 mt-4 mb-10 mx-auto flex justify-center items-center min-h-screen">
    <div class="h-auto max-w-sm sm:max-w-md md:max-w-lg px-2 my-4">
      <h1 class="text-6xl italic font-bold text-center text-gray-300 mb-5">Ech0s Panel</h1>
      <!-- 返回首页 / 登录 / 注册 -->
      <div class="flex justify-between items-center mb-5">
        <!-- 返回首页 -->
        <div class="">
          <BaseButton
            @click="router.push('/')"
            class="text-gray-600 rounded-md !shadow-none !border-none !ring-0 !bg-transparent group"
            title="返回首页"
          >
            <Arrow
              class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
            />
          </BaseButton>
        </div>
        <!-- 操作按钮 -->
        <div class="flex flex-row items-center gap-2">
          <!-- 状态 / 设置 / 个人中心 -->
          <BaseButton
            :icon="[Setting, User, Status][ShowingIndex]"
            @click="changeShow"
            class="text-gray-600 rounded-md w-8 h-8 sm:w-9 sm:h-9"
            title="状态 / 设置 / 个人中心"
          />

          <!-- 退出登录 -->
          <BaseButton
            :icon="Logout"
            @click="handleLogout"
            class="text-gray-600 rounded-md w-8 h-8 sm:w-9 sm:h-9"
            title="退出登录"
          />

          <!-- 登录 / 注册 -->
          <BaseButton
            :icon="Auth"
            @click="router.push('/auth')"
            class="text-gray-600 rounded-md w-8 h-8 sm:w-9 sm:h-9"
            title="登录 / 注册"
          />
        </div>
      </div>
      <!-- TheStatus -->
      <TheStatus v-if="Showing === ShowWhichEnum.Status" />
      <!-- TheSetting -->
      <TheSetting v-if="Showing === ShowWhichEnum.Setting" />
      <!-- TheUserCenter -->
      <TheUser v-if="Showing === ShowWhichEnum.UserCenter" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheStatus from './TheStatus.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import User from '@/components/icons/user.vue'
import Auth from '@/components/icons/auth.vue'
import Status from '@/components/icons/status.vue'
import Setting from '@/components/icons/setting.vue'
import TheSetting from './TheSetting.vue'
import TheUser from './TheUser.vue'
import Logout from '@/components/icons/logout.vue'
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()
// 控制面板的状态
enum ShowWhichEnum {
  Status = 'status',
  Setting = 'setting',
  UserCenter = 'usercenter',
}
const ShowingArray = [ShowWhichEnum.Status, ShowWhichEnum.Setting, ShowWhichEnum.UserCenter]
const ShowingIndex = ref<number>(0)
const Showing = ref<string>(ShowWhichEnum.Status)

const changeShow = () => {
  // 切换状态
  ShowingIndex.value = (ShowingIndex.value + 1) % ShowingArray.length
  Showing.value = ShowingArray[ShowingIndex.value]
}

const handleLogout = () => {
  // 弹出浏览器确认框
  if (confirm('确定要退出登录吗？')) {
    // 清除用户信息
    userStore.logout()
  }
}
</script>
