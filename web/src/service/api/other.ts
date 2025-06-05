import { request } from '../request'

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
