package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

type AuthHandler struct {
	RegisterAuthMethod *application.RegisterAuthMethodUseCase
	Authenticate       *application.AuthenticateUserUseCase
	CreateSession      *application.CreateSessionUseCase
	LogoutUseCase      *application.LogoutUseCase
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
		LogoutUseCase:      logout,
	}
}

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

	session, plainToken, err := h.CreateSession.Execute(r.Context(), u.ID, clientIP(r), r.Header.Get("User-Agent"))
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := SessionFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated"})
		return
	}

	if err := h.LogoutUseCase.Execute(r.Context(), session.ID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
