// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { ReactNode } from 'react'
import { cx } from '../lib/cx'

export function PageHead({
  kicker,
  title,
  lead,
  meta,
}: {
  kicker: string
  title: string
  lead: string
  meta?: string
}) {
  return (
    <header className="page-head">
      <div className="page-kicker">
        <span className="eyebrow">{kicker}</span>
        {meta ? <span className="eyebrow">{meta}</span> : null}
      </div>
      <h1>{title}</h1>
      <p className="lead">{lead}</p>
    </header>
  )
}

export function Section({
  id,
  index,
  title,
  note,
  children,
}: {
  id?: string
  index: string
  title: string
  note?: string
  children: ReactNode
}) {
  return (
    <section className="section" id={id}>
      <div className="section-head">
        <span className="section-index">{index}</span>
        <h2 className="section-title">{title}</h2>
        {note ? <span className="section-note">{note}</span> : null}
      </div>
      {children}
    </section>
  )
}

export function Specimen({
  name,
  note,
  children,
  footer,
  surface = false,
  className,
}: {
  name: string
  note?: string
  children: ReactNode
  footer?: ReactNode
  surface?: boolean
  className?: string
}) {
  return (
    <div className={cx('specimen', className)}>
      <div className="specimen__head">
        <span className="specimen__name">{name}</span>
        {note ? <span className="specimen__note">{note}</span> : null}
      </div>
      <div className={cx('specimen__body', surface && 'specimen__body--surface')}>{children}</div>
      {footer ? <div className="specimen__foot">{footer}</div> : null}
    </div>
  )
}
