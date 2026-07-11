package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"gestionnaire-usagers/internal/application"
	httpapi "gestionnaire-usagers/internal/infrastructure/http"
	"gestionnaire-usagers/internal/infrastructure/repository"
	"gestionnaire-usagers/internal/infrastructure/security"
)

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

	// --- Securite --------------------------------------------------------
	hasher := security.NewBcryptHasher()

	// --- Cas d'usage -----------------------------------------------------
	// Injectes avec les interfaces (UserRepository, AuthMethodRepository,
	// domain.PasswordHasher), jamais les types concrets ci-dessus.
	createUser := application.NewCreateUserUseCase(users)
	registerAuthMethod := application.NewRegisterAuthMethodUseCase(users, authMethods, hasher)
	authenticate := application.NewAuthenticateUserUseCase(users, authMethods, hasher)
	authorization := application.NewAuthorizationService(users, roles, groups, permissions)

	// --- HTTP --------------------------------------------------------------
	userHandler := httpapi.NewUserHandler(createUser, users)
	authHandler := httpapi.NewAuthHandler(registerAuthMethod, authenticate)
	serviceHandler := httpapi.NewServiceHandler(services)
	permissionsHandler := httpapi.NewPermissionsHandler(authorization)
	serviceAuth := httpapi.NewServiceAuthMiddleware(services)

	mux := httpapi.NewRouter(userHandler, authHandler, serviceHandler, permissionsHandler, serviceAuth)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Serveur demarre sur le port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
