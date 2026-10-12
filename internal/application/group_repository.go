package application

import (
	"context"

	"user-service/internal/domain"
)

type GroupRepository interface {
	Create(ctx context.Context, g *domain.Group) error
	GetByID(ctx context.Context, id string) (*domain.Group, error)
	List(ctx context.Context) ([]*domain.Group, error)

	ListByMemberUserID(ctx context.Context, userID string) ([]*domain.Group, error)

	Update(ctx context.Context, g *domain.Group) error

	AddMember(ctx context.Context, groupID, userID string) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	AssignRole(ctx context.Context, groupID, roleID string) error
	Delete(ctx context.Context, id string) error
}
