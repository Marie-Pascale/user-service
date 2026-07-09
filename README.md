# Gestionnaire d'usagers

API de gestion d'usagers écrite en Go, avec PostgreSQL comme base de données.

## Structure du projet

```
.
├── cmd/api/main.go       # Point d'entrée de l'application
├── internal/
│   ├── domain/            # Entités et logique métier pure
│   ├── application/       # Cas d'utilisation / services applicatifs
│   └── infrastructure/
│       ├── repository/    # Accès aux données (Postgres, etc.)
│       └── service/       # Intégrations avec des services externes
├── docker-compose.yml    # Orchestration des services (api + db)
├── Dockerfile             # Image Docker pour l'API Go
├── .air.toml              # Configuration du hot-reload
└── go.mod                 # Dépendances Go
```

## Démarrage

```bash
docker compose up --build
```

L'API sera accessible sur `http://localhost:8080`.
La base PostgreSQL sera accessible sur `localhost:5432` (utile pour t'y connecter avec un client comme DBeaver ou psql).

## Vérifier que ça fonctionne

```bash
curl http://localhost:8080/health
```

Tu devrais recevoir : `{"status":"ok"}`

## Arrêter les conteneurs

```bash
docker compose down
```

Pour aussi supprimer les données de la base de données :

```bash
docker compose down -v
```

## Tests

**Tests unitaires**
```bash
make test              # lance tous les tests
make test-verbose      # avec le detail de chaque cas
make coverage          # genere un rapport de couverture (coverage.html)
make race              # avec detection des race conditions
```

**Benchmarks** (performance au niveau du code)
```bash
make bench
```

**Tests de charge** (performance de l'API sous requêtes concurrentes)

Voir [`loadtest/README.md`](./loadtest/README.md). Nécessite que l'API tourne (`make up`).
```bash
make load-test-health
```

## Prochaines étapes

- Connecter l'API à PostgreSQL (driver `pgx` recommandé)
- Créer le modèle `User` (id, nom, courriel, etc.)
- Ajouter les routes CRUD : `GET /users`, `POST /users`, `GET /users/{id}`, `PUT /users/{id}`, `DELETE /users/{id}`
- Ajouter les migrations de base de données
