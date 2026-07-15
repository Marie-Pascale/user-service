//go:build integration

package repository

import (
	"context"
	"testing"

	"gestionnaire-usagers/internal/application"
)

// TestAdminRole_HasAllApplicationPermissions verifie que le role admin,
// tel que seme par les migrations (000013, 000014, 000015...), possede
// bien CHAQUE permission listee dans application.AllPermissionKeys.
//
// Sans ce test, un scenario plausible : quelqu'un ajoute une nouvelle
// route protegee par une nouvelle permission (ex: PermSomethingNew), mais
// oublie d'ecrire la migration qui l'accorde au role admin. Le code
// compile, les autres tests passent, et le bug ne se revele qu'en
// production quand un admin decouvre qu'il n'a soudain plus acces a une
// fonctionnalite -- exactement le genre de regression silencieuse que ce
// test est concu pour attraper immediatement.
func TestAdminRole_HasAllApplicationPermissions(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	roleRepo := NewPostgresRoleRepository(pool)
	permissionRepo := NewPostgresPermissionRepository(pool)

	adminRole, err := roleRepo.GetByID(ctx, application.AdminRoleID)
	if err != nil {
		t.Fatalf("GetByID(AdminRoleID): %v -- la migration 000013 a-t-elle bien ete appliquee ?", err)
	}

	grantedPermissions, err := permissionRepo.ListByIDs(ctx, adminRole.PermissionIDs)
	if err != nil {
		t.Fatalf("ListByIDs: %v", err)
	}

	grantedKeys := make(map[string]bool, len(grantedPermissions))
	for _, p := range grantedPermissions {
		grantedKeys[p.Key] = true
	}

	for _, key := range application.AllPermissionKeys {
		if !grantedKeys[key] {
			t.Errorf(
				"le role admin n'a PAS la permission %q -- verifie qu'une migration l'accorde bien via role_permissions (voir 000013/000014/000015 pour des exemples)",
				key,
			)
		}
	}
}
