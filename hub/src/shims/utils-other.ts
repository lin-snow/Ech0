// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const ABSOLUTE_URL_REGEX = /^https?:\/\//i
const joinBaseAndPath = (baseUrl: string, path: string) =>
  `${baseUrl.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`
const defaultServiceBaseUrl = String(import.meta.env.VITE_SERVICE_BASE_URL || '').trim()

const normalizeMediaPath = (path: string) => {
  if (path.startsWith('/api/') || path.startsWith('api/')) return path
  if (path.startsWith('/files/') || path.startsWith('files/'))
    return `/api/${path.replace(/^\/+/, '')}`
  return path
}

const resolveFileUrlByPath = (rawUrl?: string, baseUrl?: string) => {
  const candidate = String(rawUrl ?? '').trim()
  if (!candidate || ABSOLUTE_URL_REGEX.test(candidate)) return candidate
  const base = String(baseUrl ?? defaultServiceBaseUrl).trim()
  const p = normalizeMediaPath(candidate)
  return base ? joinBaseAndPath(base, p) : p
}

const resolveFileUrl = (
  file: Pick<App.Api.Ech0.FileObject | App.Api.Ech0.FileToAdd, 'url'> & { image_url?: string },
  baseUrl?: string,
) => resolveFileUrlByPath(file.url || file.image_url, baseUrl)

export const getFileUrl = (file: App.Api.Ech0.FileObject) => resolveFileUrl(file)

export const getImageUrl = (image: App.Api.Ech0.FileObject) => resolveFileUrl(image)

export const getHubImageUrl = (image: App.Api.Ech0.FileObject, baseurl: string) =>
  resolveFileUrl(image, baseurl)

export const getHubFileUrl = (file: App.Api.Ech0.FileObject, baseurl: string) =>
  resolveFileUrl(file, baseurl)
