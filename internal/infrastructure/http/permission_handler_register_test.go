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
	t.Run("201 pour un premier enregistrement", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		body := strings.NewReader(`{"key":"orders:read","description":"lire les commandes"}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.RegisterSelf(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var resp permissionResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if resp.ServiceID != svc.ID {
			t.Errorf("ServiceID = %q, attendu %q (celui du contexte, pas fourni par le client)", resp.ServiceID, svc.ID)
		}
		if resp.Key != "orders:read" {
			t.Errorf("Key = %q, attendu %q", resp.Key, "orders:read")
		}
	})

	t.Run("200 idempotent pour un enregistrement duplique", func(t *testing.T) {
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

		secondBody := strings.NewReader(`{"key":"orders:read","description":"description differente cette fois"}`)
		secondReq := httptest.NewRequest(http.MethodPost, "/permissions/self", secondBody)
		secondReq = secondReq.WithContext(withTestService(secondReq.Context(), svc))
		secondRec := httptest.NewRecorder()
		h.RegisterSelf(secondRec, secondReq)

		if secondRec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d (idempotent -> 200, pas une nouvelle creation)", secondRec.Code, http.StatusOK)
		}

		var second permissionResponse
		_ = json.NewDecoder(secondRec.Body).Decode(&second)
		if second.ID != first.ID {
			t.Error("le deuxieme enregistrement devrait retourner la MEME permission (meme ID), pas en creer une nouvelle")
		}
	})

	t.Run("deux services distincts peuvent avoir la meme cle", func(t *testing.T) {
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
				t.Fatalf("status = %d, attendu %d pour le service %q", rec.Code, http.StatusCreated, svc.Name)
			}
		}
	})

	t.Run("401 sans service authentifie dans le contexte", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		body := strings.NewReader(`{"key":"orders:read","description":""}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
		rec := httptest.NewRecorder()

		h.RegisterSelf(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("400 pour une cle vide", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		body := strings.NewReader(`{"key":"","description":""}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions/self", body)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.RegisterSelf(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusBadRequest)
		}
	})
}
