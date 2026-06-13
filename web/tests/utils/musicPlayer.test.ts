// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'

import { normalizeMetingTrack, parseLyrics } from '../../src/utils/musicPlayer'

describe('normalizeMetingTrack', () => {
  it('reads the first track from an array payload', () => {
    expect(
      normalizeMetingTrack([
        {
          url: 'https://cdn/song.mp3',
          name: 'Song',
          artist: 'Artist',
          cover: 'c.jpg',
          lrc: 'l.lrc',
        },
        { url: 'https://cdn/other.mp3', name: 'Other' },
      ]),
    ).toEqual({
      name: 'Song',
      artist: 'Artist',
      url: 'https://cdn/song.mp3',
      cover: 'c.jpg',
      lrc: 'l.lrc',
    })
  })

  it('accepts a single object payload', () => {
    expect(normalizeMetingTrack({ url: 'https://cdn/song.mp3', name: 'Song' })).toEqual({
      name: 'Song',
      artist: '',
      url: 'https://cdn/song.mp3',
      cover: '',
      lrc: '',
    })
  })

  it('falls back to alternate field names', () => {
    expect(
      normalizeMetingTrack([{ url: 'u', title: 'T', author: 'Au', pic: 'P', lyric: 'L' }]),
    ).toEqual({ name: 'T', artist: 'Au', url: 'u', cover: 'P', lrc: 'L' })
  })

  it('prefers primary field names over fallbacks', () => {
    expect(
      normalizeMetingTrack([
        {
          url: 'u',
          name: 'N',
          title: 'T',
          artist: 'Ar',
          author: 'Au',
          cover: 'C',
          pic: 'P',
          lrc: 'Lr',
          lyric: 'Ly',
        },
      ]),
    ).toMatchObject({ name: 'N', artist: 'Ar', cover: 'C', lrc: 'Lr' })
  })

  it('trims string fields', () => {
    expect(normalizeMetingTrack([{ url: '  u  ', name: '  N  ' }])).toMatchObject({
      url: 'u',
      name: 'N',
    })
  })

  it('ignores non-string field values', () => {
    expect(normalizeMetingTrack([{ url: 'u', name: 123, cover: null }])).toMatchObject({
      name: '',
      cover: '',
    })
  })

  it('returns null without a playable url', () => {
    expect(normalizeMetingTrack([{ name: 'N' }])).toBeNull()
    expect(normalizeMetingTrack([{ url: '   ' }])).toBeNull()
    expect(normalizeMetingTrack({})).toBeNull()
  })

  it('returns null for empty or non-object payloads', () => {
    expect(normalizeMetingTrack([])).toBeNull()
    expect(normalizeMetingTrack(null)).toBeNull()
    expect(normalizeMetingTrack(undefined)).toBeNull()
    expect(normalizeMetingTrack('nope')).toBeNull()
    expect(normalizeMetingTrack(42)).toBeNull()
  })
})

describe('parseLyrics', () => {
  it('parses basic timestamped lines', () => {
    const lines = parseLyrics('[00:00.00]First line\n[00:05.00]Second line')
    expect(lines).toHaveLength(2)
    expect(lines[0]).toEqual({ time: 0, text: 'First line', words: [] })
    expect(lines[1]).toEqual({ time: 5, text: 'Second line', words: [] })
  })

  it('sorts lines by timestamp', () => {
    const lines = parseLyrics('[00:10.00]B\n[00:05.00]A')
    expect(lines.map((line) => line.text)).toEqual(['A', 'B'])
  })

  it('expands a line carrying multiple timestamps', () => {
    const lines = parseLyrics('[00:01.00][00:05.00]Chorus')
    expect(lines).toHaveLength(2)
    expect(lines.map((line) => line.time)).toEqual([1, 5])
    expect(lines.every((line) => line.text === 'Chorus')).toBe(true)
  })

  it('ignores ID tags and lines without timestamps', () => {
    expect(parseLyrics('[ti:Title]\n[ar:Artist]\nplain text\n')).toEqual([])
  })

  it('parses minutes, seconds and fractional seconds', () => {
    const [line] = parseLyrics('[01:23.45]Test')
    expect(line?.time).toBeCloseTo(83.45, 5)
    expect(line?.text).toBe('Test')
  })

  it('handles CRLF line endings', () => {
    expect(parseLyrics('[00:00.00]A\r\n[00:01.00]B')).toHaveLength(2)
  })

  it('drops empty source', () => {
    expect(parseLyrics('')).toEqual([])
  })

  it('filters credit and instrumental lines', () => {
    expect(
      parseLyrics('[00:00.00]作词：张三\n[00:01.00]作曲 : 李四\n[00:02.00]编曲：王五'),
    ).toEqual([])
    expect(parseLyrics('[00:00.00]纯音乐，请欣赏')).toEqual([])
  })

  it('keeps lyrics that merely contain credit keywords mid-line', () => {
    const [line] = parseLyrics('[00:00.00]我想为你作词')
    expect(line?.text).toBe('我想为你作词')
  })

  it('merges a translation sharing the same timestamp', () => {
    const lines = parseLyrics('[00:01.00]Hello\n[00:01.00]你好')
    expect(lines).toHaveLength(1)
    expect(lines[0]).toEqual({ time: 1, text: 'Hello\n你好', words: [] })
  })

  it('deduplicates identical text at the same timestamp', () => {
    const lines = parseLyrics('[00:01.00]Hello\n[00:01.00]Hello')
    expect(lines).toHaveLength(1)
    expect(lines[0]?.text).toBe('Hello')
  })

  it('parses word-level (karaoke) timestamps', () => {
    const [line] = parseLyrics('[00:01.00]<00:01.00>Hello <00:01.50>world')
    expect(line?.text).toBe('Hello world')
    expect(line?.words).toEqual([
      { time: 1, text: 'Hello ' },
      { time: 1.5, text: 'world' },
    ])
  })

  it('upgrades a wordless line when a karaoke variant shares its timestamp', () => {
    const [line] = parseLyrics('[00:01.00]Hello world\n[00:01.00]<00:01.00>Hello <00:01.50>world')
    expect(line?.text).toBe('Hello world')
    expect(line?.words).toHaveLength(2)
  })

  it('appends a wordless translation onto a karaoke line', () => {
    const [line] = parseLyrics('[00:01.00]<00:01.00>Hello <00:01.50>world\n[00:01.00]你好世界')
    expect(line?.text).toBe('Hello world\n你好世界')
    expect(line?.words).toHaveLength(2)
  })
})
