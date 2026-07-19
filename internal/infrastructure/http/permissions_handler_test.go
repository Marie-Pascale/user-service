package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

// setupPermissionsHandlerTest construit un PermissionsHandler avec des
// fakes, un usager ayant une permission directe pour un service donne --
// suffisant pour tester ListForUser et CheckForUser sans base de donnees.
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
		t.Fatalf("construction permission: %v", err)
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
		t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp permissionsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decodage: %v", err)
	}
	if len(resp.Permissions) != 1 || resp.Permissions[0] != grantedKey {
		t.Errorf("Permissions = %v, attendu [%q]", resp.Permissions, grantedKey)
	}
}

func TestPermissionsHandler_CheckForUser(t *testing.T) {
	t.Run("allowed=true pour une permission accordee", func(t *testing.T) {
		h, user, svc, grantedKey := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp checkPermissionResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if !resp.Allowed {
			t.Error("Allowed devrait etre true pour une permission accordee")
		}
		if resp.Permission != grantedKey {
			t.Errorf("Permission = %q, attendu %q", resp.Permission, grantedKey)
		}
	})

	t.Run("allowed=false pour une permission non accordee", func(t *testing.T) {
		h, user, svc, _ := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission=orders:delete", nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusOK)
		}

		var resp checkPermissionResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Allowed {
			t.Error("Allowed devrait etre false pour une permission non accordee")
		}
	})

	t.Run("400 si le parametre permission est absent", func(t *testing.T) {
		h, user, svc, _ := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check", nil)
		req.SetPathValue("id", user.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("401 sans service authentifie dans le contexte", func(t *testing.T) {
		h, user, _, grantedKey := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID+"/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", user.ID)
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("404 pour un usager inexistant", func(t *testing.T) {
		h, _, svc, grantedKey := setupPermissionsHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/users/inconnu/permissions/check?permission="+grantedKey, nil)
		req.SetPathValue("id", "inconnu")
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.CheckForUser(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("un service tiers ne voit que ses propres permissions", func(t *testing.T) {
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
			t.Error("un autre service ne devrait jamais voir une permission accordee sous un service different")
		}
	})
}
