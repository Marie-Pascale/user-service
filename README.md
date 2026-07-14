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

Voir [`loadtest/README.md`](./loadtest/README.md). Tourne dans un conteneur Docker dédié (version de `vegeta` épinglée) — comme pour les tests unitaires, aucune installation locale requise.
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
| `CreateUserUseCase` | Crée un usager (sans authentification), rejette les courriels déjà utilisés |
| `RegisterAuthMethodUseCase` | Ajoute une méthode d'authentification (locale ou OAuth) à un usager **existant** |
| `RegisterUserUseCase` | Compose les deux ci-dessus : crée un usager **et** sa première méthode d'authentification, en une opération — avec suppression de compensation si la deuxième étape échoue |
| `AuthenticateUserUseCase` | Connexion par mot de passe local ; retourne toujours la même erreur en cas d'échec (anti-énumération) |
| `CreateSessionUseCase` / `LogoutUseCase` | Création et révocation de session, séparées de l'authentification elle-même |
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
| `GET` | `/docs` | Documentation Swagger UI | Aucune |
| `GET` | `/openapi.yaml` | Spec OpenAPI brute | Aucune |
| `POST` | `/users` | Crée un usager **avec** sa méthode d'authentification (`{"name", "email", "auth_method": {"password"}}` — `provider` optionnel, défaut `local`) | Aucune (auto-inscription) |
| `GET` | `/users` | Liste tous les usagers | Session + `users:read:any` |
| `GET` | `/users/{id}` | Récupère un usager | Session — soi-même OU `users:read:any` |
| `PUT` | `/users/{id}` | Met à jour nom/courriel | Session — soi-même OU `users:write:any` |
| `DELETE` | `/users/{id}` | Supprime un usager | Session — soi-même OU `users:delete:any` |
| `POST` | `/users/{id}/password` | Ajoute/remplace le mot de passe local d'un usager **existant** (`{"password"}`) | Session — soi-même OU `users:write:any` |
| `POST` | `/auth/login` | Connexion (`{"email", "password"}`) — retourne l'usager **et un jeton de session** | Aucune |
| `POST` | `/auth/logout` | Révoque la session courante | **Session** (`Authorization: Bearer`) |
| `GET` | `/me` | Le profil de l'usager connecté | **Session** (`Authorization: Bearer`) |
| `POST` | `/services` | Crée un service tiers, retourne sa clé API **une seule fois** (`{"name", "description"}`) | Session + `services:create` |
| `GET` | `/services` | Liste tous les services (jamais la clé, même hachée) | Session + `services:manage` |
| `GET` / `DELETE` | `/services/{id}` | Récupère / supprime un service | Session + `services:manage` |
| `POST` | `/services/{id}/rotate-key` | Fait pivoter la clé API, invalidant l'ancienne **immédiatement** — nouvelle clé affichée **une seule fois** | Session + `services:manage` |
| `GET` | `/users/{id}/permissions` | Permissions effectives d'un usager, **pour le service authentifié** | **Clé API** (`X-API-Key`) |
| `POST` | `/permissions` | Crée une permission (`{"service_id", "key", "description"}`) | Session + `permissions:manage` |
| `GET` | `/permissions?service_id=...` | Liste les permissions d'un service | Session + `permissions:manage` |
| `GET` / `DELETE` | `/permissions/{id}` | Récupère / supprime une permission | Session + `permissions:manage` |
| `POST` | `/roles` | Crée un rôle (`{"name", "description"}`) | Session + `roles:manage` |
| `GET` | `/roles` | Liste tous les rôles | Session + `roles:manage` |
| `GET` / `DELETE` | `/roles/{id}` | Récupère (avec ses permissions) / supprime un rôle | Session + `roles:manage` |
| `POST` | `/roles/{id}/permissions` | Accorde une permission à un rôle (`{"permission_id"}`) | Session + `roles:manage` |
| `DELETE` | `/roles/{id}/permissions/{permissionId}` | Révoque une permission d'un rôle | Session + `roles:manage` |
| `POST` | `/groups` | Crée un groupe (`{"name", "description"}`) | Session + `groups:manage` |
| `GET` | `/groups` | Liste tous les groupes | Session + `groups:manage` |
| `GET` / `DELETE` | `/groups/{id}` | Récupère (membres + rôles) / supprime un groupe | Session + `groups:manage` |
| `POST` | `/groups/{id}/members` | Ajoute un membre (`{"user_id"}`) | Session + `groups:manage` |
| `DELETE` | `/groups/{id}/members/{userId}` | Retire un membre | Session + `groups:manage` |
| `POST` | `/groups/{id}/roles` | Assigne un rôle au groupe, hérité par tous ses membres (`{"role_id"}`) | Session + `groups:manage` |

**Autorisation "self-ou-permission"** : pour les routes marquées ainsi, un usager peut toujours agir sur son propre compte. Agir sur celui d'un autre exige la permission listée — accordée typiquement via le rôle `admin`, seedé automatiquement par la migration `000013` (voir section suivante). Les routes `/services`, `/permissions`, `/roles` et `/groups` n'ont pas cette exception : aucune notion de "soi-même" n'existe pour ces entités.

**Trois mécanismes distincts, jamais confondus** :
- Services tiers → `X-API-Key: <clé>` (clé permanente jusqu'à rotation)
- Usagers → `Authorization: Bearer <jeton>` (jeton de session, expire après 24h ou après `/auth/logout`)
- Permissions → résolues via `AuthorizationService`, le **même** mécanisme protège les routes de ce service ET répond aux services tiers qui consultent les permissions d'un usager

## Amorcer le premier administrateur

Le rôle `admin` (avec toutes les permissions internes) est créé automatiquement par la migration `000013_seed_self_service`. Mais **aucun usager n'a ce rôle par défaut** — il faut l'accorder manuellement au premier compte :

```bash
make grant-admin email=alice@example.com
```

Requête SQL directe contre la base — volontairement, il n'existe **aucune route HTTP** pour accorder le rôle admin : ça éviterait sinon le problème classique de l'œuf et la poule au tout premier déploiement (qui protégerait la route qui donne les droits d'admin ?).

**Exemple d'utilisation de bout en bout** :
```bash
# 1. Creer un usager avec son mot de passe, en une seule requete
# (provider omis -- "local" est la valeur par defaut)
curl -X POST localhost:8080/users -d '{"name":"Alice","email":"alice@example.com","auth_method":{"password":"s3cret"}}'

# 2. L'accorder comme admin (une seule fois, en local)
make grant-admin email=alice@example.com

# 3. Se connecter -- recupere un jeton de session
curl -X POST localhost:8080/auth/login -d '{"email":"alice@example.com","password":"s3cret"}'
# -> {"user": {...}, "token": "a1b2c3...", "expires_at": "..."}

# 4. En tant qu'admin, lister TOUS les usagers (impossible sans le role)
curl localhost:8080/users -H "Authorization: Bearer a1b2c3..."

# 5. Creer un service tiers (reserve aux admins)
curl -X POST localhost:8080/services -H "Authorization: Bearer a1b2c3..." -d '{"name":"orders-service","description":""}'
# -> {"id":"...", "api_key":"a1b2c3..."}

# 5b. Si la cle est compromise, la faire pivoter -- l'ancienne cesse de
# fonctionner immediatement
curl -X POST localhost:8080/services/<service-id>/rotate-key -H "Authorization: Bearer a1b2c3..."
# -> {"id":"...", "api_key":"nouvelle-cle..."}

# --- Cote services tiers (mecanisme separe, cle API) ---

# 6. Utiliser cette cle pour consulter les permissions d'un usager
curl localhost:8080/users/<user-id>/permissions -H "X-API-Key: a1b2c3..."
# -> {"user_id":"...", "service_id":"...", "permissions":["orders:read", ...]}
```

## CORS (pour un frontend)

L'API autorise les requêtes cross-origin via `CORSMiddleware`, configuré par la variable d'environnement `CORS_ALLOWED_ORIGINS` (liste séparée par des virgules, ou `*` pour tout autoriser). Par défaut, `docker-compose.yml` la met à `*` — pratique en développement local avec un frontend sur un autre port (ex: `http://localhost:5173` pour Vite), mais **à restreindre explicitement** avant toute exposition au-delà de ta machine :

```yaml
# docker-compose.yml
environment:
  CORS_ALLOWED_ORIGINS: "http://localhost:5173,https://mon-frontend.example"
```

Contrairement aux autres middlewares du projet (qui protègent chacun une route précise), celui-ci enveloppe le routeur **au complet** dans `main.go` — un navigateur peut envoyer une requête *preflight* (`OPTIONS`) vers n'importe quelle route, et les en-têtes CORS doivent être présents sur toutes les réponses, succès comme erreurs.

## Prochaines étapes

- Vrai flot de vérification OAuth (Google, Facebook) : échange de code d'autorisation + validation du jeton signé auprès du fournisseur, pour pouvoir enfin brancher `RegisterUserUseCase.ExecuteOAuth` sur un handler HTTP public en toute sécurité
- Purge périodique des sessions expirées (actuellement, une session expirée n'est supprimée qu'au moment où elle est *utilisée* — voir `UserAuthMiddleware`)
- Tests de charge pour les nouveaux endpoints critiques (voir `loadtest/README.md`)
