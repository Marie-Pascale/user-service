package application

import (
	"context"
	"errors"
	"testing"

	"user-service/internal/domain"
)

func setupUserForAuth(t *testing.T, repo *fakeUserRepo) *domain.User {
	t.Helper()
	u, err := domain.NewUser(domain.NewID(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("persistance: %v", err)
	}
	return u
}

func TestRegisterAuthMethodUseCase_ExecuteLocal(t *testing.T) {
	ctx := context.Background()

	t.Run("registers a valid local method", func(t *testing.T) {
		users := newFakeUserRepo()
		u := setupUserForAuth(t, users)
		uc := NewRegisterAuthMethodUseCase(users, newFakeAuthMethodRepo(), fakeHasher{})

		am, err := uc.ExecuteLocal(ctx, u.ID, "s3cret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if am.Provider != domain.ProviderLocal {
			t.Errorf("Provider = %q, want %q", am.Provider, domain.ProviderLocal)
		}
	})

	t.Run("rejects a nonexistent user", func(t *testing.T) {
		uc := NewRegisterAuthMethodUseCase(newFakeUserRepo(), newFakeAuthMethodRepo(), fakeHasher{})

		_, err := uc.ExecuteLocal(ctx, domain.NewID(), "s3cret")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("error = %v, want %v", err, ErrNotFound)
		}
	})
}

func TestRegisterAuthMethodUseCase_ExecuteOAuth(t *testing.T) {
	ctx := context.Background()

	t.Run("registers a valid OAuth method", func(t *testing.T) {
		users := newFakeUserRepo()
		u := setupUserForAuth(t, users)
		uc := NewRegisterAuthMethodUseCase(users, newFakeAuthMethodRepo(), fakeHasher{})

		am, err := uc.ExecuteOAuth(ctx, u.ID, domain.ProviderGoogle, "google-sub-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if am.ProviderUserID != "google-sub-123" {
			t.Errorf("ProviderUserID = %q, want %q", am.ProviderUserID, "google-sub-123")
		}
	})

	t.Run("rejects a third-party account already linked to another user", func(t *testing.T) {
		users := newFakeUserRepo()
		authMethods := newFakeAuthMethodRepo()
		uc := NewRegisterAuthMethodUseCase(users, authMethods, fakeHasher{})

		u1 := setupUserForAuth(t, users)
		u2, _ := domain.NewUser(domain.NewID(), "Bob", "bob@example.com")
		_ = users.Create(ctx, u2)

		if _, err := uc.ExecuteOAuth(ctx, u1.ID, domain.ProviderGoogle, "google-sub-shared"); err != nil {
			t.Fatalf("premier lien: %v", err)
		}

		_, err := uc.ExecuteOAuth(ctx, u2.ID, domain.ProviderGoogle, "google-sub-shared")
		if !errors.Is(err, ErrProviderAccountAlreadyLinked) {
			t.Errorf("error = %v, want %v", err, ErrProviderAccountAlreadyLinked)
		}
	})
}
