# Brenox — Agent Roles

This repo uses specialized agent roles to keep code, task tracking, README, and Postman in sync.

Read this file when finishing backend work or when a hook triggers a documentation sync.

---

## When to Run Documentation Agents

Run **all three agents** after:

- Completing a task from `docs/BACKEND_TASKS.md`
- Adding or changing HTTP routes in `cmd/api/main.go`
- Adding handlers, services, migrations, or sqlc queries
- Changing request/response shapes or auth requirements

**Automatic reminder:** Cursor hooks track edits under `internal/`, `cmd/`, `sql/`, and `pkg/`. On agent stop, a follow-up prompt runs the sync workflow if backend files changed.

**Manual invocation:** Ask the agent:

> Sync documentation: update BACKEND_TASKS, README, and Postman for the work we just completed.

---

## Agent 1 — Task Tracker

**Owns:** `docs/BACKEND_TASKS.md`

**Trigger:** Any completed backend task or bug fix tied to the roadmap.

**Checklist:**

1. Find the matching task ID (e.g. `0.1`, `1.3`) and mark `[x]`.
2. Update the phase progress count (e.g. `3 / 8`).
3. Set phase status: 🔴 → 🟡 (in progress) → 🟢 (all tasks done).
4. Recalculate **Overall backend completion** if a phase completed.
5. Move items from "Known bugs / WIP" to "Already Implemented" when fixed.
6. Add a row to **Changelog** with date and summary.
7. Log non-obvious decisions in **Notes & Decisions Log**.

**Do not** mark tasks complete unless the code actually implements the exit criteria.

---

## Agent 2 — README

**Owns:** `README.md`

**Trigger:** Routes, env vars, setup steps, folder layout, or architecture changed.

**Checklist:**

1. Keep **Quick start** accurate (`make db-start`, `make migrate`, `make run`).
2. Update **Repo layout** to match actual packages (`internal/auth`, `internal/channels`, etc.).
3. Maintain an **API overview** table (method, path, auth, description).
4. Link to `docs/BACKEND_TASKS.md` and `docs/postman/`.
5. Remove stale "If you want, I can also…" placeholder sections.
6. Document required env vars (mirror `.env.example` when it exists).

**Do not** duplicate full API docs — Postman owns request examples; README owns orientation.

---

## Agent 3 — Postman

**Owns:**

- `docs/postman/brenox.postman_collection.json`
- `docs/postman/brenox.postman_environment.json`
- `docs/postman/README.md`

**Trigger:** Any new, changed, or removed HTTP endpoint.

**Checklist:**

1. Add/update/remove requests to match `cmd/api/main.go` routes exactly.
2. Use collection variables: `{{baseUrl}}`, `{{token}}`, `{{channelId}}`.
3. Auth requests: no token. Protected routes: `Authorization: Bearer {{token}}`.
4. Login request: add test script to save token:
   ```javascript
   const json = pm.response.json();
   if (json.token) pm.environment.set("token", json.token);
   ```
5. Group requests: **Auth**, **Channels**, **Messages**, **Presence**, **WebSocket** (description only).
6. Each request: name, method, URL, headers, example body, short description.
7. Update `docs/postman/README.md` import steps if files change.

**WebSocket:** Document in collection description; Postman WS requests are optional until endpoints stabilize.

---

## Code Quality (All Implementation Agents)

Applies to every Go change. See `.cursor/rules/go-backend-standards.mdc`.

- Handlers: parse, validate, respond — no business logic.
- Services: business rules and orchestration — no HTTP types.
- Document non-obvious behavior only; prefer self-explanatory names.
- Match existing package layout: one domain per `internal/<domain>/` folder.
- Leave a brief comment only when intent is not obvious from code.

---

## File Map

| File | Owner agent |
|------|-------------|
| `docs/BACKEND_TASKS.md` | Task Tracker |
| `README.md` | README |
| `docs/postman/*` | Postman |
| `internal/**/*.go` | Implementation (+ inline docs when needed) |

---

## Cursor Configuration

| Path | Purpose |
|------|---------|
| `.cursor/rules/docs-sync-workflow.mdc` | Master workflow (always on) |
| `.cursor/rules/task-tracker-agent.mdc` | Task tracker rules |
| `.cursor/rules/readme-agent.mdc` | README rules |
| `.cursor/rules/postman-agent.mdc` | Postman rules |
| `.cursor/rules/go-backend-standards.mdc` | Go code standards |
| `.cursor/hooks.json` | Auto-remind on backend edits |
