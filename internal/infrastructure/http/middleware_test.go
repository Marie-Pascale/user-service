package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"user-service/internal/domain"
)

func TestServiceAuthMiddleware_Authenticate(t *testing.T) {
	newProtectedHandler := func(repo *fakeServiceRepo) http.Handler {
		mw := NewServiceAuthMiddleware(repo)
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			svc, ok := ServiceFromContext(r.Context())
			if !ok {
				t.Fatal("the authenticated service should be present in the context")
			}
			writeJSON(w, http.StatusOK, map[string]string{"service_id": svc.ID})
		})
		return mw.Authenticate(inner)
	}

	t.Run("allows a valid API key", func(t *testing.T) {
		repo := newFakeServiceRepo()
		svc, plainKey, _ := domain.NewService(domain.NewID(), "orders-service", "")
		_ = repo.Create(context.Background(), svc)

		req := httptest.NewRequest(http.MethodGet, "/users/u1/permissions", nil)
		req.Header.Set("X-API-Key", plainKey)
		rec := httptest.NewRecorder()

		newProtectedHandler(repo).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("rejects a request without an API key", func(t *testing.T) {
		repo := newFakeServiceRepo()

		req := httptest.NewRequest(http.MethodGet, "/users/u1/permissions", nil)
		rec := httptest.NewRecorder()

		newProtectedHandler(repo).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("rejects an invalid API key", func(t *testing.T) {
		repo := newFakeServiceRepo()
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")
		_ = repo.Create(context.Background(), svc)

		req := httptest.NewRequest(http.MethodGet, "/users/u1/permissions", nil)
		req.Header.Set("X-API-Key", "invented-nonexistent-key")
		rec := httptest.NewRecorder()

		newProtectedHandler(repo).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}
