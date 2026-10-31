// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Menu } from '@base-ui/react/menu'
import { Tabs } from '@base-ui/react/tabs'
import { useState, type ReactElement, type ReactNode } from 'react'
import { connects, heatmap, heatmapEnd, me, tagCloud } from '../fixtures/mock'
import { Icon, Wordmark, type IconName } from '../lib/icons'
import type { SketchFeed } from '../lib/useSketchFeed'
import { FeedToolbar } from './FeedToolbar'
import chrome from './chrome.module.css'
import ui from './ui.module.css'
import { IconButton, Tip } from './ui'

const navItems: { icon: IconName; label: string; active?: boolean }[] = [
  { icon: 'home', label: '首页', active: true },
  { icon: 'tags', label: '标签' },
  { icon: 'globe', label: '广场' },
  { icon: 'sliders', label: '面板' },
]

const pages = [
  { value: 'home', label: '首页' },
  { value: 'tags', label: '标签' },
  { value: 'square', label: '广场' },
  { value: 'panel', label: '面板' },
]

/* ---------------------------------------------------------------- account */

function MenuLabel({ children }: { children: ReactNode }) {
  return <div className={chrome.menuLabel}>{children}</div>
}

export function AccountMenu({ trigger }: { trigger: ReactElement }) {
  return (
    <Menu.Root>
      <Menu.Trigger render={trigger} />
      <Menu.Portal>
        <Menu.Positioner side="top" align="start" sideOffset={6}>
          <Menu.Popup className={chrome.menuPopup}>
            <MenuLabel>{me.handle}</MenuLabel>
            <Menu.Item className={chrome.menuItem}>设置</Menu.Item>
            <Menu.Item className={chrome.menuItem}>数据</Menu.Item>
            <Menu.Item className={chrome.menuItem}>退出</Menu.Item>
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  )
}

export function NavMenu() {
  return (
    <Menu.Root>
      <Menu.Trigger
        render={
          <button type="button" className={ui.iconButton} aria-label="导航">
            <Icon name="grid" size={14} />
          </button>
        }
      />
      <Menu.Portal>
        <Menu.Positioner side="bottom" align="end" sideOffset={4}>
          <Menu.Popup className={chrome.menuPopup}>
            {navItems.map((item) => (
              <Menu.Item key={item.label} className={chrome.menuItem}>
              {item.label}
            </Menu.Item>
          ))}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  )
}

function RailLinks() {
  return (
    <>
      <Tip label="RSS">
        <a
          className={ui.iconButton}
          href="/rss"
          target="_blank"
          rel="noreferrer"
          aria-label="RSS"
        >
          <Icon name="rss" size={14} />
        </a>
      </Tip>
      <Tip label="GitHub">
        <a
          className={ui.iconButton}
          href="https://github.com/lin-snow/Ech0"
          target="_blank"
          rel="noreferrer"
          aria-label="GitHub"
        >
          <Icon name="github" size={14} />
        </a>
      </Tip>
    </>
  )
}

function PageMenu() {
  const [page, setPage] = useState('home')
  return (
    <Menu.Root>
      <Menu.Trigger className={chrome.pageMenuTrigger} aria-label="页面">
        {pages.find((item) => item.value === page)?.label ?? '首页'}
        <Icon name="chevron-down" size={11} className={chrome.pageMenuIcon} />
      </Menu.Trigger>
      <Menu.Portal>
        <Menu.Positioner side="bottom" align="start" sideOffset={4}>
          <Menu.Popup className={chrome.menuPopup}>
            {pages.map((item) => (
              <Menu.Item
                key={item.value}
                className={chrome.menuItem}
                data-current={page === item.value || undefined}
                onClick={() => setPage(item.value)}
              >
                {item.label}
              </Menu.Item>
            ))}
          </Menu.Popup>
        </Menu.Positioner>
      </Menu.Portal>
    </Menu.Root>
  )
}

function AccountRow() {
  return (
    <button type="button" className={chrome.accountRow}>
      {me.name}
      <Icon name="chevron-down" size={10} />
    </button>
  )
}

/* ------------------------------------------------------------- left rail */

export function SiteRail({
  composer,
  feed,
}: {
  composer?: ReactNode
  feed: SketchFeed
}) {
  return (
    <aside className={chrome.siteRail}>
      <FeedToolbar
        feed={feed}
        variant="rail"
        leading={<PageMenu />}
        trailing={<RailLinks />}
      />

      {composer}

      <ActivityTabs />

      <div className={chrome.tags}>
        <div className={chrome.tagsLabel}>TAGS</div>
        <div className={chrome.tagsCloud}>
          {tagCloud.map((tag) => (
            <button
              key={tag}
              type="button"
              className={chrome.tagItem}
              data-active={feed.tag === tag || undefined}
              onClick={() => feed.selectTag(tag)}
            >
              #{tag}
            </button>
          ))}
        </div>
      </div>

      <div className={chrome.railSpacer} />

      <AccountMenu trigger={<AccountRow />} />
    </aside>
  )
}

/* ------------------------------------------------------------- icon rail */

/* ---------------------------------------------------------- context rail */

function Heatmap() {
  const rows = 4
  const columns = Math.ceil(heatmap.length / rows)
  const end = new Date(`${heatmapEnd}T00:00:00`)
  return (
    <div
      className={chrome.heatmap}
      role="img"
      aria-label={`近 ${heatmap.length} 天发布热力图`}
      style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }}
    >
      {heatmap.map((level, index) => {
        const date = new Date(end)
        date.setDate(end.getDate() - (heatmap.length - 1 - index))
        return (
          <span
            key={index}
            className={chrome.heatCell}
            style={{ background: `var(--heatmap-bg-color-${level})` }}
            title={`${date.getMonth() + 1} 月 ${date.getDate()} 日 · ${level} 条`}
          />
        )
      })}
    </div>
  )
}

function ConnectWidget() {
  return (
    <div className={chrome.connect}>
      <div className={chrome.connectGrid}>
        {connects.map((connect) => (
          <Tip key={connect.name} label={`${connect.name} · 今日 ${connect.today}`} side="top">
            <button type="button" className={chrome.connectTile} aria-label={connect.name}>
              <img src={connect.logo} alt="" loading="lazy" />
              <span
                className={chrome.connectDot}
                style={{
                  background:
                    connect.today > 0
                      ? `var(--heatmap-bg-color-${Math.min(connect.today, 4)})`
                      : 'var(--line-strong)',
                }}
              />
            </button>
          </Tip>
        ))}
      </div>
    </div>
  )
}

export function ActivityTabs() {
  return (
    <Tabs.Root className={chrome.activityTabs} defaultValue="activity">
      <Tabs.List className={chrome.activityTabList}>
        <Tabs.Tab className={chrome.activityTab} value="activity">
          ACTIVITY
        </Tabs.Tab>
        <Tabs.Tab className={chrome.activityTab} value="connect">
          CONNECT
        </Tabs.Tab>
        <Tabs.Indicator className={chrome.activityTabIndicator} />
      </Tabs.List>
      <Tabs.Panel className={chrome.activityPanel} value="activity">
        <Heatmap />
      </Tabs.Panel>
      <Tabs.Panel className={chrome.activityPanel} value="connect">
        <ConnectWidget />
      </Tabs.Panel>
    </Tabs.Root>
  )
}

/* ------------------------------------------------------------- mobile */

export function MobileHeader({ onCompose }: { onCompose: () => void }) {
  return (
    <div className={chrome.mobileHeader}>
      <div className={chrome.mobileBrand}>
        <Wordmark compact />
      </div>
      <IconButton icon="plus" label="写一条" size={14} onClick={onCompose} />
      <NavMenu />
    </div>
  )
}

/* -------------------------------------------------------------- focus */

export function FocusTop() {
  return (
    <div className={chrome.focusTop}>
      <Wordmark compact />
      <div className={chrome.focusTopActions}>
        <AccountMenu
          trigger={
            <button type="button" className={chrome.accountTextButton}>
              {me.name}
              <Icon name="chevron-down" size={9} />
            </button>
          }
        />
      </div>
    </div>
  )
}
