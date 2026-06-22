# Brenox Runbook

Operations guide for deploy, rollback, and incident response.

## Deploy

### Docker Compose (staging / local prod-like)

```bash
cp .env.docker.example .env
# Edit .env — set DB_PASSWORD, JWT_SECRET, MINIO_* (see docs/SECRETS.md)
docker compose up -d --build
curl http://localhost:8080/health
```

Secrets live in `.env` (gitignored), not in `docker-compose.yaml` or the image. Only port **8080** (API) is exposed; Postgres, Redis, and MinIO are internal to the compose network.

Stack: API + Postgres + Redis + MinIO + migrate job. See [docker-compose.yaml](../docker-compose.yaml).

### Kubernetes

```bash
make k8s-build k8s-load-kind k8s-dev-up   # kind cluster; see docs/KUBERNETES.md
curl http://localhost:30080/health
```

Production: managed Postgres/Redis/S3 + `deploy/overlays/prod`. See [docs/KUBERNETES.md](KUBERNETES.md).

### Manual deploy

1. Run migrations: `make migrate`
2. Set env vars (see [SECRETS.md](SECRETS.md))
3. Start binary: `./brenox-api` or `make run`
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
