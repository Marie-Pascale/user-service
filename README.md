# User Management Service

A user management API written in Go, with PostgreSQL as its database.

## Design notes and disclaimers

This project was developed with assistance from Claude and GitHub Copilot, under my direction.

## Project structure

```
.
├── cmd/api/main.go       # Application entry point
├── internal/
│   ├── domain/            # Entities and pure business logic
│   ├── application/       # Use cases / application services
│   └── infrastructure/
│       ├── repository/    # Data access (Postgres, etc.)
│       └── service/       # Integrations with external services
├── docker-compose.yml    # Service orchestration (api + db)
├── Dockerfile             # Docker image for the Go API
├── .air.toml              # Hot-reload configuration
└── go.mod                 # Go dependencies
```

## Available commands

```bash
make
```
(or `make help`, which is equivalent) lists all `Makefile` targets with their descriptions. The list is generated automatically from the `## ...` comments following each target, so there is no need to maintain it manually: just add `## description` after a new target and it will appear in the menu.

**After modifying the `Dockerfile`** (adding a system dependency or installing a new tool, for example), explicitly run `make rebuild`. `ensure-up` (a dependency of most targets) only rebuilds the image if the `api` container is **not already running**. If it is already running with an outdated image, it keeps using that image and will not rebuild it automatically.

## Getting started

**Required before the very first `make up`** (only if `go.sum` does not exist in the project yet): generate it using one of the following methods.

**Without Go installed locally** — using Docker:
```bash
make mod-tidy
```
This works even if `go.sum` does not exist yet: `docker-compose.yml` mounts the project directory as a volume (`.:/app`), so anything `go mod tidy` writes *inside* the container (such as `go.sum`) appears directly on your machine, just like `coverage.html`.

**With Go installed locally**:
```bash
go mod tidy
```

Then start the containers:

```bash
make up
```

If you add a new dependency later (a new external `import`), run `make mod-tidy` again to update `go.mod`/`go.sum` accordingly.

The API will be available at `http://localhost:8080`.
PostgreSQL will be available at `localhost:5432` (for connecting with a client such as DBeaver or psql).

## Verify that it works

```bash
curl http://localhost:8080/health
```

You should receive: `{"status":"ok"}`

## Stop the containers

```bash
docker compose down
```

To also delete the database data:

```bash
docker compose down -v
```

## Tests

All commands below run **inside the Docker container**, not on your machine, so you do not need to install Go locally to run them (only if you want your IDE to provide assistance while editing files). If the containers are not running, `make` starts them in detached mode before executing the command.

**Unit tests**
```bash
make test              # run all tests (fast, does not touch the database)
make test-verbose      # show details for each test case
make coverage          # generate a coverage report (coverage.html, available on your machine)
make race               # run with the race detector
```

**Integration tests** (against a real Postgres database)
```bash
make test-integration   # apply migrations, then run tests tagged "integration"
```
These tests use a Go [build tag](https://pkg.go.dev/go/build#hdr-Build_Constraints) (`//go:build integration`) to keep them separate from unit tests. `go test ./...` does not run them; you must explicitly pass `-tags=integration`.

**Benchmarks** (code-level performance)
```bash
make bench
```

**Load tests** (API performance under concurrent requests)

See [`loadtest/README.md`](./loadtest/README.md). They run in a dedicated Docker container (with a pinned `vegeta` version), so no local installation is required, just like for unit tests.
```bash
make load-test-health
```

**Other useful commands**
```bash
make shell    # open a shell in the api container for manual debugging
make down     # stop the containers
```

## Database migrations

Managed with [`golang-migrate`](https://github.com/golang-migrate/migrate), which runs in the `api` container (installed via the `Dockerfile`). Migration files live in `migrations/` as sequentially numbered `up`/`down` pairs.

```bash
make migrate-up        # apply all pending migrations
make migrate-down      # roll back the most recent migration
make migrate-reset     # roll back ALL migrations (destructive; development only)
make migrate-version   # show the current version
make migrate-create name=migration_name   # create a new up/down pair
make clean-db           # completely reset the database (equivalent to migrate-reset + migrate-up)
```

After running `make up` for the first time, run `make migrate-up` to create the tables.

**`make clean-db`** starts from a completely empty database and reapplies all migrations in order, including those that seed data (`000013`-`000015`: self-service, the `admin` role, and internal permissions). This is useful after manual testing with `curl` to restore a clean state **without** having to run `make grant-admin` again (the admin role is recreated, but no users are assigned to it; you will need to create an account and assign it again).

Automated integration tests (`make test-integration`) clean up differently: `TestMain` empties the tables **before** the test suite starts (see `internal/infrastructure/repository/main_test.go`), ensuring a clean state even if the previous run failed before reaching its `t.Cleanup` calls. `make clean-db` is still useful for occasional manual cleanup outside the test cycle.

## Domain entities

| Entity | Purpose |
|---|---|
| `User` | A user (identity, without authentication) |
| `AuthMethod` | An authentication method linked to a user (local password, Google, Facebook, etc.); a user can have several |
| `Service` | A third-party service in the ecosystem, identified by its own ID and an API key |
| `Permission` | A specific permission defined by a service (e.g. `orders:read`) |
| `Role` | A named collection of permissions that can span multiple services |
| `Group` | A collection of users; roles assigned to a group are inherited by all its members |
| `Session` | An active user session — an opaque token stored in the database and immediately revocable (see `UserAuthMiddleware`) |

All entities have `created_at`/`updated_at` (`Permission`, `Role`, and `Group` since migration `000016`; `Service` since `000017`). `UpdatedAt` advances on **every** mutation, not only through `PUT`. For example, `Role.GrantPermission()`/`RevokePermission()`, `Group.AddMember()`/`RemoveMember()`/`AssignRole()`, and `Service.RotateAPIKey()` also update it, both in memory (domain) and in the database (Postgres repository).

`User` also has `LastLoginAt` (`null` until the user has logged in; migration `000018`), updated in **`CreateSessionUseCase`** — not in `AuthenticateUserUseCase` — so it consistently covers every authentication method that results in a session (local password today, `Google`/`Facebook` later), without duplicating this logic in each authentication flow. This is **best-effort**: a failure to update it never blocks the login itself (the same approach used by `UserAuthMiddleware` with `Sessions.Touch`).

A user's effective permissions are the sum of their direct permissions, permissions from their directly assigned roles, and permissions from roles inherited through their groups. They are calculated by `AuthorizationService.ResolvePermissionKeys` in the `application` layer.

## Use cases (`internal/application`)

| Use case | Purpose |
|---|---|
| `CreateUserUseCase` | Creates a user (without authentication) and rejects email addresses that are already in use |
| `RegisterAuthMethodUseCase` | Adds an authentication method (local or OAuth) to an **existing** user |
| `RegisterUserUseCase` | Combines the two above: creates a user **and** their first authentication method in one operation, with compensating deletion if the second step fails |
| `AuthenticateUserUseCase` | Authenticates with a local password; always returns the same error on failure (anti-enumeration) |
| `CreateSessionUseCase` / `LogoutUseCase` | Create and revoke sessions separately from authentication itself |
| `AuthorizationService` | Resolves a user's effective permissions for a service: direct permissions + directly assigned roles + roles inherited from groups |

Each is tested with **in-memory** repositories (`internal/application/fakes_test.go`) rather than Postgres — fast and deterministic, with no need for `make up` to be running (`make test` is enough).

## API documentation

The complete specification is in [`openapi.yaml`](./openapi.yaml) (OpenAPI 3.0 format). You can view it in two ways:

- **Built-in**: once `make up` is running, visit [http://localhost:8080/docs](http://localhost:8080/docs) — Swagger UI is loaded from a CDN and points to `/openapi.yaml`.
- **External**: paste the contents of `openapi.yaml` into [editor.swagger.io](https://editor.swagger.io) if you prefer not to run the API.

This specification documents the **actual current** state of the routes, including which are protected and by which mechanism. Keep it up to date whenever routes change rather than letting it drift.

## HTTP endpoints

| Method | Route | Description | Authentication |
|---|---|---|---|
| `GET` | `/health` | Check that the server is responding | None |
| `GET` | `/docs` | Swagger UI documentation | None |
| `GET` | `/openapi.yaml` | Raw OpenAPI specification | None |
| `POST` | `/users` | Create a user **with** an authentication method (`{"name", "email", "auth_method": {"password"}}` — `provider` is optional and defaults to `local`) | None (self-registration) |
| `GET` | `/users` | List all users | Session + `users:read:any` |
| `GET` | `/users/{id}` | Get a user | Session — self OR `users:read:any` |
| `PUT` | `/users/{id}` | Update name/email | Session — self OR `users:write:any` |
| `DELETE` | `/users/{id}` | Delete a user | Session — self OR `users:delete:any` |
| `POST` | `/users/{id}/password` | Add/replace the local password of an **existing** user (`{"password"}`) | Session — self OR `users:write:any` |
| `POST` | `/auth/login` | Log in (`{"email", "password"}`) — returns the user **and a session token** | None |
| `POST` | `/auth/logout` | Revoke the current session | **Session** (`Authorization: Bearer`) |
| `GET` | `/me` | Profile of the logged-in user | **Session** (`Authorization: Bearer`) |
| `POST` | `/services` | Create a third-party service; its API key is returned **only once** (`{"name", "description"}`) | Session + `services:create` |
| `GET` | `/services` | List all services (never returns the key, even hashed) | Session + `services:manage` |
| `GET` / `PUT` / `DELETE` | `/services/{id}` | Get / update (name, description) / delete a service | Session + `services:manage` |
| `POST` | `/services/{id}/rotate-key` | Rotate the API key, invalidating the old one **immediately** — the new key is displayed **only once** | Session + `services:manage` |
| `GET` | `/users/{id}/profile` | Basic user info (`id`, `name`, `email`, `status`), **for the authenticated service** | **API key** (`X-API-Key`) |
| `GET` | `/users/{id}/permissions` | Effective permissions for a user, **for the authenticated service** | **API key** (`X-API-Key`) |
| `GET` | `/users/{id}/permissions/check?permission=...` | Check **one** specific permission (`{"allowed": true/false}`) | **API key** (`X-API-Key`) |
| `POST` | `/permissions/self` | A service registers a permission for itself (`{"key", "description"}`); idempotent | **API key** (`X-API-Key`) |
| `POST` | `/permissions` | Create a permission (`{"service_id", "key", "description"}`) | Session + `permissions:manage` |
| `GET` | `/permissions?service_id=...` | List a service's permissions | Session + `permissions:manage` |
| `GET` / `DELETE` | `/permissions/{id}` | Get / delete a permission | Session + `permissions:manage` |
| `POST` | `/roles` | Create a role (`{"name", "description"}`) | Session + `roles:manage` |
| `GET` | `/roles` | List all roles | Session + `roles:manage` |
| `GET` / `PUT` / `DELETE` | `/roles/{id}` | Get (with permissions) / update (name, description) / delete a role | Session + `roles:manage` |
| `POST` | `/roles/{id}/permissions` | Grant a permission to a role (`{"permission_id"}`) | Session + `roles:manage` |
| `DELETE` | `/roles/{id}/permissions/{permissionId}` | Revoke a permission from a role | Session + `roles:manage` |
| `POST` | `/groups` | Create a group (`{"name", "description"}`) | Session + `groups:manage` |
| `GET` | `/groups` | List all groups | Session + `groups:manage` |
| `GET` / `PUT` / `DELETE` | `/groups/{id}` | Get (members + roles) / update (name, description) / delete a group | Session + `groups:manage` |
| `POST` | `/groups/{id}/members` | Add a member (`{"user_id"}`) | Session + `groups:manage` |
| `DELETE` | `/groups/{id}/members/{userId}` | Remove a member | Session + `groups:manage` |
| `POST` | `/groups/{id}/roles` | Assign a role to the group; all its members inherit it (`{"role_id"}`) | Session + `groups:manage` |

**"Self-or-permission" authorization**: for routes marked this way, a user can always act on their own account. Acting on another user's account requires the listed permission — typically granted through the `admin` role, which is seeded automatically by migration `000013` (see the next section). The `/services`, `/permissions`, `/roles`, and `/groups` routes do not have this exception: these entities have no concept of "self".

**Three distinct mechanisms that are never interchangeable**:
- Third-party services → `X-API-Key: <key>` (permanent key until rotated)
- Users → `Authorization: Bearer <token>` (session token, expires after 24 hours or after `/auth/logout`)
- Permissions → resolved through `AuthorizationService`; the **same** mechanism protects this service's routes AND answers third-party services querying a user's permissions

## Bootstrap the first administrator

The `admin` role (with all internal permissions) is created automatically by the `000013_seed_self_service` migration. However, **no user has this role by default** — you must grant it manually to the first account:

```bash
make grant-admin email=alice@example.com
```

This runs a direct SQL query against the database. There is deliberately **no HTTP route** for granting the admin role; otherwise, the first deployment would face the classic chicken-and-egg problem: who would protect the route that grants admin privileges?

**Safeguard**: `admin` must always have *all* permissions defined by the application (`application.AllPermissionKeys` in `authorization_policy.go`). An integration test (`TestAdminRole_HasAllApplicationPermissions`) checks this invariant against the real database on every `make test-integration` run. If a new permission is added in code without a corresponding migration granting it to `admin`, the test fails rather than allowing a silent regression (an admin losing access to a feature without anyone noticing until production).

**End-to-end usage example**:
```bash
# 1. Create a user with a password in a single request
# (provider omitted -- "local" is the default)
curl -X POST localhost:8080/users -d '{"name":"Alice","email":"alice@example.com","auth_method":{"password":"s3cret"}}'

# 2. Grant the admin role to the user (once, locally)
make grant-admin email=alice@example.com

# 3. Log in -- retrieve a session token
curl -X POST localhost:8080/auth/login -d '{"email":"alice@example.com","password":"s3cret"}'
# -> {"user": {...}, "token": "a1b2c3...", "expires_at": "..."}

# 4. As an admin, list ALL users (not possible without the role)
curl localhost:8080/users -H "Authorization: Bearer a1b2c3..."

# 5. Create a third-party service (admins only)
curl -X POST localhost:8080/services -H "Authorization: Bearer a1b2c3..." -d '{"name":"orders-service","description":""}'
# -> {"id":"...", "api_key":"a1b2c3..."}

# 5b. If the key is compromised, rotate it -- the old key stops working
# immediately
curl -X POST localhost:8080/services/<service-id>/rotate-key -H "Authorization: Bearer a1b2c3..."
# -> {"id":"...", "api_key":"new-key..."}

# --- Third-party services (separate API key mechanism) ---

# 6. The service registers its own permissions (typically at startup)
curl -X POST localhost:8080/permissions/self -H "X-API-Key: a1b2c3..." -d '{"key":"orders:read","description":"read orders"}'
# -> {"id":"...", "service_id":"...", "key":"orders:read", "description":"..."}
# Repeating the same command does not create a duplicate -- it returns the existing permission (200).

# 7. Get a user's basic information (never role_ids/permission_ids)
curl localhost:8080/users/<user-id>/profile -H "X-API-Key: a1b2c3..."
# -> {"id":"...", "name":"Alice", "email":"alice@example.com", "status":"active"}

# 8. Use this key to get a user's permissions
curl localhost:8080/users/<user-id>/permissions -H "X-API-Key: a1b2c3..."
# -> {"user_id":"...", "service_id":"...", "permissions":["orders:read", ...]}

# 9. Or check ONE specific permission directly (useful for a quick check
# before performing an action)
curl "localhost:8080/users/<user-id>/permissions/check?permission=orders:read" -H "X-API-Key: a1b2c3..."
# -> {"user_id":"...", "service_id":"...", "permission":"orders:read", "allowed":true}
```

## CORS (for a frontend)

The API allows cross-origin requests through `CORSMiddleware`, configured by the `CORS_ALLOWED_ORIGINS` environment variable (a comma-separated list, or `*` to allow all origins). By default, `docker-compose.yml` sets it to `*` — convenient for local development with a frontend on another port (e.g. `http://localhost:5173` for Vite), but **restrict it explicitly** before exposing the API beyond your machine:

```yaml
# docker-compose.yml
environment:
  CORS_ALLOWED_ORIGINS: "http://localhost:5173,https://my-frontend.example"
```

Unlike the other middleware in the project (each of which protects a specific route), this one wraps the **entire** router in `main.go`. A browser can send a *preflight* (`OPTIONS`) request to any route, so CORS headers must be present on every response, including both successes and errors.

## Session hijacking detection

Each `Session` records the client's IP and `User-Agent` **at login time** (`POST /auth/login`), as of migration `000019`. On every subsequent authenticated request, `UserAuthMiddleware` compares the current request's IP/`User-Agent` with the recorded values — see `domain.Session.IsSuspicious`.

**The rule**: a session is revoked only if **both** values differ at the same time.

- A user who changes networks (Wi-Fi → cellular) keeps the same `User-Agent` → allowed
- A browser updated during a session keeps the same IP → allowed
- A stolen token replayed from another device (different IP **and** `User-Agent`) → the session is immediately deleted and the request receives `401`

**Known limitation, documented in the code** (`domain/session.go`): `User-Agent` is an ordinary HTTP header and is trivial to forge. An attacker who knows the victim's exact `User-Agent` (often visible in the same logs that may have exposed the token) can copy it and bypass this check even from a different IP. This rule catches "naive" hijacking (a token simply replayed elsewhere, without forged headers), but not a targeted, sophisticated attacker. This is a deliberate trade-off between security and false positives, not an absolute guarantee.

`clientIP()` (in `internal/infrastructure/http/request_info.go`) reads `X-Forwarded-For` first, falling back to `RemoteAddr`. **This header is not currently protected by a trusted reverse proxy** in this project (which is deployed directly). Keep this in mind if the API is ever placed behind a proxy that does not filter it: a malicious client could then spoof its apparent IP address.

## Next steps

- A complete OAuth verification flow (Google, Facebook): authorization code exchange + signed token validation with the provider, so `RegisterUserUseCase.ExecuteOAuth` can eventually be safely connected to a public HTTP handler
- Periodic cleanup of expired sessions (currently, an expired session is only deleted when it is *used* — see `UserAuthMiddleware`)
- Load tests for new critical endpoints (see `loadtest/README.md`)

## License

This project is licensed under the [GNU Affero General Public License v3.0 or later](LICENSE.md) (AGPL-3.0-or-later). A French summary, for information only, is available in [LICENSE.fr.md](LICENSE.fr.md); only the English text is legally binding.

In short: you may use, modify and redistribute this software, provided you keep the same license and make the source code available. If you run a modified version as a network-accessible service, you must offer the source code of that version to its users.

Copyright © 2026 Marie-Pascale Girard
