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

**Important** : ce projet dépend maintenant de `golang.org/x/crypto` (pour le hachage des mots de passe). Après avoir décompressé le projet, lance ceci une seule fois pour télécharger la dépendance et générer `go.sum` :

```bash
go mod tidy
```

Ensuite :

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

## Entités du domaine

| Entité | Rôle |
|---|---|
| `User` | L'usager (identité, sans authentification) |
| `AuthMethod` | Une méthode d'authentification liée à un usager (mot de passe local, Google, Facebook...) — plusieurs par usager |
| `Service` | Un service tiers de l'écosystème, identifié par son propre ID et une clé API |
| `Permission` | Une permission précise, définie par un service (ex: `orders:read`) |
| `Role` | Un regroupement nommé de permissions, pouvant couvrir plusieurs services |
| `Group` | Un regroupement d'usagers ; les rôles assignés au groupe sont hérités par tous ses membres |

Les permissions effectives d'un usager = ses permissions directes + celles de ses rôles directs + celles des rôles hérités de ses groupes. Ce calcul sera fait par un service de la couche `application` (pas encore implémenté).

## Prochaines étapes

- Repositories PostgreSQL pour chaque entité (`internal/infrastructure/repository`)
- Cas d'usage dans `internal/application` : `CreateUser`, `RegisterAuthMethod`, `AuthorizationService` (résolution des permissions)
- Handlers HTTP dans `cmd/api/main.go`
- Middleware d'authentification (JWT ou sessions)
