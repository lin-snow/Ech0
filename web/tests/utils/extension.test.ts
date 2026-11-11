// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import { ExtensionType } from '@/enums/enums'
import {
  buildEchoExtension,
  isExtensionPreviewReady,
  type EchoExtensionDraft,
} from '@/utils/extension'

const draft = (overrides: Partial<EchoExtensionDraft> = {}): EchoExtensionDraft => ({
  extensionType: '',
  extension: '',
  website: { title: '', site: '' },
  location: { latitude: null, longitude: null, placeholder: '' },
  tweetUrl: '',
  ...overrides,
})

describe('buildEchoExtension', () => {
  it('builds the music, video and GitHub payloads from the raw value', () => {
    expect(
      buildEchoExtension(
        draft({
          extensionType: ExtensionType.MUSIC,
          extension: 'https://music.163.com/#/song?id=123',
        }),
      ),
    ).toEqual({
      type: ExtensionType.MUSIC,
      payload: { url: 'https://music.163.com/#/song?id=123' },
    })

    expect(
      buildEchoExtension(draft({ extensionType: ExtensionType.VIDEO, extension: 'BV1xx411c7mD' })),
    ).toEqual({ type: ExtensionType.VIDEO, payload: { videoId: 'BV1xx411c7mD' } })

    expect(
      buildEchoExtension(
        draft({
          extensionType: ExtensionType.GITHUBPROJ,
          extension: 'https://github.com/l-s/Ech0',
        }),
      ),
    ).toEqual({
      type: ExtensionType.GITHUBPROJ,
      payload: { repoUrl: 'https://github.com/l-s/Ech0' },
    })
  })

  it('requires both website fields and valid location data', () => {
    expect(
      buildEchoExtension(
        draft({ extensionType: ExtensionType.WEBSITE, website: { title: '', site: '' } }),
      ),
    ).toBeNull()
    expect(
      buildEchoExtension(
        draft({
          extensionType: ExtensionType.WEBSITE,
          website: { title: 'Example', site: 'https://example.com' },
        }),
      ),
    ).toEqual({
      type: ExtensionType.WEBSITE,
      payload: { title: 'Example', site: 'https://example.com' },
    })

    expect(
      buildEchoExtension(
        draft({
          extensionType: ExtensionType.LOCATION,
          location: { latitude: 91, longitude: 0, placeholder: 'Somewhere' },
        }),
      ),
    ).toBeNull()
    expect(
      buildEchoExtension(
        draft({
          extensionType: ExtensionType.LOCATION,
          location: { latitude: 31.2, longitude: 121.4, placeholder: '  Shanghai  ' },
        }),
      ),
    ).toEqual({
      type: ExtensionType.LOCATION,
      payload: { latitude: 31.2, longitude: 121.4, placeholder: 'Shanghai' },
    })
  })

  it('parses the tweet URL instead of relying on submit-time fields', () => {
    expect(
      buildEchoExtension(
        draft({
          extensionType: ExtensionType.TWEET,
          tweetUrl: 'https://twitter.com/lin_snow/status/1234567890',
        }),
      ),
    ).toEqual({
      type: ExtensionType.TWEET,
      payload: {
        url: 'https://x.com/lin_snow/status/1234567890',
        username: 'lin_snow',
        statusId: '1234567890',
      },
    })
    expect(
      buildEchoExtension(
        draft({ extensionType: ExtensionType.TWEET, tweetUrl: 'https://example.com/status/1' }),
      ),
    ).toBeNull()
  })
})

describe('isExtensionPreviewReady', () => {
  it('hides previews that cannot render a remote card', () => {
    expect(
      isExtensionPreviewReady({ type: ExtensionType.MUSIC, payload: { url: 'not-a-url' } }),
    ).toBe(false)
    expect(
      isExtensionPreviewReady({
        type: ExtensionType.MUSIC,
        payload: { url: 'https://music.163.com/#/song?id=1' },
      }),
    ).toBe(true)
    expect(
      isExtensionPreviewReady({
        type: ExtensionType.GITHUBPROJ,
        payload: { repoUrl: 'https://github.com/only-owner' },
      }),
    ).toBe(false)
    expect(
      isExtensionPreviewReady({
        type: ExtensionType.GITHUBPROJ,
        payload: { repoUrl: 'https://github.com/owner/repo' },
      }),
    ).toBe(true)
    expect(
      isExtensionPreviewReady({
        type: ExtensionType.WEBSITE,
        payload: { title: 'x', site: 'ftp://example.com' },
      }),
    ).toBe(false)
    expect(
      isExtensionPreviewReady({
        type: ExtensionType.WEBSITE,
        payload: { title: 'x', site: 'https://example.com' },
      }),
    ).toBe(true)
  })
})
