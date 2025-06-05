<template>
  <div class="flex justify-center items-center h-screen">
    <div class="h-1/2 max-w-sm sm:max-w-md md:max-w-lg">
      <h1 class="text-6xl italic font-bold text-center text-gray-300 mb-4">Ech0s~</h1>
      <!-- 登录 / 注册 -->
      <div v-if="AuthMode === 'login'">
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-bold text-gray-400 mb-3">登录</h2>
          <div class="mb-3">
            <button
              @click="AuthMode = 'register'"
              class="text-gray-500 hover:text-gray-700 transition duration-200"
            >
              <div class="flex flex-row gap-0 items-center">
                注册
                <Arrow class="text-2xl" />
              </div>
            </button>
          </div>
        </div>
        <BaseInput v-model="username" type="text" placeholder="请输入用户名" class="mb-4" />
        <BaseInput v-model="password" type="password" placeholder="请输入密码" class="mb-4" />
        <div class="flex justify-between items-center px-0.5">
          <BaseButton
            @click="router.push({ name: 'home' })"
            title="返回首页"
            :icon="Home"
            class="rounded-md w-9 h-9"
          />
          <BaseButton @click="handleLogin" class="rounded-md">
            <span class="text-gray-500">登录</span>
          </BaseButton>
        </div>
      </div>
      <div v-else-if="AuthMode === 'register'">
        <div class="flex justify-between items-center">
          <h2 class="text-lg font-bold text-gray-400 mb-3">注册</h2>
          <div class="mb-3">
            <button
              @click="AuthMode = 'login'"
              class="text-gray-500 hover:text-gray-700 transition duration-200"
            >
              <div class="flex flex-row gap-0 items-center">
                登录
                <Arrow class="text-2xl rotate-180" />
              </div>
            </button>
          </div>
        </div>
        <BaseInput v-model="username" type="text" placeholder="请输入用户名" class="mb-4" />
        <BaseInput v-model="password" type="password" placeholder="请输入密码" class="mb-4" />
        <div class="flex justify-between items-center px-0.5">
          <BaseButton
            @click="router.push({ name: 'home' })"
            title="返回首页"
            :icon="Home"
            class="rounded-md w-9 h-9"
          />
          <BaseButton @click="handleRegister" class="rounded-md">
            <span class="text-gray-500">注册</span>
          </BaseButton>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useUserStore } from '@/stores/user'
import Arrow from '@/components/icons/arrow.vue'
import Home from '@/components/icons/home.vue'

const AuthMode = ref<'login' | 'register'>('login') // login / register
const username = ref<string>('')
const password = ref<string>('')
const userStore = useUserStore()
const router = useRouter()

const handleLogin = async () => {
  // console.log('登录', username.value, password.value)
  await userStore.login({
    username: username.value,
    password: password.value,
  })
}

const handleRegister = async () => {
  // console.log('注册', username.value, password.value)
  if (
    await userStore.signup({
      username: username.value,
      password: password.value,
    })
  ) {
    // 注册成功，切换到登录模式
    AuthMode.value = 'login'
  }
}
</script>
