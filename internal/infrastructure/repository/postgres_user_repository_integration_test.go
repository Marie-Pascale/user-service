//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/application"
	"user-service/internal/domain"
)

func newTestRepo(t *testing.T) *PostgresUserRepository {
	t.Helper()

	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	return NewPostgresUserRepository(pool)
}

func TestPostgresUserRepository_CreateAndGetByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, err := domain.NewUser(domain.NewID(), "Test Integration", "integration-create@example.com")
	if err != nil {
		t.Fatalf("create user: %v", err)
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
		t.Errorf("Email = %q, want %q", got.Email, u.Email)
	}
	if got.RoleIDs == nil || got.PermissionIDs == nil {
		t.Error("RoleIDs and PermissionIDs should be empty slices, not nil")
	}
}

func TestPostgresUserRepository_GetByID_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	_, err := repo.GetByID(context.Background(), domain.NewID())
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, application.ErrNotFound)
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
		t.Errorf("ID = %q, want %q", got.ID, u.ID)
	}
}

func TestPostgresUserRepository_Update(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, _ := domain.NewUser(domain.NewID(), "Before", "integration-update@example.com")
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	u.Name = "After"
	u.Suspend()
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "After" {
		t.Errorf("Name = %q, want %q", got.Name, "After")
	}
	if got.Status != domain.StatusSuspended {
		t.Errorf("Status = %q, want %q", got.Status, domain.StatusSuspended)
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
		t.Errorf("error after deletion = %v, want %v", err, application.ErrNotFound)
	}
}

func TestPostgresUserRepository_Delete_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	err := repo.Delete(context.Background(), domain.NewID())
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, application.ErrNotFound)
	}
}

func TestPostgresUserRepository_AssignAndRevokeRole(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	roleRepo := NewPostgresRoleRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "Role Test", "assign-role-test@example.com")
	if err := userRepo.Create(ctx, u); err != nil {
		t.Fatalf("Create user: %v", err)
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
		t.Errorf("RoleIDs = %v, want [%q]", got.RoleIDs, role.ID)
	}

	if err := userRepo.RevokeRole(ctx, u.ID, role.ID); err != nil {
		t.Fatalf("RevokeRole: %v", err)
	}

	got, err = userRepo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID after revoke: %v", err)
	}
	if len(got.RoleIDs) != 0 {
		t.Errorf("RoleIDs = %v, want empty after RevokeRole", got.RoleIDs)
	}
}

func TestPostgresUserRepository_RecordLogin(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u, _ := domain.NewUser(domain.NewID(), "Login Test", "record-login-test@example.com")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })

	before, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID (before): %v", err)
	}
	if before.LastLoginAt != nil {
		t.Error("LastLoginAt should be nil before any call to RecordLogin")
	}

	if err := repo.RecordLogin(ctx, u.ID); err != nil {
		t.Fatalf("RecordLogin: %v", err)
	}

	after, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID (after): %v", err)
	}
	if after.LastLoginAt == nil {
		t.Fatal("LastLoginAt should no longer be nil after RecordLogin")
	}
	if !after.UpdatedAt.After(before.UpdatedAt) {
		t.Error("UpdatedAt should also have advanced in the database after RecordLogin")
	}
}

func TestPostgresUserRepository_RecordLogin_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	err := repo.RecordLogin(context.Background(), domain.NewID())
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("error = %v, want %v", err, application.ErrNotFound)
	}
}
