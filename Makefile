.PHONY: help dev api worker frontend migrate-up build clean test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Start all services (Redis + ClickHouse)
	docker-compose up -d
	@echo "✓ Services started"
	@echo "  Redis: localhost:6379"
	@echo "  ClickHouse: localhost:9000"

api: ## Run API server
	go run cmd/api/main.go

worker: ## Run worker
	go run cmd/worker/main.go

frontend: ## Start frontend dev server
	cd frontend && npm run dev

migrate-up: ## Run Postgres migrations
	@echo "Running migrations..."
	psql -U postgres -d bugvay -f migrations/002_add_indexes.sql
	@echo "✓ Migrations complete"

migrate-clickhouse: ## Run ClickHouse migrations
	@echo "Running ClickHouse migrations..."
	clickhouse-client --queries-file migrations/clickhouse/001_scan_results.sql
	@echo "✓ ClickHouse migrations complete"

build-api: ## Build API binary
	go build -o bin/api cmd/api/main.go

build-worker: ## Build worker binary
	go build -o bin/worker cmd/worker/main.go

build: build-api build-worker ## Build all binaries

test: ## Run tests
	go test -v ./...

clean: ## Stop services and clean up
	docker-compose down -v
	rm -rf bin/

deps: ## Download Go dependencies
	go mod tidy
	go mod download

lint: ## Run linters
	golangci-lint run

.DEFAULT_GOAL := help
