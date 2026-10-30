// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Select } from '@base-ui/react/select'
import { Slider } from '@base-ui/react/slider'
import { Tabs } from '@base-ui/react/tabs'
import chrome from './chrome.module.css'
import { Icon } from '../lib/icons'
import { Button, Mono } from './ui'

export type LayoutVariant = 'three' | 'focus'
export type Theme = 'light' | 'dark'

const variants: { value: LayoutVariant; label: string }[] = [
  { value: 'three', label: '侧栏' },
  { value: 'focus', label: '专注' },
]

const themes: { value: Theme; label: string }[] = [
  { value: 'light', label: '浅色' },
  { value: 'dark', label: '深色' },
]

export function DesignBar({
  variant,
  onVariant,
  theme,
  onTheme,
  width,
  onWidth,
  onReset,
}: {
  variant: LayoutVariant
  onVariant: (value: LayoutVariant) => void
  theme: Theme
  onTheme: (value: Theme) => void
  width: number
  onWidth: (value: number) => void
  onReset: () => void
}) {
  return (
    <div className={chrome.designBar} role="toolbar" aria-label="设计草图控制">
      <span className={chrome.labName}>ECH0 6.0</span>

      <Tabs.Root value={variant} onValueChange={(value) => onVariant(value as LayoutVariant)}>
        <Tabs.List className={chrome.labTabs}>
          {variants.map((item) => (
            <Tabs.Tab key={item.value} value={item.value} className={chrome.labTab}>
              {item.label}
            </Tabs.Tab>
          ))}
        </Tabs.List>
      </Tabs.Root>

      <span className={chrome.labDivider} />

      <div className={chrome.labControl}>
        <Select.Root value={theme} onValueChange={(value) => onTheme(value as Theme)}>
          <Select.Trigger className={chrome.selectTrigger} aria-label="主题">
            <Select.Value>
              {(value: Theme) => themes.find((item) => item.value === value)?.label ?? '浅色'}
            </Select.Value>
            <Select.Icon>
              <Icon name="chevron-down" size={10} />
            </Select.Icon>
          </Select.Trigger>
          <Select.Portal>
            <Select.Positioner sideOffset={4}>
              <Select.Popup className={chrome.selectPopup}>
                {themes.map((item) => (
                  <Select.Item key={item.value} value={item.value} className={chrome.selectItem}>
                    <Select.ItemText>{item.label}</Select.ItemText>
                  </Select.Item>
                ))}
              </Select.Popup>
            </Select.Positioner>
          </Select.Portal>
        </Select.Root>
      </div>

      <div className={chrome.widthControl}>
        <Mono className={chrome.widthValue}>{width}px</Mono>
        <Slider.Root
          min={390}
          max={1440}
          step={10}
          value={[width]}
          onValueChange={(values) => onWidth(values[0])}
          className={chrome.slider}
          aria-label="画布宽度"
        >
          <Slider.Control className={chrome.sliderControl}>
            <Slider.Track className={chrome.sliderTrack}>
              <Slider.Indicator className={chrome.sliderIndicator} />
            </Slider.Track>
            <Slider.Thumb className={chrome.sliderThumb} />
          </Slider.Control>
        </Slider.Root>
      </div>

      <Button size="sm" variant="ghost" icon="refresh" aria-label="重置" onClick={onReset} />

    </div>
  )
}
