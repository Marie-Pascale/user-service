//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// newTestRepo ouvre une connexion vers la vraie base Postgres (celle du
// conteneur docker-compose, via les variables d'environnement) et retourne
// un repository pret a l'emploi. Chaque test nettoie ses propres donnees
// via t.Cleanup -- pas de base partagee "sale" entre les tests.
func newTestRepo(t *testing.T) *PostgresUserRepository {
	t.Helper()

	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion a la base de test: %v", err)
	}
	t.Cleanup(pool.Close)

	return NewPostgresUserRepository(pool)
}

func TestPostgresUserRepository_CreateAndGetByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, err := domain.NewUser(domain.NewID(), "Test Integration", "integration-create@example.com")
	if err != nil {
		t.Fatalf("construction de l'usager: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != u.Email {
		t.Errorf("Email = %q, attendu %q", got.Email, u.Email)
	}
	if got.RoleIDs == nil || got.PermissionIDs == nil {
		t.Error("RoleIDs et PermissionIDs devraient etre des slices vides, pas nil")
	}
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.GetByID(context.Background(), domain.NewID())
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("erreur = %v, attendu %v", err, application.ErrNotFound)
	}
}

func TestPostgresUserRepository_GetByEmail(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, _ := domain.NewUser(domain.NewID(), "Test Email", "integration-email@example.com")
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByEmail(ctx, "integration-email@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("ID = %q, attendu %q", got.ID, u.ID)
	}
}

func TestPostgresUserRepository_Update(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, _ := domain.NewUser(domain.NewID(), "Avant", "integration-update@example.com")
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	u.Name = "Apres"
	u.Suspend()
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Apres" {
		t.Errorf("Name = %q, attendu %q", got.Name, "Apres")
	}
	if got.Status != domain.StatusSuspended {
		t.Errorf("Status = %q, attendu %q", got.Status, domain.StatusSuspended)
	}
}

func TestPostgresUserRepository_Delete(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, _ := domain.NewUser(domain.NewID(), "A supprimer", "integration-delete@example.com")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, u.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, u.ID)
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("erreur apres suppression = %v, attendu %v", err, application.ErrNotFound)
	}
}

func TestPostgresUserRepository_Delete_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	err := repo.Delete(context.Background(), domain.NewID())
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("erreur = %v, attendu %v", err, application.ErrNotFound)
	}
}

func TestPostgresUserRepository_AssignAndRevokeRole(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	roleRepo := NewPostgresRoleRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "Role Test", "assign-role-test@example.com")
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("Create usager: %v", err)
	}
	t.Cleanup(func() { _ = userRepo.Delete(ctx, u.ID) })

	role, _ := domain.NewRole(domain.NewID(), "role-assign-test", "")
	if err := roleRepo.Create(ctx, role); err != nil {
		t.Fatalf("Create role: %v", err)
	}
	t.Cleanup(func() { _ = roleRepo.Delete(ctx, role.ID) })

	if err := userRepo.AssignRole(ctx, u.ID, role.ID); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	got, err := userRepo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.RoleIDs) != 1 || got.RoleIDs[0] != role.ID {
		t.Errorf("RoleIDs = %v, attendu [%q]", got.RoleIDs, role.ID)
	}

	if err := userRepo.RevokeRole(ctx, u.ID, role.ID); err != nil {
		t.Fatalf("RevokeRole: %v", err)
	}

	got, err = userRepo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID apres revoke: %v", err)
	}
	if len(got.RoleIDs) != 0 {
		t.Errorf("RoleIDs = %v, attendu vide apres RevokeRole", got.RoleIDs)
	}
}
