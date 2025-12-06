<p align="left">
  <a href="https://hellogithub.com/repository/lin-snow/Ech0" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=neutral"
         alt="Featured｜HelloGitHub"
         width="250"
         height="54" />
  </a>
</p>

<p align="right">
  <a title="en" href="./README.en.md">
    <img src="https://img.shields.io/badge/-English-545759?style=for-the-badge" alt="English">
  </a>
  <img src="https://img.shields.io/badge/-简体中文-F54A00?style=for-the-badge" alt="简体中文">
</p>


<div align="center">
  <img alt="Ech0" src="./docs/imgs/logo.svg" width="150">

  [预览地址](https://memo.vaaat.com/) | [官网与文档](https://echo.soopy.cn/) | [Ech0 Hub](https://echohub.soopy.cn/)

  # Ech0
</div>

<div align="center">

[![GitHub release](https://img.shields.io/github/v/release/lin-snow/Ech0)](https://github.com/lin-snow/Ech0/releases) ![License](https://img.shields.io/github/license/lin-snow/Ech0) [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lin-snow/Ech0) [![Hello Github](https://api.hellogithub.com/v1/widgets/recommend.svg?rid=8f3cafdd6ef3445dbb1c0ed6dd34c8b5&claim_uid=swhbQfnJvKS0t7I&theme=small)](https://hellogithub.com/repository/lin-snow/Ech0)

</div>



> 面向个人的新一代开源、自托管、专注思想流动的轻量级联邦发布平台。

Ech0 是一款专为个人用户设计的新一代开源自托管平台，低成本、超轻量，支持 ActivityPub 协议，让你轻松发布和分享想法、文字与链接。简洁直观的界面结合高效的命令行工具，让内容管理变得简单而自由。你的数据完全自主可控，随时随地与世界联通，构建属于你的思想网络。

![界面预览](./docs/imgs/screenshot.png)

---

<details>
   <summary><strong>目录</strong></summary>

- [Ech0](#ech0)
  - [产品亮点](#产品亮点)
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

## 产品亮点

☁️ **原子级轻量**：内存占用不到**15MB**，镜像大小不到**40MB**,单SQLite文件存储架构  
🚀 **极速部署**：无需配置，从安装到使用只需1条命令  
✍️ **零干扰写作**：纯净的在线Markdown编辑器，支持丰富的Markdown插件与预览  
📦 **数据主权**：所有内容存储于本地SQLite文件，支持RSS订阅  
🔐 **安全备份机制**：支持Web、TUI、CLI三种模式下一键导出、备份完整数据,支持后台自动备份  
♻️ **无感恢复支持**：支持TUI/CLI快照恢复与Web面板在线零停机恢复机制，保障数据安全无忧  
🎉 **永久免费**：AGPL-3.0协议开源，无追踪/无订阅/无服务依赖  
🌍 **跨端适配**：完美兼容桌面/移动浏览器，支持手机、iPad、PC三端响应式布局  
👾 **PWA适配**：支持作为Web应用安装，体验更接近原生  
🏷️ **优雅的标签管理与过滤**：智能标签系统、快速过滤与精准检索，让信息组织更高效、更直观  
☁️ **S3 存储集成** —— 原生适配 S3 兼容对象存储，轻松实现云端高效存储  
🌐 **ActivityPub 联邦** —— 与 Mastodon、Misskey 等平台互联共通，构建去中心化生态  
🔑 **OAuth2 集成** —— 原生支持 OAuth2 协议，轻松接入第三方登录与 API 授权  
📝 **内置Todo管理**：轻松记录、管理每日待办事项，帮助你高效规划和追踪任务进度  
🪶 **高可用 Webhook**：支持与外部系统进行实时扩展与协作，实现事件驱动的自动化工作流  
🌗 **深色模式与主题扩展**：支持自适应系统或自由切换 Dark Mode，支持后期扩展自定义配色  
🤖 **快捷配置启动 Agent AI**：快捷配置多种大语言模型，无需动手折腾即可体验 AI  
🧰 **命令行利器**：内置高可用 CLI 工具，为开发者与高级用户提供极致掌控力与自动化体验  
🔑 **快捷访问令牌管理**：支持生成与一键吊销访问令牌，安全高效地完成 API 调用与第三方集成  
📊 **实时系统资源监控面板**：基于 WebSocket 的高性能监控模块，让你对运行状态一目了然  
📟 **极致 TUI 支持**：面向终端用户打造的友好交互界面，轻松对Ech0进行管理  
🔗 **Ech0 Connect**：全新多实例互联功能，实现Ech0实例间状态订阅与跟踪  
🎵 **无缝音乐集成**: 内置超低资源占用的音乐播放器，提供沉浸式背景音乐与专注模式  
🎥 **即时视频分享**: 原生支持哔哩哔哩/YouTube视频智能解析  
🃏 **丰富的快捷卡片**：支持网站链接、GitHub项目等多种富媒体内容一键分享，让信息展示更加直观生动  
⚙️ **高级自定义功能**：为高级用户提供便捷自定义样式与脚本，分享更具表现力  
💬 **评论系统**：支持快捷接入Twikoo评论服务，轻量、快捷、无侵入，为你的内容带来即时互动与反馈  
💻 **跨平台兼容**：原生支持 Windows、Linux 以及树莓派等 ARM 架构设备，确保多样化部署场景下稳定运行  
🔗 **Ech0 Hub 广场**：内置 Ech0 Hub 广场，轻松发现、订阅和共享优质内容  
📦 **二进制自包含**：集成完整资源，无需额外安装依赖或配置，极大简化部署流程  
🔗 **丰富的 API 支持**: 提供开放 API，方便与其他系统集成，实现更灵活的应用场景  
🃏 **内容展示支持**：支持类 X（Twitter）风格卡片展示，同时支持点赞分享等社交互动  
👤 **多账户权限管理**：支持多账户体系，提供灵活的权限控制，保障内容和功能访问的安全性与私密性  

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
  -v /opt/ech0/backup:/app/backup \
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
  -v /opt/ech0/backup:/app/backup \
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

<!-- ### 🔐 SSH 模式

在终端通过 6278 端口连接部署实例：

```shell
ssh -p 6278 ssh.vaaat.com
``` -->

---

## 常见问题

1. **Ech0是什么？**
   Ech0 是一款轻量级的开源自托管平台，专为快速发布与分享个人想法、文字和链接而设计。它提供简洁的界面，支持零干扰的写作体验，所有数据存储于本地，确保用户对内容的完全控制。

2. **Ech0不是什么？**
   Ech0不是传统的笔记软件，设计之初并不是为了专业的笔记管理和记录（如Obsidian、Notion等），Ech0的核心功能类似朋友圈/说说。

3. **Ech0 是免费的吗？**
   是的，Ech0 完全免费且开源，遵循 AGPL-3.0 协议。它没有广告、追踪、订阅或服务依赖。

4. **如何进行备份和恢复数据？**
   由于所有内容都存储在本地 SQLite 文件中，您只需备份/opt/ech0/data目录中的文件即可（具体选择部署时的映射路径）。在需要恢复时，直接将备份文件还原即可，当然也可以使用在线数据管理，直接在设置-数据管理选项内使用创建、导出、恢复快照等功能即可快速管理数据。若恢复成功后数据依然没有显示最新内容，可以手动重启一下Docker容器即可！

5. **Ech0 支持 RSS 吗？**
   是的，Ech0 支持 RSS 订阅，您可以通过 RSS 阅读器订阅您的内容更新。

6. **为什么发布失败，提示联系管理员？**
   当前版本设计上，只有管理员可以发布内容。部署后，首个注册的用户会自动被设置为系统管理员，其他用户无法发布内容（可在设置中分配权限）。

7. **为什么没有明确的权限划分？**
   Ech0 旨在保持简洁和轻量，因此在设计时没有复杂的权限系统。我们希望用户能够专注于分享内容，而不是被复杂的权限管理所困扰。为了保持流畅的使用体验，Ech0 尽量精简了功能，避免不必要的复杂性。（因此目前只有管理员与非管理员之分，所以请谨慎分配你的权限）。

8. **为什么别人无法显示自己的Connect头像？**
   要使别人显示自己的Connect头像需要在`系统设置-服务地址`中填入自己当前的实例地址，比如我自己填的是部署ech0后的域名`https://memo.vaaat.com`(注意：这里填的链接需要带上http或https)。

9.  **设置中的MetingAPI项是什么？**
   这是用于解析获取音乐流媒体直链的服务api,用于分享的音乐卡片功能，如果不设置则默认使用ech0提供的api（部署于vercel）。

10. **为什么添加后的Connect只显示了一部分？**
      因为后端会尝试获取所有connect的实例信息，如果某个实例挂了或者无法访问则会被抛弃，只返回获取到的有效connect实例的信息给前端。

11. **Ech0不建议发什么？**
      Ech0发布的内容分为三部分：文字、图片、扩展内容（如音乐、视频等播放器卡片），Ech0不建议发布同时包含`文字 + 图片 + 扩展内容`这种密集内容，因为其违反了Ech0的一些设计理念，同时在任何时候都不推荐发布扩展内容或长篇幅的文章。

12. **如何开启评论功能？**
      在设置页面的`评论API`项中填入你部署后的Twikoo后端地址后自动开启，当前仅支持[Twikoo](https://twikoo.js.org/)

13. **S3 存储如何配置？**
      在存储设置页面填入所需配置信息，注意：endpoint不需要填http或者https开头，存储桶需提供公共访问权限。

14. **如何加入联邦宇宙？**
      需要将Ech0绑定一个域名，并在设置界面的服务器地址填写域名即可自动加入联邦宇宙，填写示例如下：`https://memo.vaaat.com`

---

## 反馈与社区

- 若程序出现 bug，可在 [Issues](https://github.com/lin-snow/Ech0/issues) 中反馈。
- 针对新增或改进的需求，欢迎前往 [Discussions](https://github.com/lin-snow/Ech0/discussions) 一起交流。
- 官方 QQ 群号：1065435773

| 官方QQ交流群                                                    | 其它交流群 |
| --------------------------------------------------------------- | ---------- |
| <img src="./docs/imgs/qq.png" alt="QQ群" style="height:250px;"> | 暂无       |


---

## 项目架构

![技术架构图](./docs/imgs/Ech0技术架构图.svg)
> by ExcaliDraw
---

## 开发指南
### 后端环境要求
📌 **Go 1.25.3+**

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

📌 **Swagger**
安装[Swagger](https://github.com/swaggo/gin-swagger)用于生成和使用符合OpenAPI规范的接口文档
- 在项目根目录下执行`swag init -g internal/server/server.go -o internal/swagger`后生成或更新swagger文档
- 打开浏览器访问`http://localhost:6277/swagger/index.html`查看和使用swagger文档

### 前端环境要求
📌  **NodeJS v24.10.0+, PNPM v10.20.1+**
> 注：如需要多个nodejs版本共存可使用[fnm](https://github.com/Schniz/fnm)进行管理

---

### 启动前后端联调
**第一步： 后端（在 Ech0 根目录下）：**
```shell
go run main.go # 编译并启动后端
```
> 如果依赖注入关系发生了变化先需要在`ech0/internal/di/`下执行`wire`命令生成新的`wire_gen.go`文件

**第二步： 前端（新终端）：**
```shell
cd web # 进入前端目录

pnpm install # 如果没有安装依赖则执行

pnpm dev # 启动前端预览
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
