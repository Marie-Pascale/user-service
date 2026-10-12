package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"user-service/internal/application"
	"user-service/internal/domain"
)

type PostgresGroupRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresGroupRepository(pool *pgxpool.Pool) *PostgresGroupRepository {
	return &PostgresGroupRepository{pool: pool}
}

var _ application.GroupRepository = (*PostgresGroupRepository)(nil)

func (r *PostgresGroupRepository) Create(ctx context.Context, g *domain.Group) error {
	const query = `
		INSERT INTO groups (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, g.ID, g.Name, g.Description, g.CreatedAt, g.UpdatedAt)
	return err
}

func (r *PostgresGroupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	const query = `SELECT id, name, description, created_at, updated_at FROM groups WHERE id = $1`

	var g domain.Group
	err := r.pool.QueryRow(ctx, query, id).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	if err := r.hydrate(ctx, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *PostgresGroupRepository) hydrate(ctx context.Context, g *domain.Group) error {
	members, err := r.idsFor(ctx, "group_members", "user_id", g.ID)
	if err != nil {
		return err
	}
	g.MemberUserIDs = members

	roles, err := r.idsFor(ctx, "group_roles", "role_id", g.ID)
	if err != nil {
		return err
	}
	g.RoleIDs = roles

	return nil
}

func (r *PostgresGroupRepository) ListByMemberUserID(ctx context.Context, userID string) ([]*domain.Group, error) {
	const query = `
		SELECT g.id, g.name, g.description, g.created_at, g.updated_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	groups := []*domain.Group{}
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		groups = append(groups, &g)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, g := range groups {
		if err := r.hydrate(ctx, g); err != nil {
			return nil, err
		}
	}
	return groups, nil
}

func (r *PostgresGroupRepository) idsFor(ctx context.Context, table, column, groupID string) ([]string, error) {
	query := `SELECT ` + column + ` FROM ` + table + ` WHERE group_id = $1`

	rows, err := r.pool.Query(ctx, query, groupID)
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

func (r *PostgresGroupRepository) List(ctx context.Context) ([]*domain.Group, error) {
	const query = `SELECT id, name, description, created_at, updated_at FROM groups ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*domain.Group{}
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, rows.Err()
}

func (r *PostgresGroupRepository) Update(ctx context.Context, g *domain.Group) error {
	const query = `UPDATE groups SET name = $2, description = $3, updated_at = $4 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, g.ID, g.Name, g.Description, g.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}

func (r *PostgresGroupRepository) AddMember(ctx context.Context, groupID, userID string) error {
	const insertQuery = `
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, insertQuery, groupID, userID); err != nil {
		return err
	}
	return r.touch(ctx, groupID)
}

func (r *PostgresGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	const deleteQuery = `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`
	if _, err := r.pool.Exec(ctx, deleteQuery, groupID, userID); err != nil {
		return err
	}
	return r.touch(ctx, groupID)
}

func (r *PostgresGroupRepository) AssignRole(ctx context.Context, groupID, roleID string) error {
	const insertQuery = `
		INSERT INTO group_roles (group_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, insertQuery, groupID, roleID); err != nil {
		return err
	}
	return r.touch(ctx, groupID)
}

func (r *PostgresGroupRepository) touch(ctx context.Context, groupID string) error {
	const query = `UPDATE groups SET updated_at = now() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, groupID)
	return err
}

func (r *PostgresGroupRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM groups WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return application.ErrNotFound
	}
	return nil
}
