# 文档索引

`docs/` 只放**文本**：使用说明、开发约定与设计记录。二进制与产线工程在 [`assets/`](#素材assets)，索引课程见下方"收录规则"。

## 收录规则

1. **`docs/` 只放 `.md`**。截图、视频、PDF、可构建工程一律进 `docs/assets/`。唯一例外是 `docs/imgs/` 下的 4 个对外契约文件（见下）。
2. **新增/删除文档必须同步本页**。本页是唯一入口：读者从这里进入，孤儿文档等于不存在。
3. **冻结路径（外部 URL 契约，禁止改名或移动）**：
   - `docs/usage/*.md`、`docs/dev/capsule/spec.md` — 已被发布站 `site/docs/**` 用绝对 GitHub blob 链接引用；
   - `docs/imgs/{logo.svg,screenshot.png,pay.jpeg,qq.png}` — 被 `charts/ech0/Chart.yaml` 的 `icon`、`.github/FUNDING.yml`、4 个 README 引用（Helm 仓库与 ArtifactHub 会长期缓存）。
4. **命名**：`dev/` 下只有设计记录/待定稿用 `*-design.md` 后缀；流程与规约用平铺名词（`testing.md`、`logging.md`）。

## 使用说明（`usage/`）

面向部署者与集成方：装好之后要做什么。

| 文档 | 说明 |
|------|------|
| [usage/configuration.md](usage/configuration.md) | 环境变量与 `.env`：加载顺序、全部变量与默认值、env 与后台设置的边界 |
| [usage/mcp-usage.md](usage/mcp-usage.md) | MCP 接入：Bearer Token / scope / audience、Host 配置、工具与资源清单 |
| [usage/webhook-usage.md](usage/webhook-usage.md) | Webhook：事件、签名、管理接口与故障处理 |
| [usage/storage-migration.md](usage/storage-migration.md) | 存储迁移：更换 S3 服务商、本地 ⇄ 对象存储双向互迁 |
| [usage/capsule.md](usage/capsule.md) | 胶囊（Capsule）：导出/导入/校验、编译静态站，以及与快照的分工 |

## 开发文档（`dev/`）

按体裁分四段读：**入口 → 流程 → 规约 → 设计记录**。规约改动会影响当次改动；设计记录是历史决策，只追加不改写。

### 入口

| 文档 | 说明 |
|------|------|
| [dev/architecture-overview.md](dev/architecture-overview.md) | **架构全景**：进程启动、分层后端、业务领域、Agent/MCP、事件子系统、基础设施与 `pkg/`。新人第一读 |
| [ech0.app/arch](https://www.ech0.app/arch)（源文件 [site/public/arch/index.html](../site/public/arch/index.html)） | 交互式架构图：进程全景 + `pkg/` 内部结构，浏览器可拖拽缩放 |

### 流程

| 文档 | 说明 |
|------|------|
| [dev/development.md](dev/development.md) | 本地开发：环境、热重载、前后端联调（EN） |
| [dev/testing.md](dev/testing.md) | 后端 Go 测试体系与写法约定 |
| [dev/release-process.md](dev/release-process.md) | 发版流程（EN） |
| [dev/helm-release-validation.md](dev/helm-release-validation.md) | 发版后 Helm 仓库 / GitHub Pages 验收清单 |

### 规约（改代码时必须遵守）

| 文档 | 说明 |
|------|------|
| [dev/i18n-contract.md](dev/i18n-contract.md) | 前后端国际化契约：locale、API 错误字段、key 命名 |
| [dev/logging.md](dev/logging.md) | 日志库用法与字段 / `module` 约定 |
| [dev/table-design-standard.md](dev/table-design-standard.md) | Panel 表格组件的布局与交互规范 |
| [dev/huma-raw-gin-endpoints.md](dev/huma-raw-gin-endpoints.md) | 仍走裸 gin 的端点清单及原因（其余 JSON 端点已全量 Huma 化） |

### 设计记录（为什么这样）

| 文档 | 说明 |
|------|------|
| [dev/auth-design.md](dev/auth-design.md) | 认证体系：双 Token、OAuth / Passkey、吊销策略 |
| [dev/access-token-scope-design.md](dev/access-token-scope-design.md) | Access Token 的 `typ` / scope / audience 设计背景（实现以代码为准） |
| [dev/cli-design.md](dev/cli-design.md) | CLI（`apps/cli`）：Bun/TS 技术栈、Huma→TS/zod 生成链、三种登录与 device flow、输出契约 |
| [dev/config-and-settings-architecture.md](dev/config-and-settings-architecture.md) | config / setting / kvstore / transaction 四层依赖、"读直连写走域" |
| [dev/timezone-design.md](dev/timezone-design.md) | 时间与时区：UTC epoch 存储、本地展示、跨时区一致性 |
| [dev/job-runner-design.md](dev/job-runner-design.md) | 长时有状态作业子系统（Job Runner） |
| [dev/snapshot-design.md](dev/snapshot-design.md) | Migrator 与 Snapshot：导入/导出统一收敛 |
| [dev/llm-chat-design.md](dev/llm-chat-design.md) | Chat RAG：索引管线、检索、SSE 流式 |
| [dev/agent-toolcall-design.md](dev/agent-toolcall-design.md) | `internal/agent` 重构与 ReAct Tool Calling |
| [dev/capsule/spec.md](dev/capsule/spec.md) | Capsule v1 规范：格式与 CLI 的规范性定义（是什么） |
| [dev/capsule/capsule-design.md](dev/capsule/capsule-design.md) | Capsule 的设计依据与决策讨论（为什么这么定） |

## 商业推演（`business/`）

**假设性沙盘推演，非项目实际计划**。Ech0 当前不打算商业化，此目录只用于完整走一遍"如果要商业化"的决策过程。

| 文档 | 说明 |
|------|------|
| [business/commercialization-blueprint.md](business/commercialization-blueprint.md) | 商业化闭环推演：资产盘点、定位与红线、四条收入腿、定价、获客到复购、分仓策略、单位经济模型、风险与路线图 |
| [business/compliance-reality-check.md](business/compliance-reality-check.md) | 合规现实检验：托管 UGC 的责任转移、国内经营性 ICP / 安全评估 / 实名 / 内容审核成本、海外 DMCA / GDPR 对照 |

## 法务（`legal/`）

| 文档 | 说明 |
|------|------|
| [legal/agpl-compliance.md](legal/agpl-compliance.md) | AGPL-3.0 合规指南：使用、修改、分发的实际义务（非法律意见） |

## 素材（`assets/`）

设计与营销素材，以及它们的产线工程。**不是文档**，不进上面的索引。

| 目录 | 内容 |
|------|------|
| `assets/architecture/` | 架构图成品：`ech0-architecture-poster.pdf`、`Ech0 Arch.jpeg`、`Ech0 Architecture Atlas · Master System Diagram.html`；`Ech0技术架构图.excalidraw` 为可编辑源码及其 png/svg 导出 |
| `assets/screenshots/archive/` | 历史版本截图（v1-v2 … v4.7.3、TUI、旧 mockup、旧架构图）。当前无任何引用，仅作存档；新增截图请放回 `docs/imgs/screenshot.png` 前先确认引用方 |
| `assets/social-preview/` | GitHub 社交预览模板：`social-preview.html` + 1280×640 的 png/jpg |
| `assets/carousel/` | 六帧宣传图 `1-6.jpeg` 与设计渲染页 `Ech0_carousel/index.html` |
| `assets/video/` | 品牌宣传片成片、关键帧与可复现出片工程，说明见 [assets/video/README.md](assets/video/README.md) |

`docs/imgs/` 只保留 4 个对外契约文件：`logo.svg`、`screenshot.png`、`pay.jpeg`、`qq.png`。
