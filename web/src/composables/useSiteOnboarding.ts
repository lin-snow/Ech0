// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { DEFAULT_SITE_TITLE, PLACEHOLDER_SITE_URL_HOST } from '@/constants/site'

/** Storage key for the one-time dismissal of the first-echo card. */
export const FIRST_ECHO_CARD_SEEN_KEY = 'onboarding.v2:firstEchoSeen'

/** Storage key for the owner dismissing the setup checklist by hand. */
export const CHECKLIST_DISMISSED_KEY = 'onboarding.v2:checklistDismissed'

export type SiteOnboardingStepId = 'publish' | 'siteTitle' | 'siteUrl'

export type SiteOnboardingStep = {
  id: SiteOnboardingStepId
  done: boolean
}

export type SiteOnboardingInput = {
  echoTotal: number
  siteTitle: string
  serverURL: string
}

/** True while the operator has not given the instance its own name. */
export function isDefaultSiteTitle(title: string | undefined): boolean {
  const trimmed = (title ?? '').trim()
  return trimmed === '' || trimmed === DEFAULT_SITE_TITLE
}

/** True while `server_url` is empty or still the shipped placeholder. */
export function isPlaceholderSiteURL(raw: string | undefined): boolean {
  const trimmed = (raw ?? '').trim()
  if (trimmed === '') return true
  try {
    return new URL(trimmed).hostname.toLowerCase() === PLACEHOLDER_SITE_URL_HOST
  } catch {
    return true
  }
}

/**
 * Derives the setup checklist from real instance state. Nothing is persisted:
 * completing a step here means the live data changed, so the list can never
 * drift out of sync with the site.
 */
export function resolveSiteOnboardingSteps(input: SiteOnboardingInput): SiteOnboardingStep[] {
  return [
    { id: 'publish', done: input.echoTotal > 0 },
    { id: 'siteTitle', done: !isDefaultSiteTitle(input.siteTitle) },
    { id: 'siteUrl', done: !isPlaceholderSiteURL(input.serverURL) },
  ]
}

/**
 * The checklist belongs to the signed-in owner, and stays until the work is
 * actually finished: publishing the first echo ticks one row, it does not hide
 * the remaining ones. It disappears only when every step is done, or when the
 * owner dismisses it by hand.
 */
export function shouldShowSiteChecklist(input: {
  isOwner: boolean
  loading: boolean
  filtering: boolean
  dismissed: boolean
  echoTotal: number
  siteTitle: string
  serverURL: string
}): boolean {
  if (!input.isOwner || input.loading || input.filtering || input.dismissed) return false
  return resolveSiteOnboardingSteps(input).some((step) => !step.done)
}

/**
 * The first published echo is the moment the site becomes shareable. Show the
 * card once per browser; a private first echo is not shareable, so skip it.
 * A filtered view can narrow `echoTotal` to 1, so never show it there.
 */
export function shouldShowFirstEchoCard(input: {
  isOwner: boolean
  filtering: boolean
  echoTotal: number
  firstEchoPrivate: boolean
  dismissed: boolean
}): boolean {
  return (
    input.isOwner &&
    !input.filtering &&
    input.echoTotal === 1 &&
    !input.firstEchoPrivate &&
    !input.dismissed
  )
}
