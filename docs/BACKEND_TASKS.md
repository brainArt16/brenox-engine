# Brenox Backend — Task Tracker

> **Purpose:** Track all backend work from current state through production-ready realtime communication platform.
>
> **Last updated:** 2026-06-06
>
> **How to use:** Check off tasks as completed. Update status tags and the progress summary at the top after each sprint.

---

## Documentation Agents

Three agents keep docs in sync with code. Config: `AGENTS.md`, `.cursor/rules/`, `.cursor/hooks.json`.

| Agent | Owns | When to run |
|-------|------|-------------|
| **Task Tracker** | `docs/BACKEND_TASKS.md` | After completing any task below |
| **README** | `README.md` | Routes, setup, env vars, or layout changed |
| **Postman** | `docs/postman/*` | HTTP API added, changed, or removed |

**Automatic:** Hooks track edits to `internal/`, `cmd/`, `sql/`, `pkg/` and remind on agent stop to run all three.

**Manual:** Ask — *Sync documentation: update BACKEND_TASKS, README, and Postman.*

---

## Progress Summary

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 0 | Stabilize & Unblock | 🔴 Not started | 0 / 8 |
| 1 | Messaging APIs | 🔴 Not started | 0 / 12 |
| 2 | Channel Join / Leave | 🔴 Not started | 0 / 8 |
| 3 | Workspace Architecture | 🔴 Not started | 0 / 14 |
| 4 | Permissions System | 🔴 Not started | 0 / 12 |
| 5 | Realtime Hardening | 🔴 Not started | 0 / 10 |
| 6 | Redis & Horizontal Scale | 🔴 Not started | 0 / 10 |
| 7 | Presence (Production) | 🔴 Not started | 0 / 8 |
| 8 | Notifications | 🔴 Not started | 0 / 10 |
| 9 | File Attachments | 🔴 Not started | 0 / 10 |
| 10 | WebRTC — Voice | 🔴 Not started | 0 / 12 |
| 11 | WebRTC — Video | 🔴 Not started | 0 / 8 |
| 12 | Public Developer API | 🔴 Not started | 0 / 12 |
| 13 | SDK Support Layer | 🔴 Not started | 0 / 8 |
| 14 | Production Readiness | 🔴 Not started | 0 / 18 |

**Legend:** 🔴 Not started · 🟡 In progress · 🟢 Complete

**Overall backend completion:** ~8% (foundation only)

---

## Status Tags (per task)

Use inline tags when updating:

- `[ ]` — Not started
- `[~]` — In progress
- `[x]` — Done
- `[!]` — Blocked (add reason in notes column if using external tracker)

---

## Already Implemented (Baseline)

These exist in the repo today. Do not re-implement; extend or fix as noted.

- [x] Go module + Gin HTTP server (`cmd/api/main.go`)
- [x] PostgreSQL connection pool (`internal/database/postgres.go`)
- [x] sqlc query layer (`internal/db/`, `sql/queries/`, `sqlc.yaml`)
- [x] Database migrations: `users`, `channels`, `channel_members`, `messages`
- [x] User registration — `POST /auth/register`
- [x] User login + JWT — `POST /auth/login`
- [x] JWT auth middleware (`internal/middleware/auth.go`)
- [x] Create channel — `POST /api/channels` (creator auto-added as member)
- [x] List user channels — `GET /api/channels`
- [x] WebSocket hub — `GET /api/ws?channel_id=` (in-memory, single node)
- [x] WebSocket event model (`internal/realtime/message.go`)
- [x] `chat.Service.SaveMessage` (exists but **not wired**)
- [x] sqlc `GetChannelMessages` query (exists but **no HTTP route**)
- [x] Docker Compose for local Postgres (`docker-compose.dev.yaml`)
- [x] Makefile targets: `migration`, `migrate`, `run`, `db-start`

### Known bugs / WIP (fix in Phase 0)

- [ ] `pingPeriod` undefined — **project does not build**
- [ ] `onlineUsers` never incremented on WebSocket connect
- [ ] `GetPresence` handler written but not registered in router
- [ ] WebSocket has no channel membership validation
- [ ] Messages not persisted from WebSocket events
- [ ] No `.env.example`
- [ ] No tests anywhere in repo
- [ ] README route paths and folder layout are outdated

---

## Phase 0 — Stabilize & Unblock

**Goal:** Project builds, runs locally, presence WIP is correct, dev onboarding works.

**Exit criteria:** `go build ./...` passes; `make db-start && make migrate && make run` works with documented env vars.

| # | Task | Status |
|---|------|--------|
| 0.1 | Add `pingPeriod` (and related WS constants) to `internal/realtime` | [ ] |
| 0.2 | Fix presence: increment `onlineUsers` on client register | [ ] |
| 0.3 | Fix presence: only emit `presence.offline` when connection count reaches 0 | [ ] |
| 0.4 | Register presence HTTP route — e.g. `GET /api/presence` | [ ] |
| 0.5 | Create `.env.example` with `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `JWT_SECRET` | [ ] |
| 0.6 | Update README: correct routes, folder layout, quick-start steps | [ ] |
| 0.7 | Add Makefile target or script for sqlc codegen (`sqlc generate`) | [ ] |
| 0.8 | Verify end-to-end smoke test: register → login → create channel → WS connect | [ ] |

---

## Phase 1 — Messaging APIs

**Goal:** Close the persist → fetch → broadcast loop for chat messages.

**Exit criteria:** Messages saved to DB via REST and WebSocket; history retrievable with pagination; only channel members can send/read.

| # | Task | Status |
|---|------|--------|
| 1.1 | Create `internal/chat/handler.go` with message HTTP handlers | [ ] |
| 1.2 | Wire `chat.Service` in `cmd/api/main.go` | [ ] |
| 1.3 | `POST /api/channels/:id/messages` — create message (auth + membership check) | [ ] |
| 1.4 | `GET /api/channels/:id/messages` — paginated history (`limit`, `offset` or cursor) | [ ] |
| 1.5 | Add sqlc query: `IsChannelMember(channel_id, user_id)` (or equivalent) | [ ] |
| 1.6 | Add membership validation helper in `channels` or shared `internal/authz` package | [ ] |
| 1.7 | Define standard event type: `message.new` with payload (id, content, sender_id, created_at) | [ ] |
| 1.8 | WebSocket: handle inbound `message.send` → persist via `chat.Service` → broadcast `message.new` | [ ] |
| 1.9 | WebSocket: reject or ignore events from non-members | [ ] |
| 1.10 | Input validation: max message length, non-empty content | [ ] |
| 1.11 | Consistent JSON error responses across message endpoints | [ ] |
| 1.12 | Manual/integration test: send via REST, receive via WS; send via WS, fetch via REST | [ ] |

### API contract (target)

```http
POST /api/channels/:id/messages
Authorization: Bearer <token>
Content-Type: application/json

{ "content": "hello" }
```

```http
GET /api/channels/:id/messages?limit=50&offset=0
Authorization: Bearer <token>
```

```json
// WebSocket inbound
{ "type": "message.send", "channel_id": 1, "payload": { "content": "hello" } }

// WebSocket outbound
{ "type": "message.new", "channel_id": 1, "payload": { "id": 1, "sender_id": 2, "content": "hello", "created_at": "..." } }
```

---

## Phase 2 — Channel Join / Leave

**Goal:** Users can join and leave channels; membership changes emit realtime events.

**Exit criteria:** Join/leave REST APIs work; duplicate joins prevented; WS connect denied for non-members.

| # | Task | Status |
|---|------|--------|
| 2.1 | sqlc query: `RemoveChannelMember` | [ ] |
| 2.2 | sqlc query: `GetChannelMember` / membership lookup | [ ] |
| 2.3 | `POST /api/channels/:id/join` — add user to `channel_members` | [ ] |
| 2.4 | `POST /api/channels/:id/leave` — remove user (owner leave rules TBD) | [ ] |
| 2.5 | Prevent duplicate membership (DB constraint already exists) | [ ] |
| 2.6 | Define owner-leave policy (transfer ownership vs delete channel vs block) | [ ] |
| 2.7 | Realtime events: `member.joined`, `member.left` | [ ] |
| 2.8 | Enforce membership on WebSocket upgrade (return 403 before upgrade) | [ ] |

---

## Phase 3 — Workspace Architecture

**Goal:** Introduce multi-tenant workspace layer. **Mandatory before production and public API.**

**Exit criteria:** All channels belong to a workspace; APIs scoped by workspace; existing data migrated or reset.

| # | Task | Status |
|---|------|--------|
| 3.1 | Migration: `workspaces` table (id, name, slug, owner_id, timestamps) | [ ] |
| 3.2 | Migration: `workspace_members` table (workspace_id, user_id, role placeholder) | [ ] |
| 3.3 | Migration: add `workspace_id` FK to `channels` | [ ] |
| 3.4 | sqlc queries: create/list workspaces, add/list workspace members | [ ] |
| 3.5 | `internal/workspaces` package (handler, service, types) | [ ] |
| 3.6 | `POST /api/workspaces` — create workspace | [ ] |
| 3.7 | `GET /api/workspaces` — list user's workspaces | [ ] |
| 3.8 | `GET /api/workspaces/:id` — workspace detail | [ ] |
| 3.9 | Update channel create to require `workspace_id` | [ ] |
| 3.10 | Update channel list to filter by workspace | [ ] |
| 3.11 | Scope message APIs under workspace (path or query param — decide and document) | [ ] |
| 3.12 | Channel name uniqueness **per workspace** (not globally) | [ ] |
| 3.13 | Data migration strategy for dev/staging (document in README) | [ ] |
| 3.14 | Update WebSocket event envelope to include `workspace_id` | [ ] |

### Target data model

```text
users
  ↓
workspace_members
  ↓
workspaces
  ↓
channels (+ channel_members)
  ↓
messages
```

---

## Phase 4 — Permissions System

**Goal:** Role-based access control at workspace and channel level.

**Exit criteria:** Actions gated by role; admin can manage members; moderators can moderate channels.

| # | Task | Status |
|---|------|--------|
| 4.1 | Migration: `roles` table (or enum: admin, moderator, member) | [ ] |
| 4.2 | Migration: `workspace_member_roles` or role column on `workspace_members` | [ ] |
| 4.3 | Migration: optional `channel_roles` for channel-specific overrides | [ ] |
| 4.4 | Define permission matrix (document in `docs/PERMISSIONS.md`) | [ ] |
| 4.5 | `internal/authz` package — `Can(user, action, resource)` | [ ] |
| 4.6 | Gate: create/delete channel | [ ] |
| 4.7 | Gate: invite/remove members | [ ] |
| 4.8 | Gate: send messages (read-only channels) | [ ] |
| 4.9 | `POST /api/workspaces/:id/members` — invite/add member (admin) | [ ] |
| 4.10 | `DELETE /api/workspaces/:id/members/:user_id` — remove member | [ ] |
| 4.11 | `PATCH /api/workspaces/:id/members/:user_id` — change role | [ ] |
| 4.12 | Unit tests for permission checks | [ ] |

---

## Phase 5 — Realtime Hardening

**Goal:** Production-quality WebSocket layer on a single node before Redis.

**Exit criteria:** Stable connections, documented event protocol, browser-friendly auth, no hub deadlocks.

| # | Task | Status |
|---|------|--------|
| 5.1 | Document full event catalog in `docs/WEBSOCKET_EVENTS.md` | [ ] |
| 5.2 | Standardize event envelope: `type`, `channel_id`, `workspace_id`, `payload`, `timestamp`, `event_id` | [ ] |
| 5.3 | WebSocket auth via query param `?token=` (in addition to Authorization header) | [ ] |
| 5.4 | Restrict `CheckOrigin` to configurable allowed origins | [ ] |
| 5.5 | Fix hub deadlock risk (don't send to `broadcast` from inside hub select without buffering) | [ ] |
| 5.6 | Client `send` channel buffer size + backpressure policy | [ ] |
| 5.7 | Graceful shutdown: drain hub, close WS connections on SIGTERM | [ ] |
| 5.8 | Typing indicators: `typing.start`, `typing.stop` (ephemeral, no DB) | [ ] |
| 5.9 | Connection limits per user / per IP (basic) | [ ] |
| 5.10 | Structured logging for connect/disconnect/errors | [ ] |

---

## Phase 6 — Redis & Horizontal Scale

**Goal:** Multiple API/WS instances share events via Redis Pub/Sub.

**Exit criteria:** Two server instances; user on instance A receives messages sent from instance B.

| # | Task | Status |
|---|------|--------|
| 6.1 | Add Redis to `docker-compose.dev.yaml` | [ ] |
| 6.2 | Redis client package / config (`REDIS_URL`) | [ ] |
| 6.3 | Publish all outbound realtime events to Redis channel(s) | [ ] |
| 6.4 | Subscribe on each node; forward to local hub | [ ] |
| 6.5 | Channel-scoped Redis topics: `workspace:{id}:channel:{id}` | [ ] |
| 6.6 | Handle Redis reconnect / resubscribe | [ ] |
| 6.7 | Integration test with 2 app instances + Redis | [ ] |
| 6.8 | Document deployment topology in `docs/DEPLOYMENT.md` | [ ] |
| 6.9 | Health check endpoint: `GET /health` (DB + Redis) | [ ] |
| 6.10 | Sticky sessions vs shared state — document WS load balancer config | [ ] |

---

## Phase 7 — Presence (Production)

**Goal:** Accurate, distributed presence with Redis as source of truth.

**Exit criteria:** Presence survives instance restarts; multi-tab counting works; last_seen available via API.

| # | Task | Status |
|---|------|--------|
| 7.1 | Redis keys: `presence:{user_id}` → status, connection_count, last_seen | [ ] |
| 7.2 | Increment/decrement on WS connect/disconnect (all nodes) | [ ] |
| 7.3 | `GET /api/presence` — global online users (or workspace-scoped) | [ ] |
| 7.4 | `GET /api/workspaces/:id/presence` — workspace online members | [ ] |
| 7.5 | Status updates: online, away, offline (`PATCH /api/users/me/status`) | [ ] |
| 7.6 | Broadcast `presence.online`, `presence.offline`, `presence.status` events | [ ] |
| 7.7 | TTL / heartbeat for stale presence cleanup | [ ] |
| 7.8 | Remove in-memory-only presence maps from hub (or keep as local cache) | [ ] |

---

## Phase 8 — Notifications

**Goal:** Event-driven notifications for mentions, replies, invites, calls.

**Exit criteria:** Notification records persisted; delivered via WS; push/email stubs or integrations ready.

| # | Task | Status |
|---|------|--------|
| 8.1 | Migration: `notifications` table | [ ] |
| 8.2 | Notification types enum: mention, reply, channel_invite, workspace_invite, call_invite | [ ] |
| 8.3 | `internal/notifications` service | [ ] |
| 8.4 | `GET /api/notifications` — list with pagination | [ ] |
| 8.5 | `PATCH /api/notifications/:id/read` — mark read | [ ] |
| 8.6 | `POST /api/notifications/read-all` | [ ] |
| 8.7 | Emit `notification.new` over WebSocket | [ ] |
| 8.8 | @mention parsing in messages → create mention notifications | [ ] |
| 8.9 | Push notification adapter interface (FCM/APNs — stub first) | [ ] |
| 8.10 | Email notification adapter interface (stub first) | [ ] |

---

## Phase 9 — File Attachments

**Goal:** Upload and attach files to messages; store in S3-compatible object storage.

**Exit criteria:** Upload flow works; attachments linked to messages; URLs served securely.

| # | Task | Status |
|---|------|--------|
| 9.1 | Migration: `attachments` table (id, message_id, file_url, mime_type, size, created_at) | [ ] |
| 9.2 | S3-compatible storage client (MinIO for local dev) | [ ] |
| 9.3 | Add MinIO to `docker-compose.dev.yaml` | [ ] |
| 9.4 | `POST /api/uploads` — presigned URL or direct upload | [ ] |
| 9.5 | Attach file to message (metadata in message or separate link) | [ ] |
| 9.6 | Max file size validation | [ ] |
| 9.7 | Allowed MIME type whitelist | [ ] |
| 9.8 | `GET /api/messages/:id/attachments` | [ ] |
| 9.9 | Realtime event: `message.updated` when attachment added | [ ] |
| 9.10 | Virus scan hook (interface/stub for future) | [ ] |

---

## Phase 10 — WebRTC — Voice Calling

**Goal:** Signaling server for voice calls; no media through backend.

**Exit criteria:** Two clients can establish voice call via signaling events; call rooms managed.

| # | Task | Status |
|---|------|--------|
| 10.1 | Migration: `calls`, `call_participants` tables | [ ] |
| 10.2 | `POST /api/channels/:id/calls` — initiate call | [ ] |
| 10.3 | `POST /api/calls/:id/join` — join call room | [ ] |
| 10.4 | `POST /api/calls/:id/leave` — leave call | [ ] |
| 10.5 | WebSocket events: `call.offer`, `call.answer`, `call.ice` | [ ] |
| 10.6 | WebSocket events: `call.join`, `call.leave`, `call.end` | [ ] |
| 10.7 | ICE candidate relay through hub/Redis | [ ] |
| 10.8 | Call state machine (ringing, active, ended) | [ ] |
| 10.9 | Permission: only channel members can join channel calls | [ ] |
| 10.10 | Call invite notifications (Phase 8 integration) | [ ] |
| 10.11 | TURN server config documentation (external service) | [ ] |
| 10.12 | Integration test with mock SDP exchange | [ ] |

---

## Phase 11 — WebRTC — Video Calling

**Goal:** Extend voice signaling for video, screen share hooks.

**Exit criteria:** Video call signaling works; screen share event types defined.

| # | Task | Status |
|---|------|--------|
| 11.1 | Extend call model for video vs voice mode | [ ] |
| 11.2 | WebSocket events: `call.video.on`, `call.video.off` | [ ] |
| 11.3 | WebSocket events: `call.screen.start`, `call.screen.stop` | [ ] |
| 11.4 | Active speaker event: `call.speaker.changed` (optional) | [ ] |
| 11.5 | Recording metadata table + start/stop signaling (not media storage) | [ ] |
| 11.6 | Max participants per call config | [ ] |
| 11.7 | Bandwidth/codec preferences in signaling (optional) | [ ] |
| 11.8 | Document client-side WebRTC requirements for SDK team | [ ] |

---

## Phase 12 — Public Developer API

**Goal:** Versioned public API for third-party apps with API key auth.

**Exit criteria:** External app can create users, channels, messages using API keys; rate limited.

| # | Task | Status |
|---|------|--------|
| 12.1 | Migration: `apps`, `api_keys` tables | [ ] |
| 12.2 | API key generation, hashing, revocation | [ ] |
| 12.3 | API key auth middleware (separate from user JWT) | [ ] |
| 12.4 | Versioned router: `/v1/...` | [ ] |
| 12.5 | `POST /v1/users` — provision user for app | [ ] |
| 12.6 | `POST /v1/channels` | [ ] |
| 12.7 | `POST /v1/messages` | [ ] |
| 12.8 | `GET /v1/messages` | [ ] |
| 12.9 | App-scoped workspaces (each app = tenant boundary) | [ ] |
| 12.10 | Webhook delivery system (optional: `webhooks` table + dispatcher) | [ ] |
| 12.11 | Idempotency-Key header support on write endpoints | [ ] |
| 12.12 | OpenAPI spec generation — `docs/openapi.yaml` | [ ] |

---

## Phase 13 — SDK Support Layer

**Goal:** Backend features required for official SDKs (JS, React, Flutter).

**Exit criteria:** SDK team can integrate without undocumented behavior; token refresh documented.

| # | Task | Status |
|---|------|--------|
| 13.1 | Token refresh endpoint — `POST /auth/refresh` | [ ] |
| 13.2 | User profile endpoints — `GET/PATCH /api/users/me` | [ ] |
| 13.3 | WebSocket reconnection + missed event strategy (document) | [ ] |
| 13.4 | Server-sent event sequence numbers for gap detection | [ ] |
| 13.5 | CORS configuration for browser SDK | [ ] |
| 13.6 | SDK integration guide — `docs/SDK_INTEGRATION.md` | [ ] |
| 13.7 | Sandbox/dev API keys for testing | [ ] |
| 13.8 | Example curl/WebSocket flows in docs | [ ] |

---

## Phase 14 — Production Readiness

**Goal:** Secure, observable, deployable backend ready for real traffic.

**Exit criteria:** All items below complete; security review done; CI green.

| # | Task | Status |
|---|------|--------|
| 14.1 | Rate limiting middleware (per IP + per API key) | [ ] |
| 14.2 | Request size limits | [ ] |
| 14.3 | Refresh token rotation + revocation list | [ ] |
| 14.4 | Audit logging table + middleware for sensitive actions | [ ] |
| 14.5 | Structured JSON logging (zerolog or slog) | [ ] |
| 14.6 | Prometheus metrics endpoint | [ ] |
| 14.7 | OpenTelemetry tracing (optional) | [ ] |
| 14.8 | Dockerfile for API server | [ ] |
| 14.9 | docker-compose production-like stack (API + Postgres + Redis + MinIO) | [ ] |
| 14.10 | CI pipeline: lint, test, build (`GitHub Actions` or similar) | [ ] |
| 14.11 | Unit tests — auth, channels, chat, authz | [ ] |
| 14.12 | Integration tests — DB (Testcontainers) | [ ] |
| 14.13 | Integration tests — WebSocket + Redis multi-node | [ ] |
| 14.14 | Secrets management docs (no secrets in repo) | [ ] |
| 14.15 | SQL injection review (sqlc mitigates; verify dynamic SQL none) | [ ] |
| 14.16 | Security headers middleware | [ ] |
| 14.17 | Load test baseline (k6 or vegeta) — document RPS targets | [ ] |
| 14.18 | Runbook: `docs/RUNBOOK.md` (deploy, rollback, incident) | [ ] |

---

## Dependency Graph

```text
Phase 0 (Stabilize)
    ↓
Phase 1 (Messaging) ──────────────────────────────┐
    ↓                                              │
Phase 2 (Join/Leave)                               │
    ↓                                              │
Phase 3 (Workspaces) ← structural pivot            │
    ↓                                              │
Phase 4 (Permissions)                              │
    ↓                                              │
Phase 5 (Realtime Hardening)                       │
    ↓                                              │
Phase 6 (Redis) ← required for multi-instance      │
    ↓                                              │
Phase 7 (Presence)                                   │
    ↓                                              │
Phase 8 (Notifications) ←──────────────────────────┤
    ↓                                              │
Phase 9 (Attachments)                              │
    ↓                                              │
Phase 10 (Voice) ──→ Phase 11 (Video)              │
    ↓                                              │
Phase 12 (Public API)                              │
    ↓                                              │
Phase 13 (SDK Support)                             │
    ↓                                              │
Phase 14 (Production) ←────────────────────────────┘
```

**Parallelizable after Phase 6:** Phase 9 (Attachments) can run parallel to Phase 8 (Notifications).

---

## Recommended Sprint Plan (High Level)

| Sprint | Phases | Focus |
|--------|--------|-------|
| Sprint 1 | 0 + 1 | Build fix, messaging REST + WS persistence |
| Sprint 2 | 2 | Join/leave, WS membership enforcement |
| Sprint 3 | 3 | Workspace migration + APIs |
| Sprint 4 | 4 + 5 | Permissions + realtime hardening |
| Sprint 5 | 6 + 7 | Redis cluster + production presence |
| Sprint 6 | 8 + 9 | Notifications + attachments |
| Sprint 7 | 10 + 11 | Voice + video signaling |
| Sprint 8 | 12 + 13 | Public API + SDK docs |
| Sprint 9 | 14 | Production hardening + CI/CD |

---

## Current API Reference (As Built)

| Method | Route | Auth | Notes |
|--------|-------|------|-------|
| POST | `/auth/register` | No | Working |
| POST | `/auth/login` | No | Returns JWT |
| POST | `/api/channels` | JWT | Working |
| GET | `/api/channels` | JWT | Working |
| GET | `/api/ws?channel_id=` | JWT | WIP — no membership check |

---

## Notes & Decisions Log

Record architectural decisions here as they are made.

| Date | Decision | Rationale |
|------|----------|-----------|
| — | — | — |

---

## Changelog

| Date | Change |
|------|--------|
| 2026-06-06 | Initial task tracker created from codebase audit + platform roadmap |
