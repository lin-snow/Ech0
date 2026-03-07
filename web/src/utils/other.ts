import { MusicProvider } from '@/enums/enums'

const ABSOLUTE_URL_REGEX = /^https?:\/\//i
const joinBaseAndPath = (baseUrl: string, path: string) => `${baseUrl.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`

const normalizeLegacyMediaPath = (path: string) => {
  if (path.startsWith('/api/')) return path
  if (path.startsWith('/files/') || path.startsWith('files/')) {
    return path.startsWith('/') ? `/api${path}` : `/api/${path}`
  }
  return path
}

const resolveImageUrlByPath = (rawUrl?: string, baseUrl?: string) => {
  const candidate = String(rawUrl ?? '').trim()
  if (!candidate) return ''
  if (ABSOLUTE_URL_REGEX.test(candidate)) return candidate
  const normalizedPath = normalizeLegacyMediaPath(candidate)
  if (baseUrl) return joinBaseAndPath(baseUrl, normalizedPath)
  return normalizedPath
}

const resolveImageUrl = (
  image: Pick<App.Api.Ech0.Image, 'access_url' | 'image_url'>,
  baseUrl?: string,
) => resolveImageUrlByPath(image.access_url || image.image_url, baseUrl)

// 获取图片链接
export const getImageUrl = (image: App.Api.Ech0.Image) => resolveImageUrl(image)

// 获取待添加图片链接
export const getImageToAddUrl = (image: App.Api.Ech0.ImageToAdd) => resolveImageUrl(image)

export const formatDate = (dateInput: string | number) => {
  // 当天则显示（时：分）
  // 非当天但是三内天则显示几天前
  // 超过三天则显示（时：分 年月日）

  // 处理 Unix 时间戳（秒或毫秒）和日期字符串
  let date: Date
  if (typeof dateInput === 'number') {
    // 如果是数字，判断是秒级还是毫秒级时间戳
    // 秒级时间戳通常小于 10^12，毫秒级时间戳通常大于 10^12
    date = new Date(dateInput < 1e12 ? dateInput * 1000 : dateInput)
  } else {
    date = new Date(dateInput)
  }

  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const diffInDays = Math.floor(diff / (1000 * 60 * 60 * 24))
  const diffInHours = Math.floor(diff / (1000 * 60 * 60))
  const diffInMinutes = Math.floor(diff / (1000 * 60))

  const diffInSeconds = Math.floor(diff / 1000)
  if (diffInSeconds < 60) {
    return '刚刚'
  } else if (diffInMinutes < 60) {
    return `${diffInMinutes}分钟前`
  } else if (diffInHours < 24) {
    return `${diffInHours}小时前`
  } else if (diffInDays < 3) {
    return `${diffInDays}天前`
  } else {
    const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    const weekDay = weekDays[date.getDay()]

    return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 · ${weekDay}`
  }
}

// 解析音乐链接（网易云、QQ音乐、Apple Music）
export const parseMusicURL = (url: string) => {
  url = url.trim()

  /* =======================
   * 网易云音乐
   * ======================= */
  if (/^https:\/\/([a-z0-9-]+\.)*music\.163\.com/i.test(url)) {
    // id：永远从 query 拿
    const idMatch = url.match(/[?&]id=(\d+)/)
    if (!idMatch) return null

    let type: 'song' | 'playlist' | 'album' | undefined

    if (/(\/|#\/|\/m\/)song/.test(url)) {
      type = 'song'
    } else if (/(\/|#\/|\/m\/)playlist/.test(url)) {
      type = 'playlist'
    }
    // else if (/(\/|#\/|\/m\/)album/.test(url)) {
    //   type = 'album'
    // }

    if (!type) return null

    return {
      server: MusicProvider.NETEASE,
      type,
      id: idMatch[1],
    }
  }

  /* =======================
   * QQ 音乐
   * ======================= */
  if (/^https:\/\/([a-z0-9-]+\.)*qq\.com/i.test(url)) {
    // 新版 songDetail（字母数字混合）
    const newSongMatch = url.match(/songDetail\/([a-zA-Z0-9]+)/)
    if (newSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: newSongMatch[1],
      }
    }

    // 旧版 songid（纯数字）
    const oldSongMatch = url.match(/[?&]songid=(\d+)/)
    if (oldSongMatch) {
      return {
        server: MusicProvider.QQ,
        type: 'song',
        id: oldSongMatch[1],
      }
    }

    // 匹配 /playlist/ 后面的数字ID
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

  /* =======================
   * Apple Music
   * ======================= */
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

/**
 * 从一段文本中提取并返回最短、规范的音乐链接
 */
export const extractAndCleanMusicURL = (input: string): string | null => {
  const text = input.trim()

  // 粗暴提取第一个 URL（足够鲁棒）
  const urlMatch = text.match(/https?:\/\/[^\s]+/i)
  if (!urlMatch) return null

  const rawUrl = urlMatch[0]

  // 复用统一解析函数
  const parsed = parseMusicURL(rawUrl)
  if (!parsed) return null

  // 规范化重组
  switch (parsed.server) {
    case MusicProvider.NETEASE: {
      // 统一为 PC 可打开的最短链接
      return `https://music.163.com/#/${parsed.type}?id=${parsed.id}`
    }

    case MusicProvider.QQ: {
      // 根据 type 分别生成歌曲或歌单的规范化链接
      if (parsed.type === 'song') {
        return `https://y.qq.com/n/ryqq_v2/songDetail/${parsed.id}`
      }

      if (parsed.type === 'playlist') {
        // 使用统一的歌单路径格式
        return `https://y.qq.com/n/ryqq_v2/playlist/${parsed.id}`
      }

      // 如果是其他未知类型，则返回 null
      return null
    }

    case MusicProvider.APPLE: {
      // 去掉后面的参数
      const cleanUrl = rawUrl.split('?')[0]
      return cleanUrl ?? rawUrl
    }

    default:
      return null
  }
}

// 获取 HubEcho 的图片
export const getHubImageUrl = (image: App.Api.Ech0.Image, baseurl: string) => {
  return resolveImageUrl(image, baseurl)
}

/**
 * Base64URL to Uint8Array
 * 用于解析服务端返回的 WebAuthn publicKey
 */
export function base64urlToUint8Array(input: string): Uint8Array {
  const base64 = input.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4 === 0 ? '' : '='.repeat(4 - (base64.length % 4))
  const binary = atob(base64 + pad)
  const buffer = new ArrayBuffer(binary.length)
  const bytes = new Uint8Array(buffer)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes
}

/**
 * Uint8Array to Base64URL
 * 用于生成客户端返回的 WebAuthn publicKey
 */
export function uint8ArrayToBase64url(bytes: ArrayBuffer | Uint8Array): string {
  const u8 = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  let binary = ''
  for (let i = 0; i < u8.length; i++) binary += String.fromCharCode(u8[i]!)
  const base64 = btoa(binary)
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

export function isSafari(): boolean {
  const ua = navigator.userAgent
  // Exclude Chrome, Chromium-based Edge, and iOS Chrome/Firefox
  return (
    ua.includes('Safari') &&
    !ua.includes('Chrome') &&
    !ua.includes('CriOS') &&
    !ua.includes('FxiOS')
  )
}
