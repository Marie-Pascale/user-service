package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
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
		INSERT INTO groups (id, name, description)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, g.ID, g.Name, g.Description)
	return err
}

func (r *PostgresGroupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	const query = `SELECT id, name, description FROM groups WHERE id = $1`

	var g domain.Group
	err := r.pool.QueryRow(ctx, query, id).Scan(&g.ID, &g.Name, &g.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrNotFound
		}
		return nil, err
	}

	members, err := r.idsFor(ctx, "group_members", "user_id", id)
	if err != nil {
		return nil, err
	}
	g.MemberUserIDs = members

	roles, err := r.idsFor(ctx, "group_roles", "role_id", id)
	if err != nil {
		return nil, err
	}
	g.RoleIDs = roles

	return &g, nil
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
	const query = `SELECT id, name, description FROM groups ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*domain.Group{}
	for rows.Next() {
		var g domain.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.Description); err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, rows.Err()
}

func (r *PostgresGroupRepository) AddMember(ctx context.Context, groupID, userID string) error {
	const query = `
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, groupID, userID)
	return err
}

func (r *PostgresGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	const query = `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, groupID, userID)
	return err
}

func (r *PostgresGroupRepository) AssignRole(ctx context.Context, groupID, roleID string) error {
	const query = `
		INSERT INTO group_roles (group_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, groupID, roleID)
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
