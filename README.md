# Ech0 - 开源、自托管、专注思想流动的轻量级发布平台

<p align="center">
  <img alt="Ech0" src="./docs/imgs/logo.svg" width="100">
</p>

Ech0 是一款专为轻量级分享而设计的开源自托管平台，支持快速发布与分享你的想法、文字与链接。简单直观的操作界面，轻松管理你的内容，让分享变得更加自由，确保数据完全掌控，随时随地与世界连接。

![界面预览](./docs/imgs/screenshot.png)

[预览地址](https://soopy.cn/)
[官网地址](https://echo.soopy.cn/)

---

## 核心优势

☁️ **原子级轻量**：内存占用、可执行文件、镜像大小均不到**30MB**，单SQLite文件存储架构  
🚀 **极速部署**：无需配置，从安装到使用只需1条命令  
✍️ **零干扰写作**：纯净的在线Markdown编辑器，**支持丰富的Markdown插件与预览**  
📦 **数据主权**：所有内容存储于本地SQLite文件，支持RSS订阅  
🎉 **永久免费**：MIT协议开源，无追踪/无订阅/无服务依赖  
🌍 **跨端适配**：完美兼容桌面/移动浏览器，支持手机、iPad、PC三端响应式布局  
👾 **WPA适配**：支持作为Web应用安装  
📝 **内置Todo管理**：轻松记录、管理每日待办事项，帮助你高效规划和追踪任务进度  
🔗 **Ech0 Connect**：全新内容聚合与互联功能，支持多实例间互通、内容订阅与同步，打造属于你的去中心化内容网络  
🎵 **无缝音乐集成**: 内置超低资源占用的音乐播放器，支持本地音频流解析，提供沉浸式背景音乐与专注模式  
🎥 **即时视频分享**: 原生支持哔哩哔哩视频智能解析，输入BV号/链接即可自动提取并嵌入  
🃏 **丰富的快捷卡片**：支持网站链接、GitHub项目等多种富媒体内容一键分享，让信息展示更加直观生动  
⚙️ **高级自定义功能**：为高级用户提供便捷自定义样式与脚本，分享更具表现力  

---

## 3秒极速部署

### 🧙 脚本一键部署（推荐）
```shell
curl -fsSL "http://echo.soopy.cn/install.sh" -o install_ech0.sh && bash install_ech0.sh
```

### 🐳 docker部署（推荐）

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos!" \
  sn0wl1n/ech0:latest
```

> 💡 部署完成后访问 ip:1314 即可使用  
> 🚷 建议把`-e JWT_SECRET="Hello Echos!"`里的`Hello Echos!`改成别的内容以提高安全性  
> 📍 首次使用注册的账号会被设置为管理员（目前仅管理员支持发布内容）  
> 🎈 数据存储在/opt/ech0/data下  

### 🐋 docker-componse部署

创建一个新目录并将 `docker-compose.yml` 文件放入其中

在该目录下执行以下命令启动服务：

```shell
docker-compose up -d
```

## 如何更新

### 🔄 Docker部署如何更新 

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
  -v /opt/ech0/config/config.yaml:/app/data/config/config.yaml \
  -e JWT_SECRET="Hello Echos!" \
  sn0wl1n/ech0:latest
```

### 💎 Docker Compose部署如何更新

```shell
# 进入 compose 文件目录
cd /path/to/compose

# 拉取最新镜像并重启
docker-compose pull && \
docker-compose up -d --force-recreate

# 清理旧镜像
docker image prune -f
```

---

# 🦖 未来目标

- [x] 使用裸Vue3重写整个前端
- [x] 修复一个安全性的问题
- [ ] 重构后端，使其更加优雅高效
- [ ] 优化各项画面细节 && 增加更多实用功能
- [ ] 性能优化 && 美化界面

---

# ❓ 常见问题

1. **Ech0是什么？**
Ech0 是一款轻量级的开源自托管平台，专为快速发布与分享个人想法、文字和链接而设计。它提供简洁的界面，支持零干扰的写作体验，所有数据存储于本地，确保用户对内容的完全控制。

2. **Ech0 是免费的吗？**
是的，Ech0 完全免费且开源，遵循 MIT 协议。它没有广告、追踪、订阅或服务依赖。

3. **如何进行备份和恢复数据？**
由于所有内容都存储在本地 SQLite 文件中，您只需备份/opt/ech0/data目录中的文件即可（具体选择部署时的映射路径）。在需要恢复时，直接将备份文件还原即可。

4. **Ech0 支持 RSS 吗？**
是的，Ech0 支持 RSS 订阅，您可以通过 RSS 阅读器订阅您的内容更新。

5. **为什么发布失败，提示联系管理员？**
当前版本设计上，只有管理员可以发布内容。部署后，首个注册的用户会自动被设置为系统管理员，其他用户无法发布内容（可在设置中分配权限）。

6. **为什么没有明确的权限划分？**
Ech0 旨在保持简洁和轻量，因此在设计时没有复杂的权限系统。我们希望用户能够专注于分享内容，而不是被复杂的权限管理所困扰。为了保持流畅的使用体验，Ech0 尽量精简了功能，避免不必要的复杂性。（因此目前只有管理员与非管理员之分，所以请谨慎分配你的权限）

---

# 🛠️ 开发

🔧 依赖环境  
📌 后端： `Go 1.24.3+`  
📌 前端： `NodeJS v22.15.0, PNPM`  

🏗️ 启动  
在Ech0根目录下：

后端：
```shell
go run cmd/server/main.go
```

前端（新终端）：
```shell
cd web # 进入前端目录

pnpm install

pnpm dev
```

---

# 🥰 致谢

- 感谢 [Gin](https://github.com/gin-gonic/gin) 提供高性能的后端框架支持  
- 感谢 [Md-Editor-V3](https://github.com/imzbf/md-editor-v3) 提供强大易用的 Markdown 编辑器  
- 感谢 [Figma](https://www.figma.com/) 提供便捷的 Logo 设计工具  
- 感谢舍友的 Logo 设计  
- 感谢所有开源社区的贡献者与支持者  

---

# ☕ 支持


🌟 如果你觉得 **Ech0** 不错，欢迎为项目点个 Star！🚀  

Ech0 完全开源且免费，持续维护和优化离不开大家的支持。如果这个项目对你有所帮助，也欢迎通过赞助支持项目的持续发展。你的每一份鼓励和支持，都是我们前进的动力！  
你可以向打赏二维码付款，然后备注你的github名称，将在首页 `README.md` 页面向所有展示你的贡献  

| 支持平台 | 二维码 |
| :------: | :-------------: |
| [**爱发电**](https://afdian.com/a/l1nsn0w) | <img src="./docs/imgs/pay.jpeg" alt="Pay" width="200"> |

