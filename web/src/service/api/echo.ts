// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { request, requestWithDirectUrlAndData } from '../request'

export async function fetchQueryEchos(params: App.Api.Ech0.EchoQueryParams) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/query`,
    method: 'POST',
    data: params,
  })
}

export async function fetchGetEchosByPage(searchParams: App.Api.Ech0.ParamsByPagination) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/page`,
    method: 'POST',
    data: searchParams,
  })
}

export function fetchAddEcho(echoToAdd: App.Api.Ech0.EchoToAdd) {
  return request({
    url: `/echo`,
    method: 'POST',
    data: echoToAdd,
  })
}

export function fetchDeleteEcho(echoId: string) {
  return request({
    url: `/echo/${echoId}`,
    method: 'DELETE',
  })
}

export function fetchUpdateEcho(echo: App.Api.Ech0.EchoToUpdate) {
  return request({
    url: `/echo`,
    method: 'PUT',
    data: echo,
  })
}

export function fetchLikeEcho(echoId: string) {
  return request({
    url: `/echo/like/${echoId}`,
    method: 'PUT',
  })
}

export function fetchGetTodayEchos() {
  return request<App.Api.Ech0.Echo[]>({
    url: `/echo/today`,
    method: 'GET',
  })
}

export function fetchGetHotEchos(limit = 5) {
  return request<App.Api.Ech0.Echo[]>({
    url: `/echo/hot?limit=${limit}`,
    method: 'GET',
  })
}

export async function fetchGetEchoById(echoId: string) {
  return request<App.Api.Ech0.Echo>({
    url: `/echo/${echoId}`,
    method: 'GET',
  })
}

export function fetchGetHeatMap() {
  return request<App.Api.Ech0.HeatMap>({
    url: `/heatmap`,
    method: 'GET',
  })
}

export function fetchGetGithubRepo(githubRepo: { owner: string; repo: string }) {
  return requestWithDirectUrlAndData<App.Api.Ech0.GithubCardData>({
    dirrectUrlAndData: `https://api.github.com/repos/${githubRepo.owner}/${githubRepo.repo}`,
    url: `/github`,
    method: 'GET',
  })
}

export function fetchGetTags() {
  return request<App.Api.Ech0.Tag[]>({
    url: `/tags`,
    method: 'GET',
  })
}

export function fetchDeleteTagById(tagId: string) {
  return request({
    url: `/tag/${tagId}`,
    method: 'DELETE',
  })
}

export function fetchCreateTag(name: string) {
  return request<App.Api.Ech0.Tag>({
    url: `/tag`,
    method: 'POST',
    data: { name },
  })
}

export async function fetchGetEchosByTagId(
  tagId: string,
  searchParams: App.Api.Ech0.ParamsByPagination,
) {
  return request<App.Api.Ech0.PaginationResult>({
    url: `/echo/tag/${tagId}?page=${searchParams.page}&pageSize=${searchParams.pageSize}&search=${searchParams.search || ''}`,
    method: 'GET',
  })
}
