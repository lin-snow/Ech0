// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Button } from './Button'
import { Icon, type IconName } from './Icon'

type EmptyStateProps = {
  icon?: IconName
  title: string
  body: string
  action?: string
  onAction?: () => void
}

export function EmptyState({ icon = 'spark', title, body, action, onAction }: EmptyStateProps) {
  return (
    <div className="empty">
      <span className="empty__mark">
        <Icon name={icon} size={18} />
      </span>
      <h3 className="empty__title">{title}</h3>
      <p className="empty__body">{body}</p>
      {action ? (
        <Button variant="secondary" size="sm" onClick={onAction}>
          {action}
        </Button>
      ) : null}
    </div>
  )
}
