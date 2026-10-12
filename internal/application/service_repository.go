package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

type ServiceRepository interface {
	Create(ctx context.Context, s *domain.Service) error
	GetByID(ctx context.Context, id string) (*domain.Service, error)

	GetByAPIKeyHash(ctx context.Context, apiKeyHash string) (*domain.Service, error)

	List(ctx context.Context) ([]*domain.Service, error)

	Update(ctx context.Context, s *domain.Service) error

	UpdateAPIKeyHash(ctx context.Context, id, newAPIKeyHash string) error

	Delete(ctx context.Context, id string) error
}
