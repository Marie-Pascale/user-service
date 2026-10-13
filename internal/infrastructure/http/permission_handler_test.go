package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPermissionHandler_Create(t *testing.T) {
	t.Run("201 for a valid request", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		body := strings.NewReader(`{"service_id":"svc-1","key":"orders:read","description":""}`)
		req := httptest.NewRequest(http.MethodPost, "/permissions", body)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})

	t.Run("400 for an empty key", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		req := httptest.NewRequest(http.MethodPost, "/permissions", strings.NewReader(`{"service_id":"svc-1","key":""}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPermissionHandler_ListByService(t *testing.T) {
	t.Run("400 when service_id is missing", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
		rec := httptest.NewRecorder()

		h.ListByService(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("200 when service_id is provided", func(t *testing.T) {
		repo := newFakePermissionRepo()
		h := NewPermissionHandler(repo)

		createReq := httptest.NewRequest(http.MethodPost, "/permissions", strings.NewReader(`{"service_id":"svc-1","key":"orders:read"}`))
		createRec := httptest.NewRecorder()
		h.Create(createRec, createReq)

		req := httptest.NewRequest(http.MethodGet, "/permissions?service_id=svc-1", nil)
		rec := httptest.NewRecorder()

		h.ListByService(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}

func TestPermissionHandler_Delete(t *testing.T) {
	t.Run("404 for a nonexistent permission", func(t *testing.T) {
		h := NewPermissionHandler(newFakePermissionRepo())

		req := httptest.NewRequest(http.MethodDelete, "/permissions/unknown", nil)
		req.SetPathValue("id", "unknown")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}
