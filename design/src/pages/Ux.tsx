// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { Fragment } from 'react'
import { PageHead, Section } from '../components/Doc'

const ia = [
  { label: '时间线 / Timeline', meta: 'ROOT', child: false },
  { label: '侧栏 · 写作 / 活跃度 / 标签 / 账户', meta: '400', child: true },
  { label: '专注 · 单列阅读 + 写作', meta: '448', child: true },
  { label: '搜索 · 关键词 / 标签过滤', meta: '⌘K', child: true },
  { label: 'Connect · 友邻动态', meta: 'OPTIONAL', child: true },
  { label: '设置 · 主题 / 私密 / 数据', meta: 'ACCOUNT', child: true },
]

const flows = [
  {
    name: '01 WRITE',
    steps: ['时间线', '输入', '图片 / 标签', '发布', '乐观插入', '失败可重试'],
  },
  {
    name: '02 BROWSE',
    steps: ['首页', '标签', '过滤', '空态', '清除筛选'],
  },
  {
    name: '03 SEARCH',
    steps: ['⌘K', '输入', '结果', '打开 Echo', 'Esc 关闭'],
  },
  {
    name: '04 FOCUS',
    steps: ['专注', '单列', '返回', '保持滚动位置'],
  },
  {
    name: '05 THEME',
    steps: ['账户', '设置', '浅色 / 深色', '持久化'],
  },
]

const states = [
  {
    surface: 'Timeline',
    empty: '“还没有 Echo” + 写第一条',
    loading: '3 条骨架，保持行高',
    error: '顶部 1 行错误 + 重试',
    permission: '公开只读；隐藏写入口',
  },
  {
    surface: 'Search',
    empty: '“没有匹配” + 清除关键词',
    loading: '输入 300ms 后显示进度',
    error: '保留输入，结果区报错',
    permission: '私密结果不返回',
  },
  {
    surface: 'Composer',
    empty: '发布禁用；草稿可存',
    loading: '发布中禁用 + 文案“发布中”',
    error: '保留正文 + 重试',
    permission: '未登录时引导登录',
  },
  {
    surface: 'Echo',
    empty: '不渲染空正文；仅图片/标签',
    loading: '图片占位保持宽高比',
    error: '图片失败显示替代框',
    permission: '私密加锁 + PRIVATE',
  },
  {
    surface: 'Account',
    empty: '默认头像 + handle',
    loading: '24px 占位',
    error: '保留入口，设置页报错',
    permission: '未登录显示“登录”',
  },
]

export function Ux() {
  return (
    <div className="page">
      <PageHead
        kicker="03 / UX"
        meta="IA · FLOWS · STATES"
        title="结构与状态"
        lead="一个时间线，两种阅读模式。先定结构和状态，再加页面；所有流程都要有失败、空和权限分支。"
      />

      <Section id="ia" index="01" title="信息架构" note="ONE ROOT">
        <div className="ia">
          {ia.map((row) => (
            <div className={`ia-row${row.child ? ' ia-row--child' : ''}`} key={row.label}>
              {row.child ? <span className="ia-row__line" aria-hidden="true" /> : null}
              <span className="ia-row__label">{row.label}</span>
              <span className="ia-row__meta">{row.meta}</span>
            </div>
          ))}
        </div>
      </Section>

      <Section id="flows" index="02" title="核心流程" note="5 FLOWS">
        <div className="flow-list">
          {flows.map((flow) => (
            <div className="flow" key={flow.name}>
              <span className="flow__name">{flow.name}</span>
              {flow.steps.map((step, index) => (
                <Fragment key={step}>
                  {index > 0 ? <span className="flow__arrow">→</span> : null}
                  <span className="flow__step">{step}</span>
                </Fragment>
              ))}
            </div>
          ))}
        </div>
      </Section>

      <Section id="states" index="03" title="状态覆盖" note="EMPTY / LOADING / ERROR / PERMISSION">
        <table className="state-table">
          <thead>
            <tr>
              <th>Surface</th>
              <th>Empty</th>
              <th>Loading</th>
              <th>Error</th>
              <th>Permission</th>
            </tr>
          </thead>
          <tbody>
            {states.map((row) => (
              <tr key={row.surface}>
                <td>{row.surface}</td>
                <td>{row.empty}</td>
                <td>{row.loading}</td>
                <td>{row.error}</td>
                <td>{row.permission}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Section>

      <Section id="acceptance" index="04" title="验收" note="BEFORE HANDOFF">
        <div className="rule-list">
          <div className="rule">
            <span className="rule__num">01</span>
            <p className="rule__body">
              <strong>键盘。</strong> Tab 顺序与视觉顺序一致；Esc 关闭搜索/弹层；⌘K 可打开搜索。
            </p>
          </div>
          <div className="rule">
            <span className="rule__num">02</span>
            <p className="rule__body">
              <strong>对比度。</strong> 正文 ≥ 4.5:1；ink-3 只用于元信息，不承载关键操作。
            </p>
          </div>
          <div className="rule">
            <span className="rule__num">03</span>
            <p className="rule__body">
              <strong>文案。</strong> 中文为主，英文只做标签；空态给下一步，错误给可执行动作。
            </p>
          </div>
          <div className="rule">
            <span className="rule__num">04</span>
            <p className="rule__body">
              <strong>断点。</strong> 1024 / 880 / 560 三档；不新增第四套布局。
            </p>
          </div>
        </div>
      </Section>
    </div>
  )
}
