package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"user-service/internal/domain"
)

func TestRoleHandler_Create(t *testing.T) {
	h := NewRoleHandler(newFakeRoleRepo())

	body := strings.NewReader(`{"name":"support-agent","description":""}`)
	req := httptest.NewRequest(http.MethodPost, "/roles", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var resp roleResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decodage: %v", err)
	}
	if resp.Name != "support-agent" {
		t.Errorf("Name = %q, attendu %q", resp.Name, "support-agent")
	}
	if resp.PermissionIDs == nil {
		t.Error("PermissionIDs devrait etre une slice vide, pas nil")
	}
}

func TestRoleHandler_GrantAndRevokePermission(t *testing.T) {
	roles := newFakeRoleRepo()
	h := NewRoleHandler(roles)
	ctx := context.Background()

	role, _ := domain.NewRole(domain.NewID(), "support-agent", "")
	_ = roles.Create(ctx, role)

	t.Run("GrantPermission ajoute la permission", func(t *testing.T) {
		body := strings.NewReader(`{"permission_id":"perm-1"}`)
		req := httptest.NewRequest(http.MethodPost, "/roles/"+role.ID+"/permissions", body)
		req.SetPathValue("id", role.ID)
		rec := httptest.NewRecorder()

		h.GrantPermission(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp roleResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if len(resp.PermissionIDs) != 1 || resp.PermissionIDs[0] != "perm-1" {
			t.Errorf("PermissionIDs = %v, attendu [\"perm-1\"]", resp.PermissionIDs)
		}
	})

	t.Run("RevokePermission la retire", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/roles/"+role.ID+"/permissions/perm-1", nil)
		req.SetPathValue("id", role.ID)
		req.SetPathValue("permissionId", "perm-1")
		rec := httptest.NewRecorder()

		h.RevokePermission(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusNoContent)
		}

		got, _ := roles.GetByID(ctx, role.ID)
		if len(got.PermissionIDs) != 0 {
			t.Errorf("PermissionIDs = %v, attendu vide apres revoke", got.PermissionIDs)
		}
	})
}

func TestRoleHandler_Update(t *testing.T) {
	roles := newFakeRoleRepo()
	h := NewRoleHandler(roles)
	ctx := context.Background()

	role, _ := domain.NewRole(domain.NewID(), "support-agent", "ancienne description")
	_ = roles.Create(ctx, role)

	t.Run("200 pour une mise a jour valide", func(t *testing.T) {
		body := strings.NewReader(`{"name":"support-agent-senior","description":"nouvelle description"}`)
		req := httptest.NewRequest(http.MethodPut, "/roles/"+role.ID, body)
		req.SetPathValue("id", role.ID)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp roleResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Name != "support-agent-senior" {
			t.Errorf("Name = %q, attendu %q", resp.Name, "support-agent-senior")
		}
	})

	t.Run("400 pour un nom vide", func(t *testing.T) {
		body := strings.NewReader(`{"name":"","description":""}`)
		req := httptest.NewRequest(http.MethodPut, "/roles/"+role.ID, body)
		req.SetPathValue("id", role.ID)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("404 pour un role inexistant", func(t *testing.T) {
		body := strings.NewReader(`{"name":"x","description":""}`)
		req := httptest.NewRequest(http.MethodPut, "/roles/inconnu", body)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestRoleHandler_List(t *testing.T) {
	roles := newFakeRoleRepo()
	h := NewRoleHandler(roles)
	ctx := context.Background()

	r1, _ := domain.NewRole(domain.NewID(), "role-a", "")
	r2, _ := domain.NewRole(domain.NewID(), "role-b", "")
	_ = roles.Create(ctx, r1)
	_ = roles.Create(ctx, r2)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusOK)
	}

	var resp []roleResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("nombre de roles = %d, attendu 2", len(resp))
	}
}
