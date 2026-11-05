// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { shallowRef } from 'vue'

export type SeoOverride = {
  /** Route path the override belongs to; stale overrides are ignored. */
  path: string
  title: string
  description: string
  type: 'article' | 'website'
  image?: string
  publishedAt?: string
  jsonLd?: Record<string, unknown>
}

const override = shallowRef<SeoOverride | null>(null)

export const seoOverride = override

export const setSeoOverride = (next: SeoOverride) => {
  override.value = next
}

export const clearSeoOverride = (path?: string) => {
  if (!path || override.value?.path === path) {
    override.value = null
  }
}
