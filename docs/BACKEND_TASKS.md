# Brenox Backend — Task Tracker

> **Purpose:** Track all backend work from current state through production-ready realtime communication platform.
>
> **Last updated:** 2026-07-07 (Phase 15 — Platform admin MVP)
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
| 0 | Stabilize & Unblock | 🟢 Complete | 8 / 8 |
| 1 | Messaging APIs | 🟢 Complete | 12 / 12 |
| 2 | Channel Join / Leave | 🟢 Complete | 8 / 8 |
| 3 | Workspace Architecture | 🟢 Complete | 14 / 14 |
| 4 | Permissions System | 🟢 Complete | 12 / 12 |
| 5 | Realtime Hardening | 🟢 Complete | 10 / 10 |
| 6 | Redis & Horizontal Scale | 🟢 Complete | 10 / 10 |
| 7 | Presence (Production) | 🟢 Complete | 8 / 8 |
| 8 | Notifications | 🟢 Complete | 10 / 10 |
| 9 | File Attachments | 🟢 Complete | 10 / 10 |
| 10 | WebRTC — Voice | 🟢 Complete | 12 / 12 |
| 11 | WebRTC — Video | 🟢 Complete | 8 / 8 |
| 12 | Public Developer API | 🟢 Complete | 12 / 12 |
| 13 | SDK Support Layer | 🟢 Complete | 8 / 8 |
| 14 | Production Readiness | 🟢 Complete | 18 / 18 |

**Legend:** 🔴 Not started · 🟡 In progress · 🟢 Complete

**Overall backend completion:** 100% (Phases 0–14 complete)

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
- [x] WebSocket event model (`internal/realtime/events.go`)
- [x] Docker Compose, Makefile (`sqlc`, `build`, `migrate`), `.env.example`
- [x] Kubernetes manifests — `deploy/` (Kustomize dev + prod overlays), `docs/KUBERNETES.md`
- [x] Presence counting + `GET /api/presence`
- [x] `chat.Service` — send/list messages with membership checks
- [x] Message REST APIs — `POST/GET /api/channels/:id/messages`
- [x] WebSocket `message.send` → persist → `message.new` broadcast
- [x] Channel membership enforced on messages and WebSocket connect
- [x] `IsChannelMember` sqlc query + `channels.Service.IsMember`
- [x] Channel join/leave — `POST /api/channels/:id/join|leave`
- [x] Realtime events — `member.joined`, `member.left`
- [x] Owner-leave policy — owner cannot leave (transfer ownership deferred to Phase 4)
- [x] Workspaces — `workspaces`, `workspace_members`, workspace-scoped channels/messages
- [x] `internal/workspaces` package + migration `000004_add_workspaces`
- [x] RBAC — `internal/authz`, `docs/PERMISSIONS.md`, migration `000005_permissions`
- [x] Workspace member admin APIs + read-only channels

### Known bugs / WIP (remaining)

- [ ] No tests anywhere in repo (authz unit tests added; integration tests still pending)

---

## Phase 0 — Stabilize & Unblock

**Goal:** Project builds, runs locally, presence WIP is correct, dev onboarding works.

**Exit criteria:** `go build ./...` passes; `make db-start && make migrate && make run` works with documented env vars.

| # | Task | Status |
|---|------|--------|
| 0.1 | Add `pingPeriod` (and related WS constants) to `internal/realtime` | [x] |
| 0.2 | Fix presence: increment `onlineUsers` on client register | [x] |
| 0.3 | Fix presence: only emit `presence.offline` when connection count reaches 0 | [x] |
| 0.4 | Register presence HTTP route — e.g. `GET /api/presence` | [x] |
| 0.5 | Create `.env.example` with `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `JWT_SECRET` | [x] |
| 0.6 | Update README: correct routes, folder layout, quick-start steps | [x] |
| 0.7 | Add Makefile target or script for sqlc codegen (`sqlc generate`) | [x] |
| 0.8 | Verify end-to-end smoke test: register → login → create channel → WS connect | [x] |

---

## Phase 1 — Messaging APIs

**Goal:** Close the persist → fetch → broadcast loop for chat messages.

**Exit criteria:** Messages saved to DB via REST and WebSocket; history retrievable with pagination; only channel members can send/read.

| # | Task | Status |
|---|------|--------|
| 1.1 | Create `internal/chat/handler.go` with message HTTP handlers | [x] |
| 1.2 | Wire `chat.Service` in `cmd/api/main.go` | [x] |
| 1.3 | `POST /api/channels/:id/messages` — create message (auth + membership check) | [x] |
| 1.4 | `GET /api/channels/:id/messages` — paginated history (`limit`, `offset` or cursor) | [x] |
| 1.5 | Add sqlc query: `IsChannelMember(channel_id, user_id)` (or equivalent) | [x] |
| 1.6 | Add membership validation helper in `channels` or shared `internal/authz` package | [x] |
| 1.7 | Define standard event type: `message.new` with payload (id, content, sender_id, created_at) | [x] |
| 1.8 | WebSocket: handle inbound `message.send` → persist via `chat.Service` → broadcast `message.new` | [x] |
| 1.9 | WebSocket: reject or ignore events from non-members | [x] |
| 1.10 | Input validation: max message length, non-empty content | [x] |
| 1.11 | Consistent JSON error responses across message endpoints | [x] |
| 1.12 | Manual/integration test: send via REST, receive via WS; send via WS, fetch via REST | [x] |

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
| 2.1 | sqlc query: `RemoveChannelMember` | [x] |
| 2.2 | sqlc query: `GetChannelMember` / membership lookup | [x] |
| 2.3 | `POST /api/channels/:id/join` — add user to `channel_members` | [x] |
| 2.4 | `POST /api/channels/:id/leave` — remove user (owner leave rules TBD) | [x] |
| 2.5 | Prevent duplicate membership (DB constraint already exists) | [x] |
| 2.6 | Define owner-leave policy (transfer ownership vs delete channel vs block) | [x] |
| 2.7 | Realtime events: `member.joined`, `member.left` | [x] |
| 2.8 | Enforce membership on WebSocket upgrade (return 403 before upgrade) | [x] |

---

## Phase 3 — Workspace Architecture

**Goal:** Introduce multi-tenant workspace layer. **Mandatory before production and public API.**

**Exit criteria:** All channels belong to a workspace; APIs scoped by workspace; existing data migrated or reset.

| # | Task | Status |
|---|------|--------|
| 3.1 | Migration: `workspaces` table (id, name, slug, owner_id, timestamps) | [x] |
| 3.2 | Migration: `workspace_members` table (workspace_id, user_id, role placeholder) | [x] |
| 3.3 | Migration: add `workspace_id` FK to `channels` | [x] |
| 3.4 | sqlc queries: create/list workspaces, add/list workspace members | [x] |
| 3.5 | `internal/workspaces` package (handler, service, types) | [x] |
| 3.6 | `POST /api/workspaces` — create workspace | [x] |
| 3.7 | `GET /api/workspaces` — list user's workspaces | [x] |
| 3.8 | `GET /api/workspaces/:id` — workspace detail | [x] |
| 3.9 | Update channel create to require `workspace_id` | [x] |
| 3.10 | Update channel list to filter by workspace | [x] |
| 3.11 | Scope message APIs under workspace (path or query param — decide and document) | [x] |
| 3.12 | Channel name uniqueness **per workspace** (not globally) | [x] |
| 3.13 | Data migration strategy for dev/staging (document in README) | [x] |
| 3.14 | Update WebSocket event envelope to include `workspace_id` | [x] |

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
| 4.1 | Migration: `roles` table (or enum: admin, moderator, member) | [x] |
| 4.2 | Migration: `workspace_member_roles` or role column on `workspace_members` | [x] |
| 4.3 | Migration: optional `channel_roles` for channel-specific overrides | [x] |
| 4.4 | Define permission matrix (document in `docs/PERMISSIONS.md`) | [x] |
| 4.5 | `internal/authz` package — `Can(user, action, resource)` | [x] |
| 4.6 | Gate: create/delete channel | [x] |
| 4.7 | Gate: invite/remove members | [x] |
| 4.8 | Gate: send messages (read-only channels) | [x] |
| 4.9 | `POST /api/workspaces/:id/members` — invite/add member (admin) | [x] |
| 4.10 | `DELETE /api/workspaces/:id/members/:user_id` — remove member | [x] |
| 4.11 | `PATCH /api/workspaces/:id/members/:user_id` — change role | [x] |
| 4.12 | Unit tests for permission checks | [x] |

---

## Phase 5 — Realtime Hardening

**Goal:** Production-quality WebSocket layer on a single node before Redis.

**Exit criteria:** Stable connections, documented event protocol, browser-friendly auth, no hub deadlocks.

| # | Task | Status |
|---|------|--------|
| 5.1 | Document full event catalog in `docs/WEBSOCKET_EVENTS.md` | [x] |
| 5.2 | Standardize event envelope: `type`, `channel_id`, `workspace_id`, `payload`, `timestamp`, `event_id` | [x] |
| 5.3 | WebSocket auth via query param `?token=` (in addition to Authorization header) | [x] |
| 5.4 | Restrict `CheckOrigin` to configurable allowed origins | [x] |
| 5.5 | Fix hub deadlock risk (don't send to `broadcast` from inside hub select without buffering) | [x] |
| 5.6 | Client `send` channel buffer size + backpressure policy | [x] |
| 5.7 | Graceful shutdown: drain hub, close WS connections on SIGTERM | [x] |
| 5.8 | Typing indicators: `typing.start`, `typing.stop` (ephemeral, no DB) | [x] |
| 5.9 | Connection limits per user / per IP (basic) | [x] |
| 5.10 | Structured logging for connect/disconnect/errors | [x] |

---

## Phase 6 — Redis & Horizontal Scale

**Goal:** Multiple API/WS instances share events via Redis Pub/Sub.

**Exit criteria:** Two server instances; user on instance A receives messages sent from instance B.

| # | Task | Status |
|---|------|--------|
| 6.1 | Add Redis to `docker-compose.dev.yaml` | [x] |
| 6.2 | Redis client package / config (`REDIS_URL`) | [x] |
| 6.3 | Publish all outbound realtime events to Redis channel(s) | [x] |
| 6.4 | Subscribe on each node; forward to local hub | [x] |
| 6.5 | Channel-scoped Redis topics: `workspace:{id}:channel:{id}` | [x] |
| 6.6 | Handle Redis reconnect / resubscribe | [x] |
| 6.7 | Integration test with 2 app instances + Redis | [x] |
| 6.8 | Document deployment topology in `docs/DEPLOYMENT.md` | [x] |
| 6.9 | Health check endpoint: `GET /health` (DB + Redis) | [x] |
| 6.10 | Sticky sessions vs shared state — document WS load balancer config | [x] |

---

## Phase 7 — Presence (Production)

**Goal:** Accurate, distributed presence with Redis as source of truth.

**Exit criteria:** Presence survives instance restarts; multi-tab counting works; last_seen available via API.

| # | Task | Status |
|---|------|--------|
| 7.1 | Redis keys: `presence:{user_id}` → status, connection_count, last_seen | [x] |
| 7.2 | Increment/decrement on WS connect/disconnect (all nodes) | [x] |
| 7.3 | `GET /api/presence` — global online users (or workspace-scoped) | [x] |
| 7.4 | `GET /api/workspaces/:id/presence` — workspace online members | [x] |
| 7.5 | Status updates: online, away, offline (`PATCH /api/users/me/status`) | [x] |
| 7.6 | Broadcast `presence.online`, `presence.offline`, `presence.status` events | [x] |
| 7.7 | TTL / heartbeat for stale presence cleanup | [x] |
| 7.8 | Remove in-memory-only presence maps from hub (or keep as local cache) | [x] |

---

## Phase 8 — Notifications

**Goal:** Event-driven notifications for mentions, replies, invites, calls.

**Exit criteria:** Notification records persisted; delivered via WS; push/email stubs or integrations ready.

| # | Task | Status |
|---|------|--------|
| 8.1 | Migration: `notifications` table | [x] |
| 8.2 | Notification types enum: mention, reply, channel_invite, workspace_invite, call_invite | [x] |
| 8.3 | `internal/notifications` service | [x] |
| 8.4 | `GET /api/notifications` — list with pagination | [x] |
| 8.5 | `PATCH /api/notifications/:id/read` — mark read | [x] |
| 8.6 | `POST /api/notifications/read-all` | [x] |
| 8.7 | Emit `notification.new` over WebSocket | [x] |
| 8.8 | @mention parsing in messages → create mention notifications | [x] |
| 8.9 | Push notification adapter interface (FCM/APNs — stub first) | [x] |
| 8.10 | Email notification adapter interface (stub first) | [x] |

---

## Phase 9 — File Attachments

**Goal:** Upload and attach files to messages; store in S3-compatible object storage.

**Exit criteria:** Upload flow works; attachments linked to messages; URLs served securely.

| # | Task | Status |
|---|------|--------|
| 9.1 | Migration: `attachments` table (id, message_id, file_url, mime_type, size, created_at) | [x] |
| 9.2 | S3-compatible storage client (MinIO for local dev) | [x] |
| 9.3 | Add MinIO to `docker-compose.dev.yaml` | [x] |
| 9.4 | `POST /api/uploads` — presigned URL or direct upload | [x] |
| 9.5 | Attach file to message (metadata in message or separate link) | [x] |
| 9.6 | Max file size validation | [x] |
| 9.7 | Allowed MIME type whitelist | [x] |
| 9.8 | `GET /api/messages/:id/attachments` | [x] |
| 9.9 | Realtime event: `message.updated` when attachment added | [x] |
| 9.10 | Virus scan hook (interface/stub for future) | [x] |

---

## Phase 10 — WebRTC — Voice Calling

**Goal:** Signaling server for voice calls; no media through backend.

**Exit criteria:** Two clients can establish voice call via signaling events; call rooms managed.

| # | Task | Status |
|---|------|--------|
| 10.1 | Migration: `calls`, `call_participants` tables | [x] |
| 10.2 | `POST /api/channels/:id/calls` — initiate call | [x] |
| 10.3 | `POST /api/calls/:id/join` — join call room | [x] |
| 10.4 | `POST /api/calls/:id/leave` — leave call | [x] |
| 10.5 | WebSocket events: `call.offer`, `call.answer`, `call.ice` | [x] |
| 10.6 | WebSocket events: `call.join`, `call.leave`, `call.end` | [x] |
| 10.7 | ICE candidate relay through hub/Redis | [x] |
| 10.8 | Call state machine (ringing, active, ended) | [x] |
| 10.9 | Permission: only channel members can join channel calls | [x] |
| 10.10 | Call invite notifications (Phase 8 integration) | [x] |
| 10.11 | TURN server config documentation (external service) | [x] |
| 10.12 | Integration test with mock SDP exchange | [x] |

---

## Phase 11 — WebRTC — Video Calling

**Goal:** Extend voice signaling for video, screen share hooks.

**Exit criteria:** Video call signaling works; screen share event types defined.

| # | Task | Status |
|---|------|--------|
| 11.1 | Extend call model for video vs voice mode | [x] |
| 11.2 | WebSocket events: `call.video.on`, `call.video.off` | [x] |
| 11.3 | WebSocket events: `call.screen.start`, `call.screen.stop` | [x] |
| 11.4 | Active speaker event: `call.speaker.changed` (optional) | [x] |
| 11.5 | Recording metadata table + start/stop signaling (not media storage) | [x] |
| 11.6 | Max participants per call config | [x] |
| 11.7 | Bandwidth/codec preferences in signaling (optional) | [x] |
| 11.8 | Document client-side WebRTC requirements for SDK team | [x] |

---

## Phase 12 — Public Developer API

**Goal:** Versioned public API for third-party apps with API key auth.

**Exit criteria:** External app can create users, channels, messages using API keys; rate limited.

| # | Task | Status |
|---|------|--------|
| 12.1 | Migration: `apps`, `api_keys` tables | [x] |
| 12.2 | API key generation, hashing, revocation | [x] |
| 12.3 | API key auth middleware (separate from user JWT) | [x] |
| 12.4 | Versioned router: `/v1/...` | [x] |
| 12.5 | `POST /v1/users` — provision user for app | [x] |
| 12.6 | `POST /v1/channels` | [x] |
| 12.7 | `POST /v1/messages` | [x] |
| 12.8 | `GET /v1/messages` | [x] |
| 12.9 | App-scoped workspaces (each app = tenant boundary) | [x] |
| 12.10 | Webhook delivery system (optional: `webhooks` table + dispatcher) | [x] |
| 12.11 | Idempotency-Key header support on write endpoints | [x] |
| 12.12 | OpenAPI spec generation — `docs/openapi.yaml` | [x] |

---

## Phase 13 — SDK Support Layer

**Goal:** Backend features required for official SDKs (JS, React, Flutter).

**Exit criteria:** SDK team can integrate without undocumented behavior; token refresh documented.

| # | Task | Status |
|---|------|--------|
| 13.1 | Token refresh endpoint — `POST /auth/refresh` | [x] |
| 13.2 | User profile endpoints — `GET/PATCH /api/users/me` | [x] |
| 13.3 | WebSocket reconnection + missed event strategy (document) | [x] |
| 13.4 | Server-sent event sequence numbers for gap detection | [x] |
| 13.5 | CORS configuration for browser SDK | [x] |
| 13.6 | SDK integration guide — `docs/SDK_INTEGRATION.md` | [x] |
| 13.7 | Sandbox/dev API keys for testing | [x] |
| 13.8 | Example curl/WebSocket flows in docs | [x] |

---

## Phase 14 — Production Readiness

**Goal:** Secure, observable, deployable backend ready for real traffic.

**Exit criteria:** All items below complete; security review done; CI green.

| # | Task | Status |
|---|------|--------|
| 14.1 | Rate limiting middleware (per IP + per API key) | [x] |
| 14.2 | Request size limits | [x] |
| 14.3 | Refresh token rotation + revocation list | [x] |
| 14.4 | Audit logging table + middleware for sensitive actions | [x] |
| 14.5 | Structured JSON logging (zerolog or slog) | [x] |
| 14.6 | Prometheus metrics endpoint | [x] |
| 14.7 | OpenTelemetry tracing (optional) | [x] |
| 14.8 | Dockerfile for API server | [x] |
| 14.9 | docker-compose production-like stack (API + Postgres + Redis + MinIO) | [x] |
| 14.10 | CI pipeline: lint, test, build (`GitHub Actions` or similar) | [x] |
| 14.11 | Unit tests — auth, channels, chat, authz | [x] |
| 14.12 | Integration tests — DB (Testcontainers) | [x] |
| 14.13 | Integration tests — WebSocket + Redis multi-node | [x] |
| 14.14 | Secrets management docs (no secrets in repo) | [x] |
| 14.15 | SQL injection review (sqlc mitigates; verify dynamic SQL none) | [x] |
| 14.16 | Security headers middleware | [x] |
| 14.17 | Load test baseline (k6 or vegeta) — document RPS targets | [x] |
| 14.18 | Runbook: `docs/RUNBOOK.md` (deploy, rollback, incident) | [x] |

---

## Phase 15 — Platform Admin

**Goal:** System admin panel API for cross-tenant operations.

**Exit criteria:** Platform roles on users; admin routes gated; suspend invalidates JWTs; env bootstrap for admins.

| # | Task | Status |
|---|------|--------|
| 15.1 | Migration: `users.platform_role`, `users.suspended_at`, `users.tokens_invalidated_at` | [x] |
| 15.2 | `PLATFORM_ADMIN_EMAILS` env bootstrap on login/register/me | [x] |
| 15.3 | Admin middleware + `GET /api/admin/overview`, users, workspaces, apps, audit-logs | [x] |
| 15.4 | `PATCH /api/admin/users/:id` — suspend/unsuspend, platform role (admin only) | [x] |
| 15.5 | Extend `GET /api/users/me` with `platform_role`, `suspended` | [x] |
| 15.6 | Suspend + token invalidation in JWT validation | [x] |
| 15.7 | Web: `/admin` panel (overview, users, workspaces, apps, audit) | [x] |

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
| POST | `/api/workspaces` | JWT | Working |
| GET | `/api/workspaces` | JWT | Working |
| GET | `/api/workspaces/:workspace_id` | JWT | Working |
| GET | `/api/workspaces/:workspace_id/members` | JWT | Working |
| POST | `/api/workspaces/:workspace_id/members` | JWT | Working — owner/admin |
| DELETE | `/api/workspaces/:workspace_id/members/:user_id` | JWT | Working — owner/admin |
| PATCH | `/api/workspaces/:workspace_id/members/:user_id` | JWT | Working — owner/admin |
| POST | `/api/workspaces/:workspace_id/channels` | JWT | Working — role gated |
| GET | `/api/workspaces/:workspace_id/channels` | JWT | Working |
| POST | `/api/workspaces/:workspace_id/channels/:id/join` | JWT | Working |
| POST | `/api/workspaces/:workspace_id/channels/:id/leave` | JWT | Working — owner blocked |
| POST | `/api/workspaces/:workspace_id/channels/:id/messages` | JWT | Working |
| GET | `/api/workspaces/:workspace_id/channels/:id/messages` | JWT | Working — paginated |
| GET | `/api/presence` | JWT | Working |
| GET | `/api/ws?workspace_id=&channel_id=` | JWT | Working — member only |

---

## Notes & Decisions Log

Record architectural decisions here as they are made.

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-06-06 | Channel owner cannot leave | Prevents orphaned channels; ownership transfer in Phase 4 |
| 2026-06-06 | Workspace-scoped API paths | `/api/workspaces/:workspace_id/...` replaces flat channel routes |
| 2026-06-06 | Channel names unique per workspace | DB index on `(workspace_id, name)` |
| 2026-06-06 | Workspace roles: owner/admin/moderator/member | Enforced via `internal/authz`; read-only channels for announcements |

---

## Changelog

| Date | Change |
|------|--------|
| 2026-07-07 | Health check reports pending migrations; startup logs error when schema is behind (fixes prod audit_logs / register failures) |
| 2026-07-07 | Phase 15 MVP: platform admin API (`/api/admin/*`), roles, suspend + JWT invalidation, web admin panel |
| 2026-07-07 | Sanitized HTTP and WebSocket error responses so internal infrastructure details stay server-side |
| 2026-07-05 | `POST /v1/channels` idempotent on duplicate name — returns existing channel in app workspace (embed demo bootstrap) |
| 2026-07-05 | `POST /v1/sessions` — issue user JWT for embed SDK clients; optional channel auto-join; BrenoxServer.sessions in SDK |
| 2026-07-05 | Engine release versioning: `GET /version`, prod API `https://api.breno-x.com` |
| 2026-07-03 | User security: `PATCH /api/users/me/password`, `GET /api/users/me/status` for dashboard profile |
| 2026-07-03 | Workspace list/get include member `role`; webhook/app routes for dashboard (prior entry) |
| 2026-06-22 | Rename prod/dev K8s + image naming: `brenox-api` → `brenox-engine` (GHCR, Deployment, Service, Ingress) |
| 2026-06-22 | Kubernetes deploy scaffold: `deploy/` Kustomize overlays (dev/prod), migrate image, `docs/KUBERNETES.md`, Makefile `k8s-*` targets |
| 2026-06-06 | Phase 4 complete: RBAC, member admin APIs, read-only channels, authz tests |
| 2026-06-06 | Phase 3 complete: workspaces, workspace-scoped routes, migration 000004 |
| 2026-06-06 | Phase 2 complete: join/leave APIs, member events, owner-leave policy |
| 2026-06-06 | Phase 1 complete: message REST APIs, WS message.send, membership checks, WEBSOCKET_EVENTS.md |
| 2026-06-06 | Phase 0 complete: build fix, presence, `.env.example`, `GET /api/presence`, Makefile `sqlc`/`build` |
| 2026-06-06 | Initial task tracker created from codebase audit + platform roadmap |
