// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useState } from 'react'
import { Button } from './Button'
import { Icon } from './Icon'
import { IconButton } from './IconButton'

type ComposerProps = {
  value?: string
  onChange?: (value: string) => void
  placeholder?: string
  compact?: boolean
  hint?: string
  onPublish?: () => void
}

export function Composer({
  value,
  onChange,
  placeholder = '写点什么…',
  compact = false,
  hint,
  onPublish,
}: ComposerProps) {
  const [inner, setInner] = useState('')
  const current = value ?? inner
  const canPublish = current.trim().length > 0

  function update(next: string) {
    if (value === undefined) setInner(next)
    onChange?.(next)
  }

  return (
    <div className="composer">
      <textarea
        className="composer__body"
        value={current}
        onChange={(event) => update(event.target.value)}
        placeholder={placeholder}
        rows={compact ? 2 : 3}
        aria-label="Echo 内容"
      />
      <div className="composer__footer">
        <div className="composer__tools">
          <IconButton label="添加图片" size="sm">
            <Icon name="image" size={15} />
          </IconButton>
          <IconButton label="插入代码" size="sm">
            <Icon name="code" size={15} />
          </IconButton>
          <IconButton label="添加标签" size="sm">
            <Icon name="tag" size={15} />
          </IconButton>
          {compact ? null : <span className="composer__dot" aria-hidden="true" />}
          <span className="composer__hint">{hint ?? `${current.length} / 500`}</span>
        </div>
        <div className="composer__actions">
          {compact ? null : (
            <Button variant="ghost" size="sm">
              存为草稿
            </Button>
          )}
          <Button size="sm" disabled={!canPublish} onClick={onPublish}>
            发布
          </Button>
        </div>
      </div>
    </div>
  )
}
