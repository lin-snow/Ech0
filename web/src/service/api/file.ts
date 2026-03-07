import { request } from '../request'

// 上传文件
export function fetchUploadFile(file: File, source?: string, category = 'image') {
  const formData = new FormData()
  formData.append('file', file)

  if (source) {
    formData.append('source', source)
  }
  formData.append('category', category)

  return request<App.Api.File.FileDto>({
    url: `/files/upload`,
    method: 'POST',
    data: formData,
  })
}

// 删除文件
export function fetchDeleteFile(file: App.Api.Ech0.FileToDelete) {
  return request({
    url: `/files/delete`,
    method: 'DELETE',
    data: file,
  })
}

