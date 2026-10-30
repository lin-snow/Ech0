// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useState } from 'react'
import type { SketchFeed } from './useSketchFeed'

export type ComposerState = {
  body: string
  setBody: (value: string) => void
  tags: string[]
  setTags: (value: string[]) => void
  isPrivate: boolean
  setIsPrivate: (value: boolean) => void
  canPublish: boolean
  publish: () => void
}

/** Draft state lives above the placement so moving the editor between the
 * left rail and the feed (responsive) never drops what was typed. */
export function useComposer(feed: SketchFeed): ComposerState {
  const [body, setBody] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [isPrivate, setIsPrivate] = useState(false)

  const publish = () => {
    if (!body.trim()) return
    feed.publish({ body, tags: tags.join(' '), isPrivate })
    setBody('')
    setTags([])
    setIsPrivate(false)
  }

  return {
    body,
    setBody,
    tags,
    setTags,
    isPrivate,
    setIsPrivate,
    canPublish: body.trim().length > 0,
    publish,
  }
}
