# Kubernetes deployment

Run Brenox on Kubernetes using the manifests in [`deploy/`](../deploy/). The API is stateless; Postgres, Redis, and object storage should be **managed services** in production.

## Prerequisites

| Tool | Purpose |
|------|---------|
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | Apply manifests |
| [kind](https://kind.sigs.k8s.io/) / [minikube](https://minikube.sigs.k8s.io/) / Docker Desktop K8s | Local cluster (dev) |
| Docker | Build `brenox-engine` and `brenox-migrate` images |

For production: a cloud cluster (EKS, GKE, AKS, etc.), container registry, managed PostgreSQL, managed Redis, and S3-compatible storage.

## Architecture

```text
                    ┌─────────────────┐
                    │ Ingress / LB    │  TLS + WebSocket (prod)
                    └────────┬────────┘
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
         brenox-engine      brenox-engine      brenox-engine
              └──────────────┼──────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
         PostgreSQL        Redis           S3
```

- **Dev overlay** — all dependencies run in-cluster (ephemeral `emptyDir` volumes; data is lost on pod delete).
- **Prod overlay** — API + migrate Job only; you supply external connection strings via Secret + ConfigMap patches.

See also [DEPLOYMENT.md](DEPLOYMENT.md) for multi-instance Redis pub/sub and WebSocket load-balancer notes.

## Local dev cluster

### 1. Create a cluster

You need a running cluster **before** `make k8s-dev-up`. If `kubectl` has no context, it fails with `connection refused` on `localhost:8080` — that is the missing Kubernetes API, not the Brenox app.

**kind (recommended):**

```bash
make k8s-cluster-create   # kind create cluster --name brenox
make k8s-build k8s-load-kind k8s-dev-up
```

**Docker Desktop:** Settings → Kubernetes → Enable Kubernetes → wait until green, then retry `make k8s-dev-up`.

Verify:

```bash
kubectl config current-context
kubectl cluster-info
```

### 2. Build images

```bash
make k8s-build
```

Builds:

- `brenox-engine:dev` — from root `Dockerfile`
- `brenox-migrate:dev` — from `deploy/Dockerfile.migrate` (embeds `sql/migrations`)

### 3. Load images into the cluster

**kind:**

```bash
make k8s-load-kind
```

**minikube:**

```bash
minikube image load brenox-engine:dev
minikube image load brenox-migrate:dev
```

Skip this step if you push to a registry and reference that in the overlay.

### 4. Deploy

```bash
make k8s-dev-up
```

This applies `deploy/overlays/dev`, waits for Postgres, runs the migrate Job, and waits for completion.

### 5. Verify

**NodePort** (dev overlay exposes port `30080`):

```bash
curl http://localhost:30080/health
```

**Port-forward** (works on any overlay):

```bash
make k8s-port-forward
curl http://localhost:8080/health
```

**Ingress** (optional — dev overlay serves `brenox.local` when an ingress controller is installed):

```bash
make k8s-ingress-kind   # kind + nginx ingress
echo "127.0.0.1 brenox.local" | sudo tee -a /etc/hosts
curl http://brenox.local/health
```

### 6. Teardown

```bash
make k8s-dev-down
kind delete cluster --name brenox   # if using kind
```

## Overlays

| Overlay | Path | Use case |
|---------|------|----------|
| **dev** | `deploy/overlays/dev` | Local cluster with in-cluster Postgres, Redis, MinIO |
| **prod** | `deploy/overlays/prod` | Managed DB/Redis/S3; 3 API replicas |

Preview rendered manifests:

```bash
kubectl kustomize deploy/overlays/dev
```

## Migrations

Schema migrations run as a Kubernetes **Job** (`brenox-migrate`), using the `brenox-migrate` image.

After adding new migration files, rebuild the migrate image and re-run:

```bash
make k8s-build
make k8s-load-kind          # kind only
make k8s-migrate
```

`make k8s-migrate` deletes the previous Job (if any) and applies a fresh one.

## Production deploy

### Container images (GitHub Actions)

Pushes to `main`/`master` and version tags (`v*`) build and push to **GitHub Container Registry**:

| Image | Path |
|-------|------|
| API | `ghcr.io/<owner>/<repo>/brenox-engine` |
| Migrate | `ghcr.io/<owner>/<repo>/brenox-migrate` |

Workflow: [`.github/workflows/images.yml`](../.github/workflows/images.yml). Pull requests build images only (no push).

After the first push, make the packages **public** (or configure image pull secrets on the cluster):

```text
GitHub → Packages → brenox-engine → Package settings → Change visibility
```

Prod overlay defaults to `ghcr.io/brainart16/brenox/brenox-engine:latest` — edit `deployment-patch.yaml` / `job-migrate-patch.yaml` if your repo path differs.

### Cluster setup

1. **Create secrets** (never commit real values):

   ```bash
   cp deploy/overlays/prod/config.env.example deploy/overlays/prod/config.env
   # edit config.env
   kubectl create namespace brenox
   kubectl create secret generic brenox-secrets -n brenox \
     --from-env-file=deploy/overlays/prod/config.env
   ```

2. **Patch** `deploy/overlays/prod/configmap-patch.yaml` — `DB_HOST`, `REDIS_URL`, `S3_*`, CORS/WS origins.

3. **Patch** `deploy/overlays/prod/ingress.yaml` — replace `api.example.com` with your hostname; enable cert-manager issuer if used.

4. **Apply:**

   ```bash
   kubectl apply -k deploy/overlays/prod
   make k8s-migrate K8S_OVERLAY=prod
   ```

5. **Ingress** — prod overlay includes nginx Ingress with WebSocket timeouts and cookie affinity. Install an [nginx ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/) on the cluster. TLS secret `brenox-engine-tls` is created by cert-manager or supplied manually.

   WebSocket connections are sticky to the pod that accepted the upgrade. Redis pub/sub still fans out events across nodes ([DEPLOYMENT.md](DEPLOYMENT.md)).

## Health probes

The API Deployment uses `GET /health` for liveness and readiness. The endpoint checks Postgres and Redis (when `REDIS_URL` is set). No auth required.

Prometheus metrics: `GET /metrics`.

## Secrets

| Variable | Source |
|----------|--------|
| `JWT_SECRET` | Secret |
| `DB_USER`, `DB_PASSWORD` | Secret |
| `S3_ACCESS_KEY`, `S3_SECRET_KEY` | Secret |
| Everything else | ConfigMap (`brenox-config`) |

Dev overlay generates a **non-production** Secret via Kustomize `secretGenerator`. For prod, use [SECRETS.md](SECRETS.md) guidance (External Secrets Operator, cloud secret managers, etc.).

## Environment variables

All variables from [`.env.example`](../.env.example) apply. The ConfigMap and Secret in `deploy/base/` mirror the Docker Compose stack.

| Compose service | Dev overlay | Prod |
|-----------------|-------------|------|
| `postgres` | In-cluster `postgres` Service | Managed Postgres `DB_HOST` |
| `redis` | In-cluster `redis` Service | Managed Redis `REDIS_URL` |
| `minio` | In-cluster `minio` Service | AWS S3 / GCS with `S3_USE_SSL=true` |
| `migrate` | `brenox-migrate` Job | Same Job, registry image |

## CI / registry

| Workflow | Trigger | Action |
|----------|---------|--------|
| [`ci.yml`](../.github/workflows/ci.yml) | PR + push | Go test, build |
| [`images.yml`](../.github/workflows/images.yml) | PR + push + `v*` tags | Build/push `brenox-engine` + `brenox-migrate` to GHCR |

Manual local build remains: `make k8s-build`.

## Troubleshooting

| Symptom | Check |
|---------|-------|
| `connection refused` on `localhost:8080` during `kubectl apply` | No active kube context — `kubectl config use-context kind-brenox` or `make k8s-check` (auto-selects kind context if present) |
| API pod `CrashLoopBackOff` | `kubectl logs -n brenox deploy/brenox-engine` — often missing Secret or DB unreachable |
| Readiness never passes | Migrations not applied — run `make k8s-migrate` |
| `ImagePullBackOff` | Run `make k8s-load-kind` or set `imagePullPolicy` + registry |
| WebSocket fails via Ingress | Enable WS upgrade + long proxy timeouts on ingress controller |
| Migrate Job fails | `kubectl logs -n brenox job/brenox-migrate` — Postgres credentials or connectivity |

## Related docs

- [DEPLOYMENT.md](DEPLOYMENT.md) — topology, Redis channels, graceful shutdown
- [RUNBOOK.md](RUNBOOK.md) — rollback, incidents
- [SECRETS.md](SECRETS.md) — secret rotation
- [deploy/README.md](../deploy/README.md) — manifest layout
