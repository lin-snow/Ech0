// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { MusicProvider } from '@/enums/enums'
import { i18n, DEFAULT_LOCALE } from '@/locales'
import { timeValueToMs } from './timeValue'

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
  const path = normalizeMediaPath(candidate)
  return base ? joinBaseAndPath(base, path) : path
}

const resolveFileUrl = (
  file: Pick<App.Api.Ech0.FileObject | App.Api.Ech0.FileToAdd, 'url'> & { image_url?: string },
  baseUrl?: string,
) => resolveFileUrlByPath(file.url || file.image_url, baseUrl)

export const getFileUrl = (file: App.Api.Ech0.FileObject) => resolveFileUrl(file)

export const getFileToAddUrl = (file: App.Api.Ech0.FileToAdd) => resolveFileUrl(file)

export const getImageUrl = (image: App.Api.Ech0.FileObject) => getFileUrl(image)
export const getImageToAddUrl = (image: App.Api.Ech0.FileToAdd) => getFileToAddUrl(image)

export const formatDateTime = (dateInput: string | number | null | undefined) => {
  const ms = timeValueToMs(dateInput)
  if (ms <= 0) return ''
  const locale = i18n.global.locale.value || DEFAULT_LOCALE
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(ms))
}

export const formatDate = (dateInput: string | number | null | undefined) => {
  const ms = timeValueToMs(dateInput)
  if (ms <= 0) return ''
  const date = new Date(ms)

  const now = new Date()
  const diff = now.getTime() - date.getTime()

  const locale = i18n.global.locale.value || DEFAULT_LOCALE
  const t = (key: string, params?: Record<string, unknown>) =>
    String(i18n.global.t(key, params || {}))

  const MS_DAY = 24 * 60 * 60 * 1000

  const longFormatter = () => {
    const datePart = new Intl.DateTimeFormat(locale, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    }).format(date)
    const weekdayPart = new Intl.DateTimeFormat(locale, { weekday: 'short' }).format(date)
    return `${datePart} · ${weekdayPart}`
  }

  if (diff < 0) {
    return longFormatter()
  }

  const diffInSeconds = Math.floor(diff / 1000)
  const diffInMinutes = Math.floor(diff / (1000 * 60))
  const diffInHours = Math.floor(diff / (1000 * 60 * 60))

  if (diff < MS_DAY) {
    if (diffInSeconds < 60) {
      return t('dateTime.justNow')
    }
    if (diffInMinutes < 60) {
      return t('dateTime.minutesAgo', { count: diffInMinutes })
    }
    return t('dateTime.hoursAgo', { count: diffInHours })
  }

  if (diff < 2 * MS_DAY) {
    return t('dateTime.daysAgo', { count: 1 })
  }

  if (diff < 3 * MS_DAY) {
    return t('dateTime.daysAgo', { count: 2 })
  }

  return longFormatter()
}

export const parseMusicURL = (url: string) => {
  url = url.trim()

  if (/^https:\/\/([a-z0-9-]+\.)*music\.163\.com/i.test(url)) {
    const idMatch = url.match(/[?&]id=(\d+)/)
    if (!idMatch) return null

    let type: 'song' | 'playlist' | 'album' | undefined

    if (/(\/|#\/|\/m\/)song/.test(url)) {
      type = 'song'
    } else if (/(\/|#\/|\/m\/)playlist/.test(url)) {
      type = 'playlist'
    }

    if (!type) return null

    return {
      server: MusicProvider.NETEASE,
      type,
      id: idMatch[1],
    }
  }

  if (/^https:\/\/([a-z0-9-]+\.)*qq\.com/i.test(url)) {
    const newSongMatch = url.match(/songDetail\/([a-zA-Z0-9]+)/)
    if (newSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: newSongMatch[1],
      }
    }

    const oldSongMatch = url.match(/[?&]songid=(\d+)/)
    if (oldSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: oldSongMatch[1],
      }
    }

    const playlistMatch = url.match(/\/playlist\/(\d+)/i)
    if (playlistMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'playlist',
        id: playlistMatch[1],
      }
    }
    return null
  }

  if (/^https:\/\/music\.apple\.com/i.test(url)) {
    const appleMatch = url.match(/\/(song|album)\/[^/]+\/(\d+)/)
    if (!appleMatch) return null

    return {
      server: MusicProvider.APPLE,
      type: appleMatch[1],
      id: appleMatch[2],
    }
  }
  return null
}

export const extractAndCleanMusicURL = (input: string): string | null => {
  const text = input.trim()

  const urlMatch = text.match(/https?:\/\/[^\s]+/i)
  if (!urlMatch) return null

  const rawUrl = urlMatch[0]

  const parsed = parseMusicURL(rawUrl)
  if (!parsed) return null

  switch (parsed.server) {
    case MusicProvider.NETEASE: {
      return `https://music.163.com/#/${parsed.type}?id=${parsed.id}`
    }

    case MusicProvider.QQ: {
      if (parsed.type === 'song') {
        return `https://y.qq.com/n/ryqq_v2/songDetail/${parsed.id}`
      }

      if (parsed.type === 'playlist') {
        return `https://y.qq.com/n/ryqq_v2/playlist/${parsed.id}`
      }

      return null
    }

    case MusicProvider.APPLE: {
      const cleanUrl = rawUrl.split('?')[0]
      return cleanUrl ?? rawUrl
    }

    default:
      return null
  }
}

export const getHubImageUrl = (image: App.Api.Ech0.FileObject, baseurl: string) => {
  return resolveFileUrl(image, baseurl)
}

export const getHubFileUrl = (file: App.Api.Ech0.FileObject, baseurl: string) => {
  return resolveFileUrl(file, baseurl)
}

export function base64urlToUint8Array(input: string): Uint8Array {
  const base64 = input.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4 === 0 ? '' : '='.repeat(4 - (base64.length % 4))
  const binary = atob(base64 + pad)
  const buffer = new ArrayBuffer(binary.length)
  const bytes = new Uint8Array(buffer)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}

export function uint8ArrayToBase64url(bytes: ArrayBuffer | Uint8Array): string {
  const u8 = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  let binary = ''
  for (let i = 0; i < u8.length; i++) binary += String.fromCharCode(u8[i]!)
  const base64 = btoa(binary)
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

export function isSafari(): boolean {
  const ua = navigator.userAgent
  return (
    ua.includes('Safari') &&
    !ua.includes('Chrome') &&
    !ua.includes('CriOS') &&
    !ua.includes('FxiOS')
  )
}
