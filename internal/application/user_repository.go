package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// UserRepository definit le contrat de persistance pour User. Cette
// interface vit dans la couche application (qui l'utilise), pas dans
// l'infrastructure (qui l'implemente) -- Dependency Inversion Principle :
// la couche de haut niveau definit ce dont elle a besoin, la couche de
// bas niveau s'y conforme. Ca permet de tester les cas d'usage de
// l'application avec un faux repository en memoire, sans jamais toucher
// a Postgres.
//
// GetByID et GetByEmail retournent ErrNotFound si aucun usager ne
// correspond.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id string) error

	// AssignRole/RevokeRole et AssignPermission/RevokePermission persistent
	// les assignations DIRECTES d'un usager (voir User.RoleIDs,
	// User.PermissionIDs) dans les tables de jonction user_roles et
	// user_permissions. Ecrivent directement en base, sans passer par
	// Update -- meme patron que RoleRepository.GrantPermission et
	// GroupRepository.AssignRole.
	AssignRole(ctx context.Context, userID, roleID string) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	AssignPermission(ctx context.Context, userID, permissionID string) error
	RevokePermission(ctx context.Context, userID, permissionID string) error
}
