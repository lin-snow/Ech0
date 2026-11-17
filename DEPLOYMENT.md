# Deployment Guide

Detailed deployment and upgrade instructions for **Ech0**.

For a 60-second quickstart, see the [main README](./README.md#try-in-60-seconds).

## Table of contents

- [Quick Deployment](#quick-deployment)
  - [Environment Variables](#environment-variables)
  - [Docker (Recommended)](#docker-recommended)
  - [Docker Compose](#docker-compose)
  - [Script Deployment](#script-deployment)
  - [Kubernetes (Helm)](#kubernetes-helm)
- [Upgrading](#upgrading)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose-1)
  - [Kubernetes (Helm)](#kubernetes-helm-1)

---

## Quick Deployment

### Environment Variables

All variables, defaults, precedence rules, and the boundary between bootstrap env and admin-panel settings are documented in
[`docs/usage/configuration.md`](./docs/usage/configuration.md). The essentials:

- **Precedence:** process env → `./.env` (working directory only, no `--env-file` flag) → built-in defaults. Changes need a restart.
- **`JWT_SECRET` has no `ECH0_` prefix** (`ECH0_JWT_SECRET` is accepted as an alias). If unset, an ephemeral secret is generated and **every restart invalidates all issued tokens**.
- `ECH0_SERVER_HOST` / `ECH0_SERVER_PORT` control the bind address (default all interfaces, port `6277`); `ECH0_DB_PATH` accepts any SQLite file path.
- `ECH0_SETTING_*`, `ECH0_S3_*`, Passkey and OAuth2 redirect env vars are only **first-boot defaults** seeded into the database — once the instance is initialized, change them in the admin panel.

### 🐳 Docker (Recommended)

```shell
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

> 💡 After deployment, access `ip:6277`
> 🚷 For better security, replace `Hello Echos` in `-e JWT_SECRET="Hello Echos"` with your own secret
> 📍 The first registered account becomes administrator (currently only admins can publish)
> 🎈 Data is stored under `/opt/ech0/data`

### 🐋 Docker Compose

A ready-to-use example lives at [`docker/docker-compose.yml`](./docker/docker-compose.yml). Copy it into a new directory and run:

```shell
docker-compose up -d
```

### 🧙 Script Deployment

```shell
curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" -o ech0.sh && bash ech0.sh
```

> The script installs and manages Ech0 through systemd, so please run with root privileges when needed.
> You can run `bash ech0.sh install /your/path/ech0` to customize the install path.
>
> Already running an older `em` manager script? Refresh it once (installs the current script to `/usr/local/sbin/em`):
>
> ```shell
> curl -fsSL "https://raw.githubusercontent.com/lin-snow/Ech0/main/scripts/ech0.sh" | sudo bash -s -- self-update
> ```
>
> Afterwards `em update` keeps both the server binary and the manager script up to date.

### ☸️ Kubernetes (Helm)

If you want to deploy Ech0 in a Kubernetes cluster, use the Helm Chart provided by this project.
The chart is also listed on [Artifact Hub](https://artifacthub.io/packages/helm/ech0/ech0).

**Use the online Helm repository:**

1.  Add the Ech0 chart repository:
    ```shell
    helm repo add ech0 https://lin-snow.github.io/Ech0
    helm repo update
    ```

2.  Install with Helm:
    ```shell
    # helm install <release-name> <repo-name>/<chart-name>
    helm install ech0 ech0/ech0
    ```

    Customize the release name and namespace if needed:
    ```shell
    helm install my-ech0 ech0/ech0 --namespace my-namespace --create-namespace
    ```

**Local installation from source:**

```shell
git clone https://github.com/lin-snow/Ech0.git
cd Ech0
helm install ech0 ./charts/ech0
```

---

## Upgrading

### 🔄 Docker

```shell
# Stop current container
docker stop ech0

# Remove container
docker rm ech0

# Pull latest image
docker pull sn0wl1n/ech0:latest

# Start new version
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest
```

### 💎 Docker Compose

```shell
# Enter compose directory
cd /path/to/compose

# Pull latest image and recreate
docker-compose pull && \
docker-compose up -d --force-recreate

# Clean old images
docker image prune -f
```

### ☸️ Kubernetes (Helm)

1. Update the Helm repository index:
   ```shell
   helm repo update
   ```

2. Upgrade the Helm release:
   ```shell
   # helm upgrade <release-name> <repo-name>/<chart-name>
   helm upgrade ech0 ech0/ech0
   ```
   With a custom release name and namespace:
   ```shell
   helm upgrade my-ech0 ech0/ech0 --namespace my-namespace
   ```

---

## See also

- [Storage migration guide](./docs/usage/storage-migration.md) — local ⇄ S3 storage rules and migration.
- [Webhook usage](./docs/usage/webhook-usage.md) — webhook events and payloads.
- [MCP usage](./docs/usage/mcp-usage.md) — Model Context Protocol integration.
- [`CONTRIBUTING.md`](./CONTRIBUTING.md) — contributor workflow.
