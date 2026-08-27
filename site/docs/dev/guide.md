---
title: 开发指南
description: 克隆仓库、跑通后端与前端、仓库结构与常用命令
---

本页帮助你在本机**从源码跑起** Ech0，便于改代码、调试接口或贡献 PR。若你只想自托管使用实例，请从 [快速上手](/docs/start/getting-started) 与 [安装部署](/docs/start/installation) 开始，不必克隆全仓库。

---

## 仓库里有什么

Ech0 主仓库通常包含（以你克隆的 `main` 为准）：

| 路径           | 说明                                                                    |
| -------------- | ----------------------------------------------------------------------- |
| 根目录 Go 代码 | 后端 API、业务逻辑、静态资源嵌入等                                      |
| `web/`         | 用户-facing 前端（Vite + 现代前端栈）                                   |
| `Site/`        | **官网与文档站**（本页所在文档由这里构建；开发方式见 `Site/README.md`） |
| `charts/`      | Helm Chart                                                              |
| `docs/`        | 仓库内深度文档（用法、迁移等），与 `Site/docs` 官网文档互补             |

功能细节、环境变量、架构说明以根目录 **README.zh.md** 与 **`docs/`** 为准。

---

## 环境要求摘要

### 后端（Go）

- **Go 版本**：不低于 `go.mod` 中声明的版本（当前为 1.27+）。
- **CGO**：若使用含 SQLite 的构建，需要本机 C 编译器（Windows 可用 MinGW-w64，macOS `brew install gcc`，Linux `build-essential`）。
- **Wire**：若修改了依赖注入，在相应包执行 `wire` 生成 `wire_gen.go`（见 `internal/di/` 等）。
- **代码风格**：可用 **golangci-lint**（`golangci-lint run`、`golangci-lint fmt`）。
- **热重载（可选）**：**Air**，`just air-install` 或 `go install github.com/air-verse/air@latest`。
- **接口文档**：**swag** 生成 Swagger；本地起服务后打开 `http://localhost:6277/swagger/index.html`。

启动：

```bash
just run    # 启动后端
just dev    # 若已安装 Air，则热重载
```

### 前端（`web/`）

- 使用 **pnpm** 安装依赖；Node 版本按团队约定（可用 fnm、nvm 管理）。

```bash
cd web
pnpm install
pnpm dev
```

开发服务器端口以终端输出为准（Vite 常见 **5173**），后端默认 **6277**。前端通过代理或环境配置请求后端，详见 `web` 内 Vite 配置。

---

## 联调顺序

1. 启动后端 `just run`（在仓库根目录）。
2. 启动前端 `cd web && pnpm dev`。
3. 浏览器访问前端地址，确认接口指向本地后端。

若只改 API，可用 Swagger 或 curl 直接调 `http://localhost:6277`；若只改官网文档，在 `Site/` 下按该目录 README 启动文档站。

---

## 官网文档站（`Site/`）

- 文档正文在 **`Site/docs/**/\*.md`**，路由与列表由 `Site/app/docs/registry.ts` 注册。
- 新增文档后需在 `registry.ts` 的 `DOC_ORDER`（及可选 `DOC_HERO_SLUGS`）中加入 slug，否则排序可能靠后。
- 本地开发：`pnpm install` / `pnpm dev`（在 `Site` 目录），详见 `Site/README.md`。

---

## 更多

事件总线、日志、MCP 等见仓库内 `docs/` 与 README；本页只覆盖「把开发环境跑起来」与仓库导航。提交 PR 前请跑通项目约定的测试与 Lint（见 `justfile` 与 CI 配置）。
