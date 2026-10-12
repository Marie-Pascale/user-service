.PHONY: help test test-verbose test-integration coverage race bench load-test-health up rebuild down ensure-up shell mod-tidy grant-admin clean-db

.DEFAULT_GOAL := help

EXEC := docker compose exec api
DB_URL := postgres://usagers:usagers_pwd@db:5432/usagers_db?sslmode=disable

help: ## Display this list of commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## Container ------------------------------------------------------------

ensure-up: ## Start containers in detached mode if they are not already running
	@docker compose ps --status running --services 2>/dev/null | grep -q '^api$$' \
		|| docker compose up -d --build

up: ## Start the API and database (attached logs)
	docker compose up --build

rebuild: ## Force-rebuild the API image, even if the container is already running
	docker compose up -d --build

down: ## Stop the containers
	docker compose down

shell: ensure-up ## Open a shell in the API container
	$(EXEC) sh

mod-tidy: ensure-up ## Update project dependencies (go mod tidy); local Go installation not required
	$(EXEC) go mod tidy

grant-admin: ensure-up ## Grant the admin role to an existing user: make grant-admin email=alice@example.com
	@if [ -z "$(email)" ]; then \
		echo "Error: email is required -- use \"make grant-admin email=alice@example.com\""; \
		echo "(no positional arguments: make does not read arguments like a shell script)"; \
		exit 1; \
	fi
	@result=$$(docker compose exec db psql -U usagers -d usagers_db -t -c \
		"INSERT INTO user_roles (user_id, role_id) SELECT id, '00000000-0000-0000-0000-000000000002' FROM users WHERE email = '$(email)' ON CONFLICT DO NOTHING RETURNING user_id;"); \
	if [ -n "$$(echo "$$result" | tr -d '[:space:]')" ]; then \
		echo "Admin role granted to $(email)."; \
	else \
		echo "No user found with email '$(email)' (or they already had the admin role) -- check the address."; \
	fi

clean-db: ensure-up ## Completely reset the database: drop the schema and reapply migrations (preserves seeded data)
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" down -all
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" up

## Unit tests (run in the container) ------------------------------------

test: ensure-up ## Run all unit tests (fast, no database required)
	$(EXEC) go test ./...

test-integration: ensure-up migrate-up ## Run integration tests (requires the database; slower)
	$(EXEC) go test -tags=integration ./...

test-verbose: ensure-up ## Run all unit tests with details for each case
	$(EXEC) go test -v ./...

coverage: ensure-up ## Generate an HTML code coverage report
	$(EXEC) sh -c "go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"
	@echo "Report generated: coverage.html (available on your host via the mounted volume)"

race: ensure-up ## Run tests with the race detector
	docker compose exec -e CGO_ENABLED=1 api go test -race ./...

bench: ensure-up ## Run Go benchmarks (code-level performance)
	$(EXEC) go test -bench=. -benchmem ./...

## Load tests (run from the host via the published port) -----------------

load-test-health: ensure-up ## Load test GET /health in an ephemeral container (50 req/s, 30s)
	docker compose run --rm -T loadtest attack -targets=/loadtest/health.http -rate=50 -duration=30s -output=/loadtest/results.bin
	docker compose run --rm -T loadtest report /loadtest/results.bin

## Migrations (run in the container) -------------------------------------

migrate-up: ensure-up ## Apply all pending migrations
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ensure-up ## Roll back the most recently applied migration
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-reset: ensure-up ## Roll back ALL migrations (drops the schema)
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" down -all

migrate-version: ensure-up ## Show the current migration version
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" version

migrate-create: ensure-up ## Create a new migration: make migrate-create name=migration_name
	$(EXEC) migrate create -ext sql -dir ./migrations -seq $(name)
