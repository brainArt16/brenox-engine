.PHONY: migration migrate run db-start sqlc build

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

sqlc:
	sqlc generate

build:
	go build ./...