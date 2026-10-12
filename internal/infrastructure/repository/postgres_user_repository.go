package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"user-service/internal/application"
	"user-service/internal/domain"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

var _ application.UserRepository = (*PostgresUserRepository)(nil)

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	const query = `
		INSERT INTO users (id, name, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, u.ID, u.Name, u.Email, string(u.Status), u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `
		SELECT id, name, email, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, name, email, status, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`
	return r.scanOne(ctx, query, email)
}

func (r *PostgresUserRepository) scanOne(ctx context.Context, query string, arg any) (*domain.User, error) {
	var u domain.User
	var status string

	err := r.pool.QueryRow(ctx, query, arg).Scan(&u.ID, &u.Name, &u.Email, &status, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}
	u.Status = domain.Status(status)

	u.RoleIDs, err = r.idsFor(ctx, "user_roles", "role_id", u.ID)
	if err != nil {
		return nil, err
	}

	u.PermissionIDs, err = r.idsFor(ctx, "user_permissions", "permission_id", u.ID)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *PostgresUserRepository) idsFor(ctx context.Context, table, column, userID string) ([]string, error) {
	query := `SELECT ` + column + ` FROM ` + table + ` WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	const query = `
		SELECT id, name, email, status, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*domain.User{}
	for rows.Next() {
		var u domain.User
		var status string
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &status, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt); err != nil {
			return nil, err
		}
		u.Status = domain.Status(status)
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	const query = `
		UPDATE users
		SET name = $2, email = $3, status = $4, updated_at = $5
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, u.ID, u.Name, u.Email, string(u.Status), u.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM users WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresUserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	const query = `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *PostgresUserRepository) RevokeRole(ctx context.Context, userID, roleID string) error {
	const query = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

func (r *PostgresUserRepository) AssignPermission(ctx context.Context, userID, permissionID string) error {
	const query = `INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, userID, permissionID)
	return err
}

func (r *PostgresUserRepository) RevokePermission(ctx context.Context, userID, permissionID string) error {
	const query = `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, permissionID)
	return err
}

func (r *PostgresUserRepository) RecordLogin(ctx context.Context, userID string) error {
	const query = `UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}
