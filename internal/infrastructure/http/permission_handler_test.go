package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPermissionHandler_Create(t *testing.T) {
	t.Run("201 pour une requete valide", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		body := strings.NewReader(`{"service_id":"svc-1","key":"orders:read","description":""}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions", body)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})

	t.Run("400 pour une cle vide", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		req := httptest.NewRequest(http.MethodPost, "/permissions", strings.NewReader(`{"service_id":"svc-1","key":""}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPermissionHandler_ListByService(t *testing.T) {
	t.Run("400 si service_id absent", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
		rec := httptest.NewRecorder()

		h.ListByService(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("200 avec service_id fourni", func(t *testing.T) {
		repo := newFakePermissionRepo()
		h := NewPermissionHandler(repo)

		createReq := httptest.NewRequest(http.MethodPost, "/permissions", strings.NewReader(`{"service_id":"svc-1","key":"orders:read"}`))
		createRec := httptest.NewRecorder()
		h.Create(createRec, createReq)

		req := httptest.NewRequest(http.MethodGet, "/permissions?service_id=svc-1", nil)
		rec := httptest.NewRecorder()

		h.ListByService(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusOK)
		}
	})
}

func TestPermissionHandler_Delete(t *testing.T) {
	t.Run("404 pour une permission inexistante", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		req := httptest.NewRequest(http.MethodDelete, "/permissions/inconnu", nil)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}
