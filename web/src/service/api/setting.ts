import { request } from '../request'

// 获取系统设置
export function fetchGetSettings() {
  return request<App.Api.Setting.SystemSetting>({
    url: '/settings',
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
