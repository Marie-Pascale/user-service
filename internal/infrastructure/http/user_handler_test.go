package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gestionnaire-usagers/internal/application"
	"gestionnaire-usagers/internal/domain"
)

func newTestUserHandler() (*UserHandler, *fakeUserRepo) {
	repo := newFakeUserRepo()
	createUser := application.NewCreateUserUseCase(repo)
	return NewUserHandler(createUser, repo), repo
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("201 pour une requete valide", func(t *testing.T) {
		h, _ := newTestUserHandler()

		body := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/users", body)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var resp userResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decodage de la reponse: %v", err)
		}
		if resp.Email != "alice@example.com" {
			t.Errorf("Email = %q, attendu %q", resp.Email, "alice@example.com")
		}
	})

	t.Run("400 pour un JSON invalide", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{pas du json valide`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("400 pour une validation du domaine echouee", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"alice@example.com"}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d (nom vide)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("409 pour un courriel deja utilise", func(t *testing.T) {
		h, repo := newTestUserHandler()
		_ = repo.Create(context.Background(), mustNewUser(t, "Alice", "alice@example.com"))

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice Bis","email":"alice@example.com"}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusConflict)
		}
	})
}

func TestUserHandler_Get(t *testing.T) {
	t.Run("200 pour un usager existant", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = repo.Create(context.Background(), u)

		req := httptest.NewRequest(http.MethodGet, "/users/"+u.ID, nil)
		req.SetPathValue("id", u.ID)
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("404 pour un usager inexistant", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodGet, "/users/inconnu", nil)
		req.SetPathValue("id", "inconnu")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestUserHandler_Delete(t *testing.T) {
	t.Run("204 pour une suppression reussie", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = repo.Create(context.Background(), u)

		req := httptest.NewRequest(http.MethodDelete, "/users/"+u.ID, nil)
		req.SetPathValue("id", u.ID)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusNoContent)
		}
	})
}

// mustNewUser est un raccourci pour les tests -- echoue immediatement (via
// Fatalf) si la construction rate, ce qui ne devrait jamais arriver avec
// des valeurs de test valides.
func mustNewUser(t *testing.T, name, email string) *domain.User {
	t.Helper()
	u, err := domain.NewUser(domain.NewID(), name, email)
	if err != nil {
		t.Fatalf("construction de l'usager de test: %v", err)
	}
	return u
}
