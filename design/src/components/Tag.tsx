// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { ReactNode } from 'react'
import { cx } from '../lib/cx'

type TagProps = {
  label: string
  count?: number
  active?: boolean
  onClick?: () => void
  className?: string
}

export function Tag({ label, count, active = false, onClick, className }: TagProps) {
  const content: ReactNode = (
    <>
      <span className="tag__hash">#</span>
      <span>{label}</span>
      {count !== undefined ? <span className="tag__count">{count}</span> : null}
    </>
  )

  if (!onClick) {
    return (
      <span className={cx('tag', active && 'tag--active', className)} data-static="true">
        {content}
      </span>
    )
  }

  return (
    <button
      type="button"
      className={cx('tag', className)}
      aria-pressed={active}
      onClick={onClick}
    >
      {content}
    </button>
  )
}
