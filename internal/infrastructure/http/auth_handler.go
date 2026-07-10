package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// AuthHandler regroupe les endpoints lies a l'authentification. Il ne
// retourne pas encore de token de session -- ca viendra avec le middleware
// d'authentification, l'etape suivante. Pour l'instant, Login confirme
// juste que les identifiants sont valides et retourne l'usager.
type AuthHandler struct {
	RegisterAuthMethod *application.RegisterAuthMethodUseCase
	Authenticate       *application.AuthenticateUserUseCase
}

func NewAuthHandler(registerAuthMethod *application.RegisterAuthMethodUseCase, authenticate *application.AuthenticateUserUseCase) *AuthHandler {
	return &AuthHandler{RegisterAuthMethod: registerAuthMethod, Authenticate: authenticate}
}

// SetPassword gere POST /users/{id}/password -- ajoute une methode
// d'authentification locale a un usager existant.
func (h *AuthHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	var req setPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	am, err := h.RegisterAuthMethod.ExecuteLocal(r.Context(), userID, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newAuthMethodResponse(am))
}

// Login gere POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	u, err := h.Authenticate.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(u))
}
