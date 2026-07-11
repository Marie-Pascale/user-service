package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

var _ application.SessionRepository = (*PostgresSessionRepository)(nil)

func (r *PostgresSessionRepository) Create(ctx context.Context, s *domain.Session) error {
	const query = `
		INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, s.ID, s.UserID, s.TokenHash, s.CreatedAt, s.ExpiresAt, s.LastUsedAt)
	return err
}

func (r *PostgresSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	const query = `
		SELECT id, user_id, token_hash, created_at, expires_at, last_used_at
		FROM sessions
		WHERE token_hash = $1
	`
	var s domain.Session
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.CreatedAt, &s.ExpiresAt, &s.LastUsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresSessionRepository) Touch(ctx context.Context, id string) error {
	const query = `UPDATE sessions SET last_used_at = now() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresSessionRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM sessions WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresSessionRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	const query = `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
