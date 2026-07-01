.PHONY: migration migrate migrate-dev run dev-up dev-down dev-logs db-start sqlc build test test-integration test-integration-db
.PHONY: test-up test-down test-logs migrate-test prod-up prod-down prod-logs
.PHONY: k8s-build k8s-load-kind k8s-dev-up k8s-dev-down k8s-migrate k8s-port-forward k8s-ingress-kind k8s-check k8s-cluster-create

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

# Test / staging — managed DB/Redis/S3 by default; see .env.test.example
test-up:
	@test -f .env.test || (echo "Copy .env.test.example to .env.test and set secrets first" && exit 1)
	docker compose -f docker-compose.test.yaml --env-file .env.test up -d --build

test-down:
	docker compose -f docker-compose.test.yaml --env-file .env.test down

test-logs:
	docker compose -f docker-compose.test.yaml --env-file .env.test logs -f api

migrate-test:
	@test -f .env.test || (echo "Copy .env.test.example to .env.test first" && exit 1)
	docker compose -f docker-compose.test.yaml --profile migrate --env-file .env.test run --rm migrate

prod-up:
	@test -f .env.prod || (echo "Copy .env.prod.example to .env.prod and set secrets first" && exit 1)
	docker compose -f docker-compose.prod.yaml --env-file .env.prod up -d --build

prod-down:
	docker compose -f docker-compose.prod.yaml --env-file .env.prod down

prod-logs:
	docker compose -f docker-compose.prod.yaml --env-file .env.prod logs -f api

migrate-prod:
	@test -f .env.prod || (echo "Copy .env.prod.example to .env.prod first" && exit 1)
	docker compose -f docker-compose.prod.yaml --profile migrate --env-file .env.prod run --rm migrate

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

# --- Kubernetes (see docs/KUBERNETES.md) ---

K8S_NAMESPACE ?= brenox
K8S_OVERLAY ?= dev
KIND_CLUSTER ?= brenox

k8s-check:
	@kubectl config current-context >/dev/null 2>&1 && exit 0; \
	ctx="kind-$(KIND_CLUSTER)"; \
	if kubectl config get-contexts -o name 2>/dev/null | grep -Fxq "$$ctx"; then \
		echo "Activating kubectl context $$ctx"; \
		kubectl config use-context "$$ctx"; \
		exit 0; \
	fi; \
	echo "No Kubernetes cluster configured."; \
	echo "  kind:  make k8s-cluster-create   # then make k8s-load-kind k8s-dev-up"; \
	echo "  Or:    kubectl config use-context kind-$(KIND_CLUSTER)"; \
	echo "  Docker Desktop: enable Kubernetes in Settings, then retry"; \
	exit 1

k8s-cluster-create:
	kind create cluster --name $(KIND_CLUSTER)
	kubectl config use-context kind-$(KIND_CLUSTER)

k8s-build:
	docker build -t brenox-engine:dev .
	docker build -f deploy/Dockerfile.migrate -t brenox-migrate:dev .

k8s-load-kind:
	kind load docker-image brenox-engine:dev --name $(KIND_CLUSTER)
	kind load docker-image brenox-migrate:dev --name $(KIND_CLUSTER)

k8s-dev-up: k8s-check k8s-build
	kubectl apply -k deploy/overlays/$(K8S_OVERLAY)
	$(MAKE) k8s-migrate K8S_OVERLAY=$(K8S_OVERLAY)

k8s-migrate:
ifeq ($(K8S_OVERLAY),dev)
	kubectl wait --for=condition=available deployment/postgres -n $(K8S_NAMESPACE) --timeout=120s
endif
	kubectl delete job brenox-migrate -n $(K8S_NAMESPACE) --ignore-not-found
	kubectl apply -k deploy/overlays/$(K8S_OVERLAY)
	kubectl wait --for=condition=complete job/brenox-migrate -n $(K8S_NAMESPACE) --timeout=180s

k8s-dev-down:
	kubectl delete -k deploy/overlays/$(K8S_OVERLAY) --ignore-not-found

k8s-port-forward:
	kubectl port-forward -n $(K8S_NAMESPACE) svc/brenox-engine 8080:8080

k8s-ingress-kind:
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=180s
	@echo "Add to /etc/hosts: 127.0.0.1 brenox.local"
