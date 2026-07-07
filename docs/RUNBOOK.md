# Brenox Runbook

Operations guide for deploy, rollback, and incident response.

## Deploy

### Docker Compose

| Environment | Compose file | Env file | Make target |
|-------------|--------------|----------|-------------|
| Dev | `docker-compose.dev.yaml` | `.env` (optional) | `make dev-up` |
| Test | `deploy/compose/docker-compose.yaml` | `.env.test` | `make test-up` |
| Prod | `deploy/compose/docker-compose.yaml` | `.env.prod` | `make prod-up` |

Test and prod share one compose file; environment-specific values live in `.env.test` / `.env.prod` (`COMPOSE_PROJECT_NAME`, `DEPLOY_RESTART`, `DB_*`, `REDIS_URL`, `S3_*`, etc.).

```bash
cp .env.test.example .env.test
make migrate-test
make test-up
curl http://localhost:8080/health
```

Secrets live in env files (gitignored), not in compose YAML or the image.

**Coolify:** Base Directory = repo root. Compose file = `docker-compose.yaml` (API only) or deploy the API via Dockerfile. Set env vars in the Coolify UI (same keys as `.env.prod.example`). For Coolify-managed Postgres on the internal Docker network, set `DB_HOST` to the Postgres container name and `DB_SSLMODE=disable`.

**Migrations on deploy:** The API Docker image runs `migrate up` on container start (`scripts/docker-entrypoint.sh`). Each git push / redeploy applies pending migrations automatically. Set `RUN_MIGRATIONS_ON_START=false` to disable. Manual one-off migrate is still available:

```bash
# Option A — from repo root with .env.prod (optional; API image migrates on start)
cp .env.prod.example .env.prod   # fill in DB_* for managed Postgres
make migrate-prod

# Option B — one-off Docker command (replace placeholders)
docker run --rm \
  -v "$(pwd)/sql/migrations:/migrations" \
  migrate/migrate:v4.17.1 \
  -path=/migrations \
  -database "postgres://USER:PASS@HOST:5432/DBNAME?sslmode=require" \
  up
```

### Kubernetes

```bash
make k8s-build k8s-load-kind k8s-dev-up   # kind cluster; see docs/KUBERNETES.md
curl http://localhost:30080/health
```

Production: managed Postgres/Redis/S3 + `deploy/overlays/prod`. See [docs/KUBERNETES.md](KUBERNETES.md).

### Manual deploy

1. Run migrations: `make migrate`
2. Set env vars (see [SECRETS.md](SECRETS.md))
3. Start binary: `./brenox-engine` or `make run`
4. Verify: `GET /health` returns `"status":"ok"`

### Pre-deploy checklist

- [ ] Migrations applied
- [ ] `JWT_SECRET` rotated if compromised
- [ ] `REDIS_URL` set for multi-instance
- [ ] `CORS_ALLOWED_ORIGINS` / `WS_ALLOWED_ORIGINS` restricted
- [ ] Metrics scraped from `/metrics`

## Rollback

1. **Application:** Deploy previous container/image tag
2. **Database:** Run down migration only if the up migration is reversible:
   ```bash
   migrate -path sql/migrations -database "$DATABASE_URL" down 1
   ```
   Prefer forward-fix migrations in production instead of `down`.
3. **Redis:** Usually no rollback needed; flush only in dev (`FLUSHDB`)

## Health and metrics

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Postgres + Redis connectivity |
| `GET /metrics` | Prometheus metrics |

Key metrics:

- `brenox_http_requests_total` — request volume by route/status
- `brenox_http_request_duration_seconds` — latency histogram

## Common incidents

### API returns 503 / health degraded

1. Check Postgres: `pg_isready`, connection limits, disk
2. Check Redis: `redis-cli ping` (realtime falls back to single-node if Redis down)
3. Review API logs (JSON stdout)

### WebSocket disconnects

1. Verify load balancer supports WebSocket upgrade + sticky sessions optional
2. Check `WS_MAX_CONNECTIONS_PER_USER` / per-IP limits (429 on upgrade)
3. Confirm `REDIS_URL` consistent across all API instances

### Rate limit spikes (429)

- `HTTP_RATE_LIMIT_PER_IP` — global per-IP limit
- `API_RATE_LIMIT_PER_MINUTE` — per API key on `/v1/*`
- Tune limits or identify abusive IP/key in audit logs

### Token / auth issues

- Refresh via `POST /auth/refresh`
- Rotated tokens revoke previous `jti` — clients must store new token
- Force re-login if refresh grace exceeded (`JWT_REFRESH_GRACE_HOURS`)

## Observability

- **Logs:** Structured JSON via `slog` to stdout
- **Audit:** Mutations logged to `audit_logs` table
- **Tracing:** OpenTelemetry not enabled by default (optional future work)

## On-call contacts

Configure your team's escalation policy here.
