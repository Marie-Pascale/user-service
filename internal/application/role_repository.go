package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// RoleRepository definit le contrat de persistance pour Role.
//
// GrantPermission/RevokePermission persistent les changements sur la table
// de jonction role_permissions -- ils correspondent directement aux
// methodes du meme nom sur domain.Role, mais celles-ci ne modifient que
// l'objet en memoire. Le cas d'usage appelant doit typiquement appeler les
// deux : la methode du domaine (pour la coherence de l'objet retourne) et
// celle du repository (pour la persistance).
type RoleRepository interface {
	Create(ctx context.Context, r *domain.Role) error
	GetByID(ctx context.Context, id string) (*domain.Role, error)
	List(ctx context.Context) ([]*domain.Role, error)

	// Update persiste Name/Description uniquement -- PermissionIDs est
	// gere separement via GrantPermission/RevokePermission.
	Update(ctx context.Context, r *domain.Role) error

	GrantPermission(ctx context.Context, roleID, permissionID string) error
	RevokePermission(ctx context.Context, roleID, permissionID string) error
	Delete(ctx context.Context, id string) error
}
