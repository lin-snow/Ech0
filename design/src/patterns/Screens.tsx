// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { echos, heatmap, me, tagCloud } from '../fixtures/mock'
import { AccountRow } from '../components/AccountRow'
import { ActivityHeatmap } from '../components/ActivityHeatmap'
import { Composer } from '../components/Composer'
import { EchoEntry } from '../components/EchoEntry'
import { Icon } from '../components/Icon'
import { IconButton } from '../components/IconButton'
import { SearchField } from '../components/SearchField'
import { Tag } from '../components/Tag'

export function SidebarPattern() {
  return (
    <section className="app-frame" aria-label="侧栏模式预览">
      <aside className="app-rail">
        <div className="rail-brand">
          <span className="wordmark">echo</span>
          <span className="square" aria-hidden="true" />
        </div>

        <nav className="rail-nav" aria-label="主导航">
          <a className="rail-nav__item" data-active="true" href="#timeline">
            <Icon name="home" size={15} />
            首页
          </a>
          <a className="rail-nav__item" href="#search">
            <Icon name="search" size={15} />
            搜索
          </a>
          <a className="rail-nav__item" href="#focus">
            <Icon name="focus" size={15} />
            专注
          </a>
        </nav>

        <div className="rail-block">
          <div className="rail-label">ACTIVITY / 活跃度</div>
          <ActivityHeatmap data={heatmap} />
          <div className="heatmap-legend">
            <span>少</span>
            <span className="heat-cell" data-level="1" />
            <span className="heat-cell" data-level="2" />
            <span className="heat-cell" data-level="3" />
            <span className="heat-cell" data-level="4" />
            <span>多</span>
          </div>
        </div>

        <div className="rail-block">
          <div className="rail-label">TAGS / 标签</div>
          <div className="rail-tags">
            {tagCloud.map((tag) => (
              <Tag key={tag} label={tag} />
            ))}
          </div>
        </div>

        <div className="rail-spacer" />
        <AccountRow name={me.name} handle={me.handle} avatar={me.avatar} />
      </aside>

      <main className="app-feed">
        <div className="feed-head">
          <div className="feed-title">时间线</div>
          <SearchField plain placeholder="搜索" shortcut="⌘K" />
        </div>
        <Composer compact />
        <div className="feed-list">
          {echos.slice(0, 2).map((echo) => (
            <EchoEntry key={echo.id} echo={echo} />
          ))}
        </div>
      </main>
    </section>
  )
}

export function FocusPattern() {
  return (
    <section className="focus-frame" aria-label="专注模式预览">
      <header className="focus-bar">
        <IconButton label="返回时间线" size="sm">
          <Icon name="chevron-left" size={16} />
        </IconButton>
        <span className="focus-title">专注</span>
        <div className="focus-search">
          <SearchField plain placeholder="搜索" shortcut="⌘K" />
        </div>
      </header>

      <div className="focus-content">
        <Composer />
        <EchoEntry echo={echos[0]} clamp={7} />
        <EchoEntry echo={echos[2]} clamp={4} />
      </div>
    </section>
  )
}

export function MobilePattern() {
  return (
    <section className="mobile-frame" aria-label="移动端预览">
      <header className="mobile-head">
        <span className="wordmark">echo</span>
        <AccountRow name={me.name} handle={me.handle} avatar={me.avatar} />
      </header>

      <div className="mobile-body">
        <SearchField plain placeholder="搜索" shortcut="⌘K" />
        <Composer compact />
        <EchoEntry echo={echos[2]} clamp={5} />
      </div>
    </section>
  )
}
