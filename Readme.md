# Gestionnaire d'usagers

API de gestion d'usagers écrite en Go, avec PostgreSQL comme base de données.

## Structure du projet

```
.
├── cmd/api/main.go      # Point d'entrée de l'application
├── docker-compose.yml   # Orchestration des services (api + db)
├── Dockerfile            # Image Docker pour l'API Go
├── .air.toml             # Configuration du hot-reload
└── go.mod                # Dépendances Go
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

## Prochaines étapes

- Connecter l'API à PostgreSQL (driver `pgx` recommandé)
- Créer le modèle `User` (id, nom, courriel, etc.)
- Ajouter les routes CRUD : `GET /users`, `POST /users`, `GET /users/{id}`, `PUT /users/{id}`, `DELETE /users/{id}`
- Ajouter les migrations de base de données