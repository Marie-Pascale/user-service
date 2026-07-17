package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gestionnaire-usagers/internal/domain"
)

func TestGroupHandler_Create(t *testing.T) {
	h := NewGroupHandler(newFakeGroupRepo())

	body := strings.NewReader(`{"name":"equipe-support","description":""}`)
	req := httptest.NewRequest(http.MethodPost, "/groups", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestGroupHandler_AddAndRemoveMember(t *testing.T) {
	groups := newFakeGroupRepo()
	h := NewGroupHandler(groups)
	ctx := context.Background()

	g, _ := domain.NewGroup(domain.NewID(), "equipe-support", "")
	_ = groups.Create(ctx, g)

	t.Run("AddMember ajoute le membre", func(t *testing.T) {
		body := strings.NewReader(`{"user_id":"user-1"}`)
		req := httptest.NewRequest(http.MethodPost, "/groups/"+g.ID+"/members", body)
		req.SetPathValue("id", g.ID)
		rec := httptest.NewRecorder()

		h.AddMember(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp groupResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if len(resp.MemberUserIDs) != 1 || resp.MemberUserIDs[0] != "user-1" {
			t.Errorf("MemberUserIDs = %v, attendu [\"user-1\"]", resp.MemberUserIDs)
		}
	})

	t.Run("RemoveMember le retire", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/groups/"+g.ID+"/members/user-1", nil)
		req.SetPathValue("id", g.ID)
		req.SetPathValue("userId", "user-1")
		rec := httptest.NewRecorder()

		h.RemoveMember(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusNoContent)
		}

		got, _ := groups.GetByID(ctx, g.ID)
		if len(got.MemberUserIDs) != 0 {
			t.Errorf("MemberUserIDs = %v, attendu vide apres RemoveMember", got.MemberUserIDs)
		}
	})
}

func TestGroupHandler_AssignRole(t *testing.T) {
	groups := newFakeGroupRepo()
	h := NewGroupHandler(groups)
	ctx := context.Background()

	g, _ := domain.NewGroup(domain.NewID(), "equipe-support", "")
	_ = groups.Create(ctx, g)

	body := strings.NewReader(`{"role_id":"role-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/groups/"+g.ID+"/roles", body)
	req.SetPathValue("id", g.ID)
	rec := httptest.NewRecorder()

	h.AssignRole(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp groupResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.RoleIDs) != 1 || resp.RoleIDs[0] != "role-1" {
		t.Errorf("RoleIDs = %v, attendu [\"role-1\"]", resp.RoleIDs)
	}
}

func TestGroupHandler_Update(t *testing.T) {
	groups := newFakeGroupRepo()
	h := NewGroupHandler(groups)
	ctx := context.Background()

	g, _ := domain.NewGroup(domain.NewID(), "equipe-support", "ancienne description")
	g.AddMember("user-1")
	_ = groups.Create(ctx, g)

	t.Run("200 pour une mise a jour valide, sans affecter les membres", func(t *testing.T) {
		body := strings.NewReader(`{"name":"equipe-support-fr","description":"nouvelle description"}`)
		req := httptest.NewRequest(http.MethodPut, "/groups/"+g.ID, body)
		req.SetPathValue("id", g.ID)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp groupResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Name != "equipe-support-fr" {
			t.Errorf("Name = %q, attendu %q", resp.Name, "equipe-support-fr")
		}
		if len(resp.MemberUserIDs) != 1 {
			t.Error("MemberUserIDs ne devrait pas etre affecte par Update")
		}
	})

	t.Run("404 pour un groupe inexistant", func(t *testing.T) {
		body := strings.NewReader(`{"name":"x","description":""}`)
		req := httptest.NewRequest(http.MethodPut, "/groups/inconnu", body)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestGroupHandler_Delete(t *testing.T) {
	t.Run("404 pour un groupe inexistant", func(t *testing.T) {
		h := NewGroupHandler(newFakeGroupRepo())

		req := httptest.NewRequest(http.MethodDelete, "/groups/inconnu", nil)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}
