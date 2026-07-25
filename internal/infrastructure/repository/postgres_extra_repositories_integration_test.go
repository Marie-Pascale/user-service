//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// --- Service ---------------------------------------------------------

func TestPostgresServiceRepository_CreateAndVerify(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	repo := NewPostgresServiceRepository(pool)

	svc, plainKey, err := domain.NewService(domain.NewID(), "orders-service", "gere les commandes")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, svc.ID) })

	if err := repo.Create(ctx, svc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByAPIKeyHash(ctx, svc.APIKeyHash)
	if err != nil {
		t.Fatalf("GetByAPIKeyHash: %v", err)
	}
	if !got.VerifyAPIKey(plainKey) {
		t.Error("la cle en clair devrait etre valide pour le service retrouve")
	}
}

func TestPostgresServiceRepository_UpdateAPIKeyHash(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	repo := NewPostgresServiceRepository(pool)

	svc, oldKey, err := domain.NewService(domain.NewID(), "rotate-service", "")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, svc.ID) })

	if err := repo.Create(ctx, svc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Reproduit exactement la sequence de ServiceHandler.RotateAPIKey :
	// le domaine mute svc.APIKeyHash EN MEMOIRE, puis on persiste.
	newKey, err := svc.RotateAPIKey()
	if err != nil {
		t.Fatalf("RotateAPIKey: %v", err)
	}
	if err := repo.UpdateAPIKeyHash(ctx, svc.ID, svc.APIKeyHash); err != nil {
		t.Fatalf("UpdateAPIKeyHash: %v", err)
	}

	if _, err := repo.GetByAPIKeyHash(ctx, domain.HashToken(oldKey)); !errors.Is(err, application.ErrNotFound) {
		t.Errorf("l'ancienne cle devrait etre invalide, erreur = %v, attendu %v", err, application.ErrNotFound)
	}

	got, err := repo.GetByAPIKeyHash(ctx, domain.HashToken(newKey))
	if err != nil {
		t.Fatalf("GetByAPIKeyHash avec la nouvelle cle: %v", err)
	}
	if got.ID != svc.ID {
		t.Errorf("ID = %q, attendu %q", got.ID, svc.ID)
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Error("UpdatedAt devrait avoir avance en base apres UpdateAPIKeyHash (pas seulement en memoire)")
	}
}

func TestPostgresServiceRepository_Update(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	repo := NewPostgresServiceRepository(pool)

	svc, plainKey, err := domain.NewService(domain.NewID(), "update-service", "avant")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, svc.ID) })
	if err := repo.Create(ctx, svc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.UpdateDetails("update-service-v2", "apres"); err != nil {
		t.Fatalf("UpdateDetails: %v", err)
	}
	if err := repo.Update(ctx, svc); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, svc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "update-service-v2" || got.Description != "apres" {
		t.Errorf("got = %+v, attendu Name=update-service-v2 Description=apres", got)
	}
	if !got.VerifyAPIKey(plainKey) {
		t.Error("la cle API existante devrait rester valide apres Update")
	}
}

// --- Permission --------------------------------------------------------

func TestPostgresPermissionRepository_CreateAndListByService(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	svcRepo := NewPostgresServiceRepository(pool)
	permRepo := NewPostgresPermissionRepository(pool)

	svc, _, _ := domain.NewService(domain.NewID(), "billing-service", "")
	if err := svcRepo.Create(ctx, svc); err != nil {
		t.Fatalf("creation du service: %v", err)
	}
	t.Cleanup(func() { _ = svcRepo.Delete(ctx, svc.ID) })

	p, err := domain.NewPermission(domain.NewID(), svc.ID, "invoices:read", "")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	if err := permRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = permRepo.Delete(ctx, p.ID) })

	perms, err := permRepo.ListByServiceID(ctx, svc.ID)
	if err != nil {
		t.Fatalf("ListByServiceID: %v", err)
	}
	if len(perms) != 1 || perms[0].Key != "invoices:read" {
		t.Errorf("perms = %+v, attendu une permission \"invoices:read\"", perms)
	}
}

func TestPostgresPermissionRepository_GetByServiceIDAndKey(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	svcRepo := NewPostgresServiceRepository(pool)
	permRepo := NewPostgresPermissionRepository(pool)

	svc, _, _ := domain.NewService(domain.NewID(), "self-register-service", "")
	if err := svcRepo.Create(ctx, svc); err != nil {
		t.Fatalf("creation du service: %v", err)
	}
	t.Cleanup(func() { _ = svcRepo.Delete(ctx, svc.ID) })

	p, _ := domain.NewPermission(domain.NewID(), svc.ID, "orders:read", "")
	if err := permRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = permRepo.Delete(ctx, p.ID) })

	t.Run("retrouve une permission existante", func(t *testing.T) {
		got, err := permRepo.GetByServiceIDAndKey(ctx, svc.ID, "orders:read")
		if err != nil {
			t.Fatalf("GetByServiceIDAndKey: %v", err)
		}
		if got.ID != p.ID {
			t.Errorf("ID = %q, attendu %q", got.ID, p.ID)
		}
	})

	t.Run("ErrNotFound pour une cle inexistante", func(t *testing.T) {
		_, err := permRepo.GetByServiceIDAndKey(ctx, svc.ID, "orders:write")
		if !errors.Is(err, application.ErrNotFound) {
			t.Errorf("erreur = %v, attendu %v", err, application.ErrNotFound)
		}
	})
}

// --- Role ----------------------------------------------------------------

func TestPostgresRoleRepository_GrantAndRevokePermission(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	svcRepo := NewPostgresServiceRepository(pool)
	permRepo := NewPostgresPermissionRepository(pool)
	roleRepo := NewPostgresRoleRepository(pool)

	svc, _, _ := domain.NewService(domain.NewID(), "support-service", "")
	_ = svcRepo.Create(ctx, svc)
	t.Cleanup(func() { _ = svcRepo.Delete(ctx, svc.ID) })

	p, _ := domain.NewPermission(domain.NewID(), svc.ID, "tickets:read", "")
	_ = permRepo.Create(ctx, p)
	t.Cleanup(func() { _ = permRepo.Delete(ctx, p.ID) })

	role, err := domain.NewRole(domain.NewID(), "support-agent", "")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	if err := roleRepo.Create(ctx, role); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = roleRepo.Delete(ctx, role.ID) })

	if err := roleRepo.GrantPermission(ctx, role.ID, p.ID); err != nil {
		t.Fatalf("GrantPermission: %v", err)
	}

	got, err := roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.PermissionIDs) != 1 || got.PermissionIDs[0] != p.ID {
		t.Errorf("PermissionIDs = %v, attendu [%q]", got.PermissionIDs, p.ID)
	}
	if !got.UpdatedAt.After(role.CreatedAt) {
		t.Error("UpdatedAt devrait avoir avance en base apres GrantPermission (pas seulement en memoire)")
	}

	if err := roleRepo.RevokePermission(ctx, role.ID, p.ID); err != nil {
		t.Fatalf("RevokePermission: %v", err)
	}

	got, err = roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetByID apres revoke: %v", err)
	}
	if len(got.PermissionIDs) != 0 {
		t.Errorf("PermissionIDs = %v, attendu vide apres RevokePermission", got.PermissionIDs)
	}
}

func TestPostgresRoleRepository_Update(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	roleRepo := NewPostgresRoleRepository(pool)

	role, err := domain.NewRole(domain.NewID(), "update-role", "avant")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	t.Cleanup(func() { _ = roleRepo.Delete(ctx, role.ID) })
	if err := roleRepo.Create(ctx, role); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := role.UpdateDetails("update-role-v2", "apres"); err != nil {
		t.Fatalf("UpdateDetails: %v", err)
	}
	if err := roleRepo.Update(ctx, role); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "update-role-v2" || got.Description != "apres" {
		t.Errorf("got = %+v, attendu Name=update-role-v2 Description=apres", got)
	}
}

// --- Group -----------------------------------------------------------

func TestPostgresGroupRepository_MembersAndRoles(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	roleRepo := NewPostgresRoleRepository(pool)
	groupRepo := NewPostgresGroupRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "Membre Test", "membre-groupe@example.com")
	_ = userRepo.Create(ctx, u)
	t.Cleanup(func() { _ = userRepo.Delete(ctx, u.ID) })

	role, _ := domain.NewRole(domain.NewID(), "role-groupe-test", "")
	_ = roleRepo.Create(ctx, role)
	t.Cleanup(func() { _ = roleRepo.Delete(ctx, role.ID) })

	g, err := domain.NewGroup(domain.NewID(), "equipe-test", "")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	if err := groupRepo.Create(ctx, g); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = groupRepo.Delete(ctx, g.ID) })

	if err := groupRepo.AddMember(ctx, g.ID, u.ID); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if err := groupRepo.AssignRole(ctx, g.ID, role.ID); err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	got, err := groupRepo.GetByID(ctx, g.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.MemberUserIDs) != 1 || got.MemberUserIDs[0] != u.ID {
		t.Errorf("MemberUserIDs = %v, attendu [%q]", got.MemberUserIDs, u.ID)
	}
	if len(got.RoleIDs) != 1 || got.RoleIDs[0] != role.ID {
		t.Errorf("RoleIDs = %v, attendu [%q]", got.RoleIDs, role.ID)
	}
	if !got.UpdatedAt.After(g.CreatedAt) {
		t.Error("UpdatedAt devrait avoir avance en base apres AddMember/AssignRole (pas seulement en memoire)")
	}

	viaMembership, err := groupRepo.ListByMemberUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ListByMemberUserID: %v", err)
	}
	if len(viaMembership) != 1 || viaMembership[0].ID != g.ID {
		t.Errorf("ListByMemberUserID = %+v, attendu un groupe (%q)", viaMembership, g.ID)
	}
	if len(viaMembership[0].RoleIDs) != 1 || viaMembership[0].RoleIDs[0] != role.ID {
		t.Errorf("RoleIDs via ListByMemberUserID = %v, attendu [%q]", viaMembership[0].RoleIDs, role.ID)
	}
}

func TestPostgresGroupRepository_Update(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	groupRepo := NewPostgresGroupRepository(pool)

	g, err := domain.NewGroup(domain.NewID(), "update-group", "avant")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	t.Cleanup(func() { _ = groupRepo.Delete(ctx, g.ID) })
	if err := groupRepo.Create(ctx, g); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := g.UpdateDetails("update-group-v2", "apres"); err != nil {
		t.Fatalf("UpdateDetails: %v", err)
	}
	if err := groupRepo.Update(ctx, g); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := groupRepo.GetByID(ctx, g.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "update-group-v2" || got.Description != "apres" {
		t.Errorf("got = %+v, attendu Name=update-group-v2 Description=apres", got)
	}
}

// --- AuthMethod --------------------------------------------------------

func TestPostgresAuthMethodRepository_LocalAndOAuth(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()

	userRepo := NewPostgresUserRepository(pool)
	authRepo := NewPostgresAuthMethodRepository(pool)

	u, _ := domain.NewUser(domain.NewID(), "Auth Test", "auth-method-test@example.com")
	_ = userRepo.Create(ctx, u)
	t.Cleanup(func() { _ = userRepo.Delete(ctx, u.ID) })

	local, err := domain.NewOAuthMethod(domain.NewID(), u.ID, domain.ProviderGoogle, "google-sub-xyz")
	if err != nil {
		t.Fatalf("construction: %v", err)
	}
	if err := authRepo.Create(ctx, local); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = authRepo.Delete(ctx, local.ID) })

	got, err := authRepo.GetByProviderUserID(ctx, domain.ProviderGoogle, "google-sub-xyz")
	if err != nil {
		t.Fatalf("GetByProviderUserID: %v", err)
	}
	if got.UserID != u.ID {
		t.Errorf("UserID = %q, attendu %q", got.UserID, u.ID)
	}

	methods, err := authRepo.ListByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(methods) != 1 {
		t.Errorf("methods = %v, attendu 1 methode", methods)
	}
}

func TestPostgresAuthMethodRepository_NotFound(t *testing.T) {
	pool, err := NewPool(context.Background(), ConfigFromEnv())
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	defer pool.Close()
	authRepo := NewPostgresAuthMethodRepository(pool)

	_, err = authRepo.GetByID(context.Background(), domain.NewID())
	if !errors.Is(err, application.ErrNotFound) {
		t.Errorf("erreur = %v, attendu %v", err, application.ErrNotFound)
	}
}
