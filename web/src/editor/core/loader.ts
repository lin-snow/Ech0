// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type * as MarkdownCore from './markdown'

export type MarkdownModule = typeof MarkdownCore

let modulePromise: Promise<MarkdownModule> | null = null
let loaded: MarkdownModule | null = null

/**
 * Loads the markdown core once. The app boot awaits this so the first
 * MarkdownRenderer mount can paint final HTML synchronously instead of
 * swapping raw text for rendered HTML after the paint (which reflowed every
 * row of the virtual timeline).
 */
export function preloadMarkdown(): Promise<MarkdownModule> {
  if (loaded) return Promise.resolve(loaded)
  if (!modulePromise) {
    modulePromise = import('./markdown').then((module) => {
      loaded = module
      return module
    })
  }
  return modulePromise
}

export function loadedMarkdown(): MarkdownModule | null {
  return loaded
}
