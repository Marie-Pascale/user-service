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
	t.Run("repond 204 a une requete preflight OPTIONS sans toucher au handler", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"http://localhost:5173"})

		req := httptest.NewRequest(http.MethodOptions, "/users", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("reflete l'origine precise quand elle est dans la liste", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"http://localhost:5173"})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Errorf("Access-Control-Allow-Origin = %q, attendu %q", got, "http://localhost:5173")
		}
	})

	t.Run("n'ajoute aucun entete CORS pour une origine non autorisee", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"http://localhost:5173"})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://site-malveillant.example")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, attendu vide pour une origine non autorisee", got)
		}
	})

	t.Run("utilise le wildcard litteral quand \"*\" est configure", func(t *testing.T) {
		mw := NewCORSMiddleware([]string{"*"})

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("Origin", "http://n'importe-quoi.example")
		rec := httptest.NewRecorder()

		mw.Wrap(newOKHandler()).ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Access-Control-Allow-Origin = %q, attendu \"*\"", got)
		}
	})

	t.Run("laisse passer une requete normale (non-preflight) vers le handler", func(t *testing.T) {
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
			t.Error("le handler interne devrait avoir ete appele pour une requete non-preflight")
		}
	})
}
