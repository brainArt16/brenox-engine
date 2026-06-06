# Brenox — Realtime Communication Platform

Go backend for a reusable realtime communication infrastructure (workspaces, channels, messages, WebSocket events, presence). Uses Gin, PostgreSQL, sqlc, and gorilla/websocket.

## Documentation

| Doc | Purpose |
|-----|---------|
| [docs/BACKEND_TASKS.md](docs/BACKEND_TASKS.md) | Full task tracker and roadmap |
| [docs/postman/](docs/postman/) | Postman collection for HTTP API |
| [AGENTS.md](AGENTS.md) | Agent roles for doc sync (task tracker, README, Postman) |
| [docs/WEBSOCKET_EVENTS.md](docs/WEBSOCKET_EVENTS.md) | WebSocket event catalog |
| [docs/PERMISSIONS.md](docs/PERMISSIONS.md) | Role-based permission matrix |

## Repo layout

```text
cmd/api/              Application entrypoint
internal/
  auth/               Registration, login, JWT
  authz/              Role-based permission checks
  workspaces/         Workspace CRUD + member admin
  channels/           Channel CRUD (workspace-scoped)
  chat/               Message persistence
  realtime/           WebSocket hub and handlers
  db/                 sqlc-generated queries
  database/           Postgres pool
  middleware/         JWT auth middleware
pkg/jwt/              JWT helpers
sql/
  migrations/         Schema migrations
  queries/            sqlc query definitions
docs/
  BACKEND_TASKS.md    Task tracker
  postman/            API collection
```

## Quick start

Prerequisites: Go 1.20+, Docker, Make.

1. Copy `.env.example` to `.env` and set `DB_*`, `JWT_SECRET`, and optional WebSocket vars (`WS_ALLOWED_ORIGINS`, `WS_MAX_CONNECTIONS_PER_USER`, `WS_MAX_CONNECTIONS_PER_IP`).

2. Start database and run migrations:

```bash
make db-start
make migrate
```

3. Run the API:

```bash
make run
make test   # authz unit tests
```

Server listens on `:8080`.

### Existing databases

Migration `000004_add_workspaces` backfills a **Default Workspace** per channel owner. For a clean slate in dev, reset the database:

```bash
docker compose -f docker-compose.dev.yaml down -v
make db-start && make migrate
```

## API overview

All channel and message routes are scoped under a workspace.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Create account |
| POST | `/auth/login` | No | Login, returns JWT |
| POST | `/api/workspaces` | JWT | Create workspace |
| GET | `/api/workspaces` | JWT | List user workspaces |
| GET | `/api/workspaces/:workspace_id` | JWT | Workspace detail |
| GET | `/api/workspaces/:workspace_id/members` | JWT | List members |
| POST | `/api/workspaces/:workspace_id/members` | JWT | Invite member (owner/admin) |
| DELETE | `/api/workspaces/:workspace_id/members/:user_id` | JWT | Remove member |
| PATCH | `/api/workspaces/:workspace_id/members/:user_id` | JWT | Change member role |
| POST | `/api/workspaces/:workspace_id/channels` | JWT | Create channel (role gated) |
| GET | `/api/workspaces/:workspace_id/channels` | JWT | List workspace channels |
| POST | `/api/workspaces/:workspace_id/channels/:id/join` | JWT | Join channel |
| POST | `/api/workspaces/:workspace_id/channels/:id/leave` | JWT | Leave channel |
| POST | `/api/workspaces/:workspace_id/channels/:id/messages` | JWT | Send message |
| GET | `/api/workspaces/:workspace_id/channels/:id/messages` | JWT | Message history |
| GET | `/api/presence` | JWT | Globally online user IDs |
| GET | `/api/ws?workspace_id=&channel_id=` | JWT (header or `?token=`) | WebSocket upgrade |

WebSocket auth accepts `Authorization: Bearer …` or `?token=` on the upgrade URL. Connection limits and allowed origins are configurable via env (see `.env.example`). Graceful shutdown closes active WebSocket connections on SIGTERM.

Channel names are unique **per workspace**.

Import [docs/postman/brenox.postman_collection.json](docs/postman/brenox.postman_collection.json) for request examples.

## Development

- SQL lives in `sql/queries` and `sql/migrations`.
- Regenerate sqlc: `make sqlc` or `sqlc generate`
- New migration: `make migration <name>`
- Handlers orchestrate HTTP only; business logic in services; DB via sqlc.

After backend changes, run the documentation agents (see [AGENTS.md](AGENTS.md)).

## Testing

Add unit tests alongside packages. Use Testcontainers or `make db-start` for integration tests.

## License

No license file yet.
