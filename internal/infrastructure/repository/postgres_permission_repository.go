package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type PostgresPermissionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPermissionRepository(pool *pgxpool.Pool) *PostgresPermissionRepository {
	return &PostgresPermissionRepository{pool: pool}
}

var _ application.PermissionRepository = (*PostgresPermissionRepository)(nil)

func (r *PostgresPermissionRepository) Create(ctx context.Context, p *domain.Permission) error {
	const query = `
		INSERT INTO permissions (id, service_id, key, description)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, p.ID, p.ServiceID, p.Key, p.Description)
	return err
}

func (r *PostgresPermissionRepository) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	const query = `
		SELECT id, service_id, key, description
		FROM permissions
		WHERE id = $1
	`
	var p domain.Permission
	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.ServiceID, &p.Key, &p.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresPermissionRepository) ListByServiceID(ctx context.Context, serviceID string) ([]*domain.Permission, error) {
	const query = `
		SELECT id, service_id, key, description
		FROM permissions
		WHERE service_id = $1
		ORDER BY key ASC
	`
	return r.scanMany(ctx, query, serviceID)
}

func (r *PostgresPermissionRepository) ListByIDs(ctx context.Context, ids []string) ([]*domain.Permission, error) {
	if len(ids) == 0 {
		return []*domain.Permission{}, nil
	}
	const query = `
		SELECT id, service_id, key, description
		FROM permissions
		WHERE id = ANY($1)
	`
	return r.scanMany(ctx, query, ids)
}

func (r *PostgresPermissionRepository) scanMany(ctx context.Context, query string, arg any) ([]*domain.Permission, error) {
	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []*domain.Permission{}
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.ServiceID, &p.Key, &p.Description); err != nil {
			return nil, err
		}
		permissions = append(permissions, &p)
	}
	return permissions, rows.Err()
}

func (r *PostgresPermissionRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM permissions WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}
