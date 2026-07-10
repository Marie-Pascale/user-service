package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// ServiceRepository definit le contrat de persistance pour Service.
type ServiceRepository interface {
	Create(ctx context.Context, s *domain.Service) error
	GetByID(ctx context.Context, id string) (*domain.Service, error)

	// GetByAPIKeyHash retrouve le service correspondant a un hash de cle
	// API -- utilise pour authentifier les requetes entrantes d'un
	// service tiers de l'ecosysteme.
	GetByAPIKeyHash(ctx context.Context, apiKeyHash string) (*domain.Service, error)

	List(ctx context.Context) ([]*domain.Service, error)

	// UpdateAPIKeyHash persiste un nouveau hash suite a une rotation de cle.
	UpdateAPIKeyHash(ctx context.Context, id, newAPIKeyHash string) error

	Delete(ctx context.Context, id string) error
}
