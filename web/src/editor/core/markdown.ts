// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import markdownit, { type MarkdownIt } from 'markdown-it'

type HighlightJsModule = (typeof import('./markdown-highlight'))['default']

let hljsInstance: HighlightJsModule | null = null
let hljsPromise: Promise<HighlightJsModule> | null = null

async function ensureHighlighter(): Promise<HighlightJsModule> {
  if (hljsInstance) return hljsInstance
  if (!hljsPromise) {
    hljsPromise = import('./markdown-highlight').then((mod) => {
      hljsInstance = mod.default
      return mod.default
    })
  }
  return hljsPromise
}

const CODE_BLOCK_COLLAPSE_THRESHOLD = 18
const CODE_BLOCK_COLLAPSED_LINES = 10
const EXPAND_PLACEHOLDER = '__ECHO_MD_EXPAND__'
const COLLAPSE_PLACEHOLDER = '__ECHO_MD_COLLAPSE__'
const COPY_PLACEHOLDER = '__ECHO_MD_COPY__'
const COPIED_PLACEHOLDER = '__ECHO_MD_COPIED__'
const TASK_CHECKBOX_LABEL_PLACEHOLDER = '__ECHO_MD_TASK_CHECKBOX_LABEL__'
const TASK_LIST_MARKER_RE = /^\[( |x|X)\]\s+/
const CODE_FENCE_RE = /(^|\n)\s{0,3}(```|~~~)/
const RENDER_CACHE_MAX_SIZE = 120
const renderCache = new Map<string, string>()

function escapeHtml(input: string): string {
  return input
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

function getLineCount(code: string): number {
  if (!code) return 0
  return code.split(/\r?\n/).length
}

function renderCodeBlock(code: string, rendered: string, language?: string): string {
  const lineCount = getLineCount(code)
  const isCollapsible = lineCount >= CODE_BLOCK_COLLAPSE_THRESHOLD
  const langRaw = language?.trim() || 'text'
  const languageClass = language ? ` language-${language}` : ''
  const pre = `<pre><code class="hljs${languageClass}">${rendered}</code></pre>`

  const langLabel = escapeHtml(langRaw)
  const header =
    `<div class="code-block__header">` +
    `<span class="code-block__lang">${langLabel}</span>` +
    `<div class="code-block__actions">` +
    (isCollapsible
      ? `<button type="button" class="code-block-toggle" data-expand-label="${EXPAND_PLACEHOLDER}" data-collapse-label="${COLLAPSE_PLACEHOLDER}" aria-expanded="false">${EXPAND_PLACEHOLDER}</button>`
      : '') +
    `<button type="button" class="code-block-copy" data-copy-label="${COPY_PLACEHOLDER}" data-copied-label="${COPIED_PLACEHOLDER}">${COPY_PLACEHOLDER}</button>` +
    `</div>` +
    `</div>`

  const classes = ['code-block']
  const styleAttr = isCollapsible ? ` style="--code-max-lines:${CODE_BLOCK_COLLAPSED_LINES};"` : ''
  if (isCollapsible) {
    classes.push('code-block--collapsible', 'code-block--collapsed')
  }

  return `<div class="${classes.join(' ')}"${styleAttr}>${header}<div class="code-block__body">${pre}</div></div>`
}

function appendClass(
  token: { attrJoin: (name: string, value: string) => void },
  className: string,
): void {
  token.attrJoin('class', className)
}

function taskListPlugin(md: MarkdownIt): void {
  md.core.ruler.after('inline', 'echo_task_list', (state) => {
    const { tokens } = state

    for (let i = 0; i < tokens.length; i += 1) {
      const listItemToken = tokens[i]
      if (listItemToken.type !== 'list_item_open') continue

      let inlineTokenIndex = -1
      for (let j = i + 1; j < tokens.length; j += 1) {
        const token = tokens[j]
        if (token.type === 'list_item_close') break
        if (token.type === 'inline') {
          inlineTokenIndex = j
          break
        }
      }
      if (inlineTokenIndex < 0) continue

      const inlineToken = tokens[inlineTokenIndex]
      const firstChild = inlineToken.children?.[0]
      if (!firstChild || firstChild.type !== 'text') continue

      const markerMatch = firstChild.content.match(TASK_LIST_MARKER_RE)
      if (!markerMatch) continue

      const checked = markerMatch[1].toLowerCase() === 'x'
      firstChild.content = firstChild.content.slice(markerMatch[0].length)
      inlineToken.content = inlineToken.content.replace(TASK_LIST_MARKER_RE, '')

      appendClass(listItemToken, 'task-list-item')

      const checkbox = new state.Token('html_inline', '', 0)
      checkbox.content = `<input class="task-list-item-checkbox" type="checkbox" aria-label="${TASK_CHECKBOX_LABEL_PLACEHOLDER}" disabled${checked ? ' checked' : ''}> `
      inlineToken.children?.unshift(checkbox)
    }
  })
}

function getFromRenderCache(key: string): string | undefined {
  const value = renderCache.get(key)
  if (!value) return undefined
  renderCache.delete(key)
  renderCache.set(key, value)
  return value
}

function setRenderCache(key: string, value: string): void {
  if (renderCache.has(key)) {
    renderCache.delete(key)
  } else if (renderCache.size >= RENDER_CACHE_MAX_SIZE) {
    const oldest = renderCache.keys().next().value
    if (oldest) renderCache.delete(oldest)
  }
  renderCache.set(key, value)
}

const markdown: MarkdownIt = markdownit({
  html: false,
  linkify: true,
  typographer: false,
  langPrefix: 'language-',
  highlight(str: string, lang: string): string {
    const language = lang?.trim()
    if (hljsInstance && language && hljsInstance.getLanguage(language)) {
      try {
        const rendered = hljsInstance.highlight(str, {
          language,
          ignoreIllegals: true,
        }).value
        return renderCodeBlock(str, rendered, language)
      } catch {}
    }

    return renderCodeBlock(str, escapeHtml(str), language)
  },
})
markdown.use(taskListPlugin)

type LinkOpenRule = NonNullable<MarkdownIt['renderer']['rules']['link_open']>

const originalLinkOpen: LinkOpenRule =
  markdown.renderer.rules.link_open ??
  ((tokens, idx, options, _env, self) => self.renderToken(tokens, idx, options))

markdown.renderer.rules.link_open = (...args: Parameters<LinkOpenRule>) => {
  const [tokens, idx, options, env, self] = args
  const token = tokens[idx]
  const href = String(token.attrGet('href') ?? '')

  if (/^https?:\/\//i.test(href)) {
    token.attrSet('target', '_blank')
    token.attrSet('rel', 'noopener noreferrer')
  }

  return originalLinkOpen(tokens, idx, options, env, self)
}

export type MarkdownLabels = {
  expandLabel?: string
  collapseLabel?: string
  copyLabel?: string
  copiedLabel?: string
  taskCheckboxLabel?: string
}

type ResolvedLabels = {
  cacheKey: string
  expandLabel: string
  collapseLabel: string
  copyLabel: string
  copiedLabel: string
  taskCheckboxLabel: string
}

function resolveLabels(source: string, labels?: MarkdownLabels): ResolvedLabels {
  const expandLabel = escapeHtml(String(labels?.expandLabel || 'Expand'))
  const collapseLabel = escapeHtml(String(labels?.collapseLabel || 'Collapse'))
  const copyLabel = escapeHtml(String(labels?.copyLabel || 'Copy'))
  const copiedLabel = escapeHtml(String(labels?.copiedLabel || 'Copied'))
  const taskCheckboxLabel = escapeHtml(String(labels?.taskCheckboxLabel || 'Task item'))
  return {
    expandLabel,
    collapseLabel,
    copyLabel,
    copiedLabel,
    taskCheckboxLabel,
    cacheKey: `${source}\u241f${expandLabel}\u241f${collapseLabel}\u241f${copyLabel}\u241f${copiedLabel}\u241f${taskCheckboxLabel}`,
  }
}

function renderHtml(source: string, labels: ResolvedLabels): string {
  return markdown
    .render(source)
    .replaceAll(EXPAND_PLACEHOLDER, labels.expandLabel)
    .replaceAll(COLLAPSE_PLACEHOLDER, labels.collapseLabel)
    .replaceAll(COPY_PLACEHOLDER, labels.copyLabel)
    .replaceAll(COPIED_PLACEHOLDER, labels.copiedLabel)
    .replaceAll(TASK_CHECKBOX_LABEL_PLACEHOLDER, labels.taskCheckboxLabel)
}

/** Whether the source contains fenced code that wants the lazy highlighter. */
export function needsHighlight(source: string): boolean {
  return CODE_FENCE_RE.test(source)
}

export function isHighlighterReady(): boolean {
  return hljsInstance !== null
}

/**
 * Renders without waiting for the lazy highlighter chunk. Markdown layout is
 * identical with or without highlighting (only spans change), so callers can
 * paint the final line boxes immediately and swap in the coloured HTML later.
 * Un-highlighted fenced code is deliberately not cached so the async pass can
 * still fill the shared cache with the highlighted version.
 */
export function renderMarkdownSync(source: string, labels?: MarkdownLabels): string {
  if (!source) return ''
  const resolved = resolveLabels(source, labels)
  const cached = getFromRenderCache(resolved.cacheKey)
  if (cached) return cached

  const html = renderHtml(source, resolved)
  if (!needsHighlight(source) || hljsInstance) {
    setRenderCache(resolved.cacheKey, html)
  }
  return html
}

export async function renderMarkdown(source: string, labels?: MarkdownLabels): Promise<string> {
  if (!source) return ''
  const resolved = resolveLabels(source, labels)
  const cached = getFromRenderCache(resolved.cacheKey)
  if (cached) return cached

  if (needsHighlight(source)) {
    await ensureHighlighter()
  }

  const html = renderHtml(source, resolved)
  setRenderCache(resolved.cacheKey, html)
  return html
}
