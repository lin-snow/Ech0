// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Combobox } from '@base-ui/react/combobox'
import { Menu } from '@base-ui/react/menu'
import { Menubar } from '@base-ui/react/menubar'
import { Toggle } from '@base-ui/react/toggle'
import { useRef } from 'react'
import { me, tagCloud } from '../data/mock'
import { Icon } from '../lib/icons'
import type { ComposerState } from '../lib/useComposer'
import styles from './composer.module.css'
import { Avatar, Button, cx, Mono } from './ui'

const insertMenus: { label: string; items: { label: string; snippet: string }[] }[] = [
  {
    label: '媒体',
    items: [
      { label: '图片', snippet: '![](https://)' },
      { label: '视频', snippet: '<video controls src="https://"></video>' },
      { label: '音频', snippet: '<audio controls src="https://"></audio>' },
    ],
  },
  {
    label: '扩展',
    items: [
      { label: '网站', snippet: 'https://' },
      { label: '音乐', snippet: 'https://music.163.com/song?id=' },
      { label: 'GitHub 项目', snippet: 'https://github.com/' },
      { label: '推文', snippet: 'https://x.com/' },
    ],
  },
]

function InsertMenubar({ onInsert }: { onInsert: (snippet: string) => void }) {
  return (
    <Menubar className={styles.menubar}>
      {insertMenus.map((menu) => (
        <Menu.Root key={menu.label}>
          <Menu.Trigger className={styles.menubarTrigger}>{menu.label}</Menu.Trigger>
          <Menu.Portal>
            <Menu.Positioner side="top" align="start" sideOffset={6}>
              <Menu.Popup className={styles.menubarPopup}>
                {menu.items.map((item) => (
                  <Menu.Item
                    key={item.label}
                    className={styles.menubarItem}
                    onClick={() => onInsert(item.snippet)}
                  >
                    {item.label}
                  </Menu.Item>
                ))}
              </Menu.Popup>
            </Menu.Positioner>
          </Menu.Portal>
        </Menu.Root>
      ))}
    </Menubar>
  )
}

export function Composer({
  composer,
  compact = false,
}: {
  composer: ComposerState
  compact?: boolean
}) {
  const areaRef = useRef<HTMLTextAreaElement>(null)

  const insert = (snippet: string) => {
    const area = areaRef.current
    const start = area?.selectionStart ?? composer.body.length
    const end = area?.selectionEnd ?? start
    composer.setBody(composer.body.slice(0, start) + snippet + composer.body.slice(end))
    requestAnimationFrame(() => {
      area?.focus()
      area?.setSelectionRange(start + snippet.length, start + snippet.length)
    })
  }

  return (
    <div className={cx(styles.editor, compact && styles.editorCompact)}>
      <div className={styles.body}>
        <Avatar src={me.avatar} alt={me.name} size={compact ? 34 : 36} />
        <div className={styles.inputCol}>
          <div className={styles.who}>
            {me.name}
            <Mono className={styles.handle}>{me.handle}</Mono>
          </div>
          <textarea
            ref={areaRef}
            data-composer
            className={styles.area}
            value={composer.body}
            onChange={(event) => composer.setBody(event.target.value)}
            onKeyDown={(event) => {
              if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
                event.preventDefault()
                composer.publish()
              }
            }}
            placeholder="写点什么…"
            rows={compact ? 4 : 3}
            aria-label="写一条 echo"
          />
        </div>
      </div>
      <div className={styles.foot}>
        <InsertMenubar onInsert={insert} />
        <Combobox.Root
          multiple
          items={tagCloud}
          value={composer.tags}
          onValueChange={composer.setTags}
        >
          <Combobox.InputGroup className={styles.tagGroup}>
            <Combobox.Value>
              {(tags: string[]) => (
                <Combobox.Chips
                  className={styles.tagChips}
                  aria-label={tags.length ? '已选标签' : undefined}
                >
                  {tags.map((tag) => (
                    <Combobox.Chip key={tag} className={styles.tagChip}>
                      #{tag}
                      <Combobox.ChipRemove className={styles.tagRemove} aria-label={`移除 ${tag}`}>
                        <Icon name="close" size={9} />
                      </Combobox.ChipRemove>
                    </Combobox.Chip>
                  ))}
                  <Combobox.Input
                    className={styles.tagInput}
                    placeholder={tags.length ? '' : '#tag'}
                    aria-label="标签"
                  />
                </Combobox.Chips>
              )}
            </Combobox.Value>
          </Combobox.InputGroup>
          <Combobox.Portal>
            <Combobox.Positioner side="top" align="start" sideOffset={6}>
              <Combobox.Popup className={styles.tagPopup}>
                <Combobox.Empty className={styles.tagEmpty}>没有匹配的标签</Combobox.Empty>
                <Combobox.List>
                  {(item: string) => (
                    <Combobox.Item key={item} value={item} className={styles.tagOption}>
                      #{item}
                    </Combobox.Item>
                  )}
                </Combobox.List>
              </Combobox.Popup>
            </Combobox.Positioner>
          </Combobox.Portal>
        </Combobox.Root>
        <Toggle
          className={styles.lock}
          pressed={composer.isPrivate}
          onPressedChange={composer.setIsPrivate}
          aria-label="私密"
        >
          <Icon name="lock" size={11} />
        </Toggle>
        <Button
          size="sm"
          variant="primary"
          disabled={!composer.canPublish}
          onClick={composer.publish}
        >
          发布
        </Button>
      </div>
    </div>
  )
}
