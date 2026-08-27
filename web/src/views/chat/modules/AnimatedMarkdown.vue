<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script lang="ts">
import { defineComponent, h, computed, watch, type VNode, type PropType } from 'vue'
import markdownit, { type Token } from 'markdown-it'
import DiffText from './DiffText.vue'
import { useSmoothReveal } from './useSmoothReveal'
import '@/editor/styles/markdown.scss'

type Child = VNode | string

const md = markdownit({
  html: false,
  linkify: true,
  typographer: false,
  langPrefix: 'language-',
})

const ANIM_MAP: Record<string, string> = {
  blurIn: 'am-blur-in',
  fadeIn: 'am-fade-in',
}

interface BuildCtx {
  k: () => number
  animate: boolean
  animationClass: string
  duration: number
}

function textLeaf(text: string, ctx: BuildCtx, key: string): VNode {
  return h(DiffText, {
    key,
    text,
    animationClass: ctx.animationClass,
    duration: ctx.duration,
    animate: ctx.animate,
  })
}

function linkAttrs(token: Token): Record<string, string> {
  const href = String(token.attrGet('href') ?? '')
  const attrs: Record<string, string> = { href }
  if (/^https?:\/\//i.test(href)) {
    attrs.target = '_blank'
    attrs.rel = 'noopener noreferrer'
  }
  return attrs
}

interface Frame {
  tag: string
  attrs: Record<string, unknown>
  children: Child[]
  key: string
}

const blockKey = (token: Token, ctx: BuildCtx): string =>
  token.map ? `b${token.map[0]}` : `bx${ctx.k()}`

function renderInline(children: Token[], ctx: BuildCtx, baseLine: number): Child[] {
  const out: Child[] = []
  const stack: Frame[] = []
  const sink = (): Child[] => (stack.length ? stack[stack.length - 1].children : out)
  let seq = 0
  const nk = (): string => `${baseLine}:${seq++}`

  for (const t of children) {
    switch (t.type) {
      case 'text':
        if (t.content) sink().push(textLeaf(t.content, ctx, nk()))
        break
      case 'softbreak':
        sink().push(' ')
        break
      case 'hardbreak':
        sink().push(h('br', { key: nk() }))
        break
      case 'code_inline':
        sink().push(h('code', { key: nk(), class: 'am-code-inline' }, t.content))
        break
      case 'image':
        sink().push(
          h('img', {
            key: nk(),
            src: String(t.attrGet('src') ?? ''),
            alt: t.content,
            loading: 'lazy',
          }),
        )
        break
      case 'link_open':
        stack.push({ tag: 'a', attrs: linkAttrs(t), children: [], key: nk() })
        break
      case 'strong_open':
      case 'em_open':
      case 's_open':
        stack.push({ tag: t.tag, attrs: {}, children: [], key: nk() })
        break
      case 'link_close':
      case 'strong_close':
      case 'em_close':
      case 's_close': {
        const frame = stack.pop()
        if (frame) sink().push(h(frame.tag, { ...frame.attrs, key: frame.key }, frame.children))
        break
      }
      default:
        if (t.content) sink().push(textLeaf(t.content, ctx, nk()))
    }
  }
  while (stack.length) {
    const frame = stack.shift()
    if (frame) out.push(h(frame.tag, { ...frame.attrs, key: frame.key }, frame.children))
  }
  return out
}

function renderCode(token: Token, ctx: BuildCtx): VNode {
  const lang = token.info?.trim().split(/\s+/)[0]
  const codeClass = lang ? `hljs language-${lang}` : 'hljs'
  return h('pre', { key: blockKey(token, ctx), class: 'am-pre' }, [
    h('code', { class: codeClass }, token.content),
  ])
}

function renderBlocks(tokens: Token[], ctx: BuildCtx): Child[] {
  const out: Child[] = []
  const stack: Frame[] = []
  const sink = (): Child[] => (stack.length ? stack[stack.length - 1].children : out)

  for (const t of tokens) {
    if (t.type === 'inline') {
      const baseLine = t.map ? t.map[0] : ctx.k()
      for (const node of renderInline(t.children ?? [], ctx, baseLine)) sink().push(node)
      continue
    }
    if (t.type === 'fence' || t.type === 'code_block') {
      sink().push(renderCode(t, ctx))
      continue
    }
    if (t.type === 'hr') {
      sink().push(h('hr', { key: blockKey(t, ctx) }))
      continue
    }
    if (t.nesting === 1) {
      stack.push({ tag: t.tag || 'div', attrs: {}, children: [], key: blockKey(t, ctx) })
    } else if (t.nesting === -1) {
      const frame = stack.pop()
      if (frame) sink().push(h(frame.tag, { ...frame.attrs, key: frame.key }, frame.children))
    } else if (t.content) {
      sink().push(textLeaf(t.content, ctx, blockKey(t, ctx)))
    }
  }
  while (stack.length) {
    const frame = stack.pop()
    if (frame) sink().push(h(frame.tag, { ...frame.attrs, key: frame.key }, frame.children))
  }
  return out
}

const fenceBalanced = (s: string): boolean => {
  const m = s.match(/^ {0,3}(`{3,}|~{3,})/gm)
  return m === null || m.length % 2 === 0
}

const tailMayJoinPrefix = (tail: string): boolean => {
  for (const line of tail.split('\n')) {
    if (line.trim().length === 0) continue
    return (
      /^ {0,3}([-*+]|\d{1,9}[.)])(\s|$)/.test(line) ||
      /^ {0,3}>/.test(line) ||
      /^(?: {4,}|\t)/.test(line)
    )
  }
  return false
}

const offsetTokenMap = (t: Token, off: number): void => {
  if (t.map) t.map = [t.map[0] + off, t.map[1] + off]
}

function createIncrementalParser(): (text: string) => Token[] {
  let cachedPrefix = ''
  let cachedTokens: Token[] = []

  return (text: string): Token[] => {
    const lines = text.split('\n')
    let bi = -1
    for (let i = lines.length - 1; i >= 0; i -= 1) {
      if (lines[i].trim().length === 0) {
        bi = i
        break
      }
    }
    if (bi <= 0) return md.parse(text, {})

    const prefix = lines.slice(0, bi).join('\n') + '\n'
    const tail = lines.slice(bi).join('\n')
    if (!fenceBalanced(prefix) || tailMayJoinPrefix(tail)) return md.parse(text, {})

    if (prefix !== cachedPrefix) {
      cachedPrefix = prefix
      cachedTokens = md.parse(prefix, {})
    }
    const tailTokens = md.parse(tail, {})
    for (const t of tailTokens) offsetTokenMap(t, bi)
    return [...cachedTokens, ...tailTokens]
  }
}

export default defineComponent({
  name: 'AnimatedMarkdown',
  props: {
    content: { type: String, default: '' },
    animation: { type: String as PropType<keyof typeof ANIM_MAP>, default: 'blurIn' },
    duration: { type: Number, default: 600 },
    streaming: { type: Boolean, default: true },
  },
  emits: ['update:revealing'],
  setup(props, { emit }) {
    const reduceMotion =
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches

    const source = computed(() => props.content)
    const streamingRef = computed(() => props.streaming && !reduceMotion)
    const displayed = useSmoothReveal(source, streamingRef)

    const parseStream = createIncrementalParser()

    watch(
      [displayed, () => props.content],
      () =>
        emit('update:revealing', !reduceMotion && displayed.value.length < props.content.length),
      { immediate: true },
    )

    return () => {
      const text = reduceMotion ? props.content : displayed.value
      let counter = 0
      const ctx: BuildCtx = {
        k: () => counter++,
        animate: !reduceMotion,
        animationClass: ANIM_MAP[props.animation] ?? ANIM_MAP.fadeIn,
        duration: props.duration,
      }
      const tree = renderBlocks(parseStream(text || ''), ctx)
      return h('div', { class: ['echo-markdown', 'am-root'] }, tree)
    }
  },
})
</script>
