package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type PostgresAuthMethodRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAuthMethodRepository(pool *pgxpool.Pool) *PostgresAuthMethodRepository {
	return &PostgresAuthMethodRepository{pool: pool}
}

var _ application.AuthMethodRepository = (*PostgresAuthMethodRepository)(nil)

func (r *PostgresAuthMethodRepository) Create(ctx context.Context, am *domain.AuthMethod) error {
	const query = `
		INSERT INTO auth_methods (id, user_id, provider, password_hash, provider_user_id, created_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		am.ID, am.UserID, string(am.Provider),
		nullIfEmpty(am.PasswordHash), nullIfEmpty(am.ProviderUserID),
		am.CreatedAt, am.LastUsedAt,
	)
	return err
}

func (r *PostgresAuthMethodRepository) GetByID(ctx context.Context, id string) (*domain.AuthMethod, error) {
	const query = `
		SELECT id, user_id, provider, password_hash, provider_user_id, created_at, last_used_at
		FROM auth_methods
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *PostgresAuthMethodRepository) GetByProviderUserID(ctx context.Context, provider domain.Provider, providerUserID string) (*domain.AuthMethod, error) {
	const query = `
		SELECT id, user_id, provider, password_hash, provider_user_id, created_at, last_used_at
		FROM auth_methods
		WHERE provider = $1 AND provider_user_id = $2
	`
	return r.scanOne(ctx, query, string(provider), providerUserID)
}

func (r *PostgresAuthMethodRepository) scanOne(ctx context.Context, query string, args ...any) (*domain.AuthMethod, error) {
	var am domain.AuthMethod
	var provider string
	var passwordHash, providerUserID *string

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&am.ID, &am.UserID, &provider, &passwordHash, &providerUserID, &am.CreatedAt, &am.LastUsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	am.Provider = domain.Provider(provider)
	if passwordHash != nil {
		am.PasswordHash = *passwordHash
	}
	if providerUserID != nil {
		am.ProviderUserID = *providerUserID
	}
	return &am, nil
}

func (r *PostgresAuthMethodRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.AuthMethod, error) {
	const query = `
		SELECT id, user_id, provider, password_hash, provider_user_id, created_at, last_used_at
		FROM auth_methods
		WHERE user_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	methods := []*domain.AuthMethod{}
	for rows.Next() {
		var am domain.AuthMethod
		var provider string
		var passwordHash, providerUserID *string

		if err := rows.Scan(&am.ID, &am.UserID, &provider, &passwordHash, &providerUserID, &am.CreatedAt, &am.LastUsedAt); err != nil {
			return nil, err
		}
		am.Provider = domain.Provider(provider)
		if passwordHash != nil {
			am.PasswordHash = *passwordHash
		}
		if providerUserID != nil {
			am.ProviderUserID = *providerUserID
		}
		methods = append(methods, &am)
	}
	return methods, rows.Err()
}

func (r *PostgresAuthMethodRepository) MarkUsed(ctx context.Context, id string) error {
	const query = `UPDATE auth_methods SET last_used_at = now() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresAuthMethodRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM auth_methods WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
