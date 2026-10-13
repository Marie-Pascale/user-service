package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"user-service/internal/domain"
)

const testClientIP = "192.0.2.1"
const testUserAgent = "Mozilla/5.0 (test)"

func TestUserAuthMiddleware_Authenticate(t *testing.T) {
	newProtectedHandler := func(sessions *fakeSessionRepo, users *fakeUserRepo) http.Handler {
		mw := NewUserAuthMiddleware(sessions, users)
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				t.Fatal("the authenticated user should be present in the context")
			}
			writeJSON(w, http.StatusOK, map[string]string{"user_id": u.ID})
		})
		return mw.Authenticate(inner)
	}

	t.Run("allows a valid session (identical IP and User-Agent)", func(t *testing.T) {
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
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("allows the request when ONLY the IP differs (mobile network)", func(t *testing.T) {
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
			t.Fatalf("status = %d, want %d (a single change should not be blocking)", rec.Code, http.StatusOK)
		}
	})

	t.Run("rejects and revokes the session if BOTH IP and User-Agent differ", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = users.Create(context.Background(), u)

		session, plainToken, _ := domain.NewSession(domain.NewID(), u.ID, "203.0.113.9", "Mozilla/5.0 (other device)", time.Hour)
		_ = sessions.Create(context.Background(), session)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+plainToken)
		req.Header.Set("User-Agent", testUserAgent)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
		if _, ok := sessions.byID[session.ID]; ok {
			t.Error("a suspicious session (different IP + User-Agent) should have been revoked")
		}
	})

	t.Run("rejects a request without an Authorization header", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("rejects an invalid token", func(t *testing.T) {
		users := newFakeUserRepo()
		sessions := newFakeSessionRepo()

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer jeton-invente")
		rec := httptest.NewRecorder()

		newProtectedHandler(sessions, users).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("rejects and deletes an expired session", func(t *testing.T) {
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
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}

		if _, ok := sessions.byID[session.ID]; ok {
			t.Error("an expired session should have been deleted")
		}
	})
}
