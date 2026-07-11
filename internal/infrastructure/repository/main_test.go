//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestMain s'execute UNE FOIS avant tous les tests d'integration de ce
// package (c'est un mecanisme du package "testing" standard, pas propre
// a ce projet). On l'utilise ici pour garantir une base de donnees
// propre au debut de CHAQUE execution de "make test-integration" --
// peu importe si le run precedent a echoue avant d'atteindre ses
// t.Cleanup (ex: panique, Ctrl+C, erreur de compilation dans un fichier
// ajoute apres coup).
//
// Alternative envisagee : une base de donnees de test separee, recreee
// a chaque run. Plus isolee, mais plus lourde a maintenir (nouveau
// service docker-compose, nouvelle cible Makefile). Vider les tables
// d'une base partagee est suffisant pour ce projet.
func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, err := NewPool(ctx, ConfigFromEnv())
	if err != nil {
		fmt.Fprintf(os.Stderr, "connexion a la base de test: %v\n", err)
		os.Exit(1)
	}

	if err := truncateAllTables(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "nettoyage de la base de test: %v\n", err)
		pool.Close()
		os.Exit(1)
	}
	pool.Close()

	os.Exit(m.Run())
}

// truncateAllTables vide toutes les tables susceptibles de contenir des
// donnees de test. CASCADE se charge automatiquement des tables de
// jonction (auth_methods, role_permissions, group_members, group_roles,
// user_roles, user_permissions, sessions) qui referencent celles-ci --
// pas besoin de les lister une par une ni de respecter un ordre precis.
func truncateAllTables(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `TRUNCATE TABLE users, services, roles, groups, permissions CASCADE`
	_, err := pool.Exec(ctx, query)
	return err
}
