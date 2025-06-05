<template>
  <div class="mx-auto px-2 sm:px-5 my-4">
    <!-- Todos -->
    <div v-if="isLogin && todos.length > 0">
      <div v-for="(todo, index) in todos" :key="todo.id" class="mb-4">
        <TheTodoCard :todo="todo" :index="index" :operative="true" @refresh="getTodos" />
      </div>
    </div>
    <div v-if="isLogin && todos.length === 0">
      <div class="h-auto text-gray-300 text-center font-bold text-xl">🎉今日无事，好好休息吧！</div>
    </div>
    <div v-if="!isLogin">
      <div class="h-auto text-gray-300 text-center font-bold text-xl">📦登录后查看待办事项...</div>
    </div>

    <!-- 备案号 -->
    <div class="text-center">
      <a href="https://beian.miit.gov.cn/" target="_blank">
        <span class="text-gray-400 text-sm">
          {{ SystemSetting.ICP_number }}
        </span>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheTodoCard from '@/components/advanced/TheTodoCard.vue'
import { onMounted } from 'vue'
import { useTodoStore } from '@/stores/todo'
import { useUserStore } from '@/stores/user'
import { useSettingStore } from '@/stores/settting'
import { storeToRefs } from 'pinia'

const todoStore = useTodoStore()
const settingStore = useSettingStore()
const userStore = useUserStore()
const { SystemSetting } = storeToRefs(settingStore)
const { isLogin } = storeToRefs(userStore)
const { getTodos } = todoStore
const { todos } = storeToRefs(todoStore)

onMounted(async () => {
  // 获取数据
  await getTodos()
})
</script>
