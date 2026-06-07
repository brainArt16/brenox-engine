# Security Review Notes

## SQL injection

**Status: mitigated by design**

- All database access uses [sqlc](https://sqlc.dev/) generated queries with parameterized `$1`, `$2` placeholders
- No string-concatenated SQL in application code
- Migrations and queries live in `sql/migrations/` and `sql/queries/`

Verification:

```bash
rg 'fmt\.Sprintf.*SELECT|Query\(".*\+' internal/ cmd/ pkg/
# Expected: no matches
```

Dynamic SQL: **none found** in handlers/services. Any future raw SQL must use parameters.

## Authentication

- Passwords: bcrypt (`internal/auth`)
- JWT: HS256 with `jti` rotation on refresh; revocation list in `revoked_tokens`
- API keys: SHA-256 hash stored; plain key shown once

## Transport

- Terminate TLS at load balancer or reverse proxy in production
- WebSocket must use `wss://` in production

## Headers

Security headers middleware sets `X-Content-Type-Options`, `X-Frame-Options`, CSP, etc.

## Rate limiting

- Per-IP: `HTTP_RATE_LIMIT_PER_IP`
- Per API key: `API_RATE_LIMIT_PER_MINUTE`
- WebSocket connection caps: `WS_MAX_CONNECTIONS_PER_USER`, `WS_MAX_CONNECTIONS_PER_IP`

## Audit

Sensitive mutations logged to `audit_logs` (auth, API keys, webhooks, REST mutations).

## OpenTelemetry

Optional — not enabled. Add tracing when deploying to production if needed.
