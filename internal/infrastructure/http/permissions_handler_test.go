package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"user-service/internal/application"
	"user-service/internal/domain"
)

func setupPermissionsHandlerTest(t *testing.T) (h *PermissionsHandler, user *domain.User, svc *domain.Service, grantedKey string) {
	t.Helper()

	users := newFakeUserRepo()
	roles := newFakeRoleRepo()
	groups := newFakeGroupRepo()
	permissions := newFakePermissionRepo()
	ctx := context.Background()

	svc, _, _ = domain.NewService(domain.NewID(), "orders-service", "")

	perm, err := domain.NewPermission(domain.NewID(), svc.ID, "orders:read", "")
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}
	_ = permissions.Create(ctx, perm)

	user = mustNewUser(t, "Alice", "alice@example.com")
	user.AssignPermission(perm.ID)
	_ = users.Create(ctx, user)

	authorization := application.NewAuthorizationService(users, roles, groups, permissions)
	return NewPermissionsHandler(authorization), user, svc, perm.Key
}

func TestPermissionsHandler_ListForUser(t *testing.T) {
	h, user, svc, grantedKey := setupPermissionsHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions", nil)
	req.SetPathValue("id", user.ID)
	req = req.WithContext(withTestService(req.Context(), svc))
	rec := httptest.NewRecorder()

	h.ListForUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp permissionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decodage: %v", err)
	}
	if len(resp.Permissions) != 1 || resp.Permissions[0] != grantedKey {
		t.Errorf("Permissions = %v, want [%q]", resp.Permissions, grantedKey)
	}
}

func TestPermissionsHandler_CheckForUser(t *testing.T) {
	t.Run("allowed=true for a granted permission", func(t *testing.T) {
		h, user, svc, grantedKey := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp checkPermissionResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if !resp.Allowed {
			t.Error("Allowed should be true for a granted permission")
		}
		if resp.Permission != grantedKey {
			t.Errorf("Permission = %q, want %q", resp.Permission, grantedKey)
		}
	})

	t.Run("allowed=false for an ungranted permission", func(t *testing.T) {
		h, user, svc, _ := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission=orders:delete", nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp checkPermissionResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Allowed {
			t.Error("Allowed should be false for an ungranted permission")
		}
	})

	t.Run("400 when the permission parameter is missing", func(t *testing.T) {
		h, user, svc, _ := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check", nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("401 without an authenticated service in the context", func(t *testing.T) {
		h, user, _, grantedKey := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", user.ID)
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("404 for a nonexistent user", func(t *testing.T) {
		h, _, svc, grantedKey := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/unknown/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", "unknown")
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("a third-party service only sees its own permissions", func(t *testing.T) {
		h, user, _, grantedKey := setupPermissionsHandlerTest(t)
		autreService, _, _ := domain.NewService(domain.NewID(), "billing-service", "")

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), autreService))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		var resp checkPermissionResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Allowed {
			t.Error("another service should never see a permission granted under a different service")
		}
	})
}
