package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id string) error

	AssignRole(ctx context.Context, userID, roleID string) error
	RevokeRole(ctx context.Context, userID, roleID string) error
	AssignPermission(ctx context.Context, userID, permissionID string) error
	RevokePermission(ctx context.Context, userID, permissionID string) error

	RecordLogin(ctx context.Context, userID string) error
}
