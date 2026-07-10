package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// UserHandler regroupe les endpoints lies a User. Il ne connait que des
// interfaces (le cas d'usage CreateUserUseCase, l'interface UserRepository
// pour les lectures simples) -- jamais PostgresUserRepository directement.
type UserHandler struct {
	CreateUser *application.CreateUserUseCase
	Users      application.UserRepository
}

func NewUserHandler(createUser *application.CreateUserUseCase, users application.UserRepository) *UserHandler {
	return &UserHandler{CreateUser: createUser, Users: users}
}

// Create gere POST /users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corps de requete JSON invalide"})
		return
	}

	u, err := h.CreateUser.Execute(r.Context(), req.Name, req.Email)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(u))
}

// Get gere GET /users/{id}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	u, err := h.Users.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(u))
}

// List gere GET /users
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

// Update gere PUT /users/{id}
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

// Delete gere DELETE /users/{id}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.Users.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
