import { request, requestWithDirectUrlAndData } from '../request'

// 分页获取Echos
export function fetchGetEchosByPage(searchParams: App.Api.Ech0.ParamsByPagination) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/page`,
    method: 'POST',
    data: searchParams,
  })
}

// 上传图片
export function fetchUploadImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request<string>({
    url: `/images/upload`,
    method: 'POST',
    data: formData,
  })
}

// 添加Echo
export function fetchAddEcho(echoToAdd: App.Api.Ech0.EchoToAdd) {
  return request({
    url: `/echo`,
    method: 'POST',
    data: echoToAdd,
  })
}

// 删除Echo
export function fetchDeleteEcho(echoId: number) {
  return request({
    url: `/echo/${echoId}`,
    method: 'DELETE',
  })
}

// 更新Echo
export function fetchUpdateEcho(echo: App.Api.Ech0.EchoToUpdate) {
  return request({
    url: `/echo`,
    method: 'PUT',
    data: echo,
  })
}

// 点赞Echo
export function fetchLikeEcho(echoId: number) {
  return request({
    url: `/echo/like/${echoId}`,
    method: 'PUT',
  })
}

// 获取Echo详情
export function fetchGetEchoById(echoId: string) {
  return request<App.Api.Ech0.Echo>({
    url: `/echo/${echoId}`,
    method: 'GET',
  })
}

// 获取status
export function fetchGetStatus() {
  return request<App.Api.Ech0.Status>({
    url: `/status`,
    method: 'GET',
  })
}

// 获取一个月内的热力图
export function fetchGetHeatMap() {
  return request<App.Api.Ech0.HeatMap>({
    url: `/heatmap`,
    method: 'GET',
  })
}

// 删除Image
export function fetchDeleteImage(image: App.Api.Ech0.ImageToDelete) {
  return request({
    url: `/images/delete`,
    method: 'DELETE',
    data: image,
  })
}

// 获取Github仓库数据
export function fetchGetGithubRepo(githubRepo: { owner: string; repo: string }) {
  return requestWithDirectUrlAndData<App.Api.Ech0.GithubCardData>({
    dirrectUrlAndData: `https://api.github.com/repos/${githubRepo.owner}/${githubRepo.repo}`,
    url: `/github`,
    method: 'GET',
  })
}

// 获取预签名URL
export function fetchGetPresignedUrl(fileName: string, contentType?: string) {
  return request<App.Api.Ech0.PresignResult>({
    url: `/s3/presign`,
    method: 'PUT',
    data: {
      file_name: fileName,
      content_type: contentType,
    },
  })
}

// 获取标签列表
export function fetchGetTags() {
  return request<App.Api.Ech0.Tag[]>({
    url: `/tags`,
    method: 'GET',
  })
}

// 删除某个标签
export function fetchDeleteTagById(tagId: number) {
  return request({
    url: `/tag/${tagId}`,
    method: 'DELETE',
  })
}

// 根据标签查询Echos（支持分页）
export function fetchGetEchosByTagId(tagId: number, searchParams: App.Api.Ech0.ParamsByPagination) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/tag/${tagId}?page=${searchParams.page}&pageSize=${searchParams.pageSize}&search=${searchParams.search || ''}`,
    method: 'GET',
  })
}

// 上传3D模型
export function fetchUploadModel(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request<string>({
    url: `/models/upload`,
    method: 'POST',
    data: formData,
  })
}

// 删除3D模型
export function fetchDeleteModel(url: string) {
  return request({
    url: `/models/delete`,
    method: 'DELETE',
    data: { url },
  })
}
