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
	t.Run("201 for a valid request with the local provider", func(t *testing.T) {
		h, _ := newTestUserHandler()

		body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"local","password":"s3cret"}}`)
		req := httptest.NewRequest(http.MethodPost, "/users", body)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var resp userResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.Email != "alice@example.com" {
			t.Errorf("Email = %q, want %q", resp.Email, "alice@example.com")
		}
	})

	t.Run("201 when provider is omitted (default = local)", func(t *testing.T) {
		h, _ := newTestUserHandler()

		body := strings.NewReader(`{"name":"Bob","email":"bob@example.com","auth_method":{"password":"s3cret"}}`)
		req := httptest.NewRequest(http.MethodPost, "/users", body)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (omitted provider should default to local). Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})

	t.Run("400 for invalid JSON", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{this is not valid JSON`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("400 for a failed domain validation", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"","email":"alice@example.com","auth_method":{"provider":"local","password":"s3cret"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d (empty name)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("400 for an empty password", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"local"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d (empty password)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("400 for an unknown provider", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"mystere"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d (unknown provider)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("501 for a third-party provider that is not yet supported", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice","email":"alice@example.com","auth_method":{"provider":"google","provider_user_id":"google-sub-123"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusNotImplemented {
			t.Errorf("status = %d, want %d (Google not yet supported)", rec.Code, http.StatusNotImplemented)
		}
	})

	t.Run("409 for an email address that is already in use", func(t *testing.T) {
		h, repo := newTestUserHandler()
		_ = repo.Create(context.Background(), mustNewUser(t, "Alice", "alice@example.com"))

		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Alice Bis","email":"alice@example.com","auth_method":{"provider":"local","password":"s3cret"}}`))
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
		}
	})
}

func TestUserHandler_Get(t *testing.T) {
	t.Run("200 for an existing user", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = repo.Create(context.Background(), u)

		req := httptest.NewRequest(http.MethodGet, "/users/"+u.ID, nil)
		req.SetPathValue("id", u.ID)
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("404 for a nonexistent user", func(t *testing.T) {
		h, _ := newTestUserHandler()

		req := httptest.NewRequest(http.MethodGet, "/users/unknown", nil)
		req.SetPathValue("id", "unknown")
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestUserHandler_GetBasicInfo(t *testing.T) {
	t.Run("200 with only id/name/email/status", func(t *testing.T) {
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
			t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		body := rec.Body.String()
		if strings.Contains(body, "role_ids") || strings.Contains(body, "role-secret") {
			t.Errorf("the response should never contain RoleIDs, body: %s", body)
		}

		var resp userBasicInfoResponse
		if err := json.NewDecoder(strings.NewReader(body)).Decode(&resp); err != nil {
			t.Fatalf("decodage: %v", err)
		}
		if resp.Email != "alice@example.com" || resp.Name != "Alice" {
			t.Errorf("resp = %+v, want name=Alice email=alice@example.com", resp)
		}
	})

	t.Run("401 without an authenticated service in the context", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = repo.Create(context.Background(), u)

		req := httptest.NewRequest(http.MethodGet, "/users/"+u.ID+"/profile", nil)
		req.SetPathValue("id", u.ID)
		rec := httptest.NewRecorder()

		h.GetBasicInfo(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("404 for a nonexistent user", func(t *testing.T) {
		h, _ := newTestUserHandler()
		svc, _, _ := domain.NewService(domain.NewID(), "orders-service", "")

		req := httptest.NewRequest(http.MethodGet, "/users/unknown/profile", nil)
		req.SetPathValue("id", "unknown")
		req = req.WithContext(withTestService(req.Context(), svc))
		rec := httptest.NewRecorder()

		h.GetBasicInfo(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestUserHandler_Delete(t *testing.T) {
	t.Run("204 for a successful deletion", func(t *testing.T) {
		h, repo := newTestUserHandler()
		u := mustNewUser(t, "Alice", "alice@example.com")
		_ = repo.Create(context.Background(), u)

		req := httptest.NewRequest(http.MethodDelete, "/users/"+u.ID, nil)
		req.SetPathValue("id", u.ID)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})
}

func mustNewUser(t *testing.T, name, email string) *domain.User {
	t.Helper()
	u, err := domain.NewUser(domain.NewID(), name, email)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	return u
}
