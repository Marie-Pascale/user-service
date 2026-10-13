package application

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
)

func newTestRegisterUserUseCase() (*RegisterUserUseCase, *fakeUserRepo, *fakeAuthMethodRepo) {
	users := newFakeUserRepo()
	authMethods := newFakeAuthMethodRepo()
	hasher := fakeHasher{}

	createUser := NewCreateUserUseCase(users)
	registerAuthMethod := NewRegisterAuthMethodUseCase(users, authMethods, hasher)

	return NewRegisterUserUseCase(createUser, registerAuthMethod, users), users, authMethods
}

func TestRegisterUserUseCase_ExecuteLocal(t *testing.T) {
	t.Run("creates the user and their authentication method", func(t *testing.T) {
		uc, users, authMethods := newTestRegisterUserUseCase()

		u, err := uc.ExecuteLocal(context.Background(), "Alice", "alice@example.com", "s3cret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := users.GetByID(context.Background(), u.ID); err != nil {
			t.Errorf("the user should have been persisted: %v", err)
		}

		methods, err := authMethods.ListByUserID(context.Background(), u.ID)
		if err != nil {
			t.Fatalf("ListByUserID: %v", err)
		}
		if len(methods) != 1 || methods[0].Provider != domain.ProviderLocal {
			t.Errorf("methods = %+v, want one local method", methods)
		}
	})

	t.Run("deletes the user if the password is invalid (compensation)", func(t *testing.T) {
		uc, users, _ := newTestRegisterUserUseCase()

		_, err := uc.ExecuteLocal(context.Background(), "Alice", "alice@example.com", "")
		if err != domain.ErrEmptyPassword {
			t.Fatalf("error = %v, want %v", err, domain.ErrEmptyPassword)
		}

		_, err = users.GetByEmail(context.Background(), "alice@example.com")
		if !errors.Is(err, ErrNotFound) {
			t.Error("the user should have been deleted (compensation) after password creation failed")
		}
	})

	t.Run("rejects an email address that is already in use without changing auth methods", func(t *testing.T) {
		uc, _, authMethods := newTestRegisterUserUseCase()

		_, err := uc.ExecuteLocal(context.Background(), "Alice", "alice@example.com", "s3cret")
		if err != nil {
			t.Fatalf("first registration: %v", err)
		}

		_, err = uc.ExecuteLocal(context.Background(), "Alice Bis", "alice@example.com", "autre-mdp")
		if err != ErrEmailAlreadyUsed {
			t.Fatalf("error = %v, want %v", err, ErrEmailAlreadyUsed)
		}

		all, _ := authMethods.ListByUserID(context.Background(), "")
		_ = all
	})
}

func TestRegisterUserUseCase_ExecuteOAuth(t *testing.T) {
	uc, users, authMethods := newTestRegisterUserUseCase()

	u, err := uc.ExecuteOAuth(context.Background(), "Alice", "alice@example.com", domain.ProviderGoogle, "google-sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := users.GetByID(context.Background(), u.ID); err != nil {
		t.Errorf("the user should have been persisted: %v", err)
	}

	methods, err := authMethods.ListByUserID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(methods) != 1 || methods[0].Provider != domain.ProviderGoogle {
		t.Errorf("methods = %+v, want one Google method", methods)
	}
}
