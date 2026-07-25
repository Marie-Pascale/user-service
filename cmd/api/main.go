package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gestionnaire-usagers/internal/application"
	httpapi "gestionnaire-usagers/internal/infrastructure/http"
	"gestionnaire-usagers/internal/infrastructure/repository"
	"gestionnaire-usagers/internal/infrastructure/security"
)

// sessionTTL controle la duree de vie d'une session usager. Fixe en dur
// pour l'instant -- pourrait devenir une variable d'environnement si le
// besoin de le configurer par environnement (dev/prod) se presente.
const sessionTTL = 24 * time.Hour

// main est la racine de composition (composition root) : le SEUL endroit
// du projet ou les types concrets (PostgresUserRepository, BcryptHasher...)
// sont references directement. Tout le reste du code (application,
// httpapi) ne connait que des interfaces -- voir la discussion sur
// l'inversion de dependance plus tot dans le projet.
func main() {
	ctx := context.Background()

	pool, err := repository.NewPool(ctx, repository.ConfigFromEnv())
	if err != nil {
		log.Fatalf("connexion a la base de donnees: %v", err)
	}
	defer pool.Close()

	// --- Repositories --------------------------------------------------
	users := repository.NewPostgresUserRepository(pool)
	authMethods := repository.NewPostgresAuthMethodRepository(pool)
	services := repository.NewPostgresServiceRepository(pool)
	roles := repository.NewPostgresRoleRepository(pool)
	groups := repository.NewPostgresGroupRepository(pool)
	permissions := repository.NewPostgresPermissionRepository(pool)
	sessions := repository.NewPostgresSessionRepository(pool)

	// --- Securite --------------------------------------------------------
	hasher := security.NewBcryptHasher()

	// --- Cas d'usage -----------------------------------------------------
	// Injectes avec les interfaces (UserRepository, AuthMethodRepository,
	// domain.PasswordHasher...), jamais les types concrets ci-dessus.
	createUser := application.NewCreateUserUseCase(users)
	registerAuthMethod := application.NewRegisterAuthMethodUseCase(users, authMethods, hasher)
	registerUser := application.NewRegisterUserUseCase(createUser, registerAuthMethod, users)
	authenticate := application.NewAuthenticateUserUseCase(users, authMethods, hasher)
	authorization := application.NewAuthorizationService(users, roles, groups, permissions)
	createSession := application.NewCreateSessionUseCase(sessions, users, sessionTTL)
	logout := application.NewLogoutUseCase(sessions)

	// --- HTTP --------------------------------------------------------------
	userHandler := httpapi.NewUserHandler(registerUser, users)
	authHandler := httpapi.NewAuthHandler(registerAuthMethod, authenticate, createSession, logout)
	serviceHandler := httpapi.NewServiceHandler(services)
	permissionsHandler := httpapi.NewPermissionsHandler(authorization)
	permissionHandler := httpapi.NewPermissionHandler(permissions)
	roleHandler := httpapi.NewRoleHandler(roles)
	groupHandler := httpapi.NewGroupHandler(groups)
	serviceAuth := httpapi.NewServiceAuthMiddleware(services)
	userAuth := httpapi.NewUserAuthMiddleware(sessions, users)
	authz := httpapi.NewAuthorizationMiddleware(authorization)

	mux := httpapi.NewRouter(
		userHandler, authHandler, serviceHandler, permissionsHandler,
		permissionHandler, roleHandler, groupHandler,
		serviceAuth, userAuth, authz,
	)

	// CORS enveloppe TOUT le mux (pas une route precise) -- voir la doc
	// de CORSMiddleware pour pourquoi. "*" par defaut, pratique en
	// developpement local avec un frontend sur un autre port, mais a
	// restreindre via CORS_ALLOWED_ORIGINS (liste separee par des
	// virgules, ex: "http://localhost:5173,https://mon-frontend.example")
	// des que l'API est exposee au-dela de ta machine.
	cors := httpapi.NewCORSMiddleware(strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "*"), ","))
	handler := cors.Wrap(mux)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Serveur demarre sur le port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// getEnv lit une variable d'environnement, avec une valeur de repli si
// elle est absente ou vide.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
