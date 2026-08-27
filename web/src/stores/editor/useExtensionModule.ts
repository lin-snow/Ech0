// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, ref, type Ref } from 'vue'
import { ExtensionType } from '@/enums/enums'
import { theToast } from '@/utils/toast'
import { parseTweetUrl } from '@/utils/tweet'
import type { ExtensionToAdd, LocationToAdd, Translate, TweetToAdd, WebsiteToAdd } from './types'

type ExtensionModuleDeps = {
  echoToAdd: Ref<App.Api.Ech0.EchoToAdd>
  t: Translate
}

export function useExtensionModule({ echoToAdd, t }: ExtensionModuleDeps) {
  const websiteToAdd = ref<WebsiteToAdd>({ title: '', site: '' })
  const videoURL = ref<string>('')
  const musicURL = ref<string>('')
  const githubRepo = ref<string>('')
  const extensionToAdd = ref<ExtensionToAdd>({ extension: '', extension_type: '' })
  const locationToAdd = ref<LocationToAdd>({
    latitude: null,
    longitude: null,
    placeholder: '',
  })
  const tweetToAdd = ref<TweetToAdd>({ url: '', username: '', statusId: '' })

  const hasExtension = computed(() => {
    const ext = extensionToAdd.value.extension
    const extType = extensionToAdd.value.extension_type

    if (extType === ExtensionType.WEBSITE) {
      const { title, site } = websiteToAdd.value
      return !!title && !!site
    }

    if (extType === ExtensionType.LOCATION) {
      const { latitude, longitude, placeholder } = locationToAdd.value
      return (
        typeof latitude === 'number' &&
        typeof longitude === 'number' &&
        Number.isFinite(latitude) &&
        Number.isFinite(longitude) &&
        latitude >= -90 &&
        latitude <= 90 &&
        longitude >= -180 &&
        longitude <= 180 &&
        !!placeholder.trim()
      )
    }

    if (extType === ExtensionType.TWEET) {
      const { url, username, statusId } = tweetToAdd.value
      return !!url && !!username && !!statusId
    }

    return !!ext && !!extType
  })

  function handleWebsiteExtension(): boolean {
    const { title, site } = websiteToAdd.value

    if (title && !site) {
      theToast.error(t('editor.websiteUrlRequired'))
      return false
    }

    const finalTitle = title || (site ? t('editor.externalLink') : '')
    if (!finalTitle || !site) {
      clearExtension()
      return true
    }

    extensionToAdd.value.extension = site
    extensionToAdd.value.extension_type = ExtensionType.WEBSITE
    return true
  }

  function handleLocationExtension(): boolean {
    const { latitude, longitude, placeholder } = locationToAdd.value

    if (latitude === null || longitude === null) {
      clearExtension()
      return true
    }

    if (
      !Number.isFinite(latitude) ||
      !Number.isFinite(longitude) ||
      latitude < -90 ||
      latitude > 90 ||
      longitude < -180 ||
      longitude > 180
    ) {
      theToast.error(t('editor.locationCoordInvalid'))
      return false
    }

    if (!placeholder.trim()) {
      theToast.error(t('editor.locationPlaceholderRequired'))
      return false
    }

    extensionToAdd.value.extension_type = ExtensionType.LOCATION
    extensionToAdd.value.extension = `${latitude},${longitude}`
    return true
  }

  function handleTweetExtension(): boolean {
    const parsed = parseTweetUrl(tweetToAdd.value.url)
    if (!parsed) {
      theToast.error(t('editor.tweetUrlInvalid'))
      return false
    }

    tweetToAdd.value = parsed
    extensionToAdd.value.extension_type = ExtensionType.TWEET
    extensionToAdd.value.extension = parsed.url
    return true
  }

  function clearExtension() {
    extensionToAdd.value.extension = ''
    extensionToAdd.value.extension_type = ''
    locationToAdd.value = { latitude: null, longitude: null, placeholder: '' }
    tweetToAdd.value = { url: '', username: '', statusId: '' }
    echoToAdd.value.extension = null
  }

  function checkEchoExtension() {
    const { extension_type } = extensionToAdd.value
    if (!extension_type) {
      clearExtension()
      return
    }

    switch (extension_type) {
      case ExtensionType.WEBSITE:
        if (!handleWebsiteExtension()) return
        break
      case ExtensionType.LOCATION:
        if (!handleLocationExtension()) return
        break
      case ExtensionType.TWEET:
        if (!handleTweetExtension()) return
        break
      default:
        break
    }

    syncEchoExtension()
  }

  function syncEchoExtension() {
    const { extension, extension_type } = extensionToAdd.value
    if (!extension_type) {
      echoToAdd.value.extension = null
      return
    }

    switch (extension_type) {
      case ExtensionType.MUSIC:
        if (!extension) {
          echoToAdd.value.extension = null
          return
        }
        echoToAdd.value.extension = {
          type: ExtensionType.MUSIC,
          payload: { url: extension },
        }
        return
      case ExtensionType.VIDEO:
        if (!extension) {
          echoToAdd.value.extension = null
          return
        }
        echoToAdd.value.extension = {
          type: ExtensionType.VIDEO,
          payload: { videoId: extension },
        }
        return
      case ExtensionType.GITHUBPROJ:
        if (!extension) {
          echoToAdd.value.extension = null
          return
        }
        echoToAdd.value.extension = {
          type: ExtensionType.GITHUBPROJ,
          payload: { repoUrl: extension },
        }
        return
      case ExtensionType.WEBSITE: {
        const { title, site } = websiteToAdd.value
        if (!title || !site) {
          echoToAdd.value.extension = null
          return
        }
        echoToAdd.value.extension = {
          type: ExtensionType.WEBSITE,
          payload: { title, site },
        }
        return
      }
      case ExtensionType.LOCATION: {
        const { latitude, longitude, placeholder } = locationToAdd.value
        if (latitude === null || longitude === null || !placeholder.trim()) {
          echoToAdd.value.extension = null
          return
        }
        echoToAdd.value.extension = {
          type: ExtensionType.LOCATION,
          payload: { latitude, longitude, placeholder: placeholder.trim() },
        }
        return
      }
      case ExtensionType.TWEET: {
        const { url, username, statusId } = tweetToAdd.value
        if (!url || !username || !statusId) {
          echoToAdd.value.extension = null
          return
        }
        echoToAdd.value.extension = {
          type: ExtensionType.TWEET,
          payload: { url, username, statusId },
        }
        return
      }
      default:
        echoToAdd.value.extension = null
    }
  }

  function resetExtensionState() {
    videoURL.value = ''
    musicURL.value = ''
    githubRepo.value = ''
    extensionToAdd.value = { extension: '', extension_type: '' }
    locationToAdd.value = { latitude: null, longitude: null, placeholder: '' }
    websiteToAdd.value = { title: '', site: '' }
    tweetToAdd.value = { url: '', username: '', statusId: '' }
  }

  return {
    websiteToAdd,
    videoURL,
    musicURL,
    githubRepo,
    extensionToAdd,
    locationToAdd,
    tweetToAdd,
    hasExtension,
    checkEchoExtension,
    syncEchoExtension,
    clearExtension,
    resetExtensionState,
  }
}
