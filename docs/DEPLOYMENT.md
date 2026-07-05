# Deployment

Brenox runs as one or more stateless HTTP/WebSocket API nodes backed by PostgreSQL and Redis.

## Topology

```text
                    ┌─────────────┐
                    │ Load balancer│
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
      ┌─────────┐    ┌─────────┐    ┌─────────┐
      │ API :8080│    │ API :8080│    │ API :8080│
      │  Hub     │    │  Hub     │    │  Hub     │
      └────┬────┘    └────┬────┘    └────┬────┘
           │               │               │
           └───────────────┼───────────────┘
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
        ┌───────────┐            ┌───────────┐
        │ PostgreSQL │            │   Redis    │
        │  (state)   │            │ (pub/sub)  │
        └───────────┘            └───────────┘
```

- **PostgreSQL** — users, workspaces, channels, messages (source of truth).
- **Redis** — cross-node realtime event fan-out via pub/sub.
- **API nodes** — identical; each maintains local WebSocket connections only.

## Environment

| Variable | Required | Description |
|----------|----------|-------------|
| `DB_*` | Yes | PostgreSQL connection |
| `DB_SSLMODE` | No | Postgres TLS mode (default `prefer`; use `require` for managed DB) |
| `JWT_SECRET` | Yes | JWT signing key |
| `REDIS_URL` | Multi-node | e.g. `redis://redis:6379/0` |
| `PORT` | No | HTTP listen port (default `8080`) |
| `WS_ALLOWED_ORIGINS` | Prod | Comma-separated allowed WebSocket origins |
| `WS_MAX_CONNECTIONS_PER_USER` | No | Per-user WS limit (default 5) |
| `WS_MAX_CONNECTIONS_PER_IP` | No | Per-IP WS limit (default 20) |

If `REDIS_URL` is unset, the server runs in **local-only** mode (single node, no cross-instance fan-out).

## Redis channels

Outbound realtime events publish to:

```text
workspace:{workspace_id}:channel:{channel_id}
```

Each node subscribes only to channels where it has at least one local WebSocket client.

## Health check

```http
GET /health
```

Returns `200` when database (and Redis, if configured) are reachable:

```json
{
  "status": "ok",
  "checks": {
    "database": { "status": "up" },
    "redis": { "status": "up" }
  }
}
```

Use for load balancer and orchestrator probes. No authentication required.

## Load balancer — WebSocket

WebSocket connections are **sticky to the node that accepted the upgrade**. TCP connections cannot move between instances mid-session.

| Approach | When to use |
|----------|-------------|
| **IP hash / cookie sticky sessions** | Simple setup; route the same client to the same node for WS |
| **Separate WS subdomain** | e.g. `ws.example.com` with sticky LB in front of API pool |

HTTP REST requests do **not** require stickiness — any node can serve them.

Because message delivery uses Redis pub/sub, a user on node A receives events published from node B as long as both share the same Redis and PostgreSQL.

**Presence** is stored in Redis (`presence:{user_id}`) with global and workspace-scoped online sets. WebSocket heartbeat pings refresh key TTL (`PRESENCE_TTL_SECONDS`). Without Redis, presence falls back to in-memory on a single node.

## Local multi-instance test

```bash
make db-start          # postgres + redis
make migrate
cp .env.example .env   # ensure REDIS_URL is set

# Terminal 1
go run cmd/api/main.go

# Terminal 2
PORT=8081 go run cmd/api/main.go
```

Connect WebSocket clients to `:8080` and `:8081`. Messages sent on one instance appear on the other.

Integration test (requires Redis):

```bash
make test-integration
```

## Graceful shutdown

On `SIGTERM` / `SIGINT`:

1. Redis broker closes pub/sub
2. Hub closes all WebSocket connections
3. HTTP server drains (10s timeout)

Orchestrators should allow enough time for in-flight requests to finish.

## Production checklist

- [ ] Set strong `JWT_SECRET`
- [ ] Restrict `WS_ALLOWED_ORIGINS`
- [ ] Run ≥2 API instances behind LB with Redis
- [ ] Configure `GET /health` probes
- [ ] Enable TLS at load balancer
- [ ] Managed PostgreSQL + Redis (or Redis Cluster for HA)
