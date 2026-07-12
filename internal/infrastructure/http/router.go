package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// NewRouter cable toutes les routes de l'API sur un ServeMux (Go 1.22+,
// avec le support natif des methodes HTTP et des variables de chemin --
// voir la discussion sur ServeMux plus tot dans le projet).
//
// Chaque middleware n'enveloppe QUE les routes qui en ont besoin -- pas le
// mux entier (voir la discussion sur le placement des middlewares plus tot
// dans le projet). Pour les routes usager protegees, l'ordre de chainage
// est TOUJOURS : session (authentification) -> permission (autorisation) ->
// handler. On ne peut pas verifier une permission sans savoir d'abord QUI
// fait la requete.
func NewRouter(
	userHandler *UserHandler,
	authHandler *AuthHandler,
	serviceHandler *ServiceHandler,
	permissionsHandler *PermissionsHandler,
	serviceAuth *ServiceAuthMiddleware,
	userAuth *UserAuthMiddleware,
	authz *AuthorizationMiddleware,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /docs", docsHandler)
	mux.HandleFunc("GET /openapi.yaml", specHandler)

	// Ouverte : auto-inscription, aucune authentification requise.
	mux.HandleFunc("POST /users", userHandler.Create)

	// Ouverte : il faut bien pouvoir se connecter avant d'avoir un jeton.
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protegees par session : usager authentifie requis pour toutes les
	// routes ci-dessous.
	mux.Handle("GET /me", userAuth.Authenticate(http.HandlerFunc(userHandler.Me)))
	mux.Handle("POST /auth/logout", userAuth.Authenticate(http.HandlerFunc(authHandler.Logout)))

	mux.Handle("GET /users",
		userAuth.Authenticate(authz.RequirePermission(application.PermUsersReadAny)(http.HandlerFunc(userHandler.List))))

	mux.Handle("GET /users/{id}",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersReadAny)(http.HandlerFunc(userHandler.Get))))

	mux.Handle("PUT /users/{id}",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersWriteAny)(http.HandlerFunc(userHandler.Update))))

	mux.Handle("DELETE /users/{id}",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersDeleteAny)(http.HandlerFunc(userHandler.Delete))))

	mux.Handle("POST /users/{id}/password",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersWriteAny)(http.HandlerFunc(authHandler.SetPassword))))

	mux.Handle("POST /services",
		userAuth.Authenticate(authz.RequirePermission(application.PermServicesCreate)(http.HandlerFunc(serviceHandler.Create))))

	// Protegee par cle API (services tiers, pas des usagers).
	mux.Handle("GET /users/{id}/permissions", serviceAuth.Authenticate(http.HandlerFunc(permissionsHandler.ListForUser)))

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
