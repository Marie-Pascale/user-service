package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORSMiddleware_Wrap(t *testing.T) {
	t.Run("returns 204 for an OPTIONS preflight request without invoking the handler", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"http://localhost:5173"})

		req := httptest.NewRequest(http.MethodOptions, "/users", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("reflects the exact origin when it is in the list", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"http://localhost:5173"})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:5173")
		}
	})

	t.Run("does not add CORS headers for a disallowed origin", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"http://localhost:5173"})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://site-malveillant.example")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want empty for a disallowed origin", got)
		}
	})

	t.Run("uses the literal wildcard when \"*\" is configured", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"*"})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://n'importe-quoi.example")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q, want \"*\"", got)
		}
	})

	t.Run("passes a normal (non-preflight) request to the handler", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"*"})
		called := false
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		mw.Wrap(inner).ServeHTTP(rec, req)

		if !called {
			t.Error("the inner handler should have been called for a non-preflight request")
		}
	})
}
