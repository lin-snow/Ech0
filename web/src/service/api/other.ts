import { request, downloadFile } from '../request'

// 上传音乐
export function fetchUploadMusic(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request<string>({
    url: `/audios/upload`,
    method: 'POST',
    data: formData,
  })
}

// 删除音乐
export function fetchDeleteMusic() {
  return request({
    url: `/audios/delete`,
    method: 'DELETE',
  })
}

// 获取音乐
export function fetchGetMusic() {
  return request<string>({
    url: `/getmusic`,
    method: 'GET',
  })
}

// Hello Ech0
export function fetchHelloEch0() {
  return request<App.Api.Ech0.HelloEch0>({
    url: '/hello',
    method: 'GET',
  })
}

// 执行备份
export function fetchBackup() {
  return request({
    url: '/backup',
    method: 'GET',
  })
}

// 导出备份 - 使用专门的下载函数
export function fetchExportBackup() {
  return downloadFile({
    url: '/backup/export',
    method: 'GET',
  })
}

// 导入备份
export function fetchImportBackup(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request({
    url: '/backup/import',
    method: 'POST',
    data: formData,
  })
}

// 获取网站标题
export function fetchGetWebsiteTitle(websiteURL: string) {
  return request<string>({
    url: `/website/title?website_url=${websiteURL}`,
    method: 'GET',
  })
}