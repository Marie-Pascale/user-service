.PHONY: help test test-verbose test-integration coverage race bench load-test-health up down ensure-up shell mod-tidy

.DEFAULT_GOAL := help

EXEC := docker compose exec api
DB_URL := postgres://usagers:usagers_pwd@db:5432/usagers_db?sslmode=disable

help: ## Affiche cette liste de commandes
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

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

mod-tidy: ensure-up ## Mets a jour les dependances du projet (go mod tidy), sans besoin de Go local
	$(EXEC) go mod tidy

## Tests unitaires (executes dans le conteneur) --------------------------

test: ensure-up ## Lance tous les tests unitaires (rapide, sans base de donnees)
	$(EXEC) go test ./...

test-integration: ensure-up migrate-up ## Lance les tests d'integration (necessite la BD, plus lent)
	$(EXEC) go test -tags=integration ./...

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

## Migrations (executees dans le conteneur) -------------------------------

migrate-up: ensure-up ## Applique toutes les migrations en attente
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ensure-up ## Annule la derniere migration appliquee
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-reset: ensure-up ## Annule TOUTES les migrations (efface le schema)
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" down -all

migrate-version: ensure-up ## Affiche la version de migration actuelle
	$(EXEC) migrate -path ./migrations -database "$(DB_URL)" version

migrate-create: ensure-up ## Cree une nouvelle migration : make migrate-create name=nom_de_la_migration
	$(EXEC) migrate create -ext sql -dir ./migrations -seq $(name)
