package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// healthResponse est la structure retournee par l'endpoint /health
type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func main() {
	mux := http.NewServeMux()

	// Endpoint de base pour verifier que le serveur tourne
	mux.HandleFunc("GET /health", healthHandler)

	// Les routes du gestionnaire d'usagers seront ajoutees ici,
	// par exemple :
	// mux.HandleFunc("GET /users", listUsersHandler)
	// mux.HandleFunc("POST /users", createUserHandler)
	// mux.HandleFunc("GET /users/{id}", getUserHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur demarre sur le port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}