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

func withTestService(ctx context.Context, svc *domain.Service) context.Context {
	return context.WithValue(ctx, serviceContextKey, svc)
}

func TestPermissionHandler_RegisterSelf(t *testing.T) {
	t.Run("201 for the first registration", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		body := strings.NewReader(`{"key":"orders:read","description":"read orders"}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.RegisterSelf(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var resp permissionResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if resp.ServiceID != svc.ID {
			t.Errorf("ServiceID = %q, want %q (from the context, not provided by the client)", resp.ServiceID, svc.ID)
		}
		if resp.Key != "orders:read" {
			t.Errorf("Key = %q, want %q", resp.Key, "orders:read")
		}
	})

	t.Run("200 idempotent for a duplicate registration", func(t *testing.T) {
		repo := newFakePermissionRepo()
		h := NewPermissionHandler(repo)
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		firstBody := strings.NewReader(`{"key":"orders:read","description":""}`)
		firstReq := httptest.NewRequest(http.MethodPost, "/permissions/self", firstBody)
		firstReq = firstReq.WithContext(withTestService(firstReq.Context(), svc))
		firstRec := httptest.NewRecorder()
		h.RegisterSelf(firstRec, firstReq)

		var first permissionResponse
		_ = json.NewDecoder(firstRec.Body).Decode(&first)

		secondBody := strings.NewReader(`{"key":"orders:read","description":"different description this time"}`)
		secondReq := httptest.NewRequest(http.MethodPost, "/permissions/self", secondBody)
		secondReq = secondReq.WithContext(withTestService(secondReq.Context(), svc))
		secondRec := httptest.NewRecorder()
		h.RegisterSelf(secondRec, secondReq)

		if secondRec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (idempotent -> 200, not a new creation)", secondRec.Code, http.StatusOK)
		}

		var second permissionResponse
		_ = json.NewDecoder(secondRec.Body).Decode(&second)
		if second.ID != first.ID {
			t.Error("the second registration should return the SAME permission (same ID), not create a new one")
		}
	})

	t.Run("two distinct services can have the same key", func(t *testing.T) {
		repo := newFakePermissionRepo()
		h := NewPermissionHandler(repo)
		svcA, _, _ := domain.NewService(domain.NewID(), "orders-service", "")
		svcB, _, _ := domain.NewService(domain.NewID(), "billing-service", "")

		for _, svc := range []*domain.Service{svcA, svcB} {
			body := strings.NewReader(`{"key":"read","description":""}`)
			req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
			req = req.WithContext(withTestService(req.Context(), svc))
			rec := httptest.NewRecorder()
			h.RegisterSelf(rec, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d for service %q", rec.Code, http.StatusCreated, svc.Name)
			}
		}
	})

	t.Run("401 without an authenticated service in the context", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		body := strings.NewReader(`{"key":"orders:read","description":""}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
		rec := httptest.NewRecorder()

		h.RegisterSelf(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("400 for an empty key", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		body := strings.NewReader(`{"key":"","description":""}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.RegisterSelf(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
