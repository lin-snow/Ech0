---
title: Installation
description: Docker / Compose / binary / Helm; ports, data, and security essentials
---

> **Recommended order**: **single Docker container** → Docker Compose → install script (systemd) → online Helm repo → run the binary directly. If this is your first deployment, you may want to read [Getting started](/docs/start/getting-started) first, then come back to pick a method.

---

## Check before you deploy

| Item           | Notes                                                                                                                   |
| -------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Port           | Default **6277**. If it is taken, change the left side of `-p` to a host port, e.g. `-p 8080:6277`.                     |
| Data directory | Always map a host path with `-v`, otherwise deleting the container deletes your data.                                   |
| `JWT_SECRET`   | Signs sessions; **must** be changed to a sufficiently long random string; a leaked or weak key lets sessions be forged. |
| Public access  | On cloud servers, allow the port in the console's **security group**; for internal-only use, bind `127.0.0.1`.          |

**HTTPS**: Ech0 itself can run over HTTP; for public-facing deployments, strongly prefer putting **Nginx / Caddy / Traefik** in front for TLS termination; certificates can come from Let’s Encrypt. Point the reverse proxy at `http://127.0.0.1:6277` (or your container port).

---

## Docker (recommended)

```bash
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

- Once deployed, open `http://<server IP>:6277` in a browser (locally, `http://127.0.0.1:6277`).
- **Change** `JWT_SECRET="Hello Echos"` **to your own random string**; do not keep using the example value long-term.
- The **first account registered** becomes the administrator (Owner); in the current version only high-privilege accounts can post by default.
- Data persists in `/opt/ech0/data` in the example above (change the mount path as needed; make sure the directory is writable by the container).

### Images and versions

- The image name is typically `sn0wl1n/ech0:latest`; for a pinned version, match tags against [GitHub Releases](https://github.com/lin-snow/Ech0/releases).
- See [Updating](/docs/start/update) for the upgrade process.

### Common problems (Docker)

| Symptom                   | What to check                                                                                                       |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Browser cannot connect    | Firewall / security group, wrong port mapping, whether `docker ps` shows it running                                 |
| 502 / reverse proxy fails | Whether `proxy_pass` points to port 6277 on this host, and whether WebSocket is configured per the README if needed |
| Data is gone              | Whether `-v` was omitted, or the host data directory was deleted by mistake                                         |

---

## Docker Compose

Create a directory, put a `docker-compose.yml` in it (see the repository's [`docker/docker-compose.yml`](https://github.com/lin-snow/Ech0/blob/main/docker/docker-compose.yml) or the minimal example below), and run in that directory:

```bash
docker compose up -d
```

If your machine still uses the hyphenated `docker-compose` command, use:

```bash
docker-compose up -d
```

Example:

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
      - JWT_SECRET=change-to-a-random-strong-secret
    restart: unless-stopped
```

---

## Script install (systemd)

As in the README, fetch the install script from the repository:

```bash
curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" -o ech0.sh && bash ech0.sh
```

- The script installs and manages the service through **systemd**; starting and stopping services usually requires **root**.
- To use a custom install directory: `bash ech0.sh install /your/path/ech0` (see the script's `--help` for details).
- Machines already using the older management script (`em`) should run the bootstrap command once to switch the script to the current version (written to `/usr/local/sbin/em`):

  ```bash
  curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" | sudo bash -s -- self-update
  ```

  Afterwards, `em update` refreshes both the server binary and the management script; you can also run `em self-update` at any time to update the script alone.

---

## Kubernetes (Helm)

Installing from the online Helm repository is recommended:

```bash
helm repo add ech0 https://lin-snow.github.io/Ech0
helm repo update
helm install ech0 ech0/ech0
```

You can customize the release name and namespace, for example:

```bash
helm install my-ech0 ech0/ech0 --namespace my-namespace --create-namespace
```

`JWT_SECRET` and other values can be passed via `values.yaml` or `--set`; see the notes in the chart directory for details.

If you need to install from local source, you can also keep the old approach:

```bash
git clone https://github.com/lin-snow/Ech0.git
cd Ech0
helm install ech0 ./charts/ech0
```

---

## Binary

Download the archive for your platform from [GitHub Releases](https://github.com/lin-snow/Ech0/releases), extract it, then:

```bash
./ech0 serve
```

Common architectures include `linux/amd64`, `linux/arm64`, `linux/armv7`, `windows/amd64`, and others.

## First use

1. Open the site and **register the first account** (the Owner).
2. In **System settings**, fill in the **service address** (a full URL including the scheme) so features such as Connect and avatars resolve correctly.
3. Configure **comments**, **object storage**, **SSO**, and so on as needed; see the respective guides in the sidebar for details.

The number of users, posting permissions, and role names follow the **settings interface** of the current version. If you “cannot post,” confirm that you are using the Owner account or one that has been granted posting permission, and see the [FAQ](/docs/start/faq).

---

## Environment variables and advanced configuration

Besides `JWT_SECRET`, other variables (database path, log level, object storage, and so on) are set in `.env` or Compose depending on your deployment method; use the names in your deployment template. Time zone semantics: site-level statistics always use UTC, while user-facing endpoints (such as “today” and heatmaps) use the **`X-Timezone`** sent by the browser. For production we recommend:

- Do not use the default example secret;
- Back up the data directory and snapshots regularly;
- When you need automation, create an [access token](/docs/guide/accesstoken) in the admin panel, and never commit tokens to public repositories.
