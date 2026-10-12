package application

import (
	"context"

	"gestionnaire-usagers/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, s *domain.Session) error

	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)

	Touch(ctx context.Context, id string) error

	Delete(ctx context.Context, id string) error

	DeleteAllForUser(ctx context.Context, userID string) error
}
