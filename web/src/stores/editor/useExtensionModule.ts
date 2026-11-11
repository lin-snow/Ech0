// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, ref, watch, type Ref } from 'vue'
import { ExtensionType } from '@/enums/enums'
import {
  buildEchoExtension,
  isExtensionPreviewReady,
  type EchoExtensionDraft,
} from '@/utils/extension'
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

  const readDraft = (): EchoExtensionDraft => ({
    extensionType: extensionToAdd.value.extension_type,
    extension: extensionToAdd.value.extension,
    website: websiteToAdd.value,
    location: locationToAdd.value,
    tweetUrl: tweetToAdd.value.url,
  })

  const hasExtension = computed(() => buildEchoExtension(readDraft()) !== null)

  // Remote cards (music / GitHub / tweet) fetch as soon as they mount, so the
  // preview waits for the draft to settle; an invalid draft hides it at once
  // instead of leaving a stale card behind.
  const PREVIEW_DEBOUNCE_MS = 350
  const extensionPreview = ref<App.Api.Ech0.EchoExtension | null>(null)
  let previewTimer: ReturnType<typeof setTimeout> | null = null

  const cancelPreviewTimer = () => {
    if (previewTimer === null) return
    clearTimeout(previewTimer)
    previewTimer = null
  }

  watch(
    [extensionToAdd, websiteToAdd, locationToAdd, tweetToAdd],
    () => {
      const next = buildEchoExtension(readDraft())
      if (!next || !isExtensionPreviewReady(next)) {
        cancelPreviewTimer()
        extensionPreview.value = null
        return
      }

      cancelPreviewTimer()
      previewTimer = setTimeout(() => {
        previewTimer = null
        extensionPreview.value = next
      }, PREVIEW_DEBOUNCE_MS)
    },
    { deep: true },
  )

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
    echoToAdd.value.extension = buildEchoExtension(readDraft())
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

  function removeExtension() {
    resetExtensionState()
    echoToAdd.value.extension = null
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
    extensionPreview,
    checkEchoExtension,
    syncEchoExtension,
    clearExtension,
    resetExtensionState,
    removeExtension,
  }
}
