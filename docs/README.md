# 文档索引

本目录按用途分为 **使用说明**（部署与功能接入）与 **开发设计**（贡献代码时的约定与背景）。品牌与配图在 `imgs/`。

## 使用说明（`usage/`）

面向部署管理员与集成方：如何接入能力、如何迁移数据。

| 文档 | 说明 |
|------|------|
| [usage/mcp-usage.md](usage/mcp-usage.md) | MCP（Model Context Protocol）接入：Token、Host 配置、协议要点 |
| [usage/webhook-usage.md](usage/webhook-usage.md) | Webhook：事件、签名、管理接口与故障处理 |
| [usage/storage-migration.md](usage/storage-migration.md) | 存储迁移：本地与 S3、`key` 与路径规则、换桶与迁移注意事项 |
| [usage/capsule.md](usage/capsule.md) | 胶囊（Capsule）：内容导出/导入、校验、编译静态站，以及与快照的分工 |

## 开发设计（`dev/`）

面向本仓库贡献者：UI 约定、国际化契约、日志规范，以及鉴权模型的设计记录。

| 文档 | 说明 |
|------|------|
| [dev/architecture-overview.md](dev/architecture-overview.md) | **架构全景图**：进程启动、分层后端、业务领域、事件子系统、基础设施与 `pkg/` 自研库的关系总览（新手入口） |
| [dev/config-and-settings-architecture.md](dev/config-and-settings-architecture.md) | 配置与设置的依赖架构：config/`internal/setting`/kvstore/transaction 四层关系、两层配置、事务如何随 ctx 流动、「读直连写走域」规约 |
| [dev/table-design-standard.md](dev/table-design-standard.md) | Panel 表格组件（含管理端列表）的布局与交互规范 |
| [dev/i18n-contract.md](dev/i18n-contract.md) | 前后端国际化约定（locale、API 错误字段、key 命名） |
| [dev/logging.md](dev/logging.md) | 日志库使用与字段约定 |
| [dev/access-token-scope-design.md](dev/access-token-scope-design.md) | Access Token 的 `typ` / scope / audience 设计背景（实现以代码为准） |
| [dev/capsule/spec.md](dev/capsule/spec.md) | Capsule 交换格式与 CLI 的规范性定义（是什么） |
| [dev/capsule/capsule-design.md](dev/capsule/capsule-design.md) | Capsule 的设计依据与决策记录（为什么这么定） |

## 商业推演（`business/`）

**假设性沙盘推演，非项目实际计划**。Ech0 当前不打算商业化，此目录仅用于完整走一遍「如果要商业化」的决策过程。

| 文档 | 说明 |
|------|------|
| [business/commercialization-blueprint.md](business/commercialization-blueprint.md) | 商业化闭环推演：资产盘点、定位与红线、四条收入腿、定价、获客到复购的完整闭环、工程分仓策略、单位经济模型、风险与路线图 |
| [business/compliance-reality-check.md](business/compliance-reality-check.md) | 合规现实检验：托管 UGC 的责任转移、国内经营性 ICP/安全评估/实名/内容审核的成本量级、海外 DMCA/GDPR 对照，以及它如何否决蓝图的主力产品 |

## 资源（`imgs/`）

Logo、架构示意图等静态资源；根目录 [README.md](../README.md) 中的预览图等链接指向此处。
