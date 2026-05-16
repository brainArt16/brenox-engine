chat-platform/
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── database/
│   ├── handlers/
│   ├── services/
│   ├── repositories/
│   ├── middleware/
│   └── config/
│
├── sql/
│   ├── queries/
│   └── migrations/
│
├── db/
│   ├── sqlc.go
│   ├── models.go
│   └── querier.go
│
├── .env
├── go.mod
├── sqlc.yaml
└── docker-compose.yml

## Why This Structure?

- cmd/

Contains application entry points.
This is where app execution starts.

- internal/

Contains application business logic.
Go convention:

internal/

means:

private application code

- repositories/

Responsible ONLY for database access.

Important principle:

Database logic should not leak everywhere.
services/

Contains business logic.

- handlers/

Responsible ONLY for HTTP handling.

- sql/

Contains raw SQL.

This is important because:

SQL becomes first-class
easier optimization
easier debugging