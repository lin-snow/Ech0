<template>
  <div
    class="w-full flex flex-col gap-2 p-4 bg-white rounded-lg ring-1 ring-gray-200 ring-inset mx-auto shadow-sm hover:shadow-md"
  >
    <!-- 顶部id + 按钮 -->
    <div v-if="props.operative" class="flex justify-between items-center">
      <!-- id -->
      <div class="flex justify-start gap-1 items-center h-auto font-bold text-2xl">
        <span class="italic text-gray-300">#</span>
        <span class="text-gray-400">{{ props.index }}</span>
      </div>
      <!-- 按钮 -->
      <div class="flex gap-2">
        <BaseButton
          :icon="Delete"
          @click="handleDeleteTodo"
          class="w-7 h-7 rounded-md !text-red-200"
          title="删除待办"
        />
        <BaseButton
          :icon="Done"
          @click="handleChangeTodoStatus"
          class="w-7 h-7 rounded-md"
          title="切换待办状态"
        />
      </div>
    </div>
    <div v-else>
      <p class="text-gray-600 font-bold text-lg flex items-center">
        <Busy class="mr-1" /> 正忙着的事情：
      </p>
    </div>
    <!-- 具体内容 -->
    <div>
      <p class="text-gray-500 text-sm whitespace-pre-wrap">
        {{ props.todo.content }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import Done from '../icons/done.vue'
import Busy from '../icons/busy.vue'
import Delete from '../icons/delete.vue'
import BaseButton from '../common/BaseButton.vue'
import { fetchUpdateTodo, fetchDeleteTodo } from '@/service/api'
import { theToast } from '@/utils/toast'

const props = defineProps<{
  todo: App.Api.Todo.Todo
  index: number
  operative: boolean
}>()

const emit = defineEmits(['refresh'])

const handleDeleteTodo = () => {
  if (confirm('确定要删除待办吗？')) {
    fetchDeleteTodo(props.todo.id).then((res) => {
      if (res.code === 1) {
        theToast.success('待办已删除！')
        emit('refresh')
      }
    })
  }
}

const handleChangeTodoStatus = () => {
  if (confirm('确定要切换待办状态吗？')) {
    fetchUpdateTodo(props.todo.id).then((res) => {
      if (res.code === 1) {
        theToast.success('待办已完成！')
        emit('refresh')
      }
    })
  }
}
</script>

<style scoped></style>
