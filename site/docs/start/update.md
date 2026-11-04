---
title: 版本更新
description: 升级前备份；Docker、Compose、Helm 的具体命令
---

升级前请**备份数据目录**（以及若启用对象存储，需按策略备份桶内对象与元数据）。小版本升级通常只需拉新镜像并重建容器；**大版本**（尤其跨主版本）务必读 Release 说明再动手。

> **v3 → v4**：不能原地升级。请先在 v3 **导出快照**，再部署 v4，在 v4 面板做 **v3 迁移** 导入。详见 [数据管理](/docs/guide/datacontrol)。

---

## Docker

```bash
docker stop ech0
docker rm ech0
docker pull sn0wl1n/ech0:latest
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  -e TZ=Asia/Shanghai \
  sn0wl1n/ech0:latest
```

请把 `JWT_SECRET` 换成与你环境一致的值（与初次部署相同）；若初次部署已设置 **`TZ`**，升级命令中请一并保留。数据卷路径勿改错。若升级后无法启动，先回滚镜像标签并核对 Release 说明，不要清空数据目录试错。

---

## Docker Compose

```bash
cd /path/to/compose
docker compose pull && docker compose up -d --force-recreate
```

若使用旧命令 `docker-compose`，则：

```bash
docker-compose pull && docker-compose up -d --force-recreate
```

可选清理悬空镜像：`docker image prune -f`

---

## Kubernetes（Helm）

1. 进入克隆下来的仓库目录，拉最新代码：`git pull`
2. 升级 release：

```bash
helm upgrade ech0 ./charts/ech0
```

若使用了自定义 release 名与命名空间：

```bash
helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
```

---

## 升级后建议

1. 打开管理后台，确认版本号与健康状态（若有）。
2. 快速浏览时间线、发一条测试 Echo、检查评论与附件是否正常。
3. 若使用 Connect / Webhook，关注 Release 是否提及行为变更。

更多排错思路见 [常见问题](/docs/start/faq)。
