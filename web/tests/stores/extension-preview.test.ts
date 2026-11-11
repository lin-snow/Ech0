// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { nextTick } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useExtensionModule } from '@/stores/editor/useExtensionModule'
import { ExtensionType } from '@/enums/enums'

const PREVIEW_DEBOUNCE_MS = 350

const setup = () => useExtensionModule({ t: (key) => String(key) })

const settle = async () => {
  await nextTick()
  vi.advanceTimersByTime(PREVIEW_DEBOUNCE_MS)
}

describe('useExtensionModule extension preview', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('publishes the preview only after the draft settles', async () => {
    const { extensionToAdd, extensionPreview } = setup()

    extensionToAdd.value.extension_type = ExtensionType.MUSIC
    extensionToAdd.value.extension = 'https://music.163.com/#/song?id=123'
    await nextTick()
    expect(extensionPreview.value).toBeNull()

    vi.advanceTimersByTime(PREVIEW_DEBOUNCE_MS)
    expect(extensionPreview.value).toEqual({
      type: ExtensionType.MUSIC,
      payload: { url: 'https://music.163.com/#/song?id=123' },
    })
  })

  it('hides the preview immediately when the draft becomes invalid', async () => {
    const { extensionToAdd, extensionPreview } = setup()

    extensionToAdd.value.extension_type = ExtensionType.MUSIC
    extensionToAdd.value.extension = 'https://music.163.com/#/song?id=123'
    await settle()
    expect(extensionPreview.value).not.toBeNull()

    extensionToAdd.value.extension = 'not-a-url'
    await nextTick()
    expect(extensionPreview.value).toBeNull()
    vi.advanceTimersByTime(PREVIEW_DEBOUNCE_MS)
    expect(extensionPreview.value).toBeNull()
  })

  it('previews tweets from the URL before submit', async () => {
    const { tweetToAdd, extensionToAdd, extensionPreview, hasExtension } = setup()

    extensionToAdd.value.extension_type = ExtensionType.TWEET
    tweetToAdd.value.url = 'https://x.com/lin_snow/status/1234567890'
    await settle()

    expect(hasExtension.value).toBe(true)
    expect(extensionPreview.value).toEqual({
      type: ExtensionType.TWEET,
      payload: {
        url: 'https://x.com/lin_snow/status/1234567890',
        username: 'lin_snow',
        statusId: '1234567890',
      },
    })
  })

  it('removes the extension draft and clears every derived field', async () => {
    const {
      extensionToAdd,
      websiteToAdd,
      locationToAdd,
      tweetToAdd,
      videoURL,
      extensionPreview,
      payload,
      removeExtension,
    } = setup()

    extensionToAdd.value = { extension: 'BV1xx411c7mD', extension_type: ExtensionType.VIDEO }
    websiteToAdd.value = { title: 'Example', site: 'https://example.com' }
    Object.assign(locationToAdd.value, { latitude: 1, longitude: 2, placeholder: 'Somewhere' })
    tweetToAdd.value.url = 'https://x.com/lin_snow/status/1234567890'
    videoURL.value = 'https://www.bilibili.com/video/BV1xx411c7mD'
    await settle()
    expect(extensionPreview.value).not.toBeNull()

    removeExtension()
    await nextTick()

    expect(extensionPreview.value).toBeNull()
    expect(payload.value).toBeNull()
    expect(extensionToAdd.value).toEqual({ extension: '', extension_type: '' })
    expect(websiteToAdd.value).toEqual({ title: '', site: '' })
    expect(locationToAdd.value).toEqual({ latitude: null, longitude: null, placeholder: '' })
    expect(tweetToAdd.value).toEqual({ url: '', username: '', statusId: '' })
    expect(videoURL.value).toBe('')
  })

  it('keeps the submit payload equal to the previewed payload', async () => {
    const { extensionToAdd, locationToAdd, extensionPreview, payload } = setup()

    extensionToAdd.value.extension_type = ExtensionType.LOCATION
    Object.assign(locationToAdd.value, {
      latitude: 31.2,
      longitude: 121.4,
      placeholder: 'Shanghai',
    })
    await settle()

    expect(payload.value).toEqual({
      type: ExtensionType.LOCATION,
      payload: { latitude: 31.2, longitude: 121.4, placeholder: 'Shanghai' },
    })
    expect(payload.value).toEqual(extensionPreview.value)
  })
})
