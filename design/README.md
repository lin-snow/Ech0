# Ech0 Design

Ech0 的设计资产库。全部在 `main` 工作区维护，不单独开设计分支。

```
design/
├── src/                 # 设计系统预览（tokens / components / patterns / UX）
├── prototype/           # 可交互 UI 草图（Base UI + 真实 markdown 数据）
├── archive/             # OpenPencil 源文件与旧版截图
├── index.html           # 设计系统入口
├── package.json         # 两个应用的脚本与依赖
└── vite.config.ts       # 设计系统构建配置
```

## 两个应用

| 应用 | 入口 | 用途 |
| --- | --- | --- |
| Design System | `pnpm dev` → `:5180` | 规范、组件标本、布局模式、IA/流程/状态 |
| Prototype | `pnpm dev:prototype` → `:5174` | 真实交互的首页草图，验证布局与写作流程 |

Prototype 的 `public/` 同时供设计系统使用（`vite.config.ts` 的 `publicDir`），图片只存一份。

## 命令

```bash
cd design
pnpm install

pnpm dev              # 设计系统 :5180
pnpm dev:prototype    # 交互原型 :5174

pnpm typecheck        # 两个应用都做 TS 检查
pnpm build            # 设计系统 -> design/dist
pnpm build:prototype  # 原型 -> design/prototype/dist
```

也可以从仓库根目录：`just design dev` / `just design dev-prototype`。

构建产物 `dist/`、`prototype/dist/` 与 `node_modules/` 不提交。

## 设计系统

- 文档外壳（导航、首页、页头、section 标题）使用中性灰 `--doc-*` token；只有具体资产标本使用 Ech0 的 `--canvas/--ink/--accent`。LIGHT/DARK 只切资产预览，不切外壳。
- `src/styles/tokens.css` — 唯一 token 源：颜色、字号、间距、几何、动效；亮/暗双主题。
- `src/tokens.ts` — 给 Foundations 页面用的机器可读镜像；改 CSS 时同步这里。
- `src/fixtures/mock.ts` — 真实 echo、热力图、标签与账户数据。
- `src/components/` — Button、IconButton、SearchField、Tag、Tabs、Composer、EchoEntry、ActivityHeatmap、AccountRow、EmptyState。
- `src/patterns/` — 侧栏 400 + 480、专注 448、移动 390 三个真实布局。
- `src/pages/Home.tsx` — 资产库首页；卡片跳到页面，条目跳到 section 深链（如 `#/components/entry`）。
- `src/pages/` — 00 基础、01 组件、02 模式、03 结构与状态。
- `src/styles/` — `base` 重置、`components` 组件、`patterns` 布局、`gallery` 文档外壳。

### 规则

1. 一个强调色（accent）：只标记时间点、当前态、发布。
2. 结构靠 1px 线与留白；不用阴影、渐变、大圆角。
3. 图片 8px，头像圆形；其余 0 圆角。
4. 组件只消费语义 token，不写死颜色/尺寸。
5. 所有交互态用 `aria-*` 或属性表达，状态不复制样式。

### 与 `web/` 同步

`tokens.css` 的变量名与 `web/` 主题变量保持一致。采用这些值前，先在 `web/` 找同名变量；没有同名时先加语义 token，不要在组件里写十六进制。当前没有自动 drift 检查，改 token 时两边一起改。

## Prototype

`prototype/` 来自 `dev/design` 的草图源码，保留 Base UI 交互：页面菜单、Activity/Connect Tabs、Composer、标签 Combobox、主题/宽度/布局切换、`⌘K` 搜索。它不接 API、不参与 `just check`，只用于设计验证。

单独运行（可选）：

```bash
cd design/prototype
pnpm install
pnpm dev
```

## Archive

`archive/ech0.fig` 是 OpenPencil 版本的源文件；`archive/screenshots/` 是旧版三张截图。它们只作历史参考，不再作为事实源。
