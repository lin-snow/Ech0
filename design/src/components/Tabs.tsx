// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

type TabItem = {
  id: string
  label: string
  count?: number
}

type TabsProps = {
  items: TabItem[]
  active: string
  onChange: (id: string) => void
  label?: string
}

export function Tabs({ items, active, onChange, label = '视图' }: TabsProps) {
  return (
    <div className="tabs" role="tablist" aria-label={label}>
      {items.map((item) => (
        <button
          key={item.id}
          type="button"
          role="tab"
          aria-selected={active === item.id}
          className="tab"
          onClick={() => onChange(item.id)}
        >
          {item.label}
          {item.count !== undefined ? <span className="tab__count">{item.count}</span> : null}
        </button>
      ))}
    </div>
  )
}
