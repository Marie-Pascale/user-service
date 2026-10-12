package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"user-service/internal/application"
	"user-service/internal/domain"
)

func newTestUserHandler() (*UserHandler, *fakeUserRepo) {
	users := newFakeUserRepo()
	authMethods := newFakeAuthMethodRepo()
	hasher := fakeHasher{}

	createUser := application.NewCreateUserUseCase(users)
	registerAuthMethod := application.NewRegisterAuthMethodUseCase(users, authMethods, hasher)
	registerUser := application.NewRegisterUserUseCase(createUser, registerAuthMethod, users)

	return NewUserHandler(registerUser, users), users
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("201 pour une requete valide avec provider local", func(t *testing.T) {
		h, _ := newTestUserHandler()

		body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"local","password":"s3cret"}}`)
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

	t.Run("201 quand provider est omis (defaut = local)", func(t *testing.T) {
		h, _ := newTestUserHandler()

		body := strings.NewReader(`{"name":"Bob","email":"bob@example.com","auth_method":{"password":"s3cret"}}`)
		req := httptest.NewRequest(http.MethodPost, "/users", body)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, attendu %d (provider omis devrait defaulter a local). Corps: %s", rec.Code, http.StatusCreated, rec.Body.String())
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

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"alice@example.com","auth_method":{"provider":"local","password":"s3cret"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d (nom vide)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("400 pour un mot de passe vide", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"local"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d (mot de passe vide)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("400 pour un provider inconnu", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"mystere"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, attendu %d (provider inconnu)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("501 pour un provider tiers pas encore supporte", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"google","provider_user_id":"google-sub-123"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusNotImplemented {
			t.Errorf("status = %d, attendu %d (google pas encore supporte)", rec.Code, http.StatusNotImplemented)
		}
	})

	t.Run("409 pour un courriel deja utilise", func(t *testing.T) {
		h, repo := newTestUserHandler()
		_ = repo.Create(context.Background(), mustNewUser(t, "Alice", "alice@example.com"))

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice Bis","email":"alice@example.com","auth_method":{"provider":"local","password":"s3cret"}}`))
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

func TestUserHandler_GetBasicInfo(t *testing.T) {
	t.Run("200 avec seulement id/name/email/status", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		u.AssignRole("role-secret")
		_ = repo.Create(context.Background(), u)

		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		req := httptest.NewRequest(http.MethodGet, "/users/"+u.ID+"/profile", nil)
		req.SetPathValue("id", u.ID)
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.GetBasicInfo(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, attendu %d. Corps: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		body := rec.Body.String()
		if strings.Contains(body, "role_ids") || strings.Contains(body, "role-secret") {
			t.Errorf("la reponse ne devrait jamais contenir RoleIDs, corps: %s", body)
		}

		var resp userBasicInfoResponse
		if err := json.NewDecoder(strings.NewReader(body)).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if resp.Email != "alice@example.com" || resp.Name != "Alice" {
			t.Errorf("resp = %+v, attendu name=Alice email=alice@example.com", resp)
		}
	})

	t.Run("401 sans service authentifie dans le contexte", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = repo.Create(context.Background(), u)

		req := httptest.NewRequest(http.MethodGet, "/users/"+u.ID+"/profile", nil)
		req.SetPathValue("id", u.ID)
		rec := httptest.NewRecorder()

		h.GetBasicInfo(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, attendu %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("404 pour un usager inexistant", func(t *testing.T) {
		h, _ := newTestUserHandler()
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		req := httptest.NewRequest(http.MethodGet, "/users/inconnu/profile", nil)
		req.SetPathValue("id", "inconnu")
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.GetBasicInfo(rec, req)

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

func mustNewUser(t *testing.T, name, email string) *domain.User {
	t.Helper()
	u, err := domain.NewUser(domain.NewID(), name, email)
	if err != nil {
		t.Fatalf("construction de l'usager de test: %v", err)
	}
	return u
}
