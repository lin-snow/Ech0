// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { PageHead, Section } from '../components/Doc'
import { colorGroups, geometry, motion, spaceScale, typeScale } from '../tokens'

export function Foundations() {
  return (
    <div className="page">
      <PageHead
        kicker="00 / FOUNDATIONS"
        meta="v0.1 · 2026.10"
        title="基础"
        lead="颜色、字体、间距、几何和动效。所有值都来自同一组 CSS 变量；组件只消费语义 token，不写死数值。"
      />

      <Section id="color" index="01" title="色彩" note="LIGHT / DARK 双值">
        {colorGroups.map((group) => (
          <div className="token-group" key={group.title}>
            <div className="token-group__head">
              <h3 className="token-group__title">{group.title}</h3>
              <span className="token-group__note">{group.note}</span>
            </div>
            <div className="swatch-grid">
              {group.colors.map((color) => (
                <div className="swatch" key={color.token}>
                  <div className="swatch__color" style={{ background: `var(${color.token})` }} />
                  <div className="swatch__meta">
                    <span className="swatch__name">{color.token}</span>
                    <span className="swatch__hex">
                      {color.light} / {color.dark}
                    </span>
                    <span className="swatch__use">{color.use}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </Section>

      <Section id="type" index="02" title="字体" note="SYSTEM UI / MONO">
        <div className="type-list">
          {typeScale.map((item) => (
            <div className="type-row" key={item.token}>
              <span
                className="type-row__sample"
                style={{ fontSize: `${item.px}px`, lineHeight: item.line }}
              >
                Ech0 时间线
              </span>
              <span className="type-row__meta">
                {item.token} · {item.px}px / {item.line} · {item.use}
              </span>
            </div>
          ))}
        </div>
      </Section>

      <Section id="space" index="03" title="间距" note="4PX BASE">
        <div className="space-list">
          {spaceScale.map((item) => (
            <div className="space-row" key={item.token}>
              <span className="space-row__token">{item.token}</span>
              <span className="space-bar" style={{ width: `${item.px}px` }} />
              <span className="space-row__px">{item.px}px</span>
            </div>
          ))}
        </div>
      </Section>

      <Section id="geometry" index="04" title="几何" note="0 RADIUS / 8PX IMAGES">
        <div className="geometry-row">
          <div className="geometry-card">
            <div className="geometry-card__frame" aria-hidden="true">
              <span className="geometry-card__rail" />
              <span className="geometry-card__feed" />
            </div>
            <span className="geometry-card__label">侧栏 + 时间线</span>
            <span className="geometry-card__value">400 + 480 = 880</span>
          </div>
          <div className="geometry-card">
            <div className="geometry-card__frame" aria-hidden="true">
              <span
                className="geometry-card__feed"
                style={{ left: '30%', right: '30%', top: 20, bottom: 20 }}
              />
            </div>
            <span className="geometry-card__label">专注内容</span>
            <span className="geometry-card__value">448 / 单列</span>
          </div>
          <div className="geometry-card">
            <div className="geometry-card__frame" aria-hidden="true">
              <span
                className="geometry-card__feed"
                style={{ left: 24, right: 24, top: 16, bottom: 16 }}
              />
            </div>
            <span className="geometry-card__label">移动内容</span>
            <span className="geometry-card__value">390 − 32 = 358</span>
          </div>
        </div>

        <table className="token-table" style={{ marginTop: 20 }}>
          <thead>
            <tr>
              <th>Token</th>
              <th>值</th>
              <th>用途</th>
            </tr>
          </thead>
          <tbody>
            {geometry.map((item) => (
              <tr key={item.token}>
                <td>{item.token}</td>
                <td>{item.value}</td>
                <td>{item.use}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Section>

      <Section id="motion" index="05" title="动效" note="HOVER 演示">
        <div className="motion-row">
          {motion.map((item) => (
            <div className="motion-card" key={item.token}>
              <div className="motion-demo">HOVER</div>
              <span className="motion-card__meta">
                {item.token} · {item.value}
              </span>
              <span className="motion-card__meta">{item.use}</span>
            </div>
          ))}
        </div>
      </Section>

      <Section id="rules" index="06" title="规则" note="三条就够">
        <div className="rule-list">
          <div className="rule">
            <span className="rule__num">01</span>
            <p className="rule__body">
              <strong>一个强调色。</strong> 橙色只出现在时间点、当前态和发布按钮；其余交互靠明度与线宽。
            </p>
          </div>
          <div className="rule">
            <span className="rule__num">02</span>
            <p className="rule__body">
              <strong>结构靠线和留白。</strong> 1px 发丝线定义层级；不用阴影、不用大圆角、不用渐变。
            </p>
          </div>
          <div className="rule">
            <span className="rule__num">03</span>
            <p className="rule__body">
              <strong>图片 8px，其余 0。</strong> 头像圆形是唯一例外；控件、卡片、输入框全部直角。
            </p>
          </div>
        </div>
      </Section>
    </div>
  )
}
