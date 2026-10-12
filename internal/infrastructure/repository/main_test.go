//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, err := NewPool(ctx, ConfigFromEnv())
	if err != nil {
		fmt.Fprintf(os.Stderr, "connexion a la base de test: %v\n", err)
		os.Exit(1)
	}

	if err := cleanTestData(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "nettoyage de la base de test: %v\n", err)
		pool.Close()
		os.Exit(1)
	}
	pool.Close()

	os.Exit(m.Run())
}

func cleanTestData(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`DELETE FROM users`,
		`DELETE FROM groups`,
		`DELETE FROM roles WHERE id <> '00000000-0000-0000-0000-000000000002'`,
		`DELETE FROM services WHERE id <> '00000000-0000-0000-0000-000000000001'`,
		`DELETE FROM permissions WHERE service_id <> '00000000-0000-0000-0000-000000000001'`,
	}
	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
