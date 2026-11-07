// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { describe, expect, it } from 'vitest'
import { needsHighlight, renderMarkdown, renderMarkdownSync } from '../../src/editor/core/markdown'

describe('renderMarkdownSync', () => {
  it('同步渲染与异步渲染产出相同 HTML，首个布局帧即为最终行盒', async () => {
    const source = '第一行\n\n第二行 [站点](https://ech0.cc)'

    const syncHtml = renderMarkdownSync(source)
    const asyncHtml = await renderMarkdown(source)

    expect(syncHtml).toBe(asyncHtml)
  })

  it('代码块同步渲染即为完整代码块，不会退回纯文本占位', () => {
    const source = ['```ts', 'const answer = 42', '```'].join('\n')

    expect(needsHighlight(source)).toBe(true)
    expect(needsHighlight('普通段落')).toBe(false)

    const html = renderMarkdownSync(source)

    expect(html).toContain('code-block')
    expect(html).toContain('const answer = 42')
    expect(html).not.toContain('```')
  })
})
