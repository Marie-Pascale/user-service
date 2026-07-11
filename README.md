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

## Commandes disponibles

```bash
make
```
(ou `make help`, équivalent) Liste toutes les cibles du `Makefile` avec leur description, générée automatiquement à partir des commentaires `## ...` qui suivent chaque cible — pas besoin de maintenir cette liste à la main : ajoute simplement `## description` après une nouvelle cible et elle apparaîtra dans le menu.

## Démarrage

**Obligatoire avant le tout premier `make up`** (seulement si `go.sum` n'existe pas encore dans le projet) : génère-le avec l'une de ces deux méthodes.

**Sans Go installé localement** — via Docker :
```bash
make mod-tidy
```
Ça fonctionne même sans `go.sum` au départ : `docker-compose.yml` monte le dossier du projet en volume (`.:/app`), donc tout ce que `go mod tidy` écrit *à l'intérieur* du conteneur (le `go.sum`) apparaît directement sur ton disque, comme pour `coverage.html` plus tôt.

**Avec Go installé localement** :
```bash
go mod tidy
```

Ensuite, pour démarrer les conteneurs :

```bash
make up
```

Si tu ajoutes une nouvelle dépendance plus tard (un nouvel `import` externe), relance `make mod-tidy` pour mettre à jour `go.mod`/`go.sum` en conséquence.

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

## Cas d'usage (`internal/application`)

| Cas d'usage | Rôle |
|---|---|
| `CreateUserUseCase` | Crée un usager, rejette les courriels déjà utilisés |
| `RegisterAuthMethodUseCase` | Ajoute une méthode d'authentification (locale ou OAuth) à un usager existant |
| `AuthenticateUserUseCase` | Connexion par mot de passe local ; retourne toujours la même erreur en cas d'échec (anti-énumération) |
| `AuthorizationService` | Résout les permissions effectives d'un usager pour un service : directes + rôles directs + rôles hérités des groupes |

Chacun est testé avec des repositories **en mémoire** (`internal/application/fakes_test.go`) plutôt que Postgres — rapide, déterministe, aucun besoin que `make up` tourne pour ces tests-là (`make test` suffit).

## Endpoints HTTP

| Méthode | Route | Description | Protection |
|---|---|---|---|
| `GET` | `/health` | Vérifie que le serveur répond | Aucune |
| `POST` | `/users` | Crée un usager (`{"name", "email"}`) | Aucune |
| `GET` | `/users` | Liste tous les usagers | Aucune |
| `GET` | `/users/{id}` | Récupère un usager | Aucune |
| `PUT` | `/users/{id}` | Met à jour nom/courriel | Aucune |
| `DELETE` | `/users/{id}` | Supprime un usager | Aucune |
| `POST` | `/users/{id}/password` | Définit le mot de passe local d'un usager (`{"password"}`) | Aucune |
| `POST` | `/auth/login` | Connexion (`{"email", "password"}`) — retourne l'usager, **pas encore de token** | Aucune |
| `POST` | `/services` | Crée un service tiers, retourne sa clé API **une seule fois** (`{"name", "description"}`) | Aucune ⚠️ (voir note ci-dessous) |
| `GET` | `/users/{id}/permissions` | Permissions effectives d'un usager, **pour le service authentifié** | **Clé API** (`X-API-Key`) |

**⚠️ Sécurité connue et volontairement incomplète** : `POST /services` n'est pas encore protégée — n'importe qui peut créer un service et obtenir une clé API. Cette route devra être restreinte aux administrateurs une fois l'authentification usager en place. `GET /users/{id}/permissions` est la seule route protégée pour l'instant, via le nouveau middleware par clé API.

**Exemple d'utilisation du middleware** :
```bash
# 1. Creer un service, recuperer sa cle API (affichee UNE SEULE FOIS)
curl -X POST localhost:8080/services -d '{"name":"orders-service","description":""}'
# -> {"id":"...", "api_key":"a1b2c3..."}

# 2. Utiliser cette cle pour consulter les permissions d'un usager
curl localhost:8080/users/<user-id>/permissions -H "X-API-Key: a1b2c3..."
# -> {"user_id":"...", "service_id":"...", "permissions":["orders:read", ...]}
```

## Prochaines étapes

- Middleware de session/JWT pour les usagers — tu as choisi une **session opaque stockée en base** (révocable immédiatement, contrairement à un JWT signé). Nécessite : une entité `Session` dans le domaine, un `SessionRepository`, et `POST /auth/login` qui retourne enfin un token
- Protéger `POST /services` une fois l'authentification usager en place (probablement via un rôle "admin")
- Handlers HTTP pour `Permission`, `Role`, `Group` (même patron que `UserHandler`)
- Endpoint exposant `AuthorizationService.ResolvePermissionKeys` pour les services tiers
