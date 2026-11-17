# Ech0 Helm chart

[Ech0](https://github.com/lin-snow/Ech0) is a self-hosted, lightweight personal microblog (timeline) platform.
This chart runs the single-binary server image with a persistent volume for the SQLite database and uploaded files.

## Install

```bash
helm repo add ech0 https://lin-snow.github.io/Ech0
helm repo update
helm install ech0 ech0/ech0 --set config.JWT_SECRET="$(openssl rand -hex 32)"
```

The same package is also published as an OCI artifact to GitHub Container Registry:

```bash
helm install ech0 oci://ghcr.io/lin-snow/charts/ech0 \
  --version <chart-version> \
  --set config.JWT_SECRET="$(openssl rand -hex 32)"
```

`config.JWT_SECRET` defaults to a placeholder (`Hello Echos`) that is public in this repository, so
**always override it in production** — tokens signed with the default are forgeable.

## Configuration

| Key | Description | Default |
|-----|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` / `image.tag` | Container image and tag (tag defaults to chart `appVersion`) | `sn0wl1n/ech0` / `""` |
| `imagePullSecrets` | Pull secrets for private registries | `[]` |
| `service.type` / `service.httpPort` | Service type and port | `ClusterIP` / `6277` |
| `ingress.*` | Ingress configuration | disabled |
| `httpRoute.*` | Gateway API HTTPRoute configuration | disabled |
| `config.JWT_SECRET` | JWT signing secret. **Change this in production.** | `Hello Echos` |
| `extraEnv` / `extraEnvFrom` | Extra environment variables / env sources | `[]` |
| `persistence.enabled` / `persistence.size` / `persistence.mountPath` | PVC for `/app/data` (SQLite + uploads) | `true` / `1Gi` / `/app/data` |
| `probes.*` | Liveness and readiness probes | enabled |
| `resources`, `autoscaling`, `nodeSelector`, `tolerations`, `affinity` | Standard scheduling and scaling knobs | see `values.yaml` |

See [`values.yaml`](values.yaml) for the complete list. Server-side settings are documented in
[docs/usage/configuration.md](https://github.com/lin-snow/Ech0/blob/main/docs/usage/configuration.md);
everything under `ECH0_SETTING_*`, `ECH0_S3_*`, Passkey and OAuth2 in that guide is a first-boot default and
is changed from the admin panel afterwards.

## Upgrade

```bash
helm repo update
helm upgrade ech0 ech0/ech0
```

## Uninstall

```bash
helm uninstall ech0
```

The PVC is not removed by `helm uninstall`; delete it explicitly if you also want to drop the data.

## License

AGPL-3.0-only.
