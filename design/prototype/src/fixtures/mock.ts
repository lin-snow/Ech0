// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/* Page 1 of memo.sn0w.fyi, rendered by the timeline. Content is raw markdown;
 * images are the post attachments downloaded locally. Nothing talks to the API. */

import avatar from '../assets/avatar.jpg'

export type EchoImage = {
  src: string
  width: number
  height: number
}

export type Echo = {
  id: string
  time: string
  stamp: string
  content: string
  images?: EchoImage[]
  tags: string[]
  isPrivate?: boolean
}

export const me = {
  name: 'L1nSn0w',
  handle: '@L1nSn0w',
  avatar,
}

export const echos: Echo[] = [
  {
    id: "e1",
    time: "昨天 20:45",
    stamp: "2026-10-25T20:45",
    content: "点了点了。别为了设计模式而设计模式，别为了微服务而拆微服务，也别为了消灭重复代码造出一堆抽象。\n\n~~是谁一会 Clean Architecture, 一会 DDD ?，原来是我😭...~~\n\n1. src: https://github.com/zakirullin/cognitive-load\n2. for agents(太贴心了😚): https://github.com/zakirullin/cognitive-load/blob/main/README.agents.md",
    images: [{ src: "/echoes/e1-1.png", width: 1200, height: 936 }],
    tags: ["叽哩咕噜", "思考"],
  },
  {
    id: "e2",
    time: "10 月 9 日",
    stamp: "2026-10-09T23:13",
    content: "",
    images: [{ src: "/echoes/e2-1.jpeg", width: 1190, height: 1200 }],
    tags: ["叽哩咕噜", "思考"],
  },
  {
    id: "e3",
    time: "10 月 2 日",
    stamp: "2026-10-02T23:38",
    content: "从道家角度看，你驾驭不住的东西才会显相，情绪驾驭不住，则怒气挂脸，财富驾驭不住，则铜臭外露，权力驾驭不住，则盛气凌人，所以，真正的贵人没有骄奢气，真正的高手没有刀剑气，真正的美人也没有造作姿，所谓大相无形，大音希声，只要内在全然驾驭，外在就会返璞归真，不留痕迹。",
    tags: ["叽哩咕噜", "思考"],
  },
  {
    id: "e4",
    time: "10 月 1 日",
    stamp: "2026-10-01T00:46",
    content: "金秋十月，盛世华章。值此中华人民共和国成立77周年之际，让我们共同祝愿伟大祖国繁荣昌盛、国泰民安！\n\n七十七载风雨兼程，七十七载砥砺奋进。今日之中国，山河壮丽、万象更新，人民生活蒸蒸日上，发展成就举世瞩目。\n\n愿我们珍惜盛世荣光，心怀家国情怀，坚定理想信念，以奋斗之姿投身时代洪流，以实际行动为国家发展贡献自己的力量。\n\n祝伟大祖国繁荣昌盛、山河永固！\n祝各位国庆快乐、平安顺遂！🇨🇳\n\n```js\n/* \n * 只需粘贴到「自定义 JS」，「自定义 CSS」留空即可；过期后整套装饰自动关闭。 */\n(() => {\n  const CONFIG = {\n    from: 930, // 生效起始日 MMDD（9 月 30 日）\n    to: 1008, // 生效结束日 MMDD（10 月 8 日），过期后自动关闭装饰与配套 CSS\n    always: false, // true = 无视日期，一直生效\n    salvo: 7, // 进站时礼花数量，0 = 不放\n    clickEmoji: '🇨🇳', // 点击页面时蹦出的 emoji，留空 '' = 关闭\n  }\n\n  window.__ech0NationalDay?.destroy()\n\n  const root = document.documentElement\n  const today = new Date()\n  const mmdd = (today.getMonth() + 1) * 100 + today.getDate()\n  if (!CONFIG.always && (mmdd < CONFIG.from || mmdd > CONFIG.to)) {\n    root.dataset.nd = 'off'\n    return\n  }\n  root.dataset.nd = 'on'\n\n  const reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches\n  const timers = new Set()\n  const later = (fn, ms) => {\n    const id = setTimeout(() => {\n      timers.delete(id)\n      fn()\n    }, ms)\n    timers.add(id)\n  }\n\n  const style = document.createElement('style')\n  style.textContent = `\n#nd-layer{position:fixed;inset:0;z-index:4;pointer-events:none;overflow:hidden}\n:root:not(:has(.home-page)) #nd-layer{display:none}\n#nd-layer canvas{position:absolute;inset:0;width:100%;height:100%}\n.nd-lantern{position:absolute;top:0;width:56px;transform-origin:50% 0;pointer-events:auto;cursor:pointer;\n  animation:nd-swing var(--nd-dur,4.2s) ease-in-out infinite alternate;filter:drop-shadow(0 6px 10px rgb(168 15 24 / 28%))}\n.nd-lantern::before{content:'';display:block;width:1.5px;height:var(--nd-rope,40px);margin:0 auto;background:linear-gradient(#a8741a,#e6b95a)}\n.nd-lantern svg{display:block;width:100%;height:auto}\n@keyframes nd-swing{from{transform:rotate(-3.5deg)}to{transform:rotate(3.5deg)}}\n.nd-pop{position:absolute;font-size:24px;line-height:1;pointer-events:none;user-select:none;will-change:transform,opacity;\n  font-family:'Apple Color Emoji','Segoe UI Emoji','Noto Color Emoji',sans-serif}\n@media (max-width:1023px){.nd-lantern{display:none}}\n@media (max-width:1279px){.nd-lantern--outer{display:none}}\n@media (prefers-reduced-motion:reduce){.nd-lantern{animation:none}}`\n  document.head.appendChild(style)\n\n  const themeStyle = document.createElement('style')\n  themeStyle.textContent = `\n/* ===== Ech0 · 国庆主题 (CSS) ===== */\nhtml:root:not([data-nd='off']) {\n  /* 中国红 + 金：亮色 / 晴日主题 */\n  &.light,\n  &.sunny {\n    --color-accent: #c8161d;\n    --color-accent-soft: #fbe9c6;\n    --color-focus-ring: rgb(200 22 29 / 28%);\n    --selection-bg-color: #c8161d;\n    --selection-text-color: #ffe9a8;\n    --md-tbody-hover-bg: rgb(200 22 29 / 8%);\n    --comment-input-focus-border: #d9605f;\n    --comment-input-focus-ring-inner: rgb(200 22 29 / 38%);\n    --comment-input-focus-ring-outer: rgb(200 22 29 / 14%);\n    --nd-title: linear-gradient(100deg, #c8161d 0%, #e0480f 55%, #d19200 100%);\n  }\n\n  /* 暗色主题：金色点缀 + 暗红底 */\n  &.dark {\n    --color-accent: #ffc94d;\n    --color-accent-soft: #5a2422;\n    --color-focus-ring: rgb(255 201 77 / 34%);\n    --selection-bg-color: #ffc94d;\n    --selection-text-color: #6b0f12;\n    --nd-title: linear-gradient(100deg, #ff7a5c 0%, #ffc94d 100%);\n  }\n\n  /* 站点标题：红金渐变 */\n  .home-header__title {\n    color: transparent;\n    background: var(--nd-title);\n    background-clip: text;\n    -webkit-background-clip: text;\n  }\n\n  .home-header__cursor {\n    background: #c8161d;\n  }\n\n  /* 头像：红金描边 + 右下角金星徽章 */\n  .home-header__logo-wrap {\n    position: relative;\n  }\n\n  .home-header__logo {\n    border-color: #ffd96a;\n    box-shadow:\n      0 0 0 2px #c8161d,\n      0 2px 8px rgb(200 22 29 / 32%);\n  }\n\n  .home-header__logo-wrap::after {\n    content: '★';\n    position: absolute;\n    right: -0.3rem;\n    bottom: -0.2rem;\n    width: 0.95rem;\n    height: 0.95rem;\n    border-radius: 50%;\n    background: #c8161d;\n    color: #ffd96a;\n    font-size: 0.58rem;\n    line-height: 0.95rem;\n    text-align: center;\n    box-shadow: 0 0 0 1.5px var(--color-bg-canvas);\n  }\n\n  /* 首页 Banner：红底金字贺卡 */\n  .home-banner {\n    position: relative;\n    overflow: hidden;\n    isolation: isolate;\n    border-radius: var(--radius-md);\n    background:\n      radial-gradient(circle at 12% 0%, rgb(255 217 106 / 22%), transparent 42%),\n      linear-gradient(135deg, #d9232d 0%, #bb131c 58%, #a20e16 100%);\n    box-shadow:\n      0 3px 8px -2px rgb(168 15 24 / 32%),\n      inset 0 0 0 1px rgb(255 217 106 / 30%);\n  }\n\n  .home-banner::before {\n    content: '国庆';\n    position: absolute;\n    z-index: -1;\n    right: 2.6rem;\n    bottom: 0.55rem;\n    font-family: 'Songti SC', STSong, 'Noto Serif SC', SimSun, serif;\n    font-size: 3.2rem;\n    font-weight: 900;\n    line-height: 1;\n    letter-spacing: 0.06em;\n    color: rgb(255 217 106 / 24%);\n    white-space: nowrap;\n    pointer-events: none;\n  }\n\n  .home-banner::after {\n    content: '';\n    position: absolute;\n    z-index: -1;\n    inset: 0;\n    background:\n      radial-gradient(circle, #ffd96a 0 1.2px, transparent 1.6px) 78% 22% / 100% 100% no-repeat,\n      radial-gradient(circle, #ffd96a 0 0.9px, transparent 1.3px) 64% 38% / 100% 100% no-repeat,\n      radial-gradient(circle, #ffd96a 0 1.4px, transparent 1.8px) 91% 44% / 100% 100% no-repeat,\n      radial-gradient(circle, #ffd96a 0 0.8px, transparent 1.2px) 52% 14% / 100% 100% no-repeat;\n    opacity: 0.75;\n    pointer-events: none;\n    animation: nd-twinkle 3.2s ease-in-out infinite alternate;\n  }\n\n  .home-banner__line {\n    color: #ffe9b0;\n    text-shadow: 0 1px 0 rgb(120 8 14 / 35%);\n  }\n\n  .home-banner__powered,\n  .home-banner__about {\n    color: rgb(255 233 176 / 78%);\n  }\n\n  .home-banner__powered:hover,\n  .home-banner__about:hover {\n    color: #fff6da;\n  }\n\n  /* 时间轴圆点 → 小红星 */\n  .timeline-marker > div {\n    flex-shrink: 0;\n    width: 0.7rem;\n    height: 0.7rem;\n    border-radius: 0;\n    clip-path: polygon(\n      50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%,\n      50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%\n    );\n  }\n}\n\n@keyframes nd-twinkle {\n  from {\n    opacity: 0.35;\n  }\n\n  to {\n    opacity: 0.9;\n  }\n}\n\n@media (prefers-reduced-motion: reduce) {\n  html:root .home-banner::after {\n    animation: none;\n  }\n}\n`\n  document.head.appendChild(themeStyle)\n\n  const layer = document.createElement('div')\n  layer.id = 'nd-layer'\n  layer.setAttribute('aria-hidden', 'true')\n  const canvas = document.createElement('canvas')\n  layer.appendChild(canvas)\n\n  const lanternSvg = (char, i) => `\n<svg viewBox=\"0 0 64 104\" xmlns=\"http://www.w3.org/2000/svg\">\n  <defs><radialGradient id=\"nd-g${i}\" cx=\"38%\" cy=\"34%\" r=\"75%\">\n    <stop offset=\"0\" stop-color=\"#ff6a55\"/><stop offset=\".55\" stop-color=\"#d9232d\"/><stop offset=\"1\" stop-color=\"#9c0d15\"/>\n  </radialGradient></defs>\n  <rect x=\"21\" y=\"0\" width=\"22\" height=\"7\" rx=\"2\" fill=\"#e6b95a\"/>\n  <ellipse cx=\"32\" cy=\"32\" rx=\"27\" ry=\"26\" fill=\"url(#nd-g${i})\"/>\n  <g fill=\"none\" stroke=\"#ffd96a\" stroke-opacity=\".38\" stroke-width=\"1\">\n    <ellipse cx=\"32\" cy=\"32\" rx=\"18\" ry=\"26\"/><ellipse cx=\"32\" cy=\"32\" rx=\"8\" ry=\"26\"/>\n  </g>\n  <rect x=\"21\" y=\"57\" width=\"22\" height=\"7\" rx=\"2\" fill=\"#e6b95a\"/>\n  <text x=\"32\" y=\"40.5\" text-anchor=\"middle\" font-size=\"23\" font-weight=\"900\" fill=\"#ffe08a\"\n    font-family=\"'Songti SC',STSong,'Noto Serif SC',SimSun,serif\">${char}</text>\n  <circle cx=\"32\" cy=\"69\" r=\"3\" fill=\"#d9232d\"/>\n  <g stroke=\"#d9232d\" stroke-width=\"1.6\" stroke-linecap=\"round\">\n    <path d=\"M32 72v28M28.5 73l-2 25M35.5 73l2 25M25.5 74l-3.5 20M38.5 74l3.5 20\"/>\n  </g>\n</svg>`\n\n  // [文字, 靠哪一侧, 距内容区外缘(px), 挂绳长度(px), 摆动周期(s), 是否外侧(窄屏隐藏)]\n  ;[\n    ['欢', 'left', 170, 28, 4.6, true],\n    ['度', 'left', 90, 72, 3.9, false],\n    ['国', 'right', 90, 72, 4.3, false],\n    ['庆', 'right', 170, 28, 5.0, true],\n  ].forEach(([char, side, gap, rope, dur, outer], i) => {\n    const el = document.createElement('div')\n    el.className = 'nd-lantern' + (outer ? ' nd-lantern--outer' : '')\n    el.style.cssText = `${side}:calc(50% - 400px - ${gap}px);--nd-rope:${rope}px;--nd-dur:${dur}s`\n    el.innerHTML = lanternSvg(char, i)\n    el.addEventListener('click', () => salvo(3))\n    layer.appendChild(el)\n  })\n  document.body.appendChild(layer)\n\n  /* ---------- 点击蹦出国旗 ---------- */\n  const onClick = (e) => {\n    if (!CONFIG.clickEmoji || reduceMotion || e.button !== 0) return\n    if (!document.querySelector('.home-page')) return\n    const pop = document.createElement('span')\n    pop.className = 'nd-pop'\n    pop.textContent = CONFIG.clickEmoji\n    pop.style.left = e.clientX + 'px'\n    pop.style.top = e.clientY + 'px'\n    layer.appendChild(pop)\n    const drift = (Math.random() - 0.5) * 50\n    const tilt = (Math.random() - 0.5) * 40\n    pop.animate(\n      [\n        { transform: 'translate(-50%, -50%) scale(0.3)', opacity: 0 },\n        { transform: `translate(-50%, -110%) scale(1.3) rotate(${tilt / 2}deg)`, opacity: 1, offset: 0.25 },\n        { opacity: 1, offset: 0.6 },\n        { transform: `translate(calc(-50% + ${drift}px), calc(-50% - 90px)) scale(1) rotate(${tilt}deg)`, opacity: 0 },\n      ],\n      { duration: 1000, easing: 'cubic-bezier(0.2, 0.8, 0.3, 1)' },\n    ).onfinish = () => pop.remove()\n  }\n  document.addEventListener('click', onClick)\n\n  /* ---------- 礼花 ---------- */\n  const ctx = canvas.getContext('2d')\n  const PALETTE = {\n    light: ['#d9232d', '#f0621a', '#e0a100', '#c8161d', '#f2b705'],\n    dark: ['#ff5a4d', '#ffc94d', '#ffe9a8', '#ff8a3d', '#ffffff'],\n  }\n  let W = 0\n  let H = 0\n  let rafId = 0\n  let last = 0\n  let rockets = []\n  let sparks = []\n\n  const resize = () => {\n    const dpr = Math.min(window.devicePixelRatio || 1, 2)\n    W = window.innerWidth\n    H = window.innerHeight\n    canvas.width = W * dpr\n    canvas.height = H * dpr\n    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)\n  }\n  resize()\n  window.addEventListener('resize', resize)\n\n  const rand = (a, b) => a + Math.random() * (b - a)\n  const pick = (arr) => arr[(Math.random() * arr.length) | 0]\n\n  const launch = () => {\n    const colors = PALETTE[root.classList.contains('dark') ? 'dark' : 'light']\n    rockets.push({\n      x: rand(W * 0.1, W * 0.9),\n      y: H,\n      targetY: rand(H * 0.12, H * 0.42),\n      speed: rand(9, 12),\n      color: pick(colors),\n      colors,\n    })\n    if (!rafId) {\n      last = performance.now()\n      rafId = requestAnimationFrame(frame)\n    }\n  }\n\n  const explode = (r) => {\n    const count = (rand(56, 84) * Math.min(1, W / 900 + 0.35)) | 0\n    const second = pick(r.colors)\n    for (let i = 0; i < count; i++) {\n      const angle = rand(0, Math.PI * 2)\n      const speed = rand(1.2, 5.2)\n      sparks.push({\n        x: r.x,\n        y: r.targetY,\n        vx: Math.cos(angle) * speed,\n        vy: Math.sin(angle) * speed,\n        life: 1,\n        decay: rand(0.011, 0.02),\n        size: rand(1.2, 2.4),\n        color: i % 3 === 0 ? second : r.color,\n      })\n    }\n  }\n\n  function frame(now) {\n    const k = Math.min((now - last) / 16.67, 3)\n    last = now\n\n    ctx.globalCompositeOperation = 'destination-out'\n    ctx.fillStyle = `rgba(0,0,0,${Math.min(0.22 * k, 1)})`\n    ctx.fillRect(0, 0, W, H)\n    ctx.globalCompositeOperation = 'source-over'\n\n    rockets = rockets.filter((r) => {\n      const prevY = r.y\n      r.y -= r.speed * k\n      ctx.globalAlpha = 0.85\n      ctx.strokeStyle = r.color\n      ctx.lineWidth = 2.2\n      ctx.lineCap = 'round'\n      ctx.beginPath()\n      ctx.moveTo(r.x, prevY)\n      ctx.lineTo(r.x, r.y)\n      ctx.stroke()\n      if (r.y > r.targetY) return true\n      explode(r)\n      return false\n    })\n\n    sparks = sparks.filter((s) => {\n      const drag = Math.pow(0.975, k)\n      s.vx *= drag\n      s.vy = s.vy * drag + 0.045 * k\n      s.x += s.vx * k\n      s.y += s.vy * k\n      s.life -= s.decay * k\n      if (s.life <= 0) return false\n      ctx.globalAlpha = Math.min(1, s.life * 1.4)\n      ctx.fillStyle = s.color\n      ctx.beginPath()\n      ctx.arc(s.x, s.y, s.size, 0, Math.PI * 2)\n      ctx.fill()\n      return true\n    })\n    ctx.globalAlpha = 1\n\n    if (rockets.length || sparks.length) {\n      rafId = requestAnimationFrame(frame)\n    } else {\n      rafId = 0\n      ctx.clearRect(0, 0, W, H)\n    }\n  }\n\n  function salvo(n) {\n    if (reduceMotion || document.hidden) return\n    for (let i = 0; i < n; i++) later(launch, i * rand(320, 620))\n  }\n\n  // 等首页渲染完成、加载动画消失后再放进站礼花（最多等 15 秒）\n  let tries = 0\n  const waitHome = () => {\n    const ready = document.querySelector('.home-page') && !root.classList.contains('app-loading')\n    if (ready) return later(() => salvo(CONFIG.salvo), 500)\n    if (++tries < 50) later(waitHome, 300)\n  }\n  waitHome()\n\n  window.__ech0NationalDay = {\n    salvo,\n    destroy() {\n      timers.forEach(clearTimeout)\n      timers.clear()\n      cancelAnimationFrame(rafId)\n      window.removeEventListener('resize', resize)\n      document.removeEventListener('click', onClick)\n      layer.remove()\n      style.remove()\n      themeStyle.remove()\n      delete root.dataset.nd\n      delete window.__ech0NationalDay\n    },\n  }\n})()\n```",
    tags: ["叽哩咕噜"],
  },
  {
    id: "e5",
    time: "9 月 28 日",
    stamp: "2026-09-28T11:32",
    content: "# Ech0 v5.8.0 现已发布！🚀\n\n本版本聚焦 **Copilot Agent 稳定性、检索能力与代码质量优化**。\n\nEch0 Copilot 现已调整为 **仅管理员可用**，并全面优化上下文预算、Token 成本与多轮 Tool Calling。新增 Temperature 配置，同时完善 Context Overflow 自动恢复、截断处理及多 Provider Tool Loop 兼容性，让 Agent 运行更加稳定、可控。\n\n同时强化 **Ech0 Retrieval**：Semantic Search 现在支持 Keyword Search Fallback，并进一步完善语义索引与 Echos 数据的一致性处理，提升检索在异常场景下的可靠性。\n\n在工程层面，本版本升级至 **Go 1.27.1**，引入泛型 Job / Cache，并持续进行 **代码质量、架构与依赖更新**，进一步提升类型安全、可维护性与整体工程质量。\n\n欢迎更新体验！🥳",
    tags: ["更新日志"],
  },
  {
    id: "e6",
    time: "9 月 26 日",
    stamp: "2026-09-26T11:51",
    content: "我的腾讯云服务器将于明天到期，现已将我的 Ech0 部署到了我的一台海外的 VPS 上，目前效果还不错（国内访问的话确实还有些延迟），域名也正式取消了重定向，以后可以直接访问 [memo.sn0w.fyi ](https://memo.sn0w.fyi ) 了，按照之前发过的省钱计划，我将不会继续对国内持有的云服务资源和域名进行续费\n\n我的部署配置如下：\n- VPS: 1h 1g US\n- 形式：基于 Ech0 提供的在线脚本部署裸二进制文件\n- 反向代理和证书：使用最新版 Caddy 提供反向代理与证书自动续签\n\n",
    tags: ["叽哩咕噜"],
  },
  {
    id: "e7",
    time: "9 月 23 日",
    stamp: "2026-09-23T12:11",
    content: "Add this prompt to your Coding Agent to make it write only necessary comments, reducing conflicts and drift between comments and code while also saving output tokens.\n\n```txt\n## Code Comments\n\nKeep comments to an absolute minimum.\n\n* **Default: write no comments.**\n* Do not explain what the code does.\n* Do not explain how the code works.\n* Do not add comments for ordinary or obvious \"why\".\n* Do not narrate implementation steps, control flow, API calls, error handling, or business logic.\n* Prefer clear naming and straightforward code over comments.\n* Only add a comment when the reason is **non-obvious, important, and difficult to infer from the code itself**.\n* Such comments should explain a critical constraint, invariant, workaround, compatibility requirement, or non-obvious design decision.\n* If removing the comment would not materially hurt future maintenance, **do not write it**.\n* Never add comments merely to make the code easier to read.\n\n**When in doubt, do not write a comment.**\n\nThe desired code style is **self-explanatory code with almost no comments**.\n```",
    tags: ["摘抄"],
  },
]

export const tagCloud = ["叽哩咕噜", "思考", "更新日志", "摘抄"]


/* Public /api/echo/page: 48 days (2026-09-08 → 2026-10-25), binned 0–4. */
export const heatmapEnd = '2026-10-25'
export const heatmap: number[] = [
  0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1
]

export const connects = [
  { name: '唤青拾句', today: 0, logo: '/connects/c1.png' },
  { name: 'Jinvic', today: 0, logo: '/connects/c2.png' },
  { name: '葱苓のEch0', today: 0, logo: '/connects/c3.png' },
  { name: "XiaoTen's Echo", today: 0, logo: '/connects/c4.png' },
  { name: "iTiki's Zone", today: 0, logo: '/connects/c5.png' },
  { name: '清墨羽痕', today: 0, logo: '/connects/c6.png' },
  { name: "Zoe's Red Pill", today: 0, logo: '/connects/c7.png' },
  { name: 'Sean', today: 0, logo: '/connects/c8.png' },
  { name: 'Ech0 - Elvish', today: 1, logo: '/connects/c9.png' },
]
