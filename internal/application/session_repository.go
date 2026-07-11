package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

// SessionRepository definit le contrat de persistance pour Session.
type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error

	// GetByTokenHash retrouve une session par le hash de son jeton --
	// utilise par UserAuthMiddleware pour authentifier une requete entrante.
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)

	// Touch persiste la mise a jour de LastUsedAt (voir Session.Touch).
	Touch(ctx context.Context, id string) error

	// Delete revoque UNE session -- utilise pour la deconnexion.
	Delete(ctx context.Context, id string) error

	// DeleteAllForUser revoque TOUTES les sessions d'un usager -- utile
	// pour "se deconnecter partout" ou en cas de compromission suspectee
	// (ex: apres un changement de mot de passe).
	DeleteAllForUser(ctx context.Context, userID string) error
}
