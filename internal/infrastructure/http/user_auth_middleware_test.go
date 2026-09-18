package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gestionnaire-usagers/internal/domain"
)

const testClientIP = "192.0.2.1"
const testUserAgent = "Mozilla/5.0 (test)"

func TestUserAuthMiddleware_Authenticate(t *testing.T) {
	newProtectedHandler := func(sessions *fakeSessionRepo, users *fakeUserRepo) http.Handler {
		mw := NewUserAuthMiddleware(sessions, users)
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				t.Fatal("l'usager authentifie devrait etre present dans le contexte")
			}
			writeJSON(w, http.StatusOK, map[string]string{"user_id": u.ID})
		})
		return mw.Authenticate(inner)
	}

	t.Run("laisse passer une session valide (IP et User-Agent identiques)", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = users.Create(context.Background(), u)

		session, plainToken, _ := domain.NewSession(domain.NewID(), u.ID, testClientIP, testUserAgent, time.Hour)
		_ = sessions.Create(context.Background(), session)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+plainToken)
		req.Header.Set("User-Agent", testUserAgent)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("laisse passer si SEULE l'IP differe (reseau mobile)", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = users.Create(context.Background(), u)

		session, plainToken, _ := domain.NewSession(domain.NewID(), u.ID, "203.0.113.9", testUserAgent, time.Hour)
		_ = sessions.Create(context.Background(), session)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+plainToken)
		req.Header.Set("User-Agent", testUserAgent)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d (un seul changement ne devrait pas etre bloquant)", rec.Code, http.StatusOK)
		}
	})

	t.Run("rejette et revoque si IP ET User-Agent different tous les deux", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = users.Create(context.Background(), u)

		session, plainToken, _ := domain.NewSession(domain.NewID(), u.ID, "203.0.113.9", "Mozilla/5.0 (autre appareil)", time.Hour)
		_ = sessions.Create(context.Background(), session)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+plainToken)
		req.Header.Set("User-Agent", testUserAgent)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
		if _, ok := sessions.byID[session.ID]; ok {
			t.Error("une session suspecte (IP + User-Agent differents) devrait avoir ete revoquee")
		}
	})

	t.Run("rejette une requete sans entete Authorization", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("rejette un jeton invalide", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer jeton-invente")
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("rejette une session expiree et la supprime", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = users.Create(context.Background(), u)

		session, plainToken, _ := domain.NewSession(domain.NewID(), u.ID, testClientIP, testUserAgent, time.Hour)
		session.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		_ = sessions.Create(context.Background(), session)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+plainToken)
		req.Header.Set("User-Agent", testUserAgent)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}

		if _, ok := sessions.byID[session.ID]; ok {
			t.Error("une session expiree devrait avoir ete supprimee")
		}
	})
}
