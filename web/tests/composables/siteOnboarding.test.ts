// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import {
  isDefaultSiteTitle,
  isPlaceholderSiteURL,
  resolveSiteOnboardingSteps,
  shouldShowFirstEchoCard,
  shouldShowSiteChecklist,
} from '@/composables/useSiteOnboarding'

describe('isDefaultSiteTitle', () => {
  it('treats empty and the shipped default as unnamed', () => {
    expect(isDefaultSiteTitle('')).toBe(true)
    expect(isDefaultSiteTitle('   ')).toBe(true)
    expect(isDefaultSiteTitle(undefined)).toBe(true)
    expect(isDefaultSiteTitle('Ech0')).toBe(true)
  })

  it('accepts a custom title', () => {
    expect(isDefaultSiteTitle('晨间笔记')).toBe(false)
    expect(isDefaultSiteTitle(' my notes ')).toBe(false)
  })
})

describe('isPlaceholderSiteURL', () => {
  it('flags empty, unparsable and placeholder URLs', () => {
    expect(isPlaceholderSiteURL('')).toBe(true)
    expect(isPlaceholderSiteURL(undefined)).toBe(true)
    expect(isPlaceholderSiteURL('not a url')).toBe(true)
    expect(isPlaceholderSiteURL('https://ech0.example.com')).toBe(true)
    expect(isPlaceholderSiteURL('https://ECH0.example.com/')).toBe(true)
  })

  it('accepts a configured URL even when it differs from the current origin', () => {
    expect(isPlaceholderSiteURL('https://blog.example.org')).toBe(false)
    expect(isPlaceholderSiteURL('http://localhost:6277')).toBe(false)
  })
})

describe('resolveSiteOnboardingSteps', () => {
  it('marks every step pending on a fresh instance', () => {
    expect(resolveSiteOnboardingSteps({ echoTotal: 0, siteTitle: 'Ech0', serverURL: '' })).toEqual([
      { id: 'publish', done: false },
      { id: 'siteTitle', done: false },
      { id: 'siteUrl', done: false },
    ])
  })

  it('marks steps done from live state', () => {
    expect(
      resolveSiteOnboardingSteps({ echoTotal: 3, siteTitle: '我的站点', serverURL: 'https://a.b' }),
    ).toEqual([
      { id: 'publish', done: true },
      { id: 'siteTitle', done: true },
      { id: 'siteUrl', done: true },
    ])
  })
})

describe('shouldShowSiteChecklist', () => {
  const base = {
    isOwner: true,
    loading: false,
    filtering: false,
    dismissed: false,
    echoTotal: 0,
    siteTitle: 'Ech0',
    serverURL: '',
  }

  it('shows for the signed-in owner of a fresh instance', () => {
    expect(shouldShowSiteChecklist(base)).toBe(true)
  })

  it('stays visible after the first echo while other steps are pending', () => {
    expect(shouldShowSiteChecklist({ ...base, echoTotal: 1 })).toBe(true)
    expect(shouldShowSiteChecklist({ ...base, echoTotal: 3, siteTitle: '我的站点' })).toBe(true)
  })

  it('hides only when every step is done', () => {
    expect(
      shouldShowSiteChecklist({
        ...base,
        echoTotal: 1,
        siteTitle: '我的站点',
        serverURL: 'https://blog.example.org',
      }),
    ).toBe(false)
  })

  it('hides when the owner dismisses it', () => {
    expect(shouldShowSiteChecklist({ ...base, dismissed: true })).toBe(false)
  })

  it('hides for guests, loading and filtered views', () => {
    expect(shouldShowSiteChecklist({ ...base, isOwner: false })).toBe(false)
    expect(shouldShowSiteChecklist({ ...base, loading: true })).toBe(false)
    expect(shouldShowSiteChecklist({ ...base, filtering: true })).toBe(false)
  })
})

describe('shouldShowFirstEchoCard', () => {
  const base = { isOwner: true, echoTotal: 1, firstEchoPrivate: false, dismissed: false }

  it('shows only for the single public first echo, once', () => {
    expect(shouldShowFirstEchoCard(base)).toBe(true)
  })

  it('hides for guests, private echoes, multiple echoes and after dismissal', () => {
    expect(shouldShowFirstEchoCard({ ...base, isOwner: false })).toBe(false)
    expect(shouldShowFirstEchoCard({ ...base, firstEchoPrivate: true })).toBe(false)
    expect(shouldShowFirstEchoCard({ ...base, dismissed: true })).toBe(false)
    expect(shouldShowFirstEchoCard({ ...base, echoTotal: 2 })).toBe(false)
    expect(shouldShowFirstEchoCard({ ...base, echoTotal: 0 })).toBe(false)
  })
})
