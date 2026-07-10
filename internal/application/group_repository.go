package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// GroupRepository definit le contrat de persistance pour Group.
type GroupRepository interface {
	Create(ctx context.Context, g *domain.Group) error
	GetByID(ctx context.Context, id string) (*domain.Group, error)
	List(ctx context.Context) ([]*domain.Group, error)
	AddMember(ctx context.Context, groupID, userID string) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	AssignRole(ctx context.Context, groupID, roleID string) error
	Delete(ctx context.Context, id string) error
}
