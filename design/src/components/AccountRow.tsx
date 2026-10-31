// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Icon } from './Icon'

type AccountRowProps = {
  name: string
  handle: string
  avatar: string
  onClick?: () => void
}

export function AccountRow({ name, handle, avatar, onClick }: AccountRowProps) {
  return (
    <button type="button" className="account" onClick={onClick}>
      <img className="account__avatar" src={avatar} alt="" width={24} height={24} />
      <span className="account__name">{name}</span>
      <span className="account__handle">{handle}</span>
      <Icon name="chevron" size={14} className="account__chevron" />
    </button>
  )
}
