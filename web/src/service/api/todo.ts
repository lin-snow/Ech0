import { request } from '../request'

export function fetchGetTodos() {
  return request<App.Api.Todo.Todo[]>({
    url: '/todo',
    method: 'GET',
  })
}

export function fetchAddTodo(todo: App.Api.Todo.TodoToAdd) {
  return request<App.Api.Todo.Todo>({
    url: '/todo',
    method: 'POST',
    data: todo,
  })
}

export function fetchUpdateTodo(id: number) {
  return request<App.Api.Todo.Todo>({
    url: `/todo/${id}`,
    method: 'PUT',
  })
}

export function fetchDeleteTodo(id: number) {
  return request<App.Api.Todo.Todo>({
    url: `/todo/${id}`,
    method: 'DELETE',
  })
}
