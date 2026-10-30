// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Input } from '@base-ui/react/input'
import type { ReactNode } from 'react'
import { Icon } from '../lib/icons'
import type { SketchFeed } from '../lib/useSketchFeed'
import styles from './feed.module.css'
import { Kbd } from './ui'

function SearchField({
  value,
  onChange,
}: {
  value: string
  onChange: (value: string) => void
}) {
  return (
    <div className={styles.search}>
      <Icon name="search" size={13} className={styles.searchIcon} />
      <Input
        data-search-input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder="搜索"
        className={styles.searchInput}
        aria-label="搜索"
      />
      {value ? (
        <button
          type="button"
          className={styles.searchClear}
          aria-label="清除搜索"
          onClick={() => onChange('')}
        >
          <Icon name="close" size={11} />
        </button>
      ) : (
        <Kbd>⌘K</Kbd>
      )}
    </div>
  )
}

export function FeedToolbar({
  feed,
  variant = 'inline',
  leading,
  trailing,
}: {
  feed: SketchFeed
  variant?: 'inline' | 'rail'
  leading?: ReactNode
  trailing?: ReactNode
}) {
  const rail = variant === 'rail'
  const chip = feed.tag ? (
    <button
      type="button"
      className={styles.tagFilter}
      onClick={() => feed.setTag(null)}
      aria-label={`清除标签 ${feed.tag}`}
    >
      #{feed.tag}
      <Icon name="close" size={9} />
    </button>
  ) : null

  if (rail) {
    return (
      <div className={styles.toolbarRail}>
        <div className={styles.toolbarRailTop}>
          {leading}
          <SearchField value={feed.query} onChange={feed.setQuery} />
          {chip}
          {trailing && <div className={styles.toolbarTrailing}>{trailing}</div>}
        </div>
      </div>
    )
  }

  return (
    <div className={styles.toolbar}>
      <SearchField value={feed.query} onChange={feed.setQuery} />
      {chip}
    </div>
  )
}
