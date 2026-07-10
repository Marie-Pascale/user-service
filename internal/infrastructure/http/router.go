package httpapi

import "net/http"

// NewRouter cable toutes les routes de l'API sur un ServeMux (Go 1.22+,
// avec le support natif des methodes HTTP et des variables de chemin --
// voir la discussion sur ServeMux plus tot dans le projet).
func NewRouter(userHandler *UserHandler, authHandler *AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("GET /users/{id}", userHandler.Get)
	mux.HandleFunc("PUT /users/{id}", userHandler.Update)
	mux.HandleFunc("DELETE /users/{id}", userHandler.Delete)

	mux.HandleFunc("POST /users/{id}/password", authHandler.SetPassword)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
