package application

import (
	"context"

	"user-service/internal/domain"
)

type AuthMethodRepository interface {
	Create(ctx context.Context, am *domain.AuthMethod) error
	GetByID(ctx context.Context, id string) (*domain.AuthMethod, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.AuthMethod, error)

	GetByProviderUserID(ctx context.Context, provider domain.Provider, providerUserID string) (*domain.AuthMethod, error)

	MarkUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
