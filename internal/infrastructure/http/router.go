package httpapi

import "net/http"

// NewRouter cable toutes les routes de l'API sur un ServeMux (Go 1.22+,
// avec le support natif des methodes HTTP et des variables de chemin --
// voir la discussion sur ServeMux plus tot dans le projet).
//
// Chaque middleware n'enveloppe QUE les routes qui en ont besoin -- pas le
// mux entier. Envelopper tout le mux protegerait aussi /health ou
// POST /users, ce qui n'est pas voulu (voir la discussion sur le
// placement des middlewares plus tot dans le projet).
func NewRouter(
	userHandler *UserHandler,
	authHandler *AuthHandler,
	serviceHandler *ServiceHandler,
	permissionsHandler *PermissionsHandler,
	serviceAuth *ServiceAuthMiddleware,
	userAuth *UserAuthMiddleware,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("GET /users/{id}", userHandler.Get)
	mux.HandleFunc("PUT /users/{id}", userHandler.Update)
	mux.HandleFunc("DELETE /users/{id}", userHandler.Delete)

	mux.HandleFunc("POST /users/{id}/password", authHandler.SetPassword)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("POST /services", serviceHandler.Create)

	// Routes protegees par cle API (services tiers).
	mux.Handle("GET /users/{id}/permissions", serviceAuth.Authenticate(http.HandlerFunc(permissionsHandler.ListForUser)))

	// Routes protegees par session (usagers).
	mux.Handle("GET /me", userAuth.Authenticate(http.HandlerFunc(userHandler.Me)))
	mux.Handle("POST /auth/logout", userAuth.Authenticate(http.HandlerFunc(authHandler.Logout)))

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
