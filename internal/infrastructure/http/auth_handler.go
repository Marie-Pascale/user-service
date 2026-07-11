package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// AuthHandler regroupe les endpoints lies a l'authentification.
type AuthHandler struct {
	RegisterAuthMethod *application.RegisterAuthMethodUseCase
	Authenticate       *application.AuthenticateUserUseCase
	CreateSession      *application.CreateSessionUseCase
	Logout             *application.LogoutUseCase
}

func NewAuthHandler(
	registerAuthMethod *application.RegisterAuthMethodUseCase,
	authenticate *application.AuthenticateUserUseCase,
	createSession *application.CreateSessionUseCase,
	logout *application.LogoutUseCase,
) *AuthHandler {
	return &AuthHandler{
		RegisterAuthMethod: registerAuthMethod,
		Authenticate:       authenticate,
		CreateSession:      createSession,
		Logout:             logout,
	}
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

// Login gere POST /auth/login. Verifie les identifiants PUIS cree une
// session -- deux cas d'usage distincts appeles en sequence (voir
// CreateSessionUseCase pour pourquoi ils sont separes).
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

	session, plainToken, err := h.CreateSession.Execute(r.Context(), u.ID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		User:      newUserResponse(u),
		Token:     plainToken,
		ExpiresAt: session.ExpiresAt,
	})
}

// Logout gere POST /auth/logout -- protegee par UserAuthMiddleware, donc
// la session a revoquer vient du contexte, jamais d'un ID fourni par le
// client (sinon n'importe qui pourrait revoquer la session de n'importe
// qui d'autre en devinant un ID).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := SessionFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated"})
		return
	}

	if err := h.Logout.Execute(r.Context(), session.ID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
