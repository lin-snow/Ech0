// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { PageHead, Section } from '../components/Doc'
import { FocusPattern, MobilePattern, SidebarPattern } from '../patterns/Screens'

export function Patterns() {
  return (
    <div className="page">
      <PageHead
        kicker="02 / PATTERNS"
        meta="1024 · 880 · 390"
        title="模式"
        lead="三个真实布局：侧栏、专注、移动。尺寸来自产品，不是示意图；主题切换会同时作用于它们。"
      />

      <Section id="sidebar" index="01" title="侧栏 · 双栏" note="400 RAIL + 480 FEED">
        <div className="preview-scroll">
          <SidebarPattern />
        </div>
        <div className="preview-caption">
          <span>1024 FRAME / 880 CONTENT</span>
          <span>RAIL 400 · FEED 480 · 18/16 PADDING</span>
        </div>
      </Section>

      <Section id="focus" index="02" title="专注 · 单列" note="448 CONTENT">
        <div className="preview-scroll">
          <FocusPattern />
        </div>
        <div className="preview-caption">
          <span>1024 FRAME / 448 CONTENT</span>
          <span>COMPOSER 448 · ENTRY 448</span>
        </div>
      </Section>

      <Section id="mobile" index="03" title="移动端" note="390 × 760">
        <div className="preview-scroll">
          <MobilePattern />
        </div>
        <div className="preview-caption">
          <span>390 FRAME / 358 CONTENT</span>
          <span>NO STATUS BAR · WEB APP</span>
        </div>
      </Section>

      <Section id="breakpoints" index="04" title="断点行为" note="880 / 560">
        <div className="breakpoint-list">
          <div className="breakpoint">
            <span className="breakpoint__size">≥ 1024</span>
            <span className="breakpoint__body">侧栏 400 + 时间线 480；侧栏常驻，搜索在时间线头部。</span>
          </div>
          <div className="breakpoint">
            <span className="breakpoint__size">880 – 1023</span>
            <span className="breakpoint__body">侧栏收起为顶部导航；时间线占满，内容最大 560。</span>
          </div>
          <div className="breakpoint">
            <span className="breakpoint__size">&lt; 560</span>
            <span className="breakpoint__body">单列；内容 358（390 − 32）；写作入口固定在底部。</span>
          </div>
          <div className="breakpoint">
            <span className="breakpoint__size">THEME</span>
            <span className="breakpoint__body">浅色 / 深色共用同一组语义 token；不单独维护两套组件。</span>
          </div>
        </div>
      </Section>
    </div>
  )
}
