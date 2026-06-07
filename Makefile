.PHONY: migration migrate run db-start sqlc build test test-integration test-integration-db

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


migrate:
	migrate -path sql/migrations -database "postgres://postgres:postgres@localhost:5432/brenox?sslmode=disable" up

run:
	go run cmd/api/main.go

db-start:
	docker compose -f docker-compose.dev.yaml up -d

stack:
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