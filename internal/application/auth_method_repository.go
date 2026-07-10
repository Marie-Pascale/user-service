package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// AuthMethodRepository definit le contrat de persistance pour AuthMethod.
type AuthMethodRepository interface {
	Create(ctx context.Context, am *domain.AuthMethod) error
	GetByID(ctx context.Context, id string) (*domain.AuthMethod, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.AuthMethod, error)

	// GetByProviderUserID retrouve la methode d'authentification tierce
	// (Google, Facebook) correspondant a l'identifiant chez ce fournisseur --
	// utilise lors de la connexion OAuth, pour savoir a quel usager local
	// correspond le compte tiers.
	GetByProviderUserID(ctx context.Context, provider domain.Provider, providerUserID string) (*domain.AuthMethod, error)

	MarkUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}
