import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchGetTodos } from '@/service/api'

export const useTodoStore = defineStore('todoStore', () => {
  /**
   * State
   */
  const todos = ref<App.Api.Todo.Todo[]>([])
  const todoMode = ref<boolean>(false)

  /**
   * Actions
   */
  async function getTodos() {
    const res = await fetchGetTodos()
    todos.value = res.data
  }

  function setTodoMode(mode: boolean) {
    todoMode.value = mode
  }

  return {
    todos,
    todoMode,
    setTodoMode,
    getTodos,
  }
})
