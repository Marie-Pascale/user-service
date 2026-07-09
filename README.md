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

**Recommandé** : ce projet dépend de `golang.org/x/crypto` (pour le hachage des mots de passe). Le build Docker télécharge cette dépendance automatiquement, donc ce n'est pas strictement requis pour `make up`. Mais pour que ton IDE (VS Code / gopls) résolve correctement les imports sans erreur pendant que tu édites, lance ceci une fois localement :

```bash
go mod tidy
```

Ensuite, pour démarrer les conteneurs :

```bash
make up
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

Toutes les commandes ci-dessous s'exécutent **à l'intérieur du conteneur Docker**, pas sur ta machine — tu n'as donc pas besoin d'installer Go localement pour les lancer (seulement pour que ton IDE t'assiste en éditant les fichiers). Si les conteneurs ne sont pas démarrés, `make` les démarre automatiquement en mode détaché avant d'exécuter la commande.

**Tests unitaires**
```bash
make test              # lance tous les tests (rapide, ne touche pas a la BD)
make test-verbose      # avec le detail de chaque cas
make coverage          # genere un rapport de couverture (coverage.html, visible directement sur ton disque)
make race               # avec detection des race conditions
```

**Tests d'intégration** (contre une vraie base Postgres)
```bash
make test-integration   # applique les migrations puis lance les tests marques "integration"
```
Ces tests utilisent un [build tag](https://pkg.go.dev/go/build#hdr-Build_Constraints) Go (`//go:build integration`) pour rester séparés des tests unitaires — `go test ./...` seul ne les exécute jamais, il faut explicitement `-tags=integration`.

**Benchmarks** (performance au niveau du code)
```bash
make bench
```

**Tests de charge** (performance de l'API sous requêtes concurrentes)

Voir [`loadtest/README.md`](./loadtest/README.md). Contrairement aux tests unitaires, `vegeta` doit être installé sur ta machine — il attaque l'API depuis l'extérieur, via le port publié (`localhost:8080`), pour simuler du vrai trafic réseau.
```bash
make load-test-health
```

**Autres commandes utiles**
```bash
make shell    # ouvre un shell dans le conteneur api, pour deboguer manuellement
make down     # arrete les conteneurs
```

## Migrations

Gérées avec [`golang-migrate`](https://github.com/golang-migrate/migrate), exécuté dans le conteneur `api` (installé via le `Dockerfile`). Les fichiers vivent dans `migrations/`, en paires `up`/`down` numérotées séquentiellement.

```bash
make migrate-up        # applique toutes les migrations en attente
make migrate-down      # annule la derniere migration
make migrate-reset     # annule TOUT le schema (destructif, dev seulement)
make migrate-version   # affiche la version actuelle
make migrate-create name=nom_de_la_migration   # cree une nouvelle paire up/down
```

Après `make up` la première fois, lance `make migrate-up` pour créer les tables.

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

- Repositories PostgreSQL pour les entités restantes (`AuthMethod`, `Service`, `Permission`, `Role`, `Group`) — même patron que `PostgresUserRepository`
- Cas d'usage dans `internal/application` : `CreateUser`, `RegisterAuthMethod`, `AuthorizationService` (résolution des permissions)
- Handlers HTTP dans `cmd/api/main.go`, avec injection du repository
- Middleware d'authentification (JWT ou sessions)
