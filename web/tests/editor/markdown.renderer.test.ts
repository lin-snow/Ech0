// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import { renderMarkdown } from '../../src/editor/core/markdown'

function lines(count: number): string {
  return Array.from({ length: count }, (_, i) => `line_${i + 1}`).join('\n')
}

describe('renderMarkdown renderer behaviors', () => {
  it('为外链添加 target 与 rel 属性', async () => {
    const html = await renderMarkdown('[echo](https://example.com)')

    expect(html).toContain('target="_blank"')
    expect(html).toContain('rel="noopener noreferrer"')
  })

  it('仅折叠超过阈值的代码块并替换按钮文案', async () => {
    const source = ['```ts', lines(18), '```'].join('\n')
    const html = await renderMarkdown(source, {
      expandLabel: '展开<更多>',
      collapseLabel: '收起<更少>',
    })

    expect(html).toContain('code-block--collapsible')
    expect(html).toContain('code-block-toggle')
    expect(html).toContain('展开&lt;更多&gt;')
    expect(html).toContain('收起&lt;更少&gt;')
  })

  it('对原始 HTML 输入保持转义，避免脚本注入', async () => {
    const html = await renderMarkdown('<script>alert("xss")</script>')

    expect(html).toContain('&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;')
    expect(html).not.toContain('<script>')
  })

  it('缓存按内容与标签区分，避免错误复用本地化文案', async () => {
    const source = ['```ts', lines(18), '```'].join('\n')
    const htmlA = await renderMarkdown(source, { expandLabel: '展开A', collapseLabel: '收起A' })
    const htmlB = await renderMarkdown(source, { expandLabel: '展开B', collapseLabel: '收起B' })

    expect(htmlA).toContain('展开A')
    expect(htmlA).not.toContain('展开B')
    expect(htmlB).toContain('展开B')
    expect(htmlB).not.toContain('展开A')
  })

  it('自动成链只覆盖带协议的地址，裸域名保持纯文本', async () => {
    const html = await renderMarkdown('看 https://ech0.cc 和 ech0.cc 与 [站点](https://ech0.cc)')

    expect(html).toContain('<a href="https://ech0.cc"')
    expect(html).toContain('>站点</a>')
    expect(html).toContain('和 ech0.cc 与')
  })
})
