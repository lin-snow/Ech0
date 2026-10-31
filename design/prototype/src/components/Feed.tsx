// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Menu } from '@base-ui/react/menu'
import { useMemo, type ReactNode } from 'react'
import type { Echo } from '../fixtures/mock'
import type { ComposerState } from '../lib/useComposer'
import { Icon } from '../lib/icons'
import { renderMarkdown } from '../lib/markdown'
import type { SketchFeed } from '../lib/useSketchFeed'
import { Composer } from './Composer'
import { FeedToolbar } from './FeedToolbar'
import styles from './feed.module.css'
import uiStyles from './ui.module.css'
import { Button, Mono } from './ui'

type Variant = 'columns' | 'focus'

function EntryMenu() {
  return (
    <Menu.Root>
      <Menu.Trigger
        render={
          <button type="button" className={uiStyles.iconButton} aria-label="更多操作">
            <Icon name="more" size={14} />
          </button>
        }
      />
      <Menu.Portal>
        <Menu.Positioner side="bottom" align="end" sideOffset={4}>
          <Menu.Popup className={styles.menuPopup}>
            <Menu.Item className={styles.menuItem}>编辑</Menu.Item>
            <Menu.Item className={styles.menuItem}>删除</Menu.Item>
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  )
}

function EchoImages({ images }: { images: NonNullable<Echo['images']> }) {
  return (
    <div className={styles.images} data-count={images.length}>
      {images.map((image, index) => (
        <a
          key={`${image.src}-${index}`}
          className={styles.imageButton}
          href={image.src}
          target="_blank"
          rel="noopener noreferrer"
          aria-label={`查看图片 ${index + 1}`}
          style={{ aspectRatio: `${image.width} / ${image.height}` }}
        >
          <img
            src={image.src}
            alt=""
            width={image.width}
            height={image.height}
            loading={index === 0 ? 'eager' : 'lazy'}
          />
        </a>
      ))}
    </div>
  )
}

function TimelineEntry({ echo }: { echo: Echo }) {
  const html = useMemo(() => renderMarkdown(echo.content), [echo.content])

  return (
    <article className={styles.entry}>
      <header className={styles.entryHead}>
        <span className={styles.marker} aria-hidden="true">
          <span className={styles.dot} />
        </span>
        <time className={styles.date} dateTime={echo.stamp}>
          {echo.time}
        </time>
        {echo.isPrivate && (
          <span className={styles.privateFlag} title="私密">
            <Icon name="lock" size={10} />
          </span>
        )}
        <span className={styles.grow} />
        <EntryMenu />
      </header>

      <div className={styles.content}>
        {echo.content.trim() && (
          <div className={styles.markdown} dangerouslySetInnerHTML={{ __html: html }} />
        )}
        {echo.images && echo.images.length > 0 && <EchoImages images={echo.images} />}
      </div>
    </article>
  )
}

export type FeedProps = {
  feed: SketchFeed
  variant: Variant
  heading?: ReactNode
  composer?: ComposerState
  showToolbar?: boolean
}

export function Feed({ feed, variant, heading, composer, showToolbar = true }: FeedProps) {
  const hasFilters = Boolean(feed.query || feed.tag)
  return (
    <div className={styles.feed} data-variant={variant}>
      {showToolbar && (
        <div className={styles.stickyBlock}>
          <FeedToolbar feed={feed} />
        </div>
      )}

      <div className={styles.scrollBody}>
        {heading}
        {composer && <Composer composer={composer} />}
        {feed.filtered.length === 0 ? (
          <div className={styles.empty}>
            <Mono>NO ECHO</Mono>
            {hasFilters && (
              <Button
                size="sm"
                onClick={() => {
                  feed.setQuery('')
                  feed.setTag(null)
                }}
              >
                重置
              </Button>
            )}
          </div>
        ) : (
          feed.filtered.map((echo) => <TimelineEntry key={echo.id} echo={echo} />)
        )}
      </div>
    </div>
  )
}
