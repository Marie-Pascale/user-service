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

func TestServiceHandler_Get(t *testing.T) {
	t.Run("200 for an existing service", func(t *testing.T) {
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
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp serviceResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if resp.Name != "orders-service" {
			t.Errorf("Name = %q, want %q", resp.Name, "orders-service")
		}
	})

	t.Run("404 for a nonexistent service", func(t *testing.T) {
		h := NewServiceHandler(newFakeServiceRepo())

		req := httptest.NewRequest(http.MethodGet, "/services/unknown", nil)
		req.SetPathValue("id", "unknown")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
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
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp []serviceResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("service count = %d, want 2", len(resp))
	}
}

func TestServiceHandler_Delete(t *testing.T) {
	t.Run("204 for a successful deletion", func(t *testing.T) {
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
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("404 for a nonexistent service", func(t *testing.T) {
		h := NewServiceHandler(newFakeServiceRepo())

		req := httptest.NewRequest(http.MethodDelete, "/services/unknown", nil)
		req.SetPathValue("id", "unknown")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestServiceHandler_Update(t *testing.T) {
	repo := newFakeServiceRepo()
	h := NewServiceHandler(repo)
	ctx := context.Background()

	svc, plainKey, _ := domain.NewService(domain.NewID(), "orders-service", "old description")
	_ = repo.Create(ctx, svc)

	t.Run("200 for a valid update without changing the API key", func(t *testing.T) {
		body := strings.NewReader(`{"name":"orders-service-v2","description":"new description"}`)
		req := httptest.NewRequest(http.MethodPut, "/services/"+svc.ID, body)
		req.SetPathValue("id", svc.ID)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var resp serviceResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Name != "orders-service-v2" {
			t.Errorf("Name = %q, want %q", resp.Name, "orders-service-v2")
		}

		if !svc.VerifyAPIKey(plainKey) {
			t.Error("the existing API key should remain valid after Update")
		}
	})

	t.Run("404 for a nonexistent service", func(t *testing.T) {
		body := strings.NewReader(`{"name":"x","description":""}`)
		req := httptest.NewRequest(http.MethodPut, "/services/unknown", body)
		req.SetPathValue("id", "unknown")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
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
		t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp createServiceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decodage: %v", err)
	}
	if resp.APIKey == "" {
		t.Error("the new plaintext key should not be empty")
	}
	if resp.APIKey == oldKey {
		t.Error("the new key should not be identical to the old key")
	}

	_, err := repo.GetByAPIKeyHash(ctx, domain.HashToken(oldKey))
	if err == nil {
		t.Error("the old key should no longer retrieve the service")
	}
}
