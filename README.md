# Brenox — Chat Platform

Lightweight real-time chat server written in Go. Designed with clear separation between HTTP handlers, business services, and database repositories. Uses raw SQL and sqlc for queries and migrations for schema management.

## Repo layout

- `cmd/api` — application entrypoint
- `internal/` — business logic
  - `database/`, `handlers/`, `services/`, `repositories/`, `middleware/`, `config/`
- `sql/` — raw SQL: `queries/` and `migrations/`
- `db/` — DB integration helpers and generated/query code

## Quick start

Prerequisites: Go 1.20+, Docker & docker-compose (optional), PostgreSQL.

1. Copy `.env.example` to `.env` and configure DB connection.

2. Start a local DB (docker-compose):

```bash
docker-compose -f docker-compose.yml up -d db
```

3.Run migrations (using `psql` or your preferred tool):

```bash
psql "$DATABASE_URL" -f sql/migrations/000001_init_schema.up.sql
psql "$DATABASE_URL" -f sql/migrations/000002_create_channels.up.sql
psql "$DATABASE_URL" -f sql/migrations/000003_create_messages.up.sql
```

4.Build and run the API server:

```bash
go build ./cmd/api
./api
```

## Development notes

- Keep SQL in `sql/queries` and `sql/migrations` so schema and queries are explicit.
- Database access belongs in `internal/repositories` — business logic goes in `internal/services`.
- Handlers in `internal/handlers` should orchestrate request/response only.

## Database & tooling

- `sqlc.yaml` is present to generate type-safe DB code from SQL queries.
- Migrations are in `sql/migrations` — add matching `up.sql` and `down.sql` files for each change.

## Testing

- Add unit tests alongside packages. For DB integration tests, run them against a disposable Postgres (Testcontainers or a docker-compose DB).

## Contributing

- Open issues or PRs on the `main` branch. Keep changes focused and add tests for new behavior.

## License

This repository does not include a license file. Add one if you plan to publish.

---

If you want, I can also:

- Add a `.env.example` with recommended variables.
- Add a Makefile target to run migrations and start the app.
- Generate `sqlc` bindings and show how to run them.
