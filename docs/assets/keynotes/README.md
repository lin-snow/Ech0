# 演讲稿（Keynote）

> 素材产线，不是文档。文档索引见 [`docs/README.md`](../../README.md)。

用 HTML 写的演讲幻灯片：一份 deck 一个目录，共用 [`deck.css`](deck.css) 与 [`deck.js`](deck.js)，**无构建步骤、无依赖、无 lockfile**。双击 `index.html` 就能讲，Chrome 打印即得会议要的 PDF。

## 目录

```
keynotes/
├── deck.css            版式与主题（1600×900 逻辑画布，自动缩放）
├── deck.js             翻页 / 深链 / 备注 / 主题 / 全屏
├── _template/          所有版式的范例，复制它开始写
├── deck/               参考实现：Ech0 发布会 deck（暗场，14 页，含 PDF）
└── <slug>/index.html   一份 deck
```

`_template/` 同时是版式清单：封面、要点列表、两栏卡片、代码页、大字引语。复制目录、删掉不要的页、改文案即可。

`deck/` 是完整做出来的一份：`script.md`（**内容稿，先改这里**）+ `index.html`（14 页）+ `style.css`（各页版式）+ `ech0-deck.pdf`（打印产物）+ `shots/`（逐页截图）+ `contact.png`（14 页缩略图，一眼看完）+ `direction-approved.md`（为什么是暗场）+ `polish-log.md`（逐轮改了什么）。新 deck 想做发布会那种语气，从它抄；想要中性版式，从 `_template/` 抄。

## 写一份新的

```bash
cp -r docs/assets/keynotes/_template docs/assets/keynotes/<slug>
open docs/assets/keynotes/<slug>/index.html
```

一页 = 一个 `<section class="slide">`，正文直接写 HTML。备注写在该页末尾，屏幕上不出现，也不进 PDF：

```html
<section class="slide center">
  <p class="kicker">小标题</p>
  <h2>标题</h2>
  <ul class="list"><li>要点</li></ul>
  <aside class="notes">只有讲者看得到的话。</aside>
</section>
```

## 版式

| 类 | 作用 |
|---|---|
| `.slide` | 一页，默认顶部对齐、内边距 96px |
| `.slide.center` | 内容垂直居中 |
| `.slide.cover` | 封面/结尾：大标题 + 副标题 + 元信息，垂直居中 |
| `.kicker` | 等宽大写小标签，用强调色，放标题之上 |
| `h1` / `h2` / `h3` | 92 / 56 / 32px |
| `.lede` | 44px 陈述句，用于大字引语 |
| `.body`、`p` | 30px 正文 |
| `.meta` | 等宽 22px 元信息（链接、副标题、脚注） |
| `.list`、`.list.tight` | 带强调色横线的要点列表；`tight` 用于条目多的页 |
| `.cols`、`.cols.asym` | 两栏网格；`asym` 是 1.15 : 0.85 |
| `.card` | 白底 + 1px 发丝线的卡片，用于两栏 |
| `pre > code` | 代码块；行内高亮用 `<span class="k">`（关键字）与 `<span class="c">`（注释） |
| `.rule` | 1px 分隔线 |
| `.spacer` | 把后续内容压到底部 |

配色与字体取自 `design/src/styles/tokens.css`（暖纸色、发丝线、单一焦橙强调色）。主题跟随系统，按 `t` 可手动切换；不要在页面里写死颜色，用 `var(--accent)` 这类变量。

要一份固定暗场（发布会那种黑底），在本地 `style.css` 顶部把 `--canvas` / `--ink` / `--ink-2` / `--line` / `--accent` 同时写进 `:root`、`:root[data-theme='dark']`、`:root[data-theme='light']`，并把 `<html>` 写成 `data-theme="dark"`——`deck.css` 的暗色块靠 `prefers-color-scheme` 生效，不锁的话别人机器上是亮场。写法见 `deck/style.css` 顶部。

## 操作

| 键 | 动作 |
|---|---|
| `→` `↓` `Space` `PageDown` | 下一页 |
| `←` `↑` `PageUp` | 上一页 |
| `Home` / `End` | 首页 / 末页 |
| 数字 + `Enter` | 跳到第 N 页（Q&A 时用） |
| `f` | 全屏 |
| `n` | 显示/隐藏备注 |
| `t` | 切换明暗主题 |
| `Esc` | 收起备注 |

URL 的 `#<n>` 就是第几页，可以直接把某页发出去。触屏左右滑动也可翻页。

## 导出 PDF

Chrome 打开 `index.html` → 打印 → 目标「另存为 PDF」→ 边距「无」→ 背景图形**勾选**。

`@page` 已固定为 1600×900，每页一张幻灯片、无分页误差；备注、进度条、页码在打印样式里被移除。

## 为什么不用 Slidev / Marp

仓库里 `docs/assets/architecture/`、`carousel/`、`social-preview/`，以及 `site/public/arch/` 都是零依赖单文件 HTML，keynote 是同一体裁的产物，不值得为它引入第二套前端工具链。需要深链、备注、全屏、PDF 的，`deck.js` 168 行已经给全；真正把版式做复杂时，最终还是写 HTML，中间那层只会挡路。

需要公开讲稿时，把成品目录拷进 `site/public/keynotes/<slug>/` 随站点发布（先例：`site/public/arch/index.html` 服务于 `ech0.app/arch`）；这一步不是默认动作。

## 发布到 slides.ech0.sn0w.fyi

这个目录同时是一个 Cloudflare Worker 的静态资源根（[`wrangler.jsonc`](wrangler.jsonc)）：没有 Worker 脚本、没有构建、没有 `package.json`，`wrangler` 把整个目录原样上传，`routes[].custom_domain` 负责 DNS 与证书。

**改完 deck 只需合并到 `main`。** [`.github/workflows/deploy-slides.yml`](../../../.github/workflows/deploy-slides.yml) 在 `docs/assets/keynotes/**` 有改动时自动部署，用的是仓库已有的 `CLOUDFLARE_API_TOKEN` / `CLOUDFLARE_ACCOUNT_ID`（与 `deploy-hub.yml` 同一组 secrets）。

本地要预览同一份产物时，跑同一条命令：

```bash
npx wrangler deploy --config docs/assets/keynotes/wrangler.jsonc
```

- 站点根 `index.html` 是讲稿索引，`/deck/` 是发布会 deck，`/deck/ech0-deck.pdf` 是打印产物。
- `.assetsignore` 把 `*.md` 与配置本身排除在上传之外——讲稿正文留在仓库里，不上网。
- 首次部署时 `slides.ech0.sn0w.fyi` 会自动建 DNS 记录；证书由 Cloudflare 签发，深层子域需要等几分钟。
- 改了 deck 只需重新跑一次上面那条命令；`contact.png` / `shots/` 会一并更新。
