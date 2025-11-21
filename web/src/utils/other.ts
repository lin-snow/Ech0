import { getApiUrl } from '@/service/request/shared'
import { ImageSource, MusicProvider } from '@/enums/enums'

// 获取媒体链接（支持图片和视频）
export const getMediaUrl = (media: App.Api.Ech0.Media) => {
  if (media.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(media.media_url)
  } else if (media.media_source === ImageSource.URL) {
    return String(media.media_url)
  } else if (media.media_source === ImageSource.S3) {
    return String(media.media_url)
  } else {
    // 未知的媒体来源，按照本地媒体处理
    return getApiUrl() + String(media.media_url)
  }
}

// 获取图片链接（保留向后兼容）
export const getImageUrl = (image: App.Api.Ech0.Image) => {
  if (image.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(image.media_url)
  } else if (image.media_source === ImageSource.URL) {
    return String(image.media_url)
  } else if (image.media_source === ImageSource.S3) {
    return String(image.media_url)
  } else {
    // 未知的图片来源，按照本地图片处理
    return getApiUrl() + String(image.media_url)
  }
}

// 获取待添加媒体链接
export const getMediaToAddUrl = (media: App.Api.Ech0.MediaToAdd) => {
  if (media.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(media.media_url)
  } else if (media.media_source === ImageSource.URL) {
    return String(media.media_url)
  } else if (media.media_source === ImageSource.S3) {
    return String(media.media_url)
  } else {
    // 未知的媒体来源，按照本地媒体处理
    return getApiUrl() + String(media.media_url)
  }
}

// 获取待添加图片链接（向后兼容）
export const getImageToAddUrl = (image: App.Api.Ech0.ImageToAdd) => {
  if (image.media_source === ImageSource.LOCAL) {
    return getApiUrl() + String(image.media_url)
  } else if (image.media_source === ImageSource.URL) {
    return String(image.media_url)
  } else if (image.media_source === ImageSource.S3) {
    return String(image.media_url)
  } else {
    // 未知的图片来源，按照本地图片处理
    return getApiUrl() + String(image.media_url)
  }
}

export const formatDate = (dateString: string) => {
  // 当天则显示（时：分）
  // 非当天但是三内天则显示几天前
  // 超过三天则显示（时：分 年月日）
  const date = new Date(dateString)
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

  // 网易云音乐 - 修复id参数位置问题
  const neteaseMatch = url.match(/music\.163\.com\/(#\/)?(song|playlist|album)(\?.*)?/)
  if (neteaseMatch) {
    // 从查询参数中提取id
    const idMatch = url.match(/[?&]id=(\d+)/)
    if (idMatch) {
      return {
        server: MusicProvider.NETEASE,
        type: neteaseMatch[2], // song, playlist, album
        id: idMatch[1],
      }
    }
  }

  // QQ音乐 新格式支持，songDetail 路径，id一般是字母数字混合
  const qqNewSongMatch = url.match(/y\.qq\.com\/n\/ryqq\/songDetail\/([a-zA-Z0-9]+)/)
  if (qqNewSongMatch) {
    return {
      server: MusicProvider.QQ,
      type: 'song',
      id: qqNewSongMatch[1],
    }
  }

  // QQ音乐 - 旧格式playsong，支持各种子域名前缀
  const qqOldSongMatch = url.match(/(?:i\d*\.)?y\.qq\.com\/.*[?&]songid=(\d+)/)
  if (qqOldSongMatch) {
    return {
      server: MusicProvider.QQ,
      type: 'song',
      id: qqOldSongMatch[1],
    }
  }

  // Apple Music
  const appleMatch = url.match(/music\.apple\.com\/[a-z]{2}\/(song|album)\/[^/]+\/(\d+)/)
  if (appleMatch) {
    return {
      server: MusicProvider.APPLE,
      type: appleMatch[1], // song / album
      id: appleMatch[2],
    }
  }

  // 解析失败
  return null
}

/**
 * 从一段文本中提取并返回最短、最规范的音乐链接
 * @param input 包含音乐链接的原始文本
 * @returns 返回清理后的音乐链接字符串，如果找不到则返回 null
 */
export const extractAndCleanMusicURL = (input: string): string | null => {
  const text = input.trim()

  // 网易云音乐
  // 匹配并捕获完整的、包含id参数的网易云链接
  const neteaseMatch = text.match(
    /(https?:\/\/music\.163\.com\/(#\/)?(song|playlist|album)\?id=\d+)/i,
  )
  if (neteaseMatch) {
    // neteaseMatch[1] 就是第一个捕获组，即我们想要的完整链接
    return neteaseMatch?.[1]?.replace(/^http:\/\//, 'https://') ?? null // 统一为https
  }

  // QQ音乐 新格式
  // 匹配并捕获 songDetail 路径的链接
  const qqNewMatch = text.match(/(https?:\/\/y\.qq\.com\/n\/ryqq\/songDetail\/[a-zA-Z0-9]+)/i)
  if (qqNewMatch) {
    return qqNewMatch?.[1]?.replace(/^http:\/\//, 'https://') ?? null
  }

  // QQ音乐 旧格式
  // 这个稍微复杂，因为可能还有其他参数。我们只关心songid
  const qqOldMatch = text.match(/(https?:\/\/(?:i\d*\.)?y\.qq\.com\/\S*)[?&]songid=(\d+)/i)
  if (qqOldMatch) {
    const baseUrl = qqOldMatch?.[1]?.replace(/^http:\/\//, 'https://') ?? null
    const songId = qqOldMatch?.[2] ?? null
    // 重新构建最短链接，只保留songid参数
    return `${baseUrl}?songid=${songId}`
  }

  // Apple Music
  // 匹配并捕获 Apple Music 的 song 或 album 链接
  const appleMatch = text.match(
    /(https?:\/\/music\.apple\.com\/[a-z]{2}\/(song|album)\/[^\/]+\/\d+)/i,
  )
  if (appleMatch) {
    return appleMatch?.[1]?.replace(/^http:\/\//, 'https://') ?? null
  }

  // 如果没有匹配到任何规则
  return null
}

// 获取图片尺寸
export async function getImageSize(file: File): Promise<{ width: number; height: number }> {
  const bitmap = await createImageBitmap(file)
  return { width: bitmap.width, height: bitmap.height }
}

// 获取 HubEcho 的媒体链接（支持图片和视频）
export const getHubMediaUrl = (media: App.Api.Ech0.Media, baseurl: string) => {
  if (media.media_source === ImageSource.LOCAL) {
    return baseurl + '/api' + String(media.media_url)
  } else if (media.media_source === ImageSource.URL) {
    return String(media.media_url)
  } else if (media.media_source === ImageSource.S3) {
    return String(media.media_url)
  } else {
    // 未知的媒体来源，按照本地媒体处理
    return baseurl + '/api' + String(media.media_url)
  }
}

// 获取 HubEcho 的图片（保留向后兼容）
export const getHubImageUrl = (image: App.Api.Ech0.Image, baseurl: string) => {
  if (image.media_source === ImageSource.LOCAL) {
    return baseurl + '/api' + String(image.media_url)
  } else if (image.media_source === ImageSource.URL) {
    return String(image.media_url)
  } else if (image.media_source === ImageSource.S3) {
    return String(image.media_url)
  } else {
    // 未知的图片来源，按照本地图片处理
    return baseurl + '/api' + String(image.media_url)
  }
}
