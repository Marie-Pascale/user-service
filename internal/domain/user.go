package domain

import (
	"errors"
	"strings"
	"time"
)

// Erreurs de validation du domaine. Elles sont volontairement generiques
// (pas de code HTTP ici) : c'est a la couche infrastructure de les traduire
// en reponses HTTP appropriees (ex: 400 Bad Request).
var (
	ErrEmptyName    = errors.New("name is required")
	ErrInvalidEmail = errors.New("invalid email")
)

// Status represente l'etat du compte d'un usager.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// User est l'entite centrale du domaine. Elle ne depend d'aucune
// bibliotheque externe (pas de driver Postgres, pas de net/http) :
// c'est ce qui la rend testable et reutilisable independamment
// de la couche de stockage ou de transport.
//
// Notez que User ne contient aucune information d'authentification
// (mot de passe, etc.) -- celle-ci vit dans AuthMethod, une entite
// separee, pour permettre plusieurs methodes d'authentification par
// usager (mot de passe local, Google, Facebook...).
//
// RoleIDs et PermissionIDs representent les assignations *directes*.
// Les permissions effectives d'un usager sont la resultante de ses
// permissions directes + celles de ses roles directs + celles des
// roles herites de ses groupes. Ce calcul est fait par la couche
// application (voir AuthorizationService a venir), pas ici : User
// ne stocke que des IDs, pas les entites Role/Permission completes.
type User struct {
	ID            string
	Name          string
	Email         string
	Status        Status
	RoleIDs       []string
	PermissionIDs []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewUser cree un usager valide ou retourne une erreur de validation.
// C'est le point d'entree unique pour construire un User : ca garantit
// qu'aucun usager invalide ne peut exister ailleurs dans le systeme.
func NewUser(id, name, email string) (*User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	now := time.Now().UTC()
	return &User{
		ID:            id,
		Name:          name,
		Email:         email,
		Status:        StatusActive,
		RoleIDs:       []string{},
		PermissionIDs: []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
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

// IsActive indique si le compte de l'usager est actif.
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// Suspend desactive le compte de l'usager.
func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now().UTC()
}

// AssignRole assigne un role directement a l'usager, s'il ne l'a pas deja.
func (u *User) AssignRole(roleID string) {
	for _, id := range u.RoleIDs {
		if id == roleID {
			return
		}
	}
	u.RoleIDs = append(u.RoleIDs, roleID)
	u.UpdatedAt = time.Now().UTC()
}

// AssignPermission accorde une permission directement a l'usager,
// en plus de celles heritees de ses roles.
func (u *User) AssignPermission(permissionID string) {
	for _, id := range u.PermissionIDs {
		if id == permissionID {
			return
		}
	}
	u.PermissionIDs = append(u.PermissionIDs, permissionID)
	u.UpdatedAt = time.Now().UTC()
}

// UpdateProfile modifie le nom et le courriel d'un usager existant, avec
// la meme validation que NewUser. Centralise cette regle ici plutot que
// de la dupliquer (ou de reconstruire un User via NewUser, ce qui
// ecraserait aussi Status/RoleIDs/PermissionIDs/CreatedAt par erreur).
func (u *User) UpdateProfile(name, email string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if !isValidEmail(email) {
		return ErrInvalidEmail
	}

	u.Name = name
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	return nil
}
