---
title: 安装部署
description: Docker / Compose / 二进制 / Helm；端口、数据与安全要点
---

> **推荐顺序**：**Docker 单容器** → Docker Compose → 安装脚本（systemd）→ 在线 Helm Repo 安装 → 直接运行二进制。若你是第一次部署，可先读 [快速上手](/zh/docs/start/getting-started) 再回来选方式。

---

## 部署前请确认

| 项目         | 说明                                                                         |
| ------------ | ---------------------------------------------------------------------------- |
| 端口         | 默认 **6277**。若被占用，把 `-p` 左侧改成宿主机端口，例如 `-p 8080:6277`。   |
| 数据目录     | 务必用 `-v` 映射到宿主机路径，否则删容器会丢数据。                           |
| `JWT_SECRET` | 用于会话签名，**必须**改为足够长的随机串；泄露或弱密钥会导致会话被伪造。     |
| 公网访问     | 云服务器需在控制台 **安全组** 放行对应端口；仅内网使用可只绑定 `127.0.0.1`。 |

**HTTPS**：Ech0 本身可 HTTP 运行；对外公网强烈建议在前面加 **Nginx / Caddy / Traefik** 做 TLS 终止，证书可用 Let’s Encrypt。反代时把流量转到 `http://127.0.0.1:6277`（或你的容器端口）即可。

---

## Docker（推荐）

```bash
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

- 部署完成后浏览器访问 `http://<服务器IP>:6277`（本机即 `http://127.0.0.1:6277`）。
- **请把** `JWT_SECRET="Hello Echos"` **改成你自己的随机字符串**，不要长期使用示例值。
- **第一个注册的账号**会成为管理员（Owner）；当前版本默认只有高权限账号可以发帖。
- 数据持久化在上例的 `/opt/ech0/data`（可按需改挂载路径；确保目录对容器可写）。

### 镜像与版本

- 镜像名一般为 `sn0wl1n/ech0:latest`；需要固定版本可到 [GitHub Releases](https://github.com/lin-snow/Ech0/releases) 对照标签。
- 升级流程见 [版本更新](/zh/docs/start/update)。

### 常见故障（Docker）

| 现象           | 可检查项                                                                      |
| -------------- | ----------------------------------------------------------------------------- |
| 浏览器连不上   | 防火墙 / 安全组、端口是否映射错、`docker ps` 是否运行中                       |
| 502 / 反代失败 | 反代 `proxy_pass` 地址是否指向本机 6277、WebSocket 若需要是否已按 README 配置 |
| 数据丢了       | 是否未挂载 `-v` 或误删了宿主机数据目录                                        |

---

## Docker Compose

新建目录，放入 `docker-compose.yml`（可参考仓库 [`docker/docker-compose.yml`](https://github.com/lin-snow/Ech0/blob/main/docker/docker-compose.yml) 或下列最小示例），在该目录执行：

```bash
docker compose up -d
```

若你本机命令仍是 `docker-compose`（带横杠），则使用：

```bash
docker-compose up -d
```

示例：

```yaml
services:
  ech0:
    image: sn0wl1n/ech0:latest
    container_name: ech0
    ports:
      - "6277:6277"
    volumes:
      - ./data:/app/data
    environment:
      - JWT_SECRET=请改为随机强密码
    restart: unless-stopped
```

---

## 脚本安装（systemd）

与 README 相同，从仓库拉取安装脚本：

```bash
curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" -o ech0.sh && bash ech0.sh
```

- 脚本通过 **systemd** 安装和管理服务，涉及服务启停时通常需要 **root**。
- 自定义安装目录可执行：`bash ech0.sh install /your/path/ech0`（以脚本 `--help` 为准）。
- 已在用旧版管理脚本（`em`）的机器，先执行一次引导命令把脚本换成当前版本（写入 `/usr/local/sbin/em`）：

  ```bash
  curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" | sudo bash -s -- self-update
  ```

  之后 `em update` 会同时刷新服务端二进制与管理脚本；也可随时 `em self-update` 单独更新脚本。

---

## Kubernetes（Helm）

推荐使用在线 Helm 仓库安装：

```bash
helm repo add ech0 https://lin-snow.github.io/Ech0
helm repo update
helm install ech0 ech0/ech0
```

可自定义 release 名与命名空间，例如：

```bash
helm install my-ech0 ech0/ech0 --namespace my-namespace --create-namespace
```

`JWT_SECRET` 等可通过 `values.yaml` 或 `--set` 传入，详见 Chart 目录内说明。

如果你需要从本地源码安装，也可以保留旧方式：

```bash
git clone https://github.com/lin-snow/Ech0.git
cd Ech0
helm install ech0 ./charts/ech0
```

---

## 二进制

从 [GitHub Releases](https://github.com/lin-snow/Ech0/releases) 下载对应平台压缩包，解压后：

```bash
./ech0 serve
```

常见架构包括 `linux/amd64`、`linux/arm64`、`linux/armv7`、`windows/amd64` 等。

## 首次使用

1. 打开站点，**注册第一个账号**（即 Owner）。
2. 在 **系统设置** 中填写 **服务地址**（完整 URL，含协议），便于 Connect、头像等能力正确解析。
3. 按需配置 **评论**、**对象存储**、**SSO** 等；细节见左侧各篇指南。

用户数量、发帖权限、角色名称等以当前版本 **设置界面** 为准。若「无法发帖」，请确认是否使用 Owner 或已被授权发帖的账号，并参考 [常见问题](/zh/docs/start/faq)。

---

## 环境变量与进阶配置

除 `JWT_SECRET` 外，其余变量（数据库路径、日志级别、对象存储等）按部署方式在 `.env` 或 Compose 中设置；具体名称以你的部署模板为准。时区语义：站点级统计固定 UTC，用户视角接口（如「今日」、热力图）使用浏览器请求的 **`X-Timezone`**。生产环境建议：

- 不要用默认示例密钥；
- 定期备份数据目录与快照；
- 需要自动化时，在后台创建 [访问令牌](/zh/docs/guide/accesstoken)，勿把令牌写入公开仓库。
