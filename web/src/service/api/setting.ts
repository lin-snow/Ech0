import { request } from '../request'

// 获取系统设置
export function fetchGetSettings() {
  return request<App.Api.Setting.SystemSetting>({
    url: '/settings',
    method: 'GET',
  })
}

// 获取评论设置
export function fetchGetCommentSettings() {
  return request<App.Api.Setting.CommentSetting>({
    url: '/comment/settings',
    method: 'GET',
  })
}

// 更新系统设置
export function fetchUpdateSettings(systemSetting: App.Api.Setting.SystemSetting) {
  return request({
    url: '/settings',
    method: 'PUT',
    data: systemSetting,
  })
}

// 更新评论设置
export function fetchUpdateCommentSettings(commentSetting: App.Api.Setting.CommentSetting) {
  return request({
    url: '/comment/settings',
    method: 'PUT',
    data: commentSetting,
  })
}
