package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type PostgresServiceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresServiceRepository(pool *pgxpool.Pool) *PostgresServiceRepository {
	return &PostgresServiceRepository{pool: pool}
}

var _ application.ServiceRepository = (*PostgresServiceRepository)(nil)

func (r *PostgresServiceRepository) Create(ctx context.Context, s *domain.Service) error {
	const query = `
		INSERT INTO services (id, name, description, api_key_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, s.ID, s.Name, s.Description, s.APIKeyHash, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *PostgresServiceRepository) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	const query = `
		SELECT id, name, description, api_key_hash, created_at, updated_at
		FROM services
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *PostgresServiceRepository) GetByAPIKeyHash(ctx context.Context, apiKeyHash string) (*domain.Service, error) {
	const query = `
		SELECT id, name, description, api_key_hash, created_at, updated_at
		FROM services
		WHERE api_key_hash = $1
	`
	return r.scanOne(ctx, query, apiKeyHash)
}

func (r *PostgresServiceRepository) scanOne(ctx context.Context, query string, arg any) (*domain.Service, error) {
	var s domain.Service
	err := r.pool.QueryRow(ctx, query, arg).Scan(&s.ID, &s.Name, &s.Description, &s.APIKeyHash, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresServiceRepository) List(ctx context.Context) ([]*domain.Service, error) {
	const query = `
		SELECT id, name, description, api_key_hash, created_at, updated_at
		FROM services
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []*domain.Service{}
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.APIKeyHash, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, &s)
	}
	return services, rows.Err()
}

func (r *PostgresServiceRepository) Update(ctx context.Context, s *domain.Service) error {
	const query = `UPDATE services SET name = $2, description = $3, updated_at = $4 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, s.ID, s.Name, s.Description, s.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

// UpdateAPIKeyHash met aussi a jour updated_at -- domain.Service.RotateAPIKey()
// avance deja ce champ en memoire, cette requete synchronise la base pour
// que ce ne soit pas seulement vrai le temps de la requete en cours (meme
// remarque que pour RoleRepository.touch/GroupRepository.touch).
func (r *PostgresServiceRepository) UpdateAPIKeyHash(ctx context.Context, id, newAPIKeyHash string) error {
	const query = `UPDATE services SET api_key_hash = $2, updated_at = now() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id, newAPIKeyHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresServiceRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM services WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}
