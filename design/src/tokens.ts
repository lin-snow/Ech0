// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/*
 * Machine-readable mirror of styles/tokens.css.
 * Used by the Foundations page; keep values aligned with the CSS file.
 */

export type ColorToken = {
  token: string
  light: string
  dark: string
  use: string
}

export const colorGroups: { title: string; note: string; colors: ColorToken[] }[] = [
  {
    title: 'Surface',
    note: '画布比卡片更暖；层级靠明度，不靠阴影。',
    colors: [
      { token: '--canvas', light: '#F4F1EC', dark: '#333333', use: '页面底色' },
      { token: '--surface', light: '#FFFFFF', dark: '#201D1B', use: '卡片 / 输入框' },
      { token: '--surface-2', light: '#FDF9F4', dark: '#292623', use: '悬浮层 / 次级卡片' },
      { token: '--muted', light: '#F0E9E0', dark: '#2B2724', use: '弱背景 / hover' },
      { token: '--muted-2', light: '#E6DDD2', dark: '#36312D', use: '弱分隔 / 热力图 1' },
    ],
  },
  {
    title: 'Ink',
    note: '正文只用 ink；ink-2 放摘要，ink-3 放元信息。',
    colors: [
      { token: '--ink', light: '#33251D', dark: '#F0E9E0', use: '正文 / 标题' },
      { token: '--ink-2', light: '#5B4F46', dark: '#CFCAC4', use: '次级正文' },
      { token: '--ink-3', light: '#8E847D', dark: '#A29E99', use: '时间 / 标签 / 提示' },
      { token: '--ink-inverse', light: '#FFFDFA', dark: '#201D1B', use: '强调色上的文字' },
    ],
  },
  {
    title: 'Line',
    note: '1px 发丝线；结构线用 line，交互边界用 line-strong。',
    colors: [
      { token: '--line', light: '#E4DFDB', dark: '#3E3C3A', use: '分隔 / 默认边框' },
      { token: '--line-strong', light: '#C2B9B2', dark: '#5E5A57', use: '输入框 / hover 边框' },
      { token: '--line-ink', light: '#33251D', dark: '#F0E9E0', use: '选中 / 2px 规则线' },
    ],
  },
  {
    title: 'Accent',
    note: '只有一个强调色：时间点、发布、当前态。',
    colors: [
      { token: '--accent', light: '#B84200', dark: '#E7B063', use: '主操作 / 当前态' },
      { token: '--accent-strong', light: '#93350A', dark: '#F2C88F', use: 'hover / active' },
    ],
  },
  {
    title: 'Heatmap',
    note: '5 级，从 surface 到 ink-2；不引入额外色相。',
    colors: [
      { token: '--heat-0', light: '#FFFFFF', dark: '#201D1B', use: '无记录' },
      { token: '--heat-1', light: '#E6DDD2', dark: '#36312D', use: '1 条' },
      { token: '--heat-2', light: '#C2B9B2', dark: '#5E5A57', use: '2–3 条' },
      { token: '--heat-3', light: '#8E847D', dark: '#A29E99', use: '4–6 条' },
      { token: '--heat-4', light: '#5B4F46', dark: '#CFCAC4', use: '7+ 条' },
    ],
  },
]

export const typeScale = [
  { token: '--fs-34', px: 34, line: 1.25, use: '页面标题' },
  { token: '--fs-26', px: 26, line: 1.25, use: '区块标题' },
  { token: '--fs-20', px: 20, line: 1.25, use: '卡片标题' },
  { token: '--fs-17', px: 17, line: 1.45, use: '强调正文' },
  { token: '--fs-15', px: 15, line: 1.45, use: '正文（宽松）' },
  { token: '--fs-14', px: 14, line: 1.7, use: '正文 / Echo' },
  { token: '--fs-13', px: 13, line: 1.45, use: '控件 / 标签' },
  { token: '--fs-12', px: 12, line: 1.45, use: '辅助文字' },
  { token: '--fs-11', px: 11, line: 1.45, use: '时间 / 元信息' },
  { token: '--fs-10', px: 10, line: 1.45, use: '全大写标签' },
]

export const spaceScale = [
  { token: '--sp-1', px: 4 },
  { token: '--sp-2', px: 8 },
  { token: '--sp-3', px: 12 },
  { token: '--sp-4', px: 16 },
  { token: '--sp-5', px: 20 },
  { token: '--sp-6', px: 24 },
  { token: '--sp-8', px: 32 },
  { token: '--sp-10', px: 40 },
  { token: '--sp-12', px: 48 },
  { token: '--sp-16', px: 64 },
]

export const geometry = [
  { token: '--control-h', value: '32px', use: '桌面控件高度' },
  { token: '--control-h-sm', value: '26px', use: '紧凑控件' },
  { token: '--r-0', value: '0', use: '按钮 / 卡片 / 输入框' },
  { token: '--r-img', value: '8px', use: '仅图片' },
  { token: '--rail-w', value: '400px', use: '侧栏' },
  { token: '--feed-w', value: '480px', use: '时间线' },
  { token: '--focus-w', value: '448px', use: '专注内容' },
  { token: '--mobile-w', value: '390px', use: '移动端画布' },
]

export const motion = [
  { token: '--dur-fast', value: '100ms', use: 'hover / 颜色' },
  { token: '--dur', value: '160ms', use: '展开 / 切换' },
  { token: '--dur-slow', value: '240ms', use: '面板 / 主题' },
  { token: '--ease-out', value: 'cubic-bezier(.22,1,.36,1)', use: '默认' },
  { token: '--ease-in-out', value: 'cubic-bezier(.65,0,.35,1)', use: '双向' },
]
