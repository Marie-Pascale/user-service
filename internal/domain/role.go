package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrEmptyRoleName = errors.New("role name is required")

// Role est un regroupement nomme de permissions. Un role peut couvrir
// des permissions de plusieurs services differents (ex: un role
// "support-agent" pourrait combiner "orders:read" et "billing:read"
// meme si ces deux permissions viennent de services distincts).
//
// Un role est assigne a un User ou a un Group (voir User.RoleIDs et
// Group.RoleIDs) -- cette entite ne connait pas ses assignations,
// seulement les permissions qu'elle regroupe.
type Role struct {
	ID            string
	Name          string
	Description   string
	PermissionIDs []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewRole cree un role valide, sans permission au depart.
func NewRole(id, name, description string) (*Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyRoleName
	}

	now := time.Now().UTC()
	return &Role{
		ID:            id,
		Name:          name,
		Description:   description,
		PermissionIDs: []string{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// GrantPermission ajoute une permission au role, si elle n'y est pas deja.
func (r *Role) GrantPermission(permissionID string) {
	for _, id := range r.PermissionIDs {
		if id == permissionID {
			return
		}
	}
	r.PermissionIDs = append(r.PermissionIDs, permissionID)
	r.UpdatedAt = time.Now().UTC()
}

// RevokePermission retire une permission du role.
func (r *Role) RevokePermission(permissionID string) {
	for i, id := range r.PermissionIDs {
		if id == permissionID {
			r.PermissionIDs = append(r.PermissionIDs[:i], r.PermissionIDs[i+1:]...)
			r.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// HasPermission indique si le role regroupe une permission donnee.
func (r *Role) HasPermission(permissionID string) bool {
	for _, id := range r.PermissionIDs {
		if id == permissionID {
			return true
		}
	}
	return false
}

// UpdateDetails modifie le nom et la description d'un role existant, avec
// la meme validation que NewRole. Ne touche jamais PermissionIDs -- ca
// reste gere separement via GrantPermission/RevokePermission.
func (r *Role) UpdateDetails(name, description string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyRoleName
	}
	r.Name = name
	r.Description = description
	r.UpdatedAt = time.Now().UTC()
	return nil
}
