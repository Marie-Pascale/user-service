package application

import (
	"context"
	"testing"

	"user-service/internal/domain"
)

type authFixture struct {
	users       *fakeUserRepo
	roles       *fakeRoleRepo
	groups      *fakeGroupRepo
	permissions *fakePermissionRepo
	service     *AuthorizationService

	serviceID string
	user      *domain.User
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	ctx := context.Background()

	f := &authFixture{
		users:       newFakeUserRepo(),
		roles:       newFakeRoleRepo(),
		groups:      newFakeGroupRepo(),
		permissions: newFakePermissionRepo(),
		serviceID:   "service-orders",
	}
	f.service = NewAuthorizationService(f.users, f.roles, f.groups, f.permissions)

	u, err := domain.NewUser(domain.NewID(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("construction de l'usager: %v", err)
	}
	if err := f.users.Create(ctx, u); err != nil {
		t.Fatalf("persistance de l'usager: %v", err)
	}
	f.user = u

	return f
}

func (f *authFixture) addPermission(t *testing.T, key string) *domain.Permission {
	t.Helper()
	p, err := domain.NewPermission(domain.NewID(), f.serviceID, key, "")
	if err != nil {
		t.Fatalf("construction de la permission: %v", err)
	}
	if err := f.permissions.Create(context.Background(), p); err != nil {
		t.Fatalf("persistance de la permission: %v", err)
	}
	return p
}

func TestAuthorizationService_DirectPermission(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()

	p := f.addPermission(t, "orders:read")
	f.user.AssignPermission(p.ID)

	keys, err := f.service.ResolvePermissionKeys(ctx, f.user.ID, f.serviceID)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(keys) != 1 || keys[0] != "orders:read" {
		t.Errorf("keys = %v, attendu [\"orders:read\"]", keys)
	}
}

func TestAuthorizationService_PermissionViaDirectRole(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()

	p := f.addPermission(t, "orders:write")
	role, _ := domain.NewRole(domain.NewID(), "editeur", "")
	role.GrantPermission(p.ID)
	_ = f.roles.Create(ctx, role)

	f.user.AssignRole(role.ID)

	keys, err := f.service.ResolvePermissionKeys(ctx, f.user.ID, f.serviceID)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(keys) != 1 || keys[0] != "orders:write" {
		t.Errorf("keys = %v, attendu [\"orders:write\"]", keys)
	}
}

func TestAuthorizationService_PermissionViaGroupRole(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()

	p := f.addPermission(t, "orders:delete")
	role, _ := domain.NewRole(domain.NewID(), "administrateur", "")
	role.GrantPermission(p.ID)
	_ = f.roles.Create(ctx, role)

	group, _ := domain.NewGroup(domain.NewID(), "equipe-ops", "")
	group.AssignRole(role.ID)
	group.AddMember(f.user.ID)
	_ = f.groups.Create(ctx, group)

	keys, err := f.service.ResolvePermissionKeys(ctx, f.user.ID, f.serviceID)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(keys) != 1 || keys[0] != "orders:delete" {
		t.Errorf("keys = %v, attendu [\"orders:delete\"] (via le groupe)", keys)
	}
}

func TestAuthorizationService_FiltreParService(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()

	own := f.addPermission(t, "orders:read")
	f.user.AssignPermission(own.ID)

	other, _ := domain.NewPermission(domain.NewID(), "service-billing", "invoices:read", "")
	_ = f.permissions.Create(ctx, other)
	f.user.AssignPermission(other.ID)

	keys, err := f.service.ResolvePermissionKeys(ctx, f.user.ID, f.serviceID)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(keys) != 1 || keys[0] != "orders:read" {
		t.Errorf("keys = %v, la permission d'un autre service ne devrait pas apparaitre", keys)
	}
}

func TestAuthorizationService_AucunePermission(t *testing.T) {
	f := newAuthFixture(t)

	keys, err := f.service.ResolvePermissionKeys(context.Background(), f.user.ID, f.serviceID)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("keys = %v, attendu aucune permission", keys)
	}
}

func TestAuthorizationService_HasPermission(t *testing.T) {
	f := newAuthFixture(t)
	ctx := context.Background()

	p := f.addPermission(t, "orders:read")
	f.user.AssignPermission(p.ID)

	has, err := f.service.HasPermission(ctx, f.user.ID, f.serviceID, "orders:read")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if !has {
		t.Error("HasPermission devrait retourner true pour une permission accordee")
	}

	has, err = f.service.HasPermission(ctx, f.user.ID, f.serviceID, "orders:delete")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if has {
		t.Error("HasPermission devrait retourner false pour une permission non accordee")
	}
}
