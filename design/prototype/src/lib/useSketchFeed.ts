// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useCallback, useMemo, useState } from 'react'
import { echos as seedEchos, type Echo } from '../fixtures/mock'

export type ComposerDraft = {
  body: string
  tags: string
  isPrivate: boolean
}

export function useSketchFeed() {
  const [query, setQuery] = useState('')
  const [tag, setTag] = useState<string | null>(null)
  const [extra, setExtra] = useState<Echo[]>([])

  const all = useMemo(() => [...extra, ...seedEchos], [extra])

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase()
    return all.filter((echo) => {
      if (tag && !echo.tags.includes(tag)) return false
      if (!needle) return true
      const haystack = [echo.content, echo.tags.join(' ')].join(' ').toLowerCase()
      return haystack.includes(needle)
    })
  }, [all, tag, query])

  const publish = useCallback((draft: ComposerDraft) => {
    const body = draft.body.trim()
    if (!body) return
    const tags = draft.tags
      .split(/[,，\s]+/)
      .map((t) => t.replace(/^#/, '').trim())
      .filter(Boolean)
    const now = new Date()
    const stamp = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(
      now.getDate(),
    ).padStart(2, '0')}T${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`
    setExtra((prev) => [
      {
        id: `draft-${now.getTime()}`,
        time: '刚刚',
        stamp,
        content: body,
        tags,
        isPrivate: draft.isPrivate,
      },
      ...prev,
    ])
    setQuery('')
    setTag(null)
  }, [])

  const selectTag = useCallback((next: string) => {
    setTag((current) => (current === next ? null : next))
  }, [])

  return {
    query,
    setQuery,
    tag,
    setTag,
    selectTag,
    filtered,
    publish,
    total: all.length,
  }
}

export type SketchFeed = ReturnType<typeof useSketchFeed>
