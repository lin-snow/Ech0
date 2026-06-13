// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type MusicTrack = {
  name: string
  artist: string
  url: string
  cover: string
  lrc: string
}

export type LyricWord = {
  time: number
  text: string
}

export type LyricLine = {
  time: number
  text: string
  words: LyricWord[]
}

const lyricTimestampPattern = /\[(\d{1,3}):(\d{2})(?:[.:](\d{1,3}))?\]/g
const wordTimestampPattern = /<(\d{1,3}):(\d{2})(?:[.:](\d{1,3}))?>/g
const lyricCreditPattern = /^(?:作词|作曲|编曲|制作人|混音|母带|词|曲)\s*[:：]/
const instrumentalLyricPattern = /^纯音乐[，,、\s]*请欣赏[。！!]?$/

const getString = (value: unknown) => (typeof value === 'string' ? value.trim() : '')

export function normalizeMetingTrack(payload: unknown): MusicTrack | null {
  const first = Array.isArray(payload) ? payload[0] : payload
  if (!first || typeof first !== 'object') return null

  const data = first as Record<string, unknown>
  const url = getString(data.url)
  if (!url) return null

  return {
    name: getString(data.name) || getString(data.title),
    artist: getString(data.artist) || getString(data.author),
    url,
    cover: getString(data.cover) || getString(data.pic),
    lrc: getString(data.lrc) || getString(data.lyric),
  }
}

function parseTimestamp(match: RegExpMatchArray) {
  const minutes = Number(match[1])
  const seconds = Number(match[2])
  const fraction = Number(`0.${match[3] || 0}`)
  return minutes * 60 + seconds + fraction
}

function parseWords(text: string) {
  const matches = Array.from(text.matchAll(wordTimestampPattern))
  if (!matches.length) return []

  return matches.flatMap((match, index) => {
    const start = match.index! + match[0].length
    const end = matches[index + 1]?.index ?? text.length
    const segment = text.slice(start, end)
    const normalized = segment.trim()
    const word = normalized && /\s$/.test(segment) ? `${normalized} ` : normalized
    return word ? [{ time: parseTimestamp(match), text: word }] : []
  })
}

export function parseLyrics(source: string): LyricLine[] {
  const parsed: LyricLine[] = []
  const linesByTimestamp = new Map<number, LyricLine>()

  for (const rawLine of source.split(/\r?\n/)) {
    const lineMatches = Array.from(rawLine.matchAll(lyricTimestampPattern))
    if (!lineMatches.length) continue

    const textWithWordTimestamps = rawLine.replace(lyricTimestampPattern, '').trim()
    const words = parseWords(textWithWordTimestamps)
    const text = textWithWordTimestamps.replace(wordTimestampPattern, '').trim()
    if (!text || lyricCreditPattern.test(text) || instrumentalLyricPattern.test(text)) continue

    for (const match of lineMatches) {
      const time = parseTimestamp(match)
      const timestampKey = Math.round(time * 1000)
      const existing = linesByTimestamp.get(timestampKey)
      if (existing) {
        const existingTexts = existing.text.split('\n')
        if (!existing.words.length && words.length) {
          existing.words = words
          existing.text = [text, ...existingTexts.filter((item) => item !== text)].join('\n')
        } else if (!existingTexts.includes(text)) {
          existing.text += `\n${text}`
        }
        continue
      }

      const line = { time, text, words }
      linesByTimestamp.set(timestampKey, line)
      parsed.push(line)
    }
  }

  return parsed.sort((a, b) => a.time - b.time)
}
