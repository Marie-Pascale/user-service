package domain

import (
	"errors"
	"strings"
	"time"
)

// Role represente le role d'un usager dans le systeme.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// Erreurs de validation du domaine. Elles sont volontairement generiques
// (pas de code HTTP ici) : c'est a la couche infrastructure de les traduire
// en reponses HTTP appropriees (ex: 400 Bad Request).
var (
	ErrEmptyName    = errors.New("name is required")
	ErrInvalidEmail = errors.New("invalid email")
	ErrInvalidRole  = errors.New("invalid role")
)

// User est l'entite centrale du domaine. Elle ne depend d'aucune
// bibliotheque externe (pas de driver Postgres, pas de net/http) :
// c'est ce qui la rend testable et reutilisable independamment
// de la couche de stockage ou de transport.
type User struct {
	ID        string
	Name      string
	Email     string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser cree un usager valide ou retourne une erreur de validation.
// C'est le point d'entree unique pour construire un User : ca garantit
// qu'aucun usager invalide ne peut exister ailleurs dans le systeme.
func NewUser(id, name, email string, role Role) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	if role != RoleAdmin && role != RoleUser {
		return nil, ErrInvalidRole
	}

	now := time.Now().UTC()
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// isValidEmail fait une verification simple, pas une validation RFC complete.
// Suffisant pour un projet d'apprentissage ; a remplacer par une regex plus
// stricte ou une librairie dediee si le besoin devient reel.
func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	domain := email[at+1:]
	return strings.Contains(domain, ".")
}

// IsAdmin indique si l'usager a le role administrateur.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanView encode la regle metier dont on parlait plus tot : un admin peut
// voir n'importe quel usager, un usager normal ne peut voir que lui-meme.
// Cette regle vit ici, dans le domaine, plutot que dans un handler HTTP ou
// un middleware -- elle pourra etre testee et reutilisee sans dependre
// du transport.
func (u *User) CanView(target *User) bool {
	if u.IsAdmin() {
		return true
	}
	return u.ID == target.ID
}
