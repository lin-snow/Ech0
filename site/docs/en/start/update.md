---
title: Updating
description: Back up before upgrading; concrete commands for Docker, Compose, and Helm
---

Back up the **data directory** before upgrading (and, if object storage is enabled, back up bucket objects and metadata according to your policy). Minor upgrades usually just mean pulling the new image and recreating the container; for **major versions** (especially across major releases), read the release notes before you start.

> **v3 → v4**: an in-place upgrade is not possible. First **export a snapshot** from v3, then deploy v4 and use **v3 migration** in the v4 admin panel to import it. See [Data control](/docs/guide/datacontrol).

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

Replace `JWT_SECRET` with the value used in your environment (the same as on first deployment); if **`TZ`** was set initially, keep it in the upgrade command as well. Make sure the data volume path is correct. If the instance fails to start after upgrading, roll the image tag back and check the release notes first — do not try to troubleshoot by wiping the data directory.

---

## Docker Compose

```bash
cd /path/to/compose
docker compose pull && docker compose up -d --force-recreate
```

If you use the older `docker-compose` command:

```bash
docker-compose pull && docker-compose up -d --force-recreate
```

Optionally clean up dangling images: `docker image prune -f`

---

## Kubernetes (Helm)

1. Enter the cloned repository directory and pull the latest code: `git pull`
2. Upgrade the release:

```bash
helm upgrade ech0 ./charts/ech0
```

If you use a custom release name and namespace:

```bash
helm upgrade my-ech0 ./charts/ech0 --namespace my-namespace
```

---

## Recommended after upgrading

1. Open the admin panel and confirm the version number and health status (if shown).
2. Skim the timeline, publish a test Echo, and check that comments and attachments work.
3. If you use Connect or webhooks, check whether the release notes mention behavior changes.

For more troubleshooting ideas, see [FAQ](/docs/start/faq).
