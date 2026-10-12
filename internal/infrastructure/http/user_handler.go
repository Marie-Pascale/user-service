package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type UserHandler struct {
	RegisterUser *application.RegisterUserUseCase
	Users        application.UserRepository
}

func NewUserHandler(registerUser *application.RegisterUserUseCase, users application.UserRepository) *UserHandler {
	return &UserHandler{RegisterUser: registerUser, Users: users}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	provider := domain.Provider(req.AuthMethod.Provider)
	if provider == "" {
		provider = domain.ProviderLocal
	}

	switch provider {
	case domain.ProviderLocal:
		u, err := h.RegisterUser.ExecuteLocal(r.Context(), req.Name, req.Email, req.AuthMethod.Password)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, newUserResponse(u))

	case domain.ProviderGoogle, domain.ProviderFacebook:
		writeJSON(w, http.StatusNotImplemented, errorResponse{
			Error: "l'inscription via ce fournisseur n'est pas encore disponible",
		})

	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "auth_method.provider doit etre \"local\" (seul fournisseur supporte pour l'instant)",
		})
	}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	u, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(u))
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.Users.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]userResponse, len(users))
	for i, u := range users {
		resp[i] = newUserResponse(u)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req updateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	u, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := u.UpdateProfile(req.Name, req.Email); err != nil {
		writeError(w, err)
		return
	}

	if err := h.Users.Update(r.Context(), u); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(u))
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.Users.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated"})
		return
	}
	writeJSON(w, http.StatusOK, newUserResponse(u))
}

func (h *UserHandler) GetBasicInfo(w http.ResponseWriter, r *http.Request) {
	if _, ok := ServiceFromContext(r.Context()); !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated service"})
		return
	}

	u, err := h.Users.GetByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newUserBasicInfoResponse(u))
}
