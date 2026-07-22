package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

type PostgresRoleRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRoleRepository(pool *pgxpool.Pool) *PostgresRoleRepository {
	return &PostgresRoleRepository{pool: pool}
}

var _ application.RoleRepository = (*PostgresRoleRepository)(nil)

func (r *PostgresRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	const query = `
		INSERT INTO roles (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, role.ID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt)
	return err
}

func (r *PostgresRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	const query = `SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1`

	var role domain.Role
	err := r.pool.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	permissionIDs, err := r.permissionIDsFor(ctx, id)
	if err != nil {
		return nil, err
	}
	role.PermissionIDs = permissionIDs

	return &role, nil
}

func (r *PostgresRoleRepository) permissionIDsFor(ctx context.Context, roleID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT permission_id FROM role_permissions WHERE role_id = $1`, roleID)
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

func (r *PostgresRoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	const query = `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Note : comme pour User.List, on ne charge pas PermissionIDs pour
	// chaque role ici (evite les N+1 requetes). GetByID s'en charge pour
	// UN role precis.
	roles := []*domain.Role{}
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, rows.Err()
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	const query = `UPDATE roles SET name = $2, description = $3, updated_at = $4 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, role.ID, role.Name, role.Description, role.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

// GrantPermission insere dans la table de jonction ET met a jour
// roles.updated_at -- sans ca, updated_at resterait perime en base par
// rapport a la mutation deja faite en memoire par domain.Role.GrantPermission.
func (r *PostgresRoleRepository) GrantPermission(ctx context.Context, roleID, permissionID string) error {
	const insertQuery = `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, insertQuery, roleID, permissionID); err != nil {
		return err
	}
	return r.touch(ctx, roleID)
}

// RevokePermission -- meme remarque que GrantPermission au sujet de updated_at.
func (r *PostgresRoleRepository) RevokePermission(ctx context.Context, roleID, permissionID string) error {
	const deleteQuery = `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	if _, err := r.pool.Exec(ctx, deleteQuery, roleID, permissionID); err != nil {
		return err
	}
	return r.touch(ctx, roleID)
}

// touch met a jour updated_at a l'instant present. N'echoue pas si le
// role n'existe plus (RowsAffected == 0 ignore) -- GrantPermission/
// RevokePermission ont deja reussi leur propre operation a ce stade, pas
// la peine de faire echouer toute l'operation pour cette seule mise a jour
// de metadonnee.
func (r *PostgresRoleRepository) touch(ctx context.Context, roleID string) error {
	const query = `UPDATE roles SET updated_at = now() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, roleID)
	return err
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM roles WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}
