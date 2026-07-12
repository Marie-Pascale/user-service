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
| `Session` | Une connexion active d'un usager — jeton opaque stocké en base, révocable immédiatement (voir `UserAuthMiddleware`) |

Les permissions effectives d'un usager = ses permissions directes + celles de ses rôles directs + celles des rôles hérités de ses groupes. Calculé par `AuthorizationService.ResolvePermissionKeys` dans la couche `application`.

## Cas d'usage (`internal/application`)

| Cas d'usage | Rôle |
|---|---|
| `CreateUserUseCase` | Crée un usager, rejette les courriels déjà utilisés |
| `RegisterAuthMethodUseCase` | Ajoute une méthode d'authentification (locale ou OAuth) à un usager existant |
| `AuthenticateUserUseCase` | Connexion par mot de passe local ; retourne toujours la même erreur en cas d'échec (anti-énumération) |
| `AuthorizationService` | Résout les permissions effectives d'un usager pour un service : directes + rôles directs + rôles hérités des groupes |

Chacun est testé avec des repositories **en mémoire** (`internal/application/fakes_test.go`) plutôt que Postgres — rapide, déterministe, aucun besoin que `make up` tourne pour ces tests-là (`make test` suffit).

## Documentation API

La spec complète est dans [`openapi.yaml`](./openapi.yaml) (format OpenAPI 3.0). Deux façons de la consulter :

- **Intégrée** : une fois `make up` lancé, va sur [http://localhost:8080/docs](http://localhost:8080/docs) — Swagger UI, chargé depuis un CDN, pointé vers `/openapi.yaml`
- **Externe** : colle le contenu de `openapi.yaml` dans [editor.swagger.io](https://editor.swagger.io), si tu préfères ne pas avoir l'API en marche

Cette spec documente l'état **réel actuel** des routes (y compris lesquelles sont protégées et par quel mécanisme) — à tenir à jour à chaque changement de route, plutôt que de la laisser dériver.

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
| `POST` | `/auth/login` | Connexion (`{"email", "password"}`) — retourne l'usager **et un jeton de session** | Aucune |
| `POST` | `/auth/logout` | Révoque la session courante | **Session** (`Authorization: Bearer`) |
| `GET` | `/me` | Le profil de l'usager connecté | **Session** (`Authorization: Bearer`) |
| `POST` | `/services` | Crée un service tiers, retourne sa clé API **une seule fois** (`{"name", "description"}`) | Aucune ⚠️ (voir note ci-dessous) |
| `GET` | `/users/{id}/permissions` | Permissions effectives d'un usager, **pour le service authentifié** | **Clé API** (`X-API-Key`) |

**⚠️ Sécurité connue et volontairement incomplète** : `POST /services` n'est pas encore protégée — n'importe qui peut créer un service et obtenir une clé API. Cette route devra être restreinte aux administrateurs une fois qu'un rôle "admin" sera vérifié dans `UserAuthMiddleware` (ou un middleware d'autorisation séparé, branché sur `AuthorizationService`).

**Deux mécanismes d'authentification distincts, deux en-têtes distincts** — volontaire, pour ne jamais les confondre :
- Services tiers → `X-API-Key: <clé>` (clé permanente jusqu'à rotation)
- Usagers → `Authorization: Bearer <jeton>` (jeton de session, expire après 24h ou après `/auth/logout`)

**Exemple d'utilisation de bout en bout** :
```bash
# 1. Creer un usager et lui definir un mot de passe
curl -X POST localhost:8080/users -d '{"name":"Alice","email":"alice@example.com"}'
curl -X POST localhost:8080/users/<user-id>/password -d '{"password":"s3cret"}'

# 2. Se connecter -- recupere un jeton de session
curl -X POST localhost:8080/auth/login -d '{"email":"alice@example.com","password":"s3cret"}'
# -> {"user": {...}, "token": "a1b2c3...", "expires_at": "..."}

# 3. Utiliser ce jeton pour consulter son propre profil
curl localhost:8080/me -H "Authorization: Bearer a1b2c3..."

# 4. Se deconnecter -- revoque la session immediatement (contrairement a un JWT)
curl -X POST localhost:8080/auth/logout -H "Authorization: Bearer a1b2c3..."

# --- Cote services tiers (mecanisme separe, cle API) ---

# 5. Creer un service, recuperer sa cle API (affichee UNE SEULE FOIS)
curl -X POST localhost:8080/services -d '{"name":"orders-service","description":""}'
# -> {"id":"...", "api_key":"a1b2c3..."}

# 6. Utiliser cette cle pour consulter les permissions d'un usager
curl localhost:8080/users/<user-id>/permissions -H "X-API-Key: a1b2c3..."
# -> {"user_id":"...", "service_id":"...", "permissions":["orders:read", ...]}
```

## Prochaines étapes

- Middleware/logique d'autorisation par permission (au-dela de la simple authentification) — brancher `AuthorizationService.HasPermission` sur des routes specifiques, ex: proteger `POST /services` derriere une permission "admin"
- Purge periodique des sessions expirees (actuellement, une session expiree n'est supprimee qu'au moment ou elle est *utilisee* — voir `UserAuthMiddleware`)
- Handlers HTTP pour `Permission`, `Role`, `Group` (même patron que `UserHandler`)
- Migration des routes existantes (`GET/PUT/DELETE /users/{id}`) vers une protection par session, avec la règle "un usager ne voit que lui-même, sauf un admin" évoquée en concevant le domaine
