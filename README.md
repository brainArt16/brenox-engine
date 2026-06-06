# Brenox — Realtime Communication Platform

Go backend for a reusable realtime communication infrastructure (channels, messages, WebSocket events, presence). Uses Gin, PostgreSQL, sqlc, and gorilla/websocket.

## Documentation

| Doc | Purpose |
|-----|---------|
| [docs/BACKEND_TASKS.md](docs/BACKEND_TASKS.md) | Full task tracker and roadmap |
| [docs/postman/](docs/postman/) | Postman collection for HTTP API |
| [AGENTS.md](AGENTS.md) | Agent roles for doc sync (task tracker, README, Postman) |
| [docs/WEBSOCKET_EVENTS.md](docs/WEBSOCKET_EVENTS.md) | WebSocket event catalog |

## Repo layout

```text
cmd/api/              Application entrypoint
internal/
  auth/               Registration, login, JWT
  channels/           Channel CRUD
  chat/               Message persistence (service)
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

1. Copy `.env.example` to `.env` and set `DB_*` and `JWT_SECRET`.

2. Start database and run migrations:

```bash
make db-start
make migrate
```

3. Run the API:

```bash
make run
```

Server listens on `:8080`.

## API overview

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | No | Create account |
| POST | `/auth/login` | No | Login, returns JWT |
| POST | `/api/channels` | JWT | Create channel |
| GET | `/api/channels` | JWT | List user channels |
| POST | `/api/channels/:id/join` | JWT | Join channel |
| POST | `/api/channels/:id/leave` | JWT | Leave channel (owner blocked) |
| POST | `/api/channels/:id/messages` | JWT | Send message (member only) |
| GET | `/api/channels/:id/messages` | JWT | Message history (`limit`, `offset`) |
| GET | `/api/presence` | JWT | List globally online user IDs |
| GET | `/api/ws?channel_id=` | JWT | WebSocket upgrade |

Import [docs/postman/brenox.postman_collection.json](docs/postman/brenox.postman_collection.json) for request examples.

## Development

- SQL lives in `sql/queries` and `sql/migrations`.
- Regenerate sqlc: `sqlc generate`
- New migration: `make migration <name>`
- Handlers orchestrate HTTP only; business logic in services; DB via sqlc.

After backend changes, run the documentation agents (see [AGENTS.md](AGENTS.md)).

## Testing

Add unit tests alongside packages. Use Testcontainers or `make db-start` for integration tests.

## License

No license file yet.
