// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Icon } from '../components/Icon'
import type { PageId } from '../routes'
import { colorGroups, geometry, motion, spaceScale, typeScale } from '../tokens'

type AssetLink = {
  label: string
  anchor: string
  meta: string
}

type AssetGroup = {
  page: PageId
  num: string
  title: string
  kicker: string
  desc: string
  links: AssetLink[]
}

const groups: AssetGroup[] = [
  {
    page: 'foundations',
    num: '00',
    title: '基础',
    kicker: 'TOKENS / 6',
    desc: '颜色、字体、间距、几何与动效。所有值来自同一份 CSS 变量。',
    links: [
      { label: '色彩', anchor: 'color', meta: '19 TOKENS' },
      { label: '字体', anchor: 'type', meta: '10 级' },
      { label: '间距', anchor: 'space', meta: '4PX BASE' },
      { label: '几何', anchor: 'geometry', meta: '0 / 8 RADIUS' },
      { label: '动效', anchor: 'motion', meta: '3 DURATIONS' },
      { label: '规则', anchor: 'rules', meta: '3 RULES' },
    ],
  },
  {
    page: 'components',
    num: '01',
    title: '组件',
    kicker: 'PRIMITIVES / 6',
    desc: '可交互标本。每个组件只消费语义 token，状态用属性或 aria 表达。',
    links: [
      { label: '按钮', anchor: 'button', meta: '3 VARIANTS' },
      { label: '搜索与筛选', anchor: 'search', meta: 'SEARCH / TAG' },
      { label: '导航与写作', anchor: 'compose', meta: 'TABS / COMPOSER' },
      { label: '时间线', anchor: 'entry', meta: 'ENTRY / PRIVATE' },
      { label: '账户与空态', anchor: 'account', meta: 'HEATMAP / EMPTY' },
      { label: '契约', anchor: 'contract', meta: 'COMPONENT → TOKEN' },
    ],
  },
  {
    page: 'patterns',
    num: '02',
    title: '模式',
    kicker: 'LAYOUTS / 4',
    desc: '三个真实布局。尺寸来自产品，不是示意图；主题切换同时作用于它们。',
    links: [
      { label: '侧栏 · 双栏', anchor: 'sidebar', meta: '400 + 480' },
      { label: '专注 · 单列', anchor: 'focus', meta: '448' },
      { label: '移动端', anchor: 'mobile', meta: '390' },
      { label: '断点行为', anchor: 'breakpoints', meta: '1024 / 880 / 560' },
    ],
  },
  {
    page: 'ux',
    num: '03',
    title: '结构与状态',
    kicker: 'UX / 4',
    desc: 'IA、核心流程、状态矩阵与验收项。先定结构，再加页面。',
    links: [
      { label: '信息架构', anchor: 'ia', meta: 'ONE ROOT' },
      { label: '核心流程', anchor: 'flows', meta: '5 FLOWS' },
      { label: '状态覆盖', anchor: 'states', meta: '4 STATES' },
      { label: '验收', anchor: 'acceptance', meta: '4 CHECKS' },
    ],
  },
]

const tokenCount =
  colorGroups.reduce((sum, group) => sum + group.colors.length, 0) +
  typeScale.length +
  spaceScale.length +
  geometry.length +
  motion.length

export function Home({ onNavigate }: { onNavigate: (page: PageId, anchor?: string) => void }) {
  const entryCount = groups.reduce((sum, group) => sum + group.links.length, 0)

  return (
    <div className="page home">
      <header className="home-hero">
        <div className="page-kicker">
          <span className="eyebrow">INDEX / ECH0 DESIGN</span>
          <span className="eyebrow">
            {groups.length} GROUPS · {entryCount} ENTRIES · {tokenCount} TOKENS
          </span>
        </div>
        <h1>设计资产库</h1>
        <p className="lead">
          按“基础 → 组件 → 模式 → 结构”组织的可点击索引。点组标题进入页面，点条目直接跳到对应 section。
        </p>
      </header>

      <div className="asset-grid">
        {groups.map((group) => (
          <article className="asset-card" key={group.page}>
            <button
              type="button"
              className="asset-card__head"
              onClick={() => onNavigate(group.page)}
            >
              <span className="asset-card__num">{group.num}</span>
              <span className="asset-card__title">{group.title}</span>
              <span className="asset-card__kicker">{group.kicker}</span>
              <Icon name="arrow-right" size={16} />
            </button>

            <p className="asset-card__desc">{group.desc}</p>

            <div className="asset-links">
              {group.links.map((link) => (
                <button
                  type="button"
                  className="asset-link"
                  key={link.anchor}
                  onClick={() => onNavigate(group.page, link.anchor)}
                >
                  <span>{link.label}</span>
                  <span className="asset-link__meta">{link.meta}</span>
                  <Icon name="chevron" size={12} />
                </button>
              ))}
            </div>
          </article>
        ))}
      </div>

      <div className="home-foot">
        <span className="eyebrow">TOKEN SOURCE</span>
        <button
          type="button"
          className="home-foot__link"
          onClick={() => onNavigate('foundations', 'color')}
        >
          src/styles/tokens.css
          <Icon name="arrow-right" size={12} />
        </button>
      </div>
    </div>
  )
}
