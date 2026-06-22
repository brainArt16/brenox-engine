# Kubernetes deployment

Kustomize manifests for running Brenox on Kubernetes.

## Layout

```text
deploy/
  Dockerfile.migrate     Migrate image (bundles sql/migrations)
  base/                  API Deployment, Service, migrate Job, ConfigMap
  overlays/
    dev/                 In-cluster Postgres, Redis, MinIO (local clusters)
    prod/                API only — wire managed Postgres, Redis, S3
```

Full guide: [docs/KUBERNETES.md](../docs/KUBERNETES.md)

## Quick start (dev overlay)

Requires: Docker, a local cluster ([kind](https://kind.sigs.k8s.io/), [minikube](https://minikube.sigs.k8s.io/), or Docker Desktop Kubernetes), and `kubectl`.

```bash
# kind example
kind create cluster --name brenox

make k8s-build
make k8s-load-kind    # skip if using a registry / minikube image load
make k8s-dev-up
curl http://localhost:30080/health   # NodePort, or: make k8s-port-forward
```

Optional Ingress (`brenox.local`) — see [docs/KUBERNETES.md](../docs/KUBERNETES.md#local-dev-cluster).

Teardown:

```bash
make k8s-dev-down
```

## Production

See [overlays/prod/README.md](overlays/prod/README.md) — create secrets externally, patch hostnames, push images to your registry, apply `deploy/overlays/prod`.
