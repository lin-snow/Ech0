// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Fragment, type ReactNode } from 'react'
import type { Echo } from '../fixtures/mock'
import { Icon } from './Icon'
import { Tag } from './Tag'

type EchoEntryProps = {
  echo: Echo
  clamp?: number
  showActions?: boolean
}

function inline(text: string, keyPrefix: string): ReactNode[] {
  return text.split(/(https?:\/\/[^\s]+|~~[^~]+~~)/g).map((part, index) => {
    const key = `${keyPrefix}-${index}`
    if (/^https?:\/\//.test(part)) {
      return (
        <a key={key} href={part} target="_blank" rel="noreferrer">
          {part}
        </a>
      )
    }
    if (part.startsWith('~~') && part.endsWith('~~')) {
      return <s key={key}>{part.slice(2, -2)}</s>
    }
    return <Fragment key={key}>{part}</Fragment>
  })
}

function EchoBody({ content, clamp }: { content: string; clamp?: number }) {
  const blocks = content.split(/\n{2,}/)
  return (
    <div
      className="entry__body"
      style={
        clamp
          ? {
              display: '-webkit-box',
              WebkitBoxOrient: 'vertical',
              WebkitLineClamp: clamp,
              overflow: 'hidden',
            }
          : undefined
      }
    >
      {blocks.map((block, index) => {
        if (block.startsWith('```')) {
          const code = block.replace(/^```[a-zA-Z]*\n?/, '').replace(/```$/, '')
          return (
            <pre key={`code-${index}`}>
              <code>{code}</code>
            </pre>
          )
        }
        return (
          <p key={`p-${index}`}>
            {block.split('\n').map((line, lineIndex) => (
              <Fragment key={`line-${lineIndex}`}>
                {lineIndex > 0 ? <br /> : null}
                {inline(line, `line-${index}-${lineIndex}`)}
              </Fragment>
            ))}
          </p>
        )
      })}
    </div>
  )
}

export function EchoEntry({ echo, clamp, showActions = true }: EchoEntryProps) {
  const image = echo.images?.[0]

  return (
    <article className="entry">
      <div className="entry__head">
        <span className="entry__dot" aria-hidden="true" />
        <span className="entry__stamp">{echo.time}</span>
        {echo.isPrivate ? (
          <span className="entry__lock">
            <Icon name="lock" size={11} />
            PRIVATE
          </span>
        ) : null}
      </div>

      {echo.content ? <EchoBody content={echo.content} clamp={clamp} /> : null}

      {image ? (
        <div className="entry__media">
          <img
            src={image.src}
            width={image.width}
            height={image.height}
            alt="Echo 附件"
            loading="lazy"
          />
        </div>
      ) : null}

      {echo.tags.length > 0 ? (
        <div className="entry__tags">
          {echo.tags.map((tag) => (
            <Tag key={tag} label={tag} />
          ))}
        </div>
      ) : null}

      {showActions ? (
        <div className="entry__actions">
          <button type="button" className="entry__action" aria-label="喜欢">
            <Icon name="heart" size={14} />
            <span>12</span>
          </button>
          <button type="button" className="entry__action" aria-label="评论">
            <Icon name="comment" size={14} />
            <span>3</span>
          </button>
          <button type="button" className="entry__action" aria-label="更多操作">
            <Icon name="more" size={14} />
          </button>
        </div>
      ) : null}
    </article>
  )
}
