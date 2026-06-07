# Brenox — Realtime Communication Platform

Go backend for a reusable realtime communication infrastructure (workspaces, channels, messages, WebSocket events, presence). Uses Gin, PostgreSQL, sqlc, and gorilla/websocket.

## Documentation

| Doc | Purpose |
|-----|---------|
| [docs/BACKEND_TASKS.md](docs/BACKEND_TASKS.md) | Full task tracker and roadmap |
| [docs/postman/](docs/postman/) | Postman collection for HTTP API |
| [AGENTS.md](AGENTS.md) | Agent roles for doc sync (task tracker, README, Postman) |
| [docs/WEBSOCKET_EVENTS.md](docs/WEBSOCKET_EVENTS.md) | WebSocket event catalog |
| [docs/WEBRTC.md](docs/WEBRTC.md) | Voice/video call signaling + TURN/STUN client config |
| [docs/WEBRTC_CLIENT.md](docs/WEBRTC_CLIENT.md) | SDK integration guide for WebRTC clients |
| [docs/openapi.yaml](docs/openapi.yaml) | Public Developer API OpenAPI spec |
| [docs/SDK_INTEGRATION.md](docs/SDK_INTEGRATION.md) | SDK auth, WebSocket, reconnection guide |
| [docs/PERMISSIONS.md](docs/PERMISSIONS.md) | Role-based permission matrix |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Multi-instance topology, Redis, load balancer |

## Repo layout

```text
cmd/api/              Application entrypoint
internal/
  auth/               Registration, login, JWT
  authz/              Role-based permission checks
  workspaces/         Workspace CRUD + member admin
  channels/           Channel CRUD (workspace-scoped)
  chat/               Message persistence
  realtime/           WebSocket hub, Redis broker
  presence/           Redis-backed presence store + API
  notifications/      Notification persistence + dispatch
  storage/            S3/MinIO object storage
  attachments/        File uploads + message attachments
  calls/              Voice call rooms + WebRTC signaling
  apps/               Developer apps + API key management
  developerapi/       Public /v1 API for third-party integrations
  users/               User profile API
  webhooks/           Webhook delivery dispatcher
  ratelimit/          API key rate limiting
  redis/              Redis client wrapper
  health/             Health check handler
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

1. Copy `.env.example` to `.env` and set `DB_*`, `JWT_SECRET`, `REDIS_URL`, `S3_*`, and optional WebSocket vars.

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

Server listens on `:8080` (override with `PORT`). Set `REDIS_URL` for cross-instance realtime; see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

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
| GET | `/health` | No | DB + Redis health probe |
| POST | `/auth/register` | No | Create account |
| POST | `/auth/login` | No | Login, returns JWT |
| POST | `/auth/refresh` | No | Refresh JWT (valid or recently expired token) |
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
| POST | `/api/uploads` | JWT | Get presigned upload URL |
| POST | `/api/workspaces/:workspace_id/channels/:id/messages` | JWT | Send message (optional `attachments`) |
| GET | `/api/workspaces/:workspace_id/channels/:id/messages` | JWT | Message history |
| POST | `/api/workspaces/:workspace_id/channels/:id/messages/:message_id/attachments` | JWT | Attach files to message |
| GET | `/api/workspaces/:workspace_id/channels/:id/messages/:message_id/attachments` | JWT | List message attachments (presigned URLs) |
| GET | `/api/notifications` | JWT | List notifications (`?limit=&offset=`) |
| PATCH | `/api/notifications/:id/read` | JWT | Mark notification read |
| POST | `/api/notifications/read-all` | JWT | Mark all notifications read |
| GET | `/api/presence` | JWT | Globally online users (status, last_seen) |
| GET | `/api/workspaces/:workspace_id/presence` | JWT | Online members in workspace |
| GET | `/api/users/me` | JWT | Current user profile |
| PATCH | `/api/users/me` | JWT | Update username |
| PATCH | `/api/users/me/status` | JWT | Set presence status (`online`, `away`, `offline`) |
| POST | `/api/workspaces/:workspace_id/channels/:id/calls` | JWT | Initiate call (`mode`: `voice` or `video`) |
| POST | `/api/calls/:id/join` | JWT | Join call (channel members only) |
| POST | `/api/calls/:id/leave` | JWT | Leave call |
| POST | `/api/apps` | JWT | Create developer app (dedicated workspace) |
| POST | `/api/apps/:app_id/keys` | JWT | Create API key (secret shown once) |
| POST | `/api/apps/:app_id/webhooks` | JWT | Register webhook endpoint |
| POST | `/v1/users` | API key | Provision app-scoped user |
| POST | `/v1/channels` | API key | Create channel in app workspace |
| POST | `/v1/messages` | API key | Send message |
| GET | `/v1/messages?channel_id=` | API key | List channel messages |
| GET | `/api/ws?workspace_id=&channel_id=` | JWT (header or `?token=`) | WebSocket upgrade |

WebSocket auth accepts `Authorization: Bearer …` or `?token=` on the upgrade URL. Connection limits and allowed origins are configurable via env (see `.env.example`). Graceful shutdown closes active WebSocket connections on SIGTERM.

Voice and video call signaling (`call.offer`, `call.answer`, `call.ice`, `call.video.*`, etc.) is sent over the channel WebSocket. See [docs/WEBRTC.md](docs/WEBRTC.md) and [docs/WEBRTC_CLIENT.md](docs/WEBRTC_CLIENT.md).

**Developer API:** Authenticate with `Authorization: Bearer bx_live_...` or `X-API-Key`. Create apps/keys via JWT routes. See [docs/openapi.yaml](docs/openapi.yaml).

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
