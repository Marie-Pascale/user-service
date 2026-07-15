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

	if err := cleanTestData(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "nettoyage de la base de test: %v\n", err)
		pool.Close()
		os.Exit(1)
	}
	pool.Close()

	os.Exit(m.Run())
}

// cleanTestData supprime toutes les donnees susceptibles d'avoir ete
// laissees par un run precedent -- SAUF les lignes semees par les
// migrations (000013_seed_self_service et suivantes : le self-service,
// le role admin, leurs permissions). Un simple TRUNCATE les aurait
// egalement effacees, ce qui aurait fait echouer
// TestAdminRole_HasAllApplicationPermissions pour la mauvaise raison --
// pas un vrai bug, mais un nettoyage trop agressif qui detruit ce qu'il
// est cense proteger.
//
// users et groups n'ont aucune donnee semee : suppression totale, sans
// condition. roles/services/permissions excluent explicitement les ID
// fixes des migrations (voir application.SelfServiceID, AdminRoleID).
// Les tables de jonction (sessions, auth_methods, role_permissions,
// group_members, group_roles, user_roles, user_permissions) n'ont pas
// besoin d'etre listees : les contraintes ON DELETE CASCADE deja
// definies dans les migrations s'en chargent automatiquement.
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
