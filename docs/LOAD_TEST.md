# Load Testing Baseline

Baseline targets and scripts for capacity planning. Adjust for your hardware and SLA.

## Targets (single API instance, local docker-compose stack)

| Scenario | Target | Notes |
|----------|--------|-------|
| Health check | ≥ 2,000 RPS | No auth, no DB write |
| Authenticated REST (messages GET) | ≥ 200 RPS | JWT + Postgres read |
| WebSocket connections | ≥ 500 concurrent | Per instance; tune `WS_MAX_*` limits |
| Public API `/v1/messages` POST | ≥ 100 RPS | Rate limited per key |

These are **starting baselines**, not guarantees. Run on your infrastructure before setting SLOs.

## k6 smoke test

Install [k6](https://k6.io/) and run:

```bash
k6 run scripts/load/smoke.js
```

Default: 30 VUs for 30s against `http://localhost:8080/health`.

## vegeta alternative

```bash
echo "GET http://localhost:8080/health" | vegeta attack -duration=30s -rate=100 | vegeta report
```

## Before load testing

1. Start stack: `make test-up` (or `make dev-up` for local dev)
2. Ensure migrations applied
3. Monitor `/metrics` and Postgres connections during test
4. Increase rate limits temporarily if testing auth routes (`HTTP_RATE_LIMIT_PER_IP`)

## Bottleneck checklist

- Postgres connection pool size
- Redis pub/sub fan-out under many WebSocket clients
- Single-hub goroutine broadcast channel (256 buffer)

See [RUNBOOK.md](RUNBOOK.md) for incident response if limits are hit in production.
