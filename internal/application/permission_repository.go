package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

type PermissionRepository interface {
	Create(ctx context.Context, p *domain.Permission) error
	GetByID(ctx context.Context, id string) (*domain.Permission, error)
	ListByServiceID(ctx context.Context, serviceID string) ([]*domain.Permission, error)

	GetByServiceIDAndKey(ctx context.Context, serviceID, key string) (*domain.Permission, error)

	ListByIDs(ctx context.Context, ids []string) ([]*domain.Permission, error)

	Delete(ctx context.Context, id string) error
}
