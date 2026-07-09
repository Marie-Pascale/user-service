.PHONY: test test-verbose coverage race bench load-test-health up down ensure-up shell

EXEC := docker compose exec api

## Conteneur ------------------------------------------------------------

ensure-up: ## Demarre les conteneurs en mode detache s'ils ne tournent pas deja
	@docker compose ps --status running --services 2>/dev/null | grep -q '^api$$' \
		|| docker compose up -d --build

up: ## Demarre l'API et la base de donnees (logs attaches)
	docker compose up --build

down: ## Arrete les conteneurs
	docker compose down

shell: ensure-up ## Ouvre un shell dans le conteneur api
	$(EXEC) sh

## Tests unitaires (executes dans le conteneur) --------------------------

test: ensure-up ## Lance tous les tests unitaires
	$(EXEC) go test ./...

test-verbose: ensure-up ## Lance tous les tests unitaires avec le detail de chaque cas
	$(EXEC) go test -v ./...

coverage: ensure-up ## Genere un rapport de couverture de code (HTML)
	$(EXEC) sh -c "go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"
	@echo "Rapport genere : coverage.html (visible directement sur ton disque grace au volume monte)"

race: ensure-up ## Lance les tests avec le detecteur de race conditions
	$(EXEC) go test -race ./...

bench: ensure-up ## Lance les benchmarks Go (performance au niveau du code)
	$(EXEC) go test -bench=. -benchmem ./...

## Tests de charge (attaque depuis l'hote, via le port publie) -----------

load-test-health: ensure-up ## Test de charge sur GET /health (50 req/s pendant 30s)
	vegeta attack -targets=loadtest/health.http -rate=50 -duration=30s | vegeta report
