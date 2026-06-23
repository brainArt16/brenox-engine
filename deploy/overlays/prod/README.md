# Production overlay — managed Postgres, Redis, and S3 only.
#
# 1. Copy and edit config (or patch via your GitOps tool):
#      cp config.env.example config.env
# 2. Create the secret (never commit config.env):
#      kubectl create namespace brenox --dry-run=client -o yaml | kubectl apply -f -
#      kubectl create secret generic brenox-secrets -n brenox --from-env-file=config.env
# 3. Edit configmap-patch.yaml with your hostnames and CORS origins.
# 4. Edit ingress.yaml — set your API hostname and TLS issuer.
# 5. Images: push to main builds ghcr.io/<owner>/<repo>/brenox-engine:latest (see .github/workflows/images.yml).
#    Adjust deployment-patch.yaml / job-migrate-patch.yaml if your GHCR path differs.
# 6. Deploy:
#      kubectl apply -k deploy/overlays/prod
#      make k8s-migrate K8S_OVERLAY=prod
