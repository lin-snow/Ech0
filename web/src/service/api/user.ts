import { request } from '../request'

// 获取当前登录用户信息
export function fetchGetCurrentUser() {
  return request<App.Api.User.User>({
    url: '/user',
    method: 'GET',
  })
}

// 更新用户信息
export function fetchUpdateUser(user: App.Api.User.UserInfo) {
  return request({
    url: '/user',
    method: 'PUT',
    data: user,
  })
}

// 获取用户列表
export function fetchGetAllUsers() {
  return request<App.Api.User.User[]>({
    url: '/allusers',
    method: 'GET',
  })
}

// 更新用户权限
export function fetchUpdateUserPermission(id: number) {
  return request({
    url: `/user/admin/${id}`,
    method: 'PUT',
  })
}

// 删除用户
export function fetchDeleteUser(id: number) {
  return request({
    url: `/user/${id}`,
    method: 'DELETE',
  })
}
