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

func TestServiceHandler_Get(t *testing.T) {
	t.Run("200 pour un service existant", func(t *testing.T) {
		repo := newFakeServiceRepo()
		h := NewServiceHandler(repo)
		ctx := context.Background()

		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")
		_ = repo.Create(ctx, svc)

		req := httptest.NewRequest(http.MethodGet, "/services/"+svc.ID, nil)
		req.SetPathValue("id", svc.ID)
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp serviceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if resp.Name != "orders-service" {
			t.Errorf("Name = %q, attendu %q", resp.Name, "orders-service")
		}
	})

	t.Run("404 pour un service inexistant", func(t *testing.T) {
		h := NewServiceHandler(newFakeServiceRepo())

		req := httptest.NewRequest(http.MethodGet, "/services/inconnu", nil)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestServiceHandler_List(t *testing.T) {
	repo := newFakeServiceRepo()
	h := NewServiceHandler(repo)
	ctx := context.Background()

	s1, _, _ := domain.NewService(domain.NewID(), "orders-service", "")
	s2, _, _ := domain.NewService(domain.NewID(), "billing-service", "")
	_ = repo.Create(ctx, s1)
	_ = repo.Create(ctx, s2)

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusOK)
	}

	var resp []serviceResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("nombre de services = %d, attendu 2", len(resp))
	}
}

func TestServiceHandler_Delete(t *testing.T) {
	t.Run("204 pour une suppression reussie", func(t *testing.T) {
		repo := newFakeServiceRepo()
		h := NewServiceHandler(repo)
		ctx := context.Background()

		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")
		_ = repo.Create(ctx, svc)

		req := httptest.NewRequest(http.MethodDelete, "/services/"+svc.ID, nil)
		req.SetPathValue("id", svc.ID)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("404 pour un service inexistant", func(t *testing.T) {
		h := NewServiceHandler(newFakeServiceRepo())

		req := httptest.NewRequest(http.MethodDelete, "/services/inconnu", nil)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestServiceHandler_Update(t *testing.T) {
	repo := newFakeServiceRepo()
	h := NewServiceHandler(repo)
	ctx := context.Background()

	svc, plainKey, _ := domain.NewService(domain.NewID(), "orders-service", "ancienne description")
	_ = repo.Create(ctx, svc)

	t.Run("200 pour une mise a jour valide, sans affecter la cle API", func(t *testing.T) {
		body := strings.NewReader(`{"name":"orders-service-v2","description":"nouvelle description"}`)
		req := httptest.NewRequest(http.MethodPut, "/services/"+svc.ID, body)
		req.SetPathValue("id", svc.ID)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp serviceResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Name != "orders-service-v2" {
			t.Errorf("Name = %q, attendu %q", resp.Name, "orders-service-v2")
		}

		if !svc.VerifyAPIKey(plainKey) {
			t.Error("la cle API existante devrait rester valide apres Update")
		}
	})

	t.Run("404 pour un service inexistant", func(t *testing.T) {
		body := strings.NewReader(`{"name":"x","description":""}`)
		req := httptest.NewRequest(http.MethodPut, "/services/inconnu", body)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestServiceHandler_RotateAPIKey(t *testing.T) {
	repo := newFakeServiceRepo()
	h := NewServiceHandler(repo)
	ctx := context.Background()

	svc, oldKey, _ := domain.NewService(domain.NewID(), "orders-service", "")
	_ = repo.Create(ctx, svc)

	req := httptest.NewRequest(http.MethodPost, "/services/"+svc.ID+"/rotate-key", nil)
	req.SetPathValue("id", svc.ID)
	rec := httptest.NewRecorder()

	h.RotateAPIKey(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp createServiceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decodage: %v", err)
	}
	if resp.APIKey == "" {
		t.Error("la nouvelle cle en clair ne devrait pas etre vide")
	}
	if resp.APIKey == oldKey {
		t.Error("la nouvelle cle ne devrait pas etre identique a l'ancienne")
	}

	_, err := repo.GetByAPIKeyHash(ctx, domain.HashToken(oldKey))
	if err == nil {
		t.Error("l'ancienne cle ne devrait plus permettre de retrouver le service")
	}
}
