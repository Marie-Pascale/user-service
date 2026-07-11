package httpapi

import (
	"time"

	"gestionnaire-usagers/internal/domain"
)

// Ce fichier contient les structures de serialisation JSON (requetes et
// reponses), volontairement separees des entites du domaine. Le domaine
// ne doit rien savoir de JSON ou de HTTP -- c'est cette couche qui traduit
// entre les deux mondes.

// --- Reponses ------------------------------------------------------------

type userResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Status        string    `json:"status"`
	RoleIDs       []string  `json:"role_ids"`
	PermissionIDs []string  `json:"permission_ids"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Status:        string(u.Status),
		RoleIDs:       u.RoleIDs,
		PermissionIDs: u.PermissionIDs,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// authMethodResponse n'expose JAMAIS PasswordHash -- meme hache, il n'a
// rien a faire dans une reponse HTTP.
type authMethodResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

func newAuthMethodResponse(am *domain.AuthMethod) authMethodResponse {
	return authMethodResponse{
		ID:         am.ID,
		UserID:     am.UserID,
		Provider:   string(am.Provider),
		CreatedAt:  am.CreatedAt,
		LastUsedAt: am.LastUsedAt,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

// --- Requetes --------------------------------------------------------

type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type updateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type setPasswordRequest struct {
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// createServiceResponse est le SEUL endroit ou APIKey (en clair) apparait
// jamais dans une reponse HTTP -- exactement comme domain.NewService le
// garantit deja au niveau du domaine (voir domain/service.go).
type createServiceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	APIKey      string `json:"api_key"`
}
