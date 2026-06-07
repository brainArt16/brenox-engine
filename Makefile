.PHONY: migration migrate migrate-dev run dev-up dev-down dev-logs db-start sqlc build test test-integration test-integration-db

# Prevent make treating the migration name as a file/target
%:
	@:

migration:
	@NAME="$(word 2,$(MAKECMDGOALS))"; \
	if [ "$$NAME" = "" ]; then \
		echo "Usage: make migration <name>"; exit 1; \
	fi; \
	echo "Creating migration $$NAME"; \
	migrate create -ext sql -dir sql/migrations -seq $$NAME

# Run migrations against dev Postgres (host port 5432)
migrate:
	migrate -path sql/migrations -database "postgres://postgres:postgres@localhost:5432/brenox?sslmode=disable" up

# Run migrations via Docker (uses internal network)
migrate-dev:
	docker compose -f docker-compose.dev.yaml run --rm migrate

run:
	go run cmd/api/main.go

# Full dev stack: API + Postgres + Redis + MinIO + migrations
dev-up:
	docker compose -f docker-compose.dev.yaml up -d --build

dev-down:
	docker compose -f docker-compose.dev.yaml down

dev-logs:
	docker compose -f docker-compose.dev.yaml logs -f api

# Alias for dev-up
db-start: dev-up

stack:
	@test -f .env || (echo "Copy .env.docker.example to .env and set secrets first" && exit 1)
	docker compose up -d --build

sqlc:
	sqlc generate

build:
	go build ./...

test:
	go test ./...

test-integration:
	REDIS_URL=redis://localhost:6379/0 go test ./internal/realtime/ -run TestRedisBrokerCrossInstance -count=1

test-integration-db:
	DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=brenox JWT_SECRET=test \
	go test ./internal/integration/ -count=1
