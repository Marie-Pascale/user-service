package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"user-service/internal/application"
	httpapi "user-service/internal/infrastructure/http"
	"user-service/internal/infrastructure/repository"
	"user-service/internal/infrastructure/security"
)

const sessionTTL = 24 * time.Hour

func main() {
	ctx := context.Background()

	pool, err := repository.NewPool(ctx, repository.ConfigFromEnv())
	if err != nil {
		log.Fatalf("connexion a la base de donnees: %v", err)
	}
	defer pool.Close()

	users := repository.NewPostgresUserRepository(pool)
	authMethods := repository.NewPostgresAuthMethodRepository(pool)
	services := repository.NewPostgresServiceRepository(pool)
	roles := repository.NewPostgresRoleRepository(pool)
	groups := repository.NewPostgresGroupRepository(pool)
	permissions := repository.NewPostgresPermissionRepository(pool)
	sessions := repository.NewPostgresSessionRepository(pool)

	hasher := security.NewBcryptHasher()

	createUser := application.NewCreateUserUseCase(users)
	registerAuthMethod := application.NewRegisterAuthMethodUseCase(users, authMethods, hasher)
	registerUser := application.NewRegisterUserUseCase(createUser, registerAuthMethod, users)
	authenticate := application.NewAuthenticateUserUseCase(users, authMethods, hasher)
	authorization := application.NewAuthorizationService(users, roles, groups, permissions)
	createSession := application.NewCreateSessionUseCase(sessions, users, sessionTTL)
	logout := application.NewLogoutUseCase(sessions)

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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
