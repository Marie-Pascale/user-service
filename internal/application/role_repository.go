package application

import (
	"context"

	"user-service/internal/domain"
)

type RoleRepository interface {
	Create(ctx context.Context, r *domain.Role) error
	GetByID(ctx context.Context, id string) (*domain.Role, error)
	List(ctx context.Context) ([]*domain.Role, error)

	Update(ctx context.Context, r *domain.Role) error

	GrantPermission(ctx context.Context, roleID, permissionID string) error
	RevokePermission(ctx context.Context, roleID, permissionID string) error
	Delete(ctx context.Context, id string) error
}
