<p align="left">
  <a href="https://hellogithub.com/repository/lin-snow/Ech0" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=neutral"
         alt="Featured｜HelloGitHub"
         width="250"
         height="54" />
  </a>
</p>

<p align="right">
  <a title="en-US" href="./README.md">
    <img src="https://img.shields.io/badge/-English-545759?style=for-the-badge" alt="English">
  </a>
  <img src="https://img.shields.io/badge/-简体中文-F54A00?style=for-the-badge" alt="简体中文">
</p>


<div align="center">
  <img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

  [预览地址](https://memo.vaaat.com/) | [官网与文档](https://www.ech0.app/) | [Ech0 Hub](https://hub.ech0.app/)

  # Ech0
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/lin-snow/Ech0)](https://github.com/lin-snow/Ech0/releases) ![License](https://img.shields.io/github/license/lin-snow/Ech0) [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lin-snow/Ech0) [![Hello Github](https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=small)](https://hellogithub.com/repository/lin-snow/Ech0)

</div>



> 自托管个人微博客：你的时间线可以被分享、讨论，同时数据完全由你掌控。

像 Memos 这样的工具非常适合快速记录想法。Ech0 更关注“记录之后”的阶段：把内容发布到个人时间线，让更多人可以持续关注和互动。
你可以在自己的服务器上托管内容，保留完整控制权，同时通过可选评论与分享保持轻连接，而不是变成复杂社交平台。
它保持轻量、易部署、完全开源。

**适合你，如果你想：**
- 搭建一个自己的公开或半公开动态站
- 用统一界面发布短文、链接与媒体卡片
- 兼顾数据主权，同时保留 RSS 与评论等能力
- 让个人内容具备轻社交连接能力，而不需要重型社交产品

**不太适合你，如果你需要：**
- 双链知识库式的笔记工作流（例如 Obsidian 风格）
- 团队优先的协作文档平台（例如 Notion 风格）
- 纯私密备忘且不关注发布/时间线场景

![界面预览](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>目录</strong></summary>

- [Ech0](#ech0)
  - [1 分钟试用](#1-分钟试用)
  - [为什么是 Ech0](#为什么是-ech0)
  - [完整能力清单](#完整能力清单)
  - [极速部署](#极速部署)
    - [🐳 Docker 部署（推荐）](#-docker-部署推荐)
    - [🐋 Docker Compose](#-docker-compose)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm)
  - [版本更新](#版本更新)
    - [🔄 Docker](#-docker)
    - [💎 Docker Compose](#-docker-compose-1)
    - [☸️ Kubernetes (Helm)](#️-kubernetes-helm-1)
  - [常见问题](#常见问题)
  - [反馈与社区](#反馈与社区)
- [开源治理](#开源治理)
  - [项目架构](#项目架构)
  - [开发指南](#开发指南)
    - [后端环境要求](#后端环境要求)
    - [前端环境要求](#前端环境要求)
    - [启动前后端联调](#启动前后端联调)
  - [感谢充电支持！](#感谢充电支持)
  - [Star 增长曲线](#star-增长曲线)
  - [致谢](#致谢)
  - [支持项目](#支持项目)

</details>

---

## 1 分钟上手

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

然后访问 `http://ip:6277`：

1. 注册你的第一个账号。
2. 首个账号会自动成为 Owner（管理员权限）。
3. 默认仅高权限账号可发布内容。

更多部署方式见 [极速部署](#极速部署)。

## 为什么是 Ech0

- 📝 **为个人发布而生**：时间线优先的微博客形态，适合持续发布想法、链接和短内容。  
- 🤝 **轻社交属性可选开启**：内容可被分享、评论与互动，连接读者但不过度社交化。  
- 🧘 **专注浏览体验**：Zen 风格的低干扰时间线阅读体验。  
- ⚡ **Markdown 与媒体一体化**：写作、链接卡片、视频解析在同一条发布流里完成。  
- 🔒 **个人优先且完全可控**：默认面向个人实例，可按需启用多用户角色，同时保持自托管、可订阅 RSS、AGPL-3.0 完全开源。  

## 完整能力清单

<details>
  <summary><strong>展开查看完整能力</strong></summary>

### 产品亮点

- ☁️ **轻量高效架构**：低资源占用与小体积镜像，适合个人服务器到 ARM 设备。  
- 🚀 **极速部署体验**：开箱即用 Docker 部署，从安装到运行一条命令即可启动。  
- 📦 **自包含部署包**：提供完整二进制与容器镜像，无需额外依赖。  
- 💻 **跨平台支持**：支持 Linux、Windows 与 ARM 架构设备（如 Raspberry Pi）。  

### Storage & Data

- 🗂️ **VireFS 统一存储抽象层**：以 **VireFS** 统一本地存储与 S3 兼容对象存储的挂载与管理。  
- ☁️ **S3 对象存储支持**：原生支持 S3 兼容对象存储，便于云端资源扩展。  
- 📦 **数据主权架构**：内容与元数据由用户掌控，并支持 RSS 输出。  
- 🔄 **数据迁移机制**：支持迁移导入历史数据，配合快照导出实现迁移与归档。  
- 🔐 **自动备份系统**：支持 Web、CLI、TUI 三种导出/备份方式与后台自动备份。  

### Writing & Content

- ✍️ **Markdown 写作体验**：基于 **markdown-it** 的编辑与渲染引擎，支持插件扩展和实时预览。  
- 🧘 **Zen Mode 沉浸式阅读**：提供干扰最小化的 Timeline 浏览模式。  
- 🏷️ **标签管理系统**：支持标签分类、快速过滤与精准检索。  
- 🃏 **富媒体卡片内容**：支持网站链接、GitHub 项目等卡片展示。  
- 🎥 **视频内容解析**：支持哔哩哔哩与 YouTube 视频解析展示。  

### Media & Assets

- 📁 **可视化文件管理器**：内建文件上传、浏览与资源管理能力。  

### Social & Interaction

- 💬 **原生评论系统**：内建评论与评论管理功能，无需第三方评论服务。  
- 🃏 **内容互动能力**：支持点赞、分享等社交互动。  

### Auth & Security

- 🔑 **OAuth2 / OIDC 身份认证**：支持 OAuth2 与 OIDC 协议，便于接入第三方登录。  
- 🙈 **Passkey 无密码登录**：支持生物识别或硬件密钥登录。  
- 🔑 **访问令牌管理**：支持生成与吊销带 Scope 的访问令牌，便于 API 调用与第三方集成。
- 👤 **多账户权限管理**：支持多用户与权限控制。  

### System & Developer

- 🧱 **Busen 数据总线架构**：通过自研 Busen 实现模块解耦通信与可靠消息传递。  
- 📊 **结构化日志系统**：系统日志统一为结构化格式，提升可读性与可分析性。  
- 🖥️ **实时系统日志控制台**：内建 Web 控制台可实时查看日志流，便于调试与排障。  
- 📟 **TUI 管理界面**：提供终端交互界面，适合服务器环境管理。  
- 🧰 **CLI 工具链**：提供 CLI 工具，支持自动化管理与脚本集成。  
- 🔗 **开放 API 与 Webhook**：提供完整 API 与 Webhook，便于外部系统集成和自动化工作流。  

### Experience

- 🌍 **跨设备适配**：响应式设计，适配桌面、平板与移动浏览器。  
- 🌐 **i18n 多语言支持**：支持多语言界面切换，覆盖不同语言使用场景。  
- 👾 **PWA 支持**：支持安装为 Web App，体验更接近原生应用。  
- 🌗 **主题与 Dark Mode**：支持深色模式与主题扩展。  

### License

- 🎉 **完全开源**：基于 **AGPL-3.0** 协议发布，无追踪、无订阅、无 SaaS 依赖。  

</details>

---

## 极速部署

<!-- ### 🧙 脚本一键部署（推荐,请确保网络可以访问GitHub Release）
```shell
curl -fsSL "https://sh.soopy.cn/ech0.sh" -o ech0.sh && bash ech0.sh
``` -->

### 🐳 Docker 部署（推荐）

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

> 💡 部署完成后访问 ip:6277 即可使用  
> 🚷 建议把`-e JWT_SECRET="Hello Echos"`里的`Hello Echos`改成别的内容以提高安全性  
> 📍 首次使用注册的账号会被设置为管理员（目前仅管理员支持发布内容）  
> 🎈 数据存储在/opt/ech0/data下  

### 🐋 Docker Compose

创建一个新目录并将 `docker-compose.yml` 文件放入其中

在该目录下执行以下命令启动服务：

```shell
docker-compose up -d
```

### ☸️ Kubernetes (Helm)

如果你希望在 Kubernetes 集群中部署 Ech0，可以使用项目提供的 Helm Chart。

由于本项目暂时未提供在线 Helm 仓库，你需要先将代码库克隆到本地，然后从本地目录进行安装。

1.  **克隆代码库:**
    ```shell
    git clone https://github.com/lin-snow/Ech0.git
    cd Ech0
    ```

2.  **使用 Helm 安装:**
    ```shell
    # helm install <发布名称> <chart目录>
    helm install ech0 ./charts/ech0
    ```

    你也可以自定义发布名称和命名空间：
    ```shell
    helm install my-ech0 ./charts/ech0 --namespace my-namespace --create-namespace
    ```

---

## 版本更新

> ⚠️ 目前不支持从 v3 直接更新到 v4。请先在 v3 面板中点击“导出快照”，然后重新部署 v4，并在 v4 面板中选择“v3 迁移”即可导入原有数据。

### 🔄 Docker

```shell
# 停止当前的容器
docker stop ech0

# 移除容器
docker rm ech0

# 拉取最新的镜像
docker pull sn0wl1n/ech0:latest

# 启动新版本的容器
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

### 💎 Docker Compose

```shell
# 进入 compose 文件目录
cd /path/to/compose

# 拉取最新镜像并重启
docker-compose pull && \
docker-compose up -d --force-recreate

# 清理旧镜像
docker image prune -f
```

### ☸️ Kubernetes (Helm)

1. **更新代码库:**
   进入本地的 Ech0 代码库目录，并拉取最新的代码。
   ```shell
   cd Ech0
   git pull
   ```

2. **更新 Helm Release:**
   使用 `helm upgrade` 命令更新你的发布版本。
   ```shell
   # helm upgrade <发布名称> <chart目录>
   helm upgrade ech0 ./charts/ech0
   ```
   如果你使用了自定义的发布名称和命名空间，请使用对应的名称：
   ```shell
   helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
   ```

<!-- ---

## 访问方式

### 🖥️ TUI 模式

![TUI 模式](./docs/imgs/tui.png)

直接运行对应的二进制文件即可。例如在 Windows 中，双击 `Ech0.exe`。 -->

---

## 常见问题

1. **Ech0是什么？**
   Ech0 是一款轻量级的开源自托管平台，专为快速发布与分享个人想法、文字和链接而设计。它提供简洁的界面与零干扰体验，并确保数据始终由你自己掌控。

2. **Ech0不是什么？**
   Ech0不是传统的笔记软件，设计之初并不是为了专业的笔记管理和记录（如Obsidian、Notion等），Ech0的核心功能类似朋友圈/说说。

3. **Ech0 是免费的吗？**
   是的，Ech0 完全免费且开源，遵循 AGPL-3.0 协议。它没有广告、追踪、订阅或服务依赖。

4. **如何进行备份和导入数据？**
   Ech0 支持通过“快照导出 / 迁移导入”进行数据迁移与恢复。部署层面建议定期备份你映射的数据目录（如 `/opt/ech0/data`）。默认情况下核心数据位于本地数据库；若启用了对象存储，媒体文件会按存储配置写入对应后端。

5. **Ech0 支持 RSS 吗？**
   是的，Ech0 支持 RSS 订阅，您可以通过 RSS 阅读器订阅您的内容更新。

6. **为什么发布失败，提示联系管理员？**
   当前版本下，发布权限默认受限于高权限账号。初始化时创建的首个账号为 Owner（同时具备管理权限），普通用户默认不能发布，需要由高权限用户按实际策略授权。若是首次部署，请先对照 [1 分钟试用](#1-分钟试用) 确认首个账号是否为 Owner。

7. **为什么没有明确的权限划分？**
   Ech0 当前采用轻量权限模型（Owner / Admin / 普通用户），目标是降低管理复杂度并保持日常使用流畅。后续会根据社区反馈持续迭代。

8. **为什么别人无法显示自己的Connect头像？**
   要使别人显示自己的Connect头像需要在`系统设置-服务地址`中填入自己当前的实例地址，比如我自己填的是部署ech0后的域名`https://memo.vaaat.com`(注意：这里填的链接需要带上http或https)。

9.  **设置中的MetingAPI项是什么？**
   这是音乐卡片解析所使用的 API 地址。你可以填写自建或可信的解析服务；未配置时会使用系统默认解析地址。建议在生产环境中优先使用你可控的服务端点。

10. **为什么添加后的Connect只显示了一部分？**
      因为后端会尝试获取所有connect的实例信息，如果某个实例挂了或者无法访问则会被抛弃，只返回获取到的有效connect实例的信息给前端。

11. **如何开启评论功能？**
      在面板的评论管理页面开启评论并按需配置审核与验证码参数即可。当前为内建评论系统，无需额外接入第三方评论平台。

12. **S3 存储如何配置？**
      在存储设置中填写 Provider、Endpoint、Bucket、Access Key、Secret Key 等信息。`endpoint` 建议填写不含 `http/https` 的地址；若前端需直接访问媒体资源，请确保对象具备可访问策略（如 public-read 或等效 CDN/网关配置）。

13. **如何启用 Passkey 无密码登录？**
      在 `SSO - Passkey` 页面先配置 `WebAuthn RP ID` 与 `WebAuthn Origins`，保存并显示“Passkey就绪”后，再按浏览器提示绑定你常用的生物识别或安全密钥设备即可使用。

---

## 反馈与社区

- 若程序出现 bug，可在 [Issues](https://github.com/lin-snow/Ech0/issues) 中反馈。
- 针对新增或改进的需求，欢迎前往 [Discussions](https://github.com/lin-snow/Ech0/discussions) 一起交流。
- 官方 QQ 群号：1065435773

| 官方QQ交流群                                                    | 其它交流群 |
| --------------------------------------------------------------- | ---------- |
| <img src="./docs/imgs/qq.png" alt="QQ群" style="height:250px;"> | 暂无       |


---

## 开源治理

- [贡献指南](./CONTRIBUTING.md)
- [行为准则](./CODE_OF_CONDUCT.md)
- [安全策略](./SECURITY.md)
- [许可证](./LICENSE)

---

## 项目架构

- 后端事件总线已切换为 [Busen](https://github.com/lin-snow/Busen)：采用 typed-first in-process 架构，并通过显式背压、hooks 与 drain shutdown 提升稳定性。
---

## 开发指南
### 后端环境要求
📌 **Go 1.26.0+**

📌 **C 编译器**
使用 `go-sqlite3` 等需要 CGO 的库时，需安装：
- Windows：
    - [MinGW-w64](https://winlibs.com/)
    - 解压后将bin目录添加到PATH
- macOS： `brew install gcc`
- Linux： `sudo apt install build-essential`

📌 **Google Wire**
安装[wire](https://github.com/google/wire)用于依赖注入文件生成:
- `go install github.com/google/wire/cmd/wire@latest`

📌 **Golangci-Lint**
安装[Golangci-Lint](https://golangci-lint.run/)用于lint和fmt:
- 在项目根目录下执行`golangci-lint run`进行lint
- 在项目根目录下执行`golangci-lint fmt`进行格式化

📌 **Air（后端热重载，可选）**
- 推荐通过 Makefile 安装：`make air-install`
- 或手动安装：`go install github.com/air-verse/air@latest`

📌 **Swagger**
安装[Swagger](https://github.com/swaggo/gin-swagger)用于生成和使用符合OpenAPI规范的接口文档
- 在项目根目录下执行`swag init -g internal/server/server.go -o internal/swagger`后生成或更新swagger文档
- 打开浏览器访问`http://localhost:6277/swagger/index.html`查看和使用swagger文档

📌 **Event 运行参数（Busen）**
- `ECH0_EVENT_DEFAULT_BUFFER` / `ECH0_EVENT_DEFAULT_OVERFLOW`
- `ECH0_EVENT_DEADLETTER_BUFFER` / `ECH0_EVENT_SYSTEM_BUFFER`
- `ECH0_EVENT_AGENT_BUFFER` / `ECH0_EVENT_AGENT_PARALLELISM`
- `ECH0_EVENT_INBOX_BUFFER`
- `ECH0_EVENT_WEBHOOK_POOL_WORKERS` / `ECH0_EVENT_WEBHOOK_POOL_QUEUE`

### 前端环境要求
📌  **NodeJS v25.5.0+, PNPM v10.30.0+**
> 注：如需要多个nodejs版本共存可使用[fnm](https://github.com/Schniz/fnm)进行管理

---

### 启动前后端联调
**第一步： 后端（在 Ech0 根目录下）：**
```shell
make run # 普通启动后端（等价于 go run main.go serve）
make dev # 使用 Air 启动后端热重载
```
> 如果依赖注入关系发生了变化先需要在`ech0/internal/di/`下执行`wire`命令生成新的`wire_gen.go`文件

**第二步： 前端（新终端）：**
```shell
cd web # 进入前端目录

pnpm install # 如果没有安装依赖则执行

pnpm dev # 启动前端预览
# 或在项目根目录执行：make web-dev
```

**第三步： 前后端启动后访问：**
前端预览： http://localhost:5173 （端口在启动后可在控制台查看）  
后端预览： http://localhost:6277 （默认后端端口为6277）  

> 对使用**层次化架构的包**进行导入时，请使用**规范的 alias 命名**：  
> model 层： `xxxModel`  
> util 层： `xxxUtil`  
> handler 层： `xxxHandler`  
> service 层： `xxxService`  
> repository 层： `xxxRepository`  

---

## 感谢充电支持！

感谢所有为项目充电的朋友！你们的支持让项目持续发光发热 💡✨


|                        ⚙️ 用户                        | 🔋 充电日期 | 💬 留言                 |
| :--------------------------------------------------: | :--------: | :--------------------- |
|                     🧑‍💻 匿名小伙伴                     | 2025-5-19  | 笨比程序员买杯糖水喝吧 |
|        🧑‍💻 [@sseaan](https://github.com/sseaan)        | 2025-7-27  | Ech0是个好东西🥳        |
| 🧑‍💻 [@QYG2297248353](https://github.com/QYG2297248353) | 2025-10-10 | 无                     |
|    🧑‍💻 [@continue33](https://github.com/continue33)    | 2025-10-23 | 感谢修复R2             |
|    🧑‍💻 [@hoochanlon](https://github.com/hoochanlon)   | 2025-10-28 | 无        |
|       🧑‍💻 [@Rvn0xsy](https://github.com/Rvn0xsy)       | 2025-11-12 | 很棒的项目，我会持续关注！|
|                     🧑‍💻 王贼臣                     | 2025-11-20 | 感谢www.cardopt.cn             |
|       🧑‍💻 [@ljxme](https://github.com/ljxme)    | 2025-11-30 | 略尽绵薄之力😋             |
|       🧑‍💻 [@he9ab2l](https://github.com/he9ab2l)    | 2025-12-23 | 无            |
|       🧑‍💻 鸿运当头(windfore)    | 2026-1-6 | 感谢你创造ech0           |
|       🧑‍💻 匿名用户    | 2026-01-23  | 无           |


---

## Star 增长曲线

<a href="https://www.star-history.com/#lin-snow/Ech0&Timeline">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=lin-snow/Ech0&type=Timeline" />
 </picture>
</a>

---

## 致谢

- 感谢广大用户提供的各种改进建议和问题反馈
- 感谢所有开源社区的贡献者与支持者

[![Contributors](https://contrib.rocks/image?repo=lin-snow/Ech0)](https://contrib.rocks/image?repo=lin-snow/Ech0)

![Alt](https://repobeats.axiom.co/api/embed/d69b9177e4a121e31aaed95354ff862c928ca22d.svg "Repobeats analytics image")

---

## 支持项目


🌟 如果你觉得 **Ech0** 不错，欢迎为项目点个 Star！🚀

Ech0 完全开源且免费，持续维护和优化离不开大家的支持。如果这个项目对你有所帮助，也欢迎通过赞助支持项目的持续发展。你的每一份鼓励和支持，都是我们前进的动力！
你可以向打赏二维码付款，然后备注你的github名称，将在首页 `README.md` 页面向所有展示你的贡献

|                  支持平台                  |                         二维码                         |
| :----------------------------------------: | :----------------------------------------------------: |
| [**爱发电**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

---


```cpp

███████╗     ██████╗    ██╗  ██╗     ██████╗
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝

```
