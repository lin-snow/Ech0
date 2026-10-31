# Ech0 6.x — 首页 UI/UX 草图

纯前端设计稿，不接后端、不复用现有 Vue 前端。中间时间线直接使用 [memo.sn0w.fyi](https://memo.sn0w.fyi) 第一页的 7 条公开 echo（含图片），样式对齐现有 `TheEchoCard` 的 timeline（日期轴/圆点/2px 竖线/11-12 正文/88% 媒体）。交互只发生在当前页面。

## 方向

- **沿用现有主题**：背景、文字、边框使用 `web/` 的 light（`#f4f1ec`）与 dark（`#333`）色板；时间线日期/圆点/链接使用现有 accent。没有渐变与卡片阴影，只有弹出层使用 Base UI 示例里的硬偏移阴影。
- **无侧栏分割线**：去掉左右两根竖线，靠留白分区；时间线自身的 2px 轴保留。
- **系统字体 + 等宽字体**：正文系统 sans，时间/计数/代码/标签用 mono。没有衬线标题。
- **微博客宽度**：正文列固定 448–480px（接近现有 `28rem`），窗口再宽也不拉伸。
- **首页内联写作**：不跳页、不弹窗。大屏（>880px）把编辑器（含头像/名称）放进 400px 左栏；变窄后回到正文顶部，草稿在重排时保留。
- **Activity + Connect**：颜色对齐现有 `web/src/components/advanced/widget/`：heatmap 12×4（48 天）、236px 宽、16px 格子、4px 间距、无圆角；活动色阶跟随纸色（surface → ink），不再用绿色；Connect 使用同一组状态色和真实实例头像。ACTIVITY / CONNECT 放在同一组 Base UI `Tabs` 里切换：宽屏在左栏，窄屏在正文顶部；专注布局不展示。

## 两种布局

| 变体 | 结构 | 正文列 | 控件/活动 |
| --- | --- | --- | --- |
| 侧栏 | 400 页面/搜索/写作/ACTIVITY·CONNECT Tabs/标签 + 480 时间线 | 480 | 左栏 Tabs |
| 专注 | 448 单列 + 40px 顶栏 | 448 | 无（不显示 Tabs） |

侧栏 `>880px` 时左栏承担页面导航/搜索/写作/ACTIVITY·CONNECT/标签；`<=880px` 左栏收起，写作框与 ACTIVITY/CONNECT 回到正文顶部。专注布局只有顶栏、写作与时间线，不带 ACTIVITY/CONNECT。

## 交互与控件

- **页面切换**：左栏最顶部与搜索并排一个 Base UI `Menu`（首页/标签/广场/面板），触发器只显示当前页并带下拉箭头，当前页在菜单里加粗；搜索限宽 180px，右侧是 RSS 与 GitHub 图标（带 Tooltip）。
- **Activity / Connect**：一组 Base UI `Tabs`（ACTIVITY / CONNECT），带滑动下划线指示器；切换时只显示热力图或实例头像。
- **筛选**：只有搜索框和可清除的标签 chip，没有类型 filter；标签集合在左栏 TAGS 下。
- **时间线**：只保留日期轴、正文/图片和更多菜单；菜单按钮是扁平 Base UI icon button，标签与互动不塞进每条 echo，和原 timeline 一致；大屏时时间线顶部没有工具条。
- **写作**：内联编辑器（大屏在 400px 左栏、输入区 132px 高、2px 边框（line-strong，不加深颜色）+ 500 字重；窄屏在正文顶部），`⌘/Ctrl + ↵` 发布；锁图标是 Base UI `Toggle`，发布后直接置顶。编辑器 foot 左侧是 Base UI `Menubar`（媒体：图片/视频/音频；扩展：网站/音乐/GitHub/推文），选中项在光标处插入 markdown/URL 模板；标签用 Base UI `Combobox` 从标签云多选，chip 可单独移除。
- **其他**：账户 Menu、明暗主题、画布宽度 Slider、布局 Tabs、`1/2` 切布局、`T` 切主题、`⌘K` 搜索。

## Base UI 组件

`Menu`（页面切换/账户/移动导航/条目操作）、`Menubar`（编辑器插入媒体/扩展）、`Combobox`（标签多选）、`Tabs`（布局、ACTIVITY·CONNECT）、`Toggle`（私密）、`Select`（主题）、`Slider`（画布宽度）、`Tooltip`（图标）、`Input`（搜索）。

## 数据

- 中间 7 条 echo 来自 memo.sn0w.fyi 第一页的公开数据（含两张本地化图片），markdown 直接渲染；标签集合来自这页内容。
- 长文中的代码块限高 260px 内部滚动；图片点击在新标签页打开。
- heatmap 为近 48 天真实发布量分档（0–4），12×4、236px 宽 / 16px 格子 / 4px 间距；Connect 使用公开实例名与今日发帖数，头像为各实例公开头像。

## 运行

从 `design/` 根目录（推荐，依赖已由根 `package.json` 管理）：

```bash
pnpm dev:prototype    # http://localhost:5174
pnpm build:prototype
```

也可以在本目录单独安装运行：

```bash
pnpm install
pnpm dev
pnpm typecheck
pnpm build
```

## 结构

```
src/
  App.tsx               # 画布、主题、布局切换、快捷键
  components/DesignBar  # 底部工具条
  components/Chrome     # 侧栏/导航、ACTIVITY·CONNECT Tabs、移动顶栏
  components/Composer   # 左栏 / 正文顶部的同一个编辑器
  components/Feed       # timeline 条目、markdown、图片
  layouts/              # ThreeColumn / Focus
  fixtures/mock.ts      # 真实 echo / heatmap / connect 数据
  styles/tokens.css     # 现有主题 token、heatmap 变量
```

> 设计稿，不是生产代码：无路由、无 API、无测试，不参与 `just check`。
