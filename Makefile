.PHONY: test test-verbose coverage race bench load-test-health up down

## Tests unitaires ---------------------------------------------------------

test: ## Lance tous les tests unitaires
	go test ./...

test-verbose: ## Lance tous les tests unitaires avec le detail de chaque cas
	go test -v ./...

coverage: ## Genere un rapport de couverture de code (HTML)
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Rapport genere : coverage.html"

race: ## Lance les tests avec le detecteur de race conditions
	go test -race ./...

bench: ## Lance les benchmarks Go (performance au niveau du code)
	go test -bench=. -benchmem ./...

## Tests de charge (necessite l'API en cours d'execution) -----------------

load-test-health: ## Test de charge sur GET /health (50 req/s pendant 30s)
	vegeta attack -targets=loadtest/health.http -rate=50 -duration=30s | vegeta report

## Docker -------------------------------------------------------------------

up: ## Demarre l'API et la base de donnees
	docker compose up --build

down: ## Arrete les conteneurs
	docker compose down
