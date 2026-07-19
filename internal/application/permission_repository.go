package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// PermissionRepository definit le contrat de persistance pour Permission.
type PermissionRepository interface {
	Create(ctx context.Context, p *domain.Permission) error
	GetByID(ctx context.Context, id string) (*domain.Permission, error)
	ListByServiceID(ctx context.Context, serviceID string) ([]*domain.Permission, error)

	// GetByServiceIDAndKey retrouve une permission par sa cle, au sein
	// d'un service donne (la contrainte UNIQUE(service_id, key) en base
	// garantit qu'il y en a au plus une). Utilise pour l'auto-enregistrement
	// idempotent des services tiers -- voir RegisterPermissionUseCase.
	GetByServiceIDAndKey(ctx context.Context, serviceID, key string) (*domain.Permission, error)

	// ListByIDs charge plusieurs permissions d'un coup, a partir d'une
	// liste d'IDs -- utile pour resoudre les permissions d'un role sans
	// faire une requete par permission (evite le probleme des N+1 requetes).
	ListByIDs(ctx context.Context, ids []string) ([]*domain.Permission, error)

	Delete(ctx context.Context, id string) error
}
