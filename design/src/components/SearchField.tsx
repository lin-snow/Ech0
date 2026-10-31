// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useId, useState } from 'react'
import { Icon } from './Icon'
import { cx } from '../lib/cx'

type SearchFieldProps = {
  value?: string
  onChange?: (value: string) => void
  placeholder?: string
  shortcut?: string
  plain?: boolean
  label?: string
  className?: string
}

export function SearchField({
  value,
  onChange,
  placeholder = '搜索',
  shortcut,
  plain = false,
  label = '搜索',
  className,
}: SearchFieldProps) {
  const id = useId()
  const [inner, setInner] = useState('')
  const current = value ?? inner

  function update(next: string) {
    if (value === undefined) setInner(next)
    onChange?.(next)
  }

  return (
    <div className={cx('search', plain && 'search--plain', className)}>
      <Icon name="search" size={15} />
      <label htmlFor={id} className="sr-only">
        {label}
      </label>
      <input
        id={id}
        className="search__input"
        value={current}
        onChange={(event) => update(event.target.value)}
        placeholder={placeholder}
        autoComplete="off"
        spellCheck={false}
      />
      {shortcut ? <kbd className="kbd search__kbd">{shortcut}</kbd> : null}
    </div>
  )
}
