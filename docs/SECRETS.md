# Secrets Management

Brenox never commits secrets to the repository. All sensitive values are supplied via environment variables or your platform's secret store.

## Required secrets

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | HS256 signing key — **min 32 random bytes in production** |
| `DB_PASSWORD` | PostgreSQL password |
| `S3_SECRET_KEY` | Object storage secret (MinIO/S3) |

## Recommended secret store

| Environment | Approach |
|-------------|----------|
| Local dev | `.env` file (gitignored), copy from `.env.example` |
| Staging/Prod | AWS Secrets Manager, GCP Secret Manager, Vault, or K8s secrets |
| CI | GitHub Actions secrets (never log values) |

## Rotation

| Secret | Rotation procedure |
|--------|-------------------|
| `JWT_SECRET` | Rotate secret → all users re-login; plan maintenance window |
| API keys | Revoke via `DELETE /api/apps/:id/keys/:key_id`, issue new key |
| DB password | Update secret store → rolling restart API pods |
| Webhook secrets | Re-register webhook to get new secret |

## Do not

- Commit `.env` files
- Log tokens, API keys, or passwords
- Use default `JWT_SECRET=dev-secret-change-me` in production
- Share sandbox (`bx_test_`) keys in production builds

## Docker / Compose

| File | Purpose |
|------|---------|
| `.env.docker.example` | Template for `docker compose` — copy to `.env` |
| `.env` | **Gitignored** — real secrets for local stack or deploy |
| `docker-compose.yaml` | References `${VAR}` and `env_file: .env` — no embedded passwords |
| `Dockerfile` | Builds binary only; no secrets baked in |

For cloud production, inject env vars from your secret store (K8s secrets, AWS Secrets Manager, etc.) instead of a `.env` file on disk.

## Verification

```bash
# Ensure .env is gitignored
git check-ignore -v .env

# Scan for accidental secrets (example)
rg -i 'password\s*=\s*["\'][^"\']+["\']' --glob '!*.example' .
```
