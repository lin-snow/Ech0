// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { useState, type ReactNode } from 'react'
import { AccountRow } from '../components/AccountRow'
import { ActivityHeatmap } from '../components/ActivityHeatmap'
import { Button } from '../components/Button'
import { Composer } from '../components/Composer'
import { PageHead, Section, Specimen } from '../components/Doc'
import { EchoEntry } from '../components/EchoEntry'
import { EmptyState } from '../components/EmptyState'
import { Icon } from '../components/Icon'
import { IconButton } from '../components/IconButton'
import { SearchField } from '../components/SearchField'
import { Tabs } from '../components/Tabs'
import { Tag } from '../components/Tag'
import { echos, heatmap, me } from '../fixtures/mock'

function State({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="state-label">
      <span className="state-label__caption">{label}</span>
      {children}
    </div>
  )
}

const privateEcho = { ...echos[2], id: 'private-demo', isPrivate: true }

export function Components() {
  const [activeTag, setActiveTag] = useState('思考')
  const [tab, setTab] = useState('timeline')
  const [draft, setDraft] = useState('今天把设计系统收进仓库，先定 token，再做组件。')

  return (
    <div className="page">
      <PageHead
        kicker="01 / COMPONENTS"
        meta="LIVE · INTERACTIVE"
        title="组件"
        lead="可交互的组件标本。每个组件只消费语义 token；状态用属性或 aria 表达，不复制样式。"
      />

      <Section id="button" index="01" title="按钮" note="PRIMARY / SECONDARY / GHOST">
        <div className="specimen-grid">
          <Specimen name="BUTTON / VARIANTS" note="MD">
            <Button>发布</Button>
            <Button variant="secondary">取消</Button>
            <Button variant="ghost">更多</Button>
            <Button variant="secondary" disabled>
              不可用
            </Button>
          </Specimen>

          <Specimen name="BUTTON / SIZES" note="MD · SM">
            <Button>发布</Button>
            <Button size="sm">发布</Button>
            <Button variant="secondary">保存</Button>
            <Button variant="secondary" size="sm">
              保存
            </Button>
          </Specimen>

          <Specimen name="BUTTON / STATES" note="PRIMARY">
            <State label="DEFAULT">
              <Button>发布</Button>
            </State>
            <State label="HOVER">
              <Button className="is-hover">发布</Button>
            </State>
            <State label="FOCUS">
              <Button className="is-focus">发布</Button>
            </State>
            <State label="PRESSED">
              <Button className="is-pressed">发布</Button>
            </State>
            <State label="DISABLED">
              <Button disabled>发布</Button>
            </State>
          </Specimen>

          <Specimen name="ICON BUTTON" note="32 / 26">
            <IconButton label="搜索">
              <Icon name="search" size={16} />
            </IconButton>
            <IconButton label="写作">
              <Icon name="pen" size={16} />
            </IconButton>
            <IconButton label="专注" aria-pressed="true">
              <Icon name="focus" size={16} />
            </IconButton>
            <IconButton label="更多" size="sm">
              <Icon name="more" size={14} />
            </IconButton>
            <IconButton label="主题" size="sm">
              <Icon name="moon" size={14} />
            </IconButton>
          </Specimen>
        </div>
      </Section>

      <Section id="search" index="02" title="搜索与筛选" note="DEFAULT · FOCUS · FILLED">
        <div className="specimen-grid">
          <Specimen name="SEARCH / STATES" note="WIDTH 100%">
            <div className="demo-col">
              <State label="DEFAULT">
                <SearchField />
              </State>
              <State label="FOCUS">
                <SearchField className="is-focus" />
              </State>
              <State label="FILLED">
                <SearchField value="设计系统" onChange={() => {}} shortcut="⌘K" />
              </State>
              <State label="PLAIN / 侧栏">
                <SearchField plain placeholder="搜索" shortcut="⌘K" />
              </State>
            </div>
          </Specimen>

          <Specimen name="TAG / FILTER" note="MULTI-SELECT">
            <div className="demo-col">
              <div className="demo-row">
                <Tag label="叽哩咕噜" />
                <Tag label="思考" active />
                <Tag label="更新日志" />
                <Tag label="摘抄" count={12} />
              </div>
              <div className="demo-row">
                <Tag
                  label="思考"
                  active={activeTag === '思考'}
                  onClick={() => setActiveTag('思考')}
                />
                <Tag
                  label="更新日志"
                  active={activeTag === '更新日志'}
                  onClick={() => setActiveTag('更新日志')}
                />
                <Tag
                  label="摘抄"
                  active={activeTag === '摘抄'}
                  onClick={() => setActiveTag('摘抄')}
                />
              </div>
              <span className="state-label__caption">CLICK TO TOGGLE · ACTIVE = ACCENT</span>
            </div>
          </Specimen>
        </div>
      </Section>

      <Section id="compose" index="03" title="导航与写作" note="TABS / COMPOSER">
        <div className="specimen-grid">
          <Specimen name="TABS" note="UNDERLINE = ACCENT">
            <div className="demo-col">
              <Tabs
                items={[
                  { id: 'timeline', label: '时间线', count: 128 },
                  { id: 'private', label: '私密', count: 9 },
                  { id: 'drafts', label: '草稿', count: 3 },
                ]}
                active={tab}
                onChange={setTab}
              />
              <span className="state-label__caption">ACTIVE: {tab.toUpperCase()}</span>
            </div>
          </Specimen>

          <Specimen name="COMPOSER / COMPACT" note="368 WIDE IN RAIL">
            <div style={{ width: '100%' }}>
              <Composer compact />
            </div>
          </Specimen>

          <Specimen name="COMPOSER / FULL" note="448 WIDE IN FOCUS" className="specimen--full">
            <div style={{ width: '100%', maxWidth: 448 }}>
              <Composer
                value={draft}
                onChange={setDraft}
                hint={`${draft.length} / 500`}
              />
            </div>
          </Specimen>
        </div>
      </Section>

      <Section id="entry" index="04" title="时间线" note="ANATOMY + PRIVATE">
        <div className="specimen-grid">
          <Specimen name="TIMELINE ENTRY / ANATOMY" note="448 WIDE" className="specimen--full">
            <div className="anatomy">
              <div className="anatomy__row">
                <span className="anatomy__key">TIME</span>
                8px 橙点 + 等宽时间戳；橙点只标记“有新内容”。
              </div>
              <div className="anatomy__row">
                <span className="anatomy__key">BODY</span>
                14 / 1.7；段间距 10px；链接用 accent + 下划线。
              </div>
              <div className="anatomy__row">
                <span className="anatomy__key">MEDIA</span>
                1px 线 + 8px 圆角；按原始宽高比，不裁切。
              </div>
              <div className="anatomy__row">
                <span className="anatomy__key">TAGS</span>
                22px 高、0 圆角；hover 只变边框和文字。
              </div>
              <div className="anatomy__row">
                <span className="anatomy__key">ACTIONS</span>
                24px 命中区；默认 ink-3，hover accent。
              </div>
            </div>
          </Specimen>

          <Specimen name="ECHO / PUBLIC" note="REAL CONTENT" className="specimen--full">
            <div style={{ width: '100%', maxWidth: 448 }}>
              <EchoEntry echo={echos[0]} />
            </div>
          </Specimen>

          <Specimen name="ECHO / PRIVATE" note="LOCK + LABEL" className="specimen--full">
            <div style={{ width: '100%', maxWidth: 448 }}>
              <EchoEntry echo={privateEcho} clamp={4} />
            </div>
          </Specimen>
        </div>
      </Section>

      <Section id="account" index="05" title="账户与空态" note="HEATMAP / EMPTY">
        <div className="specimen-grid">
          <Specimen name="ACTIVITY / HEATMAP" note="12 × 4 · 16PX">
            <div className="demo-col">
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
          </Specimen>

          <Specimen name="ACCOUNT ROW" note="24PX AVATAR">
            <div style={{ width: '100%', maxWidth: 260 }}>
              <AccountRow name={me.name} handle={me.handle} avatar={me.avatar} />
            </div>
          </Specimen>

          <Specimen name="EMPTY / NO CONTENT" note="ONE ACTION" className="specimen--full">
            <div style={{ width: '100%', maxWidth: 448 }}>
              <EmptyState
                icon="pen"
                title="还没有 Echo"
                body="写下第一条。之后可以随时编辑、设为私密，或只留在本地。"
                action="写第一条"
              />
            </div>
          </Specimen>
        </div>
      </Section>

      <Section id="contract" index="06" title="契约" note="COMPONENT → TOKEN">
        <table className="token-table">
          <thead>
            <tr>
              <th>组件</th>
              <th>关键值</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>Button</td>
              <td>h 32 / 26 · px 14 / 10 · radius 0</td>
              <td>hover accent-strong · disabled 38%</td>
            </tr>
            <tr>
              <td>SearchField</td>
              <td>h 32 · border 1px line</td>
              <td>focus border line-ink</td>
            </tr>
            <tr>
              <td>Tag</td>
              <td>h 22 · px 7 · radius 0</td>
              <td>active accent + accent-soft</td>
            </tr>
            <tr>
              <td>Composer</td>
              <td>p 12 · textarea 14 / 1.7</td>
              <td>focus 2px inset line-ink</td>
            </tr>
            <tr>
              <td>EchoEntry</td>
              <td>py 18 / 20 · media radius 8</td>
              <td>private lock + mono label</td>
            </tr>
            <tr>
              <td>Heatmap</td>
              <td>16px cell · 3px gap</td>
              <td>5 levels, heat-0…4</td>
            </tr>
          </tbody>
        </table>
      </Section>
    </div>
  )
}
