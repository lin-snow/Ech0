// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const codeFencePattern = /```[\s\S]*?```/g
const htmlTagPattern = /<\/?[a-zA-Z][^>]*>/g
const imagePattern = /!\[[^\]]*\]\([^)]*\)/g
const linkPattern = /\[([^\]]*)\]\([^)]*\)/g
const listPrefixPattern = /^[ \t]*(?:[-*+]|\d+\.)[ \t]+/
const spaceRunPattern = /[ \t]+/g
const punctuationPattern = /[ \t]+([.,!?;:，。！？；：])/g
const markdownNoisePattern = /[*_`~#>|\\]/g

/** Mirrors the server-side strip so the hydrated head matches the crawlable one. */
export const seoPlainText = (raw?: string | null): string => {
  if (!raw) return ''

  const text = raw
    .replace(codeFencePattern, ' ')
    .replace(imagePattern, ' ')
    .replace(linkPattern, '$1')
    .replace(htmlTagPattern, ' ')
    .replace(/\r\n/g, '\n')

  return text
    .split('\n')
    .map((line) =>
      line
        .replace(listPrefixPattern, '')
        .replace(markdownNoisePattern, '')
        .replace(spaceRunPattern, ' ')
        .replace(punctuationPattern, '$1')
        .trim(),
    )
    .filter(Boolean)
    .join('\n')
}

export const seoTruncate = (text: string, limit: number): string => {
  const runes = Array.from(text)
  if (runes.length <= limit) return text
  return `${runes
    .slice(0, limit - 1)
    .join('')
    .trim()}…`
}

export const absoluteSeoUrl = (raw?: string | null): string | undefined => {
  const value = String(raw || '').trim()
  if (!value) return undefined

  try {
    return new URL(value, window.location.origin).toString()
  } catch {
    return undefined
  }
}
