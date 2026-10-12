package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

func setupAuthzTest(t *testing.T, permissionKey string) (authz *AuthorizationMiddleware, adminUser, plainUser *domain.User) {
	t.Helper()

	users := newFakeUserRepo()
	roles := newFakeRoleRepo()
	groups := newFakeGroupRepo()
	permissions := newFakePermissionRepo()

	admin := mustNewUser(t, "Admin", "admin@example.com")
	plain := mustNewUser(t, "Usager Ordinaire", "plain@example.com")
	_ = users.Create(context.Background(), admin)
	_ = users.Create(context.Background(), plain)

	perm, err := domain.NewPermission(domain.NewID(), application.SelfServiceID, permissionKey, "")
	if err != nil {
		t.Fatalf("construction permission: %v", err)
	}
	_ = permissions.Create(context.Background(), perm)

	role, _ := domain.NewRole(domain.NewID(), "role-test", "")
	_ = roles.Create(context.Background(), role)
	_ = roles.GrantPermission(context.Background(), role.ID, perm.ID)

	admin.AssignRole(role.ID)

	authorization := application.NewAuthorizationService(users, roles, groups, permissions)
	return NewAuthorizationMiddleware(authorization), admin, plain
}

func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
	authz, admin, plain := setupAuthzTest(t, application.PermServicesCreate)

	newHandler := func() http.Handler {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		return authz.RequirePermission(application.PermServicesCreate)(inner)
	}

	t.Run("autorise un usager avec la permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services", nil)
		req = req.WithContext(withTestUser(req.Context(), admin))
		rec := httptest.NewRecorder()

		newHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("refuse un usager sans la permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services", nil)
		req = req.WithContext(withTestUser(req.Context(), plain))
		rec := httptest.NewRecorder()

		newHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("refuse une requete non authentifiee", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services", nil)
		rec := httptest.NewRecorder()

		newHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func TestAuthorizationMiddleware_RequireSelfOrPermission(t *testing.T) {
	authz, admin, plain := setupAuthzTest(t, application.PermUsersReadAny)

	newHandler := func() http.Handler {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		return authz.RequireSelfOrPermission(application.PermUsersReadAny)(inner)
	}

	t.Run("autorise un usager a acceder a son propre compte", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+plain.ID, nil)
		req.SetPathValue("id", plain.ID)
		req = req.WithContext(withTestUser(req.Context(), plain))
		rec := httptest.NewRecorder()

		newHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("refuse un usager ordinaire sur le compte d'un autre", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+admin.ID, nil)
		req.SetPathValue("id", admin.ID)
		req = req.WithContext(withTestUser(req.Context(), plain))
		rec := httptest.NewRecorder()

		newHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("autorise un admin sur le compte d'un autre", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+plain.ID, nil)
		req.SetPathValue("id", plain.ID)
		req = req.WithContext(withTestUser(req.Context(), admin))
		rec := httptest.NewRecorder()

		newHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusOK)
		}
	})
}

func withTestUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}
