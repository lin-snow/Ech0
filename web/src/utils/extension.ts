// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ExtensionType } from '@/enums/enums'
import { parseMusicURL } from '@/utils/other'
import { parseTweetUrl } from '@/utils/tweet'

export type EchoExtensionDraft = {
  extensionType: string
  extension: string
  website: { title: string; site: string }
  location: { latitude: number | null; longitude: number | null; placeholder: string }
  tweetUrl: string
}

function isFiniteCoordinate(value: number | null): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

function isHttpUrl(value: string): boolean {
  try {
    const url = new URL(value.trim())
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

/** Builds the canonical extension payload from the editor draft. */
export function buildEchoExtension(draft: EchoExtensionDraft): App.Api.Ech0.EchoExtension | null {
  const extension = draft.extension

  switch (draft.extensionType) {
    case ExtensionType.MUSIC:
      return extension ? { type: ExtensionType.MUSIC, payload: { url: extension } } : null
    case ExtensionType.VIDEO:
      return extension ? { type: ExtensionType.VIDEO, payload: { videoId: extension } } : null
    case ExtensionType.GITHUBPROJ:
      return extension ? { type: ExtensionType.GITHUBPROJ, payload: { repoUrl: extension } } : null
    case ExtensionType.WEBSITE: {
      const { title, site } = draft.website
      return title && site ? { type: ExtensionType.WEBSITE, payload: { title, site } } : null
    }
    case ExtensionType.LOCATION: {
      const { latitude, longitude, placeholder } = draft.location
      if (!isFiniteCoordinate(latitude) || !isFiniteCoordinate(longitude)) return null
      if (latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) return null
      const name = placeholder.trim()
      if (!name) return null
      return {
        type: ExtensionType.LOCATION,
        payload: { latitude, longitude, placeholder: name },
      }
    }
    case ExtensionType.TWEET: {
      const tweet = parseTweetUrl(draft.tweetUrl)
      return tweet ? { type: ExtensionType.TWEET, payload: tweet } : null
    }
    default:
      return null
  }
}

/**
 * Cards for these types load remote data as soon as they mount, so the preview
 * must stay hidden until the payload can actually render.
 */
export function isExtensionPreviewReady(extension: App.Api.Ech0.EchoExtension): boolean {
  switch (extension.type) {
    case ExtensionType.MUSIC:
      return parseMusicURL(extension.payload.url) !== null
    case ExtensionType.VIDEO:
      return extension.payload.videoId.trim().length > 0
    case ExtensionType.GITHUBPROJ:
      return /^https?:\/\/github\.com\/[^/\s]+\/[^/\s]+/i.test(extension.payload.repoUrl.trim())
    case ExtensionType.WEBSITE:
      return isHttpUrl(extension.payload.site)
    case ExtensionType.LOCATION:
    case ExtensionType.TWEET:
      return true
    default:
      return false
  }
}
