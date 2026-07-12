package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// UserHandler regroupe les endpoints lies a User. Il ne connait que des
// interfaces (les cas d'usage, l'interface UserRepository pour les
// lectures simples) -- jamais PostgresUserRepository directement.
type UserHandler struct {
	RegisterUser *application.RegisterUserUseCase
	Users        application.UserRepository
}

func NewUserHandler(registerUser *application.RegisterUserUseCase, users application.UserRepository) *UserHandler {
	return &UserHandler{RegisterUser: registerUser, Users: users}
}

// Create gere POST /users. La methode d'authentification est fournie DES
// LA CREATION -- plus d'usager "orphelin" sans aucun moyen de se
// connecter (voir RegisterUserUseCase).
//
// auth_method.provider est optionnel et vaut "local" par defaut -- le cas
// le plus courant n'a donc pas besoin de le preciser explicitement.
//
// Seul le provider "local" est accepte pour l'instant. Les fournisseurs
// tiers (Google, Facebook) sont volontairement refuses ici : accepter un
// "provider_user_id" venant directement du client, sans jamais verifier
// aupres du fournisseur que ce jeton/identifiant est authentique,
// permettrait a n'importe qui de pretendre etre n'importe quel compte
// Google. RegisterUserUseCase.ExecuteOAuth existe deja et fonctionne --
// il attend seulement un vrai flot de verification OAuth en amont
// (echange de code, validation de jeton signe) avant d'etre branche ici.
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

// Me gere GET /me -- protegee par UserAuthMiddleware. C'est l'endpoint
// evoque tout au debut du projet : un usager peut voir SON PROPRE profil
// sans connaitre (ni avoir besoin de connaitre) son propre ID -- celui-ci
// vient du contexte, pose par le middleware a partir de la session, jamais
// d'un parametre d'URL.
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthenticated"})
		return
	}
	writeJSON(w, http.StatusOK, newUserResponse(u))
}
